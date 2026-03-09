package up

import (
	"context"
	"fmt"

	"github.com/previewctl/previewctl-cli/pkg/deployment"
	"github.com/previewctl/previewctl-cli/pkg/docker"
	"github.com/previewctl/previewctl-cli/pkg/resolver"
	"github.com/previewctl/previewctl-cli/pkg/types"
)

func HandleUp(previewID string, config types.PreviewConfig, secrets map[string]string) error {
	resolvedConfig, err := resolver.ResolveConfig(config, previewID, secrets)
	if err != nil {
		return fmt.Errorf("failed to resolve config variables: %w", err)
	}

	deploymentOrder, err := deployment.ResolveServiceDeploymentOrderFromConfig(resolvedConfig)
	if err != nil {
		return err
	}

	ctx := context.Background()

	cli, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	fmt.Printf("creating network %q...\n", previewID)
	_, err = docker.EnsureNetwork(ctx, cli, previewID)
	if err != nil {
		return err
	}

	for _, serviceName := range deploymentOrder {
		svc := resolvedConfig.Services[serviceName]
		fmt.Printf("deploying %s...\n", serviceName)

		if err := docker.EnsureImage(ctx, cli, svc.Image); err != nil {
			return fmt.Errorf("service %q: %w", serviceName, err)
		}

		containerID, err := docker.RunService(ctx, cli, previewID, serviceName, svc)
		if err != nil {
			return fmt.Errorf("service %q: %w", serviceName, err)
		}

		fmt.Printf("  started %s (container %s)\n", serviceName, containerID[:12])
	}

	fmt.Println("all services deployed successfully")
	return nil
}
