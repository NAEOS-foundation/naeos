package module

import "testing"

func TestModule_ZeroValue(t *testing.T) {
	var m Module
	if m.Name != "" {
		t.Error("expected empty Name")
	}
	if m.Path != "" {
		t.Error("expected empty Path")
	}
	if m.Packages != nil {
		t.Error("expected nil Packages")
	}
}

func TestModule_Full(t *testing.T) {
	m := Module{
		Name:         "core",
		Path:         "./core",
		Description:  "Core module",
		Packages:     []string{"domain", "application"},
		Dependencies: []string{"shared"},
		Attributes:   map[string]string{"language": "go"},
	}
	if m.Name != "core" {
		t.Errorf("expected core, got %s", m.Name)
	}
	if m.Path != "./core" {
		t.Errorf("expected ./core, got %s", m.Path)
	}
	if len(m.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(m.Packages))
	}
	if len(m.Dependencies) != 1 || m.Dependencies[0] != "shared" {
		t.Errorf("expected [shared], got %v", m.Dependencies)
	}
	if m.Attributes["language"] != "go" {
		t.Errorf("expected go, got %s", m.Attributes["language"])
	}
}
