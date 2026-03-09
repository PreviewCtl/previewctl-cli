package docker

import (
	"context"
	"fmt"
	"net/netip"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/previewctl/previewctl-cli/pkg/constants"
	"github.com/previewctl/previewctl-cli/pkg/types"
)

// RunService creates and starts a container for the given service.
// The container is named "{networkName}-{serviceName}" and attached to the
// specified Docker network with serviceName as a network alias for DNS resolution.
// It returns the container ID and the dynamically assigned host port (0 if no port exposed).
func RunService(ctx context.Context, cli *client.Client, networkName, serviceName string, svc types.ServiceConfig, workingDir string) (string, int, error) {
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

	// Host config with port bindings and volumes
	binds, err := resolveVolumePaths(svc.Volumes, serviceName, workingDir)
	if err != nil {
		return "", 0, fmt.Errorf("failed to prepare volumes: %w", err)
	}
	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	if svc.Port > 0 {
		port := network.MustParsePort(strconv.Itoa(svc.Port) + "/tcp")
		containerConfig.ExposedPorts = network.PortSet{
			port: struct{}{},
		}
		hostConfig.PortBindings = network.PortMap{
			port: []network.PortBinding{
				{HostIP: netip.IPv4Unspecified(), HostPort: "0"},
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
		return "", 0, fmt.Errorf("failed to create container %q: %w", containerName, err)
	}

	if _, err := cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", 0, fmt.Errorf("failed to start container %q: %w", containerName, err)
	}

	// Discover the dynamically assigned host port
	var hostPort int
	if svc.Port > 0 {
		inspect, err := cli.ContainerInspect(ctx, resp.ID, client.ContainerInspectOptions{})
		if err != nil {
			return "", 0, fmt.Errorf("failed to inspect container %q: %w", containerName, err)
		}
		portKey := network.MustParsePort(strconv.Itoa(svc.Port) + "/tcp")
		if bindings, ok := inspect.Container.NetworkSettings.Ports[portKey]; ok && len(bindings) > 0 {
			hostPort, _ = strconv.Atoi(bindings[0].HostPort)
		}
	}

	return resp.ID, hostPort, nil
}

// resolveVolumePaths maps container paths to host paths under .previewctrl/data/{serviceName}/.
// Each volume entry is a container path (e.g. "/var/lib/postgresql/data").
// The host path is derived as {workingDir}/.previewctrl/data/{serviceName}/{sanitized-container-path}.
func resolveVolumePaths(volumes []string, serviceName, workingDir string) ([]string, error) {
	binds := make([]string, 0, len(volumes))
	for _, containerPath := range volumes {
		sanitized := strings.ReplaceAll(strings.Trim(containerPath, "/"), "/", "_")
		hostPath := filepath.Join(constants.PreviewCtrlConfigDirPath(workingDir), "data", serviceName, sanitized)
		// if err := os.MkdirAll(hostPath, 0o777); err != nil {
		// 	return nil, fmt.Errorf("failed to create volume directory %q: %w", hostPath, err)
		// }
		binds = append(binds, hostPath+":"+containerPath)
	}
	return binds, nil
}
