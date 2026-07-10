package builder

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
)

type Builder interface {
	Build(resolved any) (*NEIR, error)
}

type NEIR struct {
	Project  any
	Modules  []any
	Metadata map[string]any
}

type DefaultBuilder struct{}

func NewBuilder() Builder {
	return DefaultBuilder{}
}

func (DefaultBuilder) Build(resolved any) (*NEIR, error) {
	if resolved == nil {
		return nil, fmt.Errorf("resolved spec is nil")
	}

	resolvedSpec, ok := resolved.(*resolver.ResolvedSpec)
	if !ok {
		return &NEIR{Project: resolved, Modules: []any{}, Metadata: map[string]any{"version": "0.1"}}, nil
	}

	modules := []any{}
	if rawModules, ok := resolvedSpec.Context["modules"].([]map[string]any); ok {
		for _, module := range rawModules {
			modules = append(modules, module)
		}
	}

	return &NEIR{
		Project: resolvedSpec.Context["project"],
		Modules: modules,
		Metadata: map[string]any{"version": "0.1", "source": resolvedSpec.Context},
	}, nil
}
