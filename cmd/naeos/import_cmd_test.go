package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testHCLSpec = `project "demo" {
  version = "1.0.0"
  description = "test project"
}

service "api" {
  port = 8080
  type = "backend"
}

infra "main" {
  engine = "docker"
}
`

func TestImportYAMLToStdout(t *testing.T) {
	dir := t.TempDir()
	hclPath := filepath.Join(dir, "spec.hcl")
	writeTestFile(t, dir, "spec.hcl", testHCLSpec)

	root := NewRootCommand()
	output, err := executeCommand(root, "import", "--input", hclPath)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if !strings.Contains(output, "project:") || !strings.Contains(output, "name: demo") {
		t.Fatalf("expected yaml project output, got %q", output)
	}
	if !strings.Contains(output, "port: 8080") {
		t.Fatalf("expected service port in output, got %q", output)
	}
}

func TestImportJSONToFile(t *testing.T) {
	dir := t.TempDir()
	hclPath := filepath.Join(dir, "spec.hcl")
	writeTestFile(t, dir, "spec.hcl", testHCLSpec)
	outputPath := filepath.Join(dir, "spec.json")

	root := NewRootCommand()
	output, err := executeCommand(root, "import", "--input", hclPath, "--format", "json", "--output", outputPath)
	if err != nil {
		t.Fatalf("import json failed: %v", err)
	}
	if !strings.Contains(output, "Imported") || !strings.Contains(output, "json format") {
		t.Fatalf("expected import summary, got %q", output)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(data), `"project"`) || !strings.Contains(string(data), "demo") {
		t.Fatalf("expected json output, got %q", string(data))
	}
}

func TestImportMissingInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "import")
	if err == nil {
		t.Fatal("expected error when --input is missing")
	}
}

func TestImportNonexistentFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "import", "--input", "/nonexistent/spec.hcl")
	if err == nil {
		t.Fatal("expected error for nonexistent input file")
	}
}

func TestImportInvalidHCL(t *testing.T) {
	dir := t.TempDir()
	hclPath := filepath.Join(dir, "bad.hcl")
	writeTestFile(t, dir, "bad.hcl", "service \"api\" {\n  port = \"not-a-number\"\n}\n")

	root := NewRootCommand()
	_, err := executeCommand(root, "import", "--input", hclPath)
	if err == nil {
		t.Fatal("expected error for invalid HCL input")
	}
}
