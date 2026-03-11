package deployment

import (
	"testing"

	"github.com/previewctl/previewctl-core/types"
)

func TestResolveServiceDeploymentOrderFromConfig_Linear(t *testing.T) {
	config := types.PreviewConfig{
		Services: map[string]types.ServiceConfig{
			"frontend": {DependsOn: []string{"api"}},
			"api":      {DependsOn: []string{"db"}},
			"db":       {},
		},
	}

	order, err := ResolveServiceDeploymentOrderFromConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idx := indexMap(order)
	if idx["db"] >= idx["api"] {
		t.Errorf("db should come before api, got: %v", order)
	}
	if idx["api"] >= idx["frontend"] {
		t.Errorf("api should come before frontend, got: %v", order)
	}
}

func TestResolveServiceDeploymentOrderFromConfig_NoDeps(t *testing.T) {
	config := types.PreviewConfig{
		Services: map[string]types.ServiceConfig{
			"a": {},
			"b": {},
			"c": {},
		},
	}

	order, err := ResolveServiceDeploymentOrderFromConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(order) != 3 {
		t.Errorf("expected 3 services, got %d", len(order))
	}
}

func TestResolveServiceDeploymentOrderFromConfig_Diamond(t *testing.T) {
	config := types.PreviewConfig{
		Services: map[string]types.ServiceConfig{
			"a": {DependsOn: []string{"b", "c"}},
			"b": {DependsOn: []string{"d"}},
			"c": {DependsOn: []string{"d"}},
			"d": {},
		},
	}

	order, err := ResolveServiceDeploymentOrderFromConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idx := indexMap(order)
	if idx["d"] >= idx["b"] || idx["d"] >= idx["c"] {
		t.Errorf("d should come before b and c, got: %v", order)
	}
	if idx["b"] >= idx["a"] || idx["c"] >= idx["a"] {
		t.Errorf("b and c should come before a, got: %v", order)
	}
}

func TestResolveServiceDeploymentOrderFromConfig_Cycle(t *testing.T) {
	config := types.PreviewConfig{
		Services: map[string]types.ServiceConfig{
			"a": {DependsOn: []string{"b"}},
			"b": {DependsOn: []string{"a"}},
		},
	}

	_, err := ResolveServiceDeploymentOrderFromConfig(config)
	if err == nil {
		t.Error("expected cycle error, got nil")
	}
}

func indexMap(order []string) map[string]int {
	m := make(map[string]int, len(order))
	for i, v := range order {
		m[v] = i
	}
	return m
}
