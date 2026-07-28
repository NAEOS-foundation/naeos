package migration

import (
	"errors"
	"strings"
	"testing"
)

var errTest = errors.New("test error")

func TestParseVersion(t *testing.T) {
	v, err := ParseVersion("1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("expected 1.2.3, got %v", v)
	}
}

func TestParseVersionInvalid(t *testing.T) {
	_, err := ParseVersion("invalid")
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestVersionLess(t *testing.T) {
	v1 := Version{Major: 1, Minor: 0, Patch: 0}
	v2 := Version{Major: 2, Minor: 0, Patch: 0}
	if !v1.Less(v2) {
		t.Error("expected 1.0.0 < 2.0.0")
	}
	if v2.Less(v1) {
		t.Error("expected 2.0.0 not < 1.0.0")
	}
}

func TestVersionString(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	if v.String() != "1.2.3" {
		t.Errorf("expected 1.2.3, got %s", v.String())
	}
}

func TestPlannerPlan(t *testing.T) {
	planner := NewPlanner()
	plan, err := planner.Plan("0.1.0", "0.3.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan) < 1 {
		t.Error("expected at least 1 migration step")
	}
}

func TestPlannerPlanNoMigrationNeeded(t *testing.T) {
	planner := NewPlanner()
	_, err := planner.Plan("0.3.0", "0.1.0")
	if err == nil {
		t.Fatal("expected error for downgrade")
	}
}

func TestPlannerMigrate(t *testing.T) {
	planner := NewPlanner()
	spec := []byte("project: test\n")
	result, err := planner.Migrate(spec, "0.1.0", "0.3.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// --- New tests for improved coverage ---

func TestMigrationEngineMigrate010To030(t *testing.T) {
	engine := NewMigrationEngine()
	data := map[string]any{
		"version": "0.1.0",
		"name":    "testproject",
		"modules": []any{
			map[string]any{
				"name": "Core",
			},
		},
	}
	result, err := engine.Migrate(data, "0.1.0", "0.3.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["version"] != "0.3.0" {
		t.Errorf("expected version 0.3.0, got %v", result["version"])
	}
	// Verify module got path and dependencies from 0.1.0->0.2.0 step
	mods, ok := result["modules"].([]any)
	if !ok || len(mods) == 0 {
		t.Fatal("expected modules to be a non-empty slice")
	}
	mod := mods[0].(map[string]any)
	if _, ok := mod["path"]; !ok {
		t.Error("expected module to have path after migration")
	}
	if _, ok := mod["dependencies"]; !ok {
		t.Error("expected module to have dependencies after migration")
	}
	// Verify architecture, security, testing from 0.2.0->0.3.0 step
	if _, ok := result["architecture"]; !ok {
		t.Error("expected architecture to be set")
	}
	if _, ok := result["security"]; !ok {
		t.Error("expected security to be set")
	}
	if _, ok := result["testing"]; !ok {
		t.Error("expected testing to be set")
	}
}

func TestMigrationEngineMigrateNoPath(t *testing.T) {
	engine := NewMigrationEngine()
	data := map[string]any{"version": "9.0.0"}
	_, err := engine.Migrate(data, "9.0.0", "0.3.0")
	if err == nil {
		t.Fatal("expected error for no migration path")
	}
}

func TestMigrationEnginePlan(t *testing.T) {
	engine := NewMigrationEngine()
	plan := engine.Plan("0.1.0", "0.3.0")
	if len(plan) != 2 {
		t.Errorf("expected 2 steps, got %d", len(plan))
	}
}

func TestMigrationEngineAvailableVersions(t *testing.T) {
	engine := NewMigrationEngine()
	versions := engine.AvailableVersions()
	expected := []string{"0.1.0", "0.2.0", "0.3.0"}
	if len(versions) != len(expected) {
		t.Fatalf("expected %d versions, got %d", len(expected), len(versions))
	}
	for i, v := range versions {
		if v != expected[i] {
			t.Errorf("expected version[%d] = %s, got %s", i, expected[i], v)
		}
	}
}

func TestMigrationEngineVersionBetween(t *testing.T) {
	engine := NewMigrationEngine()
	tests := []struct {
		from, to string
		want     bool
	}{
		{"0.1.0", "0.3.0", true},
		{"0.1.0", "0.2.0", true},
		{"0.2.0", "0.3.0", true},
		{"0.3.0", "0.1.0", false},
		{"0.2.0", "0.1.0", false},
		{"0.1.0", "0.1.0", false},
		{"9.0.0", "0.3.0", false},
		{"0.1.0", "9.0.0", false},
	}
	for _, tc := range tests {
		got := engine.VersionBetween(tc.from, tc.to)
		if got != tc.want {
			t.Errorf("VersionBetween(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestFormatMigrationPlanEmpty(t *testing.T) {
	got := FormatMigrationPlan(nil)
	if got != "No migrations needed." {
		t.Errorf("expected %q, got %q", "No migrations needed.", got)
	}
}

func TestFormatMigrationPlanMultiple(t *testing.T) {
	plan := []TransformStep{
		{FromVersion: "0.1.0", ToVersion: "0.2.0", Description: "Step one"},
		{FromVersion: "0.2.0", ToVersion: "0.3.0", Description: "Step two"},
	}
	got := FormatMigrationPlan(plan)
	if !strings.Contains(got, "Migration Plan:") {
		t.Error("expected output to contain header")
	}
	if !strings.Contains(got, "0.1.0") || !strings.Contains(got, "0.2.0") || !strings.Contains(got, "0.3.0") {
		t.Error("expected output to contain version numbers")
	}
	if !strings.Contains(got, "Step one") || !strings.Contains(got, "Step two") {
		t.Error("expected output to contain step descriptions")
	}
}

func TestPlannerAddStep(t *testing.T) {
	planner := NewPlanner()
	called := false
	planner.AddStep(MigrationStep{
		FromVersion: "0.3.0",
		ToVersion:   "0.4.0",
		Description: "Custom step",
		Migrate: func(spec []byte) ([]byte, error) {
			called = true
			return spec, nil
		},
	})
	plan, err := planner.Plan("0.3.0", "0.4.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan) == 0 {
		t.Fatal("expected at least 1 step in plan")
	}
	_, err = planner.Migrate([]byte("test"), "0.3.0", "0.4.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected custom step to be called")
	}
}

func TestPlannerPlanInvalidVersion(t *testing.T) {
	planner := NewPlanner()
	_, err := planner.Plan("invalid", "0.3.0")
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
	_, err = planner.Plan("0.1.0", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid target version")
	}
}

func TestPlannerPlanSameVersion(t *testing.T) {
	planner := NewPlanner()
	_, err := planner.Plan("0.1.0", "0.1.0")
	if err == nil {
		t.Fatal("expected error when from == to")
	}
}

func TestPlannerMigrateMultiStep(t *testing.T) {
	planner := NewPlanner()
	spec := []byte("project: myapp\n")
	result, err := planner.Migrate(spec, "0.1.0", "0.3.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(result)
	if !strings.Contains(content, "generation:") {
		t.Error("expected output to contain generation section")
	}
	if !strings.Contains(content, "testing:") {
		t.Error("expected output to contain testing section")
	}
}

func TestBuiltinTransform010To020(t *testing.T) {
	engine := NewMigrationEngine()
	// Omit version key so the transform sets it to "0.2.0" by default
	data := map[string]any{
		"modules": []any{
			map[string]any{
				"name": "Auth Module",
			},
		},
	}
	result, err := engine.Migrate(data, "0.1.0", "0.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When version key is absent, the transform defaults it to 0.2.0
	if result["version"] != "0.2.0" {
		t.Errorf("expected version 0.2.0, got %v", result["version"])
	}
	mods := result["modules"].([]any)
	mod := mods[0].(map[string]any)
	if mod["path"] == nil || mod["path"] == "" {
		t.Error("expected module path to be set")
	}
	if mod["dependencies"] == nil {
		t.Error("expected module dependencies to be set")
	}
	if _, ok := result["generation"]; !ok {
		t.Error("expected generation to be set")
	}
}

func TestBuiltinTransform020To030(t *testing.T) {
	engine := NewMigrationEngine()
	data := map[string]any{
		"version": "0.2.0",
	}
	result, err := engine.Migrate(data, "0.2.0", "0.3.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["version"] != "0.3.0" {
		t.Errorf("expected version 0.3.0, got %v", result["version"])
	}
	if _, ok := result["architecture"]; !ok {
		t.Error("expected architecture to be set")
	}
	if _, ok := result["security"]; !ok {
		t.Error("expected security to be set")
	}
	if _, ok := result["testing"]; !ok {
		t.Error("expected testing to be set")
	}
}

func TestVersionLessEqual(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	if v.Less(v) {
		t.Error("expected same version to not be Less")
	}
}

func TestVersionLessPatch(t *testing.T) {
	v1 := Version{Major: 1, Minor: 2, Patch: 1}
	v2 := Version{Major: 1, Minor: 2, Patch: 3}
	if !v1.Less(v2) {
		t.Error("expected 1.2.1 < 1.2.3")
	}
	if v2.Less(v1) {
		t.Error("expected 1.2.3 not < 1.2.1")
	}
}

func TestPlannerPlanSingleStep(t *testing.T) {
	planner := NewPlanner()
	plan, err := planner.Plan("0.1.0", "0.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan) != 1 {
		t.Errorf("expected 1 step, got %d", len(plan))
	}
}

func TestPlannerMigrateError(t *testing.T) {
	planner := NewPlanner()
	planner.AddStep(MigrationStep{
		FromVersion: "0.3.0",
		ToVersion:   "0.4.0",
		Description: "Failing step",
		Migrate: func(spec []byte) ([]byte, error) {
			return nil, errTest
		},
	})
	_, err := planner.Migrate([]byte("test"), "0.3.0", "0.4.0")
	if err == nil {
		t.Fatal("expected error from failing migration step")
	}
}

func TestMigrationEngineMigrateTransformError(t *testing.T) {
	engine := NewMigrationEngine()
	engine.Register(TransformStep{
		FromVersion: "0.3.0",
		ToVersion:   "0.4.0",
		Description: "Failing transform",
		Transform: func(data map[string]any) (map[string]any, error) {
			return nil, errTest
		},
	})
	_, err := engine.Migrate(map[string]any{}, "0.3.0", "0.4.0")
	if err == nil {
		t.Fatal("expected error from failing transform step")
	}
}

func TestMigrationEnginePlanPartial(t *testing.T) {
	engine := NewMigrationEngine()
	plan := engine.Plan("0.2.0", "0.3.0")
	if len(plan) != 1 {
		t.Errorf("expected 1 step, got %d", len(plan))
	}
}

func TestMigrationEnginePlanBeyondKnown(t *testing.T) {
	engine := NewMigrationEngine()
	plan := engine.Plan("0.3.0", "9.0.0")
	if len(plan) != 0 {
		t.Errorf("expected 0 steps for unknown path, got %d", len(plan))
	}
}

func TestBuiltinTransform010To020ExistingVersion(t *testing.T) {
	engine := NewMigrationEngine()
	data := map[string]any{
		"version": "0.1.0",
	}
	result, err := engine.Migrate(data, "0.1.0", "0.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["version"] != "0.1.0" {
		t.Errorf("expected version to remain 0.1.0, got %v", result["version"])
	}
}

func TestBuiltinTransform020To030WithServices(t *testing.T) {
	engine := NewMigrationEngine()
	data := map[string]any{
		"version": "0.2.0",
		"services": []any{
			map[string]any{
				"name": "api",
			},
		},
		"architecture": map[string]any{
			"pattern": "layered",
		},
	}
	result, err := engine.Migrate(data, "0.2.0", "0.3.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svcs := result["services"].([]any)
	svc := svcs[0].(map[string]any)
	if svc["kind"] != "http" {
		t.Errorf("expected service kind 'http', got %v", svc["kind"])
	}
	if result["architecture"] != nil {
		arch := result["architecture"].(map[string]any)
		if arch["pattern"] != "layered" {
			t.Errorf("expected architecture pattern 'layered', got %v", arch["pattern"])
		}
	}
}

func TestBuiltinEngineRegisterOverride(t *testing.T) {
	engine := NewMigrationEngine()
	called := false
	engine.Register(TransformStep{
		FromVersion: "0.1.0",
		ToVersion:   "0.2.0",
		Description: "Override",
		Transform: func(data map[string]any) (map[string]any, error) {
			called = true
			data["version"] = "0.2.0"
			return data, nil
		},
	})
	_, err := engine.Migrate(map[string]any{}, "0.1.0", "0.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected overridden step to be called")
	}
}

func TestFormatMigrationPlanNil(t *testing.T) {
	got := FormatMigrationPlan([]TransformStep{})
	if got != "No migrations needed." {
		t.Errorf("expected %q, got %q", "No migrations needed.", got)
	}
}
