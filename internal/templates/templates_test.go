// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager("")
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestListBuiltins(t *testing.T) {
	mgr := NewManager("")
	tmpls, err := mgr.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tmpls) < 5 {
		t.Errorf("expected at least 5 built-in templates, got %d", len(tmpls))
	}
}

func TestGetBuiltinTemplate(t *testing.T) {
	mgr := NewManager("")
	tmpl, err := mgr.Get("readme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl == nil {
		t.Fatal("expected non-nil template")
	}
}

func TestRenderBuiltinTemplate(t *testing.T) {
	mgr := NewManager("")
	result, err := mgr.Render("readme", map[string]string{
		"ProjectName": "test-project",
		"Description": "A test project",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestGetNonexistentTemplate(t *testing.T) {
	mgr := NewManager("")
	_, err := mgr.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
}

func TestAddCustomTemplate(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	err := mgr.AddCustom("custom", "Hello {{.Name}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpls, _ := mgr.List()
	found := false
	for _, tmpl := range tmpls {
		if tmpl.Name == "custom" && tmpl.IsCustom {
			found = true
		}
	}
	if !found {
		t.Error("expected custom template in list")
	}
}

func TestRemoveCustomTemplate(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	mgr.AddCustom("custom", "Hello {{.Name}}")
	err := mgr.RemoveCustom("custom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveNonexistentTemplate(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	err := mgr.RemoveCustom("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
}
