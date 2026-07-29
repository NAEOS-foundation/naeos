package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffRequiresInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "diff")
	if err == nil {
		t.Fatal("expected error for missing --input/--input-file")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected 'missing' error, got: %v", err)
	}
}

func TestDiffWithInput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input", "project: test", "--output-dir", dir)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected Summary in output, got %q", output)
	}
}

func TestDiffVisual(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input", "project: visual-test", "--output-dir", dir, "--visual")
	if err != nil {
		t.Fatalf("diff --visual failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected Summary, got %q", output)
	}
}

func TestDiffFormatUnified(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input", "project: unified-test", "--output-dir", dir, "--format", "unified")
	if err != nil {
		t.Fatalf("diff --format unified failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected Summary, got %q", output)
	}
}

func TestBuildWithInput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "build", "--config", configPath, "--input", "project: build-test")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if !strings.Contains(output, "pipeline") {
		t.Fatalf("expected pipeline output, got %q", output)
	}
}

func TestBuildDryRun(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "build", "--config", configPath, "--input", "project: dryrun-test", "--dry-run")
	if err != nil {
		t.Fatalf("build --dry-run failed: %v", err)
	}
	if !strings.Contains(output, "pipeline") {
		t.Fatalf("expected pipeline output, got %q", output)
	}
}

func TestLSPCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "lsp", "--help")
	if err != nil {
		t.Fatalf("lsp --help failed: %v", err)
	}
	if !strings.Contains(output, "Language Server Protocol") {
		t.Fatalf("expected LSP help, got %q", output)
	}
}

func TestDXCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "dx", "--help")
	if err != nil {
		t.Fatalf("dx --help failed: %v", err)
	}
	if !strings.Contains(output, "VS Code") {
		t.Fatalf("expected DX help about VS Code, got %q", output)
	}
}

func TestDXVSCodeGen(t *testing.T) {
	dir := t.TempDir()

	root := NewRootCommand()
	output, err := executeCommand(root, "dx", "vscode-gen", "--output", dir)
	if err != nil {
		t.Fatalf("dx vscode-gen failed: %v", err)
	}
	if !strings.Contains(output, "VS Code extension generated") {
		t.Fatalf("expected success message, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "package.json")); os.IsNotExist(err) {
		t.Fatal("expected package.json to be generated")
	}
	if _, err := os.Stat(filepath.Join(dir, "extension.js")); os.IsNotExist(err) {
		t.Fatal("expected extension.js to be generated")
	}
}