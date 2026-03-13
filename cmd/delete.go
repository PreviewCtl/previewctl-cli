package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/previewctl/previewctl-cli/internal/build/docker"
	"github.com/previewctl/previewctl-cli/internal/store"
	"github.com/previewctl/previewctl-core/constants"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <name or id>",
	Short: "Delete a preview environment",
	Long: `Delete a preview environment by name.

This stops and removes all Docker containers and the network associated with
the preview, then removes the environment and its port mappings from the
local database.`,
	Aliases: []string{"rm"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		var env *store.PreviewEnvironment
		var err error
		if len(args) == 1 {
			nameOrId := strings.TrimSpace(args[0])
			env, err = findEnvByNameOrID(ctx, nameOrId)
			if err != nil {
				return err
			}
		} else {
			env, err = findCurrentPreviewIfOnce(ctx)
			if err != nil {
				return err
			}
		}

		// Tear down Docker resources
		cli, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create docker client: %w", err)
		}
		defer cli.Close()

		// Stop and remove all containers on the preview network
		removed, err := docker.StopAndRemoveContainersByNetwork(ctx, cli, env.Name)
		if err != nil {
			fmt.Printf("warning: %v\n", err)
		}
		for _, name := range removed {
			fmt.Printf("stopped %s\n", name)
		}

		fmt.Printf("removing network %q...\n", env.Name)
		if err := docker.RemoveNetwork(ctx, cli, env.Name); err != nil {
			fmt.Printf("  warning: %v\n", err)
		}

		// Clean up preview data directory
		dataDir := filepath.Join(constants.PreviewCtlConfigDirPath(workingDir), "data", env.Name)
		if err := os.RemoveAll(dataDir); err != nil {
			fmt.Printf("  warning: failed to remove data directory: %v\n", err)
		}

		// Clean up database records
		if err := portStore.DeleteByPreviewEnv(ctx, env.ID); err != nil {
			return fmt.Errorf("failed to delete port mappings: %w", err)
		}

		if err := secretStore.DeleteByPreviewEnv(ctx, env.ID); err != nil {
			return fmt.Errorf("failed to delete generated secrets: %w", err)
		}

		if err := envStore.Delete(ctx, env.ID); err != nil {
			return fmt.Errorf("failed to delete preview environment: %w", err)
		}

		fmt.Printf("preview environment %s : %q deleted\n", env.ID, env.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
