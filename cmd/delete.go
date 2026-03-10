package cmd

import (
	"errors"
	"fmt"

	"github.com/previewctl/previewctl-cli/internal/build/docker"
	"github.com/previewctl/previewctl-cli/internal/store"
	"github.com/previewctl/previewctl-cli/internal/store/database"
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
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrId := args[0]

		envStore := database.NewPreviewEnvironmentStore(DB)
		portStore := database.NewPortMappingStore(DB)
		var env *store.PreviewEnvironment
		var err error
		env, err = envStore.FindByName(cmd.Context(), nameOrId)
		if err != nil && !errors.Is(err, store.ErrResourceNotFound) {
			return err
		}

		if env == nil {
			env, err = envStore.Find(cmd.Context(), nameOrId)
			if err != nil && !errors.Is(err, store.ErrResourceNotFound) {
				return err
			}
		}

		if env == nil {
			return fmt.Errorf("preview environment %q not found", nameOrId)
		}

		// Tear down Docker resources
		cli, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create docker client: %w", err)
		}
		defer cli.Close()

		// Stop and remove all containers on the preview network
		removed, err := docker.StopAndRemoveContainersByNetwork(cmd.Context(), cli, env.Name)
		if err != nil {
			fmt.Printf("warning: %v\n", err)
		}
		for _, name := range removed {
			fmt.Printf("stopped %s\n", name)
		}

		fmt.Printf("removing network %q...\n", env.Name)
		if err := docker.RemoveNetwork(cmd.Context(), cli, env.Name); err != nil {
			fmt.Printf("  warning: %v\n", err)
		}

		// Clean up database records
		if err := portStore.DeleteByPreviewEnv(cmd.Context(), env.ID); err != nil {
			return fmt.Errorf("failed to delete port mappings: %w", err)
		}

		if err := envStore.Delete(cmd.Context(), env.ID); err != nil {
			return fmt.Errorf("failed to delete preview environment: %w", err)
		}

		fmt.Printf("preview environment %q deleted\n", nameOrId)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
