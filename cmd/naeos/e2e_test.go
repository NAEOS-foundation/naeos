package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusCommandShowsConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: status-test\n  mode: production\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := newRootCommand()
	output, err := executeCommand(root, "status", "--config", configPath)
	if err != nil {
		t.Fatalf("execute status failed: %v", err)
	}
	if !strings.Contains(output, "status-test") {
		t.Fatalf("expected status output to contain pipeline name, got %q", output)
	}
	if !strings.Contains(output, "production") {
		t.Fatalf("expected status output to contain mode, got %q", output)
	}
}

func TestStatusAutoDetectsConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: auto-detect\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	root := newRootCommand()
	output, err := executeCommand(root, "status")
	if err != nil {
		t.Fatalf("execute status failed: %v", err)
	}
	if !strings.Contains(output, "auto-detect") {
		t.Fatalf("expected auto-detect config, got %q", output)
	}
}

func TestRunDryRunFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	outputDir := filepath.Join(dir, "out")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: dryrun-test\n  mode: development\n  output_dir: "+outputDir+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input", "test spec", "--dry-run"})
	if err != nil {
		t.Fatalf("run with dry-run returned error: %v", err)
	}

	if _, err := os.Stat(outputDir); err == nil {
		t.Fatal("expected output directory NOT to exist with dry-run")
	}
}

func TestExportDryRunFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	outputDir := filepath.Join(dir, "generated")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: export-dry\n  mode: development\n  output_dir: "+outputDir+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"export", "--config", configPath, "--input", "test spec", "--dry-run"})
	if err != nil {
		t.Fatalf("export dry-run returned error: %v", err)
	}

	if _, err := os.Stat(outputDir); err == nil {
		t.Fatal("expected output directory NOT to exist with dry-run")
	}
}

func TestGlobalDryRunFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	outputDir := filepath.Join(dir, "out")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: global-dry\n  mode: development\n  output_dir: "+outputDir+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"--dry-run", "run", "--config", configPath, "--input", "test spec"})
	if err != nil {
		t.Fatalf("global dry-run returned error: %v", err)
	}

	if _, err := os.Stat(outputDir); err == nil {
		t.Fatal("expected output directory NOT to exist with global dry-run")
	}
}

func TestDoctorAutoDetectsConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: doctor-auto\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	root := newRootCommand()
	_, err := executeCommand(root, "doctor")
	if err != nil {
		t.Fatalf("execute doctor auto-detect failed: %v", err)
	}
}

func TestE2EFullSpecWithMinimal(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	outputDir := filepath.Join(dir, "out")

	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: e2e-minimal\n  mode: development\n  output_dir: "+outputDir+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("project: my-e2e-project\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input-file", specPath})
	if err != nil {
		t.Fatalf("e2e run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "go.mod")); err != nil {
		t.Fatal("expected go.mod to be generated")
	}
	if _, err := os.Stat(filepath.Join(outputDir, "README.md")); err != nil {
		t.Fatal("expected README.md to be generated")
	}
	if _, err := os.Stat(filepath.Join(outputDir, "Dockerfile")); err != nil {
		t.Fatal("expected Dockerfile to be generated")
	}
}

func TestE2EFullSpecWithModules(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	outputDir := filepath.Join(dir, "out")

	spec := `project: full-e2e
modules:
  - name: api
    path: ./internal/api
  - name: core
    path: ./internal/core
services:
  - name: web
    kind: http
    port: 8080
`
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: e2e-full\n  mode: development\n  output_dir: "+outputDir+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input-file", specPath})
	if err != nil {
		t.Fatalf("e2e full run failed: %v", err)
	}

	expectedFiles := []string{
		"README.md", "go.mod", "Dockerfile",
		".github/workflows/ci.yml",
		"internal/api/handler.go",
		"internal/api/repository.go",
		"internal/api/service.go",
		"internal/core/handler.go",
		"internal/core/domain/model.go",
		"internal/web/config.yaml",
	}
	for _, name := range expectedFiles {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Errorf("expected %s to exist", name)
		}
	}
}

func TestE2EValidateMinimalSpec(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: validate-e2e\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"validate", "--config", configPath, "--input", "project: validate-test\n"})
	if err != nil {
		t.Fatalf("validate e2e failed: %v", err)
	}
}

func TestE2ELintSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(specPath, []byte("project: lint-test\nmodules:\n  - name: core\n    path: ./internal/core\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"lint", "--input-file", specPath})
	if err != nil {
		t.Fatalf("lint e2e failed: %v", err)
	}
}

func TestE2EDiffWithNoOutputDir(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: diff-e2e\n  mode: development\n  output_dir: ./nonexistent\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("project: diff-test\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"diff", "--config", configPath, "--input-file", specPath})
	if err != nil {
		t.Fatalf("diff e2e failed: %v", err)
	}
}

func TestE2EMigratePlan(t *testing.T) {
	err := run([]string{"migrate", "plan", "--from", "0.1.0", "--to", "0.3.0"})
	if err != nil {
		t.Fatalf("migrate plan e2e failed: %v", err)
	}
}

func TestE2EProfile(t *testing.T) {
	err := run([]string{"profile"})
	if err != nil {
		t.Fatalf("profile e2e failed: %v", err)
	}
}

func TestE2EVersionJSON(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "version")
	if err != nil {
		t.Fatalf("execute version failed: %v", err)
	}
	if !strings.Contains(output, "naeos") {
		t.Fatalf("expected version output, got %q", output)
	}
}

func TestAutoDetectYmlExtension(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: yml-ext\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	root := newRootCommand()
	output, err := executeCommand(root, "status")
	if err != nil {
		t.Fatalf("execute status with .yml failed: %v", err)
	}
	if !strings.Contains(output, "yml-ext") {
		t.Fatalf("expected yml-ext, got %q", output)
	}
}

func TestAutoDetectNaeosYaml(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "naeos.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: naeos-yaml\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	root := newRootCommand()
	output, err := executeCommand(root, "status")
	if err != nil {
		t.Fatalf("execute status with naeos.yaml failed: %v", err)
	}
	if !strings.Contains(output, "naeos-yaml") {
		t.Fatalf("expected naeos-yaml, got %q", output)
	}
}
