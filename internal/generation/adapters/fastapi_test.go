package adapters

import (
	"strings"
	"testing"
)

func TestFastAPIAdapter_GenerateProject(t *testing.T) {
	a := FastAPIAdapter{}
	artifacts := a.GenerateProject("myproj")
	if len(artifacts) == 0 {
		t.Fatal("expected at least one artifact")
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	expected := []string{"README.md", "pyproject.toml", "myproj/__init__.py", "myproj/__main__.py", "myproj/app.py"}
	for _, p := range expected {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	// Check content of app.py contains FastAPI
	for _, a := range artifacts {
		if a.Path == "myproj/app.py" {
			if !strings.Contains(string(a.Content), "FastAPI") {
				t.Errorf("app.py should contain FastAPI")
			}
			break
		}
	}
}

func TestFastAPIAdapter_GenerateModule(t *testing.T) {
	a := FastAPIAdapter{}
	artifacts := a.GenerateModule("users", "./internal/users", "myproj")
	if len(artifacts) == 0 {
		t.Fatal("expected at least one artifact")
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	expected := []string{"src/users/__init__.py", "src/users/router.py", "src/users/service.py", "src/users/models.py", "tests/test_users.py"}
	for _, p := range expected {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	// Check router.py contains router and prefix
	for _, a := range artifacts {
		if a.Path == "src/users/router.py" {
			content := string(a.Content)
			if !strings.Contains(content, "APIRouter") {
				t.Errorf("router.py should contain APIRouter")
			}
			if !strings.Contains(content, "prefix=\"/users\"") {
				t.Errorf("router.py should contain prefix")
			}
			break
		}
	}
}

func TestFastAPIAdapter_GenerateService(t *testing.T) {
	a := FastAPIAdapter{}
	artifacts := a.GenerateService("api-gateway", "http", 8080, "myproj")
	if len(artifacts) == 0 {
		t.Fatal("expected at least one artifact")
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	// Expect __init__.py and server.py
	if !paths["src/services/api-gateway/__init__.py"] {
		t.Errorf("missing __init__.py")
	}
	if !paths["src/services/api-gateway/server.py"] {
		t.Errorf("missing server.py")
	}
	// Check server.py contains uvicorn and the port
	for _, a := range artifacts {
		if a.Path == "src/services/api-gateway/server.py" {
			content := string(a.Content)
			if !strings.Contains(content, "uvicorn") {
				t.Errorf("server.py should contain uvicorn")
			}
			if !strings.Contains(content, "port=8080") {
				t.Errorf("server.py should contain port 8080")
			}
			break
		}
	}
}

func TestFastAPIAdapter_GenerateDockerfile(t *testing.T) {
	a := FastAPIAdapter{}
	artifacts := a.GenerateDockerfile("myproj")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Path != "Dockerfile" {
		t.Errorf("expected path Dockerfile, got %s", artifacts[0].Path)
	}
	if !strings.Contains(string(artifacts[0].Content), "EXPOSE 8000") {
		t.Errorf("Dockerfile should expose port 8000")
	}
	if !strings.Contains(string(artifacts[0].Content), "uvicorn") {
		t.Errorf("Dockerfile should use uvicorn")
	}
}

func TestFastAPIAdapter_GenerateCI(t *testing.T) {
	a := FastAPIAdapter{}
	artifacts := a.GenerateCI("myproj")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Path != ".github/workflows/ci.yml" {
		t.Errorf("expected path .github/workflows/ci.yml, got %s", artifacts[0].Path)
	}
	if !strings.Contains(string(artifacts[0].Content), "pytest") {
		t.Errorf("CI should run pytest")
	}
}

func TestFastAPIAdapter_GenerateDockerCompose(t *testing.T) {
	a := FastAPIAdapter{}
	artifacts := a.GenerateDockerCompose("myproj")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Path != "docker-compose.yml" {
		t.Errorf("expected path docker-compose.yml, got %s", artifacts[0].Path)
	}
	if !strings.Contains(string(artifacts[0].Content), "8000:8000") {
		t.Errorf("docker-compose should map port 8000")
	}
}

func TestFastAPIAdapter_GenerateArchitectureDoc(t *testing.T) {
	a := FastAPIAdapter{}
	artifacts := a.GenerateArchitectureDoc("myproj", "hexagonal")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Path != "docs/architecture.md" {
		t.Errorf("expected path docs/architecture.md, got %s", artifacts[0].Path)
	}
	if !strings.Contains(string(artifacts[0].Content), "hexagonal") {
		t.Errorf("architecture doc should contain pattern")
	}
	if !strings.Contains(string(artifacts[0].Content), "myproj") {
		t.Errorf("architecture doc should contain project name")
	}
}
