package parser

import (
	"testing"
)

func TestResolveRef(t *testing.T) {
	r := NewVariableResolver()
	r.SetRef("service.url", "https://example.com")
	result, err := r.Resolve("api: $ref{service.url}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "api: https://example.com" {
		t.Errorf("expected 'api: https://example.com', got %q", result)
	}
}

func TestResolveRefMissing(t *testing.T) {
	r := NewVariableResolver()
	result, err := r.Resolve("api: $ref{missing.ref}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "api: $ref{missing.ref}" {
		t.Errorf("expected unchanged, got %q", result)
	}
}

func TestSetVars(t *testing.T) {
	r := NewVariableResolver()
	r.SetVars(map[string]string{"host": "localhost", "port": "8080"})
	result, err := r.Resolve("${host}:${port}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "localhost:8080" {
		t.Errorf("expected 'localhost:8080', got %q", result)
	}
}

func TestResolveAdjacentPatterns(t *testing.T) {
	r := NewVariableResolver()
	r.SetVar("a", "x")
	r.SetVar("b", "y")
	result, err := r.Resolve("${a}${b}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "xy" {
		t.Errorf("expected 'xy', got %q", result)
	}
}

func TestResolveEmptyInput(t *testing.T) {
	r := NewVariableResolver()
	result, err := r.Resolve("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestResolveMapNested(t *testing.T) {
	r := NewVariableResolver()
	r.SetVar("name", "test")
	r.SetVar("port", "3000")

	data := map[string]any{
		"service": map[string]any{
			"name": "${name}",
			"port": "${port}",
		},
		"tags": []any{"${name}", "static"},
	}
	resolved, err := r.ResolveMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc := resolved["service"].(map[string]any)
	if svc["name"] != "test" {
		t.Errorf("expected test, got %v", svc["name"])
	}
	tags := resolved["tags"].([]any)
	if tags[0] != "test" {
		t.Errorf("expected test, got %v", tags[0])
	}
}

func TestResolveMapPassthrough(t *testing.T) {
	r := NewVariableResolver()
	data := map[string]any{
		"count": 42,
		"ratio": 3.14,
		"flag":  true,
		"nested": map[string]any{
			"items": []any{1, 2, 3},
		},
	}
	resolved, err := r.ResolveMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved["count"] != 42 {
		t.Errorf("expected 42, got %v", resolved["count"])
	}
}

func TestSpecValidatorValidateNoIssues(t *testing.T) {
	v := NewSpecValidator()
	result := v.Validate(map[string]any{
		"name": "test",
	})
	if !result.Valid {
		t.Errorf("expected valid, got %d issues", len(result.Issues))
	}
}

func TestSpecValidatorValidateRefNotFound(t *testing.T) {
	v := NewSpecValidator()
	result := v.Validate(map[string]any{
		"endpoint": "$ref{missing.endpoint}",
	})
	if result.Valid {
		t.Error("expected invalid for missing ref")
	}
	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}
	if result.Issues[0].Rule != "ref-not-found" {
		t.Errorf("expected ref-not-found, got %s", result.Issues[0].Rule)
	}
}

func TestSpecValidatorValidateRefExists(t *testing.T) {
	resolver := NewVariableResolver()
	resolver.SetRef("existing.ref", "value")
	v := &SpecValidator{resolver: resolver}

	result := v.Validate(map[string]any{
		"endpoint": "$ref{existing.ref}",
	})
	if !result.Valid {
		t.Errorf("expected valid, got issues: %v", result.Issues)
	}
}

func TestSpecValidatorValidateNestedMap(t *testing.T) {
	v := NewSpecValidator()
	issues := 0
	_ = issues
	result := v.Validate(map[string]any{
		"config": map[string]any{
			"db":     "${db_url}",
			"secret": "$ref{db.password}",
		},
		"list": []any{
			"$ref{item1}",
			"simple",
		},
	})
	if result.Valid {
		t.Error("expected invalid")
	}
}

func TestSpecValidatorValidateNonMap(t *testing.T) {
	v := NewSpecValidator()
	result := v.Validate("just a string")
	if !result.Valid {
		t.Errorf("expected valid for non-map data")
	}
}

func TestValidateModulesEmptyName(t *testing.T) {
	v := NewSpecValidator()
	issues := v.ValidateModules([]Module{
		{Name: "", Dependencies: nil},
	})
	if len(issues) == 0 {
		t.Error("expected issue for empty name")
	}
}

func TestValidateServicesEmptyName(t *testing.T) {
	v := NewSpecValidator()
	issues := v.ValidateServices([]Service{
		{Name: "", Port: 80},
	})
	if len(issues) == 0 {
		t.Error("expected issue for empty name")
	}
}

func TestDetectCyclesSelfCycle(t *testing.T) {
	cycles := detectCycles(map[string][]string{
		"a": {"a"},
	})
	if len(cycles) == 0 {
		t.Error("expected self-cycle detection")
	}
}

func TestValidateServicesNegativePort(t *testing.T) {
	v := NewSpecValidator()
	issues := v.ValidateServices([]Service{
		{Name: "svc", Port: -1},
	})
	if len(issues) == 0 {
		t.Error("expected issue for negative port")
	}
}
