package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "init")
	if err != nil {
		t.Fatalf("execute init failed: %v", err)
	}
}

func TestInitDefaultTemplate(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	root := NewRootCommand()
	_, err := executeCommand(root, "init", "--output", output)
	if err != nil {
		t.Fatalf("execute init failed: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if !strings.Contains(string(data), "my-project") {
		t.Fatalf("expected basic template content, got %q", string(data))
	}
}

func TestInitMicroservicesTemplate(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	root := NewRootCommand()
	_, err := executeCommand(root, "init", "--template", "microservices", "--output", output)
	if err != nil {
		t.Fatalf("execute init failed: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if !strings.Contains(string(data), "api-gateway") || !strings.Contains(string(data), "user-service") {
		t.Fatalf("expected microservices template content, got %q", string(data))
	}
}

func TestInitRESTAPITemplate(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	root := NewRootCommand()
	_, err := executeCommand(root, "init", "--template", "rest-api", "--output", output)
	if err != nil {
		t.Fatalf("execute init failed: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if !strings.Contains(string(data), "my-rest-api") {
		t.Fatalf("expected rest-api template content, got %q", string(data))
	}
}

func TestInitFullstackTemplate(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	root := NewRootCommand()
	_, err := executeCommand(root, "init", "--template", "fullstack", "--output", output)
	if err != nil {
		t.Fatalf("execute init failed: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if !strings.Contains(string(data), "frontend") || !strings.Contains(string(data), "worker") {
		t.Fatalf("expected fullstack template content, got %q", string(data))
	}
}

func TestInitKubernetesTemplate(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	root := NewRootCommand()
	_, err := executeCommand(root, "init", "--template", "kubernetes", "--output", output)
	if err != nil {
		t.Fatalf("execute init failed: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if !strings.Contains(string(data), "kubernetes") || !strings.Contains(string(data), "replicas") {
		t.Fatalf("expected kubernetes template content, got %q", string(data))
	}
}

func TestInitWithProjectName(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	root := NewRootCommand()
	_, err := executeCommand(root, "init", "--output", output, "--name", "my-app")
	if err != nil {
		t.Fatalf("execute init failed: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if !strings.Contains(string(data), "my-app") {
		t.Fatalf("expected project name 'my-app', got %q", string(data))
	}
	if strings.Contains(string(data), "my-project") {
		t.Fatalf("expected default name to be replaced, got %q", string(data))
	}
}

func TestInitListTemplates(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "init", "--list-templates")
	if err != nil {
		t.Fatalf("execute init --list-templates failed: %v", err)
	}
	if !strings.Contains(output, "basic") || !strings.Contains(output, "microservices") {
		t.Fatalf("expected template list, got %q", output)
	}
}

func TestInitUnknownTemplate(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	root := NewRootCommand()
	_, err := executeCommand(root, "init", "--template", "nonexistent", "--output", output)
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
}
