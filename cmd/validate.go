package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/previewctl/previewctl-cli/pkg/constants"
	"github.com/previewctl/previewctl-cli/pkg/types"
	"github.com/previewctl/previewctl-cli/pkg/validator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

		configPath := constants.PreviewCtrlConfigFilePath(workingDir)
		rawConfig, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file at %s: %w", configPath, err)
		}

		decoder := yaml.NewDecoder(bytes.NewReader(rawConfig))
		decoder.KnownFields(true)

		var config types.PreviewConfig
		if err := decoder.Decode(&config); err != nil {
			return fmt.Errorf("invalid yaml schema in %s: %w", configPath, err)
		}

		if err := validator.ValidateConfig(config); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

		fmt.Println("preview config is valid")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
