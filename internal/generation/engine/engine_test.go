package engine

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/builder"
)

func TestGeneratorCreatesArtifactsFromNEIR(t *testing.T) {
	neir := &builder.NEIR{
		Project: "acme-api",
		Modules: []any{map[string]any{"name": "auth", "path": "./internal/auth"}},
		Metadata: map[string]any{"version": "0.1"},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(artifacts) < 2 {
		t.Fatalf("expected at least two artifacts, got %d", len(artifacts))
	}
	if artifacts[0].Path != "README.md" {
		t.Fatalf("expected README artifact first, got %s", artifacts[0].Path)
	}
}
