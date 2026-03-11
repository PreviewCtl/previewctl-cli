package constants

import (
	"path/filepath"
	"testing"
)

func TestPreviewCtrlConfigDirPath(t *testing.T) {
	got := PreviewCtrlConfigDirPath("/home/user/project")
	want := filepath.Join("/home/user/project", PreviewCtrlConfigDir)
	if got != want {
		t.Errorf("PreviewCtrlConfigDirPath() = %q, want %q", got, want)
	}
}

func TestPreviewCtrlConfigFilePath(t *testing.T) {
	got := PreviewCtrlConfigFilePath("/home/user/project")
	want := filepath.Join("/home/user/project", PreviewCtrlConfigDir, PreviewCtrlConfigFile)
	if got != want {
		t.Errorf("PreviewCtrlConfigFilePath() = %q, want %q", got, want)
	}
}
