package docker

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveVolumePaths(t *testing.T) {
	workingDir := "/home/user/project"
	if runtime.GOOS == "windows" {
		workingDir = `C:\Users\user\project`
	}

	volumes := []string{
		"/var/lib/postgresql/data",
		"/data",
	}

	binds, err := resolveVolumePaths(volumes, "my-preview", "postgres", workingDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(binds) != 2 {
		t.Fatalf("expected 2 binds, got %d", len(binds))
	}

	// First bind: /var/lib/postgresql/data -> sanitized path
	expectedHost := filepath.Join(workingDir, ".previewctl", "data", "my-preview", "postgres", "var_lib_postgresql_data")
	expectedBind := expectedHost + ":/var/lib/postgresql/data"
	if binds[0] != expectedBind {
		t.Errorf("bind[0] = %q, want %q", binds[0], expectedBind)
	}

	// Second bind: /data
	expectedHost2 := filepath.Join(workingDir, ".previewctl", "data", "my-preview", "postgres", "data")
	expectedBind2 := expectedHost2 + ":/data"
	if binds[1] != expectedBind2 {
		t.Errorf("bind[1] = %q, want %q", binds[1], expectedBind2)
	}
}

func TestResolveVolumePaths_Empty(t *testing.T) {
	binds, err := resolveVolumePaths(nil, "preview-1", "svc", "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(binds) != 0 {
		t.Errorf("expected 0 binds, got %d", len(binds))
	}
}
