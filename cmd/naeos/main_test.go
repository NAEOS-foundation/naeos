package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCreatesConfigFile(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	err := run([]string{"init", "--output", output})
	if err != nil {
		t.Fatalf("run init returned error: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected generated config file to contain content")
	}
}

func TestValidateUsesConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"validate", "--config", configPath, "--input", "sample specification"})
	if err != nil {
		t.Fatalf("run validate returned error: %v", err)
	}
}

func TestRunSupportsJSONOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input", "sample specification", "--output", "json"})
	if err != nil {
		t.Fatalf("run run returned error: %v", err)
	}
}

func TestRunSupportsYAMLOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input", "sample specification", "--output", "yaml"})
	if err != nil {
		t.Fatalf("run run returned error: %v", err)
	}
}

func TestRunWritesOutputToFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	outputPath := filepath.Join(dir, "result.json")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input", "sample specification", "--output", "json", "--output-file", outputPath})
	if err != nil {
		t.Fatalf("run run returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected output file to contain content")
	}
}
