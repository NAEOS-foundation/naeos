package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigResolveTable(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	writeTestFile(t, dir, "config.json", `{"name":"myproject","region":"us-east-1"}`)

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "resolve", cfg)
	if err != nil {
		t.Fatalf("config resolve failed: %v", err)
	}
	if !strings.Contains(output, "resolved 2 keys from "+cfg) {
		t.Fatalf("expected resolve summary, got %q", output)
	}
	if !strings.Contains(output, "myproject") {
		t.Fatalf("expected resolved value in table, got %q", output)
	}
}

func TestConfigResolveJSONOutput(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	writeTestFile(t, dir, "config.json", `{"name":"myproject"}`)

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "resolve", cfg, "--output", "json")
	if err != nil {
		t.Fatalf("config resolve failed: %v", err)
	}
	if !strings.Contains(output, `"key": "name"`) || !strings.Contains(output, `"value": "myproject"`) {
		t.Fatalf("expected json output, got %q", output)
	}
}

func TestConfigResolveEnvReference(t *testing.T) {
	t.Setenv("NAEOS_TEST_TOKEN", "secretvalue")

	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	writeTestFile(t, dir, "config.json", `{"token":"env:NAEOS_TEST_TOKEN"}`)

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "resolve", cfg)
	if err != nil {
		t.Fatalf("config resolve failed: %v", err)
	}
	if !strings.Contains(output, "secretvalue") {
		t.Fatalf("expected resolved env value, got %q", output)
	}
}

func TestConfigResolveMissingFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "resolve", "/nonexistent/config.json")
	if err == nil {
		t.Fatal("expected error for nonexistent config file")
	}
}

func TestConfigResolveMissingArg(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "resolve")
	if err == nil {
		t.Fatal("expected error when path argument is missing")
	}
}

func TestConfigResolveInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	writeTestFile(t, dir, "config.json", `not json`)

	root := NewRootCommand()
	if _, err := executeCommand(root, "config", "resolve", cfg); err == nil {
		t.Fatal("expected error for invalid json config")
	}
}

func TestConfigTestEnvReference(t *testing.T) {
	t.Setenv("NAEOS_TEST_REF", "hello")

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "test", "env:NAEOS_TEST_REF")
	if err != nil {
		t.Fatalf("config test failed: %v", err)
	}
	if !strings.Contains(output, "Provider: env") || !strings.Contains(output, "hello") {
		t.Fatalf("expected resolved ref output, got %q", output)
	}
}

func TestConfigTestMissingEnvReference(t *testing.T) {
	t.Setenv("NAEOS_TEST_REF", "")
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "test", "env:NAEOS_TEST_UNDEFINED_XYZ")
	if err == nil {
		t.Fatal("expected error for missing env reference")
	}
}

func TestConfigTestMissingArg(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "test")
	if err == nil {
		t.Fatal("expected error when reference argument is missing")
	}
}

func TestConfigSources(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "config", "sources")
	if err != nil {
		t.Fatalf("config sources failed: %v", err)
	}
	for _, want := range []string{"env", "file", "secret", "vault", "plain"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in sources output, got %q", want, output)
		}
	}
}
