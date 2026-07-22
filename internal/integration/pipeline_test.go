//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	"github.com/NAEOS-foundation/naeos/internal/compiler/adapters"
	contextbundle "github.com/NAEOS-foundation/naeos/internal/context/bundle"
	"github.com/NAEOS-foundation/naeos/internal/neir/builder"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/validator"
	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
)

const testSpec = `project: testapp
version: "1.0.0"
architecture:
  pattern: hexagonal
  principles:
    - separation of concerns
    - dependency inversion
modules:
  - name: core
    path: ./internal/core
    description: Core business logic
    dependencies: []
  - name: api
    path: ./internal/api
    description: HTTP API layer
    dependencies:
      - core
services:
  - name: http-server
    kind: http
    port: 8080
    endpoints:
      - method: GET
        path: /health
        action: healthcheck
      - method: POST
        path: /api/v1/items
        action: createItem
deployment:
  strategy: blue-green
  environments:
    - staging
    - production
testing:
  strategy: unit+integration
generation:
  languages:
    - go
    - typescript
`

func parseAndBuild(t *testing.T, spec string) (*model.NEIR, *parser.SpecDocument) {
	t.Helper()
	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	n := normalizer.NewNormalizer()
	norm, err := n.Normalize(doc)
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}

	r := resolver.NewResolver()
	resolved, err := r.Resolve(norm)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	b := builder.NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("build NEIR failed: %v", err)
	}
	return neir, doc
}

func TestFullPipeline(t *testing.T) {
	neir, doc := parseAndBuild(t, testSpec)

	if doc.Project != "testapp" {
		t.Errorf("expected project 'testapp', got %s", doc.Project)
	}
	if len(doc.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(doc.Modules))
	}
	if len(doc.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(doc.Services))
	}

	if neir.Project == nil {
		t.Fatal("expected non-nil project in NEIR")
	}
	if neir.Project.Name != "testapp" {
		t.Errorf("expected project name 'testapp', got %s", neir.Project.Name)
	}
	if len(neir.Modules) != 2 {
		t.Errorf("expected 2 modules in NEIR, got %d", len(neir.Modules))
	}

	v := validator.NewValidator()
	err := v.Validate(neir)
	if err != nil {
		t.Errorf("validation failed: %v", err)
	}

	c := compiler.New()
	c.Register(adapters.NewCopilotAdapter(nil))
	c.Register(adapters.NewClaudeAdapter(nil))
	c.Register(adapters.NewCursorAdapter(nil))

	outputs := c.CompileAll(neir)
	if len(outputs) != 3 {
		t.Fatalf("expected 3 compiled outputs, got %d", len(outputs))
	}

	for target, out := range outputs {
		if out == nil {
			t.Errorf("nil output for target %s", target)
			continue
		}
		if len(out.Files) == 0 {
			t.Errorf("no files generated for target %s", target)
		}
		for _, f := range out.Files {
			if f.Content == "" {
				t.Errorf("empty content in file %s for target %s", f.Path, target)
			}
		}
	}
}

func TestSpecToNEIRToCompileRoundTrip(t *testing.T) {
	neir, _ := parseAndBuild(t, testSpec)

	c := compiler.New()
	c.Register(adapters.NewCopilotAdapter(nil))
	out, err := c.Compile(neir, compiler.TargetCopilot)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if out.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if len(out.Files) != 3 {
		t.Errorf("expected 3 copilot files, got %d", len(out.Files))
	}

	instructionsFile := out.Files[0]
	if instructionsFile.Path != ".github/copilot-instructions.md" {
		t.Errorf("unexpected path: %s", instructionsFile.Path)
	}
	if !strings.Contains(instructionsFile.Content, "hexagonal") {
		t.Error("instructions missing architecture pattern")
	}
}

func TestSpecMinimal(t *testing.T) {
	neir, _ := parseAndBuild(t, "project: min\n")

	c := compiler.New()
	c.Register(adapters.NewClaudeAdapter(nil))
	out, err := c.Compile(neir, compiler.TargetClaude)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if len(out.Files) == 0 {
		t.Error("expected files for minimal spec")
	}
}

func TestSpecToContextBundle(t *testing.T) {
	p := parser.NewParser(".")
	doc, err := p.Parse(testSpec)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	c := compiler.New()
	gen := contextbundle.NewGenerator(c)
	bundle := gen.GenerateFromSpec(doc)

	if bundle.Project != "testapp" {
		t.Errorf("expected project 'testapp', got %s", bundle.Project)
	}
	if len(bundle.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(bundle.Modules))
	}
	if len(bundle.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(bundle.Services))
	}
	if len(bundle.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(bundle.Languages))
	}

	md := bundle.ToMarkdown()
	if !strings.Contains(md, "# testapp") {
		t.Error("markdown missing project name")
	}
	if !strings.Contains(md, "## Modules") {
		t.Error("markdown missing modules section")
	}

	plain := bundle.ToPlainText()
	if !strings.Contains(plain, "Project: testapp") {
		t.Error("plain text missing project name")
	}
}
