package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/previewctl/previewctl-cli/internal/identity"
	"github.com/previewctl/previewctl-cli/internal/store"
	"github.com/previewctl/previewctl-cli/internal/store/database"
	"github.com/previewctl/previewctl-cli/internal/up"
	"github.com/previewctl/previewctl-core/secrets"
	"github.com/previewctl/previewctl-core/validator"
	"github.com/spf13/cobra"
)

var (
	previewID    string
	secretInputs []string
	envFile      string
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Build and deploy preview services to Docker",
	Long: `Read .previewctl/preview.yml, resolve service dependencies, and start
the preview stack in Docker.

The up command will build services (for example Dockerfile and Nixpacks
builds), create the runtime network, and deploy all configured services in the
required dependency order.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		branch, err := currentGitBranch(workingDir)
		if err != nil {
			fmt.Printf("failed to determine git branch: %v\n", err)
			fmt.Println("using default branch name: main")
			branch = "main"
		}

		config, err := validator.LoadAndValidateConfig(workingDir)
		if err != nil {
			return err
		}

		resolutionSecrets, userSecrets, err := getSecrets()
		if err != nil {
			return err
		}

		envStore := database.NewPreviewEnvironmentStore(DB)
		portStore := database.NewPortMappingStore(DB)

		previewEnvName, err := resolvePreviewEnv(cmd.Context(), envStore, branch)
		if err != nil {
			return err
		}
		previewEnv, err := updatePreviewEnvPre(cmd.Context(), envStore, previewEnvName, branch)
		if err != nil {
			return err
		}

		if err := up.HandleUp(cmd.Context(), previewEnvName, previewEnv.ID, config, resolutionSecrets, userSecrets, portStore, workingDir); err != nil {
			return err
		}

		if err := envStore.UpdateStatus(cmd.Context(), previewEnv.ID, "active"); err != nil {
			return fmt.Errorf("failed to update preview environment status: %w", err)
		}

		return nil
	},
}

// currentGitBranch returns the current branch name for the given directory.
func currentGitBranch(dir string) (string, error) {
	gitCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	gitCmd.Dir = dir
	out, err := gitCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// getSecrets returns two maps:
//   - resolutionSecrets: OS env + .env + flags (for resolving ${secrets.X} in config)
//   - userSecrets: .env + flags only (for injecting into containers and builds)
func getSecrets() (map[string]string, map[string]string, error) {
	if envFile == "" {
		envFile = filepath.Join(workingDir, ".env")
	}

	envSecrets, err := secrets.ParseEnvFile(envFile)
	if err != nil {
		return nil, nil, err
	}

	flagSecrets, err := secrets.ParseKeyValues(secretInputs)
	if err != nil {
		return nil, nil, err
	}

	userSecrets := secrets.Merge(envSecrets, flagSecrets)
	resolutionSecrets := secrets.Merge(secrets.ParseOSEnv(), userSecrets)

	return resolutionSecrets, userSecrets, nil
}

// resovlve Preview env
func resolvePreviewEnv(ctx context.Context, envStore *database.PreviewEnvironmentStore, branch string) (string, error) {
	if strings.TrimSpace(previewID) != "" {
		return identity.ResolvePreviewID(previewID, workingDir, branch)
	}

	previews, err := envStore.FindByWorkspaceAndBranch(ctx, workingDir, branch)
	if err != nil && !errors.Is(store.ErrResourceNotFound, err) {
		return "", err
	}
	if previews != nil {
		return previews.Name, nil
	}
	return identity.ResolvePreviewID(previewID, workingDir, branch)
}

// updatePreviewEnvPre
func updatePreviewEnvPre(ctx context.Context, envStore *database.PreviewEnvironmentStore, previewEnvName string, branch string) (*store.PreviewEnvironment, error) {
	existing, err := envStore.FindByName(ctx, previewEnvName)
	if err != nil {
		if errors.Is(store.ErrResourceNotFound, err) {
			created, err := envStore.Create(ctx, &store.PreviewEnvironment{
				Name:      previewEnvName,
				Workspace: workingDir,
				Branch:    branch,
				Status:    "deploying",
			})
			if err != nil {
				return nil, fmt.Errorf("failed to store preview environment: %w", err)
			}
			return created, nil
		} else {
			return nil, err
		}
	}

	if err := envStore.UpdateStatus(ctx, existing.ID, "deploying"); err != nil {
		return nil, fmt.Errorf("failed to update preview environment status: %w", err)
	}
	return existing, nil
}

func init() {
	rootCmd.AddCommand(upCmd)

	upCmd.Flags().StringVar(&previewID, "preview-id", "", "Preview ID to deploy (defaults to generated value)")
	upCmd.Flags().StringArrayVar(&secretInputs, "secret", nil, "Secret in KEY=VALUE format (optional, repeatable)")
	upCmd.Flags().StringVar(&envFile, "env-file", "", "Path to .env file (defaults to .env in working directory)")
}
