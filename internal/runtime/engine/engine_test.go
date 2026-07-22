package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine()
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestRunNilArtifact(t *testing.T) {
	e := NewEngine()
	err := e.Run(nil)
	if err == nil {
		t.Fatal("expected error for nil artifact")
	}
}

func TestRunValidArtifact(t *testing.T) {
	e := NewEngine()
	err := e.Run("something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteEmptyPath(t *testing.T) {
	e := NewEngine()
	_, err := e.Execute(Artifact{Path: "", Content: []byte("test")})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestExecute(t *testing.T) {
	e := NewEngine()
	result, err := e.Execute(Artifact{Path: "test.go", Content: []byte("package main")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed status, got %s", result.Status)
	}
	if e.ExecutedCount() != 1 {
		t.Fatalf("expected 1 executed artifact, got %d", e.ExecutedCount())
	}
}

func TestExecuteDuplicate(t *testing.T) {
	e := NewEngine()
	_, _ = e.Execute(Artifact{Path: "test.go", Content: []byte("package main")})
	result, err := e.Execute(Artifact{Path: "test.go", Content: []byte("package main")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "skipped" {
		t.Fatalf("expected skipped status for duplicate, got %s", result.Status)
	}
	if e.ExecutedCount() != 1 {
		t.Fatalf("expected 1 executed artifact, got %d", e.ExecutedCount())
	}
}

func TestHistory(t *testing.T) {
	e := NewEngine()
	_, _ = e.Execute(Artifact{Path: "a.go", Content: []byte("package a")})
	_, _ = e.Execute(Artifact{Path: "b.go", Content: []byte("package b")})

	history := e.History()
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
}

func TestReset(t *testing.T) {
	e := NewEngine()
	_, _ = e.Execute(Artifact{Path: "test.go", Content: []byte("package main")})
	e.Reset()

	if e.ExecutedCount() != 0 {
		t.Fatalf("expected 0 executed after reset, got %d", e.ExecutedCount())
	}
	history := e.History()
	if len(history) != 0 {
		t.Fatalf("expected 0 history after reset, got %d", len(history))
	}
}

func TestValidateGoFileMissingPackage(t *testing.T) {
	e := NewEngine()
	err := e.Validate(Artifact{Path: "test.go", Content: []byte("fmt.Println")})
	if err == nil {
		t.Fatal("expected error for go file missing package declaration")
	}
}

func TestValidateGoFileEmptyContent(t *testing.T) {
	e := NewEngine()
	err := e.Validate(Artifact{Path: "test.go", Content: []byte{}})
	if err == nil {
		t.Fatal("expected error for empty go file")
	}
}

func TestValidateYamlFile(t *testing.T) {
	e := NewEngine()
	err := e.Validate(Artifact{Path: "config.yaml", Content: []byte("key: value")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateEmptyPath(t *testing.T) {
	e := NewEngine()
	err := e.Validate(Artifact{Path: "", Content: []byte("test")})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestExecuteAll(t *testing.T) {
	e := NewEngine()
	artifacts := []Artifact{
		{Path: "a.go", Content: []byte("package a")},
		{Path: "b.go", Content: []byte("package b")},
		{Path: "c.yaml", Content: []byte("key: value")},
	}
	results, err := e.ExecuteAll(artifacts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != "completed" {
			t.Fatalf("expected completed status for %s, got %s", r.Artifact.Path, r.Status)
		}
	}
}

func TestExecuteAllEmpty(t *testing.T) {
	e := NewEngine()
	_, err := e.ExecuteAll([]Artifact{})
	if err == nil {
		t.Fatal("expected error for empty artifacts")
	}
}

func TestExecuteAllValidationFailure(t *testing.T) {
	e := NewEngine()
	artifacts := []Artifact{
		{Path: "a.go", Content: []byte("package a")},
		{Path: "b.go", Content: []byte("no package")},
	}
	_, err := e.ExecuteAll(artifacts)
	if err == nil {
		t.Fatal("expected error for validation failure")
	}
}

func TestFailedCount(t *testing.T) {
	e := NewEngine()
	_, _ = e.Execute(Artifact{Path: "a.go", Content: []byte("package a")})
	if e.FailedCount() != 0 {
		t.Fatalf("expected 0 failed, got %d", e.FailedCount())
	}
}

func TestExecuteWritesFile(t *testing.T) {
	dir := t.TempDir()
	e := NewEngine()
	e.SetOutputDir(dir)

	result, err := e.Execute(Artifact{Path: "test.txt", Content: []byte("hello")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed status, got %s", result.Status)
	}
	if !strings.Contains(result.Output, dir) {
		t.Fatalf("expected output to contain dir, got %s", result.Output)
	}

	data, err := os.ReadFile(filepath.Join(dir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}
}

func TestExecuteWritesNestedFile(t *testing.T) {
	dir := t.TempDir()
	e := NewEngine()
	e.SetOutputDir(dir)

	result, err := e.Execute(Artifact{Path: "internal/app/main.go", Content: []byte("package main")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed status, got %s", result.Status)
	}

	data, err := os.ReadFile(filepath.Join(dir, "internal/app/main.go"))
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != "package main" {
		t.Fatalf("expected 'package main', got '%s'", string(data))
	}
}
