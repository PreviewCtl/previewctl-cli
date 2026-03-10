package docker

import (
	"context"
	"fmt"

	"github.com/moby/moby/client"
)

// RemoveNetwork removes a Docker network by name. It is a no-op if the network
// does not exist.
func RemoveNetwork(ctx context.Context, cli *client.Client, networkName string) error {
	result, err := cli.NetworkList(ctx, client.NetworkListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	for _, n := range result.Items {
		if n.Name == networkName {
			if _, err := cli.NetworkRemove(ctx, n.ID, client.NetworkRemoveOptions{}); err != nil {
				return fmt.Errorf("failed to remove network %q: %w", networkName, err)
			}
			return nil
		}
	}

	return nil
}

// EnsureNetwork creates a bridge network with the given name if it doesn't already exist.
// Returns the network ID.
func EnsureNetwork(ctx context.Context, cli *client.Client, networkName string) (string, error) {
	// Check if network already exists
	result, err := cli.NetworkList(ctx, client.NetworkListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %w", err)
	}

	for _, n := range result.Items {
		if n.Name == networkName {
			return n.ID, nil
		}
	}

	resp, err := cli.NetworkCreate(ctx, networkName, client.NetworkCreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create network %q: %w", networkName, err)
	}

	return resp.ID, nil
}
