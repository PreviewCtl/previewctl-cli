package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/previewctl/previewctl-cli/internal/up"
	"github.com/previewctl/previewctl-cli/pkg/identity"
	"github.com/previewctl/previewctl-cli/pkg/secrets"
	"github.com/previewctl/previewctl-cli/pkg/validator"
	"github.com/spf13/cobra"
)

var (
	previewID    string
	secretInputs []string
	envFile      string
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Build and deploy preview services to Docker",
	Long: `Read .previewctrl/preview.yml, resolve service dependencies, and start
the preview stack in Docker.

The up command will build services (for example Dockerfile and Nixpacks
builds), create the runtime network, and deploy all configured services in the
required dependency order.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		resolvedPreviewID, generated, err := identity.ResolvePreviewID(previewID, workingDir)
		if err != nil {
			return err
		}
		if generated {
			fmt.Println("generated preview id:", resolvedPreviewID)
		}

		config, err := validator.LoadAndValidateConfig(workingDir)
		if err != nil {
			return err
		}

		envFilePath := envFile
		if envFilePath == "" {
			envFilePath = filepath.Join(workingDir, ".env")
		}

		envSecrets, err := secrets.ParseEnvFile(envFilePath)
		if err != nil {
			return err
		}

		flagSecrets, err := secrets.ParseKeyValues(secretInputs)
		if err != nil {
			return err
		}

		merged := secrets.Merge(envSecrets, flagSecrets)

		if err := up.HandleUp(resolvedPreviewID, config, merged); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	upCmd.Flags().StringVar(&previewID, "preview-id", "", "Preview ID to deploy (defaults to generated value)")
	upCmd.Flags().StringArrayVar(&secretInputs, "secret", nil, "Secret in KEY=VALUE format (optional, repeatable)")
	upCmd.Flags().StringVar(&envFile, "env-file", "", "Path to .env file (defaults to .env in working directory)")
}
