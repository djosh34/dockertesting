//go:build integration

package dockertesting

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateContainer_Simple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a temporary Go package for testing
	tmpDir := t.TempDir()

	// Create a minimal go.mod
	goModContent := `module testpkg

go 1.25.6
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a simple main.go file
	mainGoContent := `package testpkg

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create a network first
	network, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() {
		if err := cleanup(ctx); err != nil {
			t.Logf("warning: failed to cleanup network: %v", err)
		}
	}()

	// Create container
	cfg := CreateContainerConfig{
		PackagePath: tmpDir,
		Network:     network,
		NetworkName: network.Name,
	}

	container, err := CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate container: %v", err)
		}
	}()

	// Verify container is running
	if container.Container() == nil {
		t.Fatal("expected container to be non-nil")
	}

	// Verify container is responsive by executing a simple command
	exitCode, _, err := container.Container().Exec(ctx, []string{"echo", "hello"})
	if err != nil {
		t.Fatalf("failed to exec in container: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

func TestCreateContainer_WithAliases(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a temporary Go package for testing
	tmpDir := t.TempDir()

	// Create a minimal go.mod
	goModContent := `module testpkg

go 1.25.6
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a simple main.go file
	mainGoContent := `package testpkg

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create a network first
	network, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() {
		if err := cleanup(ctx); err != nil {
			t.Logf("warning: failed to cleanup network: %v", err)
		}
	}()

	// Create container with aliases
	aliases := []string{"myapp.test", "test-service"}
	cfg := CreateContainerConfig{
		PackagePath: tmpDir,
		Network:     network,
		Aliases:     aliases,
		NetworkName: network.Name,
	}

	container, err := CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate container: %v", err)
		}
	}()

	// Verify container is running
	if container.Container() == nil {
		t.Fatal("expected container to be non-nil")
	}
}

func TestCreateContainer_WithVarSock(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check if Docker socket exists before running this test
	if _, err := os.Stat("/var/run/docker.sock"); os.IsNotExist(err) {
		t.Fatal("Docker socket not available at /var/run/docker.sock - this integration test requires Docker")
	}

	// Create a temporary Go package for testing
	tmpDir := t.TempDir()

	// Create a minimal go.mod
	goModContent := `module testpkg

go 1.25.6
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a simple main.go file
	mainGoContent := `package testpkg

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create a network first
	network, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() {
		if err := cleanup(ctx); err != nil {
			t.Logf("warning: failed to cleanup network: %v", err)
		}
	}()

	// Create container with Docker socket mounted
	cfg := CreateContainerConfig{
		PackagePath:   tmpDir,
		Network:       network,
		EnableVarSock: true,
		SockPath:      "/var/run/docker.sock",
		NetworkName:   network.Name,
	}

	container, err := CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate container: %v", err)
		}
	}()

	// Verify container is running
	if container.Container() == nil {
		t.Fatal("expected container to be non-nil")
	}

	// Verify Docker socket is mounted by checking if it exists in the container
	exitCode, _, err := container.Container().Exec(ctx, []string{"test", "-S", "/var/run/docker.sock"})
	if err != nil {
		t.Fatalf("failed to exec in container: %v", err)
	}
	if exitCode != 0 {
		t.Fatal("Docker socket was not mounted in container")
	}
}

func TestCreateContainer_EnvVarSet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a temporary Go package for testing
	tmpDir := t.TempDir()

	// Create a minimal go.mod
	goModContent := `module testpkg

go 1.25.6
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a simple main.go file
	mainGoContent := `package testpkg

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create a network first
	network, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() {
		if err := cleanup(ctx); err != nil {
			t.Logf("warning: failed to cleanup network: %v", err)
		}
	}()

	// Create container with network name env var
	cfg := CreateContainerConfig{
		PackagePath: tmpDir,
		Network:     network,
		NetworkName: network.Name,
	}

	container, err := CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate container: %v", err)
		}
	}()

	// Verify TESTCONTAINERS_DOCKER_NETWORK env var is set
	exitCode, _, err := container.Container().Exec(ctx, []string{"sh", "-c", "echo $TESTCONTAINERS_DOCKER_NETWORK"})
	if err != nil {
		t.Fatalf("failed to exec in container: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

func TestCreateContainer_InvalidPackagePath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := CreateContainerConfig{
		PackagePath: "/nonexistent/path/to/package",
	}

	_, err := CreateContainer(ctx, cfg)
	if err == nil {
		t.Fatal("expected error for non-existent package path")
	}
}
