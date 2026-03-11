package yaml

import (
	"strings"
	"testing"
)

func TestGetDefaultYamlV1(t *testing.T) {
	data, err := GetDefaultYamlV1()
	if err != nil {
		t.Fatalf("GetDefaultYamlV1() error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty default YAML")
	}

	content := string(data)
	if !strings.Contains(content, "version:") {
		t.Error("default YAML should contain 'version:'")
	}
	if !strings.Contains(content, "services:") {
		t.Error("default YAML should contain 'services:'")
	}
}
