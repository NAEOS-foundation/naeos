package main

import (
	"strings"
	"testing"
)

func TestCloudCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "cloud")
	if err != nil {
		t.Fatalf("execute cloud failed: %v", err)
	}
}

func TestCloudTypesListsResources(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "types")
	if err != nil {
		t.Fatalf("execute cloud types failed: %v", err)
	}
	if !strings.Contains(output, "Supported resource types") {
		t.Fatalf("expected resource types header, got %q", output)
	}
}

func TestCloudTypesJSONOutput(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "types", "--output-format", "json")
	if err != nil {
		t.Fatalf("execute cloud types json failed: %v", err)
	}
	if !strings.Contains(output, "[") {
		t.Fatalf("expected JSON array output, got %q", output)
	}
}

func TestCloudDeployRequiresConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "cloud", "deploy", "--provider", "aws", "--region", "us-east-1", "--project", "test")
	if err == nil {
		t.Fatal("expected error when deploy has no resources")
	}
}

func TestCloudPlanRequiresConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "cloud", "plan", "--provider", "aws", "--region", "us-east-1", "--project", "test")
	if err == nil {
		t.Fatal("expected error when plan has no resources")
	}
}

func TestCloudExportRequiresConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "cloud", "export", "--provider", "aws", "--region", "us-east-1", "--project", "test")
	if err == nil {
		t.Fatal("expected error when export has no resources")
	}
}

func TestCloudStatusEmpty(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "status")
	if err != nil {
		t.Fatalf("execute cloud status failed: %v", err)
	}
	if !strings.Contains(output, "No deployments") && !strings.Contains(output, "Deployed") {
		t.Fatalf("expected status output, got %q", output)
	}
}
