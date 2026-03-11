package constants

import (
	"path/filepath"
	"testing"
)

func TestPreviewCtlConfigDirPath(t *testing.T) {
	got := PreviewCtlConfigDirPath("/home/user/project")
	want := filepath.Join("/home/user/project", PreviewCtlConfigDir)
	if got != want {
		t.Errorf("PreviewCtlConfigDirPath() = %q, want %q", got, want)
	}
}

func TestPreviewCtlConfigFilePath(t *testing.T) {
	got := PreviewCtlConfigFilePath("/home/user/project")
	want := filepath.Join("/home/user/project", PreviewCtlConfigDir, PreviewCtlConfigFile)
	if got != want {
		t.Errorf("PreviewCtlConfigFilePath() = %q, want %q", got, want)
	}
}
