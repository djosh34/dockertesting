package dockertesting

import (
	"archive/tar"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
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

	// DockerfilePath is the path to a custom Dockerfile (optional).
	DockerfilePath string
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

	if _, err = os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package path does not exist: %s", absPath)
	}

	contextArchive, err := CreateTarContext(absPath, cfg.DockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar context: %w", err)
	}

	// Build container request
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			ContextArchive: contextArchive,
			Dockerfile:     "Dockerfile",
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

// CreateTarContext creates a tar archive of the contextPath directory,
// adding the Dockerfile from dockerfilePath.
// If dockerfilePath is empty, it adds the embedded Dockerfile template instead.
func CreateTarContext(contextPath string, dockerfilePath string) (io.ReadSeeker, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Get the Dockerfile content
	var dockerfileContent []byte
	if dockerfilePath == "" {
		// Use the default embedded Dockerfile template
		dockerfileContent = []byte(dockerfileTemplate)
	} else {
		// Read the custom Dockerfile
		// Support both relative (relative to contextPath) and absolute paths
		var fullPath string
		if filepath.IsAbs(dockerfilePath) {
			fullPath = dockerfilePath
		} else {
			fullPath = filepath.Join(contextPath, dockerfilePath)
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read custom Dockerfile at %s: %w", fullPath, err)
		}
		dockerfileContent = content
	}

	// Walk the context directory and add all files to the tar
	contextFS := os.DirFS(contextPath)
	err := fs.WalkDir(contextFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory entry
		if path == "." {
			return nil
		}

		// Skip any file named "Dockerfile" - we'll add our own
		if filepath.Base(path) == "Dockerfile" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", path, err)
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", path, err)
		}
		header.Name = path

		// Handle symlinks
		if info.Mode()&fs.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(filepath.Join(contextPath, path))
			if err != nil {
				return fmt.Errorf("failed to read symlink %s: %w", path, err)
			}
			header.Linkname = linkTarget
		}

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", path, err)
		}

		// For regular files, write the content
		if info.Mode().IsRegular() {
			fullPath := filepath.Join(contextPath, path)
			file, err := os.Open(fullPath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}

			_, copyErr := io.Copy(tw, file)
			closeErr := file.Close()

			if copyErr != nil {
				return fmt.Errorf("failed to write file content for %s: %w", path, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("failed to close file %s: %w", path, closeErr)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk context directory: %w", err)
	}

	// Add the Dockerfile to the tar archive
	dockerfileHeader := &tar.Header{
		Name: "Dockerfile",
		Mode: 0644,
		Size: int64(len(dockerfileContent)),
	}
	if err := tw.WriteHeader(dockerfileHeader); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile header: %w", err)
	}
	if _, err := tw.Write(dockerfileContent); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile content: %w", err)
	}

	// Close the tar writer
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
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
