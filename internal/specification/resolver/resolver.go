package resolver

import "fmt"

type Resolver interface {
	Resolve(spec any) (*ResolvedSpec, error)
}

type ResolvedSpec struct {
	Context map[string]any
}

type DefaultResolver struct{}

func NewResolver() Resolver {
	return DefaultResolver{}
}

func (DefaultResolver) Resolve(spec any) (*ResolvedSpec, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec is nil")
	}
	return &ResolvedSpec{Context: map[string]any{"resolved": true}}, nil
}
