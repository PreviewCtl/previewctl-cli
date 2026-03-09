package cmd

import (
	"fmt"
	"os"

	"github.com/previewctl/previewctl-cli/pkg/validator"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate .previewctrl/preview.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

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
