package contextbundle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func TestGenerateFromNEIR(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test-project"},
		Modules: []module.Module{
			{Name: "auth", Path: "./auth", Dependencies: []string{"core"}},
			{Name: "api", Path: "./api", Dependencies: []string{"auth"}},
		},
		Services: []service.Service{
			{Name: "gateway", Kind: service.KindHTTP, Port: 8080},
		},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{"go", "typescript"},
		},
	}

	gen := NewGenerator(nil)
	bundle := gen.GenerateFromNEIR(neir)

	if bundle.Project != "test-project" {
		t.Errorf("project = %q, want test-project", bundle.Project)
	}
	if len(bundle.Modules) != 2 {
		t.Errorf("modules = %d, want 2", len(bundle.Modules))
	}
	if bundle.Modules[0].Name != "auth" {
		t.Errorf("modules[0].name = %q, want auth", bundle.Modules[0].Name)
	}
	if bundle.Modules[0].Dependencies[0] != "core" {
		t.Errorf("modules[0].deps[0] = %q, want core", bundle.Modules[0].Dependencies[0])
	}
	if len(bundle.Services) != 1 {
		t.Errorf("services = %d, want 1", len(bundle.Services))
	}
	if bundle.Services[0].Port != 8080 {
		t.Errorf("services[0].port = %d, want 8080", bundle.Services[0].Port)
	}
	if len(bundle.Languages) != 2 {
		t.Errorf("languages = %d, want 2", len(bundle.Languages))
	}
	if bundle.Summary == "" {
		t.Error("summary should not be empty")
	}
	if bundle.Metadata["module_count"] != "2" {
		t.Errorf("metadata module_count = %q, want 2", bundle.Metadata["module_count"])
	}
}

func TestGenerateFromSpec(t *testing.T) {
	doc := &parser.SpecDocument{
		Project: "my-app",
		Modules: []parser.Module{
			{Name: "web", Path: "./web", Description: "web frontend"},
		},
		Services: []parser.Service{
			{Name: "api-server", Kind: "rest", Port: 3000, Endpoints: []parser.Endpoint{
				{Method: "GET", Path: "/users", Action: "listUsers"},
			}},
		},
		Generation: &parser.Generation{Languages: []string{"go"}},
	}

	gen := NewGenerator(nil)
	bundle := gen.GenerateFromSpec(doc)

	if bundle.Project != "my-app" {
		t.Errorf("project = %q, want my-app", bundle.Project)
	}
	if bundle.Modules[0].Description != "web frontend" {
		t.Errorf("description = %q, want web frontend", bundle.Modules[0].Description)
	}
	if bundle.Services[0].Endpoints[0].Action != "listUsers" {
		t.Errorf("endpoint action = %q, want listUsers", bundle.Services[0].Endpoints[0].Action)
	}
}

func TestToMarkdown(t *testing.T) {
	bundle := &Bundle{
		Project:   "test",
		Modules:   []ModuleContext{{Name: "auth", Path: "./auth", Dependencies: []string{"core"}}},
		Services:  []ServiceContext{{Name: "api", Kind: "rest", Port: 8080}},
		Languages: []string{"go"},
		Summary:   "Project: test",
	}

	md := bundle.ToMarkdown()
	if !strings.Contains(md, "# test") {
		t.Error("markdown should contain project title")
	}
	if !strings.Contains(md, "## Modules") {
		t.Error("markdown should contain modules section")
	}
	if !strings.Contains(md, "auth") {
		t.Error("markdown should contain module name")
	}
	if !strings.Contains(md, "core") {
		t.Error("markdown should contain dependency")
	}
}

func TestToPlainText(t *testing.T) {
	bundle := &Bundle{
		Project:  "test",
		Modules:  []ModuleContext{{Name: "auth", Path: "./auth"}},
		Services: []ServiceContext{{Name: "api", Kind: "rest", Port: 8080}},
	}

	plain := bundle.ToPlainText()
	if !strings.Contains(plain, "Project: test") {
		t.Error("plain text should contain project")
	}
	if !strings.Contains(plain, "Module: auth") {
		t.Error("plain text should contain module")
	}
}

func TestSupportedTargets(t *testing.T) {
	bundle := &Bundle{
		Modules: []ModuleContext{{Name: "x"}},
	}

	targets := bundle.SupportedTargets()
	found := false
	for _, tgt := range targets {
		if tgt == "markdown" {
			found = true
		}
	}
	if !found {
		t.Error("markdown should be a supported target")
	}
}

func TestBundleMetadata(t *testing.T) {
	bundle := &Bundle{
		Project:  "meta-test",
		Metadata: make(map[string]string),
		Modules:  []ModuleContext{{Name: "a"}, {Name: "b"}},
		Services: []ServiceContext{{Name: "s1"}},
	}
	bundle.Metadata["module_count"] = fmt.Sprintf("%d", len(bundle.Modules))
	bundle.Metadata["service_count"] = fmt.Sprintf("%d", len(bundle.Services))

	if bundle.Metadata["module_count"] != "2" {
		t.Errorf("module_count = %q, want 2", bundle.Metadata["module_count"])
	}
	if bundle.Metadata["service_count"] != "1" {
		t.Errorf("service_count = %q, want 1", bundle.Metadata["service_count"])
	}
}

func TestGenerateFromNEIREmpty(t *testing.T) {
	neir := &model.NEIR{}

	gen := NewGenerator(nil)
	bundle := gen.GenerateFromNEIR(neir)

	if bundle.Project != "" {
		t.Errorf("project should be empty, got %q", bundle.Project)
	}
	if len(bundle.Modules) != 0 {
		t.Errorf("modules should be empty, got %d", len(bundle.Modules))
	}
	_ = bundle.Summary
}
