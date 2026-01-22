package dockertesting

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// dockerfileTemplate is the embedded Dockerfile template for building test containers.
//
//go:embed template.Dockerfile
var dockerfileTemplate string

// TestContainer wraps a testcontainers container for running Go tests.
type TestContainer struct {
	// container is the underlying testcontainers container.
	ctr testcontainers.Container
}

// CreateContainerConfig holds the configuration needed to create a test container.
type CreateContainerConfig struct {
	// PackagePath is the absolute path to the Go package to test.
	PackagePath string

	// Network is the Docker network to attach the container to.
	Network *DockerNetwork

	// Aliases are DNS aliases for the container within the network.
	Aliases []string

	// EnableVarSock enables mounting the Docker socket into the container.
	EnableVarSock bool

	// SockPath is the path to the Docker socket on the host.
	SockPath string

	// NetworkName is the name of the Docker network (for env var).
	NetworkName string
}

// CreateContainer builds and creates a Docker container for running Go tests.
// The container is built from the package at PackagePath using the embedded Dockerfile template.
// The container is attached to the provided network with optional aliases.
//
// The container starts with "sleep infinity" to keep it alive for executing tests via Exec.
//
// The caller is responsible for terminating the container by calling Terminate().
func CreateContainer(ctx context.Context, cfg CreateContainerConfig) (*TestContainer, error) {
	// Validate package path exists
	absPath, err := filepath.Abs(cfg.PackagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for package: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package path does not exist: %s", absPath)
	}

	// Write Dockerfile to the package directory temporarily
	dockerfilePath := filepath.Join(absPath, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileTemplate), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Ensure cleanup of temporary Dockerfile
	defer func() {
		_ = os.Remove(dockerfilePath)
	}()

	// Build container request
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    absPath,
			Dockerfile: "Dockerfile",
		},
		// Keep container alive for exec commands
		WaitingFor: wait.ForExec([]string{"echo", "ready"}),
	}

	// Set environment variables
	req.Env = make(map[string]string)
	if cfg.NetworkName != "" {
		req.Env["TESTCONTAINERS_DOCKER_NETWORK"] = cfg.NetworkName
	}

	// Configure network and aliases
	if cfg.Network != nil {
		req.Networks = []string{cfg.Network.Name}
		if len(cfg.Aliases) > 0 {
			req.NetworkAliases = map[string][]string{
				cfg.Network.Name: cfg.Aliases,
			}
		}
	}

	// Build the generic container request
	genReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	// Apply network option if network is provided
	if cfg.Network != nil && cfg.Network.Network() != nil {
		networkOpt := network.WithNetwork(cfg.Aliases, cfg.Network.Network())
		if err := networkOpt.Customize(&genReq); err != nil {
			return nil, fmt.Errorf("failed to apply network option: %w", err)
		}
	}

	// Mount Docker socket if enabled using HostConfigModifier
	if cfg.EnableVarSock {
		sockPath := cfg.SockPath
		if sockPath == "" {
			sockPath = DefaultSockPath
		}
		hostConfigOpt := testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.Mounts = append(hc.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: sockPath,
				Target: "/var/run/docker.sock",
			})
		})
		if err := hostConfigOpt.Customize(&genReq); err != nil {
			return nil, fmt.Errorf("failed to apply host config option: %w", err)
		}
	}

	// Create container using GenericContainer
	ctr, err := testcontainers.GenericContainer(ctx, genReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return &TestContainer{
		ctr: ctr,
	}, nil
}

// Terminate stops and removes the container.
func (c *TestContainer) Terminate(ctx context.Context) error {
	if c.ctr == nil {
		return nil
	}
	if err := c.ctr.Terminate(ctx); err != nil {
		return fmt.Errorf("failed to terminate container: %w", err)
	}
	return nil
}

// Container returns the underlying testcontainers.Container.
func (c *TestContainer) Container() testcontainers.Container {
	return c.ctr
}
