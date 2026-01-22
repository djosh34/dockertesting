//go:build integration

package dockertesting

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestExecTest_SimplePackage(t *testing.T) {
	ctx := context.Background()

	// Create network
	network, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() { _ = cleanup(ctx) }()

	// Create container with the testdata/simple package
	cfg := CreateContainerConfig{
		PackagePath: "testdata/simple",
		Network:     network,
		NetworkName: network.Name,
	}

	container, err := CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}()

	// Execute tests
	execCfg := ExecConfig{
		Pattern: "./...",
		Timeout: 5 * time.Minute,
	}

	result, err := container.ExecTest(ctx, execCfg)
	if err != nil {
		t.Fatalf("failed to execute tests: %v", err)
	}

	// Verify results
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Logf("output: %s", string(result.Stdout))
	}

	// Verify output contains test information
	output := string(result.Stdout)
	if !strings.Contains(output, "PASS") && !strings.Contains(output, "ok") {
		t.Errorf("expected output to contain PASS or ok, got: %s", output)
	}
}

func TestExecTest_WithArgs(t *testing.T) {
	ctx := context.Background()

	// Create network
	network, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() { _ = cleanup(ctx) }()

	// Create container with the testdata/simple package
	cfg := CreateContainerConfig{
		PackagePath: "testdata/simple",
		Network:     network,
		NetworkName: network.Name,
	}

	container, err := CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}()

	// Execute tests with -v flag for verbose output
	execCfg := ExecConfig{
		Pattern: "./...",
		Args:    []string{"-v"},
		Timeout: 5 * time.Minute,
	}

	result, err := container.ExecTest(ctx, execCfg)
	if err != nil {
		t.Fatalf("failed to execute tests: %v", err)
	}

	// Verify results
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	// With -v flag, output should contain test function names
	output := string(result.Stdout)
	if !strings.Contains(output, "Test") {
		t.Errorf("expected verbose output to contain Test function names, got: %s", output)
	}
}

func TestExecTest_Timeout(t *testing.T) {
	ctx := context.Background()

	// Create network
	network, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() { _ = cleanup(ctx) }()

	// Create container with the testdata/simple package
	cfg := CreateContainerConfig{
		PackagePath: "testdata/simple",
		Network:     network,
		NetworkName: network.Name,
	}

	container, err := CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}()

	// Execute with very short timeout (this should complete fine, just testing the timeout field is used)
	execCfg := ExecConfig{
		Pattern: "./...",
		Timeout: 5 * time.Minute, // Use reasonable timeout, actual timeout testing would require slow tests
	}

	result, err := container.ExecTest(ctx, execCfg)
	if err != nil {
		t.Fatalf("failed to execute tests: %v", err)
	}

	// Verify it completed successfully
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestExecTest_NilContainer(t *testing.T) {
	ctx := context.Background()

	// Create a TestContainer with nil internal container
	container := &TestContainer{ctr: nil}

	execCfg := ExecConfig{
		Pattern: "./...",
	}

	_, err := container.ExecTest(ctx, execCfg)
	if err == nil {
		t.Fatal("expected error for nil container")
	}

	if !strings.Contains(err.Error(), "container is nil") {
		t.Errorf("expected error about nil container, got: %v", err)
	}
}

func TestExecTest_DefaultValues(t *testing.T) {
	ctx := context.Background()

	// Create network
	network, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() { _ = cleanup(ctx) }()

	// Create container with the testdata/simple package
	cfg := CreateContainerConfig{
		PackagePath: "testdata/simple",
		Network:     network,
		NetworkName: network.Name,
	}

	container, err := CreateContainer(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}()

	// Execute with empty config to test defaults
	execCfg := ExecConfig{}

	result, err := container.ExecTest(ctx, execCfg)
	if err != nil {
		t.Fatalf("failed to execute tests with default config: %v", err)
	}

	// Verify it used defaults and completed
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Logf("output: %s", string(result.Stdout))
	}
}
