package normalizer

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func TestNormalizerConvertsParsedSpecToStructuredValues(t *testing.T) {
	doc := &parser.SpecDocument{
		Project: "acme-api",
		Modules: []parser.Module{{Name: "auth", Path: "./internal/auth"}},
		Services: []parser.Service{{Name: "gateway", Kind: "http", Port: 8080}},
	}

	normalizer := NewNormalizer()
	normalized, err := normalizer.Normalize(doc)
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}
	if normalized.Values["project"] != "acme-api" {
		t.Fatalf("expected project acme-api, got %v", normalized.Values["project"])
	}
	if len(normalized.Values["modules"].([]map[string]any)) != 1 {
		t.Fatalf("expected one module, got %d", len(normalized.Values["modules"].([]map[string]any)))
	}
	if normalized.Values["services"].([]map[string]any)[0]["name"] != "gateway" {
		t.Fatalf("expected gateway service, got %v", normalized.Values["services"].([]map[string]any)[0]["name"])
	}
}
