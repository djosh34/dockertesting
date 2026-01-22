# dockertesting

A Go library for running `go test` inside Docker containers using [testcontainers-go](https://golang.testcontainers.org/).

## Why

Some tests require capabilities that are difficult or impossible to achieve when running directly on the host:

- Custom DNS resolution so containers can reach each other by hostname
- Nested testcontainers where tests spin up their own Postgres, Redis, etc. and need to communicate with them via the same Docker network
- Isolated environments that do not affect the host

This library wraps your Go package in a Docker container, runs the tests inside, and returns the results. All Docker resources (networks, containers) are cleaned up automatically.

## Install

```
go get github.com/djosh34/dockertesting
```

## Basic Usage

Run tests for a package inside a Docker container:

```go
// From integration_test.go TestRun_SimplePackage
packagePath, _ := filepath.Abs("testdata/simple")
result, err := dockertesting.Run(ctx, packagePath)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Exit code: %d\n", result.ExitCode)
```

## DNS Aliases

Make the test container reachable via custom hostnames within the Docker network. This is useful when tests need to connect to themselves or other services via specific DNS names.

```go
// From integration_test.go TestRun_DNSAlias
packagePath, _ := filepath.Abs("testdata/dnsalias")
result, err := dockertesting.Run(ctx, packagePath, dockertesting.WithAliases("myapp.test"))
```

The test inside the container can then make HTTP requests to `http://myapp.test:port/` and the DNS will resolve correctly within the Docker network.

## Nested Testcontainers

When your tests use testcontainers-go internally to spin up additional containers (databases, message queues, etc.), you need two things:

1. Docker socket access so testcontainers-go can create containers
2. The network name so nested containers attach to the same network

Enable this with `WithVarSock()`:

```go
// From integration_test.go TestRun_NestedTestcontainers
packagePath, _ := filepath.Abs("testdata/nested")
result, err := dockertesting.Run(ctx, packagePath, dockertesting.WithVarSock())
```

When `WithVarSock()` is enabled:

- The Docker socket (`/var/run/docker.sock`) is mounted into the container
- The `TESTCONTAINERS_DOCKER_NETWORK` environment variable is set to the network name

Your nested tests read this environment variable to attach their containers to the same network:

```go
// From testdata/nested/nested_test.go TestNestedContainer
networkName := os.Getenv("TESTCONTAINERS_DOCKER_NETWORK")
if networkName == "" {
    t.Skip("TESTCONTAINERS_DOCKER_NETWORK not set")
}

nginxContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
    ContainerRequest: testcontainers.ContainerRequest{
        Image:    "nginx:alpine",
        Networks: []string{networkName},
        NetworkAliases: map[string][]string{
            networkName: {"nested-service.test"},
        },
    },
    Started: true,
})
```

This allows the outer container (running your tests) and the inner container (nginx in this example) to communicate via DNS aliases.

## Options

All options use the functional options pattern and can be combined:

```go
result, err := dockertesting.Run(ctx, packagePath,
    dockertesting.WithPattern("./..."),
    dockertesting.WithArgs("-v", "-race"),
    dockertesting.WithAliases("myapp.test", "api.local"),
    dockertesting.WithVarSock(),
    dockertesting.WithTimeout(5 * time.Minute),
)
```

## WithPattern

Set the test pattern passed to `go test`. Defaults to `./...`.

```go
dockertesting.WithPattern("./api/...")
```

## WithArgs

Pass additional arguments to `go test`. Multiple calls are cumulative.

```go
dockertesting.WithArgs("-v", "-race", "-count=1")
```

## WithAliases

Add DNS aliases for the container. Multiple calls are cumulative. Other containers on the same network can reach this container using these hostnames.

```go
dockertesting.WithAliases("myapp.test", "api.local")
```

## WithVarSock

Mount the Docker socket into the container. Required when tests use testcontainers-go or otherwise need Docker access. Also sets `TESTCONTAINERS_DOCKER_NETWORK` env var.

```go
dockertesting.WithVarSock()
```

## WithSockPath

Override the Docker socket path on the host. Only relevant when using `WithVarSock()`. Defaults to `/var/run/docker.sock`.

```go
dockertesting.WithSockPath("/custom/docker.sock")
```

## WithTimeout

Set the maximum duration for the entire test execution. Defaults to 10 minutes.

```go
dockertesting.WithTimeout(5 * time.Minute)
```

## Result

The `Run` function returns a `Result` struct:

```go
type Result struct {
    Stdout   []byte  // Combined stdout/stderr from test execution
    Coverage []byte  // Coverage profile bytes from -coverprofile
    ExitCode int     // Exit code from go test (0 = success)
}
```

## Coverage Retrieval

Coverage is automatically collected via `-coverprofile` and returned in `Result.Coverage`. The coverage file is written to `/tmp/coverage.txt` inside the container and copied out after test execution.

## Real-time Output

Stdout and stderr are forwarded to `os.Stdout` and `os.Stderr` in real-time during test execution. The output is also captured and returned in `Result.Stdout`.

## Cleanup

All Docker resources are cleaned up automatically via deferred cleanup functions, regardless of success or failure. No manual cleanup is required.

## Timeout Handling

When a timeout occurs, the error returned is wrapped as a `TimeoutError`:

```go
result, err := dockertesting.Run(ctx, packagePath, dockertesting.WithTimeout(1*time.Second))
if err != nil {
    var timeoutErr *dockertesting.TimeoutError
    if errors.As(err, &timeoutErr) {
        fmt.Printf("Operation %s timed out\n", timeoutErr.Operation)
    }
}
```

## Run Tests

```
make help              # show available commands
make test              # run unit tests
make test-integration  # run integration tests (requires Docker)
```

## Links

- [testcontainers-go documentation](https://golang.testcontainers.org/)
