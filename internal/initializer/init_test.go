package initializer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/previewctl/previewctl-core/constants"
)

func TestInitRepo_CreatesConfigDirAndFile(t *testing.T) {
	dir := t.TempDir()

	if err := InitRepo(dir); err != nil {
		t.Fatalf("InitRepo() error: %v", err)
	}

	configDir := constants.PreviewCtrlConfigDirPath(dir)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("expected config directory to be created")
	}

	configFile := constants.PreviewCtrlConfigFilePath(dir)
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if len(data) == 0 {
		t.Error("config file should not be empty")
	}
	if !strings.Contains(string(data), "version:") {
		t.Error("config file should contain 'version:'")
	}
}

func TestInitRepo_Idempotent(t *testing.T) {
	dir := t.TempDir()

	if err := InitRepo(dir); err != nil {
		t.Fatalf("first InitRepo() error: %v", err)
	}
	if err := InitRepo(dir); err != nil {
		t.Fatalf("second InitRepo() error: %v", err)
	}
}

func TestInitRepo_AddsDataDirToGitignore(t *testing.T) {
	dir := t.TempDir()

	// Create a .gitignore first
	gitignore := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignore, []byte("node_modules/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := InitRepo(dir); err != nil {
		t.Fatalf("InitRepo() error: %v", err)
	}

	data, err := os.ReadFile(gitignore)
	if err != nil {
		t.Fatal(err)
	}

	entry := constants.PreviewCtrlConfigDir + "/data/"
	if !strings.Contains(string(data), entry) {
		t.Errorf(".gitignore should contain %q, got:\n%s", entry, string(data))
	}
}

func TestInitRepo_GitignoreNotDuplicated(t *testing.T) {
	dir := t.TempDir()

	gitignore := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignore, []byte("node_modules/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run twice
	if err := InitRepo(dir); err != nil {
		t.Fatal(err)
	}
	if err := InitRepo(dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(gitignore)
	if err != nil {
		t.Fatal(err)
	}

	entry := constants.PreviewCtrlConfigDir + "/data/"
	count := strings.Count(string(data), entry)
	if count != 1 {
		t.Errorf("expected entry to appear once, appeared %d times", count)
	}
}

func TestInitRepo_NoGitignore(t *testing.T) {
	dir := t.TempDir()

	// No .gitignore exists — InitRepo should not fail
	if err := InitRepo(dir); err != nil {
		t.Fatalf("InitRepo() error: %v", err)
	}
}
