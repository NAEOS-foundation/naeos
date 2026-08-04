package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
)

func TestDXCompletions(t *testing.T) {
	root := NewRootCommand()
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"dx", "completion-bash"}, "complete"},
		{[]string{"dx", "completion-zsh"}, "#compdef"},
		{[]string{"dx", "completion-powershell"}, "Register-ArgumentCompleter"},
	} {
		output, err := executeCommand(root, tc.args...)
		if err != nil {
			t.Fatalf("%v: %v", tc.args, err)
		}
		if !strings.Contains(output, tc.want) {
			t.Errorf("%v: expected %q in output, got %q", tc.args, tc.want, output)
		}
	}
}

func TestDXSnippetListAndGet(t *testing.T) {
	root := NewRootCommand()

	output, err := executeCommand(root, "dx", "snippet-list")
	if err != nil {
		t.Fatalf("snippet-list: %v", err)
	}
	if !strings.Contains(output, "SNIPPET") || !strings.Contains(output, "neir-spec") {
		t.Errorf("unexpected output: %q", output)
	}

	output, err = executeCommand(root, "dx", "snippet-get", "--name", "module")
	if err != nil {
		t.Fatalf("snippet-get: %v", err)
	}
	if !strings.Contains(output, "module-name") {
		t.Errorf("unexpected output: %q", output)
	}

	_, err = executeCommand(NewRootCommand(), "dx", "snippet-get", "--name", "missing")
	if err == nil {
		t.Error("expected error for unknown snippet")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = executeCommand(NewRootCommand(), "dx", "snippet-get")
	if err == nil {
		t.Error("expected error when --name is missing")
	}
}

func resetHTTPMux() {
	http.DefaultServeMux = http.NewServeMux()
}

func TestWebSocketBadInputFile(t *testing.T) {
	resetHTTPMux()
	root := NewRootCommand()
	_, err := executeCommand(root, "ws", "--port", ":0", "--input-file", "/nonexistent/spec.yaml")
	if err == nil {
		t.Fatal("expected error for missing input file")
	}
	if !strings.Contains(err.Error(), "read input file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebSocketBadConfig(t *testing.T) {
	resetHTTPMux()
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(spec, []byte("project: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := NewRootCommand()
	_, err := executeCommand(root, "ws", "--port", ":0", "--input-file", spec, "--config", "/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for bad config")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebSocketPortInUse(t *testing.T) {
	resetHTTPMux()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().String()

	root := NewRootCommand()
	_, err = executeCommand(root, "ws", "--port", port)
	if err == nil {
		t.Fatal("expected error when port already in use")
	}
	if !strings.Contains(err.Error(), "address already in use") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestArtifactsListEmpty(t *testing.T) {
	dir := t.TempDir()
	root := NewRootCommand()
	output, err := executeCommand(root, "artifacts", "--dir", dir, "list")
	if err != nil {
		t.Fatalf("artifacts list: %v", err)
	}
	if !strings.Contains(output, "No artifacts tracked.") {
		t.Errorf("unexpected output: %q", output)
	}
}

func seedArtifactStore(t *testing.T, dir string) {
	t.Helper()
	store := artifacts.NewStore(dir)
	if _, err := store.Add("src/main.go", []byte("package main\n"), artifacts.KindCode); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add("README.md", []byte("# Docs\n"), artifacts.KindDocs); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add("src/dupe.go", []byte("package main\n"), artifacts.KindCode); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteToDisk(); err != nil {
		t.Fatal(err)
	}
}

func TestArtifactsListInfoSummary(t *testing.T) {
	dir := t.TempDir()
	seedArtifactStore(t, dir)
	root := NewRootCommand()

	output, err := executeCommand(root, "artifacts", "--dir", dir, "list")
	if err != nil {
		t.Fatalf("artifacts list: %v", err)
	}
	if !strings.Contains(output, "src/main.go") || !strings.Contains(output, "code") {
		t.Errorf("unexpected list output: %q", output)
	}

	output, err = executeCommand(root, "artifacts", "--dir", dir, "info", "src/main.go")
	if err != nil {
		t.Fatalf("artifacts info: %v", err)
	}
	if !strings.Contains(output, "Path: src/main.go") || !strings.Contains(output, "Kind: code") {
		t.Errorf("unexpected info output: %q", output)
	}

	_, err = executeCommand(root, "artifacts", "--dir", dir, "info", "/nonexistent")
	if err == nil {
		t.Error("expected error for unknown artifact")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}

	output, err = executeCommand(root, "artifacts", "--dir", dir, "summary")
	if err != nil {
		t.Fatalf("artifacts summary: %v", err)
	}
	if !strings.Contains(output, "code") || !strings.Contains(output, "docs") {
		t.Errorf("unexpected summary output: %q", output)
	}
}

func writeDuplicateManifest(t *testing.T, dir string) {
	t.Helper()
	content := []byte("package main\n")
	hash := fmt.Sprintf("%x", sha256.Sum256(content))
	manifest := artifacts.StoreManifest{
		Version:   "1.0",
		Project:   "dup-test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Artifacts: []artifacts.Artifact{
			{
				ID:          "a",
				Path:        "src/main.go",
				ContentHash: hash,
				Kind:        artifacts.KindCode,
				Size:        int64(len(content)),
				CreatedAt:   time.Now(),
			},
			{
				ID:          "b",
				Path:        "src/dupe.go",
				ContentHash: hash,
				Kind:        artifacts.KindCode,
				Size:        int64(len(content)),
				CreatedAt:   time.Now(),
			},
		},
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".artifacts.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"src/main.go", "src/dupe.go"} {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestArtifactsDedup(t *testing.T) {
	dir := t.TempDir()
	writeDuplicateManifest(t, dir)
	root := NewRootCommand()

	output, err := executeCommand(root, "artifacts", "--dir", dir, "dedup")
	if err != nil {
		t.Fatalf("artifacts dedup: %v", err)
	}
	if !strings.Contains(output, "Removed 1 duplicate artifacts.") {
		t.Errorf("unexpected dedup output: %q", output)
	}

	output, err = executeCommand(NewRootCommand(), "artifacts", "--dir", dir, "dedup")
	if err != nil {
		t.Fatalf("second dedup: %v", err)
	}
	if !strings.Contains(output, "Removed 0 duplicate artifacts.") {
		t.Errorf("expected zero duplicates, got %q", output)
	}
}

func TestArtifactsBadManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".artifacts.json"), []byte("{{{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := NewRootCommand()
	_, err := executeCommand(root, "artifacts", "--dir", dir, "list")
	if err == nil {
		t.Fatal("expected error for corrupted manifest")
	}
}

func TestArtifactsDedupBadManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".artifacts.json"), []byte("{{{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := NewRootCommand()
	_, err := executeCommand(root, "artifacts", "--dir", dir, "summary")
	if err == nil {
		t.Fatal("expected error for corrupted manifest")
	}
}
