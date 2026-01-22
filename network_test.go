package dockertesting

import (
	"context"
	"testing"
)

func TestCreateNetwork(t *testing.T) {
	ctx := context.Background()

	net, cleanup, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("createNetwork() error = %v, want nil", err)
	}

	// Verify network was created with a name
	if net.Name == "" {
		t.Error("createNetwork() returned empty network name")
	}

	t.Logf("Created network with name: %s", net.Name)

	// Cleanup the network
	if err := cleanup(ctx); err != nil {
		t.Fatalf("cleanup() error = %v, want nil", err)
	}
}

func TestCreateNetwork_CleanupViaMethod(t *testing.T) {
	ctx := context.Background()

	net, _, err := CreateNetwork(ctx)
	if err != nil {
		t.Fatalf("createNetwork() error = %v, want nil", err)
	}

	// Verify network was created with a name
	if net.Name == "" {
		t.Error("createNetwork() returned empty network name")
	}

	// Cleanup via the Remove method
	if err := net.Remove(ctx); err != nil {
		t.Fatalf("Remove() error = %v, want nil", err)
	}
}

func TestDockerNetwork_Remove_NilNetwork(t *testing.T) {
	// Test that Remove handles nil network gracefully
	dn := &DockerNetwork{
		Name:    "test",
		network: nil,
	}

	ctx := context.Background()
	if err := dn.Remove(ctx); err != nil {
		t.Errorf("Remove() on nil network error = %v, want nil", err)
	}
}
