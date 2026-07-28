package main

import (
	"strings"
	"testing"
)

func TestWorkflowCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "workflow")
	if err != nil {
		t.Fatalf("execute workflow failed: %v", err)
	}
}

func TestWorkflowCreateAndList(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "workflow", "create", "--name", "deploy", "--steps", "build,test,deploy")
	if err != nil {
		t.Fatalf("workflow create failed: %v", err)
	}
	if !strings.Contains(output, "Created workflow") {
		t.Fatalf("expected create success message, got %q", output)
	}

	output, err = executeCommand(root, "workflow", "list")
	if err != nil {
		t.Fatalf("workflow list failed: %v", err)
	}
	if !strings.Contains(output, "deploy") {
		t.Fatalf("expected workflow list to contain 'deploy', got %q", output)
	}
}

func TestWorkflowExecuteNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "workflow", "execute", "--name", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent workflow")
	}
}

func TestWorkflowRequestsEmpty(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "workflow", "requests", "--status", "pending")
	if err != nil {
		t.Fatalf("workflow requests failed: %v", err)
	}
	if !strings.Contains(output, "No pending requests") {
		t.Fatalf("expected no requests message, got %q", output)
	}
}
