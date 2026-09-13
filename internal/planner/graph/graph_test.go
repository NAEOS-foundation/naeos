// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"strings"
	"testing"
)

func TestNewGraphIsEmpty(t *testing.T) {
	g := New()
	if g.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestAddNode(t *testing.T) {
	g := New()
	err := g.AddNode(Node{ID: "a", Kind: NodeKindModule, Name: "module-a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", g.NodeCount())
	}
}

func TestAddNodeEmptyID(t *testing.T) {
	g := New()
	err := g.AddNode(Node{ID: "", Kind: NodeKindModule})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestAddNodeDuplicate(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	err := g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	if err == nil {
		t.Fatal("expected error for duplicate node")
	}
}

func TestAddEdge(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	err := g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", g.EdgeCount())
	}
}

func TestAddEdgeMissingNode(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	err := g.AddEdge(Edge{From: "a", To: "missing", Kind: EdgeKindDependency})
	if err == nil {
		t.Fatal("expected error for missing target node")
	}
}

func TestAddEdgeDuplicate(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	err := g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	if err == nil {
		t.Fatal("expected error for duplicate edge")
	}
}

func TestRemoveNode(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	err := g.RemoveNode("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.NodeCount() != 1 {
		t.Fatalf("expected 1 node after removal, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges after node removal, got %d", g.EdgeCount())
	}
}

func TestRemoveNodeNotFound(t *testing.T) {
	g := New()
	err := g.RemoveNode("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing node")
	}
}

func TestGetNeighbors(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "a", To: "c", Kind: EdgeKindDependency})
	neighbors := g.GetNeighbors("a")
	if len(neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(neighbors))
	}
}

func TestGetNeighborsNone(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	neighbors := g.GetNeighbors("a")
	if len(neighbors) != 0 {
		t.Fatalf("expected 0 neighbors, got %d", len(neighbors))
	}
}

func TestNodesAndEdges(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindService})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDataflow})

	nodes := g.Nodes()
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	edges := g.Edges()
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
}

func TestTopologicalSort(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDependency})

	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3 nodes in sort, got %d", len(sorted))
	}

	index := make(map[string]int)
	for i, n := range sorted {
		index[n.ID] = i
	}
	if index["a"] >= index["b"] || index["b"] >= index["c"] {
		t.Fatalf("topological order incorrect: %v", sorted)
	}
}

func TestTopologicalSortDetectsCycle(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "a", Kind: EdgeKindDependency})

	_, err := g.TopologicalSort()
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
}

func TestHasCycle(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})

	if g.HasCycle() {
		t.Fatal("expected no cycle")
	}

	_ = g.AddEdge(Edge{From: "b", To: "a", Kind: EdgeKindDependency})
	if !g.HasCycle() {
		t.Fatal("expected cycle detected")
	}
}

func TestInDegree(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "a", Kind: EdgeKindDependency})

	if g.InDegree("a") != 1 {
		t.Fatalf("expected in-degree 1 for a, got %d", g.InDegree("a"))
	}
	if g.OutDegree("a") != 1 {
		t.Fatalf("expected out-degree 1 for a, got %d", g.OutDegree("a"))
	}
}

func TestGetNode(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "x", Kind: NodeKindAPI, Name: "my-api"})
	n, ok := g.GetNode("x")
	if !ok {
		t.Fatal("expected node to be found")
	}
	if n.Name != "my-api" {
		t.Fatalf("expected name 'my-api', got %q", n.Name)
	}
	_, ok = g.GetNode("missing")
	if ok {
		t.Fatal("expected node not found")
	}
}

// --- New tests ---

func TestSubgraph(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindService})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindAPI})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDataflow})
	_ = g.AddEdge(Edge{From: "a", To: "c", Kind: EdgeKindExecution})

	sub := g.Subgraph([]string{"a", "c"})
	if sub.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes in subgraph, got %d", sub.NodeCount())
	}
	if sub.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge in subgraph (a->c), got %d", sub.EdgeCount())
	}
	_, ok := sub.GetNode("b")
	if ok {
		t.Fatal("node b should not be in subgraph")
	}
}

func TestSubgraphEmpty(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	sub := g.Subgraph([]string{})
	if sub.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", sub.NodeCount())
	}
}

func TestBFS(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "1", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "2", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "3", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "4", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "1", To: "2", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "1", To: "3", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "2", To: "4", Kind: EdgeKindDependency})

	result := g.BFS("1")
	if len(result) != 4 {
		t.Fatalf("expected 4 nodes in BFS, got %d", len(result))
	}
	if result[0].ID != "1" {
		t.Fatalf("expected first node to be 1, got %s", result[0].ID)
	}
	if result[1].ID != "2" && result[1].ID != "3" {
		t.Fatalf("expected second node to be 2 or 3, got %s", result[1].ID)
	}
}

func TestBFSInvalidStart(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	result := g.BFS("missing")
	if result != nil {
		t.Fatal("expected nil for missing start node")
	}
}

func TestDFSSingleNode(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	result := g.DFS("a")
	if len(result) != 1 {
		t.Fatalf("expected 1 node in DFS, got %d", len(result))
	}
	if result[0].ID != "a" {
		t.Fatalf("expected node a, got %s", result[0].ID)
	}
}

func TestDFSLinearChain(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDependency})

	result := g.DFS("a")
	if len(result) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(result))
	}
	if result[0].ID != "a" || result[1].ID != "b" || result[2].ID != "c" {
		t.Fatalf("unexpected DFS order: %v", result)
	}
}

func TestDFSInvalidStart(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	result := g.DFS("missing")
	if result != nil {
		t.Fatal("expected nil for missing start node")
	}
}

func TestFindAllPathsLinear(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "a", To: "c", Kind: EdgeKindExecution})

	paths := g.FindAllPaths("a", "c")
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
	for _, p := range paths {
		if len(p) == 0 {
			t.Fatal("expected non-empty path")
		}
		if p[0].ID != "a" {
			t.Fatalf("expected path to start with a, got %s", p[0].ID)
		}
		if p[len(p)-1].ID != "c" {
			t.Fatalf("expected path to end with c, got %s", p[len(p)-1].ID)
		}
	}
}

func TestFindAllPathsSameNode(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	paths := g.FindAllPaths("a", "a")
	if len(paths) != 1 {
		t.Fatalf("expected 1 path (trivial), got %d", len(paths))
	}
	if len(paths[0]) != 1 || paths[0][0].ID != "a" {
		t.Fatalf("expected [a], got %v", paths[0])
	}
}

func TestFindAllPathsNoPath(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "b", To: "a", Kind: EdgeKindDependency})

	paths := g.FindAllPaths("a", "b")
	if len(paths) != 0 {
		t.Fatalf("expected 0 paths, got %d", len(paths))
	}
}

func TestFindAllPathsInvalidNodes(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	paths := g.FindAllPaths("a", "missing")
	if paths != nil {
		t.Fatal("expected nil for missing target")
	}
	paths = g.FindAllPaths("missing", "a")
	if paths != nil {
		t.Fatal("expected nil for missing source")
	}
}

func TestMerge(t *testing.T) {
	g1 := New()
	_ = g1.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g1.AddEdge(Edge{From: "a", To: "a", Kind: EdgeKindDependency})

	g2 := New()
	_ = g2.AddNode(Node{ID: "b", Kind: NodeKindService})
	_ = g2.AddNode(Node{ID: "c", Kind: NodeKindAPI})
	_ = g2.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDataflow})

	err := g1.Merge(g2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g1.NodeCount() != 3 {
		t.Fatalf("expected 3 nodes after merge, got %d", g1.NodeCount())
	}
	if g1.EdgeCount() != 2 {
		t.Fatalf("expected 2 edges after merge, got %d", g1.EdgeCount())
	}
}

func TestMergeConflict(t *testing.T) {
	g1 := New()
	_ = g1.AddNode(Node{ID: "a", Kind: NodeKindModule})

	g2 := New()
	_ = g2.AddNode(Node{ID: "a", Kind: NodeKindService})

	err := g1.Merge(g2)
	if err == nil {
		t.Fatal("expected error for conflicting node IDs")
	}
}

func TestStronglyConnectedComponents(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "1", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "2", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "3", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "4", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "1", To: "2", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "2", To: "1", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "2", To: "3", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "3", To: "4", Kind: EdgeKindDependency})

	sccs := g.StronglyConnectedComponents()
	if len(sccs) != 3 {
		t.Fatalf("expected 3 SCCs, got %d", len(sccs))
	}

	foundAB := false
	for _, scc := range sccs {
		if len(scc) == 2 {
			ids := make(map[string]bool)
			for _, n := range scc {
				ids[n.ID] = true
			}
			if ids["1"] && ids["2"] {
				foundAB = true
			}
		}
	}
	if !foundAB {
		t.Fatal("expected SCC containing nodes 1 and 2")
	}
}

func TestStronglyConnectedComponentsNoEdges(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})

	sccs := g.StronglyConnectedComponents()
	if len(sccs) != 2 {
		t.Fatalf("expected 2 SCCs (each node alone), got %d", len(sccs))
	}
}

func TestTransitiveReduction(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "a", To: "c", Kind: EdgeKindDependency})

	reduced := g.TransitiveReduction()
	if reduced.NodeCount() != 3 {
		t.Fatalf("expected 3 nodes, got %d", reduced.NodeCount())
	}
	if reduced.EdgeCount() != 2 {
		t.Fatalf("expected 2 edges after reduction (a->c removed), got %d", reduced.EdgeCount())
	}
}

func TestTransitiveReductionNoRedundant(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})

	reduced := g.TransitiveReduction()
	if reduced.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", reduced.EdgeCount())
	}
}

func TestToDOT(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule, Name: "modA"})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindService, Name: "svcB"})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})

	dot := g.ToDOT()
	if !strings.Contains(dot, "digraph G {") {
		t.Fatal("expected DOT to contain 'digraph G {'")
	}
	if !strings.Contains(dot, `"a"`) {
		t.Fatal("expected DOT to contain node a")
	}
	if !strings.Contains(dot, `"b"`) {
		t.Fatal("expected DOT to contain node b")
	}
	if !strings.Contains(dot, "->") {
		t.Fatal("expected DOT to contain edge arrow")
	}
	if !strings.Contains(dot, "modA") {
		t.Fatal("expected DOT to contain label modA")
	}
}

func TestFromDOT(t *testing.T) {
	dot := `digraph G {
  "a" [label="modA", kind="module"];
  "b" [label="svcB", kind="service"];
  "a" -> "b" [kind="dependency"];
}`
	g, err := FromDOT(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", g.EdgeCount())
	}
	n, ok := g.GetNode("a")
	if !ok || n.Name != "modA" {
		t.Fatalf("expected node a with name modA, got %v", n)
	}
}

func TestFromDOTMinimal(t *testing.T) {
	dot := `digraph G {
  "x" -> "y";
}`
	g, err := FromDOT(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", g.EdgeCount())
	}
}

func TestFromDOTElim(t *testing.T) {
	dot := `digraph G {
}`
	g, err := FromDOT(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", g.NodeCount())
	}
}

func TestShortestPath(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "d", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "a", To: "d", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "d", To: "c", Kind: EdgeKindDependency})

	path, err := g.ShortestPath("a", "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 3 {
		t.Fatalf("expected 3 nodes in shortest path, got %d", len(path))
	}
	if path[0].ID != "a" || path[len(path)-1].ID != "c" {
		t.Fatalf("unexpected path start/end: %v", path)
	}
}

func TestShortestPathSameNode(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	path, err := g.ShortestPath("a", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 1 || path[0].ID != "a" {
		t.Fatalf("expected [a], got %v", path)
	}
}

func TestShortestPathNoPath(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_, err := g.ShortestPath("a", "b")
	if err == nil {
		t.Fatal("expected error for unreachable node")
	}
}

func TestShortestPathInvalidNodes(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_, err := g.ShortestPath("a", "missing")
	if err == nil {
		t.Fatal("expected error for missing target")
	}
	_, err = g.ShortestPath("missing", "a")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestClone(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule, Name: "modA"})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindService, Name: "svcB"})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDataflow})

	clone := g.Clone()
	if clone.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes in clone, got %d", clone.NodeCount())
	}
	if clone.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge in clone, got %d", clone.EdgeCount())
	}

	_ = g.AddNode(Node{ID: "c", Kind: NodeKindAPI})
	if g.NodeCount() != 3 {
		t.Fatalf("expected 3 nodes in original, got %d", g.NodeCount())
	}
	if clone.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes in clone after modification, got %d", clone.NodeCount())
	}
}

func TestCloneEmpty(t *testing.T) {
	g := New()
	clone := g.Clone()
	if clone.NodeCount() != 0 || clone.EdgeCount() != 0 {
		t.Fatal("expected empty clone")
	}
}

func TestNodesEdgesCopy(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "a", Kind: EdgeKindDependency})

	nodes := g.Nodes()
	nodes[0].ID = "changed"
	_, ok := g.GetNode("changed")
	if ok {
		t.Fatal("modifying returned slice should not affect graph")
	}

	edges := g.Edges()
	edges[0].From = "changed"
	if g.EdgeCount() != 1 {
		t.Fatal("modifying returned edge slice should not affect graph")
	}
}

func TestBFSDisconnected(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	// no edges
	result := g.BFS("a")
	if len(result) != 1 {
		t.Fatalf("expected 1 node (only start), got %d", len(result))
	}
}

func TestTopologicalSortSingle(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "x", Kind: NodeKindAPI})
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 1 || sorted[0].ID != "x" {
		t.Fatalf("expected [x], got %v", sorted)
	}
}

func TestStronglyConnectedComponentsCycle(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "c", To: "a", Kind: EdgeKindDependency})

	sccs := g.StronglyConnectedComponents()
	if len(sccs) != 1 {
		t.Fatalf("expected 1 SCC containing all nodes, got %d", len(sccs))
	}
	if len(sccs[0]) != 3 {
		t.Fatalf("expected 3 nodes in SCC, got %d", len(sccs[0]))
	}
}

func TestFromDOTInvalid(t *testing.T) {
	_, err := FromDOT("not a dot graph")
	if err != nil {
		// parser is lenient, it should not crash
		t.Logf("got error (acceptable): %v", err)
	}
}

func TestMergeEmpty(t *testing.T) {
	g1 := New()
	_ = g1.AddNode(Node{ID: "a", Kind: NodeKindModule})
	g2 := New()
	err := g1.Merge(g2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g1.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", g1.NodeCount())
	}
}

func TestTransitiveReductionPreservesReachability(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "c", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "d", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "a", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "c", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "c", To: "d", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "a", To: "c", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "a", To: "d", Kind: EdgeKindDependency})

	reduced := g.TransitiveReduction()
	if reduced.EdgeCount() != 3 {
		t.Fatalf("expected 3 edges after reduction, got %d", reduced.EdgeCount())
	}
	for _, e := range g.edges {
		path, err := reduced.ShortestPath(e.From, e.To)
		if err != nil {
			t.Fatalf("reachability lost for %s -> %s: %v", e.From, e.To, err)
		}
		if path[0].ID != e.From || path[len(path)-1].ID != e.To {
			t.Fatalf("path mismatch for %s -> %s", e.From, e.To)
		}
	}
}

func TestShortestPathDiamond(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "s", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "t", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "s", To: "a", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "s", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "a", To: "t", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "t", Kind: EdgeKindDependency})

	path, err := g.ShortestPath("s", "t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 3 {
		t.Fatalf("expected length 3, got %d", len(path))
	}
}

func TestFindAllPathsBranching(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{ID: "s", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "a", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "b", Kind: NodeKindModule})
	_ = g.AddNode(Node{ID: "t", Kind: NodeKindModule})
	_ = g.AddEdge(Edge{From: "s", To: "a", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "s", To: "b", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "a", To: "t", Kind: EdgeKindDependency})
	_ = g.AddEdge(Edge{From: "b", To: "t", Kind: EdgeKindDependency})

	paths := g.FindAllPaths("s", "t")
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
}
