package constants

import "path/filepath"

func PreviewCtrlConfigDirPath(workingDir string) string {
	return filepath.Join(workingDir, PreviewCtrlConfigDir)
}

func PreviewCtrlConfigFilePath(workingDir string) string {
	return filepath.Join(PreviewCtrlConfigDirPath(workingDir), PreviewCtrlConfigFile)
}
