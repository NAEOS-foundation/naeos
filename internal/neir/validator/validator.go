package validator

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/neir/builder"
)

type Validator interface {
	Validate(neir any) error
}

type DefaultValidator struct{}

func NewValidator() Validator {
	return DefaultValidator{}
}

func (DefaultValidator) Validate(neir any) error {
	if neir == nil {
		return fmt.Errorf("neir is nil")
	}

	neirStruct, ok := neir.(*builder.NEIR)
	if !ok {
		return fmt.Errorf("neir is not a builder.NEIR")
	}
	if neirStruct.Project == nil || fmt.Sprint(neirStruct.Project) == "" {
		return fmt.Errorf("neir project must be set")
	}
	if len(neirStruct.Modules) == 0 {
		return fmt.Errorf("neir must contain at least one module")
	}
	return nil
}
