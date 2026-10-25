// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package clihelpers

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create dir for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestLoadInputWithInputFlag(t *testing.T) {
	result, err := LoadInput("hello spec", "")
	if err != nil {
		t.Fatalf("LoadInput returned error: %v", err)
	}
	if result != "hello spec" {
		t.Fatalf("expected 'hello spec', got %q", result)
	}
}

func TestLoadInputWithInputFileFlag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(path, []byte("project: file-test"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	result, err := LoadInput("", path)
	if err != nil {
		t.Fatalf("LoadInput returned error: %v", err)
	}
	if result != "project: file-test" {
		t.Fatalf("expected file content, got %q", result)
	}
}

func TestLoadInputMissingBothFlags(t *testing.T) {
	_, err := LoadInput("", "")
	if err == nil {
		t.Fatal("expected error when both flags are empty")
	}
}

func TestLoadInputBothFlagsProvided(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(path, []byte("project: file"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	_, err := LoadInput("inline spec", path)
	if err == nil {
		t.Fatal("expected error when both flags are provided")
	}
}

func TestLoadInputNonexistentFile(t *testing.T) {
	_, err := LoadInput("", "/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestResolveInputEmpty(t *testing.T) {
	result, err := ResolveInput("")
	if err != nil {
		t.Fatalf("ResolveInput returned error: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}

func TestResolveInputNonFile(t *testing.T) {
	result, err := ResolveInput("project: inline")
	if err != nil {
		t.Fatalf("ResolveInput returned error: %v", err)
	}
	if result != "project: inline" {
		t.Fatalf("expected passthrough, got %q", result)
	}
}

func TestResolveInputReadsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(path, []byte("project: from-file"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	result, err := ResolveInput(path)
	if err != nil {
		t.Fatalf("ResolveInput returned error: %v", err)
	}
	if result != "project: from-file" {
		t.Fatalf("expected file content, got %q", result)
	}
}

func TestResolveInputDirectory(t *testing.T) {
	dir := t.TempDir()
	result, err := ResolveInput(dir)
	if err != nil {
		t.Fatalf("ResolveInput returned error: %v", err)
	}
	if result != dir {
		t.Fatalf("expected directory path passthrough, got %q", result)
	}
}

func TestRenderOutputJSON(t *testing.T) {
	data := map[string]string{"key": "val"}
	result, err := RenderOutput(data, "json", nil)
	if err != nil {
		t.Fatalf("RenderOutput json returned error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestRenderOutputYAML(t *testing.T) {
	data := map[string]string{"key": "val"}
	result, err := RenderOutput(data, "yaml", nil)
	if err != nil {
		t.Fatalf("RenderOutput yaml returned error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestRenderOutputDefault(t *testing.T) {
	data := "default content"
	result, err := RenderOutput(data, "text", func() []byte {
		return []byte(data)
	})
	if err != nil {
		t.Fatalf("RenderOutput default returned error: %v", err)
	}
	if string(result) != "default content" {
		t.Fatalf("expected 'default content', got %q", string(result))
	}
}

func TestWriteFileInDir(t *testing.T) {
	dir := t.TempDir()
	err := WriteFileInDir(dir, "sub/file.txt", "hello")
	if err != nil {
		t.Fatalf("WriteFileInDir returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "sub", "file.txt"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(content))
	}
}

func TestResolveConfigPathExplicit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("pipeline:\n  name: test"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	resolved, err := ResolveConfigPath(path)
	if err != nil {
		t.Fatalf("ResolveConfigPath returned error: %v", err)
	}
	if resolved != path {
		t.Fatalf("expected %q, got %q", path, resolved)
	}
}

func TestResolveConfigPathMissing(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	_, err := ResolveConfigPath("")
	if err == nil {
		t.Fatal("expected error when no config found")
	}
}
