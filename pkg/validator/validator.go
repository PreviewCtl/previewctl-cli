package validator

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/previewctl/previewctl-cli/pkg/constants"
	"github.com/previewctl/previewctl-cli/pkg/dag"
	"github.com/previewctl/previewctl-cli/pkg/types"
	"gopkg.in/yaml.v3"
)

// LoadAndValidateConfig reads and validates .previewctrl/preview.yml from workingDir.
func LoadAndValidateConfig(workingDir string) (types.PreviewConfig, error) {
	configPath := constants.PreviewCtrlConfigFilePath(workingDir)
	rawConfig, err := os.ReadFile(configPath)
	if err != nil {
		return types.PreviewConfig{}, fmt.Errorf("failed to read config file at %s: %w", configPath, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(rawConfig))
	decoder.KnownFields(true)

	var config types.PreviewConfig
	if err := decoder.Decode(&config); err != nil {
		return types.PreviewConfig{}, fmt.Errorf("invalid yaml schema in %s: %w", configPath, err)
	}

	if err := ValidateConfig(config); err != nil {
		return types.PreviewConfig{}, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

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

	generateExpr := regexp.MustCompile(`\$\{Generate\((\d+)\)\}`)
	templateVar := regexp.MustCompile(`\$\{([^}]+)\}`)
	serviceEnvRef := regexp.MustCompile(`^services\.([^.]+)\.env\.(.+)$`)
	const maxGenerateLength = 100

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
			switch service.Build.Type {
			case types.BuildTypeDockerfile, types.BuildTypeNixpacks, types.BuildTypeRailpack:
				// valid
			default:
				return fmt.Errorf("services.%s.build.type: unsupported build type %q (must be %q, %q, or %q)", serviceName, service.Build.Type, types.BuildTypeDockerfile, types.BuildTypeNixpacks, types.BuildTypeRailpack)
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

		for envKey, envValue := range service.Env {
			for _, match := range generateExpr.FindAllStringSubmatch(envValue, -1) {
				length, err := strconv.Atoi(match[1])
				if err != nil {
					return fmt.Errorf("services.%s.env.%s: invalid Generate length %q", serviceName, envKey, match[1])
				}
				if length < 1 || length > maxGenerateLength {
					return fmt.Errorf("services.%s.env.%s: Generate length must be between 1 and %d, got %d", serviceName, envKey, maxGenerateLength, length)
				}
			}
		}

		// Validate bare env var references and detect circular dependencies.
		envGraph := dag.NewGraph[string]()
		for key := range service.Env {
			envGraph.AddVertex(key)
		}

		depSet := make(map[string]bool, len(service.DependsOn))
		for _, dep := range service.DependsOn {
			depSet[dep] = true
		}

		for envKey, envValue := range service.Env {
			for _, match := range templateVar.FindAllStringSubmatch(envValue, -1) {
				expr := match[1]
				if generateExpr.MatchString("${" + expr + "}") {
					continue
				}

				// Validate ${services.X.env.Y} references.
				if m := serviceEnvRef.FindStringSubmatch(expr); m != nil {
					refService := m[1]
					refEnvKey := m[2]
					if _, exists := config.Services[refService]; !exists {
						return fmt.Errorf("services.%s.env.%s: references undefined service %q", serviceName, envKey, refService)
					}
					if refService != serviceName {
						if !depSet[refService] {
							return fmt.Errorf("services.%s.env.%s: references env of service %q which is not in depends_on", serviceName, envKey, refService)
						}
					}
					refSvc := config.Services[refService]
					if _, exists := refSvc.Env[refEnvKey]; !exists {
						return fmt.Errorf("services.%s.env.%s: service %q has no env var %q", serviceName, envKey, refService, refEnvKey)
					}
					continue
				}

				if strings.Contains(expr, ".") {
					continue
				}
				if _, exists := service.Env[expr]; !exists {
					return fmt.Errorf("services.%s.env.%s: references undefined env var ${%s}", serviceName, envKey, expr)
				}
				envGraph.AddEdge(expr, envKey)
			}
		}
		if _, err := envGraph.TopoSort(); err != nil {
			return fmt.Errorf("services.%s: env vars contain circular references", serviceName)
		}
	}

	if _, err := serviceGraph.TopoSort(); err != nil {
		return fmt.Errorf("services.depends_on contains cyclic dependencies")
	}

	return nil
}
