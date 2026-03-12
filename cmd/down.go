package cmd

import (
	"errors"
	"fmt"

	"github.com/previewctl/previewctl-cli/internal/build/docker"
	"github.com/previewctl/previewctl-cli/internal/store"
	"github.com/previewctl/previewctl-cli/internal/store/database"
	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down <name or id>",
	Short: "Stop a preview environment",
	Long: `Stop a preview environment by name.

This stops and removes all Docker containers and the network, but preserves
the environment record, port mappings, and data directory so the preview can
be brought back up with "previewctl up".

Use "previewctl delete" to permanently remove a preview environment.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrId := args[0]

		envStore := database.NewPreviewEnvironmentStore(DB)
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

		if err := envStore.UpdateStatus(cmd.Context(), env.ID, "stopped"); err != nil {
			return fmt.Errorf("failed to update environment status: %w", err)
		}

		fmt.Printf("preview environment %q stopped\n", nameOrId)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
