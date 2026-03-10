package up

import (
	"context"
	"fmt"

	"github.com/previewctl/previewctl-cli/internal/build/docker"
	"github.com/previewctl/previewctl-cli/internal/build/nixpacks"
	"github.com/previewctl/previewctl-cli/internal/build/railpack"
	"github.com/previewctl/previewctl-cli/internal/store"
	"github.com/previewctl/previewctl-cli/internal/store/database"
	"github.com/previewctl/previewctl-core/deployment"
	"github.com/previewctl/previewctl-core/resolver"
	"github.com/previewctl/previewctl-core/types"
)

func HandleUp(ctx context.Context, previewID string, previewEnvID string, config types.PreviewConfig, secrets map[string]string, userSecrets map[string]string, portStore *database.PortMappingStore, workingDir string) error {
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

	// Load existing port mappings for this preview environment
	savedPorts, err := portStore.FindByPreviewEnv(ctx, previewEnvID)
	if err != nil {
		return fmt.Errorf("failed to load port mappings: %w", err)
	}
	portLookup := make(map[string]int, len(savedPorts))
	for _, pm := range savedPorts {
		portLookup[pm.ServiceName] = pm.HostPort
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

		if svc.Build != nil {
			imageTag := previewID + "-" + serviceName + ":latest"
			switch svc.Build.Type {
			case types.BuildTypeDockerfile:
				if err := docker.BuildImage(ctx, imageTag, *svc.Build, userSecrets, workingDir); err != nil {
					return fmt.Errorf("service %q: %w", serviceName, err)
				}
			case types.BuildTypeNixpacks:
				if err := nixpacks.NixpacksBuild(ctx, imageTag, *svc.Build, userSecrets, workingDir); err != nil {
					return fmt.Errorf("service %q: %w", serviceName, err)
				}
			case types.BuildTypeRailpack:
				if err := railpack.RailpackBuild(ctx, imageTag, *svc.Build, userSecrets, workingDir); err != nil {
					return fmt.Errorf("service %q: %w", serviceName, err)
				}
			default:
				return fmt.Errorf("service %q: unsupported build type %q", serviceName, svc.Build.Type)
			}
			svc.Image = imageTag
		} else {
			if err := docker.EnsureImage(ctx, cli, svc.Image); err != nil {
				return fmt.Errorf("service %q: %w", serviceName, err)
			}
		}

		preferredPort := portLookup[serviceName]

		containerID, hostPort, err := docker.RunService(ctx, cli, previewID, serviceName, svc, userSecrets, preferredPort, workingDir)
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

			// Persist the port mapping
			if err := portStore.Upsert(ctx, &store.PortMapping{
				PreviewEnvID:  previewEnvID,
				ServiceName:   serviceName,
				ContainerPort: svc.Port,
				HostPort:      hostPort,
			}); err != nil {
				return fmt.Errorf("failed to save port mapping for %s: %w", serviceName, err)
			}
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
