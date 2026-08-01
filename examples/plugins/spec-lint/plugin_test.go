package main

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

func TestSpecLintCleanSpec(t *testing.T) {
	p := New()
	ctx := &pluginhost.PluginContext{}
	if err := p.Initialize(ctx); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	t.Cleanup(func() { _ = p.Shutdown() })

	spec := strings.Join([]string{
		"project: demo",
		"modules:",
		"  - name: auth",
		"    path: ./auth",
		"services:",
		"  - name: gateway",
		"    port: 8080",
	}, "\n")

	result, err := p.Execute("lint", map[string]any{"spec": spec})
	if err != nil {
		t.Fatalf("Execute lint: %v", err)
	}

	r, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if r["ok"] != true {
		t.Errorf("expected ok=true, got %v (violations: %v)", r["ok"], r["violations"])
	}
}

func TestSpecLintDetectsViolations(t *testing.T) {
	p := New()

	spec := strings.Join([]string{
		"modules:",
		"  - name: Auth_Service",
		"    path: ./auth",
		"services:",
		"  - name: gateway",
		"    port: 08080",
	}, "\n")

	result, err := p.Execute("lint", map[string]any{"spec": spec})
	if err != nil {
		t.Fatalf("Execute lint: %v", err)
	}

	r := result.(map[string]any)
	if r["ok"] != false {
		t.Fatalf("expected ok=false, got %v", r["ok"])
	}

	violations, ok := r["violations"].([]violation)
	if !ok {
		t.Fatalf("expected []violation, got %T", r["violations"])
	}
	if len(violations) < 3 {
		t.Errorf("expected at least 3 violations (case, format, port), got %d: %v", len(violations), violations)
	}
}

func TestSpecLintMissingSpec(t *testing.T) {
	p := New()

	_, err := p.Execute("lint", nil)
	if err == nil {
		t.Fatal("expected error when spec param is missing")
	}
}

func TestSpecLintPing(t *testing.T) {
	p := New()

	result, err := p.Execute("ping", nil)
	if err != nil {
		t.Fatalf("Execute ping: %v", err)
	}
	if result.(map[string]string)["status"] != "ok" {
		t.Fatalf("unexpected ping result: %v", result)
	}
}

func TestSpecLintUnknownAction(t *testing.T) {
	p := New()

	_, err := p.Execute("nope", nil)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}
