package resolver

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/previewctl/previewctl-cli/pkg/dag"
	"github.com/previewctl/previewctl-cli/pkg/deployment"
	"github.com/previewctl/previewctl-cli/pkg/types"
)

var (
	templateVar  = regexp.MustCompile(`\$\{([^}]+)\}`)
	generateExpr = regexp.MustCompile(`^Generate\((\d+)\)$`)
)

// ResolveConfig resolves all ${...} template variables in service env values.
// Services are resolved in dependency order so that ${services.<name>.env.<KEY>}
// references can read already-resolved env vars from upstream services.
func ResolveConfig(config types.PreviewConfig, previewID string, secrets map[string]string) (types.PreviewConfig, error) {
	order, err := deployment.ResolveServiceDeploymentOrderFromConfig(config)
	if err != nil {
		return types.PreviewConfig{}, fmt.Errorf("failed to determine service resolution order: %w", err)
	}

	resolved := types.PreviewConfig{
		Version:  config.Version,
		Preview:  config.Preview,
		Services: make(map[string]types.ServiceConfig, len(config.Services)),
	}

	for _, name := range order {
		svc := config.Services[name]
		resolvedEnv, err := resolveServiceEnv(name, svc.Env, config, previewID, secrets, resolved.Services)
		if err != nil {
			return types.PreviewConfig{}, err
		}

		resolvedSvc := svc
		resolvedSvc.Env = resolvedEnv
		resolved.Services[name] = resolvedSvc
	}

	return resolved, nil
}

// resolveServiceEnv resolves env vars for a single service in dependency order.
// Bare ${VARNAME} references are resolved from already-resolved env vars in the same service.
func resolveServiceEnv(serviceName string, env map[string]string, config types.PreviewConfig, previewID string, secrets map[string]string, resolvedServices map[string]types.ServiceConfig) (map[string]string, error) {
	envGraph := dag.NewGraph[string]()
	for key := range env {
		envGraph.AddVertex(key)
	}

	for key, value := range env {
		for _, match := range templateVar.FindAllStringSubmatch(value, -1) {
			expr := match[1]
			if isSelfEnvRef(expr) {
				envGraph.AddEdge(expr, key)
			}
		}
	}

	order, err := envGraph.TopoSort()
	if err != nil {
		return nil, fmt.Errorf("services.%s: env vars contain circular references", serviceName)
	}

	resolvedEnv := make(map[string]string, len(env))
	for _, key := range order {
		val, err := resolveValue(env[key], config, previewID, secrets, resolvedEnv, resolvedServices)
		if err != nil {
			return nil, fmt.Errorf("services.%s.env.%s: %w", serviceName, key, err)
		}
		resolvedEnv[key] = val
	}

	return resolvedEnv, nil
}

// isSelfEnvRef returns true if the expression is a bare env var name
// (no dots, not a Generate() call).
func isSelfEnvRef(expr string) bool {
	if generateExpr.MatchString(expr) {
		return false
	}
	return !strings.Contains(expr, ".")
}

func resolveValue(value string, config types.PreviewConfig, previewID string, secrets map[string]string, resolvedEnv map[string]string, resolvedServices map[string]types.ServiceConfig) (string, error) {
	var resolveErr error

	result := templateVar.ReplaceAllStringFunc(value, func(match string) string {
		if resolveErr != nil {
			return match
		}

		// Extract expression inside ${ ... }
		expr := match[2 : len(match)-1]

		resolved, err := resolveExpression(expr, config, previewID, secrets, resolvedEnv, resolvedServices)
		if err != nil {
			resolveErr = err
			return match
		}

		return resolved
	})

	if resolveErr != nil {
		return "", resolveErr
	}

	return result, nil
}

func resolveExpression(expr string, config types.PreviewConfig, previewID string, secrets map[string]string, resolvedEnv map[string]string, resolvedServices map[string]types.ServiceConfig) (string, error) {
	if m := generateExpr.FindStringSubmatch(expr); m != nil {
		return resolveGenerateExpr(m[1])
	}

	parts := strings.Split(expr, ".")

	switch parts[0] {
	case "services":
		return resolveServiceExpr(parts[1:], config, resolvedServices)
	case "secrets":
		return resolveSecretExpr(parts[1:], secrets)
	case "preview":
		return resolvePreviewExpr(parts[1:], previewID)
	default:
		if !strings.Contains(expr, ".") {
			if val, exists := resolvedEnv[expr]; exists {
				return val, nil
			}
			return "", fmt.Errorf("undefined env var ${%s}", expr)
		}
		return "", fmt.Errorf("unknown variable ${%s}", expr)
	}
}

func resolveServiceExpr(parts []string, config types.PreviewConfig, resolvedServices map[string]types.ServiceConfig) (string, error) {
	if len(parts) < 2 || len(parts) > 3 {
		return "", fmt.Errorf("invalid service reference: expected ${services.<name>.<property>} or ${services.<name>.env.<KEY>}")
	}

	serviceName := parts[0]
	property := parts[1]

	svc, exists := config.Services[serviceName]
	if !exists {
		return "", fmt.Errorf("unknown service %q", serviceName)
	}

	switch property {
	case "host":
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid service reference: ${services.%s.host} takes no sub-property", serviceName)
		}
		return serviceName, nil
	case "port":
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid service reference: ${services.%s.port} takes no sub-property", serviceName)
		}
		if svc.Port == 0 {
			return "", fmt.Errorf("service %q has no port configured", serviceName)
		}
		return strconv.Itoa(svc.Port), nil
	case "env":
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid service env reference: expected ${services.%s.env.<KEY>}", serviceName)
		}
		envKey := parts[2]
		resolved, exists := resolvedServices[serviceName]
		if !exists {
			return "", fmt.Errorf("service %q has not been resolved yet; ensure it is listed in depends_on", serviceName)
		}
		val, exists := resolved.Env[envKey]
		if !exists {
			return "", fmt.Errorf("service %q has no env var %q", serviceName, envKey)
		}
		return val, nil
	default:
		return "", fmt.Errorf("unknown service property %q (supported: host, port, env)", property)
	}
}

func resolveSecretExpr(parts []string, secrets map[string]string) (string, error) {
	if len(parts) != 1 {
		return "", fmt.Errorf("invalid secret reference: expected ${secrets.<KEY>}")
	}

	key := parts[0]
	value, exists := secrets[key]
	if !exists {
		return "", fmt.Errorf("secret %q not provided", key)
	}

	return value, nil
}

func resolvePreviewExpr(parts []string, previewID string) (string, error) {
	if len(parts) != 1 {
		return "", fmt.Errorf("invalid preview reference: expected ${preview.<property>}")
	}

	switch parts[0] {
	case "id":
		return previewID, nil
	default:
		return "", fmt.Errorf("unknown preview property %q (supported: id)", parts[0])
	}
}

const maxGenerateLength = 100

func resolveGenerateExpr(lengthStr string) (string, error) {
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid Generate length %q: %w", lengthStr, err)
	}
	if length < 1 || length > maxGenerateLength {
		return "", fmt.Errorf("Generate length must be between 1 and %d, got %d", maxGenerateLength, length)
	}
	return randomString(length)
}

func randomString(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}
