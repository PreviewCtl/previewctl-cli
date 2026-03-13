package cmd

import (
	"fmt"
	"strings"

	"github.com/previewctl/previewctl-cli/internal/build/docker"
	"github.com/previewctl/previewctl-cli/internal/store"
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
	Args: cobra.MaximumNArgs(1),
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

		if err := envStore.UpdateStatus(ctx, env.ID, "stopped"); err != nil {
			return fmt.Errorf("failed to update environment status: %w", err)
		}

		fmt.Printf("preview environment %s : %q stopped\n", env.ID, env.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
