package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/profiling"
)

func TestRunBuildDistributedHappy(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: dist-build\n  mode: development\n  output_dir: ./out\n")
	inputFile := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: dist\nmodules:\n  - name: core\n    path: ./core\n")

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := runBuildDistributed(cmd, configPath, "", inputFile, 2)
	if err != nil {
		t.Fatalf("runBuildDistributed failed: %v", err)
	}
}

func TestRunBuildDistributedMissingInput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: dist\n  mode: development\n")

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})

	err := runBuildDistributed(cmd, configPath, "", filepath.Join(dir, "missing.yaml"), 1)
	if err == nil {
		t.Fatal("expected error for missing input file")
	}
}

func TestRunBuildDistributedInvalidSpec(t *testing.T) {
	dir := t.TempDir()
	inputFile := filepath.Join(dir, "bad.yaml")
	writeTestFile(t, dir, "bad.yaml", "{{{{invalid")

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})

	err := runBuildDistributed(cmd, "/nonexistent/config.yaml", "", inputFile, 1)
	if err == nil {
		t.Fatal("expected error for invalid spec")
	}
}
func TestRunBuildDistributedInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	inputFile := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: dist\nmodules:\n  - name: core\n    path: ./core\n")

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})

	// Workers fail per-task; the function itself returns no error.
	err := runBuildDistributed(cmd, "/nonexistent/config.yaml", "", inputFile, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveMemProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memprofile.json")

	mp := profiling.NewMemProfiler()
	mp.Snapshot("validate")
	mp.Snapshot("generate")

	if err := saveMemProfile(path, mp); err != nil {
		t.Fatalf("saveMemProfile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "snapshots") {
		t.Error("expected snapshots key in output")
	}
	if !strings.Contains(string(data), "validate") {
		t.Error("expected validate label in output")
	}
}

func TestSaveMemProfileErrorPath(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	mp := profiling.NewMemProfiler()
	mp.Snapshot("start")

	err := saveMemProfile(filepath.Join(blocker, "out.json"), mp)
	if err == nil {
		t.Fatal("expected error when target path parent is a file")
	}
}

func TestParseSpecInputModules(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"yaml modules", "project: x\nmodules:\n  - name: a\n  - name: b\n  - name: c\n", 3},
		{"json modules", `{"modules":[{"name":"a"},{"name":"b"}]}`, 2},
		{"no modules", "project: x\n", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parseSpecInput(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			modules, _ := spec["modules"].([]any)
			if len(modules) != tt.expect {
				t.Errorf("expected %d modules, got %d", tt.expect, len(modules))
			}
		})
	}
}
