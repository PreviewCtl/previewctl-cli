package database

import (
	"context"
	"testing"

	"github.com/previewctl/previewctl-cli/internal/store"
)

// setupTestDB creates an in-memory SQLite database with migrations applied.
func setupTestDB(t *testing.T) *PreviewEnvironmentStore {
	t.Helper()
	ctx := context.Background()
	db, err := ConnectAndMigrate(ctx, ":memory:", Migrate)
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewPreviewEnvironmentStore(db)
}

func setupPortMappingDB(t *testing.T) (*PortMappingStore, *PreviewEnvironmentStore) {
	t.Helper()
	ctx := context.Background()
	db, err := ConnectAndMigrate(ctx, ":memory:", Migrate)
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewPortMappingStore(db), NewPreviewEnvironmentStore(db)
}

func TestPreviewEnvironmentStore_CreateAndFind(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	env := &store.PreviewEnvironment{
		Name:      "test-preview",
		Workspace: "/home/user/project",
		Branch:    "main",
		Status:    "active",
	}

	created, err := s.Create(ctx, env)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected non-empty ID")
	}
	if created.CreatedAt == 0 {
		t.Error("expected non-zero CreatedAt")
	}

	found, err := s.Find(ctx, created.ID)
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if found.Name != "test-preview" {
		t.Errorf("Name = %q, want %q", found.Name, "test-preview")
	}
	if found.Workspace != "/home/user/project" {
		t.Errorf("Workspace = %q, want %q", found.Workspace, "/home/user/project")
	}
}

func TestPreviewEnvironmentStore_FindNotFound(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	_, err := s.Find(ctx, "nonexistent")
	if err != store.ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got: %v", err)
	}
}

func TestPreviewEnvironmentStore_FindByWorkspace(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	_, err := s.Create(ctx, &store.PreviewEnvironment{
		Name:      "p1",
		Workspace: "/workspace/a",
		Branch:    "main",
		Status:    "active",
	})
	if err != nil {
		t.Fatal(err)
	}

	found, err := s.FindByWorkspace(ctx, "/workspace/a")
	if err != nil {
		t.Fatalf("FindByWorkspace() error: %v", err)
	}
	if found.Name != "p1" {
		t.Errorf("Name = %q, want %q", found.Name, "p1")
	}
}

func TestPreviewEnvironmentStore_FindByName(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	_, err := s.Create(ctx, &store.PreviewEnvironment{
		Name:      "my-preview",
		Workspace: "/workspace/b",
		Branch:    "dev",
		Status:    "active",
	})
	if err != nil {
		t.Fatal(err)
	}

	found, err := s.FindByName(ctx, "my-preview")
	if err != nil {
		t.Fatalf("FindByName() error: %v", err)
	}
	if found.Branch != "dev" {
		t.Errorf("Branch = %q, want %q", found.Branch, "dev")
	}
}

func TestPreviewEnvironmentStore_FindByWorkspaceAndBranch(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	_, err := s.Create(ctx, &store.PreviewEnvironment{
		Name:      "p-main",
		Workspace: "/ws",
		Branch:    "main",
		Status:    "active",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Create(ctx, &store.PreviewEnvironment{
		Name:      "p-dev",
		Workspace: "/ws",
		Branch:    "dev",
		Status:    "active",
	})
	if err != nil {
		t.Fatal(err)
	}

	found, err := s.FindByWorkspaceAndBranch(ctx, "/ws", "dev")
	if err != nil {
		t.Fatalf("FindByWorkspaceAndBranch() error: %v", err)
	}
	if found.Name != "p-dev" {
		t.Errorf("Name = %q, want %q", found.Name, "p-dev")
	}
}

func TestPreviewEnvironmentStore_List(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	for _, name := range []string{"a", "b", "c"} {
		_, err := s.Create(ctx, &store.PreviewEnvironment{
			Name:      name,
			Workspace: "/ws/" + name,
			Branch:    "main",
			Status:    "active",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	list, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 envs, got %d", len(list))
	}
}

func TestPreviewEnvironmentStore_ListByWorkspace(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	_, _ = s.Create(ctx, &store.PreviewEnvironment{Name: "a", Workspace: "/ws1", Branch: "main", Status: "active"})
	_, _ = s.Create(ctx, &store.PreviewEnvironment{Name: "b", Workspace: "/ws1", Branch: "dev", Status: "active"})
	_, _ = s.Create(ctx, &store.PreviewEnvironment{Name: "c", Workspace: "/ws2", Branch: "main", Status: "active"})

	list, err := s.ListByWorkspace(ctx, "/ws1")
	if err != nil {
		t.Fatalf("ListByWorkspace() error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 envs for /ws1, got %d", len(list))
	}
}

func TestPreviewEnvironmentStore_Update(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	created, _ := s.Create(ctx, &store.PreviewEnvironment{
		Name:      "orig",
		Workspace: "/ws",
		Branch:    "main",
		Status:    "active",
	})

	created.Name = "updated"
	created.Status = "stopped"
	if err := s.Update(ctx, created); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	found, _ := s.Find(ctx, created.ID)
	if found.Name != "updated" {
		t.Errorf("Name = %q, want %q", found.Name, "updated")
	}
	if found.Status != "stopped" {
		t.Errorf("Status = %q, want %q", found.Status, "stopped")
	}
}

func TestPreviewEnvironmentStore_UpdateNotFound(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	err := s.Update(ctx, &store.PreviewEnvironment{ID: "nonexistent", Name: "x", Workspace: "/ws", Status: "active"})
	if err != store.ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got: %v", err)
	}
}

func TestPreviewEnvironmentStore_UpdateStatus(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	created, _ := s.Create(ctx, &store.PreviewEnvironment{
		Name:      "p1",
		Workspace: "/ws",
		Branch:    "main",
		Status:    "active",
	})

	if err := s.UpdateStatus(ctx, created.ID, "stopped"); err != nil {
		t.Fatalf("UpdateStatus() error: %v", err)
	}

	found, _ := s.Find(ctx, created.ID)
	if found.Status != "stopped" {
		t.Errorf("Status = %q, want %q", found.Status, "stopped")
	}
}

func TestPreviewEnvironmentStore_UpdateStatusNotFound(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	err := s.UpdateStatus(ctx, "nonexistent", "stopped")
	if err != store.ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got: %v", err)
	}
}

func TestPreviewEnvironmentStore_Delete(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	created, _ := s.Create(ctx, &store.PreviewEnvironment{
		Name:      "to-delete",
		Workspace: "/ws",
		Branch:    "main",
		Status:    "active",
	})

	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := s.Find(ctx, created.ID)
	if err != store.ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound after delete, got: %v", err)
	}
}

func TestPreviewEnvironmentStore_DeleteNotFound(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	err := s.Delete(ctx, "nonexistent")
	if err != store.ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got: %v", err)
	}
}

// --- Port Mapping Tests ---

func TestPortMappingStore_UpsertAndFind(t *testing.T) {
	portStore, envStore := setupPortMappingDB(t)
	ctx := context.Background()

	env, _ := envStore.Create(ctx, &store.PreviewEnvironment{
		Name: "p1", Workspace: "/ws", Branch: "main", Status: "active",
	})

	err := portStore.Upsert(ctx, &store.PortMapping{
		PreviewEnvID:  env.ID,
		ServiceName:   "api",
		ContainerPort: 8080,
		HostPort:      32000,
	})
	if err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	mappings, err := portStore.FindByPreviewEnv(ctx, env.ID)
	if err != nil {
		t.Fatalf("FindByPreviewEnv() error: %v", err)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(mappings))
	}
	if mappings[0].ServiceName != "api" {
		t.Errorf("ServiceName = %q, want %q", mappings[0].ServiceName, "api")
	}
	if mappings[0].HostPort != 32000 {
		t.Errorf("HostPort = %d, want %d", mappings[0].HostPort, 32000)
	}
}

func TestPortMappingStore_UpsertOverwrite(t *testing.T) {
	portStore, envStore := setupPortMappingDB(t)
	ctx := context.Background()

	env, _ := envStore.Create(ctx, &store.PreviewEnvironment{
		Name: "p1", Workspace: "/ws", Branch: "main", Status: "active",
	})

	_ = portStore.Upsert(ctx, &store.PortMapping{
		PreviewEnvID: env.ID, ServiceName: "api", ContainerPort: 8080, HostPort: 32000,
	})
	// Upsert again with different port
	_ = portStore.Upsert(ctx, &store.PortMapping{
		PreviewEnvID: env.ID, ServiceName: "api", ContainerPort: 8080, HostPort: 33000,
	})

	mappings, _ := portStore.FindByPreviewEnv(ctx, env.ID)
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping after upsert, got %d", len(mappings))
	}
	if mappings[0].HostPort != 33000 {
		t.Errorf("HostPort = %d, want %d after upsert", mappings[0].HostPort, 33000)
	}
}

func TestPortMappingStore_DeleteByPreviewEnv(t *testing.T) {
	portStore, envStore := setupPortMappingDB(t)
	ctx := context.Background()

	env, _ := envStore.Create(ctx, &store.PreviewEnvironment{
		Name: "p1", Workspace: "/ws", Branch: "main", Status: "active",
	})

	_ = portStore.Upsert(ctx, &store.PortMapping{
		PreviewEnvID: env.ID, ServiceName: "api", ContainerPort: 8080, HostPort: 32000,
	})
	_ = portStore.Upsert(ctx, &store.PortMapping{
		PreviewEnvID: env.ID, ServiceName: "db", ContainerPort: 5432, HostPort: 32001,
	})

	if err := portStore.DeleteByPreviewEnv(ctx, env.ID); err != nil {
		t.Fatalf("DeleteByPreviewEnv() error: %v", err)
	}

	mappings, _ := portStore.FindByPreviewEnv(ctx, env.ID)
	if len(mappings) != 0 {
		t.Errorf("expected 0 mappings after delete, got %d", len(mappings))
	}
}

func TestPortMappingStore_FindByPreviewEnv_Empty(t *testing.T) {
	portStore, _ := setupPortMappingDB(t)
	ctx := context.Background()

	mappings, err := portStore.FindByPreviewEnv(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("FindByPreviewEnv() error: %v", err)
	}
	if len(mappings) != 0 {
		t.Errorf("expected 0 mappings, got %d", len(mappings))
	}
}

// --- Migration & Connect tests ---

func TestConnectAndMigrate(t *testing.T) {
	ctx := context.Background()
	db, err := ConnectAndMigrate(ctx, ":memory:", Migrate)
	if err != nil {
		t.Fatalf("ConnectAndMigrate() error: %v", err)
	}
	defer db.Close()

	// Verify tables exist by querying them
	var count int
	if err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM preview_environments"); err != nil {
		t.Fatalf("preview_environments table should exist: %v", err)
	}
	if err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM port_mappings"); err != nil {
		t.Fatalf("port_mappings table should exist: %v", err)
	}
}
