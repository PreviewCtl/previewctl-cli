package database

import (
	"context"

	"github.com/jmoiron/sqlx"
)

const createTablePreviewEnvironments = `
CREATE TABLE IF NOT EXISTS preview_environments (
	 id          TEXT    PRIMARY KEY,
	 name        TEXT    NOT NULL,
	 workspace   TEXT    NOT NULL,
	 branch      TEXT    NOT NULL DEFAULT '',
	 status      TEXT    NOT NULL DEFAULT 'active',
	 created_at  INTEGER NOT NULL,
	 updated_at  INTEGER NOT NULL
);
`

const createIndexWorkspace = `
CREATE INDEX IF NOT EXISTS idx_preview_environments_workspace
ON preview_environments (workspace);
`

const createIndexName = `
CREATE INDEX IF NOT EXISTS idx_preview_environments_name
ON preview_environments (name);
`

const createTablePortMappings = `
CREATE TABLE IF NOT EXISTS port_mappings (
	 id              TEXT    PRIMARY KEY,
	 preview_env_id  TEXT    NOT NULL,
	 service_name    TEXT    NOT NULL,
	 container_port  INTEGER NOT NULL,
	 host_port       INTEGER NOT NULL,
	 created_at      INTEGER NOT NULL,
	 UNIQUE(preview_env_id, service_name)
);
`

const createIndexPortMappingsPreviewEnv = `
CREATE INDEX IF NOT EXISTS idx_port_mappings_preview_env_id
ON port_mappings (preview_env_id);
`

const createTableGeneratedSecrets = `
CREATE TABLE IF NOT EXISTS generated_secrets (
	 id              TEXT    PRIMARY KEY,
	 preview_env_id  TEXT    NOT NULL,
	 service_name    TEXT    NOT NULL,
	 env_key         TEXT    NOT NULL,
	 value           TEXT    NOT NULL,
	 created_at      INTEGER NOT NULL,
	 UNIQUE(preview_env_id, service_name, env_key)
);
`

const createIndexGeneratedSecretsPreviewEnv = `
CREATE INDEX IF NOT EXISTS idx_generated_secrets_preview_env_id
ON generated_secrets (preview_env_id);
`

// Migrate applies the database migrations.
func Migrate(ctx context.Context, db *sqlx.DB) error {
	if _, err := db.ExecContext(ctx, createTablePreviewEnvironments); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, createIndexWorkspace); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, createIndexName); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, createTablePortMappings); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, createIndexPortMappingsPreviewEnv); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, createTableGeneratedSecrets); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, createIndexGeneratedSecretsPreviewEnv); err != nil {
		return err
	}
	return nil
}
