/*
Package dockertesting provides a Go library for running go test inside Docker containers.

This library enables running tests for any Go package inside a Docker container with
custom DNS aliases, network configuration, and support for nested testcontainers.
It is particularly useful when tests need:

  - Custom DNS resolution via container network aliases
  - Tests that spin up their own testcontainers (e.g., PostgreSQL, Redis) while
    communicating with them via the same Docker network
  - Isolated test environments that don't affect the host

# Basic Usage

The main entry point is the Run function which takes a package path and options:

	result, err := dockertesting.Run(ctx, "./mypackage")
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Printf("Exit code: %d\n", result.ExitCode)
	fmt.Printf("Coverage:\n%s\n", result.Coverage)

# DNS Aliases

To make the test container accessible via custom DNS names within the network:

	result, err := dockertesting.Run(ctx, "./mypackage",
	    dockertesting.WithAliases("myapp.test", "api.local"),
	)

# Nested Testcontainers

For tests that use testcontainers-go internally, enable Docker socket mounting:

	result, err := dockertesting.Run(ctx, "./mypackage",
	    dockertesting.WithVarSock(),
	)

The TESTCONTAINERS_DOCKER_NETWORK environment variable is automatically set in the
container, allowing nested testcontainers to attach to the same network.

# Additional Options

Configure test patterns, timeouts, and additional go test arguments:

	result, err := dockertesting.Run(ctx, "./mypackage",
	    dockertesting.WithPattern("./..."),
	    dockertesting.WithArgs("-v", "-race"),
	    dockertesting.WithTimeout(5 * time.Minute),
	)

# Results

The Run function returns a Result containing:
  - Stdout: Combined stdout/stderr from the test execution
  - Coverage: The coverage profile bytes (from -coverprofile)
  - ExitCode: The exit code from go test (0 = success)

Output is also streamed to os.Stdout/os.Stderr in real-time during execution.

# Cleanup

All Docker resources (networks and containers) are automatically cleaned up
regardless of success or failure.
*/
package dockertesting
