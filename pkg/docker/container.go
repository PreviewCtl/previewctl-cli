package docker

import (
	"context"
	"fmt"
	"net/netip"
	"strconv"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/previewctl/previewctl-cli/pkg/types"
)

// RunService creates and starts a container for the given service.
// The container is named "{networkName}-{serviceName}" and attached to the
// specified Docker network with serviceName as a network alias for DNS resolution.
func RunService(ctx context.Context, cli *client.Client, networkName, serviceName string, svc types.ServiceConfig) (string, error) {
	containerName := networkName + "-" + serviceName

	// Stop and remove existing container if present (idempotent re-runs)
	if _, err := cli.ContainerInspect(ctx, containerName, client.ContainerInspectOptions{}); err == nil {
		_, _ = cli.ContainerStop(ctx, containerName, client.ContainerStopOptions{})
		_, _ = cli.ContainerRemove(ctx, containerName, client.ContainerRemoveOptions{})
	}

	// Build env slice
	env := make([]string, 0, len(svc.Env))
	for k, v := range svc.Env {
		env = append(env, k+"="+v)
	}

	// Container config
	containerConfig := &container.Config{
		Image: svc.Image,
		Env:   env,
	}

	// Host config with port bindings
	hostConfig := &container.HostConfig{}

	if svc.Port > 0 {
		port := network.MustParsePort(strconv.Itoa(svc.Port) + "/tcp")
		containerConfig.ExposedPorts = network.PortSet{
			port: struct{}{},
		}
		hostConfig.PortBindings = network.PortMap{
			port: []network.PortBinding{
				{HostIP: netip.IPv4Unspecified(), HostPort: strconv.Itoa(svc.Port)},
			},
		}
	}

	// Network config: attach to the preview network with service name as alias
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {
				Aliases: []string{serviceName},
			},
		},
	}

	resp, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:           containerConfig,
		HostConfig:       hostConfig,
		NetworkingConfig: networkConfig,
		Name:             containerName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create container %q: %w", containerName, err)
	}

	if _, err := cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container %q: %w", containerName, err)
	}

	return resp.ID, nil
}
