package dockertesting

import (
	"errors"
	"time"
)

// DefaultPattern is the default test pattern used when none is specified.
const DefaultPattern = "./..."

// DefaultSockPath is the default Docker socket path.
const DefaultSockPath = "/var/run/docker.sock"

// DefaultTimeout is the default timeout for test execution (10 minutes).
const DefaultTimeout = 10 * time.Minute

// Options holds the configuration for running tests in a Docker container.
type Options struct {
	// PackagePath is the path to the Go package to test (required).
	PackagePath string

	// Pattern is the test pattern to run (default: "./...").
	Pattern string

	// Args are additional arguments to pass to go test.
	Args []string

	// Aliases are DNS aliases for the container.
	Aliases []string

	// EnableVarSock enables mounting the Docker socket into the container.
	EnableVarSock bool

	// SockPath is the path to the Docker socket on the host (default: "/var/run/docker.sock").
	SockPath string

	// Timeout is the maximum duration for the entire test execution (default: 10 minutes).
	Timeout time.Duration
}

// Option is a functional option for configuring Options.
type Option func(*Options)

// WithPattern sets the test pattern to run. The pattern is passed to go test
// and follows the same syntax (e.g., "./...", "./pkg/...", "TestFoo").
// If not set, defaults to "./..." to run all tests.
//
// Example:
//
//	dockertesting.Run(ctx, path, dockertesting.WithPattern("./api/..."))
func WithPattern(pattern string) Option {
	return func(o *Options) {
		o.Pattern = pattern
	}
}

// WithArgs sets additional arguments to pass to go test. These are appended
// after the pattern. Common examples include "-v" for verbose output,
// "-race" for race detection, or "-count=1" to disable test caching.
//
// Multiple calls to WithArgs are cumulative.
//
// Example:
//
//	dockertesting.Run(ctx, path, dockertesting.WithArgs("-v", "-race"))
func WithArgs(args ...string) Option {
	return func(o *Options) {
		o.Args = append(o.Args, args...)
	}
}

// WithAliases sets DNS aliases for the container within the Docker network.
// Other containers on the same network can reach this container using these names.
// This is useful for tests that need to connect to services via custom hostnames.
//
// Multiple calls to WithAliases are cumulative.
//
// Example:
//
//	dockertesting.Run(ctx, path, dockertesting.WithAliases("myapp.test", "api.local"))
func WithAliases(aliases ...string) Option {
	return func(o *Options) {
		o.Aliases = append(o.Aliases, aliases...)
	}
}

// WithVarSock enables mounting the Docker socket into the container.
// This is required when the tests inside the container use testcontainers-go
// or otherwise need to interact with Docker.
//
// The socket is mounted from the host's SockPath (default: /var/run/docker.sock)
// to /var/run/docker.sock inside the container.
//
// When enabled, the TESTCONTAINERS_DOCKER_NETWORK environment variable is also
// set in the container, allowing nested testcontainers to attach to the same network.
//
// Example:
//
//	dockertesting.Run(ctx, path, dockertesting.WithVarSock())
func WithVarSock() Option {
	return func(o *Options) {
		o.EnableVarSock = true
	}
}

// WithSockPath sets the path to the Docker socket on the host.
// Only relevant when WithVarSock() is also used.
// Defaults to "/var/run/docker.sock".
//
// Example:
//
//	dockertesting.Run(ctx, path,
//	    dockertesting.WithVarSock(),
//	    dockertesting.WithSockPath("/custom/docker.sock"),
//	)
func WithSockPath(path string) Option {
	return func(o *Options) {
		o.SockPath = path
	}
}

// WithTimeout sets the maximum duration for the entire test execution.
// If the timeout is reached, the operation will be cancelled and a TimeoutError returned.
// Defaults to 10 minutes.
//
// Example:
//
//	dockertesting.Run(ctx, path, dockertesting.WithTimeout(5 * time.Minute))
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// NewOptions creates a new Options with the given package path and functional options.
// It returns an error if the package path is empty.
func NewOptions(packagePath string, opts ...Option) (*Options, error) {
	if packagePath == "" {
		return nil, errors.New("package path is required")
	}

	o := &Options{
		PackagePath: packagePath,
		Pattern:     DefaultPattern,
		SockPath:    DefaultSockPath,
		Timeout:     DefaultTimeout,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o, nil
}
