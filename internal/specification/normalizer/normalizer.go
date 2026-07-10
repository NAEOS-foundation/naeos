package normalizer

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type Normalizer interface {
	Normalize(doc any) (*NormalizedSpec, error)
}

type NormalizedSpec struct {
	Values map[string]any
}

type DefaultNormalizer struct{}

func NewNormalizer() Normalizer {
	return DefaultNormalizer{}
}

func (DefaultNormalizer) Normalize(doc any) (*NormalizedSpec, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	specDoc, ok := doc.(*parser.SpecDocument)
	if !ok {
		return &NormalizedSpec{Values: map[string]any{"source": doc}}, nil
	}

	modules := make([]map[string]any, 0, len(specDoc.Modules))
	for _, module := range specDoc.Modules {
		modules = append(modules, map[string]any{"name": module.Name, "path": module.Path})
	}

	services := make([]map[string]any, 0, len(specDoc.Services))
	for _, service := range specDoc.Services {
		services = append(services, map[string]any{"name": service.Name, "kind": service.Kind, "port": service.Port})
	}

	return &NormalizedSpec{Values: map[string]any{
		"project":  specDoc.Project,
		"modules":  modules,
		"services": services,
		"source":   specDoc,
	}}, nil
}
