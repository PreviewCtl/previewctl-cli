package cmd

import (
	"fmt"

	"github.com/previewctl/previewctl-core/validator"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate .previewctrl/preview.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := validator.LoadAndValidateConfig(workingDir); err != nil {
			return err
		}

		fmt.Println("preview config is valid")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
