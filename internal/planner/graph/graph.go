package graph

type PlannerGraph struct {
	Nodes []Node
	Edges []Edge
}

type Node struct {
	ID   string
	Kind string
}

type Edge struct {
	From string
	To   string
}
