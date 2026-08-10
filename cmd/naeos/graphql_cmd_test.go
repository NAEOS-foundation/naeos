package main

import (
	"strings"
	"testing"
)

func TestGraphQLHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "graphql", "--help")
	if err != nil {
		t.Fatalf("graphql --help failed: %v", err)
	}
	if !strings.Contains(output, "Start GraphQL API server") {
		t.Fatalf("expected graphql help text, got %q", output)
	}
}

func TestGraphQLUnknownFlag(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "graphql", "--bogus-flag")
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestGraphQLListenError(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "graphql", "--port", "-1")
	if err == nil {
		t.Fatal("expected error when ListenAndServe fails")
	}
}
