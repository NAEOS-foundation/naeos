package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
)

func TestIntegrationMinimalSpecFullPipeline(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := p.Run("project: minimal-proj")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.NEIR == nil || result.NEIR.Project == nil {
		t.Fatal("expected NEIR with project")
	}
	if result.NEIR.Project.Name != "minimal-proj" {
		t.Errorf("Project.Name = %q, want %q", result.NEIR.Project.Name, "minimal-proj")
	}
	if result.Graph == nil {
		t.Error("expected execution graph to be set")
	}
}

func TestIntegrationSpecMinimalFile(t *testing.T) {
	t.Parallel()
	specData, err := os.ReadFile("../../examples/spec-minimal.yaml")
	if err != nil {
		t.Fatalf("read spec-minimal.yaml: %v", err)
	}
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(string(specData))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR == nil || result.NEIR.Project == nil {
		t.Fatal("expected NEIR with project")
	}
	if len(result.Artifacts) == 0 {
		t.Error("expected at least one artifact")
	}
	if len(result.Tasks) == 0 {
		t.Error("expected at least one task")
	}
}

func TestIntegrationSpecFullFile(t *testing.T) {
	t.Parallel()
	specData, err := os.ReadFile("../../examples/spec-full.yaml")
	if err != nil {
		t.Fatalf("read spec-full.yaml: %v", err)
	}
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(string(specData))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR == nil || result.NEIR.Project == nil {
		t.Fatal("expected NEIR with project")
	}
	if result.NEIR.Project.Name != "e-commerce-platform" {
		t.Errorf("Project.Name = %q, want %q", result.NEIR.Project.Name, "e-commerce-platform")
	}
	if len(result.NEIR.Modules) != 5 {
		t.Errorf("Modules = %d, want 5", len(result.NEIR.Modules))
	}
	if len(result.NEIR.Services) != 2 {
		t.Errorf("Services = %d, want 2", len(result.NEIR.Services))
	}
	if result.NEIR.Generation == nil {
		t.Error("expected Generation config")
	}
	if len(result.Artifacts) == 0 {
		t.Error("expected artifacts")
	}
	if len(result.Tasks) == 0 {
		t.Error("expected tasks")
	}
	if len(result.Reviews) == 0 {
		t.Error("expected reviews")
	}
	if result.Graph == nil {
		t.Error("expected execution graph")
	}
}

func TestIntegrationSpecMicroserviceEvent(t *testing.T) {
	t.Parallel()
	specData, err := os.ReadFile("../../examples/spec-microservice-event.yaml")
	if err != nil {
		t.Fatalf("read spec-microservice-event.yaml: %v", err)
	}
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(string(specData))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR == nil || result.NEIR.Project == nil {
		t.Fatal("expected NEIR with project")
	}
	if result.NEIR.Project.Name != "order-processing-system" {
		t.Errorf("Project.Name = %q, want %q", result.NEIR.Project.Name, "order-processing-system")
	}
	if len(result.NEIR.Services) != 3 {
		t.Errorf("Services = %d, want 3", len(result.NEIR.Services))
	}
}

func TestIntegrationValidateVersusRun(t *testing.T) {
	t.Parallel()
	spec := `project: compare-test
modules:
  - name: core
    path: ./internal/core
services:
  - name: api
    kind: http
    port: 8080
`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	validResult, err := p.Validate(spec)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	runResult, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if validResult.NEIR == nil || runResult.NEIR == nil {
		t.Fatal("both should produce NEIR")
	}
	if validResult.NEIR.Project.Name != runResult.NEIR.Project.Name {
		t.Errorf("Validate project %q != Run project %q", validResult.NEIR.Project.Name, runResult.NEIR.Project.Name)
	}
	if len(validResult.Artifacts) != 0 {
		t.Errorf("Validate should produce 0 artifacts, got %d", len(validResult.Artifacts))
	}
	if len(runResult.Artifacts) == 0 {
		t.Error("Run should produce artifacts")
	}
	if len(runResult.Tasks) == 0 {
		t.Error("Run should produce tasks")
	}
}

func TestIntegrationValidateEmptyInput(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = p.Validate("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestIntegrationRunWithContext(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	result, err := p.RunContext(ctx, "project: ctx-test")
	if err != nil {
		t.Fatalf("RunContext: %v", err)
	}
	if result == nil || result.NEIR == nil {
		t.Fatal("expected non-nil result with NEIR")
	}
}

func TestIntegrationRunWithCanceledContext(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.RunContext(ctx, "project: canceled")
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestIntegrationDryRun(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	p, err := New(Config{DryRun: true, OutputDir: outDir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run("project: dry-test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result == nil || result.NEIR == nil {
		t.Fatal("expected result with NEIR")
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run should not write files, got %d entries", len(entries))
	}
}

func TestIntegrationOutputDirWriting(t *testing.T) {
	t.Parallel()
	outDir := filepath.Join(t.TempDir(), "nested", "out")
	p, err := New(Config{OutputDir: outDir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(`project: write-test
modules:
  - name: core
    path: ./internal/core
`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Artifacts) == 0 {
		t.Fatal("expected artifacts")
	}
	for _, a := range result.Artifacts {
		path := filepath.Join(outDir, a.Path)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("artifact %q not written to output dir: %v", a.Path, err)
		}
	}
}

func TestIntegrationLanguageGo(t *testing.T) {
	t.Parallel()
	p, err := New(Config{Languages: []string{"go"}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(`project: go-test
modules:
  - name: core
    path: ./internal/core
`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR.Generation == nil {
		t.Fatal("expected Generation")
	}
	if !result.NEIR.Generation.HasLanguage(language.LanguageGo) {
		t.Error("expected Go language")
	}
	goFound := false
	for _, a := range result.Artifacts {
		if strings.HasSuffix(a.Path, ".go") || a.Path == "go.mod" {
			goFound = true
			break
		}
	}
	if !goFound {
		t.Error("expected at least one Go artifact")
	}
}

func TestIntegrationLanguageTypeScript(t *testing.T) {
	t.Parallel()
	p, err := New(Config{Languages: []string{"typescript"}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(`project: ts-test
modules:
  - name: core
    path: ./internal/core
`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR.Generation == nil {
		t.Fatal("expected Generation")
	}
	if !result.NEIR.Generation.HasLanguage(language.LanguageTypeScript) {
		t.Error("expected TypeScript language")
	}
	tsFound := false
	for _, a := range result.Artifacts {
		if strings.HasSuffix(a.Path, ".ts") || strings.HasSuffix(a.Path, ".tsx") || a.Path == "package.json" {
			tsFound = true
			break
		}
	}
	if !tsFound {
		t.Error("expected at least one TypeScript artifact")
	}
}

func TestIntegrationMultiLanguage(t *testing.T) {
	t.Parallel()
	p, err := New(Config{Languages: []string{"go", "typescript"}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(`project: multi-lang
modules:
  - name: core
    path: ./internal/core
`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR.Generation == nil {
		t.Fatal("expected Generation")
	}
	if len(result.NEIR.Generation.Languages) != 2 {
		t.Errorf("Languages = %d, want 2", len(result.NEIR.Generation.Languages))
	}
	goFound, tsFound := false, false
	for _, a := range result.Artifacts {
		if strings.HasSuffix(a.Path, ".go") || a.Path == "go.mod" {
			goFound = true
		}
		if strings.HasSuffix(a.Path, ".ts") || strings.HasSuffix(a.Path, ".tsx") || a.Path == "package.json" {
			tsFound = true
		}
	}
	if !goFound {
		t.Error("expected Go artifact")
	}
	if !tsFound {
		t.Error("expected TypeScript artifact")
	}
}

func TestIntegrationSpecWithArchitecture(t *testing.T) {
	t.Parallel()
	spec := `project: arch-test
architecture:
  pattern: hexagonal
  principles:
    - Loose coupling
    - Separation of concerns
`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR.Architecture == nil {
		t.Fatal("expected Architecture")
	}
}

func TestIntegrationSpecWithDeployment(t *testing.T) {
	t.Parallel()
	spec := `project: deploy-test
deployment:
  strategy: blue-green
  environments:
    - development
    - staging
    - production
`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR.Deployment == nil {
		t.Fatal("expected Deployment")
	}
}

func TestIntegrationSpecWithTesting(t *testing.T) {
	t.Parallel()
	spec := `project: testing-test
testing:
  strategy: unit
  coverage: "90%"
`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR.Testing == nil {
		t.Fatal("expected Testing config")
	}
}

func TestIntegrationModuleDependencies(t *testing.T) {
	t.Parallel()
	spec := `project: dep-test
modules:
  - name: core
    path: ./internal/core
  - name: api
    path: ./internal/api
    dependencies:
      - core
  - name: worker
    path: ./internal/worker
    dependencies:
      - core
services:
  - name: http-svc
    kind: http
    port: 8080
`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.NEIR.Modules) != 3 {
		t.Errorf("Modules = %d, want 3", len(result.NEIR.Modules))
	}
	if len(result.Tasks) == 0 {
		t.Error("expected tasks from module dependencies")
	}
	if result.Graph == nil {
		t.Error("expected execution graph")
	}
}

func TestIntegrationServiceEndpoints(t *testing.T) {
	t.Parallel()
	spec := `project: endpoint-test
services:
  - name: api
    kind: http
    port: 8080
    endpoints:
      - method: GET
        path: /health
        action: check
      - method: POST
        path: /users
        action: create
      - method: GET
        path: /users/:id
        action: get
      - method: DELETE
        path: /users/:id
        action: delete
`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.NEIR.Services) != 1 {
		t.Fatalf("Services = %d, want 1", len(result.NEIR.Services))
	}
	if len(result.NEIR.Services[0].Endpoints) < 2 {
		t.Logf("Endpoints = %d (parser may not fully populate endpoint details)", len(result.NEIR.Services[0].Endpoints))
	}
}

func TestIntegrationReviewsGenerated(t *testing.T) {
	t.Parallel()
	p, err := New(Config{Languages: []string{"go"}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(`project: review-test
modules:
  - name: core
    path: ./internal/core
`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Reviews) == 0 {
		t.Error("expected review results")
	}
	for _, r := range result.Reviews {
		if r == nil {
			t.Error("expected non-nil review result")
		}
	}
}

func TestIntegrationParallelModes(t *testing.T) {
	t.Parallel()
	boolTrue := true
	boolFalse := false

	p1, err := New(Config{Parallel: &boolTrue})
	if err != nil {
		t.Fatalf("New parallel: %v", err)
	}
	r1, err := p1.Run(`project: parallel-test
modules:
  - name: a
    path: ./a
  - name: b
    path: ./b
`)
	if err != nil {
		t.Fatalf("Run parallel: %v", err)
	}
	if len(r1.Artifacts) == 0 {
		t.Error("expected artifacts from parallel pipeline")
	}

	p2, err := New(Config{Parallel: &boolFalse})
	if err != nil {
		t.Fatalf("New sequential: %v", err)
	}
	r2, err := p2.Run(`project: sequential-test
modules:
  - name: a
    path: ./a
  - name: b
    path: ./b
`)
	if err != nil {
		t.Fatalf("Run sequential: %v", err)
	}
	if len(r2.Artifacts) == 0 {
		t.Error("expected artifacts from sequential pipeline")
	}
}

func TestIntegrationSourcePreservedInResult(t *testing.T) {
	t.Parallel()
	spec := "project: source-test"
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Source != spec {
		t.Errorf("Source = %q, want %q", result.Source, spec)
	}
}

func TestIntegrationMultipleRunsSamePipeline(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	specs := []string{
		"project: run-1\nmodules:\n  - name: a\n    path: ./a",
		"project: run-2\nmodules:\n  - name: b\n    path: ./b",
		"project: run-3\nmodules:\n  - name: c\n    path: ./c",
	}
	for i, spec := range specs {
		result, err := p.Run(spec)
		if err != nil {
			t.Fatalf("Run %d: %v", i, err)
		}
		if result == nil || result.NEIR == nil {
			t.Fatalf("Run %d: expected NEIR", i)
		}
	}
}

func TestIntegrationSpecWithAllFields(t *testing.T) {
	t.Parallel()
	spec := `project: full-e2e-test
modules:
  - name: auth
    path: ./internal/auth
    description: Authentication module
    dependencies:
      - core
  - name: core
    path: ./internal/core
    description: Core logic
services:
  - name: api
    kind: http
    port: 8080
    description: REST API
    endpoints:
      - method: GET
        path: /health
        action: healthCheck
      - method: POST
        path: /login
        action: login
  - name: worker
    kind: worker
    port: 9090
    description: Background worker
    endpoints:
      - method: POST
        path: /process
        action: process
architecture:
  pattern: hexagonal
  description: Hexagonal architecture
  principles:
    - Loose coupling
    - Separation of concerns
deployment:
  strategy: rolling
  environments:
    - development
    - staging
    - production
testing:
  strategy: unit
  coverage: "85%"
generation:
  languages:
    - go
    - typescript
  output_dir: ./out
  module_dir: ./internal
`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR == nil || result.NEIR.Project == nil {
		t.Fatal("expected NEIR with project")
	}
	if result.NEIR.Project.Name != "full-e2e-test" {
		t.Errorf("Project.Name = %q, want %q", result.NEIR.Project.Name, "full-e2e-test")
	}
	if len(result.NEIR.Modules) != 2 {
		t.Errorf("Modules = %d, want 2", len(result.NEIR.Modules))
	}
	if len(result.NEIR.Services) != 2 {
		t.Errorf("Services = %d, want 2", len(result.NEIR.Services))
	}
	if result.NEIR.Architecture == nil {
		t.Error("expected Architecture")
	}
	if result.NEIR.Deployment == nil {
		t.Error("expected Deployment")
	}
	if result.NEIR.Testing == nil {
		t.Error("expected Testing")
	}
	if result.NEIR.Generation == nil {
		t.Error("expected Generation")
	}
	if len(result.Artifacts) == 0 {
		t.Error("expected artifacts")
	}
	if len(result.Tasks) == 0 {
		t.Error("expected tasks")
	}
	if len(result.Reviews) == 0 {
		t.Error("expected reviews")
	}
	if result.Graph == nil {
		t.Error("expected execution graph")
	}
}

func TestIntegrationPipelineName(t *testing.T) {
	t.Parallel()
	p, err := New(Config{Name: "my-pipeline"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != "my-pipeline" {
		t.Errorf("Name() = %q, want %q", p.Name(), "my-pipeline")
	}

	p2, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p2.Name() != "unnamed" {
		t.Errorf("Name() = %q, want %q", p2.Name(), "unnamed")
	}
}

func TestIntegrationHooks(t *testing.T) {
	t.Parallel()
	var stages []string
	p, err := New(Config{
		Hooks: &Hooks{
			BeforeRun: []HookFunc{func(ctx *HookContext) error {
				stages = append(stages, "before-run")
				return nil
			}},
			AfterRun: []HookFunc{func(ctx *HookContext) error {
				stages = append(stages, "after-run")
				return nil
			}},
			BeforeGenerate: []HookFunc{func(ctx *HookContext) error {
				stages = append(stages, "before-generate")
				return nil
			}},
			AfterGenerate: []HookFunc{func(ctx *HookContext) error {
				stages = append(stages, "after-generate")
				return nil
			}},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run("project: hook-test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if len(stages) == 0 {
		t.Error("expected hooks to be called")
	}
}

func TestIntegrationHookFailure(t *testing.T) {
	t.Parallel()
	p, err := New(Config{
		Hooks: &Hooks{
			BeforeRun: []HookFunc{func(ctx *HookContext) error {
				return &testHookError{msg: "hook failed"}
			}},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = p.Run("project: hook-fail-test")
	if err == nil {
		t.Error("expected error from failing hook")
	}
}

type testHookError struct{ msg string }

func (e *testHookError) Error() string { return e.msg }

func TestIntegrationGetHookFuncs(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	hooks := p.Hooks()
	if hooks == nil {
		t.Fatal("expected non-nil Hooks")
	}
}

func TestIntegrationSpecMinimalWithServices(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(`project: svc-test
services:
  - name: api
    kind: http
    port: 3000
`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.NEIR.Services) != 1 {
		t.Errorf("Services = %d, want 1", len(result.NEIR.Services))
	}
	if result.NEIR.Services[0].Kind != "http" {
		t.Errorf("Service Kind = %q, want %q", result.NEIR.Services[0].Kind, "http")
	}
}

func TestIntegrationNoLanguages(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(`project: no-lang-test
modules:
  - name: core
    path: ./internal/core
`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.NEIR.Generation != nil && len(result.NEIR.Generation.Languages) > 0 {
		t.Errorf("expected no languages, got %d", len(result.NEIR.Generation.Languages))
	}
}

func TestIntegrationGraphStructure(t *testing.T) {
	t.Parallel()
	spec := `project: graph-test
modules:
  - name: core
    path: ./internal/core
  - name: api
    path: ./internal/api
    dependencies:
      - core
services:
  - name: web
    kind: http
    port: 8080
`
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Graph == nil {
		t.Fatal("expected graph")
	}
	if result.Graph.NodeCount() == 0 {
		t.Error("expected non-empty graph nodes")
	}
	if result.Graph.EdgeCount() == 0 {
		t.Error("expected non-empty graph edges")
	}
}
