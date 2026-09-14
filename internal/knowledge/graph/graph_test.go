// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"strings"
	"testing"
)

func TestNewKnowledgeGraph(t *testing.T) {
	kg := New()
	if kg == nil {
		t.Fatal("expected non-nil knowledge graph")
	}
	if kg.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", kg.NodeCount())
	}
}

func TestAddNode(t *testing.T) {
	kg := New()
	err := kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision, Topic: "architecture", Component: "api"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kg.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", kg.NodeCount())
	}
}

func TestAddNodeEmptyID(t *testing.T) {
	kg := New()
	err := kg.AddNode(Node{ID: "", Type: NodeTypeDecision})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestAddNodeDuplicate(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	err := kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	if err == nil {
		t.Fatal("expected error for duplicate node")
	}
}

func TestGetNode(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision, Topic: "test"})
	node, ok := kg.GetNode("n1")
	if !ok {
		t.Fatal("expected to find node")
	}
	if node.Topic != "test" {
		t.Fatalf("expected topic 'test', got %s", node.Topic)
	}
}

func TestGetNodeNotFound(t *testing.T) {
	kg := New()
	_, ok := kg.GetNode("nonexistent")
	if ok {
		t.Fatal("expected not to find nonexistent node")
	}
}

func TestRemoveNode(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})

	err := kg.RemoveNode("n1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kg.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", kg.NodeCount())
	}
	if kg.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", kg.EdgeCount())
	}
}

func TestRemoveNodeNotFound(t *testing.T) {
	kg := New()
	err := kg.RemoveNode("nonexistent")
	if err == nil {
		t.Fatal("expected error for removing nonexistent node")
	}
}

func TestAddEdge(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	err := kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kg.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", kg.EdgeCount())
	}
}

func TestAddEdgeMissingNode(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	err := kg.AddEdge(Edge{From: "n1", To: "missing", Type: EdgeTypeDependsOn})
	if err == nil {
		t.Fatal("expected error for missing target node")
	}
}

func TestAddEdgeMissingSource(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeDecision})
	err := kg.AddEdge(Edge{From: "missing", To: "n2", Type: EdgeTypeDependsOn})
	if err == nil {
		t.Fatal("expected error for missing source node")
	}
}

func TestFindByTopic(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision, Topic: "auth"})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeDecision, Topic: "api"})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeDecision, Topic: "auth"})

	nodes := kg.FindByTopic("auth")
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestFindByComponent(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeComponent, Component: "api"})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent, Component: "db"})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeComponent, Component: "api"})

	nodes := kg.FindByComponent("api")
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestFindByType(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeDecision})

	nodes := kg.FindByType(NodeTypeDecision)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 decision nodes, got %d", len(nodes))
	}
}

func TestFindByVersion(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeModule, Version: "1.0"})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeModule, Version: "2.0"})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeModule, Version: "1.0"})

	nodes := kg.FindByVersion("1.0")
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestGetNeighbors(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "n1", To: "n3", Type: EdgeTypeDependsOn})

	neighbors := kg.GetNeighbors("n1")
	if len(neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(neighbors))
	}
}

func TestNodesAndEdges(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})

	nodes := kg.Nodes()
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	edges := kg.Edges()
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
}

func TestGetPredecessors(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})

	predecessors := kg.GetPredecessors("n2")
	if len(predecessors) != 1 {
		t.Fatalf("expected 1 predecessor, got %d", len(predecessors))
	}
	if predecessors[0].ID != "n1" {
		t.Fatalf("expected predecessor n1, got %s", predecessors[0].ID)
	}
}

func TestGetConnected(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "n2", To: "n3", Type: EdgeTypeDependsOn})

	connected := kg.GetConnected("n1")
	if len(connected) != 1 {
		t.Fatalf("expected 1 connected node (n2), got %d", len(connected))
	}
}

func TestFindByEdgeType(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "n1", To: "n3", Type: EdgeTypeImplements})

	edges := kg.FindByEdgeType(EdgeTypeDependsOn)
	if len(edges) != 1 {
		t.Fatalf("expected 1 depends_on edge, got %d", len(edges))
	}
}

func TestFindEdgesFrom(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})

	edges := kg.FindEdgesFrom("n1")
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge from n1, got %d", len(edges))
	}
}

func TestFindEdgesTo(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})

	edges := kg.FindEdgesTo("n2")
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge to n2, got %d", len(edges))
	}
}

func TestFindByMetadata(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision, Metadata: map[string]string{"env": "prod"}})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeDecision, Metadata: map[string]string{"env": "staging"}})

	nodes := kg.FindByMetadata("env", "prod")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
}

func TestHasPath(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeComponent})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeComponent})
	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "n2", To: "n3", Type: EdgeTypeDependsOn})

	if !kg.HasPath("n1", "n3") {
		t.Fatal("expected path from n1 to n3")
	}
	if kg.HasPath("n3", "n1") {
		t.Fatal("expected no path from n3 to n1")
	}
}

func TestFindByContentSubstring(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeDecision, Content: "use postgres for persistence"})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeDecision, Content: "use redis for caching"})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeComponent, Content: "auth service"})

	nodes := kg.FindByContentSubstring("postgres")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node with 'postgres', got %d", len(nodes))
	}
	if nodes[0].ID != "n1" {
		t.Fatalf("expected node n1, got %s", nodes[0].ID)
	}

	nodes = kg.FindByContentSubstring("use")
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes with 'use', got %d", len(nodes))
	}

	nodes = kg.FindByContentSubstring("nonexistent")
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestNewNodeTypes(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "svc1", Type: NodeTypeService, Topic: "api"})
	_ = kg.AddNode(Node{ID: "mod1", Type: NodeTypeModule, Topic: "auth"})
	_ = kg.AddNode(Node{ID: "api1", Type: NodeTypeAPI, Topic: "rest"})
	_ = kg.AddNode(Node{ID: "db1", Type: NodeTypeStorage, Topic: "postgres"})
	_ = kg.AddNode(Node{ID: "dep1", Type: NodeTypeDeployment, Topic: "k8s"})
	_ = kg.AddNode(Node{ID: "test1", Type: NodeTypeTesting, Topic: "unit"})
	_ = kg.AddNode(Node{ID: "sec1", Type: NodeTypeSecurity, Topic: "auth"})

	if kg.NodeCount() != 7 {
		t.Fatalf("expected 7 nodes, got %d", kg.NodeCount())
	}

	services := kg.FindByType(NodeTypeService)
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
}

func TestNewEdgeTypes(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "n1", Type: NodeTypeService})
	_ = kg.AddNode(Node{ID: "n2", Type: NodeTypeModule})
	_ = kg.AddNode(Node{ID: "n3", Type: NodeTypeStorage})

	_ = kg.AddEdge(Edge{From: "n1", To: "n2", Type: EdgeTypeContains})
	_ = kg.AddEdge(Edge{From: "n1", To: "n3", Type: EdgeTypeConnectsTo})

	contains := kg.FindByEdgeType(EdgeTypeContains)
	if len(contains) != 1 {
		t.Fatalf("expected 1 contains edge, got %d", len(contains))
	}
}

func TestShortestPathDirect(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeDependsOn})

	path := kg.ShortestPath("a", "b")
	if path == nil {
		t.Fatal("expected path")
	}
	if len(path) != 2 || path[0] != "a" || path[1] != "b" {
		t.Fatalf("expected [a b], got %v", path)
	}
}

func TestShortestPathThroughIntermediate(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})
	_ = kg.AddNode(Node{ID: "c"})
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "b", To: "c", Type: EdgeTypeDependsOn})

	path := kg.ShortestPath("a", "c")
	if path == nil {
		t.Fatal("expected path")
	}
	if len(path) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(path))
	}
	if path[0] != "a" || path[1] != "b" || path[2] != "c" {
		t.Fatalf("expected [a b c], got %v", path)
	}
}

func TestShortestPathSameNode(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})

	path := kg.ShortestPath("a", "a")
	if path == nil {
		t.Fatal("expected path")
	}
	if len(path) != 1 || path[0] != "a" {
		t.Fatalf("expected [a], got %v", path)
	}
}

func TestShortestPathNoPath(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})

	path := kg.ShortestPath("a", "b")
	if path != nil {
		t.Fatalf("expected nil, got %v", path)
	}
}

func TestShortestPathNonexistentNode(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})

	if kg.ShortestPath("a", "missing") != nil {
		t.Fatal("expected nil for missing target")
	}
	if kg.ShortestPath("missing", "a") != nil {
		t.Fatal("expected nil for missing source")
	}
}

func TestShortestPathTakesShorterRoute(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})
	_ = kg.AddNode(Node{ID: "c"})
	_ = kg.AddNode(Node{ID: "d"})
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "b", To: "d", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "a", To: "c", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "c", To: "d", Type: EdgeTypeDependsOn})

	path := kg.ShortestPath("a", "d")
	if path == nil {
		t.Fatal("expected path")
	}
	if len(path) != 3 {
		t.Fatalf("expected shortest path of 3, got %d", len(path))
	}
}

func TestTopologicalSort(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})
	_ = kg.AddNode(Node{ID: "c"})
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "b", To: "c", Type: EdgeTypeDependsOn})

	order, err := kg.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}

	indexOf := func(id string) int {
		for i, n := range order {
			if n == id {
				return i
			}
		}
		return -1
	}

	if indexOf("a") > indexOf("b") {
		t.Error("a should come before b")
	}
	if indexOf("b") > indexOf("c") {
		t.Error("b should come before c")
	}
}

func TestTopologicalSortCycle(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})
	_ = kg.AddNode(Node{ID: "c"})
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "b", To: "c", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "c", To: "a", Type: EdgeTypeDependsOn})

	_, err := kg.TopologicalSort()
	if err == nil {
		t.Fatal("expected error for cyclic graph")
	}
}

func TestTopologicalSortNoEdges(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})

	order, err := kg.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(order))
	}
}

func TestDetectCycles(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})
	_ = kg.AddNode(Node{ID: "c"})
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "b", To: "c", Type: EdgeTypeDependsOn})
	_ = kg.AddEdge(Edge{From: "c", To: "a", Type: EdgeTypeDependsOn})

	cycles := kg.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("expected to detect cycles")
	}

	foundCycle := false
	for _, cycle := range cycles {
		if len(cycle) >= 3 {
			foundCycle = true
			break
		}
	}
	if !foundCycle {
		t.Fatalf("expected cycle of length >= 3, got cycles: %v", cycles)
	}
}

func TestDetectCyclesNoCycle(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddNode(Node{ID: "b"})
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeDependsOn})

	cycles := kg.DetectCycles()
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %d", len(cycles))
	}
}

func TestDetectCyclesEmpty(t *testing.T) {
	kg := New()
	cycles := kg.DetectCycles()
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles in empty graph, got %d", len(cycles))
	}
}

func TestExportDOT(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a", Type: NodeTypeService, Topic: "api"})
	_ = kg.AddNode(Node{ID: "b", Type: NodeTypeModule, Topic: "auth"})
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeContains, Weight: 2})

	dot := kg.ExportDOT()
	if !strings.Contains(dot, "digraph knowledge") {
		t.Error("expected 'digraph knowledge' header")
	}
	if !strings.Contains(dot, `label="a`) {
		t.Error("expected node a label")
	}
	if !strings.Contains(dot, `label="b`) {
		t.Error("expected node b label")
	}
	if !strings.Contains(dot, "contains") {
		t.Error("expected edge label 'contains'")
	}
	if !strings.Contains(dot, "penwidth=2") {
		t.Error("expected penwidth=2 for weighted edge")
	}
}

func TestExportDOTEmpty(t *testing.T) {
	kg := New()
	dot := kg.ExportDOT()
	if !strings.Contains(dot, "digraph knowledge") {
		t.Error("expected header even for empty graph")
	}
}

func TestNodeColor(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		expected string
	}{
		{NodeTypeDecision, "#4A90D9"},
		{NodeTypeService, "#50C878"},
		{NodeTypeModule, "#FFD700"},
		{NodeTypeAPI, "#FF6347"},
		{NodeTypeStorage, "#9370DB"},
		{NodeTypeSecurity, "#FF4500"},
		{NodeTypeDeployment, "#20B2AA"},
		{NodeTypeTesting, "#DDA0DD"},
		{NodeTypeRequirement, "#E8E8E8"},
	}

	for _, tt := range tests {
		t.Run(string(tt.nodeType), func(t *testing.T) {
			color := nodeColor(tt.nodeType)
			if color != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, color)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	kg := New()
	for i := 0; i < 100; i++ {
		_ = kg.AddNode(Node{ID: string(rune('a' + i%26)), Type: NodeTypeDecision})
	}
	_ = kg.AddEdge(Edge{From: "a", To: "b", Type: EdgeTypeDependsOn})

	done := make(chan bool, 4)
	go func() {
		for i := 0; i < 50; i++ {
			kg.NodeCount()
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 50; i++ {
			kg.EdgeCount()
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 50; i++ {
			kg.HasPath("a", "b")
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 50; i++ {
			kg.Nodes()
		}
		done <- true
	}()
	for i := 0; i < 4; i++ {
		<-done
	}
}

func TestSelfLoopCycle(t *testing.T) {
	kg := New()
	_ = kg.AddNode(Node{ID: "a"})
	_ = kg.AddEdge(Edge{From: "a", To: "a", Type: EdgeTypeDependsOn})

	cycles := kg.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("expected to detect self-loop cycle")
	}
}
