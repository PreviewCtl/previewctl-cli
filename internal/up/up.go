package up

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/previewctl/previewctl-cli/internal/build/docker"
	"github.com/previewctl/previewctl-cli/internal/build/nixpacks"
	"github.com/previewctl/previewctl-cli/internal/build/railpack"
	"github.com/previewctl/previewctl-cli/internal/identity"
	"github.com/previewctl/previewctl-cli/internal/store"
	"github.com/previewctl/previewctl-cli/internal/store/database"
	"github.com/previewctl/previewctl-core/deployment"
	"github.com/previewctl/previewctl-core/resolver"
	"github.com/previewctl/previewctl-core/types"
)

func HandleUp(ctx context.Context, previewID string, previewEnvID string, branch string, config types.PreviewConfig, secrets map[string]string, userSecrets map[string]string, portStore *database.PortMappingStore, workingDir string) error {
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
			sanitizedBranch := identity.SanitizeResourceName(branch)
			imageBase := strings.TrimSuffix(previewID, "-"+sanitizedBranch)
			imageTag := imageBase + "-" + serviceName + ":" + sanitizedBranch
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

		containerName := previewID + "-" + serviceName

		// Phase 1: Create container (not started yet)
		containerID, err := docker.CreateService(ctx, cli, previewID, serviceName, svc, userSecrets, preferredPort, workingDir)
		if err != nil {
			return fmt.Errorf("service %q: %w", serviceName, err)
		}

		// Phase 2: Run prestart seeds (filesystem seeds — copy before container starts)
		if svc.Seed != nil && len(svc.Seed.Prestart) > 0 {
			fmt.Printf("  running prestart seeds for %s...\n", serviceName)
			if err := docker.RunPrestartSeeds(ctx, cli, containerID, svc.Seed.Prestart, workingDir); err != nil {
				return fmt.Errorf("service %q prestart seed: %w", serviceName, err)
			}
		}

		// Phase 3: Start the container
		hostPort, err := docker.StartService(ctx, cli, containerID, containerName, svc)
		if err != nil {
			return fmt.Errorf("service %q: %w", serviceName, err)
		}

		fmt.Printf("  started %s (container %s)\n", serviceName, containerID[:12])

		// Phase 4: Run poststart seeds (runtime seeds — after container is healthy)
		if svc.Seed != nil && len(svc.Seed.Poststart) > 0 {
			fmt.Printf("  waiting for %s to be ready...\n", serviceName)
			if err := docker.WaitHealthy(ctx, cli, containerID, 30*time.Second); err != nil {
				return fmt.Errorf("service %q: %w", serviceName, err)
			}
			fmt.Printf("  running poststart seeds for %s...\n", serviceName)
			if err := docker.RunPoststartSeeds(ctx, cli, containerID, svc.Seed.Poststart, workingDir); err != nil {
				return fmt.Errorf("service %q poststart seed: %w", serviceName, err)
			}
		}

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
