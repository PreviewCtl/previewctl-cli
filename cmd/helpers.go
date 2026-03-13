package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/previewctl/previewctl-cli/internal/identity"
	"github.com/previewctl/previewctl-cli/internal/store"
)

func resolveCurrentGitBranch() string {
	branch, err := currentGitBranch(workingDir)
	if err != nil {
		fmt.Printf("failed to determine git branch: %v\n", err)
		fmt.Println("using default branch name: main")
		branch = "main"
	}

	return branch
}

// resovlve Preview env
func resolvePreviewEnv(ctx context.Context, previewID string) (string, error) {
	if strings.TrimSpace(previewID) != "" {
		return identity.ResolvePreviewID(previewID, workingDir, gitBranch)
	}

	previews, err := envStore.FindByWorkspaceAndBranch(ctx, workingDir, gitBranch)
	if err != nil && !errors.Is(store.ErrResourceNotFound, err) {
		return "", err
	}
	if previews != nil {
		return previews.Name, nil
	}
	return identity.ResolvePreviewID(previewID, workingDir, gitBranch)
}

func findEnvByNameOrID(ctx context.Context, nameOrId string) (*store.PreviewEnvironment, error) {
	env, err := envStore.FindByName(ctx, nameOrId)
	if err != nil && !errors.Is(err, store.ErrResourceNotFound) {
		return nil, err
	}

	if env == nil {
		env, err = envStore.Find(ctx, nameOrId)
		if err != nil && !errors.Is(err, store.ErrResourceNotFound) {
			return nil, err
		}
	}

	if env == nil {
		return nil, fmt.Errorf("preview environment %q not found", nameOrId)
	}

	return env, nil
}

func findCurrentPreviewIfOnce(ctx context.Context) (*store.PreviewEnvironment, error) {
	envs, err := envStore.ListByWorkspace(ctx, workingDir)
	if err != nil && !errors.Is(err, store.ErrResourceNotFound) {
		return nil, err
	}

	switch len(envs) {
	case 0:
		return nil, fmt.Errorf("no preview environment found in this workspace")
	case 1:
		return envs[0], nil
	default:
		printEnvTable(envs)
		return nil, fmt.Errorf("Found multiple preview environments in this workspace, please select one with name or id")
	}

}

func printEnvTable(envs []*store.PreviewEnvironment) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tBRANCH\tSTATUS\tWORKSPACE\tCREATED")
	for _, e := range envs {
		created := time.Unix(e.CreatedAt, 0).Format(time.DateTime)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", e.ID, e.Name, e.Branch, e.Status, e.Workspace, created)
	}
	w.Flush()
}
