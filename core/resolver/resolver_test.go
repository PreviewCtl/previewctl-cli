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
	resolved, _, err := ResolveConfig(baseConfig(), "test-preview", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["api"].Env["DB_HOST"] != "db" {
		t.Errorf("DB_HOST = %q, want %q", resolved.Services["api"].Env["DB_HOST"], "db")
	}
}

func TestResolveConfig_ServicePort(t *testing.T) {
	resolved, _, err := ResolveConfig(baseConfig(), "test-preview", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["api"].Env["DB_PORT"] != "5432" {
		t.Errorf("DB_PORT = %q, want %q", resolved.Services["api"].Env["DB_PORT"], "5432")
	}
}

func TestResolveConfig_ServiceEnvRef(t *testing.T) {
	resolved, _, err := ResolveConfig(baseConfig(), "test-preview", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["api"].Env["DB_USER"] != "postgres" {
		t.Errorf("DB_USER = %q, want %q", resolved.Services["api"].Env["DB_USER"], "postgres")
	}
}

func TestResolveConfig_Generate(t *testing.T) {
	resolved, _, err := ResolveConfig(baseConfig(), "test-preview", nil, nil)
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
	resolved, _, err := ResolveConfig(config, "p1", secrets, nil)
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

	resolved, _, err := ResolveConfig(config, "my-preview-123", nil, nil)
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

	resolved, _, err := ResolveConfig(config, "p1", nil, nil)
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

	_, _, err := ResolveConfig(config, "p1", nil, nil)
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

	_, _, err := ResolveConfig(config, "p1", nil, nil)
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

	resolved, _, err := ResolveConfig(config, "p1", nil, nil)
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

	resolved, _, err := ResolveConfig(config, "p1", nil, nil)
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

	_, _, err := ResolveConfig(config, "p1", nil, nil)
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

	_, _, err := ResolveConfig(config, "p1", nil, nil)
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

	_, _, err := ResolveConfig(config, "p1", nil, nil)
	if err == nil {
		t.Error("expected error for unknown variable namespace, got nil")
	}
}

func TestResolveConfig_PreservesMetadata(t *testing.T) {
	config := baseConfig()
	resolved, _, err := ResolveConfig(config, "p1", nil, nil)
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

	_, _, err := ResolveConfig(config, "p1", nil, nil)
	if err == nil {
		t.Error("expected error for missing port, got nil")
	}
	if !strings.Contains(err.Error(), "no port") {
		t.Errorf("error should mention 'no port', got: %v", err)
	}
}

// --- Seed cmd resolution tests ---

func TestResolveConfig_SeedPoststartCmdBareEnvRef(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Port:  5432,
				Env: map[string]string{
					"POSTGRES_USER": "postgres",
					"POSTGRES_DB":   "mydb",
				},
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "db/seed.sql", Destination: "/tmp/seed.sql", Cmd: "psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /tmp/seed.sql"},
					},
				},
			},
		},
	}

	resolved, _, err := ResolveConfig(config, "p1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "psql -U postgres -d mydb -f /tmp/seed.sql"
	got := resolved.Services["db"].Seed.Poststart[0].Cmd
	if got != want {
		t.Errorf("seed cmd = %q, want %q", got, want)
	}
}

func TestResolveConfig_SeedPoststartCmdCrossServiceEnvRef(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Port:  5432,
				Env: map[string]string{
					"POSTGRES_USER": "admin",
					"POSTGRES_PWD":  "secret",
				},
			},
			"api": {
				Image:     "api:latest",
				DependsOn: []string{"db"},
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "db/seed", Destination: "/seed", Cmd: "/seed/migrate.sh --POSTGRES_USER ${services.db.env.POSTGRES_USER} ${services.db.env.POSTGRES_PWD}"},
					},
				},
			},
		},
	}

	resolved, _, err := ResolveConfig(config, "p1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "/seed/migrate.sh --POSTGRES_USER admin secret"
	got := resolved.Services["api"].Seed.Poststart[0].Cmd
	if got != want {
		t.Errorf("seed cmd = %q, want %q", got, want)
	}
}

func TestResolveConfig_SeedPoststartCmdSecretRef(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "s.sh", Destination: "/s.sh", Cmd: "/s.sh ${secrets.DB_TOKEN}"},
					},
				},
			},
		},
	}

	secrets := map[string]string{"DB_TOKEN": "tok123"}
	resolved, _, err := ResolveConfig(config, "p1", secrets, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "/s.sh tok123"
	got := resolved.Services["db"].Seed.Poststart[0].Cmd
	if got != want {
		t.Errorf("seed cmd = %q, want %q", got, want)
	}
}

func TestResolveConfig_SeedPoststartCmdPreviewID(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"app": {
				Image: "app:latest",
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "s.sh", Destination: "/s.sh", Cmd: "/s.sh --preview ${preview.id}"},
					},
				},
			},
		},
	}

	resolved, _, err := ResolveConfig(config, "my-preview-42", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "/s.sh --preview my-preview-42"
	got := resolved.Services["app"].Seed.Poststart[0].Cmd
	if got != want {
		t.Errorf("seed cmd = %q, want %q", got, want)
	}
}

func TestResolveConfig_SeedPrestartPassedThrough(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Seed: &types.SeedConfig{
					Prestart: []types.SeedEntry{
						{Source: "db/data.sqlite", Destination: "/app/data/app.db"},
						{Source: "fixtures/", Destination: "/app/fixtures/"},
					},
				},
			},
		},
	}

	resolved, _, err := ResolveConfig(config, "p1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seed := resolved.Services["api"].Seed
	if seed == nil {
		t.Fatal("expected seed to be preserved, got nil")
	}
	if len(seed.Prestart) != 2 {
		t.Fatalf("expected 2 prestart entries, got %d", len(seed.Prestart))
	}
	if seed.Prestart[0].Source != "db/data.sqlite" || seed.Prestart[0].Destination != "/app/data/app.db" {
		t.Errorf("prestart[0] = %+v, want source=db/data.sqlite dest=/app/data/app.db", seed.Prestart[0])
	}
}

func TestResolveConfig_SeedNilPassedThrough(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"api": {Image: "api:latest"},
		},
	}

	resolved, _, err := ResolveConfig(config, "p1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["api"].Seed != nil {
		t.Error("expected nil seed, got non-nil")
	}
}

// --- Generated secrets persistence tests ---

func TestResolveConfig_GenerateReturnsGeneratedSecrets(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env: map[string]string{
					"POSTGRES_PASSWORD": "${Generate(16)}",
				},
			},
		},
	}

	_, generated, err := ResolveConfig(config, "p1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := generated["db.POSTGRES_PASSWORD"]
	if !ok {
		t.Fatal("expected generated secret for db.POSTGRES_PASSWORD")
	}
	if len(val) != 16 {
		t.Errorf("generated secret length = %d, want 16", len(val))
	}
}

func TestResolveConfig_GenerateReusesSavedSecrets(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env: map[string]string{
					"POSTGRES_PASSWORD": "${Generate(16)}",
				},
			},
		},
	}

	saved := map[string]string{
		"db.POSTGRES_PASSWORD": "previously_saved!",
	}

	resolved, generated, err := ResolveConfig(config, "p1", nil, saved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["db"].Env["POSTGRES_PASSWORD"] != "previously_saved!" {
		t.Errorf("POSTGRES_PASSWORD = %q, want %q", resolved.Services["db"].Env["POSTGRES_PASSWORD"], "previously_saved!")
	}
	if generated["db.POSTGRES_PASSWORD"] != "previously_saved!" {
		t.Errorf("generated map should contain saved value, got %q", generated["db.POSTGRES_PASSWORD"])
	}
}

func TestResolveConfig_GenerateStableAcrossRuns(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env: map[string]string{
					"PASSWORD": "${Generate(20)}",
				},
			},
		},
	}

	// First run: no saved secrets
	_, generated1, err := ResolveConfig(config, "p1", nil, nil)
	if err != nil {
		t.Fatalf("run 1: unexpected error: %v", err)
	}

	// Second run: pass generated secrets from first run
	resolved2, generated2, err := ResolveConfig(config, "p1", nil, generated1)
	if err != nil {
		t.Fatalf("run 2: unexpected error: %v", err)
	}

	if resolved2.Services["db"].Env["PASSWORD"] != generated1["db.PASSWORD"] {
		t.Errorf("second run should reuse first run's value: got %q, want %q",
			resolved2.Services["db"].Env["PASSWORD"], generated1["db.PASSWORD"])
	}
	if generated2["db.PASSWORD"] != generated1["db.PASSWORD"] {
		t.Errorf("generated secrets should be stable across runs")
	}
}

func TestResolveConfig_GenerateCrossServiceRefUsesSaved(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Port:  5432,
				Env: map[string]string{
					"POSTGRES_PASSWORD": "${Generate(16)}",
				},
			},
			"api": {
				Image:     "api:latest",
				DependsOn: []string{"db"},
				Env: map[string]string{
					"DB_PASS": "${services.db.env.POSTGRES_PASSWORD}",
				},
			},
		},
	}

	saved := map[string]string{
		"db.POSTGRES_PASSWORD": "stable_password!",
	}

	resolved, _, err := ResolveConfig(config, "p1", nil, saved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Services["db"].Env["POSTGRES_PASSWORD"] != "stable_password!" {
		t.Errorf("DB POSTGRES_PASSWORD = %q, want %q", resolved.Services["db"].Env["POSTGRES_PASSWORD"], "stable_password!")
	}
	if resolved.Services["api"].Env["DB_PASS"] != "stable_password!" {
		t.Errorf("API DB_PASS = %q, want %q (should follow saved db secret)", resolved.Services["api"].Env["DB_PASS"], "stable_password!")
	}
}

func TestResolveConfig_NonGenerateEnvNotInGeneratedMap(t *testing.T) {
	config := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Env: map[string]string{
					"STATIC": "hello",
					"SECRET": "${secrets.MY_KEY}",
				},
			},
		},
	}

	secrets := map[string]string{"MY_KEY": "val"}
	_, generated, err := ResolveConfig(config, "p1", secrets, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(generated) != 0 {
		t.Errorf("expected no generated secrets, got %d: %v", len(generated), generated)
	}
}
