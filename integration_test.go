//go:build integration

package dockertesting

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestRun_SimplePackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get absolute path to testdata/simple
	packagePath, err := filepath.Abs("testdata/simple")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// Run tests in Docker container
	result, err := Run(ctx, packagePath)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Verify exit code is 0 (tests passed)
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Logf("stdout:\n%s", string(result.Stdout))
	}

	// Verify coverage bytes are non-empty
	if len(result.Coverage) == 0 {
		t.Error("expected coverage to be non-empty")
	} else {
		// Verify coverage has valid content (starts with "mode:")
		if !strings.HasPrefix(string(result.Coverage), "mode:") {
			t.Errorf("coverage does not look valid, expected to start with 'mode:', got: %s", string(result.Coverage)[:min(50, len(result.Coverage))])
		}
		t.Logf("coverage file (%d bytes):\n%s", len(result.Coverage), string(result.Coverage))
	}

	// Verify stdout contains test output
	stdout := string(result.Stdout)
	if len(stdout) == 0 {
		t.Error("expected stdout to be non-empty")
	}

	// Check for typical go test output indicators
	// Note: Output may contain PASS, ok, or test function names
	if !strings.Contains(stdout, "PASS") && !strings.Contains(stdout, "ok") {
		t.Errorf("stdout does not contain expected test output (PASS or ok), got:\n%s", stdout)
	}

	t.Logf("stdout:\n%s", stdout)
}

func TestRun_DNSAlias(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get absolute path to testdata/dnsalias
	packagePath, err := filepath.Abs("testdata/dnsalias")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// Run tests in Docker container with DNS alias "myapp.test"
	result, err := Run(ctx, packagePath, WithAliases("myapp.test"))
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Verify exit code is 0 (tests passed)
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Logf("stdout:\n%s", string(result.Stdout))
	}

	// Verify stdout contains test output
	stdout := string(result.Stdout)
	if len(stdout) == 0 {
		t.Error("expected stdout to be non-empty")
	}

	// Verify that the DNS alias test was able to connect successfully
	// The test output should indicate successful connection via the alias
	if !strings.Contains(stdout, "PASS") && !strings.Contains(stdout, "ok") {
		t.Errorf("stdout does not contain expected test output (PASS or ok), got:\n%s", stdout)
	}

	// Check for indications of successful DNS alias connection
	if strings.Contains(stdout, "Successfully connected via DNS alias") {
		t.Logf("DNS alias connection confirmed in output")
	}

	t.Logf("stdout:\n%s", stdout)
}

func TestRun_NestedTestcontainers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Get absolute path to testdata/nested
	packagePath, err := filepath.Abs("testdata/nested")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// Run tests in Docker container with Docker socket mounted
	// This allows the nested package to spin up its own testcontainers
	result, err := Run(ctx, packagePath, WithVarSock())
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Verify exit code is 0 (tests passed)
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Logf("stdout:\n%s", string(result.Stdout))
	}

	// Verify stdout contains test output
	stdout := string(result.Stdout)
	if len(stdout) == 0 {
		t.Error("expected stdout to be non-empty")
	}

	// Verify that the tests passed
	if !strings.Contains(stdout, "PASS") && !strings.Contains(stdout, "ok") {
		t.Errorf("stdout does not contain expected test output (PASS or ok), got:\n%s", stdout)
	}

	// Check for indications of successful nested container connection
	// The nested test should have connected to the nginx container via its alias
	if strings.Contains(stdout, "TestNestedContainer") {
		t.Logf("TestNestedContainer test was executed")
	}

	// Verify coverage bytes are non-empty (tests actually ran)
	if len(result.Coverage) == 0 {
		t.Error("expected coverage to be non-empty")
	} else {
		t.Logf("coverage file (%d bytes):\n%s", len(result.Coverage), string(result.Coverage))
	}

	t.Logf("stdout:\n%s", stdout)
}

// TestRun_NoGarbageFilesProduced verifies that running tests does not leave
// any garbage files (coverage.txt, binaries, temp files) in the working directory.
// NOTE: This test does NOT use t.Parallel() because it checks for filesystem
// state before and after Run(), which would be affected by other parallel tests.
func TestRun_NoGarbageFilesProduced(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get absolute path to testdata/simple
	packagePath, err := filepath.Abs("testdata/simple")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// Record files in package directory BEFORE running tests
	filesBefore, err := listFiles(packagePath)
	if err != nil {
		t.Fatalf("failed to list files before Run(): %v", err)
	}

	// Also record files in the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	cwdFilesBefore, err := listFiles(cwd)
	if err != nil {
		t.Fatalf("failed to list files in cwd before Run(): %v", err)
	}

	// Run tests in Docker container
	result, err := Run(ctx, packagePath)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Verify exit code is 0 (tests passed)
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Logf("stdout:\n%s", string(result.Stdout))
	}

	// Record files AFTER running tests
	filesAfter, err := listFiles(packagePath)
	if err != nil {
		t.Fatalf("failed to list files after Run(): %v", err)
	}
	cwdFilesAfter, err := listFiles(cwd)
	if err != nil {
		t.Fatalf("failed to list files in cwd after Run(): %v", err)
	}

	// Verify no new files were created in the package directory
	newFilesInPackage := findNewFiles(filesBefore, filesAfter)
	if len(newFilesInPackage) > 0 {
		t.Errorf("garbage files were created in package directory: %v", newFilesInPackage)
	}

	// Verify no new files were created in the current working directory
	// Filter out expected transient files (like .testcontainers, Docker logs)
	newFilesInCwd := findNewFiles(cwdFilesBefore, cwdFilesAfter)
	garbageFiles := filterNonTransientFiles(newFilesInCwd)
	if len(garbageFiles) > 0 {
		t.Errorf("garbage files were created in working directory: %v", garbageFiles)
	}

	t.Logf("Package directory verified clean: %d files before, %d files after", len(filesBefore), len(filesAfter))
	t.Logf("Working directory verified clean: %d files before, %d files after", len(cwdFilesBefore), len(cwdFilesAfter))
}

// listFiles returns a sorted list of files in the given directory (non-recursive).
func listFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	sort.Strings(files)
	return files, nil
}

// findNewFiles returns files that are in 'after' but not in 'before'.
func findNewFiles(before, after []string) []string {
	beforeSet := make(map[string]bool)
	for _, f := range before {
		beforeSet[f] = true
	}
	var newFiles []string
	for _, f := range after {
		if !beforeSet[f] {
			newFiles = append(newFiles, f)
		}
	}
	return newFiles
}

// filterNonTransientFiles filters out known transient files that may be
// created by the test framework or Docker and are not considered garbage.
func filterNonTransientFiles(files []string) []string {
	// Files/directories that are known to be transient and not garbage
	transient := map[string]bool{
		".testcontainers":      true, // testcontainers config dir
		".testcontainers.yaml": true, // testcontainers config file
	}
	var nonTransient []string
	for _, f := range files {
		if !transient[f] {
			nonTransient = append(nonTransient, f)
		}
	}
	return nonTransient
}
