package compiler

import (
	"fmt"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/component"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
	"github.com/NAEOS-foundation/naeos/internal/promptlib"
)

func TestNewWithLibrary(t *testing.T) {
	t.Parallel()
	lib, err := promptlib.New()
	if err != nil {
		t.Fatalf("failed to create library: %v", err)
	}
	c := NewWithLibrary(lib)
	if c == nil {
		t.Fatal("expected non-nil compiler")
	}
	if c.Library() != lib {
		t.Error("expected library to match")
	}
}

func TestNewWithLibraryNil(t *testing.T) {
	t.Parallel()
	c := NewWithLibrary(nil)
	if c == nil {
		t.Fatal("expected non-nil compiler")
	}
	if c.Library() != nil {
		t.Error("expected nil library")
	}
}

func TestCompilerLibraryDefault(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Library() != nil {
		t.Error("expected nil library for default compiler")
	}
}

func TestCompilerTargetsSorted(t *testing.T) {
	t.Parallel()
	c := New()
	c.Register(&stubAdapter{target: TargetOpenCode})
	c.Register(&stubAdapter{target: TargetCopilot})
	c.Register(&stubAdapter{target: TargetClaude})

	targets := c.Targets()
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	if targets[0] != TargetClaude || targets[1] != TargetCopilot || targets[2] != TargetOpenCode {
		t.Errorf("targets not sorted: %v", targets)
	}
}

func TestCompilerCompileAllWithError(t *testing.T) {
	t.Parallel()
	c := New()
	c.Register(&stubAdapter{target: TargetCopilot})
	c.Register(&errAdapter{target: TargetClaude})

	results := c.CompileAll(&model.NEIR{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[TargetCopilot].Summary == "" {
		t.Error("expected non-empty summary for copilot")
	}
	if results[TargetClaude].Summary == "" {
		t.Error("expected non-empty summary for claude error")
	}
}

func TestBuildProjectContextFull(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{
			Name:        "full-proj",
			Description: "Full test project",
			Version:     "2.0.0",
		},
		Architecture: &architecture.Architecture{
			Pattern:    "clean",
			Principles: []string{"KISS", "YAGNI"},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./core", Description: "Core logic"},
			{Name: "api", Path: "./api", Description: "API layer", Dependencies: []string{"core"}},
		},
		Services: []service.Service{
			{
				Name: "http-svc",
				Kind: service.KindHTTP,
				Port: 8080,
				Endpoints: []service.Endpoint{
					{Method: "GET", Path: "/health", Action: "check"},
					{Method: "POST", Path: "/api/v1/users", Action: "createUser"},
				},
			},
		},
		Components: []component.Component{
			{Name: "handler", Kind: component.KindHandler, Module: "api"},
			{Name: "repo", Kind: component.KindRepository, Module: "core"},
		},
		Security: &security.Security{
			Authentication: &security.Authentication{
				Method:   "jwt",
				Provider: "auth0",
			},
			Authorization: &security.Authorization{
				Model: "rbac",
				Roles: []string{"admin", "user"},
			},
		},
		Deployment: &deployment.Deployment{
			Strategy: deployment.StrategyRolling,
		},
		Testing: &testingmodel.Testing{
			Strategy: testingmodel.StrategyUnit,
			Coverage: &testingmodel.Coverage{
				MinPercent: 80.0,
			},
		},
	}

	ctx := buildProjectContext(neir)
	checks := []string{"full-proj", "2.0.0", "clean", "KISS", "core", "api", "http-svc", "GET", "handler", "repo", "jwt", "auth0", "rbac", "admin", "rolling", "unit", "80%"}
	for _, check := range checks {
		if !containsStr(ctx, check) {
			t.Errorf("context missing %q", check)
		}
	}
}

func TestBuildProjectContextEmpty(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{}
	ctx := buildProjectContext(neir)
	if ctx != "" {
		t.Errorf("expected empty context for empty NEIR, got %q", ctx)
	}
}

func TestResolveLanguagesWithGeneration(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguagePython, language.LanguageTypeScript},
		},
	}
	langs := resolveLanguages(neir)
	if len(langs) != 2 {
		t.Errorf("expected 2 languages, got %d", len(langs))
	}
	if langs[0] != language.LanguagePython {
		t.Errorf("expected python, got %s", langs[0])
	}
}

func TestResolveLanguagesEmptyGeneration(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Generation: &generation.GenerationConfig{},
	}
	langs := resolveLanguages(neir)
	if len(langs) != 1 {
		t.Errorf("expected 1 default language, got %d", len(langs))
	}
	if langs[0] != language.LanguageGo {
		t.Errorf("expected go, got %s", langs[0])
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type errAdapter struct {
	target Target
}

func (a *errAdapter) Target() Target { return a.target }
func (a *errAdapter) Compile(neir *model.NEIR) (*CompiledOutput, error) {
	return nil, fmt.Errorf("compile failed for %s", a.target)
}
