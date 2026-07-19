package promptlib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

func testNEIR() *model.NEIR {
	return &model.NEIR{
		Project: &project.Project{
			Name:        "test-project",
			Description: "A test project",
			Version:     "1.0.0",
		},
		Architecture: &architecture.Architecture{
			Pattern:    "hexagonal",
			Principles: []string{"DI", "SRP"},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./core", Description: "Core module", Dependencies: []string{}},
			{Name: "api", Path: "./api", Description: "API module", Dependencies: []string{"core"}},
		},
		Services: []service.Service{
			{Name: "http-api", Kind: "http", Port: 8080},
		},
	}
}

func TestNewWithDefaults(t *testing.T) {
	lib := NewWithDefaults()
	if lib == nil {
		t.Fatal("expected non-nil library")
	}

	llmPrompts := lib.ListLLMPrompts()
	if len(llmPrompts) != 3 {
		t.Errorf("expected 3 LLM prompts, got %d: %v", len(llmPrompts), llmPrompts)
	}

	compilerTpls := lib.ListCompilerTemplates()
	if len(compilerTpls) != 6 {
		t.Errorf("expected 6 compiler templates, got %d: %v", len(compilerTpls), compilerTpls)
	}
}

func TestGetLLMPrompt(t *testing.T) {
	lib := NewWithDefaults()

	p, ok := lib.GetLLMPrompt("enrich-spec")
	if !ok {
		t.Fatal("expected to find enrich-spec prompt")
	}
	if p.Name != "enrich-spec" {
		t.Errorf("expected name enrich-spec, got %s", p.Name)
	}
	if p.Kind != "llm" {
		t.Errorf("expected kind llm, got %s", p.Kind)
	}
	if p.User == "" {
		t.Error("expected non-empty user template")
	}

	_, ok = lib.GetLLMPrompt("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent prompt")
	}
}

func TestGetCompilerTemplate(t *testing.T) {
	lib := NewWithDefaults()

	tmpl, ok := lib.GetCompilerTemplate("claude")
	if !ok {
		t.Fatal("expected to find claude template")
	}
	if tmpl.Name != "claude" {
		t.Errorf("expected name claude, got %s", tmpl.Name)
	}
	if tmpl.Target != "claude" {
		t.Errorf("expected target claude, got %s", tmpl.Target)
	}
	if len(tmpl.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(tmpl.Files))
	}

	_, ok = lib.GetCompilerTemplate("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent template")
	}
}

func TestRenderLLM(t *testing.T) {
	lib := NewWithDefaults()

	result, err := lib.RenderLLM("enrich-spec", map[string]any{
		"SpecContent": "project: my-app",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User == "" {
		t.Error("expected non-empty user prompt")
	}
	if !strings.Contains(result.User, "project: my-app") {
		t.Errorf("expected rendered prompt to contain spec content, got: %s", result.User)
	}
	if result.System == "" {
		t.Error("expected non-empty system prompt")
	}
	if result.MaxTokens != 2048 {
		t.Errorf("expected max_tokens 2048, got %d", result.MaxTokens)
	}
}

func TestRenderLLMNotFound(t *testing.T) {
	lib := NewWithDefaults()

	_, err := lib.RenderLLM("nonexistent", map[string]any{})
	if err == nil {
		t.Error("expected error for nonexistent prompt")
	}
}

func TestRenderCompiler(t *testing.T) {
	lib := NewWithDefaults()

	neir := testNEIR()
	files, err := lib.RenderCompiler("claude", neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	for _, f := range files {
		if f.Content == "" {
			t.Errorf("empty content for file %s", f.Path)
		}
		if f.Path == "" {
			t.Error("empty path for file")
		}
		if f.Kind == "" {
			t.Errorf("empty kind for file %s", f.Path)
		}
	}

	if !strings.Contains(files[0].Content, "test-project") {
		t.Error("expected instructions to contain project name")
	}
	if !strings.Contains(files[0].Content, "hexagonal") {
		t.Error("expected instructions to contain architecture pattern")
	}
}

func TestRenderCompilerAllTargets(t *testing.T) {
	lib := NewWithDefaults()
	neir := testNEIR()

	targets := []string{"copilot", "claude", "cursor", "gemini", "codex", "opencode"}
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			files, err := lib.RenderCompiler(target, neir)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", target, err)
			}
			if len(files) == 0 {
				t.Errorf("expected at least 1 file for %s", target)
			}
			for _, f := range files {
				if f.Content == "" {
					t.Errorf("empty content for %s/%s", target, f.Path)
				}
			}
		})
	}
}

func TestRenderCompilerNotFound(t *testing.T) {
	lib := NewWithDefaults()

	_, err := lib.RenderCompiler("nonexistent", testNEIR())
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestRegisterLLMPrompt(t *testing.T) {
	lib := NewWithDefaults()

	customYAML := `name: custom-prompt
kind: llm
version: "1.0.0"
description: "Custom test prompt"
user: "Hello {{.Name}}"
variables:
  - name: Name
    type: string
    required: true
`
	p, err := ParseLLMPrompt([]byte(customYAML))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	lib.RegisterLLMPrompt("custom-prompt", p)

	found, ok := lib.GetLLMPrompt("custom-prompt")
	if !ok {
		t.Fatal("expected to find custom prompt")
	}
	if found.Name != "custom-prompt" {
		t.Errorf("expected name custom-prompt, got %s", found.Name)
	}
}

func TestRegisterCompilerTemplate(t *testing.T) {
	lib := NewWithDefaults()

	customYAML := `name: custom-compiler
kind: compiler
version: "1.0.0"
target: custom
files:
  - path: "CUSTOM.md"
    kind: instructions
    template: "Project: {{.Project.Name}}"
`
	ct, err := ParseCompilerTemplate([]byte(customYAML))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	lib.RegisterCompilerTemplate("custom-compiler", ct)

	found, ok := lib.GetCompilerTemplate("custom-compiler")
	if !ok {
		t.Fatal("expected to find custom template")
	}
	if found.Target != "custom" {
		t.Errorf("expected target custom, got %s", found.Target)
	}

	files, err := lib.RenderCompiler("custom-compiler", testNEIR())
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !strings.Contains(files[0].Content, "test-project") {
		t.Error("expected rendered content to contain project name")
	}
}

func TestListAll(t *testing.T) {
	lib := NewWithDefaults()

	all := lib.ListAll()
	if len(all) != 9 {
		t.Errorf("expected 9 total items (3 LLM + 6 compiler), got %d", len(all))
	}

	for _, m := range all {
		if m.Name == "" {
			t.Error("expected non-empty name")
		}
		if m.Kind != "llm" && m.Kind != "compiler" {
			t.Errorf("unexpected kind: %s", m.Kind)
		}
	}
}

func TestParseLLMPrompt(t *testing.T) {
	yaml := `name: test
kind: llm
user: "Hello"
`
	p, err := ParseLLMPrompt([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "test" {
		t.Errorf("expected name test, got %s", p.Name)
	}
	if p.Constraints == nil {
		t.Error("expected default constraints")
	} else {
		if p.Constraints.MaxTokens != 1024 {
			t.Errorf("expected default max_tokens 1024, got %d", p.Constraints.MaxTokens)
		}
		if p.Constraints.Temperature != 0.3 {
			t.Errorf("expected default temperature 0.3, got %f", p.Constraints.Temperature)
		}
	}
}

func TestParseLLMPromptMissingName(t *testing.T) {
	yaml := `kind: llm
user: "Hello"
`
	_, err := ParseLLMPrompt([]byte(yaml))
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestParseLLMPromptMissingUser(t *testing.T) {
	yaml := `name: test
kind: llm
`
	_, err := ParseLLMPrompt([]byte(yaml))
	if err == nil {
		t.Error("expected error for missing user template")
	}
}

func TestParseCompilerTemplate(t *testing.T) {
	yaml := `name: test
kind: compiler
target: test
files:
  - path: "TEST.md"
    kind: instructions
    template: "Hello"
`
	ct, err := ParseCompilerTemplate([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct.Name != "test" {
		t.Errorf("expected name test, got %s", ct.Name)
	}
	if len(ct.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(ct.Files))
	}
}

func TestParseCompilerTemplateMissingTarget(t *testing.T) {
	yaml := `name: test
kind: compiler
files:
  - path: "TEST.md"
    template: "Hello"
`
	_, err := ParseCompilerTemplate([]byte(yaml))
	if err == nil {
		t.Error("expected error for missing target")
	}
}

func TestParseCompilerTemplateEmptyFiles(t *testing.T) {
	yaml := `name: test
kind: compiler
target: test
files: []
`
	_, err := ParseCompilerTemplate([]byte(yaml))
	if err == nil {
		t.Error("expected error for empty files")
	}
}

func TestFuncMap(t *testing.T) {
	if _, ok := FuncMap["join"]; !ok {
		t.Error("expected join function")
	}
	if _, ok := FuncMap["bt"]; !ok {
		t.Error("expected bt function")
	}
	if _, ok := FuncMap["code"]; !ok {
		t.Error("expected code function")
	}
	if _, ok := FuncMap["json"]; !ok {
		t.Error("expected json function")
	}
}

func TestManifestFilterByKind(t *testing.T) {
	yaml := `prompts:
  - name: a
    kind: llm
  - name: b
    kind: compiler
  - name: c
    kind: llm
`
	m, err := LoadManifest([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	llm := m.FilterByKind("llm")
	if len(llm) != 2 {
		t.Errorf("expected 2 LLM prompts, got %d", len(llm))
	}

	compiler := m.FilterByKind("compiler")
	if len(compiler) != 1 {
		t.Errorf("expected 1 compiler template, got %d", len(compiler))
	}
}

func TestManifestFindByName(t *testing.T) {
	yaml := `prompts:
  - name: a
    kind: llm
  - name: b
    kind: compiler
`
	m, err := LoadManifest([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := m.FindByName("a")
	if found == nil {
		t.Fatal("expected to find prompt 'a'")
	}
	if found.Kind != "llm" {
		t.Errorf("expected kind llm, got %s", found.Kind)
	}

	if m.FindByName("z") != nil {
		t.Error("expected nil for nonexistent prompt")
	}
}

func TestLoadOverrides_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("invalid: [yaml: broken"), 0o644)

	_, err := New(WithOverridesDir(dir))
	if err == nil {
		t.Fatal("expected error for invalid override YAML")
	}
	if !strings.Contains(err.Error(), "override load error") {
		t.Errorf("expected override load error message, got: %v", err)
	}
}

func TestLoadOverrides_UnknownKind(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "unknown.yaml"), []byte("kind: unknown\nname: test"), 0o644)

	_, err := New(WithOverridesDir(dir))
	if err == nil {
		t.Fatal("expected error for unknown kind")
	}
	if !strings.Contains(err.Error(), "unknown kind") {
		t.Errorf("expected unknown kind error, got: %v", err)
	}
}

func TestLoadOverrides_ValidOverride(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "custom.yaml"), []byte(`kind: llm
name: custom-prompt
description: A custom prompt
user: "Hello {{.Name}}"
`), 0o644)

	l, err := New(WithOverridesDir(dir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, ok := l.GetLLMPrompt("custom-prompt")
	if !ok {
		t.Fatal("expected custom prompt to be loaded")
	}
	if p.User != "Hello {{.Name}}" {
		t.Errorf("expected 'Hello {{.Name}}', got %s", p.User)
	}
}
