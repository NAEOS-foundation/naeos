package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAICommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "ai")
	if err != nil {
		t.Fatalf("execute ai failed: %v", err)
	}
}

func TestAISuggestRequiresInputFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "ai", "suggest")
	if err == nil {
		t.Fatal("expected error when --input-file is missing")
	}
}

func TestAISuggestWithFile(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(specPath, []byte("project: test\nmodules:\n  - name: api\n    path: ./api\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "ai", "suggest", "--input-file", specPath)
	if err != nil {
		t.Fatalf("execute ai suggest failed: %v", err)
	}
	if len(strings.TrimSpace(output)) == 0 {
		t.Fatal("expected non-empty suggestion output")
	}
}

func TestAIExplainRequiresArgs(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "ai", "explain")
	if err == nil {
		t.Fatal("expected error when no topic is provided")
	}
}

func TestAIExplainPipeline(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "ai", "explain", "pipeline")
	if err != nil {
		t.Fatalf("execute ai explain failed: %v", err)
	}
	if len(strings.TrimSpace(output)) == 0 {
		t.Fatal("expected non-empty explanation output")
	}
}

func TestAIExplainNEIR(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "ai", "explain", "neir")
	if err != nil {
		t.Fatalf("execute ai explain neir failed: %v", err)
	}
	if len(strings.TrimSpace(output)) == 0 {
		t.Fatal("expected non-empty explanation output")
	}
}
