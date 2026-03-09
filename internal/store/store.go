package store

import (
	"context"
	"errors"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrDuplicate        = errors.New("duplicate resource")
)

// PreviewEnvironmentStore defines the interface for managing preview environments.
type PreviewEnvironmentStore interface {
	Create(ctx context.Context, env *PreviewEnvironment) (*PreviewEnvironment, error)
	Find(ctx context.Context, id string) (*PreviewEnvironment, error)
	FindByName(ctx context.Context, name string) (*PreviewEnvironment, error)
	FindByWorkspace(ctx context.Context, workspace string) (*PreviewEnvironment, error)
	FindByWorkspaceAndBranch(ctx context.Context, workspace string, branch string) (*PreviewEnvironment, error)
	List(ctx context.Context) ([]*PreviewEnvironment, error)
	Update(ctx context.Context, env *PreviewEnvironment) error
	UpdateStatus(ctx context.Context, id string, status string) error
	Delete(ctx context.Context, id string) error
}
