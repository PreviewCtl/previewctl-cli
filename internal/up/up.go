package up

import (
	"context"
	"fmt"

	"github.com/previewctl/previewctl-cli/pkg/deployment"
	"github.com/previewctl/previewctl-cli/pkg/docker"
	"github.com/previewctl/previewctl-cli/pkg/resolver"
	"github.com/previewctl/previewctl-cli/pkg/types"
)

func HandleUp(ctx context.Context, previewID string, config types.PreviewConfig, secrets map[string]string, workingDir string) error {
	resolvedConfig, err := resolver.ResolveConfig(config, previewID, secrets)
	if err != nil {
		return fmt.Errorf("failed to resolve config variables: %w", err)
	}

	deploymentOrder, err := deployment.ResolveServiceDeploymentOrderFromConfig(resolvedConfig)
	if err != nil {
		return err
	}

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

	type portMapping struct {
		service       string
		containerPort int
		hostPort      int
	}
	var portMappings []portMapping

	for _, serviceName := range deploymentOrder {
		svc := resolvedConfig.Services[serviceName]
		fmt.Printf("deploying %s...\n", serviceName)

		if err := docker.EnsureImage(ctx, cli, svc.Image); err != nil {
			return fmt.Errorf("service %q: %w", serviceName, err)
		}

		containerID, hostPort, err := docker.RunService(ctx, cli, previewID, serviceName, svc, workingDir)
		if err != nil {
			return fmt.Errorf("service %q: %w", serviceName, err)
		}

		fmt.Printf("  started %s (container %s)\n", serviceName, containerID[:12])

		if svc.Port > 0 && hostPort > 0 {
			portMappings = append(portMappings, portMapping{
				service:       serviceName,
				containerPort: svc.Port,
				hostPort:      hostPort,
			})
		}
	}

	fmt.Println("\nall services deployed successfully")

	if len(portMappings) > 0 {
		fmt.Println("\nPort mappings:")
		for _, pm := range portMappings {
			fmt.Printf("  %-20s localhost:%d -> :%d/tcp\n", pm.service, pm.hostPort, pm.containerPort)
		}
	}

	return nil
}
