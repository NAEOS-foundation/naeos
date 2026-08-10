package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeExportTestConfig(t *testing.T, dir, outputDir string) string {
	t.Helper()
	configPath := filepath.Join(dir, "config.yaml")
	content := "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n"
	if outputDir != "" {
		content += "  output_dir: " + outputDir + "\n"
	}
	writeTestFile(t, dir, "config.yaml", content)
	return configPath
}

func TestExportMissingInput(t *testing.T) {
	dir := t.TempDir()
	configPath := writeExportTestConfig(t, dir, "")

	root := NewRootCommand()
	_, err := executeCommand(root, "export", "--config", configPath)
	if err == nil {
		t.Fatal("expected error when --input is missing")
	}
}

func TestExportMissingOutputDir(t *testing.T) {
	dir := t.TempDir()
	configPath := writeExportTestConfig(t, dir, "")

	root := NewRootCommand()
	_, err := executeCommand(root, "export", "--config", configPath, "--input", "project: demo")
	if err == nil {
		t.Fatal("expected error when no output directory is configured")
	}
}

func TestExportWritesArtifacts(t *testing.T) {
	dir := t.TempDir()
	configPath := writeExportTestConfig(t, dir, "")
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: export-demo\nmodules:\n  - name: api\n    path: ./internal/api\n")
	outputDir := filepath.Join(dir, "generated")

	root := NewRootCommand()
	output, err := executeCommand(root, "export", "--config", configPath, "--input", specPath, "--output-dir", outputDir)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if !strings.Contains(output, "exported") {
		t.Fatalf("expected export summary, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "README.md")); err != nil {
		t.Fatalf("expected exported README.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "Dockerfile")); err != nil {
		t.Fatalf("expected exported Dockerfile: %v", err)
	}
}

func TestExportDryRunPreview(t *testing.T) {
	dir := t.TempDir()
	configPath := writeExportTestConfig(t, dir, "")
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: export-demo\nmodules:\n  - name: api\n    path: ./internal/api\n")
	outputDir := filepath.Join(dir, "generated")

	root := NewRootCommand()
	output, err := executeCommand(root, "export", "--config", configPath, "--input", specPath, "--output-dir", outputDir, "--dry-run")
	if err != nil {
		t.Fatalf("export dry-run failed: %v", err)
	}
	if !strings.Contains(output, "Export Preview") || !strings.Contains(output, "Total:") {
		t.Fatalf("expected preview output, got %q", output)
	}
	entries, err := os.ReadDir(outputDir)
	if err == nil && len(entries) > 0 {
		t.Fatalf("expected no files written in dry-run, found %d", len(entries))
	}
}

func TestExportLanguageFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := writeExportTestConfig(t, dir, "")
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: export-demo\nmodules:\n  - name: api\n    path: ./internal/api\n")
	outputDir := filepath.Join(dir, "generated")

	root := NewRootCommand()
	output, err := executeCommand(root, "export", "--config", configPath, "--input", specPath, "--output-dir", outputDir, "--language", "go")
	if err != nil {
		t.Fatalf("export with language failed: %v", err)
	}
	if !strings.Contains(output, "languages: go") {
		t.Fatalf("expected language line in output, got %q", output)
	}
}

func TestExportInvalidConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "export", "--config", "/nonexistent/config.yaml", "--input", "project: demo", "--output-dir", t.TempDir())
	if err == nil {
		t.Fatal("expected error for nonexistent config")
	}
}
