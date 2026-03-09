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
	return nil
}
