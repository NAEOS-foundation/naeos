package main

import (
	"strings"
	"testing"
)

func TestWorkflowCommandShowsHelp(t *testing.T) {
	root := newRootCommand()
	_, err := executeCommand(root, "workflow")
	if err != nil {
		t.Fatalf("execute workflow failed: %v", err)
	}
}

func TestWorkflowListEmpty(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "workflow", "list")
	if err != nil {
		t.Fatalf("workflow list failed: %v", err)
	}
	if !strings.Contains(output, "No workflows defined") {
		t.Fatalf("expected no workflows message, got %q", output)
	}
}

func TestWorkflowCreate(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "workflow", "create", "--name", "deploy", "--steps", "build,test,deploy")
	if err != nil {
		t.Fatalf("workflow create failed: %v", err)
	}
	if !strings.Contains(output, "Created workflow") {
		t.Fatalf("expected create success message, got %q", output)
	}
}

func TestWorkflowExecuteNotFound(t *testing.T) {
	root := newRootCommand()
	_, err := executeCommand(root, "workflow", "execute", "--name", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent workflow")
	}
}

func TestWorkflowRequestsEmpty(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "workflow", "requests", "--status", "pending")
	if err != nil {
		t.Fatalf("workflow requests failed: %v", err)
	}
	if !strings.Contains(output, "No pending requests") {
		t.Fatalf("expected no requests message, got %q", output)
	}
}
