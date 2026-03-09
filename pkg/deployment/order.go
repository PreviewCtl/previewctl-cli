package deployment

import (
	"fmt"

	"github.com/previewctl/previewctl-cli/pkg/dag"
	"github.com/previewctl/previewctl-cli/pkg/types"
)

// ResolveServiceDeploymentOrderFromConfig returns service deployment order from a validated config.
func ResolveServiceDeploymentOrderFromConfig(config types.PreviewConfig) ([]string, error) {
	serviceGraph := dag.NewGraph[string]()

	for serviceName, service := range config.Services {
		serviceGraph.AddVertex(serviceName)

		for _, dep := range service.DependsOn {
			// Dependency existence and cycle safety are expected to be pre-validated.
			serviceGraph.AddEdge(dep, serviceName)
		}
	}

	order, err := serviceGraph.TopoSort()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve deployment order: %w", err)
	}

	return order, nil
}
