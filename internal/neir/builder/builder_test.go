package builder

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
)

func TestBuilderCreatesNEIRFromResolvedSpec(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "acme-api",
		"modules": []map[string]any{{"name": "auth", "path": "./internal/auth"}},
		"services": []map[string]any{{"name": "gateway", "kind": "http", "port": 8080}},
	}}

	builder := NewBuilder()
	neir, err := builder.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Project != "acme-api" {
		t.Fatalf("expected project acme-api, got %v", neir.Project)
	}
	if len(neir.Modules) != 1 {
		t.Fatalf("expected one module, got %d", len(neir.Modules))
	}
}
