package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/previewctl/previewctl-cli/internal/store"
	"github.com/previewctl/previewctl-cli/internal/store/database"
	"github.com/spf13/cobra"
)

var listAll bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List preview environments",
	Long: `List preview environments tracked by PreviewCtrl.

By default only previews belonging to the current workspace are shown.
Use --all to display every preview environment across all workspaces.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		envStore := database.NewPreviewEnvironmentStore(DB)

		var envs []*store.PreviewEnvironment

		if listAll {
			list, err := envStore.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list preview environments: %w", err)
			}
			envs = list
		} else {
			list, err := envStore.ListByWorkspace(cmd.Context(), workingDir)
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

func printEnvTable(envs []*store.PreviewEnvironment) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tBRANCH\tSTATUS\tWORKSPACE\tCREATED")
	for _, e := range envs {
		created := time.Unix(e.CreatedAt, 0).Format(time.DateTime)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", e.ID, e.Name, e.Branch, e.Status, e.Workspace, created)
	}
	w.Flush()
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "List preview environments across all workspaces")
}
