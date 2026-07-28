package adapters

import (
	"strings"
	"testing"
)

func TestPythonAdapter_GenerateProject(t *testing.T) {
	t.Parallel()
	a := PythonAdapter{}
	artifacts := a.GenerateProject("MyProject")
	if len(artifacts) < 4 {
		t.Fatalf("expected at least 4 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{"README.md", "pyproject.toml"} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	for _, a := range artifacts {
		if a.Path == "pyproject.toml" {
			content := string(a.Content)
			if !strings.Contains(content, "setuptools>=68.0") {
				t.Errorf("pyproject.toml should contain setuptools")
			}
			if !strings.Contains(content, "requires-python") {
				t.Errorf("pyproject.toml should contain requires-python")
			}
		}
	}
}

func TestPythonAdapter_GenerateModule(t *testing.T) {
	t.Parallel()
	a := PythonAdapter{}
	artifacts := a.GenerateModule("users", "./internal/users", "MyProject")
	if len(artifacts) < 6 {
		t.Fatalf("expected at least 6 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{
		"src/users/__init__.py",
		"src/users/handler.py",
		"src/users/service.py",
		"src/users/repository.py",
		"src/users/models.py",
		"tests/test_users.py",
	} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
}

func TestPythonAdapter_GenerateService(t *testing.T) {
	t.Parallel()
	a := PythonAdapter{}
	artifacts := a.GenerateService("api-gateway", "http", 8080, "MyProject")
	if len(artifacts) < 1 {
		t.Fatalf("expected at least 1 artifact, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	if !paths["src/services/api-gateway/__init__.py"] {
		t.Error("missing __init__.py")
	}
	if !paths["src/services/api-gateway/server.py"] {
		t.Error("missing server.py")
	}
}

func TestPythonAdapter_GenerateDockerfile(t *testing.T) {
	t.Parallel()
	a := PythonAdapter{}
	artifacts := a.GenerateDockerfile("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "python:3.12-slim") {
		t.Errorf("Dockerfile should use python:3.12-slim")
	}
	if !strings.Contains(content, "pip install") {
		t.Errorf("Dockerfile should use pip install")
	}
}

func TestPythonAdapter_GenerateCI(t *testing.T) {
	t.Parallel()
	a := PythonAdapter{}
	artifacts := a.GenerateCI("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "actions/setup-python@v5") {
		t.Errorf("CI should use actions/setup-python@v5")
	}
	if !strings.Contains(content, "pytest") {
		t.Errorf("CI should run pytest")
	}
}

func TestPythonAdapter_GenerateDockerCompose(t *testing.T) {
	t.Parallel()
	a := PythonAdapter{}
	artifacts := a.GenerateDockerCompose("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "8000:8000") {
		t.Errorf("docker-compose should map port 8000")
	}
}

func TestPythonAdapter_GenerateArchitectureDoc(t *testing.T) {
	t.Parallel()
	a := PythonAdapter{}
	artifacts := a.GenerateArchitectureDoc("MyProject", "clean")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "clean") {
		t.Errorf("architecture doc should contain pattern")
	}
	if !strings.Contains(content, "Python") {
		t.Errorf("architecture doc should contain Python")
	}
}
