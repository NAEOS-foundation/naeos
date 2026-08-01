package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCICDGitHubDefault(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "ci.yml")

	root := NewRootCommand()
	_, err := executeCommand(root, "cicd", "--project", "myapp", "--output", out)
	if err != nil {
		t.Fatalf("cicd github failed: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "CI/CD Pipeline") || !strings.Contains(content, "jobs:") {
		t.Fatalf("expected GitHub Actions workflow, got %q", content)
	}
}

func TestCICDGitLab(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "gitlab-ci.yml")

	root := NewRootCommand()
	_, err := executeCommand(root, "cicd", "--platform", "gitlab", "--languages", "python", "--output", out)
	if err != nil {
		t.Fatalf("cicd gitlab failed: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if !strings.Contains(string(data), "stages:") {
		t.Fatalf("expected GitLab CI stages, got %q", string(data))
	}
}

func TestCICDJenkins(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "Jenkinsfile")

	root := NewRootCommand()
	_, err := executeCommand(root, "cicd", "--platform", "jenkins", "--output", out)
	if err != nil {
		t.Fatalf("cicd jenkins failed: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty Jenkinsfile")
	}
}

func TestCICDUnsupportedPlatform(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "cicd", "--platform", "circleci")
	if err == nil {
		t.Fatal("expected error for unsupported platform")
	}
}

func TestCICDNoOutputFlag(t *testing.T) {
	root := NewRootCommand()
	if _, err := executeCommand(root, "cicd", "--project", "myapp"); err != nil {
		t.Fatalf("cicd without --output failed: %v", err)
	}
}

func TestCICDInputFileYAML(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	out := filepath.Join(dir, "ci.yml")
	writeTestFile(t, dir, "spec.yaml", "project: from-file\nlanguages: [python]\n")

	root := NewRootCommand()
	if _, err := executeCommand(root, "cicd", "--input-file", spec, "--output", out); err != nil {
		t.Fatalf("cicd with yaml input failed: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if !strings.Contains(string(data), "Set up Python") {
		t.Fatalf("expected python language from input file, got %q", string(data))
	}
}

func TestCICDInputFileJSON(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.json")
	out := filepath.Join(dir, "ci.yml")
	writeTestFile(t, dir, "spec.json", `{"project":"jsonproj","languages":["go"]}`)

	root := NewRootCommand()
	if _, err := executeCommand(root, "cicd", "--input-file", spec, "--output", out); err != nil {
		t.Fatalf("cicd with json input failed: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected generated workflow content")
	}
}

func TestCICDInputFileMissing(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "cicd", "--input-file", "/nonexistent/spec.yaml")
	if err == nil {
		t.Fatal("expected error for missing input file")
	}
}

func TestCICDInputFileInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "not: [valid: yaml")

	root := NewRootCommand()
	_, err := executeCommand(root, "cicd", "--input-file", spec)
	if err == nil {
		t.Fatal("expected error for invalid yaml input")
	}
}
