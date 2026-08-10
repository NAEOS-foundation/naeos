package main

import (
	"os"
	"strings"
	"testing"
)

func TestLSPInvalidArgs(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "lsp", "extra-arg")
	if err == nil {
		t.Fatal("expected error when args are provided")
	}
}

func TestLSPStdioEOFExitsCleanly(t *testing.T) {
	origStdin := os.Stdin
	t.Cleanup(func() {
		os.Stdin = origStdin
	})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	w.Close()
	os.Stdin = r

	root := NewRootCommand()
	output, err := executeCommand(root, "lsp")
	if err != nil {
		t.Fatalf("lsp with closed stdin should exit cleanly: %v", err)
	}
	if !strings.Contains(output, "NEIR LSP server started on stdio") {
		t.Fatalf("expected server start message, got %q", output)
	}
}
