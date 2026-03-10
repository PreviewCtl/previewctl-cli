package database

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config specifies the config for the database package.
type Config struct {
	Datasource string
}

// DefaultDatasource returns the default database path under ~/.previewctl/data/previewctl.db.
func DefaultDatasource() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	fmt.Printf("Database location: %s\n", home)
	return filepath.Join(home, ".previewctl", "data", "previewctl.db"), nil
}
