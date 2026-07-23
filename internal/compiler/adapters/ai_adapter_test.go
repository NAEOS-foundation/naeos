package adapters

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/testutil"
)

func TestNewAICompilerAdapter(t *testing.T) {
	t.Parallel()
	llm := newMockLLMService()
	a := NewAICompilerAdapter(compiler.TargetClaude, llm)
	if a == nil {
		t.Fatal("expected non-nil adapter")
	}
	if a.Target() != compiler.TargetClaude {
		t.Errorf("expected claude target, got %s", a.Target())
	}
}

func TestAICompilerAdapterTarget(t *testing.T) {
	t.Parallel()
	llm := newMockLLMService()
	tests := []compiler.Target{
		compiler.TargetCopilot,
		compiler.TargetClaude,
		compiler.TargetCursor,
		compiler.TargetGemini,
		compiler.TargetCodex,
		compiler.TargetOpenCode,
		compiler.TargetWindsurf,
	}
	for _, tt := range tests {
		a := NewAICompilerAdapter(tt, llm)
		if a.Target() != tt {
			t.Errorf("expected target %s, got %s", tt, a.Target())
		}
	}
}

func TestStructuredBufferWriteAndString(t *testing.T) {
	t.Parallel()
	var buf structuredBuffer
	n, err := buf.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	n, err = buf.Write([]byte(" world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 6 {
		t.Errorf("expected 6 bytes written, got %d", n)
	}
	if buf.String() != "hello world" {
		t.Errorf("expected 'hello world', got %q", buf.String())
	}
}

func TestStructuredBufferEmpty(t *testing.T) {
	t.Parallel()
	var buf structuredBuffer
	if buf.String() != "" {
		t.Errorf("expected empty string, got %q", buf.String())
	}
}

func TestStructuredBufferMultipleWrites(t *testing.T) {
	t.Parallel()
	var buf structuredBuffer
	buf.Write([]byte("one"))
	buf.Write([]byte("two"))
	buf.Write([]byte("three"))
	if buf.String() != "onetwothree" {
		t.Errorf("expected 'onetwothree', got %q", buf.String())
	}
}

func TestParseCompiledFilesNonJSON(t *testing.T) {
	t.Parallel()
	_, err := parseCompiledFiles("this is not json at all")
	if err == nil {
		t.Error("expected error for non-JSON input")
	}
}

func TestParseCompiledFilesObjectInsteadOfArray(t *testing.T) {
	t.Parallel()
	_, err := parseCompiledFiles(`{"path":"x","content":"y","kind":"z"}`)
	if err == nil {
		t.Error("expected error for object instead of array")
	}
}

func TestBuildNEIRContextEmpty(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{}
	ctx := buildNEIRContext(neir)
	if ctx != "" {
		t.Errorf("expected empty context for empty NEIR, got %q", ctx)
	}
}

func TestBuildNEIRContextProjectOnly(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{
			Name:        "test-proj",
			Description: "A test",
			Version:     "1.0.0",
		},
	}
	ctx := buildNEIRContext(neir)
	if !testutil.Contains(ctx, "test-proj") {
		t.Error("expected project name in context")
	}
	if !testutil.Contains(ctx, "A test") {
		t.Error("expected project description")
	}
	if !testutil.Contains(ctx, "1.0.0") {
		t.Error("expected project version")
	}
}

func TestBuildNEIRContextWithNilFields(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
	}
	ctx := buildNEIRContext(neir)
	if !testutil.Contains(ctx, "test") {
		t.Error("expected project name")
	}
	if testutil.Contains(ctx, "Security") {
		t.Error("should not contain Security for nil security")
	}
	if testutil.Contains(ctx, "Deployment") {
		t.Error("should not contain Deployment for nil deployment")
	}
	if testutil.Contains(ctx, "Testing") {
		t.Error("should not contain Testing for nil testing")
	}
}
