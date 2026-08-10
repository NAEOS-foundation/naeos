package main

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

func newTestPlugin(t *testing.T) *Plugin {
	t.Helper()
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	ctx := &pluginhost.PluginContext{}
	if err := p.Initialize(ctx); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	t.Cleanup(func() { _ = p.Shutdown() })
	return p
}

func TestHelloPing(t *testing.T) {
	p := newTestPlugin(t)

	result, err := p.Execute("ping", nil)
	if err != nil {
		t.Fatalf("Execute ping: %v", err)
	}

	r, ok := result.(map[string]string)
	if !ok || r["status"] != "ok" {
		t.Errorf("expected {status: ok}, got %v", result)
	}
}

func TestHelloDescribe(t *testing.T) {
	p := newTestPlugin(t)

	result, err := p.Execute("describe", nil)
	if err != nil {
		t.Fatalf("Execute describe: %v", err)
	}

	r, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if r["name"] != "hello" {
		t.Errorf("expected name %q, got %v", "hello", r["name"])
	}
	if r["version"] != "0.1.0" {
		t.Errorf("expected version %q, got %v", "0.1.0", r["version"])
	}
}

func TestHelloGreet(t *testing.T) {
	p := newTestPlugin(t)

	result, err := p.Execute("greet", map[string]any{"name": "NAEOS"})
	if err != nil {
		t.Fatalf("Execute greet: %v", err)
	}

	r, ok := result.(map[string]string)
	if !ok || r["message"] != "Hello, NAEOS!" {
		t.Errorf("expected greeting, got %v", result)
	}
}

func TestHelloGreetDefault(t *testing.T) {
	p := newTestPlugin(t)

	result, err := p.Execute("greet", nil)
	if err != nil {
		t.Fatalf("Execute greet: %v", err)
	}

	r, ok := result.(map[string]string)
	if !ok || !strings.Contains(r["message"], "world") {
		t.Errorf("expected default greeting, got %v", result)
	}
}

func TestHelloUnknownAction(t *testing.T) {
	p := newTestPlugin(t)

	_, err := p.Execute("nope", nil)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}
