package profiledetect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/foo\ngo 1.22"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Language != "go" {
		t.Errorf("expected go, got %s", result.Language)
	}
	if result.Confidence <= 0 {
		t.Error("expected positive confidence")
	}
}

func TestDetectTypeScript(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies": {"typescript": "^5.0"}}`), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Language != "typescript" {
		t.Errorf("expected typescript, got %s", result.Language)
	}
}

func TestDetectPython(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask==3.0"), 0644)
	os.WriteFile(filepath.Join(dir, "app.py"), []byte("from flask import Flask"), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Language != "python" {
		t.Errorf("expected python, got %s", result.Language)
	}
}

func TestDetectRust(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"myapp\""), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Language != "rust" {
		t.Errorf("expected rust, got %s", result.Language)
	}
}

func TestDetectJava(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project></project>"), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Language != "java" {
		t.Errorf("expected java, got %s", result.Language)
	}
}

func TestDetectFrameworkReact(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies": {"react": "^18.0"}}`), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Framework != "react" {
		t.Errorf("expected react framework, got %s", result.Framework)
	}
}

func TestDetectFrameworkNextjs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies": {"next": "^14.0", "react": "^18.0"}}`), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Framework != "nextjs" {
		t.Errorf("expected nextjs framework, got %s", result.Framework)
	}
}

func TestDetectFrameworkDjango(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[tool.poetry]\n[tool.django]\n"), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Framework != "django" {
		t.Errorf("expected django framework, got %s", result.Framework)
	}
}

func TestDetectFrameworkGin(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/foo\nrequire github.com/gin-gonic/gin v1.9.0"), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Framework != "gin" {
		t.Errorf("expected gin framework, got %s", result.Framework)
	}
}

func TestDetectEmpty(t *testing.T) {
	dir := t.TempDir()
	d := NewDetector(dir)
	result := d.Detect()

	if result.Language != "unknown" {
		t.Errorf("expected unknown, got %s", result.Language)
	}
	if result.Confidence != 0 {
		t.Error("expected 0 confidence for unknown")
	}
}

func TestDetectMultipleSignals(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/foo"), 0644)
	os.WriteFile(filepath.Join(dir, "go.sum"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	d := NewDetector(dir)
	result := d.Detect()

	if result.Language != "go" {
		t.Errorf("expected go, got %s", result.Language)
	}
	if len(result.Files) < 2 {
		t.Errorf("expected at least 2 matched files, got %d", len(result.Files))
	}
}
