package adapters

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
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

func TestParseCompiledFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		count   int
		wantErr bool
	}{
		{
			name:    "valid JSON array",
			input:   `[{"path":"test.md","content":"hello","kind":"instructions"}]`,
			count:   1,
			wantErr: false,
		},
		{
			name:    "multiple files",
			input:   `[{"path":"a.md","content":"a","kind":"docs"},{"path":"b.md","content":"b","kind":"rules"}]`,
			count:   2,
			wantErr: false,
		},
		{
			name:    "empty array",
			input:   `[]`,
			count:   0,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			count:   0,
			wantErr: true,
		},
		{
			name:    "JSON in code block",
			input:   "```json\n[{\"path\":\"x\",\"content\":\"y\",\"kind\":\"z\"}]\n```",
			count:   1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := parseCompiledFiles(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(files) != tt.count {
				t.Errorf("got %d files, want %d", len(files), tt.count)
			}
		})
	}
}

func TestBuildNEIRContext(t *testing.T) {
	t.Parallel()

	neir := testNEIR()
	ctx := buildNEIRContext(neir)
	if ctx == "" {
		t.Fatal("expected non-empty context")
	}
	if !strings.Contains(ctx, "test-project") {
		t.Error("expected project name")
	}
	if !strings.Contains(ctx, "hexagonal") {
		t.Error("expected architecture pattern")
	}
	if !strings.Contains(ctx, "core") {
		t.Error("expected module name")
	}
	if !strings.Contains(ctx, "http-api") {
		t.Error("expected service name")
	}
}

func TestCopilotAdapter(t *testing.T) {
	a := NewCopilotAdapter(nil)
	if a.Target() != compiler.TargetCopilot {
		t.Errorf("expected copilot target, got %s", a.Target())
	}

	out, err := a.Compile(testNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(out.Files))
	}
	if out.Target != compiler.TargetCopilot {
		t.Errorf("expected copilot target in output")
	}
}

func TestCopilotAdapterNil(t *testing.T) {
	a := NewCopilotAdapter(nil)
	_, err := a.Compile(nil)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestClaudeAdapter(t *testing.T) {
	a := NewClaudeAdapter(nil)
	if a.Target() != compiler.TargetClaude {
		t.Errorf("expected claude target, got %s", a.Target())
	}

	out, err := a.Compile(testNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(out.Files))
	}
}

func TestClaudeAdapterNil(t *testing.T) {
	a := NewClaudeAdapter(nil)
	_, err := a.Compile(nil)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestCursorAdapter(t *testing.T) {
	a := NewCursorAdapter(nil)
	if a.Target() != compiler.TargetCursor {
		t.Errorf("expected cursor target, got %s", a.Target())
	}

	out, err := a.Compile(testNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(out.Files))
	}
}

func TestCursorAdapterNil(t *testing.T) {
	a := NewCursorAdapter(nil)
	_, err := a.Compile(nil)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestGeminiAdapter(t *testing.T) {
	a := NewGeminiAdapter(nil)
	if a.Target() != compiler.TargetGemini {
		t.Errorf("expected gemini target, got %s", a.Target())
	}

	out, err := a.Compile(testNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(out.Files))
	}
}

func TestGeminiAdapterNil(t *testing.T) {
	a := NewGeminiAdapter(nil)
	_, err := a.Compile(nil)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestCodexAdapter(t *testing.T) {
	a := NewCodexAdapter(nil)
	if a.Target() != compiler.TargetCodex {
		t.Errorf("expected codex target, got %s", a.Target())
	}

	out, err := a.Compile(testNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(out.Files))
	}
}

func TestCodexAdapterNil(t *testing.T) {
	a := NewCodexAdapter(nil)
	_, err := a.Compile(nil)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestOpenCodeAdapter(t *testing.T) {
	a := NewOpenCodeAdapter(nil)
	if a.Target() != compiler.TargetOpenCode {
		t.Errorf("expected opencode target, got %s", a.Target())
	}

	out, err := a.Compile(testNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(out.Files))
	}
}

func TestOpenCodeAdapterNil(t *testing.T) {
	a := NewOpenCodeAdapter(nil)
	_, err := a.Compile(nil)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestWindsurfAdapter(t *testing.T) {
	a := NewWindsurfAdapter(nil)
	if a.Target() != compiler.TargetWindsurf {
		t.Errorf("expected windsurf target, got %s", a.Target())
	}

	out, err := a.Compile(testNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(out.Files))
	}
}

func TestWindsurfAdapterNil(t *testing.T) {
	a := NewWindsurfAdapter(nil)
	_, err := a.Compile(nil)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestAllAdaptersContent(t *testing.T) {
	adapters := []compiler.Adapter{
		NewCopilotAdapter(nil),
		NewClaudeAdapter(nil),
		NewCursorAdapter(nil),
		NewGeminiAdapter(nil),
		NewCodexAdapter(nil),
		NewOpenCodeAdapter(nil),
		NewWindsurfAdapter(nil),
	}

	neir := testNEIR()
	for _, a := range adapters {
		out, err := a.Compile(neir)
		if err != nil {
			t.Fatalf("adapter %s: %v", a.Target(), err)
		}
		if out.Summary == "" {
			t.Errorf("adapter %s: empty summary", a.Target())
		}
		for _, f := range out.Files {
			if f.Content == "" {
				t.Errorf("adapter %s: empty file content for %s", a.Target(), f.Path)
			}
		}
	}
}
