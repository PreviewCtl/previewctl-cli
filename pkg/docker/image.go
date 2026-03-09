package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/moby/moby/client"
)

// EnsureImage pulls the image if it is not already present locally.
func EnsureImage(ctx context.Context, cli *client.Client, imageName string) error {
	// Check if image exists locally
	_, err := cli.ImageInspect(ctx, imageName)
	if err == nil {
		return nil // already present
	}

	fmt.Printf("  pulling %s...\n", imageName)
	resp, err := cli.ImagePull(ctx, imageName, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %q: %w", imageName, err)
	}
	defer resp.Close()

	// Consume the pull output to wait for completion
	_, err = io.Copy(io.Discard, resp)
	if err != nil {
		return fmt.Errorf("error reading pull output for %q: %w", imageName, err)
	}

	return nil
}
