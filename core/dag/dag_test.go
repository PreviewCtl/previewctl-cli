package dag

import (
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph[string]()
	if g.adjList == nil || g.inDegree == nil || g.vertices == nil {
		t.Fatal("NewGraph should initialize all internal maps")
	}
}

func TestAddVertex(t *testing.T) {
	g := NewGraph[string]()
	g.AddVertex("a")

	if !g.vertices["a"] {
		t.Error("vertex 'a' should exist")
	}
	if g.inDegree["a"] != 0 {
		t.Error("in-degree of new vertex should be 0")
	}
}

func TestAddVertexIdempotent(t *testing.T) {
	g := NewGraph[string]()
	g.AddVertex("a")
	g.AddEdge("b", "a") // inDegree["a"] = 1
	g.AddVertex("a")    // should not reset

	if g.inDegree["a"] != 1 {
		t.Error("AddVertex should not reset in-degree of existing vertex")
	}
}

func TestAddEdge(t *testing.T) {
	g := NewGraph[string]()
	g.AddEdge("a", "b")

	if !g.vertices["a"] || !g.vertices["b"] {
		t.Error("AddEdge should create both vertices")
	}
	if g.inDegree["b"] != 1 {
		t.Error("in-degree of 'b' should be 1")
	}
	if len(g.adjList["a"]) != 1 || g.adjList["a"][0] != "b" {
		t.Error("adjacency list of 'a' should contain 'b'")
	}
}

func TestTopoSortLinearChain(t *testing.T) {
	g := NewGraph[string]()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")

	sorted, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idx := make(map[string]int, len(sorted))
	for i, v := range sorted {
		idx[v] = i
	}

	if idx["a"] >= idx["b"] || idx["b"] >= idx["c"] {
		t.Errorf("expected a < b < c, got order: %v", sorted)
	}
}

func TestTopoSortDiamond(t *testing.T) {
	g := NewGraph[string]()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "d")

	sorted, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idx := make(map[string]int, len(sorted))
	for i, v := range sorted {
		idx[v] = i
	}

	if idx["a"] >= idx["b"] || idx["a"] >= idx["c"] {
		t.Errorf("'a' should come before 'b' and 'c', got: %v", sorted)
	}
	if idx["b"] >= idx["d"] || idx["c"] >= idx["d"] {
		t.Errorf("'d' should come after 'b' and 'c', got: %v", sorted)
	}
}

func TestTopoSortSingleVertex(t *testing.T) {
	g := NewGraph[string]()
	g.AddVertex("only")

	sorted, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 1 || sorted[0] != "only" {
		t.Errorf("expected [only], got %v", sorted)
	}
}

func TestTopoSortDisconnected(t *testing.T) {
	g := NewGraph[string]()
	g.AddVertex("a")
	g.AddVertex("b")
	g.AddVertex("c")

	sorted, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 3 {
		t.Errorf("expected 3 vertices, got %d", len(sorted))
	}
}

func TestTopoSortCycleDetection(t *testing.T) {
	g := NewGraph[string]()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	g.AddEdge("c", "a")

	_, err := g.TopoSort()
	if err == nil {
		t.Error("expected cycle error, got nil")
	}
}

func TestTopoSortSelfLoop(t *testing.T) {
	g := NewGraph[string]()
	g.AddEdge("a", "a")

	_, err := g.TopoSort()
	if err == nil {
		t.Error("expected cycle error for self-loop, got nil")
	}
}

func TestTopoSortIntGraph(t *testing.T) {
	g := NewGraph[int]()
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)

	sorted, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idx := make(map[int]int, len(sorted))
	for i, v := range sorted {
		idx[v] = i
	}
	if idx[1] >= idx[2] || idx[2] >= idx[3] {
		t.Errorf("expected 1 < 2 < 3, got order: %v", sorted)
	}
}
