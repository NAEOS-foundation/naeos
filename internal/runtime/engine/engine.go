package engine

import "fmt"

type RuntimeEngine interface {
	Run(artifact any) error
}

type DefaultRuntimeEngine struct{}

func NewEngine() RuntimeEngine {
	return DefaultRuntimeEngine{}
}

func (DefaultRuntimeEngine) Run(artifact any) error {
	if artifact == nil {
		return fmt.Errorf("artifact is nil")
	}
	return nil
}
