package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileRunHappyPath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	profileOut := filepath.Join(dir, "profile.json")
	memOut := filepath.Join(dir, "mem.json")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: "+filepath.Join(dir, "out")+"\n")
	writeTestFile(t, dir, "spec.yaml", "project: profile-demo\nmodules:\n  - name: api\n    path: ./internal/api\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "profile", "run", "--config", configPath, "--input-file", specPath, "--profile", profileOut, "--memprofile", memOut)
	if err != nil {
		t.Fatalf("profile run failed: %v", err)
	}
	if !strings.Contains(output, "Pipeline complete:") {
		t.Fatalf("expected pipeline summary, got %q", output)
	}
	if !strings.Contains(output, "Memory Profile Summary") {
		t.Fatalf("expected memory analysis output, got %q", output)
	}
	for _, name := range []string{profileOut, memOut} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if len(data) == 0 {
			t.Fatalf("expected %s to contain content", name)
		}
	}
}

func TestProfileRunNoInput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n")

	root := NewRootCommand()
	_, err := executeCommand(root, "profile", "run", "--config", configPath)
	if err == nil || !strings.Contains(err.Error(), "missing required --input or --input-file") {
		t.Fatalf("expected missing input error, got %v", err)
	}
}

func TestProfileRunBothInputs(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n")
	writeTestFile(t, dir, "spec.yaml", "project: both-demo\n")

	root := NewRootCommand()
	_, err := executeCommand(root, "profile", "run", "--config", configPath, "--input", "project: inline", "--input-file", specPath)
	if err == nil || !strings.Contains(err.Error(), "cannot use both") {
		t.Fatalf("expected both-inputs error, got %v", err)
	}
}

func TestProfileRunMissingInputFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n")

	root := NewRootCommand()
	_, err := executeCommand(root, "profile", "run", "--config", configPath, "--input-file", filepath.Join(dir, "missing.yaml"))
	if err == nil || !strings.Contains(err.Error(), "read input file") {
		t.Fatalf("expected missing input file error, got %v", err)
	}
}
