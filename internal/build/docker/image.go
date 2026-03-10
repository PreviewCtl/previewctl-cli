package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/moby/moby/client"
	"github.com/previewctl/previewctl-cli/common/types"
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

// BuildImage builds a Docker image by shelling out to `docker build`.
// This supports BuildKit and all Dockerfile features (e.g. RUN --mount).
func BuildImage(ctx context.Context, imageTag string, build types.BuildConfig, secrets map[string]string, workingDir string) error {
	contextDir := filepath.Join(workingDir, build.Context)

	info, err := os.Stat(contextDir)
	if err != nil {
		return fmt.Errorf("build context %q: %w", contextDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("build context %q is not a directory", contextDir)
	}

	dockerfile := build.Dockerfile
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	fmt.Printf("  building %s (context: %s)...\n", imageTag, build.Context)

	args := []string{"build", "-t", imageTag, "-f", dockerfile}
	for k, v := range secrets {
		args = append(args, "--build-arg", k+"="+v)
	}
	args = append(args, ".")

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = contextDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed for %q: %w", imageTag, err)
	}

	return nil
}
