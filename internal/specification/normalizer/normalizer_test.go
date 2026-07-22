package normalizer

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func TestNormalizerConvertsParsedSpecToStructuredValues(t *testing.T) {
	doc := &parser.SpecDocument{
		Project:  "acme-api",
		Modules:  []parser.Module{{Name: "auth", Path: "./internal/auth"}},
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

func TestNormalizerHandlesArchitectureDeploymentTesting(t *testing.T) {
	doc := &parser.SpecDocument{
		Project: "full-api",
		Modules: []parser.Module{
			{Name: "auth", Path: "./internal/auth", Description: "Auth module", Dependencies: []string{"crypto"}},
		},
		Services: []parser.Service{
			{Name: "gateway", Kind: "http", Port: 8080, Description: "API gateway", Endpoints: []parser.Endpoint{
				{Method: "GET", Path: "/health", Action: "healthCheck"},
			}},
		},
		Architecture: &parser.Architecture{
			Pattern:     "hexagonal",
			Description: "Clean architecture",
			Principles:  []string{"separation of concerns", "dependency inversion"},
		},
		Deployment: &parser.Deployment{
			Strategy:     "rolling",
			Environments: []string{"staging", "production"},
		},
		Testing: &parser.Testing{
			Strategy: "unit-integration",
			Coverage: "80%",
		},
	}

	normalized, err := NewNormalizer().Normalize(doc)
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}

	arch := normalized.Values["architecture"].(map[string]any)
	if arch["pattern"] != "hexagonal" {
		t.Fatalf("expected pattern hexagonal, got %v", arch["pattern"])
	}
	principles := arch["principles"].([]string)
	if len(principles) != 2 {
		t.Fatalf("expected 2 principles, got %d", len(principles))
	}

	deploy := normalized.Values["deployment"].(map[string]any)
	if deploy["strategy"] != "rolling" {
		t.Fatalf("expected strategy rolling, got %v", deploy["strategy"])
	}
	envs := deploy["environments"].([]map[string]any)
	if len(envs) != 2 {
		t.Fatalf("expected 2 environments, got %d", len(envs))
	}

	test := normalized.Values["testing"].(map[string]any)
	if test["strategy"] != "unit-integration" {
		t.Fatalf("expected strategy unit-integration, got %v", test["strategy"])
	}
	if test["coverage"] != "80%" {
		t.Fatalf("expected coverage 80%%, got %v", test["coverage"])
	}

	modules := normalized.Values["modules"].([]map[string]any)
	if modules[0]["description"] != "Auth module" {
		t.Fatalf("expected module description, got %v", modules[0]["description"])
	}
	deps := modules[0]["dependencies"].([]string)
	if deps[0] != "crypto" {
		t.Fatalf("expected dependency crypto, got %v", deps[0])
	}

	services := normalized.Values["services"].([]map[string]any)
	if services[0]["description"] != "API gateway" {
		t.Fatalf("expected service description, got %v", services[0]["description"])
	}
	eps := services[0]["endpoints"].([]map[string]any)
	if eps[0]["method"] != "GET" {
		t.Fatalf("expected endpoint method GET, got %v", eps[0]["method"])
	}
}

func TestNormalizerNilDocumentReturnsError(t *testing.T) {
	_, err := NewNormalizer().Normalize(nil)
	if err == nil {
		t.Fatal("expected error for nil document")
	}
}

func TestNormalizerNonSpecDocumentReturnsSource(t *testing.T) {
	normalized, err := NewNormalizer().Normalize("plain string")
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}
	if normalized.Values["source"] != "plain string" {
		t.Fatalf("expected source to be plain string, got %v", normalized.Values["source"])
	}
}

func TestVersion(t *testing.T) {
	if v := Version(); v != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", v)
	}
}

func TestNormalizeRaw(t *testing.T) {
	t.Run("basic spec", func(t *testing.T) {
		data := map[string]any{
			"project": "myapp",
			"modules": []any{
				map[string]any{"name": "auth", "path": "./auth"},
			},
			"services": []any{
				map[string]any{"name": "api", "kind": "http", "port": 8080},
			},
		}
		spec, err := NormalizeRaw(data)
		if err != nil {
			t.Fatal(err)
		}
		if spec.Values["project"] != "myapp" {
			t.Errorf("expected myapp, got %v", spec.Values["project"])
		}
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := NormalizeRaw(nil)
		if err == nil {
			t.Error("expected error for nil input")
		}
	})

	t.Run("architecture section", func(t *testing.T) {
		data := map[string]any{
			"project":      "test",
			"architecture": map[string]any{"pattern": "hexagonal"},
		}
		spec, err := NormalizeRaw(data)
		if err != nil {
			t.Fatal(err)
		}
		arch := spec.Values["architecture"].(map[string]any)
		if arch["pattern"] != "hexagonal" {
			t.Errorf("expected hexagonal, got %v", arch["pattern"])
		}
	})

	t.Run("invalid modules type", func(t *testing.T) {
		data := map[string]any{
			"project": "test",
			"modules": "not-an-array",
		}
		_, err := NormalizeRaw(data)
		if err == nil {
			t.Error("expected error for invalid modules type")
		}
	})

	t.Run("invalid services type", func(t *testing.T) {
		data := map[string]any{
			"project":  "test",
			"services": "not-an-array",
		}
		_, err := NormalizeRaw(data)
		if err == nil {
			t.Error("expected error for invalid services type")
		}
	})
}

func TestFlattenUnflattenRoundtrip(t *testing.T) {
	t.Run("nested map roundtrip", func(t *testing.T) {
		original := map[string]any{
			"project": "myapp",
			"architecture": map[string]any{
				"pattern":    "hexagonal",
				"principles": []string{"solid"},
			},
		}
		flat := Flatten(original)
		if flat["project"] != "myapp" {
			t.Errorf("expected myapp, got %v", flat["project"])
		}
		if flat["architecture.pattern"] != "hexagonal" {
			t.Errorf("expected hexagonal, got %v", flat["architecture.pattern"])
		}
		unflat := Unflatten(flat)
		if unflat["project"] != "myapp" {
			t.Errorf("expected myapp, got %v", unflat["project"])
		}
		arch := unflat["architecture"].(map[string]any)
		if arch["pattern"] != "hexagonal" {
			t.Errorf("expected hexagonal, got %v", arch["pattern"])
		}
	})

	t.Run("empty map", func(t *testing.T) {
		flat := Flatten(map[string]any{})
		if len(flat) != 0 {
			t.Errorf("expected empty, got %v", flat)
		}
		unflat := Unflatten(map[string]any{})
		if len(unflat) != 0 {
			t.Errorf("expected empty, got %v", unflat)
		}
	})

	t.Run("unflatten overwrites non-map value", func(t *testing.T) {
		flat := map[string]any{"a.b": "c", "a": "d"}
		result := Unflatten(flat)
		if result["a"] == nil {
			t.Error("expected a to be set")
		}
	})
}

func TestMergeNormalized(t *testing.T) {
	t.Run("merge two specs", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{"project": "a", "version": "1"}}
		b := &NormalizedSpec{Values: map[string]any{"project": "b", "lang": "go"}}
		result := MergeNormalized(a, b)
		if result.Values["project"] != "b" {
			t.Errorf("expected b, got %v", result.Values["project"])
		}
		if result.Values["version"] != "1" {
			t.Errorf("expected 1, got %v", result.Values["version"])
		}
		if result.Values["lang"] != "go" {
			t.Errorf("expected go, got %v", result.Values["lang"])
		}
	})

	t.Run("both nil", func(t *testing.T) {
		result := MergeNormalized(nil, nil)
		if result == nil || result.Values == nil {
			t.Error("expected empty spec, not nil")
		}
	})

	t.Run("a nil", func(t *testing.T) {
		b := &NormalizedSpec{Values: map[string]any{"project": "b"}}
		result := MergeNormalized(nil, b)
		if result.Values["project"] != "b" {
			t.Errorf("expected b, got %v", result.Values["project"])
		}
	})

	t.Run("b nil", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{"project": "a"}}
		result := MergeNormalized(a, nil)
		if result.Values["project"] != "a" {
			t.Errorf("expected a, got %v", result.Values["project"])
		}
	})

	t.Run("deep merge nested maps", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{
			"arch": map[string]any{"pattern": "hex", "version": "1"},
		}}
		b := &NormalizedSpec{Values: map[string]any{
			"arch": map[string]any{"pattern": "layered", "lang": "go"},
		}}
		result := MergeNormalized(a, b)
		arch := result.Values["arch"].(map[string]any)
		if arch["pattern"] != "layered" {
			t.Errorf("expected layered, got %v", arch["pattern"])
		}
		if arch["version"] != "1" {
			t.Errorf("expected 1 from a, got %v", arch["version"])
		}
		if arch["lang"] != "go" {
			t.Errorf("expected go from b, got %v", arch["lang"])
		}
	})
}

func TestDiffNormalized(t *testing.T) {
	t.Run("identical specs", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{"project": "myapp"}}
		b := &NormalizedSpec{Values: map[string]any{"project": "myapp"}}
		diffs := DiffNormalized(a, b)
		if len(diffs) != 0 {
			t.Errorf("expected no diffs, got %d", len(diffs))
		}
	})

	t.Run("added field", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{}}
		b := &NormalizedSpec{Values: map[string]any{"project": "myapp"}}
		diffs := DiffNormalized(a, b)
		if len(diffs) != 1 || diffs[0].Type != "added" {
			t.Errorf("expected 1 added diff, got %d %v", len(diffs), diffs)
		}
	})

	t.Run("removed field", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{"project": "myapp"}}
		b := &NormalizedSpec{Values: map[string]any{}}
		diffs := DiffNormalized(a, b)
		if len(diffs) != 1 || diffs[0].Type != "removed" {
			t.Errorf("expected 1 removed diff, got %d %v", len(diffs), diffs)
		}
	})

	t.Run("changed field", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{"project": "old"}}
		b := &NormalizedSpec{Values: map[string]any{"project": "new"}}
		diffs := DiffNormalized(a, b)
		if len(diffs) != 1 || diffs[0].Type != "changed" {
			t.Errorf("expected 1 changed diff, got %d %v", len(diffs), diffs)
		}
	})

	t.Run("both nil", func(t *testing.T) {
		diffs := DiffNormalized(nil, nil)
		if diffs != nil {
			t.Errorf("expected nil, got %v", diffs)
		}
	})

	t.Run("a nil treated as empty", func(t *testing.T) {
		b := &NormalizedSpec{Values: map[string]any{"project": "b"}}
		diffs := DiffNormalized(nil, b)
		if len(diffs) != 1 || diffs[0].Type != "added" {
			t.Errorf("expected 1 added, got %d", len(diffs))
		}
	})

	t.Run("b nil treated as empty", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{"project": "a"}}
		diffs := DiffNormalized(a, nil)
		if len(diffs) != 1 || diffs[0].Type != "removed" {
			t.Errorf("expected 1 removed, got %d", len(diffs))
		}
	})

	t.Run("nested diff", func(t *testing.T) {
		a := &NormalizedSpec{Values: map[string]any{
			"arch": map[string]any{"pattern": "hex"},
		}}
		b := &NormalizedSpec{Values: map[string]any{
			"arch": map[string]any{"pattern": "layered", "lang": "go"},
		}}
		diffs := DiffNormalized(a, b)
		if len(diffs) != 2 {
			t.Errorf("expected 2 diffs (changed + added), got %d", len(diffs))
		}
	})
}

func TestInferTypes(t *testing.T) {
	t.Run("detects types correctly", func(t *testing.T) {
		values := map[string]any{
			"name":   "myapp",
			"port":   8080,
			"secure": true,
			"cost":   9.99,
			"tags":   []any{"go", "api"},
			"meta":   map[string]any{"key": "val"},
			"empty":  nil,
		}
		result := InferTypes(values)
		check := func(key, expected string) {
			m := result[key].(map[string]any)
			if m["_type"] != expected {
				t.Errorf("%s: expected %s, got %v", key, expected, m["_type"])
			}
		}
		check("name", "string")
		check("port", "int")
		check("secure", "bool")
		check("cost", "float")
		check("empty", "null")
	})

	t.Run("array type inference", func(t *testing.T) {
		values := map[string]any{
			"items": []any{1, 2, 3},
		}
		result := InferTypes(values)
		arr := result["items"].(map[string]any)
		if arr["_type"] != "array" {
			t.Errorf("expected array, got %v", arr["_type"])
		}
		if arr["_count"] != 3 {
			t.Errorf("expected count 3, got %v", arr["_count"])
		}
	})

	t.Run("empty array", func(t *testing.T) {
		values := map[string]any{
			"items": []any{},
		}
		result := InferTypes(values)
		arr := result["items"].(map[string]any)
		if arr["_items"] != "empty" {
			t.Errorf("expected empty items, got %v", arr["_items"])
		}
	})
}

func TestExtractSchema(t *testing.T) {
	t.Run("nil spec returns empty", func(t *testing.T) {
		schema := ExtractSchema(nil)
		if len(schema) != 0 {
			t.Errorf("expected empty, got %v", schema)
		}
	})

	t.Run("extracts flat types", func(t *testing.T) {
		spec := &NormalizedSpec{Values: map[string]any{
			"project": "myapp",
			"version": 2,
		}}
		schema := ExtractSchema(spec)
		if schema["project"] != "string" {
			t.Errorf("expected string, got %s", schema["project"])
		}
		if schema["version"] != "int" {
			t.Errorf("expected int, got %s", schema["version"])
		}
	})

	t.Run("extracts nested types", func(t *testing.T) {
		spec := &NormalizedSpec{Values: map[string]any{
			"arch": map[string]any{"pattern": "hex"},
		}}
		schema := ExtractSchema(spec)
		if schema["arch.pattern"] != "string" {
			t.Errorf("expected string, got %s", schema["arch.pattern"])
		}
	})
}
