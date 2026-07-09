package builder

import "fmt"

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
	return &NEIR{Project: resolved, Modules: []any{}, Metadata: map[string]any{"version": "0.1"}}, nil
}
