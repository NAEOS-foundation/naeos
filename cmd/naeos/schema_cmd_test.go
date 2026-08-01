package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const testSchemaJSON = `{
  "title": "NEIR Schema",
  "description": "Canonical NEIR spec schema",
  "$id": "neir://v1",
  "properties": {
    "project": {"type": "string"},
    "mode": {"type": "string", "enum": ["dev", "prod"]}
  },
  "required": ["project"],
  "definitions": {}
}`

func schemaRegistryFlag(t *testing.T) string {
	t.Helper()
	return "file://" + filepath.Join(t.TempDir(), "neir.json")
}

func TestSchemaValidateValid(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "neir.json")
	writeTestFile(t, dir, "neir.json", testSchemaJSON)
	writeTestFile(t, dir, "spec.yaml", "project: demo\nmode: dev\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "schema", "validate", filepath.Join(dir, "spec.yaml"), "--registry", "file://"+schemaPath)
	if err != nil {
		t.Fatalf("schema validate failed: %v", err)
	}
	if !strings.Contains(output, "Valid") {
		t.Fatalf("expected valid message, got %q", output)
	}
}

func TestSchemaValidateInvalid(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "neir.json")
	writeTestFile(t, dir, "neir.json", testSchemaJSON)
	writeTestFile(t, dir, "spec.yaml", "mode: staging\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "schema", "validate", filepath.Join(dir, "spec.yaml"), "--registry", "file://"+schemaPath)
	if err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failed error, got %v", err)
	}
	if !strings.Contains(output, "Invalid") {
		t.Fatalf("expected invalid message, got %q", output)
	}
}

func TestSchemaValidateJSONOutput(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "neir.json")
	writeTestFile(t, dir, "neir.json", testSchemaJSON)
	writeTestFile(t, dir, "spec.yaml", "project: demo\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "schema", "validate", filepath.Join(dir, "spec.yaml"), "--registry", "file://"+schemaPath, "--output", "json")
	if err != nil {
		t.Fatalf("schema validate json failed: %v", err)
	}
	if !strings.Contains(output, `"valid": true`) {
		t.Fatalf("expected json valid field, got %q", output)
	}
}

func TestSchemaValidateYAMLOutput(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "neir.json")
	writeTestFile(t, dir, "neir.json", testSchemaJSON)
	writeTestFile(t, dir, "spec.yaml", "mode: staging\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "schema", "validate", filepath.Join(dir, "spec.yaml"), "--registry", "file://"+schemaPath, "--output", "yaml")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(output, "valid: false") {
		t.Fatalf("expected yaml valid field, got %q", output)
	}
}

func TestSchemaValidateMissingFile(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "neir.json")
	writeTestFile(t, dir, "neir.json", testSchemaJSON)

	root := NewRootCommand()
	_, err := executeCommand(root, "schema", "validate", filepath.Join(dir, "missing.yaml"), "--registry", "file://"+schemaPath)
	if err == nil {
		t.Fatal("expected error for missing spec file")
	}
}

func TestSchemaValidateRegistryUnreachable(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "spec.yaml", "project: demo\n")

	root := NewRootCommand()
	_, err := executeCommand(root, "schema", "validate", filepath.Join(dir, "spec.yaml"), "--registry", "http://127.0.0.1:1/neir.json")
	if err == nil || !strings.Contains(err.Error(), "fetch schema") {
		t.Fatalf("expected fetch schema error, got %v", err)
	}
}

func TestSchemaInfo(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "neir.json")
	writeTestFile(t, dir, "neir.json", testSchemaJSON)

	root := NewRootCommand()
	output, err := executeCommand(root, "schema", "info", "--registry", "file://"+schemaPath)
	if err != nil {
		t.Fatalf("schema info failed: %v", err)
	}
	if !strings.Contains(output, "Schema Registry") || !strings.Contains(output, "NEIR Schema") {
		t.Fatalf("expected schema info output, got %q", output)
	}
	if !strings.Contains(output, "Required fields:    project") {
		t.Fatalf("expected required fields line, got %q", output)
	}
}

func TestSchemaInfoRegistryUnreachable(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "schema", "info", "--registry", "http://127.0.0.1:1/neir.json")
	if err == nil || !strings.Contains(err.Error(), "fetch schema") {
		t.Fatalf("expected fetch schema error, got %v", err)
	}
}
