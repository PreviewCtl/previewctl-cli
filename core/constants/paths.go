package constants

import "path/filepath"

func PreviewCtlConfigDirPath(workingDir string) string {
	return filepath.Join(workingDir, PreviewCtlConfigDir)
}

func PreviewCtlConfigFilePath(workingDir string) string {
	return filepath.Join(PreviewCtlConfigDirPath(workingDir), PreviewCtlConfigFile)
}
