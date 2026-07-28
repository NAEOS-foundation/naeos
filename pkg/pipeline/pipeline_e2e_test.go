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

func TestEndToEndFullSpec(t *testing.T) {
	spec := `
project: full-spec
version: "0.3.0"
modules:
  - name: auth
    path: ./internal/auth
  - name: api
    path: ./internal/api
    dependencies: [auth]
services:
  - name: gateway
    kind: http
    port: 8080
    endpoints:
      - method: GET
        path: /api/v1/users
architecture:
  pattern: hexagonal
  principles:
    - loose-coupling
    - high-cohesion
security:
  audit_logging: true
  encryption: tls
testing:
  strategy: unit
  coverage: "80"
generation:
  languages:
    - go
`
	p, err := New(Config{
		Name:      "e2e-full",
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

	if result.NEIR.Project.Name != "full-spec" {
		t.Errorf("Project.Name = %q, want %q", result.NEIR.Project.Name, "full-spec")
	}
	if len(result.NEIR.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(result.NEIR.Modules))
	}
	if len(result.NEIR.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(result.NEIR.Services))
	}
	if len(result.Artifacts) == 0 {
		t.Error("expected artifacts to be generated")
	}
}

func TestEndToEndModuleDependencies(t *testing.T) {
	spec := `
project: dep-test
modules:
  - name: core
    path: ./core
  - name: auth
    path: ./auth
    dependencies: [core]
  - name: api
    path: ./api
    dependencies: [auth, core]
`
	p, err := New(Config{
		Name:      "e2e-dep",
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

	if len(result.NEIR.Modules) != 3 {
		t.Fatalf("expected 3 modules, got %d", len(result.NEIR.Modules))
	}

	modMap := make(map[string]int)
	for i, m := range result.NEIR.Modules {
		modMap[m.Name] = i
	}

	names := make([]string, len(result.NEIR.Modules))
	for i, m := range result.NEIR.Modules {
		names[i] = m.Name
	}
	expected := []string{"core", "auth", "api"}
	for _, name := range expected {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected module %q not found in %v", name, names)
		}
	}
}

func TestEndToEndServices(t *testing.T) {
	spec := `
project: svc-test
services:
  - name: web
    kind: http
    port: 80
  - name: admin
    kind: http
    port: 8080
  - name: worker
    kind: grpc
    port: 50051
`
	p, err := New(Config{
		Name:      "e2e-svc",
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

	if len(result.NEIR.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(result.NEIR.Services))
	}
}

func TestEndToEndSpecNotFound(t *testing.T) {
	spec := `project: missing-modules`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run should not error on minimal spec: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

func TestEndToEndEmptySpec(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = p.Run("")
	if err == nil {
		t.Fatal("expected error for empty spec")
	}
}
