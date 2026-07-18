package adapters

import (
	"strings"
	"testing"
)

func TestGoAdapter_GenerateProject(t *testing.T) {
	a := GoAdapter{}
	artifacts := a.GenerateProject("MyProject")
	if len(artifacts) < 4 {
		t.Fatalf("expected at least 4 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{"README.md", "go.mod", ".gitignore", "cmd/app/main.go"} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	for _, a := range artifacts {
		if a.Path == "go.mod" {
			content := string(a.Content)
			if !strings.Contains(content, "module github.com/example/") {
				t.Errorf("go.mod should contain module declaration")
			}
			if !strings.Contains(content, "go 1.22") {
				t.Errorf("go.mod should specify go version")
			}
		}
	}
}

func TestGoAdapter_GenerateModule(t *testing.T) {
	a := GoAdapter{}
	artifacts := a.GenerateModule("users", "./internal/users", "MyProject")
	if len(artifacts) < 8 {
		t.Fatalf("expected at least 8 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{
		"internal/users/handler.go",
		"internal/users/repository.go",
		"internal/users/service.go",
		"internal/users/domain/model.go",
		"internal/users/http/handler.go",
		"internal/users/http/router.go",
		"internal/users/middleware/logging.go",
		"internal/users/config/config.go",
		"internal/users/config/load.go",
		"internal/users/handler_test.go",
	} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
}

func TestGoAdapter_GenerateService(t *testing.T) {
	a := GoAdapter{}
	artifacts := a.GenerateService("api-gateway", "http", 8080, "MyProject")
	if len(artifacts) < 1 {
		t.Fatalf("expected at least 1 artifact, got %d", len(artifacts))
	}
	found := false
	for _, art := range artifacts {
		if strings.Contains(art.Path, "server.go") {
			found = true
			content := string(art.Content)
			if !strings.Contains(content, "net/http") {
				t.Errorf("server.go should use net/http")
			}
			if !strings.Contains(content, "api-gateway") {
				t.Errorf("server.go should contain service name")
			}
			if !strings.Contains(content, "Copyright 2026 NAEOS Foundation") {
				t.Errorf("server.go should include license header")
			}
		}
		if strings.Contains(art.Path, "server_test.go") {
			content := string(art.Content)
			if !strings.Contains(content, "Copyright 2026 NAEOS Foundation") {
				t.Errorf("server_test.go should include license header")
			}
		}
	}
	if !found {
		t.Error("expected server.go for http service")
	}
}

func TestGoAdapter_GenerateServiceNonHTTP(t *testing.T) {
	a := GoAdapter{}
	artifacts := a.GenerateService("worker", "grpc", 9090, "MyProject")
	if len(artifacts) != 1 {
		t.Errorf("expected 1 artifact (config.yaml) for non-http service, got %d", len(artifacts))
	}
}

func TestGoAdapter_GenerateDockerfile(t *testing.T) {
	a := GoAdapter{}
	artifacts := a.GenerateDockerfile("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "golang:1.22-alpine") {
		t.Errorf("Dockerfile should use golang:1.22-alpine")
	}
	if !strings.Contains(content, "go build") {
		t.Errorf("Dockerfile should use go build")
	}
}

func TestGoAdapter_GenerateCI(t *testing.T) {
	a := GoAdapter{}
	artifacts := a.GenerateCI("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "actions/setup-go@v5") {
		t.Errorf("CI should use actions/setup-go@v5")
	}
	if !strings.Contains(content, "go test") {
		t.Errorf("CI should run go test")
	}
}

func TestGoAdapter_GenerateDockerCompose(t *testing.T) {
	a := GoAdapter{}
	artifacts := a.GenerateDockerCompose("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if strings.Contains(content, "version:") {
		t.Errorf("docker-compose should not contain deprecated version key")
	}
	if !strings.Contains(content, "8080:8080") {
		t.Errorf("docker-compose should map port 8080")
	}
}

func TestGoAdapter_GenerateArchitectureDoc(t *testing.T) {
	a := GoAdapter{}
	artifacts := a.GenerateArchitectureDoc("MyProject", "hexagonal")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "hexagonal") {
		t.Errorf("architecture doc should contain pattern")
	}
	if !strings.Contains(content, "MyProject") {
		t.Errorf("architecture doc should contain project name")
	}
}
