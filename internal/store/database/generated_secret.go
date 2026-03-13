package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/previewctl/previewctl-cli/internal/store"
)

// GeneratedSecretStore manages generated secrets backed by sqlite.
type GeneratedSecretStore struct {
	db *sqlx.DB
}

// NewGeneratedSecretStore returns a new GeneratedSecretStore.
func NewGeneratedSecretStore(db *sqlx.DB) *GeneratedSecretStore {
	return &GeneratedSecretStore{db: db}
}

// FindByPreviewEnv returns all generated secrets for a given preview environment ID.
func (s *GeneratedSecretStore) FindByPreviewEnv(ctx context.Context, previewEnvID string) ([]*store.GeneratedSecret, error) {
	query, args, err := Builder.
		Select("id", "preview_env_id", "service_name", "env_key", "value", "created_at").
		From("generated_secrets").
		Where(squirrel.Eq{"preview_env_id": previewEnvID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var secrets []*store.GeneratedSecret
	if err := s.db.SelectContext(ctx, &secrets, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to find generated secrets for preview env %s", previewEnvID)
	}

	return secrets, nil
}

// Upsert inserts or updates a generated secret for a service env key.
func (s *GeneratedSecretStore) Upsert(ctx context.Context, secret *store.GeneratedSecret) error {
	now := time.Now().Unix()

	query := `
		INSERT INTO generated_secrets (id, preview_env_id, service_name, env_key, value, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(preview_env_id, service_name, env_key) DO UPDATE SET
			value = excluded.value
	`

	id, err := generateID()
	if err != nil {
		return fmt.Errorf("failed to generate id: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, id, secret.PreviewEnvID, secret.ServiceName, secret.EnvKey, secret.Value, now); err != nil {
		return processSQLErrorf(ctx, err, "failed to upsert generated secret for service %s key %s", secret.ServiceName, secret.EnvKey)
	}

	return nil
}

// DeleteByPreviewEnv removes all generated secrets for a given preview environment.
func (s *GeneratedSecretStore) DeleteByPreviewEnv(ctx context.Context, previewEnvID string) error {
	query, args, err := Builder.
		Delete("generated_secrets").
		Where(squirrel.Eq{"preview_env_id": previewEnvID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return processSQLErrorf(ctx, err, "failed to delete generated secrets for preview env %s", previewEnvID)
	}

	return nil
}
