package dockertesting

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/testcontainers/testcontainers-go/exec"
)

// DefaultExecTimeout is the default timeout for test execution.
const DefaultExecTimeout = 10 * time.Minute

// ExecConfig holds the configuration for executing tests in the container.
type ExecConfig struct {
	// Pattern is the test pattern to run (e.g., "./...").
	Pattern string

	// Args are additional arguments to pass to go test.
	Args []string

	// CoverageFile is the path inside the container where coverage output is written.
	CoverageFile string

	// Timeout is the maximum duration for test execution.
	Timeout time.Duration
}

// ExecResult holds the result of test execution.
type ExecResult struct {
	// Stdout contains the combined stdout/stderr output from the test execution.
	Stdout []byte

	// ExitCode is the exit code from the test execution.
	// 0 indicates success, non-zero indicates failure.
	ExitCode int
}

// ExecTest runs `go test` inside the container and returns the result.
// The command includes coverage profiling to /tmp/coverage.txt by default.
//
// The method captures stdout/stderr and returns them along with the exit code.
// A non-zero exit code typically indicates test failures.
func (c *TestContainer) ExecTest(ctx context.Context, cfg ExecConfig) (*ExecResult, error) {
	if c.ctr == nil {
		return nil, fmt.Errorf("container is nil")
	}

	// Apply defaults
	if cfg.Pattern == "" {
		cfg.Pattern = DefaultPattern
	}
	if cfg.CoverageFile == "" {
		cfg.CoverageFile = "/tmp/coverage.txt"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultExecTimeout
	}

	// Build the go test command
	cmd := []string{
		"go", "test",
		"-coverprofile=" + cfg.CoverageFile,
		cfg.Pattern,
	}
	// Append additional arguments
	cmd = append(cmd, cfg.Args...)

	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	// Execute the command in the container
	// Using Multiplexed() to combine stdout and stderr into a single stream
	exitCode, reader, err := c.ctr.Exec(execCtx, cmd, exec.Multiplexed())
	if err != nil {
		// Check if this is a context timeout error
		if execCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("test execution timed out after %v: %w", cfg.Timeout, err)
		}
		return nil, fmt.Errorf("failed to execute test command: %w", err)
	}

	// Read all output
	var output []byte
	if reader != nil {
		output, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read test output: %w", err)
		}
	}

	return &ExecResult{
		Stdout:   output,
		ExitCode: exitCode,
	}, nil
}
