package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/previewctl/previewctl-cli/internal/store"
)

// PortMappingStore manages port mappings backed by sqlite.
type PortMappingStore struct {
	db *sqlx.DB
}

// NewPortMappingStore returns a new PortMappingStore.
func NewPortMappingStore(db *sqlx.DB) *PortMappingStore {
	return &PortMappingStore{db: db}
}

// FindByPreviewEnv returns all port mappings for a given preview environment ID.
func (s *PortMappingStore) FindByPreviewEnv(ctx context.Context, previewEnvID string) ([]*store.PortMapping, error) {
	query, args, err := Builder.
		Select("id", "preview_env_id", "service_name", "container_port", "host_port", "created_at").
		From("port_mappings").
		Where(squirrel.Eq{"preview_env_id": previewEnvID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var mappings []*store.PortMapping
	if err := s.db.SelectContext(ctx, &mappings, query, args...); err != nil {
		return nil, processSQLErrorf(ctx, err, "failed to find port mappings for preview env %s", previewEnvID)
	}

	return mappings, nil
}

// Upsert inserts or updates a port mapping for a service.
func (s *PortMappingStore) Upsert(ctx context.Context, mapping *store.PortMapping) error {
	now := time.Now().Unix()

	query := `
		INSERT INTO port_mappings (id, preview_env_id, service_name, container_port, host_port, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(preview_env_id, service_name) DO UPDATE SET
			container_port = excluded.container_port,
			host_port = excluded.host_port
	`

	id, err := generateID()
	if err != nil {
		return fmt.Errorf("failed to generate id: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, id, mapping.PreviewEnvID, mapping.ServiceName, mapping.ContainerPort, mapping.HostPort, now); err != nil {
		return processSQLErrorf(ctx, err, "failed to upsert port mapping for service %s", mapping.ServiceName)
	}

	return nil
}
