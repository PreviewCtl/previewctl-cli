package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	"github.com/previewctl/previewctl-cli/internal/store/database"
)

// DB is the global database handle, available to all subcommands.
var DB *sqlx.DB

var workingDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "previewctl",
	Short: "Spin up ephemeral preview environments with Docker",
	Long: `PreviewCtrl is a CLI tool for creating and managing ephemeral
preview environments locally using Docker.

Define your services, builds, and dependencies in a .previewctrl/preview.yml
config file and bring them up with a single command.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		datasource, err := database.DefaultDatasource()
		if err != nil {
			return fmt.Errorf("failed to resolve database path: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(datasource), 0o700); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}

		db, err := database.ConnectAndMigrate(cmd.Context(), datasource, database.Migrate)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		DB = db

		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		workingDir = workingDirectory
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.previewctl-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
