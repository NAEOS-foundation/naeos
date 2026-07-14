//go:build integration

package pipeline

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
)

func TestEndToEndMinimalSpec(t *testing.T) {
	spec := `
project: test-api
modules:
  - name: auth
    path: ./internal/auth
services:
  - name: gateway
    kind: http
    port: 8080
`
	p, err := New(Config{
		Name:      "e2e-test",
		Mode:      "development",
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.NEIR == nil {
		t.Fatal("NEIR should not be nil")
	}
	if result.NEIR.Project.Name != "test-api" {
		t.Errorf("Project.Name = %q, want %q", result.NEIR.Project.Name, "test-api")
	}
	if len(result.NEIR.Modules) != 1 {
		t.Errorf("Modules has %d entries, want 1", len(result.NEIR.Modules))
	}
	if result.NEIR.Modules[0].Name != "auth" {
		t.Errorf("Modules[0].Name = %q, want %q", result.NEIR.Modules[0].Name, "auth")
	}
	if len(result.NEIR.Services) != 1 {
		t.Errorf("Services has %d entries, want 1", len(result.NEIR.Services))
	}
	if result.NEIR.Services[0].Name != "gateway" {
		t.Errorf("Services[0].Name = %q, want %q", result.NEIR.Services[0].Name, "gateway")
	}
	if len(result.Artifacts) == 0 {
		t.Error("Artifacts should not be empty")
	}
	if len(result.Tasks) == 0 {
		t.Error("Tasks should not be empty")
	}
}

func TestEndToEndWithLanguages(t *testing.T) {
	spec := `
project: multi-lang-api
modules:
  - name: user
    path: ./internal/user
services:
  - name: api
    kind: http
    port: 9090
generation:
  languages:
    - go
    - typescript
`
	p, err := New(Config{
		Name:      "e2e-multi-lang",
		Mode:      "development",
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.NEIR.Generation == nil {
		t.Fatal("Generation should not be nil")
	}
	if !result.NEIR.Generation.HasLanguage(language.LanguageGo) {
		t.Error("Generation should contain Go")
	}
	if !result.NEIR.Generation.HasLanguage(language.LanguageTypeScript) {
		t.Error("Generation should contain TypeScript")
	}
	if result.NEIR.Generation.HasLanguage(language.LanguagePython) {
		t.Error("Generation should not contain Python")
	}

	goCount := 0
	tsCount := 0
	for _, a := range result.Artifacts {
		if len(a.Content) > 0 {
			switch {
			case a.Path == "go.mod" || a.Path == "Dockerfile":
				goCount++
			case a.Path == "package.json" || a.Path == "tsconfig.json":
				tsCount++
			}
		}
	}
	if goCount == 0 {
		t.Error("Expected Go artifacts (go.mod)")
	}
	if tsCount == 0 {
		t.Error("Expected TypeScript artifacts (package.json)")
	}
}

func TestEndToEndValidateOnly(t *testing.T) {
	spec := `
project: validate-test
modules:
  - name: core
    path: ./internal/core
`
	p, err := New(Config{
		Name: "validate-test",
		Mode: "development",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := p.Validate(spec)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.NEIR == nil {
		t.Fatal("NEIR should not be nil after validation")
	}
	if result.NEIR.Project.Name != "validate-test" {
		t.Errorf("Project.Name = %q, want %q", result.NEIR.Project.Name, "validate-test")
	}
}
