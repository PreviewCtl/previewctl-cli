package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/previewctl/previewctl-cli/internal/store"
)

var _ store.PreviewEnvironmentStore = (*PreviewEnvironmentStore)(nil)

// PreviewEnvironmentStore implements store.PreviewEnvironmentStore backed by sqlite.
type PreviewEnvironmentStore struct {
	db *sqlx.DB
}

// NewPreviewEnvironmentStore returns a new PreviewEnvironmentStore.
func NewPreviewEnvironmentStore(db *sqlx.DB) *PreviewEnvironmentStore {
	return &PreviewEnvironmentStore{db: db}
}

// Create inserts a new preview environment. A random hex ID is generated automatically.
func (s *PreviewEnvironmentStore) Create(ctx context.Context, env *store.PreviewEnvironment) (*store.PreviewEnvironment, error) {
	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}
	env.ID = id

	now := time.Now().Unix()
	env.CreatedAt = now
	env.UpdatedAt = now

	query, args, err := Builder.
		Insert("preview_environments").
		Columns("id", "name", "workspace", "branch", "status", "created_at", "updated_at").
		Values(env.ID, env.Name, env.Workspace, env.Branch, env.Status, env.CreatedAt, env.UpdatedAt).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to create preview environment")
	}

	return env, nil
}

// generateID returns a 16-character random hex string (like a short Docker container ID).
func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Find returns a preview environment by ID.
func (s *PreviewEnvironmentStore) Find(ctx context.Context, id string) (*store.PreviewEnvironment, error) {
	query, args, err := Builder.
		Select("id", "name", "workspace", "branch", "status", "created_at", "updated_at").
		From("preview_environments").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	env := &store.PreviewEnvironment{}
	if err := s.db.GetContext(ctx, env, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to find preview environment by id %s", id)
	}

	return env, nil
}

// FindByWorkspace returns a preview environment by workspace path.
func (s *PreviewEnvironmentStore) FindByWorkspace(ctx context.Context, workspace string) (*store.PreviewEnvironment, error) {
	query, args, err := Builder.
		Select("id", "name", "workspace", "branch", "status", "created_at", "updated_at").
		From("preview_environments").
		Where(squirrel.Eq{"workspace": workspace}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	env := &store.PreviewEnvironment{}
	if err := s.db.GetContext(ctx, env, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to find preview environment by workspace %s", workspace)
	}

	return env, nil
}

// FindByWorkspaceAndBranch returns a preview environment matching the workspace and branch.
func (s *PreviewEnvironmentStore) FindByName(ctx context.Context, name string) (*store.PreviewEnvironment, error) {
	query, args, err := Builder.
		Select("id", "name", "workspace", "branch", "status", "created_at", "updated_at").
		From("preview_environments").
		Where(squirrel.Eq{"name": name}).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	env := &store.PreviewEnvironment{}
	if err := s.db.GetContext(ctx, env, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to find preview environment by name %s", name)
	}

	return env, nil
}

// FindByWorkspaceAndBranch returns a preview environment matching the workspace and branch.
func (s *PreviewEnvironmentStore) FindByWorkspaceAndBranch(ctx context.Context, workspace string, branch string) (*store.PreviewEnvironment, error) {
	query, args, err := Builder.
		Select("id", "name", "workspace", "branch", "status", "created_at", "updated_at").
		From("preview_environments").
		Where(squirrel.Eq{"workspace": workspace, "branch": branch}).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	env := &store.PreviewEnvironment{}
	if err := s.db.GetContext(ctx, env, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to find preview environment by workspace %s and branch %s", workspace, branch)
	}

	return env, nil
}

// List returns all preview environments.
func (s *PreviewEnvironmentStore) List(ctx context.Context) ([]*store.PreviewEnvironment, error) {
	query, args, err := Builder.
		Select("id", "name", "workspace", "branch", "status", "created_at", "updated_at").
		From("preview_environments").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var envs []*store.PreviewEnvironment
	if err := s.db.SelectContext(ctx, &envs, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to list preview environments")
	}

	return envs, nil
}

// ListByWorkspace returns all preview environments for a given workspace.
func (s *PreviewEnvironmentStore) ListByWorkspace(ctx context.Context, workspace string) ([]*store.PreviewEnvironment, error) {
	query, args, err := Builder.
		Select("id", "name", "workspace", "branch", "status", "created_at", "updated_at").
		From("preview_environments").
		Where(squirrel.Eq{"workspace": workspace}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var envs []*store.PreviewEnvironment
	if err := s.db.SelectContext(ctx, &envs, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to list preview environments for workspace %s", workspace)
	}

	return envs, nil
}

// Update updates an existing preview environment.
func (s *PreviewEnvironmentStore) Update(ctx context.Context, env *store.PreviewEnvironment) error {
	env.UpdatedAt = time.Now().Unix()

	query, args, err := Builder.
		Update("preview_environments").
		Set("name", env.Name).
		Set("workspace", env.Workspace).
		Set("branch", env.Branch).
		Set("status", env.Status).
		Set("updated_at", env.UpdatedAt).
		Where(squirrel.Eq{"id": env.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return processSQLErrorf(ctx, err, "failed to update preview environment %s", env.ID)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if count == 0 {
		return store.ErrResourceNotFound
	}

	return nil
}

// UpdateStatus updates only the status of a preview environment.
func (s *PreviewEnvironmentStore) UpdateStatus(ctx context.Context, id string, status string) error {
	now := time.Now().Unix()

	query, args, err := Builder.
		Update("preview_environments").
		Set("status", status).
		Set("updated_at", now).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update status query: %w", err)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return processSQLErrorf(ctx, err, "failed to update status for preview environment %s", id)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if count == 0 {
		return store.ErrResourceNotFound
	}

	return nil
}

// Delete removes a preview environment by ID.
func (s *PreviewEnvironmentStore) Delete(ctx context.Context, id string) error {
	query, args, err := Builder.
		Delete("preview_environments").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return processSQLErrorf(ctx, err, "failed to delete preview environment %s", id)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if count == 0 {
		return store.ErrResourceNotFound
	}

	return nil
}

// processSQLErrorf translates known SQL errors to store-level errors.
func processSQLErrorf(_ context.Context, err error, format string, args ...interface{}) error {
	translatedError := err

	switch {
	case isErrNoRows(err):
		translatedError = store.ErrResourceNotFound
	case isSQLUniqueConstraintError(err):
		translatedError = store.ErrDuplicate
	default:
		translatedError = fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), translatedError)
	}

	return translatedError
}

func isErrNoRows(err error) bool {
	return err == sql.ErrNoRows
}
