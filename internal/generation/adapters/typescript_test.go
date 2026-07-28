package adapters

import (
	"strings"
	"testing"
)

func TestTypeScriptAdapter_GenerateProject(t *testing.T) {
	t.Parallel()
	a := TypeScriptAdapter{}
	artifacts := a.GenerateProject("MyProject")
	if len(artifacts) < 4 {
		t.Fatalf("expected at least 4 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{"README.md", "package.json", "tsconfig.json", "src/index.ts"} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	for _, a := range artifacts {
		if a.Path == "package.json" {
			content := string(a.Content)
			if !strings.Contains(content, "typescript") {
				t.Errorf("package.json should contain typescript")
			}
			if !strings.Contains(content, "vitest") {
				t.Errorf("package.json should contain vitest")
			}
		}
	}
}

func TestTypeScriptAdapter_GenerateModule(t *testing.T) {
	t.Parallel()
	a := TypeScriptAdapter{}
	artifacts := a.GenerateModule("users", "./internal/users", "MyProject")
	if len(artifacts) < 6 {
		t.Fatalf("expected at least 6 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{
		"src/users/index.ts",
		"src/users/handler.ts",
		"src/users/service.ts",
		"src/users/repository.ts",
		"src/users/types.ts",
		"src/users/handler.test.ts",
	} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
}

func TestTypeScriptAdapter_GenerateService(t *testing.T) {
	t.Parallel()
	a := TypeScriptAdapter{}
	artifacts := a.GenerateService("api-gateway", "http", 8080, "MyProject")
	if len(artifacts) < 1 {
		t.Fatalf("expected at least 1 artifact, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	if !paths["src/services/api-gateway/index.ts"] {
		t.Error("missing index.ts")
	}
	if !paths["src/services/api-gateway/server.ts"] {
		t.Error("missing server.ts")
	}
}

func TestTypeScriptAdapter_GenerateDockerfile(t *testing.T) {
	t.Parallel()
	a := TypeScriptAdapter{}
	artifacts := a.GenerateDockerfile("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "node:22-alpine") {
		t.Errorf("Dockerfile should use node:22-alpine")
	}
	if !strings.Contains(content, "npm ci") {
		t.Errorf("Dockerfile should use npm ci")
	}
}

func TestTypeScriptAdapter_GenerateCI(t *testing.T) {
	t.Parallel()
	a := TypeScriptAdapter{}
	artifacts := a.GenerateCI("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "actions/setup-node@v4") {
		t.Errorf("CI should use actions/setup-node@v4")
	}
	if !strings.Contains(content, "npm test") {
		t.Errorf("CI should run npm test")
	}
}

func TestTypeScriptAdapter_GenerateDockerCompose(t *testing.T) {
	t.Parallel()
	a := TypeScriptAdapter{}
	artifacts := a.GenerateDockerCompose("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "3000:3000") {
		t.Errorf("docker-compose should map port 3000")
	}
}

func TestTypeScriptAdapter_GenerateArchitectureDoc(t *testing.T) {
	t.Parallel()
	a := TypeScriptAdapter{}
	artifacts := a.GenerateArchitectureDoc("MyProject", "clean")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "clean") {
		t.Errorf("architecture doc should contain pattern")
	}
	if !strings.Contains(content, "TypeScript") {
		t.Errorf("architecture doc should contain TypeScript")
	}
}
