package graph

import (
	"fmt"
	"strings"
	"sync"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type NodeType string

const (
	NodeTypeDecision       NodeType = "decision"
	NodeTypeRequirement    NodeType = "requirement"
	NodeTypeRationale      NodeType = "rationale"
	NodeTypeComponent      NodeType = "component"
	NodeTypePolicy         NodeType = "policy"
	NodeTypeImplementation NodeType = "implementation"
	NodeTypeHistorical     NodeType = "historical"
	NodeTypeService        NodeType = "service"
	NodeTypeModule         NodeType = "module"
	NodeTypeAPI            NodeType = "api"
	NodeTypeStorage        NodeType = "storage"
	NodeTypeDeployment     NodeType = "deployment"
	NodeTypeTesting        NodeType = "testing"
	NodeTypeSecurity       NodeType = "security"
)

type EdgeType string

const (
	EdgeTypeDependsOn     EdgeType = "depends_on"
	EdgeTypeImplements    EdgeType = "implements"
	EdgeTypeRelatedTo     EdgeType = "related_to"
	EdgeTypeSupersedes    EdgeType = "supersedes"
	EdgeTypeConflictsWith EdgeType = "conflicts_with"
	EdgeTypeContains      EdgeType = "contains"
	EdgeTypeExposes       EdgeType = "exposes"
	EdgeTypeConnectsTo    EdgeType = "connects_to"
	EdgeTypeDeploysTo     EdgeType = "deploys_to"
	EdgeTypeTests         EdgeType = "tests"
	EdgeTypeSecures       EdgeType = "secures"
	EdgeTypeUses          EdgeType = "uses"
	EdgeTypeExtends       EdgeType = "extends"
)

type Node struct {
	ID        string
	Type      NodeType
	Topic     string
	Component string
	Version   string
	Content   string
	Metadata  map[string]string
}

type Edge struct {
	From   string
	To     string
	Type   EdgeType
	Weight int
}

type KnowledgeGraph struct {
	mu    sync.RWMutex
	nodes map[string]*Node
	edges []Edge
}

func New() *KnowledgeGraph {
	return &KnowledgeGraph{
		nodes: make(map[string]*Node),
	}
}

func (kg *KnowledgeGraph) AddNode(n Node) error {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	if n.ID == "" {
		return naeoserr.New(naeoserr.ErrValidation, "node ID must not be empty")
	}
	if _, exists := kg.nodes[n.ID]; exists {
		return naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("node %s already exists", n.ID))
	}
	kg.nodes[n.ID] = &n
	return nil
}

func (kg *KnowledgeGraph) GetNode(id string) (*Node, bool) {
	kg.mu.RLock()
	defer kg.mu.RUnlock()
	n, ok := kg.nodes[id]
	return n, ok
}

func (kg *KnowledgeGraph) RemoveNode(id string) error {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	if _, exists := kg.nodes[id]; !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("node %s not found", id))
	}
	delete(kg.nodes, id)

	filtered := kg.edges[:0]
	for _, e := range kg.edges {
		if e.From != id && e.To != id {
			filtered = append(filtered, e)
		}
	}
	kg.edges = filtered
	return nil
}

func (kg *KnowledgeGraph) AddEdge(e Edge) error {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	if _, exists := kg.nodes[e.From]; !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("source node %s not found", e.From))
	}
	if _, exists := kg.nodes[e.To]; !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("target node %s not found", e.To))
	}
	kg.edges = append(kg.edges, e)
	return nil
}

func (kg *KnowledgeGraph) FindByTopic(topic string) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []*Node
	for _, n := range kg.nodes {
		if n.Topic == topic {
			result = append(result, n)
		}
	}
	return result
}

func (kg *KnowledgeGraph) FindByComponent(component string) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []*Node
	for _, n := range kg.nodes {
		if n.Component == component {
			result = append(result, n)
		}
	}
	return result
}

func (kg *KnowledgeGraph) FindByVersion(version string) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []*Node
	for _, n := range kg.nodes {
		if n.Version == version {
			result = append(result, n)
		}
	}
	return result
}

func (kg *KnowledgeGraph) FindByType(nodeType NodeType) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []*Node
	for _, n := range kg.nodes {
		if n.Type == nodeType {
			result = append(result, n)
		}
	}
	return result
}

func (kg *KnowledgeGraph) FindByEdgeType(edgeType EdgeType) []Edge {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []Edge
	for _, e := range kg.edges {
		if e.Type == edgeType {
			result = append(result, e)
		}
	}
	return result
}

func (kg *KnowledgeGraph) FindEdgesFrom(id string) []Edge {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []Edge
	for _, e := range kg.edges {
		if e.From == id {
			result = append(result, e)
		}
	}
	return result
}

func (kg *KnowledgeGraph) FindEdgesTo(id string) []Edge {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []Edge
	for _, e := range kg.edges {
		if e.To == id {
			result = append(result, e)
		}
	}
	return result
}

func (kg *KnowledgeGraph) Nodes() []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	result := make([]*Node, 0, len(kg.nodes))
	for _, n := range kg.nodes {
		result = append(result, n)
	}
	return result
}

func (kg *KnowledgeGraph) Edges() []Edge {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	result := make([]Edge, len(kg.edges))
	copy(result, kg.edges)
	return result
}

func (kg *KnowledgeGraph) NodeCount() int {
	kg.mu.RLock()
	defer kg.mu.RUnlock()
	return len(kg.nodes)
}

func (kg *KnowledgeGraph) EdgeCount() int {
	kg.mu.RLock()
	defer kg.mu.RUnlock()
	return len(kg.edges)
}

func (kg *KnowledgeGraph) GetNeighbors(id string) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []*Node
	for _, e := range kg.edges {
		if e.From == id {
			if n, ok := kg.nodes[e.To]; ok {
				result = append(result, n)
			}
		}
	}
	return result
}

func (kg *KnowledgeGraph) GetPredecessors(id string) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []*Node
	for _, e := range kg.edges {
		if e.To == id {
			if n, ok := kg.nodes[e.From]; ok {
				result = append(result, n)
			}
		}
	}
	return result
}

func (kg *KnowledgeGraph) GetConnected(id string) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	seen := make(map[string]bool)
	var result []*Node
	for _, e := range kg.edges {
		if e.From == id {
			if !seen[e.To] {
				if n, ok := kg.nodes[e.To]; ok {
					result = append(result, n)
					seen[e.To] = true
				}
			}
		}
		if e.To == id {
			if !seen[e.From] {
				if n, ok := kg.nodes[e.From]; ok {
					result = append(result, n)
					seen[e.From] = true
				}
			}
		}
	}
	return result
}

func (kg *KnowledgeGraph) FindByMetadata(key, value string) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []*Node
	for _, n := range kg.nodes {
		if n.Metadata != nil && n.Metadata[key] == value {
			result = append(result, n)
		}
	}
	return result
}

func (kg *KnowledgeGraph) FindByContentSubstring(substring string) []*Node {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var result []*Node
	for _, n := range kg.nodes {
		if strings.Contains(n.Content, substring) {
			result = append(result, n)
		}
	}
	return result
}

func (kg *KnowledgeGraph) HasPath(from, to string) bool {
	path := kg.ShortestPath(from, to)
	return path != nil
}

func (kg *KnowledgeGraph) ShortestPath(from, to string) []string {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	if from == to {
		if _, ok := kg.nodes[from]; ok {
			return []string{from}
		}
		return nil
	}

	if _, ok := kg.nodes[from]; !ok {
		return nil
	}
	if _, ok := kg.nodes[to]; !ok {
		return nil
	}

	type entry struct {
		id   string
		path []string
	}

	visited := make(map[string]bool)
	queue := []entry{{id: from, path: []string{from}}}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, e := range kg.edges {
			if e.From == current.id && !visited[e.To] {
				newPath := make([]string, len(current.path), len(current.path)+1)
				copy(newPath, current.path)
				newPath = append(newPath, e.To)
				if e.To == to {
					return newPath
				}
				visited[e.To] = true
				queue = append(queue, entry{id: e.To, path: newPath})
			}
		}
	}
	return nil
}

func (kg *KnowledgeGraph) TopologicalSort() ([]string, error) {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	inDegree := make(map[string]int)
	for id := range kg.nodes {
		inDegree[id] = 0
	}
	for _, e := range kg.edges {
		inDegree[e.To]++
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var order []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		order = append(order, n)

		for _, e := range kg.edges {
			if e.From == n {
				inDegree[e.To]--
				if inDegree[e.To] == 0 {
					queue = append(queue, e.To)
				}
			}
		}
	}

	if len(order) != len(kg.nodes) {
		return nil, naeoserr.New(naeoserr.ErrValidation, "graph contains a cycle")
	}
	return order, nil
}

func (kg *KnowledgeGraph) DetectCycles() [][]string {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(id string)
	dfs = func(id string) {
		visited[id] = true
		recStack[id] = true
		path = append(path, id)

		for _, e := range kg.edges {
			if e.From == id {
				if !visited[e.To] {
					dfs(e.To)
				} else if recStack[e.To] {
					cycle := []string{e.To}
					for i := len(path) - 1; i >= 0; i-- {
						cycle = append(cycle, path[i])
						if path[i] == e.To {
							break
						}
					}
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[id] = false
	}

	for id := range kg.nodes {
		if !visited[id] {
			dfs(id)
		}
	}
	return cycles
}

func (kg *KnowledgeGraph) ExportDOT() string {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("digraph knowledge {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box];\n\n")

	for _, n := range kg.nodes {
		label := n.ID
		if n.Topic != "" {
			label = n.ID + "\\n(" + n.Topic + ")"
		}
		attrs := fmt.Sprintf(`label="%s"`, label)
		if n.Type != "" {
			attrs += fmt.Sprintf(` style=filled fillcolor="%s"`, nodeColor(n.Type))
		}
		fmt.Fprintf(&sb, "  %s [%s];\n", n.ID, attrs)
	}

	sb.WriteString("\n")
	for _, e := range kg.edges {
		attrs := fmt.Sprintf(`label="%s"`, e.Type)
		if e.Weight > 1 {
			attrs += fmt.Sprintf(` penwidth=%d`, e.Weight)
		}
		fmt.Fprintf(&sb, "  %s -> %s [%s];\n", e.From, e.To, attrs)
	}

	sb.WriteString("}\n")
	return sb.String()
}

func nodeColor(t NodeType) string {
	switch t {
	case NodeTypeDecision:
		return "#4A90D9"
	case NodeTypeService:
		return "#50C878"
	case NodeTypeModule:
		return "#FFD700"
	case NodeTypeAPI:
		return "#FF6347"
	case NodeTypeStorage:
		return "#9370DB"
	case NodeTypeSecurity:
		return "#FF4500"
	case NodeTypeDeployment:
		return "#20B2AA"
	case NodeTypeTesting:
		return "#DDA0DD"
	default:
		return "#E8E8E8"
	}
}
