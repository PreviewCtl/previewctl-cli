package cmd

import (
	"fmt"

	"github.com/previewctl/previewctl-cli/internal/store"
	"github.com/spf13/cobra"
)

var listAll bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List preview environments",
	Long: `List preview environments tracked by PreviewCtl.

By default only previews belonging to the current workspace are shown.
Use --all to display every preview environment across all workspaces.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		var envs []*store.PreviewEnvironment

		if listAll {
			list, err := envStore.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list preview environments: %w", err)
			}
			envs = list
		} else {
			list, err := envStore.ListByWorkspace(ctx, workingDir)
			if err != nil {
				return fmt.Errorf("failed to list preview environments: %w", err)
			}
			envs = list
		}

		if len(envs) == 0 {
			fmt.Println("No preview environments found.")
			return nil
		}

		printEnvTable(envs)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "List preview environments across all workspaces")
}
