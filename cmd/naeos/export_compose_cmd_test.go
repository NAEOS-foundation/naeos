package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportComposeMissingInput(t *testing.T) {
	dir := t.TempDir()
	configPath := writeExportTestConfig(t, dir, "")

	root := NewRootCommand()
	_, err := executeCommand(root, "export", "compose", "--config", configPath)
	if err == nil {
		t.Fatal("expected error when --input is missing")
	}
}

func TestExportComposeGeneratesFiles(t *testing.T) {
	dir := t.TempDir()
	configPath := writeExportTestConfig(t, dir, "")
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: compose-demo\nservices:\n  - name: api\n    port: 8080\nmodules:\n  - name: api\n    path: ./internal/api\n")
	outputDir := filepath.Join(dir, "docker")

	root := NewRootCommand()
	output, err := executeCommand(root, "export", "compose", "--config", configPath, "--input", specPath, "--output-dir", outputDir)
	if err != nil {
		t.Fatalf("export compose failed: %v", err)
	}
	if !strings.Contains(output, "files generated in") {
		t.Fatalf("expected generation summary, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "Dockerfile")); err != nil {
		t.Fatalf("expected generated Dockerfile: %v", err)
	}
}

func TestExportComposeInvalidConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "export", "compose", "--config", "/nonexistent/config.yaml", "--input", "project: demo", "--output-dir", t.TempDir())
	if err == nil {
		t.Fatal("expected error for nonexistent config")
	}
}

func TestExportComposePipelineFailure(t *testing.T) {
	dir := t.TempDir()
	configPath := writeExportTestConfig(t, dir, "")

	root := NewRootCommand()
	_, err := executeCommand(root, "export", "compose", "--config", configPath, "--input", ":::: not a spec ::::", "--output-dir", filepath.Join(dir, "out"))
	if err == nil {
		t.Fatal("expected error for unparseable spec input")
	}
}
