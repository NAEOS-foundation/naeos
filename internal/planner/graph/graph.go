package graph

import (
	"fmt"
	"sort"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type NodeKind string

const (
	NodeKindService       NodeKind = "service"
	NodeKindModule        NodeKind = "module"
	NodeKindComponent     NodeKind = "component"
	NodeKindAPI           NodeKind = "api"
	NodeKindStorage       NodeKind = "storage"
	NodeKindInfra         NodeKind = "infrastructure"
	NodeKindDocumentation NodeKind = "documentation"
	NodeKindTesting       NodeKind = "testing"
	NodeKindDeployment    NodeKind = "deployment"
)

type EdgeKind string

const (
	EdgeKindDependency EdgeKind = "dependency"
	EdgeKindExecution  EdgeKind = "execution"
	EdgeKindDataflow   EdgeKind = "dataflow"
	EdgeKindPolicy     EdgeKind = "policy"
)

type Node struct {
	ID   string
	Kind NodeKind
	Name string
}

type Edge struct {
	From string
	To   string
	Kind EdgeKind
}

type PlannerGraph struct {
	nodes map[string]Node
	edges []Edge
}

func New() *PlannerGraph {
	return &PlannerGraph{
		nodes: make(map[string]Node),
	}
}

func (g *PlannerGraph) AddNode(n Node) error {
	if n.ID == "" {
		return naeoserr.New(naeoserr.ErrValidation, "node ID must not be empty")
	}
	if _, exists := g.nodes[n.ID]; exists {
		return naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("node %s already exists", n.ID))
	}
	g.nodes[n.ID] = n
	return nil
}

func (g *PlannerGraph) GetNode(id string) (Node, bool) {
	n, ok := g.nodes[id]
	return n, ok
}

func (g *PlannerGraph) RemoveNode(id string) error {
	if _, exists := g.nodes[id]; !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("node %s not found", id))
	}
	delete(g.nodes, id)
	filtered := g.edges[:0]
	for _, e := range g.edges {
		if e.From != id && e.To != id {
			filtered = append(filtered, e)
		}
	}
	g.edges = filtered
	return nil
}

func (g *PlannerGraph) AddEdge(e Edge) error {
	if _, exists := g.nodes[e.From]; !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("source node %s not found", e.From))
	}
	if _, exists := g.nodes[e.To]; !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("target node %s not found", e.To))
	}
	for _, existing := range g.edges {
		if existing.From == e.From && existing.To == e.To && existing.Kind == e.Kind {
			return naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("edge from %s to %s with kind %s already exists", e.From, e.To, e.Kind))
		}
	}
	g.edges = append(g.edges, e)
	return nil
}

func (g *PlannerGraph) GetNeighbors(id string) []Node {
	var neighbors []Node
	for _, e := range g.edges {
		if e.From == id {
			if n, ok := g.nodes[e.To]; ok {
				neighbors = append(neighbors, n)
			}
		}
	}
	return neighbors
}

func (g *PlannerGraph) Nodes() []Node {
	result := make([]Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		result = append(result, n)
	}
	return result
}

func (g *PlannerGraph) Edges() []Edge {
	result := make([]Edge, len(g.edges))
	copy(result, g.edges)
	return result
}

func (g *PlannerGraph) NodeCount() int {
	return len(g.nodes)
}

func (g *PlannerGraph) EdgeCount() int {
	return len(g.edges)
}

func (g *PlannerGraph) TopologicalSort() ([]Node, error) {
	inDegree := make(map[string]int)
	for id := range g.nodes {
		inDegree[id] = 0
	}
	for _, e := range g.edges {
		inDegree[e.To]++
	}

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var sorted []Node
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		sorted = append(sorted, g.nodes[id])

		for _, e := range g.edges {
			if e.From == id {
				inDegree[e.To]--
				if inDegree[e.To] == 0 {
					queue = append(queue, e.To)
				}
			}
		}
	}

	if len(sorted) != len(g.nodes) {
		return nil, naeoserr.New(naeoserr.ErrPipeline, "cycle detected in graph")
	}

	return sorted, nil
}

func (g *PlannerGraph) HasCycle() bool {
	_, err := g.TopologicalSort()
	return err != nil
}

func (g *PlannerGraph) InDegree(id string) int {
	count := 0
	for _, e := range g.edges {
		if e.To == id {
			count++
		}
	}
	return count
}

func (g *PlannerGraph) OutDegree(id string) int {
	count := 0
	for _, e := range g.edges {
		if e.From == id {
			count++
		}
	}
	return count
}

func (g *PlannerGraph) Subgraph(ids []string) *PlannerGraph {
	sub := New()
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	for _, id := range ids {
		if n, ok := g.nodes[id]; ok {
			_ = sub.AddNode(n)
		}
	}
	for _, e := range g.edges {
		if idSet[e.From] && idSet[e.To] {
			_ = sub.AddEdge(e)
		}
	}
	return sub
}

func (g *PlannerGraph) BFS(startID string) []Node {
	if _, ok := g.nodes[startID]; !ok {
		return nil
	}
	visited := make(map[string]bool)
	var result []Node
	queue := []string{startID}
	visited[startID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, g.nodes[current])

		for _, e := range g.edges {
			if e.From == current && !visited[e.To] {
				visited[e.To] = true
				queue = append(queue, e.To)
			}
		}
	}
	return result
}

func (g *PlannerGraph) DFS(startID string) []Node {
	if _, ok := g.nodes[startID]; !ok {
		return nil
	}
	visited := make(map[string]bool)
	var result []Node
	g.dfsVisit(startID, visited, &result)
	return result
}

func (g *PlannerGraph) dfsVisit(id string, visited map[string]bool, result *[]Node) {
	visited[id] = true
	*result = append(*result, g.nodes[id])
	for _, e := range g.edges {
		if e.From == id && !visited[e.To] {
			g.dfsVisit(e.To, visited, result)
		}
	}
}

func (g *PlannerGraph) FindAllPaths(from, to string) [][]Node {
	if _, ok := g.nodes[from]; !ok {
		return nil
	}
	if _, ok := g.nodes[to]; !ok {
		return nil
	}
	var paths [][]Node
	visited := make(map[string]bool)
	var current []Node
	g.findAllPathsDFS(from, to, visited, &current, &paths)
	return paths
}

func (g *PlannerGraph) findAllPathsDFS(current, target string, visited map[string]bool, path *[]Node, paths *[][]Node) {
	visited[current] = true
	*path = append(*path, g.nodes[current])

	if current == target {
		pathCopy := make([]Node, len(*path))
		copy(pathCopy, *path)
		*paths = append(*paths, pathCopy)
	} else {
		for _, e := range g.edges {
			if e.From == current && !visited[e.To] {
				g.findAllPathsDFS(e.To, target, visited, path, paths)
			}
		}
	}

	*path = (*path)[:len(*path)-1]
	visited[current] = false
}

func (g *PlannerGraph) Merge(other *PlannerGraph) error {
	for _, n := range other.nodes {
		if _, exists := g.nodes[n.ID]; exists {
			return naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("node %s already exists in graph", n.ID))
		}
	}
	for _, n := range other.nodes {
		g.nodes[n.ID] = n
	}
	g.edges = append(g.edges, other.edges...)
	return nil
}

func (g *PlannerGraph) StronglyConnectedComponents() [][]Node {
	indexCounter := 0
	stack := []string{}
	onStack := make(map[string]bool)
	indices := make(map[string]int)
	lowlinks := make(map[string]int)
	var components [][]Node

	var strongconnect func(id string)
	strongconnect = func(id string) {
		indices[id] = indexCounter
		lowlinks[id] = indexCounter
		indexCounter++
		stack = append(stack, id)
		onStack[id] = true

		for _, e := range g.edges {
			if e.From == id {
				if _, ok := indices[e.To]; !ok {
					strongconnect(e.To)
					if lowlinks[e.To] < lowlinks[id] {
						lowlinks[id] = lowlinks[e.To]
					}
				} else if onStack[e.To] {
					if indices[e.To] < lowlinks[id] {
						lowlinks[id] = indices[e.To]
					}
				}
			}
		}

		if lowlinks[id] == indices[id] {
			var component []Node
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				component = append(component, g.nodes[w])
				if w == id {
					break
				}
			}
			components = append(components, component)
		}
	}

	for id := range g.nodes {
		if _, ok := indices[id]; !ok {
			strongconnect(id)
		}
	}
	return components
}

func (g *PlannerGraph) TransitiveReduction() *PlannerGraph {
	reduced := New()
	for _, n := range g.nodes {
		_ = reduced.AddNode(n)
	}
	for _, e := range g.edges {
		_ = reduced.AddEdge(e)
	}

	reachable := g.computeReachability()

	for _, e := range g.edges {
		g.removeEdge(reduced, e)
		if !g.reachableWithoutEdge(reachable, e, reduced) {
			_ = reduced.AddEdge(e)
		}
	}
	return reduced
}

func (g *PlannerGraph) computeReachability() map[string]map[string]bool {
	reachable := make(map[string]map[string]bool)
	for id := range g.nodes {
		reachable[id] = make(map[string]bool)
		for _, e := range g.edges {
			if e.From == id {
				reachable[id][e.To] = true
			}
		}
	}

	changed := true
	for changed {
		changed = false
		for id := range g.nodes {
			for intermediate := range reachable[id] {
				for dest := range reachable[intermediate] {
					if !reachable[id][dest] {
						reachable[id][dest] = true
						changed = true
					}
				}
			}
		}
	}
	return reachable
}

func (g *PlannerGraph) removeEdge(pg *PlannerGraph, e Edge) {
	filtered := pg.edges[:0]
	for _, existing := range pg.edges {
		if existing.From != e.From || existing.To != e.To || existing.Kind != e.Kind {
			filtered = append(filtered, existing)
		}
	}
	pg.edges = filtered
}

func (g *PlannerGraph) reachableWithoutEdge(_ map[string]map[string]bool, removed Edge, temp *PlannerGraph) bool {
	// Check if there's still a path from removed.From to removed.To in temp graph
	// using BFS on temp graph
	if _, ok := temp.nodes[removed.From]; !ok {
		return false
	}
	visited := make(map[string]bool)
	queue := []string{removed.From}
	visited[removed.From] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == removed.To {
			return true
		}
		for _, e := range temp.edges {
			if e.From == current && !visited[e.To] {
				visited[e.To] = true
				queue = append(queue, e.To)
			}
		}
	}
	return false
}

func (g *PlannerGraph) ToDOT() string {
	var sb strings.Builder
	sb.WriteString("digraph G {\n")
	ids := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		n := g.nodes[id]
		label := n.Name
		if label == "" {
			label = n.ID
		}
		fmt.Fprintf(&sb, "  %q [label=%q, kind=%q];\n", n.ID, label, n.Kind)
	}
	for _, e := range g.edges {
		fmt.Fprintf(&sb, "  %q -> %q [kind=%q];\n", e.From, e.To, e.Kind)
	}
	sb.WriteString("}\n")
	return sb.String()
}

func FromDOT(dot string) (*PlannerGraph, error) {
	g := New()
	lines := strings.Split(dot, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "digraph G {" || line == "}" {
			continue
		}

		if strings.Contains(line, "->") {
			e, err := parseDOTEdge(line)
			if err != nil {
				return nil, err
			}
			if _, ok := g.nodes[e.From]; !ok {
				_ = g.AddNode(Node{ID: e.From, Kind: NodeKindModule})
			}
			if _, ok := g.nodes[e.To]; !ok {
				_ = g.AddNode(Node{ID: e.To, Kind: NodeKindModule})
			}
			_ = g.AddEdge(e)
		} else if strings.Contains(line, "[") {
			n, err := parseDOTNode(line)
			if err != nil {
				return nil, err
			}
			_ = g.AddNode(n)
		}
	}
	return g, nil
}

func parseDOTNode(line string) (Node, error) {
	parts := strings.SplitN(line, " [", 2)
	if len(parts) < 2 {
		return Node{}, naeoserr.New(naeoserr.ErrParse, fmt.Sprintf("invalid DOT node: %s", line))
	}
	id := strings.TrimSpace(parts[0])
	id = strings.Trim(id, "\"")

	n := Node{ID: id, Kind: NodeKindModule}
	attrs := strings.TrimSuffix(parts[1], ";")
	attrs = strings.Trim(attrs, "[]")
	for _, attr := range strings.Split(attrs, ",") {
		kv := strings.SplitN(strings.TrimSpace(attr), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
		switch key {
		case "label":
			n.Name = val
		case "kind":
			n.Kind = NodeKind(val)
		}
	}
	return n, nil
}

func parseDOTEdge(line string) (Edge, error) {
	parts := strings.SplitN(line, " -> ", 2)
	if len(parts) < 2 {
		return Edge{}, naeoserr.New(naeoserr.ErrParse, fmt.Sprintf("invalid DOT edge: %s", line))
	}
	from := strings.Trim(strings.TrimSpace(parts[0]), "\"")
	right := strings.SplitN(parts[1], " [", 2)
	to := strings.Trim(strings.TrimSpace(right[0]), "\"")

	e := Edge{From: from, To: to, Kind: EdgeKindDependency}
	if len(right) > 1 {
		attrs := strings.TrimSuffix(right[1], ";")
		attrs = strings.Trim(attrs, "[]")
		for _, attr := range strings.Split(attrs, ",") {
			kv := strings.SplitN(strings.TrimSpace(attr), "=", 2)
			if len(kv) != 2 {
				continue
			}
			key := strings.TrimSpace(kv[0])
			val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
			if key == "kind" {
				e.Kind = EdgeKind(val)
			}
		}
	}
	return e, nil
}

func (g *PlannerGraph) ShortestPath(from, to string) ([]Node, error) {
	if _, ok := g.nodes[from]; !ok {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("node %s not found", from))
	}
	if _, ok := g.nodes[to]; !ok {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("node %s not found", to))
	}
	if from == to {
		return []Node{g.nodes[from]}, nil
	}

	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{from}
	visited[from] = true

	found := false
	for len(queue) > 0 && !found {
		current := queue[0]
		queue = queue[1:]
		for _, e := range g.edges {
			if e.From == current && !visited[e.To] {
				visited[e.To] = true
				parent[e.To] = current
				if e.To == to {
					found = true
					break
				}
				queue = append(queue, e.To)
			}
		}
	}

	if !found {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("no path from %s to %s", from, to))
	}

	var path []Node
	for cur := to; cur != ""; cur = parent[cur] {
		path = append([]Node{g.nodes[cur]}, path...)
		if cur == from {
			break
		}
	}
	return path, nil
}

func (g *PlannerGraph) Clone() *PlannerGraph {
	clone := New()
	for _, n := range g.nodes {
		_ = clone.AddNode(Node{ID: n.ID, Kind: n.Kind, Name: n.Name})
	}
	for _, e := range g.edges {
		_ = clone.AddEdge(Edge{From: e.From, To: e.To, Kind: e.Kind})
	}
	return clone
}
