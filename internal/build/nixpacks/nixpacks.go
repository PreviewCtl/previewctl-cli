package nixpacks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/previewctl/previewctl-cli/common/types"
)

// NixpacksBuild builds a Docker image using the Nixpacks CLI.
// It shells out to `nixpacks build` in the same way BuildImage shells out to `docker build`.
func NixpacksBuild(ctx context.Context, imageTag string, build types.BuildConfig, secrets map[string]string, workingDir string) error {
	if _, err := exec.LookPath("nixpacks"); err != nil {
		return fmt.Errorf("nixpacks CLI not found in PATH: %w", err)
	}

	contextDir := filepath.Join(workingDir, build.Context)

	info, err := os.Stat(contextDir)
	if err != nil {
		return fmt.Errorf("build context %q: %w", contextDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("build context %q is not a directory", contextDir)
	}

	fmt.Printf("  building %s with nixpacks (context: %s)...\n", imageTag, build.Context)

	args := []string{"build", ".", "--name", imageTag}
	for k, v := range secrets {
		args = append(args, "--env", k+"="+v)
	}

	cmd := exec.CommandContext(ctx, "nixpacks", args...)
	cmd.Dir = contextDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nixpacks build failed for %q: %w", imageTag, err)
	}

	return nil
}
