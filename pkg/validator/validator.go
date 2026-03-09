package validator

import (
	"fmt"
	"strings"

	"github.com/previewctl/previewctl-cli/pkg/dag"
	"github.com/previewctl/previewctl-cli/pkg/types"
)

// ValidateConfig validates semantic rules for preview config.
func ValidateConfig(config types.PreviewConfig) error {
	if config.Version <= 0 {
		return fmt.Errorf("version must be a positive integer")
	}

	if strings.TrimSpace(config.Preview.TTL) == "" {
		return fmt.Errorf("preview.ttl is required")
	}

	if len(config.Services) == 0 {
		return fmt.Errorf("services must define at least one service")
	}

	serviceGraph := dag.NewGraph[string]()

	for serviceName, service := range config.Services {
		serviceGraph.AddVertex(serviceName)

		hasBuild := service.Build != nil
		hasImage := strings.TrimSpace(service.Image) != ""

		if !hasBuild && !hasImage {
			return fmt.Errorf("services.%s must define either build or image", serviceName)
		}

		if hasBuild {
			if strings.TrimSpace(service.Build.Type) == "" {
				return fmt.Errorf("services.%s.build.type is required", serviceName)
			}
			if strings.TrimSpace(service.Build.Context) == "" {
				return fmt.Errorf("services.%s.build.context is required", serviceName)
			}
		}

		for _, dep := range service.DependsOn {
			if _, exists := config.Services[dep]; !exists {
				return fmt.Errorf("services.%s depends_on references undefined service %q", serviceName, dep)
			}

			// If A depends_on B, B must come before A in topological ordering.
			serviceGraph.AddEdge(dep, serviceName)
		}
	}

	if _, err := serviceGraph.TopoSort(); err != nil {
		return fmt.Errorf("services.depends_on contains cyclic dependencies")
	}

	return nil
}
