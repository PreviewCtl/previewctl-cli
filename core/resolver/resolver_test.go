package resolver

import (
	"strings"
	"testing"

	"github.com/previewctl/previewctl-core/types"
)

func baseConfig() types.PreviewConfig {
	return types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Port:  5432,
				Env: map[string]string{
					"POSTGRES_DB":       "mydb",
					"POSTGRES_USER":     "postgres",
					"POSTGRES_PASSWORD": "${Generate(16)}",
				},
			},
			"api": {
				Build:     &types.BuildConfig{Type: "dockerfile", Context: "."},
				Port:      8080,
				DependsOn: []string{"db"},
				Env: map[string]string{
					"DB_HOST": "${services.db.host}",
					"DB_PORT": "${services.db.port}",
					"DB_USER": "${services.db.env.POSTGRES_USER}",
				},
			},
		},
	}
}

func TestResolveConfig_ServiceHost(t *testing.T) {
	resolved, err := ResolveConfig(baseConfig(), "test-preview", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["api"].Env["DB_HOST"] != "db" {
		t.Errorf("DB_HOST = %q, want %q", resolved.Services["api"].Env["DB_HOST"], "db")
	}
}

func TestResolveConfig_ServicePort(t *testing.T) {
	resolved, err := ResolveConfig(baseConfig(), "test-preview", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["api"].Env["DB_PORT"] != "5432" {
		t.Errorf("DB_PORT = %q, want %q", resolved.Services["api"].Env["DB_PORT"], "5432")
	}
}

func TestResolveConfig_ServiceEnvRef(t *testing.T) {
	resolved, err := ResolveConfig(baseConfig(), "test-preview", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["api"].Env["DB_USER"] != "postgres" {
		t.Errorf("DB_USER = %q, want %q", resolved.Services["api"].Env["DB_USER"], "postgres")
	}
}

func TestResolveConfig_Generate(t *testing.T) {
	resolved, err := ResolveConfig(baseConfig(), "test-preview", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pw := resolved.Services["db"].Env["POSTGRES_PASSWORD"]
	if len(pw) != 16 {
		t.Errorf("POSTGRES_PASSWORD length = %d, want 16", len(pw))
	}
}

func TestResolveConfig_Secrets(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Env: map[string]string{
					"API_KEY": "${secrets.MY_KEY}",
				},
			},
		},
	}

	secrets := map[string]string{"MY_KEY": "secret123"}
	resolved, err := ResolveConfig(config, "p1", secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["api"].Env["API_KEY"] != "secret123" {
		t.Errorf("API_KEY = %q, want %q", resolved.Services["api"].Env["API_KEY"], "secret123")
	}
}

func TestResolveConfig_PreviewID(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"app": {
				Image: "app:latest",
				Env: map[string]string{
					"PREVIEW_ID": "${preview.id}",
				},
			},
		},
	}

	resolved, err := ResolveConfig(config, "my-preview-123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["app"].Env["PREVIEW_ID"] != "my-preview-123" {
		t.Errorf("PREVIEW_ID = %q, want %q", resolved.Services["app"].Env["PREVIEW_ID"], "my-preview-123")
	}
}

func TestResolveConfig_SelfEnvRef(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Port:  5432,
				Env: map[string]string{
					"USER":     "admin",
					"PASSWORD": "pass",
					"URL":      "postgres://${USER}:${PASSWORD}@localhost",
				},
			},
		},
	}

	resolved, err := ResolveConfig(config, "p1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["db"].Env["URL"] != "postgres://admin:pass@localhost" {
		t.Errorf("URL = %q, want %q", resolved.Services["db"].Env["URL"], "postgres://admin:pass@localhost")
	}
}

func TestResolveConfig_MissingSecret(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Env: map[string]string{
					"KEY": "${secrets.MISSING}",
				},
			},
		},
	}

	_, err := ResolveConfig(config, "p1", nil)
	if err == nil {
		t.Error("expected error for missing secret, got nil")
	}
}

func TestResolveConfig_UnknownServiceRef(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Env: map[string]string{
					"HOST": "${services.unknown.host}",
				},
			},
		},
	}

	_, err := ResolveConfig(config, "p1", nil)
	if err == nil {
		t.Error("expected error for unknown service, got nil")
	}
}

func TestResolveConfig_TemplateInString(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Port:  5432,
				Env:   map[string]string{"DB": "mydb"},
			},
			"api": {
				Image:     "api:latest",
				DependsOn: []string{"db"},
				Env: map[string]string{
					"CONN": "host=${services.db.host} port=${services.db.port}",
				},
			},
		},
	}

	resolved, err := ResolveConfig(config, "p1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "host=db port=5432"
	if resolved.Services["api"].Env["CONN"] != want {
		t.Errorf("CONN = %q, want %q", resolved.Services["api"].Env["CONN"], want)
	}
}

func TestResolveConfig_GenerateLength(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"app": {
				Image: "app:latest",
				Env: map[string]string{
					"TOKEN": "${Generate(32)}",
				},
			},
		},
	}

	resolved, err := ResolveConfig(config, "p1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resolved.Services["app"].Env["TOKEN"]) != 32 {
		t.Errorf("TOKEN length = %d, want 32", len(resolved.Services["app"].Env["TOKEN"]))
	}
}

func TestResolveConfig_GenerateInvalidLength(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"app": {
				Image: "app:latest",
				Env: map[string]string{
					"TOKEN": "${Generate(0)}",
				},
			},
		},
	}

	_, err := ResolveConfig(config, "p1", nil)
	if err == nil {
		t.Error("expected error for Generate(0), got nil")
	}
}

func TestResolveConfig_UnknownPreviewProperty(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"app": {
				Image: "app:latest",
				Env: map[string]string{
					"X": "${preview.unknown}",
				},
			},
		},
	}

	_, err := ResolveConfig(config, "p1", nil)
	if err == nil {
		t.Error("expected error for unknown preview property, got nil")
	}
}

func TestResolveConfig_UnknownVariable(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"app": {
				Image: "app:latest",
				Env: map[string]string{
					"X": "${foo.bar.baz}",
				},
			},
		},
	}

	_, err := ResolveConfig(config, "p1", nil)
	if err == nil {
		t.Error("expected error for unknown variable namespace, got nil")
	}
}

func TestResolveConfig_PreservesMetadata(t *testing.T) {
	config := baseConfig()
	resolved, err := ResolveConfig(config, "p1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Version != config.Version {
		t.Errorf("Version = %d, want %d", resolved.Version, config.Version)
	}
	if resolved.Preview.TTL != config.Preview.TTL {
		t.Errorf("TTL = %q, want %q", resolved.Preview.TTL, config.Preview.TTL)
	}
}

func TestResolveConfig_ServiceNoPort(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"worker": {Image: "worker:latest"},
			"api": {
				Image:     "api:latest",
				DependsOn: []string{"worker"},
				Env: map[string]string{
					"PORT": "${services.worker.port}",
				},
			},
		},
	}

	_, err := ResolveConfig(config, "p1", nil)
	if err == nil {
		t.Error("expected error for missing port, got nil")
	}
	if !strings.Contains(err.Error(), "no port") {
		t.Errorf("error should mention 'no port', got: %v", err)
	}
}
