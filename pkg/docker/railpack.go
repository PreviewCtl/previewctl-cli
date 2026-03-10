package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/previewctl/previewctl-cli/pkg/types"
)

// RailpackBuild builds a Docker image using the Railpack CLI.
// It shells out to `railpack build` in the same way BuildImage shells out to `docker build`.
func RailpackBuild(ctx context.Context, imageTag string, build types.BuildConfig, secrets map[string]string, workingDir string) error {
	if _, err := exec.LookPath("railpack"); err != nil {
		return fmt.Errorf("railpack CLI not found in PATH: %w", err)
	}

	contextDir := filepath.Join(workingDir, build.Context)

	info, err := os.Stat(contextDir)
	if err != nil {
		return fmt.Errorf("build context %q: %w", contextDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("build context %q is not a directory", contextDir)
	}

	if err := ensureBuildkit(ctx); err != nil {
		return fmt.Errorf("failed to start buildkit: %w", err)
	}

	fmt.Printf("  building %s with railpack (context: %s)...\n", imageTag, build.Context)

	args := []string{"build", "--name", imageTag}
	for k, v := range secrets {
		args = append(args, "--env", k+"="+v)
	}
	args = append(args, ".")

	cmd := exec.CommandContext(ctx, "railpack", args...)
	cmd.Dir = contextDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if os.Getenv("BUILDKIT_HOST") == "" {
		cmd.Env = append(cmd.Env, "BUILDKIT_HOST=docker-container://buildkitd")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("railpack build failed for %q: %w", imageTag, err)
	}

	return nil
}

const buildkitContainerName = "buildkitd"

// ensureBuildkit checks if the buildkitd container is running and starts it if not.
func ensureBuildkit(ctx context.Context) error {
	// Check if container exists and is running
	inspect := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", buildkitContainerName)
	out, err := inspect.Output()
	if err == nil && string(out) == "true\n" {
		return nil
	}

	// Remove stopped container if it exists
	_ = exec.CommandContext(ctx, "docker", "rm", "-f", buildkitContainerName).Run()

	fmt.Println("  starting buildkitd container...")
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--privileged", "-d", "--name", buildkitContainerName, "moby/buildkit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start buildkitd container: %w", err)
	}

	return nil
}
