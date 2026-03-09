package cmd

import (
	"github.com/previewctl/previewctl-cli/internal/initializer"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize PreviewCtrl in this repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		return initializer.InitRepo()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
