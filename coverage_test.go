//go:build integration

package dockertesting

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCopyCoverage_AfterSuccessfulTest(t *testing.T) {
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

	// Execute tests to generate coverage
	execCfg := ExecConfig{
		Pattern: "./...",
		Timeout: 5 * time.Minute,
	}

	result, err := container.ExecTest(ctx, execCfg)
	if err != nil {
		t.Fatalf("failed to execute tests: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("tests failed with exit code %d: %s", result.ExitCode, string(result.Stdout))
	}

	// Now copy the coverage file
	coverage, err := container.CopyCoverage(ctx)
	if err != nil {
		t.Fatalf("failed to copy coverage: %v", err)
	}

	// Verify coverage content is non-empty
	if len(coverage) == 0 {
		t.Error("expected non-empty coverage file")
	}

	// Verify coverage file has expected format (starts with "mode:")
	coverageStr := string(coverage)
	if !strings.HasPrefix(coverageStr, "mode:") {
		t.Errorf("coverage file should start with 'mode:', got: %s", coverageStr[:min(50, len(coverageStr))])
	}

	t.Logf("Coverage file content (%d bytes):\n%s", len(coverage), coverageStr)
}

func TestCopyCoverageFromPath_CustomPath(t *testing.T) {
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

	// Execute tests with custom coverage path
	customPath := "/tmp/custom_coverage.txt"
	execCfg := ExecConfig{
		Pattern:      "./...",
		CoverageFile: customPath,
		Timeout:      5 * time.Minute,
	}

	result, err := container.ExecTest(ctx, execCfg)
	if err != nil {
		t.Fatalf("failed to execute tests: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("tests failed with exit code %d: %s", result.ExitCode, string(result.Stdout))
	}

	// Now copy the coverage file from custom path
	coverage, err := container.CopyCoverageFromPath(ctx, customPath)
	if err != nil {
		t.Fatalf("failed to copy coverage from custom path: %v", err)
	}

	// Verify coverage content is non-empty
	if len(coverage) == 0 {
		t.Error("expected non-empty coverage file")
	}

	// Verify coverage file has expected format
	coverageStr := string(coverage)
	if !strings.HasPrefix(coverageStr, "mode:") {
		t.Errorf("coverage file should start with 'mode:', got: %s", coverageStr[:min(50, len(coverageStr))])
	}
}

func TestCopyCoverageFromPath_EmptyPath(t *testing.T) {
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

	if result.ExitCode != 0 {
		t.Fatalf("tests failed with exit code %d", result.ExitCode)
	}

	// Test that empty path defaults to DefaultCoverageFile
	coverage, err := container.CopyCoverageFromPath(ctx, "")
	if err != nil {
		t.Fatalf("failed to copy coverage with empty path: %v", err)
	}

	if len(coverage) == 0 {
		t.Error("expected non-empty coverage file when using empty path (should default)")
	}
}

func TestCopyFileFromContainer_NonExistentFile(t *testing.T) {
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

	// Try to copy a file that doesn't exist
	content, err := container.CopyFileFromContainer(ctx, "/nonexistent/file.txt")

	// Should return nil content and nil error for non-existent files
	if err != nil {
		t.Errorf("expected nil error for non-existent file, got: %v", err)
	}

	if content != nil {
		t.Errorf("expected nil content for non-existent file, got: %v", content)
	}
}

func TestCopyFileFromContainer_NilContainer(t *testing.T) {
	ctx := context.Background()

	// Create a TestContainer with nil internal container
	container := &TestContainer{ctr: nil}

	_, err := container.CopyFileFromContainer(ctx, "/tmp/coverage.txt")
	if err == nil {
		t.Fatal("expected error for nil container")
	}

	if !strings.Contains(err.Error(), "container is nil") {
		t.Errorf("expected error about nil container, got: %v", err)
	}
}

func TestCopyCoverage_NilContainer(t *testing.T) {
	ctx := context.Background()

	// Create a TestContainer with nil internal container
	container := &TestContainer{ctr: nil}

	_, err := container.CopyCoverage(ctx)
	if err == nil {
		t.Fatal("expected error for nil container")
	}

	if !strings.Contains(err.Error(), "container is nil") {
		t.Errorf("expected error about nil container, got: %v", err)
	}
}

func TestCopyCoverage_BeforeTestExecution(t *testing.T) {
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

	// Try to copy coverage before running tests (file shouldn't exist)
	coverage, err := container.CopyCoverage(ctx)

	// Should return nil content and nil error since file doesn't exist
	if err != nil {
		t.Errorf("expected nil error when coverage file doesn't exist, got: %v", err)
	}

	if coverage != nil {
		t.Errorf("expected nil coverage when file doesn't exist, got: %v", coverage)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
