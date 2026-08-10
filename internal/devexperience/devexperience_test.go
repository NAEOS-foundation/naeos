package devexperience

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVSCodeExtension(t *testing.T) {
	ext := NewVSCodeExtension("naeos", "1.0.0", "Test extension", "Test Author", []string{"yaml", "json"})

	if ext.Name != "naeos" {
		t.Error("expected name 'naeos'")
	}

	pkg := ext.GeneratePackageJSON()
	if !strings.Contains(pkg, "naeos") {
		t.Error("expected package JSON to contain name")
	}
	if !strings.Contains(pkg, "lsp.path") {
		t.Error("expected package JSON to contain lsp config")
	}
	if !strings.Contains(pkg, "naeos-yaml") {
		t.Error("expected package JSON to contain naeos-yaml language")
	}
	if !strings.Contains(pkg, "keybindings") {
		t.Error("expected package JSON to contain keybindings")
	}

	syntax := ext.GenerateSyntaxJSON()
	if !strings.Contains(syntax, "naeos.yaml") {
		t.Error("expected syntax to contain naeos.yaml")
	}
	if !strings.Contains(syntax, "$if") {
		t.Error("expected syntax to contain conditionals")
	}
	if !strings.Contains(syntax, "source.naeos") {
		t.Error("expected syntax scopeName")
	}
}

func TestVSCodeExtensionGenerateExtension(t *testing.T) {
	ext := NewVSCodeExtension("naeos-test", "1.0.0", "Test", "Test", nil)
	tmpDir := t.TempDir()

	if err := ext.GenerateExtension(tmpDir); err != nil {
		t.Fatalf("GenerateExtension failed: %v", err)
	}

	expectedFiles := []string{
		"package.json",
		"syntaxes/naeos.tmLanguage.json",
		"extension.js",
		".vscode/launch.json",
		"README.md",
	}
	for _, f := range expectedFiles {
		fullPath := filepath.Join(tmpDir, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestVSCodeExtensionExtensionJS(t *testing.T) {
	ext := NewVSCodeExtension("naeos", "1.0.0", "", "", nil)
	js := ext.GenerateExtensionJS()
	if !strings.Contains(js, "activate") {
		t.Error("expected extension.js to contain activate function")
	}
	if !strings.Contains(js, "naeos.lspStart") {
		t.Error("expected extension.js to contain LSP start command")
	}
	if !strings.Contains(js, "lspClient") {
		t.Error("expected extension.js to contain LSP client")
	}
}

func TestVSCodeExtensionLaunchJSON(t *testing.T) {
	ext := NewVSCodeExtension("naeos", "1.0.0", "", "", nil)
	launch := ext.GenerateLaunchJSON()
	if !strings.Contains(launch, "extensionDevelopmentPath") {
		t.Error("expected launch.json to contain extensionDevelopmentPath")
	}
}

func TestCompletionEngine(t *testing.T) {
	e := NewCompletionEngine()

	completions := e.Complete("")
	if len(completions) != len(e.commands) {
		t.Errorf("expected %d completions, got %d", len(e.commands), len(completions))
	}

	completions = e.Complete("co")
	if len(completions) != 1 {
		t.Errorf("expected 1 completion, got %d", len(completions))
	}

	completions = e.Complete("c")
	if len(completions) != 3 {
		t.Errorf("expected 3 completions, got %d", len(completions))
	}

	completions = e.Complete("compile")
	if len(completions) != 1 {
		t.Errorf("expected 1 completion, got %d", len(completions))
	}
}

func TestCompletionEngineOptions(t *testing.T) {
	e := NewCompletionEngine()

	completions := e.Complete("compile --")
	if len(completions) != 3 {
		t.Errorf("expected 3 completions, got %d", len(completions))
	}

	completions = e.Complete("compile --in")
	if len(completions) != 1 {
		t.Errorf("expected 1 completion, got %d", len(completions))
	}
}

func TestCompletionEngineShellScripts(t *testing.T) {
	e := NewCompletionEngine()

	bash := e.GenerateBashCompletion()
	if !strings.Contains(bash, "_naeos_completions") {
		t.Error("expected bash completion function")
	}

	zsh := e.GenerateZshCompletion()
	if !strings.Contains(zsh, "_naeos") {
		t.Error("expected zsh completion function")
	}

	ps := e.GeneratePowerShellCompletion()
	if !strings.Contains(ps, "Register-ArgumentCompleter") {
		t.Error("expected PowerShell completer")
	}
}

func TestSnippetManager(t *testing.T) {
	sm := NewSnippetManager()

	snippet, ok := sm.Get("neir-spec")
	if !ok {
		t.Error("expected snippet to exist")
	}
	if !strings.Contains(snippet, "project:") {
		t.Error("expected snippet to contain project")
	}

	snippets := sm.List()
	if len(snippets) != 3 {
		t.Errorf("expected 3 snippets, got %d", len(snippets))
	}
}

func TestSnippetManagerAdd(t *testing.T) {
	sm := NewSnippetManager()

	sm.Add("custom", "custom snippet")
	snippet, ok := sm.Get("custom")
	if !ok || snippet != "custom snippet" {
		t.Error("expected custom snippet")
	}
}

func TestDevExperienceStack(t *testing.T) {
	stack := NewStack()

	if stack.Extension == nil {
		t.Error("expected extension")
	}
	if stack.Engine == nil {
		t.Error("expected engine")
	}
	if stack.Snippets == nil {
		t.Error("expected snippets")
	}
}
