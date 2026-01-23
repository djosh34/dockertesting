package dockertesting

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTarContext_DefaultDockerfile(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create some test files
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.25.6\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create tar with default Dockerfile
	reader, err := CreateTarContext(tmpDir, "")
	if err != nil {
		t.Fatalf("CreateTarContext failed: %v", err)
	}

	// Read and verify the tar contents
	files := readTarContents(t, reader)

	// Should contain go.mod, main.go, and Dockerfile
	if len(files) != 3 {
		t.Errorf("expected 3 files in tar, got %d: %v", len(files), getFileNames(files))
	}

	// Verify Dockerfile contains the default template content
	dockerfile, ok := files["Dockerfile"]
	if !ok {
		t.Fatal("Dockerfile not found in tar")
	}
	if !strings.Contains(dockerfile, "ARG GO_VERSION") {
		t.Error("Dockerfile does not contain expected ARG GO_VERSION")
	}
	if !strings.Contains(dockerfile, "golang:${GO_VERSION}") {
		t.Error("Dockerfile does not contain expected FROM golang:${GO_VERSION}")
	}

	// Verify go.mod is present
	if _, ok := files["go.mod"]; !ok {
		t.Error("go.mod not found in tar")
	}

	// Verify main.go is present
	if _, ok := files["main.go"]; !ok {
		t.Error("main.go not found in tar")
	}
}

func TestCreateTarContext_CustomDockerfile_RelativePath(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create some test files
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.25.6\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a custom Dockerfile
	customDockerfileContent := "FROM alpine:latest\nRUN echo 'custom'\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "custom.Dockerfile"), []byte(customDockerfileContent), 0644); err != nil {
		t.Fatalf("failed to write custom.Dockerfile: %v", err)
	}

	// Create tar with custom Dockerfile (relative path)
	reader, err := CreateTarContext(tmpDir, "custom.Dockerfile")
	if err != nil {
		t.Fatalf("CreateTarContext failed: %v", err)
	}

	// Read and verify the tar contents
	files := readTarContents(t, reader)

	// Verify Dockerfile contains the custom content
	dockerfile, ok := files["Dockerfile"]
	if !ok {
		t.Fatal("Dockerfile not found in tar")
	}
	if !strings.Contains(dockerfile, "FROM alpine:latest") {
		t.Error("Dockerfile does not contain expected custom content")
	}
	if !strings.Contains(dockerfile, "RUN echo 'custom'") {
		t.Error("Dockerfile does not contain expected custom RUN command")
	}

	// The original custom.Dockerfile should also be in the tar (since we didn't exclude it)
	if _, ok := files["custom.Dockerfile"]; !ok {
		t.Error("custom.Dockerfile not found in tar")
	}
}

func TestCreateTarContext_CustomDockerfile_AbsolutePath(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the context
	contextDir := t.TempDir()

	// Create a separate directory for the custom Dockerfile
	dockerfileDir := t.TempDir()

	// Create some test files in context
	if err := os.WriteFile(filepath.Join(contextDir, "go.mod"), []byte("module test\n\ngo 1.25.6\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a custom Dockerfile in separate directory
	customDockerfileContent := "FROM golang:1.21\nWORKDIR /custom\n"
	customDockerfilePath := filepath.Join(dockerfileDir, "external.Dockerfile")
	if err := os.WriteFile(customDockerfilePath, []byte(customDockerfileContent), 0644); err != nil {
		t.Fatalf("failed to write external.Dockerfile: %v", err)
	}

	// Create tar with custom Dockerfile (absolute path)
	reader, err := CreateTarContext(contextDir, customDockerfilePath)
	if err != nil {
		t.Fatalf("CreateTarContext failed: %v", err)
	}

	// Read and verify the tar contents
	files := readTarContents(t, reader)

	// Verify Dockerfile contains the custom content
	dockerfile, ok := files["Dockerfile"]
	if !ok {
		t.Fatal("Dockerfile not found in tar")
	}
	if !strings.Contains(dockerfile, "FROM golang:1.21") {
		t.Error("Dockerfile does not contain expected custom FROM")
	}
	if !strings.Contains(dockerfile, "WORKDIR /custom") {
		t.Error("Dockerfile does not contain expected WORKDIR")
	}
}

func TestCreateTarContext_InvalidDockerfilePath(t *testing.T) {
	t.Parallel()

	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create tar with non-existent Dockerfile path
	_, err := CreateTarContext(tmpDir, "nonexistent.Dockerfile")
	if err == nil {
		t.Fatal("expected error for non-existent Dockerfile path, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read custom Dockerfile") {
		t.Errorf("expected error message to mention 'failed to read custom Dockerfile', got: %v", err)
	}
}

func TestCreateTarContext_ExcludesExistingDockerfile(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create a Dockerfile that should be excluded
	originalDockerfile := "FROM original:latest\nRUN echo 'should be excluded'\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(originalDockerfile), 0644); err != nil {
		t.Fatalf("failed to write Dockerfile: %v", err)
	}

	// Create some test files
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.25.6\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create tar with default Dockerfile (should replace existing)
	reader, err := CreateTarContext(tmpDir, "")
	if err != nil {
		t.Fatalf("CreateTarContext failed: %v", err)
	}

	// Read and verify the tar contents
	files := readTarContents(t, reader)

	// Verify Dockerfile contains the default template content, NOT the original
	dockerfile, ok := files["Dockerfile"]
	if !ok {
		t.Fatal("Dockerfile not found in tar")
	}
	if strings.Contains(dockerfile, "FROM original:latest") {
		t.Error("Dockerfile contains original content - should have been replaced")
	}
	if strings.Contains(dockerfile, "should be excluded") {
		t.Error("Dockerfile contains original RUN command - should have been replaced")
	}
	if !strings.Contains(dockerfile, "ARG GO_VERSION") {
		t.Error("Dockerfile does not contain expected default template content")
	}
}

func TestCreateTarContext_WithSubdirectory(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with nested structure
	tmpDir := t.TempDir()

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create files
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "helper.go"), []byte("package subdir\n"), 0644); err != nil {
		t.Fatalf("failed to write helper.go: %v", err)
	}

	// Create tar
	reader, err := CreateTarContext(tmpDir, "")
	if err != nil {
		t.Fatalf("CreateTarContext failed: %v", err)
	}

	// Read and verify the tar contents
	files := readTarContents(t, reader)

	// Verify subdirectory and file are present
	if _, ok := files["subdir"]; !ok {
		t.Error("subdir not found in tar")
	}
	if _, ok := files["subdir/helper.go"]; !ok {
		t.Error("subdir/helper.go not found in tar")
	}
}

func TestWithDockerfilePath(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithDockerfilePath("./custom.Dockerfile"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.DockerfilePath != "./custom.Dockerfile" {
		t.Errorf("expected DockerfilePath './custom.Dockerfile', got %q", opts.DockerfilePath)
	}
}

func TestNewOptions_DefaultDockerfilePath(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.DockerfilePath != "" {
		t.Errorf("expected DockerfilePath to be empty by default, got %q", opts.DockerfilePath)
	}
}

// Helper function to read tar contents into a map
func readTarContents(t *testing.T, reader io.ReadSeeker) map[string]string {
	t.Helper()

	// Seek to beginning
	if _, err := reader.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek reader: %v", err)
	}

	tr := tar.NewReader(reader)
	files := make(map[string]string)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read tar header: %v", err)
		}

		switch header.Typeflag {
		case tar.TypeReg:
			content, err := io.ReadAll(tr)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", header.Name, err)
			}
			files[header.Name] = string(content)
		case tar.TypeDir:
			files[header.Name] = ""
		}
	}

	return files
}

// Helper function to get file names from the files map
func getFileNames(files map[string]string) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}
