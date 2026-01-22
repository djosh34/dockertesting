package dockertesting

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/testcontainers/testcontainers-go/exec"
)

// TimeoutError represents an error that occurred due to a timeout.
type TimeoutError struct {
	Operation string
	Err       error
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout during %s: %v", e.Operation, e.Err)
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}

// wrapTimeoutError wraps an error with timeout context if the context was cancelled due to timeout.
func wrapTimeoutError(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return &TimeoutError{Operation: operation, Err: err}
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// Result holds the result of running tests in a Docker container.
type Result struct {
	// Stdout contains the combined stdout/stderr output from the test execution.
	Stdout []byte

	// Coverage contains the coverage file contents as bytes.
	// May be nil if coverage was not generated (e.g., tests failed early).
	Coverage []byte

	// ExitCode is the exit code from the test execution.
	// 0 indicates success, non-zero indicates test failures.
	ExitCode int
}

// Run executes go test for the given package path inside a Docker container.
//
// It creates a Docker network, builds and starts a container with the package,
// executes the tests with coverage profiling, and returns the results.
//
// The function ensures cleanup of all resources (network and container) regardless
// of success or failure. Stdout/stderr are forwarded to os.Stdout/os.Stderr in real-time.
//
// Example:
//
//	result, err := dockertesting.Run(ctx, "./mypackage",
//	    dockertesting.WithAliases("myapp.test"),
//	    dockertesting.WithVarSock(),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Exit code: %d\n", result.ExitCode)
//	fmt.Printf("Coverage:\n%s\n", result.Coverage)
func Run(ctx context.Context, packagePath string, opts ...Option) (*Result, error) {
	// Parse options
	options, err := NewOptions(packagePath, opts...)
	if err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Apply timeout to context if configured
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Create network
	network, cleanupNetwork, err := CreateNetwork(ctx)
	if err != nil {
		return nil, wrapTimeoutError(ctx, err, "create network")
	}

	// Ensure network cleanup always happens
	defer func() {
		if cleanupNetwork != nil {
			_ = cleanupNetwork(ctx)
		}
	}()

	// Create container
	container, err := CreateContainer(ctx, CreateContainerConfig{
		PackagePath:   options.PackagePath,
		Network:       network,
		Aliases:       options.Aliases,
		EnableVarSock: options.EnableVarSock,
		SockPath:      options.SockPath,
		NetworkName:   network.Name,
	})
	if err != nil {
		return nil, wrapTimeoutError(ctx, err, "create container")
	}

	// Ensure container cleanup always happens
	defer func() {
		if container != nil {
			_ = container.Terminate(ctx)
		}
	}()

	// Execute tests with real-time output forwarding
	result, err := execTestWithStreaming(ctx, container, options)
	if err != nil {
		return nil, wrapTimeoutError(ctx, err, "execute tests")
	}

	// Copy coverage file from container
	coverage, err := container.CopyCoverage(ctx)
	if err != nil {
		// Non-fatal: coverage may not exist if tests failed early
		coverage = nil
	}

	return &Result{
		Stdout:   result.Stdout,
		Coverage: coverage,
		ExitCode: result.ExitCode,
	}, nil
}

// execTestWithStreaming executes tests and streams output to stdout in real-time.
func execTestWithStreaming(ctx context.Context, container *TestContainer, options *Options) (*ExecResult, error) {
	if container.ctr == nil {
		return nil, fmt.Errorf("container is nil")
	}

	// Build the go test command
	cmd := []string{
		"go", "test",
		"-coverprofile=" + DefaultCoverageFile,
		options.Pattern,
	}
	// Append additional arguments
	cmd = append(cmd, options.Args...)

	// Execute the command in the container with multiplexed output
	exitCode, reader, err := container.ctr.Exec(ctx, cmd, exec.Multiplexed())
	if err != nil {
		return nil, wrapTimeoutError(ctx, err, "execute test command")
	}

	// Stream output to os.Stdout while also capturing it
	var output []byte
	if reader != nil {
		// Create a TeeReader to write to stdout while also capturing the output
		output, err = io.ReadAll(io.TeeReader(reader, os.Stdout))
		if err != nil {
			return nil, wrapTimeoutError(ctx, err, "read test output")
		}
	}

	return &ExecResult{
		Stdout:   output,
		ExitCode: exitCode,
	}, nil
}
