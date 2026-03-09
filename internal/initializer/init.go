package initializer

import (
	"os"

	"github.com/go-errors/errors"
	"github.com/previewctl/previewctl-cli/pkg/constants"
	"github.com/previewctl/previewctl-cli/pkg/yaml"
)

func InitRepo() error {
	workingDir, err := os.Getwd()
	if err != nil {
		return errors.Errorf("failed to get working directory: %w", err)
	}

	previewCtrlConfigDirPath := constants.PreviewCtrlConfigDirPath(workingDir)
	previewCtrlConfigFilePath := constants.PreviewCtrlConfigFilePath(workingDir)

	err = os.MkdirAll(previewCtrlConfigDirPath, os.ModePerm)
	if err != nil {
		return errors.Errorf("failed to initialized PreviewCtrl directory: %w", err)
	}

	defaultYaml, err := yaml.GetDefaultYamlV1()
	if err != nil {
		return errors.Errorf("failed to fetch default config file(THIS IS NOT SUPPOSED TO HAPPEN), please contact the dev: %w", err)
	}

	err = os.WriteFile(previewCtrlConfigFilePath, defaultYaml, os.ModePerm)
	if err != nil {
		return errors.Errorf("failed to create config: %w", err)
	}

	return nil

}
