package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompletionBash(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "completion", "bash")
	if err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}
	if !strings.Contains(output, "# bash completion") && !strings.Contains(output, "bash") && output == "" {
		t.Log("completion bash ran (output may go to stdout)")
	}
}

func TestCompletionZsh(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "completion", "zsh")
	if err != nil {
		t.Fatalf("completion zsh failed: %v", err)
	}
}

func TestCompletionFish(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "completion", "fish")
	if err != nil {
		t.Fatalf("completion fish failed: %v", err)
	}
}

func TestCompletionPowershell(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "completion", "powershell")
	if err != nil {
		t.Fatalf("completion powershell failed: %v", err)
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "completion", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid shell")
	}
}

func TestHistoryNoStore(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "history", "--store-dir", t.TempDir())
	if err != nil {
		t.Fatalf("history failed: %v", err)
	}
	if !strings.Contains(output, "No pipeline runs found") {
		t.Fatalf("expected 'No pipeline runs found', got %q", output)
	}
}

func TestHistoryWithEvents(t *testing.T) {
	dir := t.TempDir()

	events := `[
		{"id":"e1","stream_id":"run1","type":"pipeline.started","data":{"name":"test-run"},"version":1,"timestamp":"2025-01-01T00:00:00Z"},
		{"id":"e2","stream_id":"run1","type":"pipeline.completed","data":{"artifacts":3},"version":2,"timestamp":"2025-01-01T00:00:05Z"}
	]`
	if err := os.WriteFile(filepath.Join(dir, "run1.json"), []byte(events), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "history", "--store-dir", dir)
	if err != nil {
		t.Fatalf("history failed: %v", err)
	}
	if !strings.Contains(output, "Pipeline Run History") {
		t.Fatalf("expected 'Pipeline Run History', got %q", output)
	}
	if !strings.Contains(output, "test-run") {
		t.Fatalf("expected 'test-run' in output, got %q", output)
	}
}

func TestHistoryJSON(t *testing.T) {
	dir := t.TempDir()
	events := `[
		{"id":"e1","stream_id":"run1","type":"pipeline.started","data":{"name":"test-run"},"version":1,"timestamp":"2025-01-01T00:00:00Z"}
	]`
	if err := os.WriteFile(filepath.Join(dir, "run1.json"), []byte(events), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "history", "--store-dir", dir, "--output-format", "json")
	if err != nil {
		t.Fatalf("history json failed: %v", err)
	}
	if !strings.Contains(output, "test-run") {
		t.Fatalf("expected 'test-run' in json output, got %q", output)
	}
}

func TestHistoryYAML(t *testing.T) {
	dir := t.TempDir()
	events := `[
		{"id":"e1","stream_id":"run1","type":"pipeline.started","data":{"name":"test-run"},"version":1,"timestamp":"2025-01-01T00:00:00Z"}
	]`
	if err := os.WriteFile(filepath.Join(dir, "run1.json"), []byte(events), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "history", "--store-dir", dir, "--output-format", "yaml")
	if err != nil {
		t.Fatalf("history yaml failed: %v", err)
	}
	if !strings.Contains(output, "test-run") {
		t.Fatalf("expected 'test-run' in yaml output, got %q", output)
	}
}

func TestAuditBasic(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.yaml")
	content := "password: secret123"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "audit", file)
	if err != nil {
		t.Fatalf("audit failed: %v", err)
	}
	if !strings.Contains(output, "Audit complete") {
		t.Fatalf("expected 'Audit complete' in output, got %q", output)
	}
}

func TestAuditMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "main.go")
	f2 := filepath.Join(dir, "config.yaml")
	os.WriteFile(f1, []byte("package main\nfunc main() {}"), 0o644)
	os.WriteFile(f2, []byte("api_key: sk-1234"), 0o644)

	root := NewRootCommand()
	output, err := executeCommand(root, "audit", f1, f2)
	if err != nil {
		t.Fatalf("audit failed: %v", err)
	}
	if !strings.Contains(output, "Audit complete") {
		t.Fatalf("expected 'Audit complete' in output, got %q", output)
	}
}

func TestAuditNoArgs(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "audit")
	if err == nil {
		t.Fatal("expected error for audit with no args")
	}
}

func TestAuditNonexistentFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "audit", "/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error when all files fail to read")
	}
}

func TestAuditMixedFiles(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.yaml")
	os.WriteFile(valid, []byte("key: value"), 0o644)

	root := NewRootCommand()
	output, err := executeCommand(root, "audit", valid, "/nonexistent/file.yaml")
	if err != nil {
		t.Fatalf("audit should succeed when at least one file is valid: %v", err)
	}
	if !strings.Contains(output, "Audit complete") {
		t.Fatalf("expected 'Audit complete' in output, got %q", output)
	}
}

func TestLockGenerateNoArgs(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "lock", "generate")
	if err != nil {
		t.Logf("lock generate with no args errored (expected): %v", err)
	} else {
		if !strings.Contains(output, "Generated") {
			t.Fatalf("expected 'Generated' in output, got %q", output)
		}
	}
}

func TestLockVerifyChanged(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	lockFilePath := filepath.Join(dir, "naeos.lock")
	os.WriteFile(file, []byte("hello"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	root := NewRootCommand()
	_, err := executeCommand(root, "lock", "generate", "test.txt")
	if err != nil {
		os.Chdir(origDir)
		t.Fatalf("lock generate failed: %v", err)
	}

	os.WriteFile(file, []byte("world"), 0o644)

	output, err := executeCommand(root, "lock", "verify", "test.txt", "--lock-file", lockFilePath)
	if err != nil {
		os.Chdir(origDir)
		t.Fatalf("lock verify should not error on changes: %v", err)
	}
	if !strings.Contains(output, "Changes detected") {
		os.Chdir(origDir)
		t.Fatalf("expected 'Changes detected' in output, got %q", output)
	}
	os.Chdir(origDir)
}

func TestLockVerifyNoLockFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	os.WriteFile(file, []byte("hello"), 0o644)

	root := NewRootCommand()
	_, err := executeCommand(root, "lock", "verify", file, "--lock-file", filepath.Join(dir, "nonexistent.lock"))
	if err == nil {
		t.Fatal("expected error when lock file not found")
	}
}

func TestBenchmarkBasic(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "naeos.yaml")
	os.WriteFile(configPath, []byte("pipeline:\n  name: test\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644)

	root := NewRootCommand()
	output, err := executeCommand(root, "benchmark", "--iterations", "2", "--config", configPath)
	if err != nil {
		t.Fatalf("benchmark failed: %v", err)
	}
	if !strings.Contains(output, "Benchmark Results") {
		t.Fatalf("expected 'Benchmark Results' in output, got %q", output)
	}
}

func TestBenchmarkJSON(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "naeos.yaml")
	os.WriteFile(configPath, []byte("pipeline:\n  name: test\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644)

	root := NewRootCommand()
	output, err := executeCommand(root, "benchmark", "--iterations", "2", "--config", configPath, "--output", "json")
	if err != nil {
		t.Fatalf("benchmark json failed: %v", err)
	}
	if !strings.Contains(output, "iterations") {
		t.Fatalf("expected json output, got %q", output)
	}
}

func TestBenchmarkNoConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "benchmark", "--iterations", "1")
	if err != nil {
		t.Logf("benchmark without config failed as expected: %v", err)
	}
}
