package initializer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-errors/errors"
	"github.com/previewctl/previewctl-core/constants"
	"github.com/previewctl/previewctl-core/yaml"
)

func InitRepo(workingDir string) error {
	previewCtlConfigDirPath := constants.PreviewCtlConfigDirPath(workingDir)
	previewCtlConfigFilePath := constants.PreviewCtlConfigFilePath(workingDir)

	err := os.MkdirAll(previewCtlConfigDirPath, os.ModePerm)
	if err != nil {
		return errors.Errorf("failed to initialized PreviewCtl directory: %w", err)
	}

	defaultYaml, err := yaml.GetDefaultYamlV1()
	if err != nil {
		return errors.Errorf("failed to fetch default config file(THIS IS NOT SUPPOSED TO HAPPEN), please contact the dev: %w", err)
	}

	err = os.WriteFile(previewCtlConfigFilePath, defaultYaml, os.ModePerm)
	if err != nil {
		return errors.Errorf("failed to create config: %w", err)
	}

	if err := addDataDirToGitignore(workingDir); err != nil {
		return err
	}

	return nil
}

func addDataDirToGitignore(workingDir string) error {
	gitignorePath := filepath.Join(workingDir, ".gitignore")
	entry := "**/" + constants.PreviewCtlConfigDir + "/data/"

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Errorf("failed to read .gitignore: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Errorf("failed to open .gitignore: %w", err)
	}
	defer f.Close()

	prefix := "\n"
	if len(data) > 0 && data[len(data)-1] == '\n' {
		prefix = ""
	}

	if _, err := f.WriteString(prefix + "\n# PreviewCtl data directories\n" + entry + "\n"); err != nil {
		return errors.Errorf("failed to write to .gitignore: %w", err)
	}

	return nil
}
