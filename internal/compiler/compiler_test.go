package compiler

import (
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/testutil"
)

type stubAdapter struct {
	target Target
}

func (a *stubAdapter) Target() Target { return a.target }
func (a *stubAdapter) Compile(neir *model.NEIR) (*CompiledOutput, error) {
	return &CompiledOutput{
		Target:     a.target,
		Files:      []OutputFile{{Path: "test.md", Content: "hello", Kind: "instructions"}},
		Summary:    "stub output",
		CompiledAt: time.Now(),
	}, nil
}

func TestCompilerRegisterAndGet(t *testing.T) {
	c := New()
	c.Register(&stubAdapter{target: TargetCopilot})

	out, err := c.Compile(&model.NEIR{}, TargetCopilot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Target != TargetCopilot {
		t.Errorf("expected target copilot, got %s", out.Target)
	}
	if len(out.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(out.Files))
	}
}

func TestCompilerUnknownTarget(t *testing.T) {
	c := New()
	_, err := c.Compile(&model.NEIR{}, "unknown")
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestCompilerTargets(t *testing.T) {
	c := New()
	c.Register(&stubAdapter{target: TargetCopilot})
	c.Register(&stubAdapter{target: TargetClaude})
	c.Register(&stubAdapter{target: TargetCursor})

	targets := c.Targets()
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
}

func TestCompilerCompileAll(t *testing.T) {
	c := New()
	c.Register(&stubAdapter{target: TargetCopilot})
	c.Register(&stubAdapter{target: TargetClaude})

	results := c.CompileAll(&model.NEIR{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[TargetCopilot] == nil {
		t.Error("expected copilot result")
	}
	if results[TargetClaude] == nil {
		t.Error("expected claude result")
	}
}

func TestBuildProjectContext(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{
			Name:        "test-proj",
			Description: "A test project",
			Version:     "1.0.0",
		},
		Architecture: &architecture.Architecture{
			Pattern:    "hexagonal",
			Principles: []string{"DI", "SRP"},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./core", Description: "Core module"},
		},
		Services: []service.Service{
			{Name: "api", Kind: "http", Port: 8080},
		},
	}

	ctx := buildProjectContext(neir)
	if ctx == "" {
		t.Error("expected non-empty context")
	}
	if !testutil.Contains(ctx, "test-proj") {
		t.Error("expected project name in context")
	}
	if !testutil.Contains(ctx, "hexagonal") {
		t.Error("expected architecture in context")
	}
}

func TestResolveLanguages(t *testing.T) {
	neir := &model.NEIR{}
	langs := resolveLanguages(neir)
	if len(langs) != 1 {
		t.Errorf("expected 1 default language, got %d", len(langs))
	}
}
