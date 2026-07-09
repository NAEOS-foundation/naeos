package normalizer

import "fmt"

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
	return &NormalizedSpec{Values: map[string]any{"source": doc}}, nil
}
