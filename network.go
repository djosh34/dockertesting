package dockertesting

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// DockerNetwork wraps a testcontainers Docker network and provides
// access to the network name and cleanup functionality.
type DockerNetwork struct {
	// Name is the name of the Docker network.
	Name string

	// network is the underlying testcontainers network.
	network *testcontainers.DockerNetwork
}

// CreateNetwork creates a new Docker network using testcontainers-go.
// The network is created with an auto-generated name and can be used
// to attach containers.
//
// The caller is responsible for cleaning up the network by calling
// the cleanup function returned, or by calling network.Remove(ctx).
func CreateNetwork(ctx context.Context) (*DockerNetwork, func(context.Context) error, error) {
	net, err := network.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create docker network: %w", err)
	}

	dn := &DockerNetwork{
		Name:    net.Name,
		network: net,
	}

	cleanup := func(ctx context.Context) error {
		if err := net.Remove(ctx); err != nil {
			return fmt.Errorf("failed to remove docker network: %w", err)
		}
		return nil
	}

	return dn, cleanup, nil
}

// Remove removes the Docker network. This should be called when the
// network is no longer needed to clean up resources.
func (n *DockerNetwork) Remove(ctx context.Context) error {
	if n.network == nil {
		return nil
	}
	if err := n.network.Remove(ctx); err != nil {
		return fmt.Errorf("failed to remove docker network: %w", err)
	}
	return nil
}

// Network returns the underlying testcontainers.DockerNetwork for use
// with network.WithNetwork() when attaching containers.
func (n *DockerNetwork) Network() *testcontainers.DockerNetwork {
	return n.network
}
