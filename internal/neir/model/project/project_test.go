package project

import "testing"

func TestProject_ZeroValue(t *testing.T) {
	var p Project
	if p.Name != "" {
		t.Error("expected empty Name")
	}
	if p.Authors != nil {
		t.Error("expected nil Authors")
	}
	if p.Attributes != nil {
		t.Error("expected nil Attributes")
	}
}

func TestProject_Full(t *testing.T) {
	p := Project{
		Name:        "naeos",
		Version:     "2.1.0",
		Description: "Platform engineering system",
		License:     "MIT",
		Authors:     []string{"team"},
		Repository:  "github.com/NAEOS-foundation/naeos",
		Tags:        []string{"platform", "engineering"},
		Attributes:  map[string]string{"key": "val"},
	}
	if p.Name != "naeos" {
		t.Errorf("expected naeos, got %s", p.Name)
	}
	if p.Version != "2.1.0" {
		t.Errorf("expected 2.1.0, got %s", p.Version)
	}
	if p.License != "MIT" {
		t.Errorf("expected MIT, got %s", p.License)
	}
	if len(p.Authors) != 1 {
		t.Errorf("expected 1 author, got %d", len(p.Authors))
	}
	if len(p.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(p.Tags))
	}
}
