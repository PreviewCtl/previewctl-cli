package validator

import (
	"strings"
	"testing"

	"github.com/previewctl/previewctl-core/types"
)

func validConfig() types.PreviewConfig {
	return types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "24h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Build: &types.BuildConfig{Type: "dockerfile", Context: "."},
				Port:  8080,
				Env:   map[string]string{"PORT": "8080"},
			},
		},
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	if err := ValidateConfig(validConfig()); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidateConfig_VersionZero(t *testing.T) {
	c := validConfig()
	c.Version = 0
	assertValidationError(t, c, "version must be a positive integer")
}

func TestValidateConfig_VersionNegative(t *testing.T) {
	c := validConfig()
	c.Version = -1
	assertValidationError(t, c, "version must be a positive integer")
}

func TestValidateConfig_EmptyTTL(t *testing.T) {
	c := validConfig()
	c.Preview.TTL = ""
	assertValidationError(t, c, "preview.ttl is required")
}

func TestValidateConfig_WhitespaceTTL(t *testing.T) {
	c := validConfig()
	c.Preview.TTL = "   "
	assertValidationError(t, c, "preview.ttl is required")
}

func TestValidateConfig_NoServices(t *testing.T) {
	c := validConfig()
	c.Services = map[string]types.ServiceConfig{}
	assertValidationError(t, c, "services must define at least one service")
}

func TestValidateConfig_NoBuildOrImage(t *testing.T) {
	c := validConfig()
	c.Services = map[string]types.ServiceConfig{
		"api": {Port: 8080},
	}
	assertValidationError(t, c, "must define either build or image")
}

func TestValidateConfig_BuildMissingType(t *testing.T) {
	c := validConfig()
	c.Services = map[string]types.ServiceConfig{
		"api": {Build: &types.BuildConfig{Context: "."}},
	}
	assertValidationError(t, c, "build.type is required")
}

func TestValidateConfig_BuildInvalidType(t *testing.T) {
	c := validConfig()
	c.Services = map[string]types.ServiceConfig{
		"api": {Build: &types.BuildConfig{Type: "invalid", Context: "."}},
	}
	assertValidationError(t, c, "unsupported build type")
}

func TestValidateConfig_BuildMissingContext(t *testing.T) {
	c := validConfig()
	c.Services = map[string]types.ServiceConfig{
		"api": {Build: &types.BuildConfig{Type: "dockerfile"}},
	}
	assertValidationError(t, c, "build.context is required")
}

func TestValidateConfig_AllBuildTypes(t *testing.T) {
	for _, bt := range []string{types.BuildTypeDockerfile, types.BuildTypeNixpacks, types.BuildTypeRailpack} {
		t.Run(bt, func(t *testing.T) {
			c := validConfig()
			c.Services = map[string]types.ServiceConfig{
				"svc": {Build: &types.BuildConfig{Type: bt, Context: "."}},
			}
			if err := ValidateConfig(c); err != nil {
				t.Errorf("build type %q should be valid, got error: %v", bt, err)
			}
		})
	}
}

func TestValidateConfig_ImageOnly(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {Image: "postgres:16", Port: 5432},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("image-only service should be valid, got: %v", err)
	}
}

func TestValidateConfig_DependsOnUndefined(t *testing.T) {
	c := validConfig()
	c.Services["api"] = types.ServiceConfig{
		Build:     &types.BuildConfig{Type: "dockerfile", Context: "."},
		DependsOn: []string{"nonexistent"},
	}
	assertValidationError(t, c, "references undefined service")
}

func TestValidateConfig_DependsOnCycle(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"a": {Image: "a:latest", DependsOn: []string{"b"}},
			"b": {Image: "b:latest", DependsOn: []string{"a"}},
		},
	}
	assertValidationError(t, c, "cyclic dependencies")
}

func TestValidateConfig_GenerateValid(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env:   map[string]string{"PW": "${Generate(16)}"},
			},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("Generate(16) should be valid, got: %v", err)
	}
}

func TestValidateConfig_GenerateZero(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env:   map[string]string{"PW": "${Generate(0)}"},
			},
		},
	}
	assertValidationError(t, c, "Generate length must be between 1 and")
}

func TestValidateConfig_GenerateTooLarge(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env:   map[string]string{"PW": "${Generate(999)}"},
			},
		},
	}
	assertValidationError(t, c, "Generate length must be between 1 and")
}

func TestValidateConfig_EnvCircularRef(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"app": {
				Image: "app:latest",
				Env: map[string]string{
					"A": "${B}",
					"B": "${A}",
				},
			},
		},
	}
	assertValidationError(t, c, "circular references")
}

func TestValidateConfig_EnvUndefinedSelfRef(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"app": {
				Image: "app:latest",
				Env: map[string]string{
					"A": "${NOTDEFINED}",
				},
			},
		},
	}
	assertValidationError(t, c, "references undefined env var")
}

func TestValidateConfig_ServiceEnvRefNoDependsOn(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db":  {Image: "postgres:16", Env: map[string]string{"USER": "admin"}},
			"api": {Image: "api:latest", Env: map[string]string{"U": "${services.db.env.USER}"}},
		},
	}
	assertValidationError(t, c, "not in depends_on")
}

func TestValidateConfig_ServiceEnvRefUndefinedKey(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {Image: "postgres:16", Env: map[string]string{"USER": "admin"}},
			"api": {
				Image:     "api:latest",
				DependsOn: []string{"db"},
				Env:       map[string]string{"U": "${services.db.env.MISSING}"},
			},
		},
	}
	assertValidationError(t, c, "has no env var")
}

func TestValidateConfig_ServiceEnvRefToUndefinedService(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Env:   map[string]string{"U": "${services.ghost.env.KEY}"},
			},
		},
	}
	assertValidationError(t, c, "references undefined service")
}

func TestValidateConfig_EnvSelfRefValid(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
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
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("self-referencing env should be valid, got: %v", err)
	}
}

func TestValidateConfig_MultipleDeps(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db":    {Image: "postgres:16", Env: map[string]string{"USER": "pg"}},
			"redis": {Image: "redis:7", Port: 6379},
			"api": {
				Image:     "api:latest",
				DependsOn: []string{"db", "redis"},
				Env: map[string]string{
					"DB_USER":    "${services.db.env.USER}",
					"REDIS_HOST": "${services.redis.host}",
				},
			},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
}

// --- Seed validation tests ---

func TestValidateConfig_SeedPrestartValid(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Seed: &types.SeedConfig{
					Prestart: []types.SeedEntry{
						{Source: "db/data.sqlite", Destination: "/app/data/app.db"},
					},
				},
			},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("valid prestart seed should pass, got: %v", err)
	}
}

func TestValidateConfig_SeedPoststartValid(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env:   map[string]string{"POSTGRES_USER": "postgres"},
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "db/seed.sql", Destination: "/tmp/seed.sql", Cmd: "psql -f /tmp/seed.sql"},
					},
				},
			},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("valid poststart seed should pass, got: %v", err)
	}
}

func TestValidateConfig_SeedPoststartNoCmdValid(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "config/runtime.json", Destination: "/app/config.json"},
					},
				},
			},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("poststart without cmd should pass, got: %v", err)
	}
}

func TestValidateConfig_SeedPrestartMissingSource(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Seed: &types.SeedConfig{
					Prestart: []types.SeedEntry{
						{Destination: "/app/data/app.db"},
					},
				},
			},
		},
	}
	assertValidationError(t, c, "source is required")
}

func TestValidateConfig_SeedPrestartMissingDestination(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Seed: &types.SeedConfig{
					Prestart: []types.SeedEntry{
						{Source: "db/data.sqlite"},
					},
				},
			},
		},
	}
	assertValidationError(t, c, "destination is required")
}

func TestValidateConfig_SeedPrestartWithCmdFails(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Seed: &types.SeedConfig{
					Prestart: []types.SeedEntry{
						{Source: "db/seed.sql", Destination: "/tmp/seed.sql", Cmd: "psql -f /tmp/seed.sql"},
					},
				},
			},
		},
	}
	assertValidationError(t, c, "prestart seeds cannot have cmd")
}

func TestValidateConfig_SeedPoststartMissingSource(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Destination: "/tmp/seed.sql", Cmd: "psql -f /tmp/seed.sql"},
					},
				},
			},
		},
	}
	assertValidationError(t, c, "source is required")
}

func TestValidateConfig_SeedPoststartMissingDestination(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "db/seed.sql", Cmd: "psql -f /tmp/seed.sql"},
					},
				},
			},
		},
	}
	assertValidationError(t, c, "destination is required")
}

func TestValidateConfig_SeedPoststartCmdCrossServiceWithoutDependsOn(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db":  {Image: "postgres:16", Env: map[string]string{"POSTGRES_USER": "pg"}},
			"api": {
				Image: "api:latest",
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "s.sh", Destination: "/s.sh", Cmd: "${services.db.env.POSTGRES_USER}"},
					},
				},
			},
		},
	}
	assertValidationError(t, c, "not in depends_on")
}

func TestValidateConfig_SeedPoststartCmdCrossServiceValid(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {Image: "postgres:16", Env: map[string]string{"POSTGRES_USER": "pg"}},
			"api": {
				Image:     "api:latest",
				DependsOn: []string{"db"},
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "s.sh", Destination: "/s.sh", Cmd: "/s.sh --user ${services.db.env.POSTGRES_USER}"},
					},
				},
			},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("poststart cmd with valid cross-service ref should pass, got: %v", err)
	}
}

func TestValidateConfig_SeedPoststartCmdUndefinedBareVar(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env:   map[string]string{"POSTGRES_USER": "pg"},
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "s.sh", Destination: "/s.sh", Cmd: "${UNDEFINED_VAR}"},
					},
				},
			},
		},
	}
	assertValidationError(t, c, "references undefined env var")
}

func TestValidateConfig_SeedPoststartCmdBareVarValid(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"db": {
				Image: "postgres:16",
				Env:   map[string]string{"POSTGRES_USER": "pg", "POSTGRES_DB": "mydb"},
				Seed: &types.SeedConfig{
					Poststart: []types.SeedEntry{
						{Source: "s.sql", Destination: "/s.sql", Cmd: "psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /s.sql"},
					},
				},
			},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("poststart cmd with valid bare env ref should pass, got: %v", err)
	}
}

func TestValidateConfig_SeedBothPhasesValid(t *testing.T) {
	c := types.PreviewConfig{
		Version: 1,
		Preview: types.PreviewSettings{TTL: "1h"},
		Services: map[string]types.ServiceConfig{
			"api": {
				Image: "api:latest",
				Env:   map[string]string{"DB_NAME": "mydb"},
				Seed: &types.SeedConfig{
					Prestart: []types.SeedEntry{
						{Source: "fixtures/", Destination: "/app/fixtures/"},
					},
					Poststart: []types.SeedEntry{
						{Source: "db/seed.sql", Destination: "/tmp/seed.sql", Cmd: "psql -d ${DB_NAME} -f /tmp/seed.sql"},
					},
				},
			},
		},
	}
	if err := ValidateConfig(c); err != nil {
		t.Fatalf("both phases valid should pass, got: %v", err)
	}
}

func assertValidationError(t *testing.T, config types.PreviewConfig, contains string) {
	t.Helper()
	err := ValidateConfig(config)
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Errorf("expected error containing %q, got: %v", contains, err)
	}
}
