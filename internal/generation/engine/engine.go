package engine

import "fmt"

type GeneratorEngine interface {
	Generate(neir any) ([]Artifact, error)
}

type Artifact struct {
	Path    string
	Content []byte
}

type DefaultEngine struct{}

func NewEngine() GeneratorEngine {
	return DefaultEngine{}
}

func (DefaultEngine) Generate(neir any) ([]Artifact, error) {
	if neir == nil {
		return nil, fmt.Errorf("neir is nil")
	}
	return []Artifact{{Path: "README.md", Content: []byte("# Generated artifact\n")}}, nil
}
