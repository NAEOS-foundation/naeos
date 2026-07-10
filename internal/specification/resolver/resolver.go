package resolver

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
)

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

	normalized, ok := spec.(*normalizer.NormalizedSpec)
	if !ok {
		return &ResolvedSpec{Context: map[string]any{"resolved": true}}, nil
	}

	context := map[string]any{}
	for key, value := range normalized.Values {
		context[key] = value
	}

	return &ResolvedSpec{Context: context}, nil
}
