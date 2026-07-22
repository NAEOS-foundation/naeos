package templates

import (
	"testing"
)

func TestGetCachedTemplate(t *testing.T) {
	m := NewManager("")
	tmpl1, _ := m.Get("readme")
	tmpl2, _ := m.Get("readme")
	if tmpl1 != tmpl2 {
		t.Error("expected same cached instance on second Get")
	}
}

func TestAddCustomNoDir(t *testing.T) {
	m := NewManager("")
	err := m.AddCustom("x", "content")
	if err == nil {
		t.Error("expected error when no templates directory is configured")
	}
}

func TestRenderError(t *testing.T) {
	m := NewManager("")
	_, err := m.Render("nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestGetCustomTemplate(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	err := m.AddCustom("custom", "Hello {{.Name}}")
	if err != nil {
		t.Fatalf("add custom: %v", err)
	}
	tmpl, err := m.Get("custom")
	if err != nil {
		t.Fatalf("get custom: %v", err)
	}
	if tmpl == nil {
		t.Error("expected non-nil template")
	}
}

func TestGetCustomTemplateCached(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	m.AddCustom("c", "content")
	m.Get("c")
	tmpl, err := m.Get("c")
	if err != nil {
		t.Fatalf("second get: %v", err)
	}
	if tmpl == nil {
		t.Error("expected non-nil template on cache hit")
	}
}
