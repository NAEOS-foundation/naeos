package validator

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/builder"
)

func TestValidatorAcceptsValidNEIR(t *testing.T) {
	neir := &builder.NEIR{Project: "acme-api", Modules: []any{map[string]any{"name": "auth"}}}
	v := NewValidator()
	if err := v.Validate(neir); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidatorRejectsIncompleteNEIR(t *testing.T) {
	neir := &builder.NEIR{Project: "", Modules: []any{}}
	v := NewValidator()
	if err := v.Validate(neir); err == nil {
		t.Fatalf("expected validation error for incomplete NEIR")
	}
}
