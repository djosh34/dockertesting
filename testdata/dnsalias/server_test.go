package dnsalias

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestServerViaDNSAlias starts an HTTP server and connects to it via the DNS alias.
// This test will ONLY pass when:
// 1. Running inside a Docker container
// 2. The container has the DNS alias "myapp.test" configured via network aliases
//
// When running via `go test` directly (not in dockertest), this test will fail
// because "myapp.test" will not resolve.
func TestServerViaDNSAlias(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Start the server
	server, err := StartServer(port)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Ensure server shuts down when test completes
	defer server.Close()

	// Connect via DNS alias - this is the critical test
	// This URL will only work if the container has the alias "myapp.test"
	url := fmt.Sprintf("http://%s:%d/", DNSAlias, port)
	t.Logf("Attempting to connect to: %s", url)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to connect via DNS alias %s: %v\n"+
			"This test requires the container to have the DNS alias '%s' configured.\n"+
			"If running via dockertest, use WithAliases(\"%s\").",
			DNSAlias, err, DNSAlias, DNSAlias)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := fmt.Sprintf("Hello from %s!", DNSAlias)
	if string(body) != expected {
		t.Errorf("Expected body %q, got %q", expected, string(body))
	}

	t.Logf("Successfully connected via DNS alias: %s", DNSAlias)
	t.Logf("Response: %s", string(body))
}

// TestHelloHandler tests the handler in isolation (this test always passes).
// This provides some coverage even when DNS alias is not available.
func TestHelloHandler(t *testing.T) {
	handler := HelloHandler()

	// We're not testing the full HTTP flow here, just that the handler exists
	if handler == nil {
		t.Error("HelloHandler returned nil")
	}
}

// TestStartServer tests server creation (this test always passes).
func TestStartServer(t *testing.T) {
	server, err := StartServer(0)
	if err != nil {
		t.Fatalf("StartServer failed: %v", err)
	}
	if server == nil {
		t.Error("StartServer returned nil server")
	}
}

// TestConstants verifies the constants are set correctly.
func TestConstants(t *testing.T) {
	if DNSAlias != "myapp.test" {
		t.Errorf("DNSAlias = %q, want %q", DNSAlias, "myapp.test")
	}
	if DefaultPort != 8080 {
		t.Errorf("DefaultPort = %d, want %d", DefaultPort, 8080)
	}
}
