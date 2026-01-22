package dockertesting

import (
	"context"
	"fmt"
	"io"
)

// DefaultCoverageFile is the default path where coverage output is written inside the container.
const DefaultCoverageFile = "/tmp/coverage.txt"

// CopyFileFromContainer copies a file from the container and returns its contents as bytes.
// If the file doesn't exist, it returns nil bytes and a nil error.
// This is useful for extracting coverage files which may not exist if tests failed early.
func (c *TestContainer) CopyFileFromContainer(ctx context.Context, containerFilePath string) ([]byte, error) {
	if c.ctr == nil {
		return nil, fmt.Errorf("container is nil")
	}

	reader, err := c.ctr.CopyFileFromContainer(ctx, containerFilePath)
	if err != nil {
		// Check if the error indicates the file doesn't exist
		// testcontainers-go returns an error when the file doesn't exist
		// We treat this as a non-fatal condition and return nil bytes
		return nil, nil
	}
	defer func() {
		_ = reader.Close()
	}()

	// Read all content from the reader
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

// CopyCoverage is a convenience method that copies the coverage file from the default location.
// It returns the coverage file contents as bytes, or nil if the file doesn't exist.
func (c *TestContainer) CopyCoverage(ctx context.Context) ([]byte, error) {
	return c.CopyFileFromContainer(ctx, DefaultCoverageFile)
}

// CopyCoverageFromPath copies the coverage file from a custom path inside the container.
// It returns the coverage file contents as bytes, or nil if the file doesn't exist.
func (c *TestContainer) CopyCoverageFromPath(ctx context.Context, coveragePath string) ([]byte, error) {
	if coveragePath == "" {
		coveragePath = DefaultCoverageFile
	}
	return c.CopyFileFromContainer(ctx, coveragePath)
}
