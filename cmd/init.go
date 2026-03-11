package cmd

import (
	"github.com/previewctl/previewctl-cli/internal/initializer"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize PreviewCtl in this repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		return initializer.InitRepo(workingDir)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
