package nested

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestGetMessage tests the basic GetMessage function.
func TestGetMessage(t *testing.T) {
	msg := GetMessage()
	if msg != "Hello from nested package" {
		t.Errorf("unexpected message: %s", msg)
	}
}

// TestNestedContainer tests spinning up an nginx container and connecting to it
// via a network alias. This test requires:
// 1. Docker socket to be mounted (for testcontainers-go to work)
// 2. TESTCONTAINERS_DOCKER_NETWORK env var to be set (to attach to the same network)
func TestNestedContainer(t *testing.T) {
	ctx := context.Background()

	// Get the network name from environment variable
	networkName := os.Getenv(NetworkEnvVar)
	if networkName == "" {
		t.Skipf("Skipping test: %s environment variable not set. "+
			"This test must be run inside dockertesting with WithVarSock() enabled.", NetworkEnvVar)
	}

	t.Logf("Using network: %s", networkName)

	// Create an nginx container attached to the provided network with an alias
	// Using Networks and NetworkAliases fields directly to attach to the existing network by name
	nginxContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nginx:alpine",
			ExposedPorts: []string{"80/tcp"},
			Networks:     []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {NestedServiceAlias},
			},
			WaitingFor: wait.ForHTTP("/").WithPort("80/tcp").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to create nginx container: %v", err)
	}
	defer func() {
		if err := nginxContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate nginx container: %v", err)
		}
	}()

	t.Logf("Nginx container started with alias: %s", NestedServiceAlias)

	// Make an HTTP request to the nginx container via its DNS alias
	url := fmt.Sprintf("http://%s", NestedServiceAlias)
	t.Logf("Making request to: %s", url)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to connect to %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	t.Logf("Successfully connected via DNS alias! Response length: %d bytes", len(body))

	// Verify we got the nginx welcome page
	if len(body) == 0 {
		t.Error("Response body is empty")
	}
}
