package validator

import "fmt"

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
	return nil
}
