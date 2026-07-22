package resolver

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
)

func TestResolverBuildsContextFromNormalizedSpec(t *testing.T) {
	norm := &normalizer.NormalizedSpec{Values: map[string]any{
		"project":  "acme-api",
		"modules":  []map[string]any{{"name": "auth", "path": "./internal/auth"}},
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

func TestResolverResolvesModuleDependencies(t *testing.T) {
	norm := &normalizer.NormalizedSpec{Values: map[string]any{
		"project": "acme-api",
		"modules": []map[string]any{
			{"name": "auth", "path": "./internal/auth"},
			{"name": "user", "path": "./internal/user", "dependencies": []any{"auth", "nonexistent"}},
		},
	}}

	resolver := NewResolver()
	resolved, err := resolver.Resolve(norm)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	modules := resolved.Context["modules"].([]map[string]any)
	userMod := modules[1]
	deps := userMod["dependencies"].([]any)
	if len(deps) != 1 {
		t.Fatalf("expected 1 valid dependency, got %d", len(deps))
	}
	if deps[0] != "auth" {
		t.Fatalf("expected dependency 'auth', got %v", deps[0])
	}
}

func TestResolverNormalizesEndpoints(t *testing.T) {
	norm := &normalizer.NormalizedSpec{Values: map[string]any{
		"project": "acme-api",
		"services": []map[string]any{
			{
				"name": "api",
				"kind": "http",
				"port": 8080,
				"endpoints": []map[string]any{
					{"method": "GET", "path": "users"},
				},
			},
		},
	}}

	resolver := NewResolver()
	resolved, err := resolver.Resolve(norm)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	services := resolved.Context["services"].([]map[string]any)
	eps := services[0]["endpoints"].([]map[string]any)
	if eps[0]["path"] != "/users" {
		t.Fatalf("expected path '/users', got %v", eps[0]["path"])
	}
}

func TestResolverPopulatesDefaults(t *testing.T) {
	norm := &normalizer.NormalizedSpec{Values: map[string]any{
		"project": "acme-api",
		"modules": []map[string]any{
			{"name": "auth"},
		},
	}}

	resolver := NewResolver()
	resolved, err := resolver.Resolve(norm)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if resolved.Context["architecture"] == nil {
		t.Fatal("expected default architecture to be populated")
	}

	modules := resolved.Context["modules"].([]map[string]any)
	if modules[0]["path"] != "./internal/auth" {
		t.Fatalf("expected default path './internal/auth', got %v", modules[0]["path"])
	}
}

func TestResolverNilSpec(t *testing.T) {
	_, err := NewResolver().Resolve(nil)
	if err == nil {
		t.Error("expected error for nil spec")
	}
}

func TestResolverNonNormalizedSpec(t *testing.T) {
	resolved, err := NewResolver().Resolve("plain string")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Context["resolved"] != true {
		t.Error("expected resolved=true for non-normalized spec")
	}
}

func TestResolveWithTrace(t *testing.T) {
	t.Run("trace includes steps", func(t *testing.T) {
		norm := &normalizer.NormalizedSpec{Values: map[string]any{
			"project":  "myapp",
			"modules":  []map[string]any{{"name": "auth", "path": "./auth"}},
			"services": []map[string]any{{"name": "api", "kind": "http"}},
		}}
		resolved, trace, err := ResolveWithTrace(norm)
		if err != nil {
			t.Fatal(err)
		}
		if resolved == nil {
			t.Fatal("expected resolved spec")
		}
		if len(trace.Steps) == 0 {
			t.Error("expected at least one trace step")
		}
	})

	t.Run("nil spec returns error", func(t *testing.T) {
		_, trace, err := ResolveWithTrace(nil)
		if err == nil {
			t.Error("expected error for nil spec")
		}
		if trace == nil {
			t.Error("expected non-nil trace even on error")
		}
	})

	t.Run("non-normalized spec", func(t *testing.T) {
		_, trace, err := ResolveWithTrace("raw")
		if err != nil {
			t.Fatal(err)
		}
		if len(trace.Steps) != 1 {
			t.Errorf("expected 1 step, got %d", len(trace.Steps))
		}
	})

	t.Run("warnings for missing port", func(t *testing.T) {
		norm := &normalizer.NormalizedSpec{Values: map[string]any{
			"project":  "myapp",
			"services": []map[string]any{{"name": "api", "kind": "http"}},
		}}
		_, trace, err := ResolveWithTrace(norm)
		if err != nil {
			t.Fatal(err)
		}
		if len(trace.Warnings) == 0 {
			t.Error("expected warning for missing port")
		}
	})
}

func TestResolveEnvironmentVariables(t *testing.T) {
	SetEnvForTest("DB_HOST", "localhost")
	SetEnvForTest("DB_PORT", "5432")
	defer ClearEnvForTest()

	context := map[string]any{
		"host":  "${DB_HOST}",
		"port":  "${DB_PORT}",
		"fixed": "hello",
	}
	resolved := ResolveEnvironmentVariables(context)
	if resolved["host"] != "localhost" {
		t.Errorf("expected localhost, got %v", resolved["host"])
	}
	if resolved["port"] != "5432" {
		t.Errorf("expected 5432, got %v", resolved["port"])
	}
	if resolved["fixed"] != "hello" {
		t.Errorf("expected hello, got %v", resolved["fixed"])
	}
}

func TestResolveEnvironmentVariablesNoMatch(t *testing.T) {
	ClearEnvForTest()
	context := map[string]any{
		"key": "${MISSING_VAR}",
	}
	resolved := ResolveEnvironmentVariables(context)
	if resolved["key"] != "${MISSING_VAR}" {
		t.Errorf("expected unchanged ${MISSING_VAR}, got %v", resolved["key"])
	}
}

func TestResolveReferences(t *testing.T) {
	t.Run("resolves ref to root field", func(t *testing.T) {
		context := map[string]any{
			"project": "myapp",
			"title":   "${ref:project.name}",
		}
		resolved := ResolveReferences(context)
		if resolved["title"] != "${ref:project.name}" {
			t.Errorf("expected unresolved ref, got %v", resolved["title"])
		}
	})

	t.Run("no ref returns same", func(t *testing.T) {
		context := map[string]any{
			"project": "myapp",
		}
		resolved := ResolveReferences(context)
		if resolved["project"] != "myapp" {
			t.Errorf("expected myapp, got %v", resolved["project"])
		}
	})
}

func TestValidateSpec(t *testing.T) {
	t.Run("nil spec", func(t *testing.T) {
		result := ValidateSpec(nil)
		if result.Valid {
			t.Error("expected invalid for nil spec")
		}
	})

	t.Run("valid spec", func(t *testing.T) {
		spec := &ResolvedSpec{Context: map[string]any{
			"project": "myapp",
			"modules": []map[string]any{
				{"name": "auth", "path": "./auth", "dependencies": []any{}},
			},
			"architecture": map[string]any{"pattern": "layered"},
		}}
		result := ValidateSpec(spec)
		if !result.Valid {
			t.Errorf("expected valid, got errors: %v", result.Errors)
		}
	})

	t.Run("detects duplicate modules", func(t *testing.T) {
		spec := &ResolvedSpec{Context: map[string]any{
			"modules": []map[string]any{
				{"name": "auth"},
				{"name": "auth"},
			},
		}}
		result := ValidateSpec(spec)
		if result.Valid {
			t.Error("expected invalid for duplicate modules")
		}
	})

	t.Run("detects invalid port", func(t *testing.T) {
		spec := &ResolvedSpec{Context: map[string]any{
			"services": []map[string]any{
				{"name": "api", "port": 99999},
			},
		}}
		result := ValidateSpec(spec)
		if result.Valid {
			t.Error("expected invalid for out-of-range port")
		}
	})
}

func TestCrossValidateServices(t *testing.T) {
	t.Run("no errors for valid services", func(t *testing.T) {
		errs := CrossValidateServices(map[string]any{
			"modules": []map[string]any{
				{"name": "auth"},
			},
			"services": []map[string]any{
				{"name": "api", "module": "auth"},
			},
		})
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})

	t.Run("detects non-existent module reference", func(t *testing.T) {
		errs := CrossValidateServices(map[string]any{
			"modules": []map[string]any{
				{"name": "auth"},
			},
			"services": []map[string]any{
				{"name": "api", "module": "nonexistent"},
			},
		})
		if len(errs) == 0 {
			t.Error("expected error for non-existent module reference")
		}
	})

	t.Run("no services returns empty", func(t *testing.T) {
		errs := CrossValidateServices(map[string]any{"project": "x"})
		if errs != nil {
			t.Errorf("expected nil, got %v", errs)
		}
	})
}

func TestConflictDetection(t *testing.T) {
	t.Run("detects port conflicts", func(t *testing.T) {
		conflicts := ConflictDetection(map[string]any{
			"services": []map[string]any{
				{"name": "api", "port": 8080},
				{"name": "web", "port": 8080},
			},
		})
		if len(conflicts) == 0 {
			t.Error("expected port conflict")
		}
	})

	t.Run("no conflicts for distinct ports", func(t *testing.T) {
		conflicts := ConflictDetection(map[string]any{
			"services": []map[string]any{
				{"name": "api", "port": 8080},
				{"name": "web", "port": 9090},
			},
		})
		if len(conflicts) != 0 {
			t.Errorf("expected no conflicts, got %v", conflicts)
		}
	})

	t.Run("architecture pattern conflict", func(t *testing.T) {
		conflicts := ConflictDetection(map[string]any{
			"architecture": map[string]any{
				"pattern":    "microservices",
				"principles": []any{"monolith"},
			},
		})
		if len(conflicts) == 0 {
			t.Error("expected architecture conflict")
		}
	})
}

func TestValidationErrorSeverity(t *testing.T) {
	e := ValidationError{Field: "test", Message: "msg", Severity: SeverityError}
	if e.Severity != SeverityError {
		t.Errorf("expected error severity, got %s", e.Severity)
	}
	w := ValidationError{Field: "test", Message: "msg", Severity: SeverityWarning}
	if w.Severity != SeverityWarning {
		t.Errorf("expected warning severity, got %s", w.Severity)
	}
}

func TestResolutionContext(t *testing.T) {
	rc := &ResolutionContext{}
	rc.AddStep("step1", "first step")
	rc.AddWarning("beware")
	if len(rc.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(rc.Steps))
	}
	if len(rc.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(rc.Warnings))
	}
}
