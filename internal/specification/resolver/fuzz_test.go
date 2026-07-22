package resolver

import (
	"testing"
)

func FuzzResolveEnvironmentVariables(f *testing.F) {
	f.Add("hello ${NAME}", "world")
	f.Add("${A}-${B}", "test")
	f.Add("no vars here", "")
	f.Add("${_UNDERSCORE_123}", "val")
	f.Add("$$nested ${X}", "y")

	f.Fuzz(func(t *testing.T, key, value string) {
		context := map[string]any{
			"test_key": key,
			"other":    value,
		}
		result := ResolveEnvironmentVariables(context)
		if result == nil {
			t.Fatal("result should not be nil")
		}
		if len(result) != 2 {
			t.Errorf("expected 2 entries, got %d", len(result))
		}
	})
}

func FuzzValidateSpec(f *testing.F) {
	f.Add("api", "http", 8080)
	f.Add("worker", "grpc", 9090)
	f.Add("svc", "rest", 0)
	f.Add("svc", "rest", 99999)
	f.Add("", "", -1)

	f.Fuzz(func(t *testing.T, name, kind string, port int) {
		spec := &ResolvedSpec{
			Context: map[string]any{
				"modules": []map[string]any{
					{"name": name, "path": "./" + name},
				},
				"services": []map[string]any{
					{"name": name, "kind": kind, "port": port},
				},
			},
		}
		result := ValidateSpec(spec)
		for _, e := range result.Errors {
			if e.Severity != SeverityError && e.Severity != SeverityWarning {
				t.Errorf("invalid severity: %s", e.Severity)
			}
			if e.Field == "" {
				t.Error("field should not be empty")
			}
			if e.Message == "" {
				t.Error("message should not be empty")
			}
		}
	})
}

func FuzzResolveWithTrace(f *testing.F) {
	f.Add("test-module", "http", "8080")
	f.Add("auth", "grpc", "9090")

	f.Fuzz(func(t *testing.T, moduleName, svcKind, portStr string) {
		spec := map[string]any{
			"modules": []map[string]any{
				{"name": moduleName, "path": "./" + moduleName},
			},
			"services": []map[string]any{
				{"name": moduleName, "kind": svcKind, "port": portStr},
			},
		}
		resolved, ctx, err := ResolveWithTrace(spec)
		if err != nil {
			return
		}
		if resolved == nil {
			t.Fatal("resolved should not be nil")
		}
		if ctx == nil {
			t.Fatal("context should not be nil")
		}
		if len(ctx.Steps) == 0 {
			t.Error("expected at least one step")
		}
	})
}

func FuzzConflictDetection(f *testing.F) {
	f.Add("api", "./api", "worker", "./api", 8080, 8080)
	f.Add("a", "./a", "b", "./b", 3000, 4000)
	f.Add("x", "", "y", "", 0, 0)

	f.Fuzz(func(t *testing.T, name1, path1, name2, path2 string, port1, port2 int) {
		context := map[string]any{
			"modules": []map[string]any{
				{"name": name1, "path": path1},
				{"name": name2, "path": path2},
			},
			"services": []map[string]any{
				{"name": name1, "port": port1},
				{"name": name2, "port": port2},
			},
			"architecture": map[string]any{
				"pattern":    "microservices",
				"principles": []any{"monolith"},
			},
		}
		conflicts := ConflictDetection(context)
		for _, c := range conflicts {
			if c.Message == "" {
				t.Error("conflict message should not be empty")
			}
			if c.Path == "" {
				t.Error("conflict path should not be empty")
			}
		}
	})
}

func FuzzCrossValidateServices(f *testing.F) {
	f.Add("api", "core")
	f.Add("worker", "auth")
	f.Add("svc", "nonexistent")

	f.Fuzz(func(t *testing.T, svcName, moduleName string) {
		context := map[string]any{
			"modules": []map[string]any{
				{"name": "core", "path": "./core"},
			},
			"services": []map[string]any{
				{
					"name":   svcName,
					"module": moduleName,
					"endpoints": []map[string]any{
						{"method": "GET", "path": "/test"},
					},
				},
			},
		}
		errs := CrossValidateServices(context)
		for _, e := range errs {
			if e.Severity != SeverityError && e.Severity != SeverityWarning {
				t.Errorf("invalid severity: %s", e.Severity)
			}
		}
	})
}
