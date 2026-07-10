package resolver

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
)

func TestResolverBuildsContextFromNormalizedSpec(t *testing.T) {
	norm := &normalizer.NormalizedSpec{Values: map[string]any{
		"project": "acme-api",
		"modules": []map[string]any{{"name": "auth", "path": "./internal/auth"}},
		"services": []map[string]any{{"name": "gateway", "kind": "http", "port": 8080}},
	}}

	resolver := NewResolver()
	resolved, err := resolver.Resolve(norm)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.Context["project"] != "acme-api" {
		t.Fatalf("expected project acme-api, got %v", resolved.Context["project"])
	}
	if len(resolved.Context["modules"].([]map[string]any)) != 1 {
		t.Fatalf("expected one module, got %d", len(resolved.Context["modules"].([]map[string]any)))
	}
}
