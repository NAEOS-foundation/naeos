// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package docgen

import (
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

func TestNewDocGenerator(t *testing.T) {
	g := NewDocGenerator()
	if g == nil {
		t.Error("expected non-nil generator")
	}
}

func TestGenerateFromSpec(t *testing.T) {
	g := NewDocGenerator()
	doc := &parser.SpecDocument{
		Project: "test-project",
		Modules: []parser.Module{
			{Name: "core", Path: "./core", Description: "Core module"},
		},
		Services: []parser.Service{
			{
				Name: "api",
				Kind: "http",
				Port: 8080,
				Endpoints: []parser.Endpoint{
					{Method: "GET", Path: "/health", Action: "check"},
				},
			},
		},
		Architecture: &parser.Architecture{
			Pattern:    "hexagonal",
			Principles: []string{"separation of concerns"},
		},
		Deployment: &parser.Deployment{
			Strategy:     "blue-green",
			Environments: []string{"staging", "production"},
		},
		Testing: &parser.Testing{
			Strategy: "unit",
			Coverage: "80",
		},
		Generation: &parser.Generation{
			Languages: []string{"go", "typescript"},
			OutputDir: "./generated",
		},
	}

	result := g.GenerateFromSpec(doc)

	if !strings.Contains(result, "# test-project") {
		t.Error("expected project name in output")
	}
	if !strings.Contains(result, "## Modules") {
		t.Error("expected Modules section")
	}
	if !strings.Contains(result, "## Services") {
		t.Error("expected Services section")
	}
	if !strings.Contains(result, "## Architecture") {
		t.Error("expected Architecture section")
	}
	if !strings.Contains(result, "hexagonal") {
		t.Error("expected architecture pattern")
	}
	if !strings.Contains(result, "## Deployment") {
		t.Error("expected Deployment section")
	}
	if !strings.Contains(result, "## Testing") {
		t.Error("expected Testing section")
	}
	if !strings.Contains(result, "## Generation") {
		t.Error("expected Generation section")
	}
}

func TestGenerateFromSpecMinimal(t *testing.T) {
	g := NewDocGenerator()
	doc := &parser.SpecDocument{
		Project: "minimal",
	}

	result := g.GenerateFromSpec(doc)

	if !strings.Contains(result, "# minimal") {
		t.Error("expected project name")
	}
	if strings.Contains(result, "## Modules") {
		t.Error("should not contain empty Modules section")
	}
}

func TestGenerateFromNEIR(t *testing.T) {
	g := NewDocGenerator()
	neir := &model.NEIR{
		Project: &project.Project{
			Name:        "neir-project",
			Description: "A test project",
			Version:     "1.0.0",
		},
		Modules: []module.Module{
			{Name: "auth", Path: "./auth", Description: "Auth module"},
		},
		Services: []service.Service{
			{
				Name: "api-svc",
				Kind: service.KindHTTP,
				Port: 3000,
				Endpoints: []service.Endpoint{
					{Method: "POST", Path: "/login", Action: "authenticate"},
				},
			},
		},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguageGo},
		},
	}

	result := g.GenerateFromNEIR(neir)

	if !strings.Contains(result, "# neir-project") {
		t.Error("expected project name")
	}
	if !strings.Contains(result, "A test project") {
		t.Error("expected project description")
	}
	if !strings.Contains(result, "1.0.0") {
		t.Error("expected version")
	}
	if !strings.Contains(result, "## Modules") {
		t.Error("expected Modules section")
	}
	if !strings.Contains(result, "## Services") {
		t.Error("expected Services section")
	}
}

func TestGenerateFromNEIRNil(t *testing.T) {
	g := NewDocGenerator()
	result := g.GenerateFromNEIR(&model.NEIR{})
	if result != "" {
		t.Errorf("expected empty result for empty NEIR, got: %q", result)
	}
}

func TestGenerateAPIDoc(t *testing.T) {
	g := NewDocGenerator()
	doc := &parser.SpecDocument{
		Project: "api-doc-project",
		Services: []parser.Service{
			{
				Name: "users",
				Endpoints: []parser.Endpoint{
					{Method: "GET", Path: "/users", Action: "list"},
					{Method: "POST", Path: "/users", Action: "create"},
				},
			},
		},
	}

	result := g.GenerateAPIDoc(doc)

	if !strings.Contains(result, "# api-doc-project — API Reference") {
		t.Error("expected API reference header")
	}
	if !strings.Contains(result, "## users") {
		t.Error("expected users service section")
	}
	if !strings.Contains(result, "GET /users") {
		t.Error("expected GET endpoint")
	}
	if !strings.Contains(result, "POST /users") {
		t.Error("expected POST endpoint")
	}
}

func TestGenerateModuleDocs(t *testing.T) {
	g := NewDocGenerator()
	doc := &parser.SpecDocument{
		Project: "mod-docs",
		Modules: []parser.Module{
			{
				Name:         "core",
				Path:         "./core",
				Description:  "Core logic",
				Dependencies: []string{"utils", "config"},
			},
		},
	}

	result := g.GenerateModuleDocs(doc)

	if !strings.Contains(result, "# mod-docs — Module Documentation") {
		t.Error("expected module doc header")
	}
	if !strings.Contains(result, "## core") {
		t.Error("expected core module section")
	}
	if !strings.Contains(result, "Core logic") {
		t.Error("expected module description")
	}
	if !strings.Contains(result, "Dependencies:") {
		t.Error("expected dependencies section")
	}
}

func TestSupportedLanguages(t *testing.T) {
	g := NewDocGenerator()
	langs := g.SupportedLanguages()
	if len(langs) != 5 {
		t.Errorf("expected 5 languages, got %d", len(langs))
	}

	expected := map[language.Language]bool{
		language.LanguageGo:         true,
		language.LanguageTypeScript: true,
		language.LanguagePython:     true,
		language.LanguageJava:       true,
		language.LanguageRust:       true,
	}
	for _, l := range langs {
		if !expected[l] {
			t.Errorf("unexpected language: %s", l)
		}
	}
}
