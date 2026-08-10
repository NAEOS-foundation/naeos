package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func writePipelineConfig(t *testing.T, dir string) string {
	t.Helper()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: diff-test\n  mode: development\n  output_dir: "+filepath.Join(dir, "out")+"\n")
	return configPath
}

func TestDiffMissingInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "diff", "--config", writePipelineConfig(t, t.TempDir()))
	if err == nil {
		t.Fatal("expected error when both --input and --input-file are missing")
	}
}

func TestDiffInlineInput(t *testing.T) {
	dir := t.TempDir()
	configPath := writePipelineConfig(t, dir)

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input", "project: diff-inline")
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected summary in output, got %q", output)
	}
}

func TestDiffInputFile(t *testing.T) {
	dir := t.TempDir()
	configPath := writePipelineConfig(t, dir)
	spec := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: diff-file\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input-file", spec)
	if err != nil {
		t.Fatalf("diff input-file failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected summary in output, got %q", output)
	}
}

func TestDiffOutputDirFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := writePipelineConfig(t, dir)
	outDir := filepath.Join(dir, "existing-out")

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input", "project: diff-out", "--output-dir", outDir)
	if err != nil {
		t.Fatalf("diff output-dir failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected summary in output, got %q", output)
	}
}

func TestDiffBadConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "diff", "--config", "/nonexistent/config.yaml", "--input", "project: x")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestDiffMissingInputFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "diff", "--config", writePipelineConfig(t, t.TempDir()), "--input-file", "/nonexistent/spec.yaml")
	if err == nil {
		t.Fatal("expected error for missing input file")
	}
}
