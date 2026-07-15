package adapters

import (
	"strings"
	"testing"
)

func TestActixWebAdapter_GenerateProject(t *testing.T) {
	a := ActixWebAdapter{}
	artifacts := a.GenerateProject("myproj")
	if len(artifacts) == 0 {
		t.Fatal("expected at least one artifact")
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	expected := []string{"README.md", "Cargo.toml", "src/main.rs", "src/lib.rs"}
	for _, p := range expected {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	// Check Cargo.toml contains actix-web
	for _, a := range artifacts {
		if a.Path == "Cargo.toml" {
			if !strings.Contains(string(a.Content), "actix-web") {
				t.Errorf("Cargo.toml should contain actix-web")
			}
			break
		}
	}
	// Check src/main.rs contains actix_web::main
	for _, a := range artifacts {
		if a.Path == "src/main.rs" {
			if !strings.Contains(string(a.Content), "actix_web::main") {
				t.Errorf("src/main.rs should contain actix_web::main")
			}
			break
		}
	}
}

func TestActixWebAdapter_GenerateModule(t *testing.T) {
	a := ActixWebAdapter{}
	artifacts := a.GenerateModule("users", "./internal/users", "myproj")
	if len(artifacts) == 0 {
		t.Fatal("expected at least one artifact")
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	// We generate specific files, check they exist
	for _, p := range []string{
		"src/users/mod.rs",
		"src/users/handler.rs",
		"src/users/service.rs",
		"src/users/models.rs",
		"tests/users_test.rs",
	} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	// handler.rs contains actix_web imports
	for _, a := range artifacts {
		if a.Path == "src/users/handler.rs" {
			content := string(a.Content)
			if !strings.Contains(content, "actix_web") {
				t.Errorf("handler.rs should import actix web")
			}
			break
		}
	}
	// service.rs defines trait
	for _, a := range artifacts {
		if a.Path == "src/users/service.rs" {
			content := string(a.Content)
			if !strings.Contains(content, "trait Service") {
				t.Errorf("service.rs should have trait")
			}
			break
		}
	}
	// models.rs uses serde
	for _, a := range artifacts {
		if a.Path == "src/users/models.rs" {
			content := string(a.Content)
			if !strings.Contains(content, "serde") {
				t.Errorf("models.rs should import serde")
			}
			if !strings.Contains(content, "Serialize") {
				t.Errorf("models.rs should have Serialize derive")
			}
			break
		}
	}
}

func TestActixWebAdapter_GenerateService(t *testing.T) {
	a := ActixWebAdapter{}
	artifacts := a.GenerateService("api-gateway", "http", 8080, "myproj")
	if len(artifacts) == 0 {
		t.Fatal("expected at least one artifact")
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	if !paths["src/services/api-gateway/server.rs"] {
		t.Errorf("missing server.rs")
	}
	for _, a := range artifacts {
		if a.Path == "src/services/api-gateway/server.rs" {
			content := string(a.Content)
			if !strings.Contains(content, "actix_web::main") {
				t.Errorf("server.rs should have actix_web::main")
			}
			if !strings.Contains(content, "8080") {
				t.Errorf("server.rs should contain port 8080")
			}
			break
		}
	}
}

func TestActixWebAdapter_GenerateDockerfile(t *testing.T) {
	a := ActixWebAdapter{}
	artifacts := a.GenerateDockerfile("myproj")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Path != "Dockerfile" {
		t.Errorf("expected path Dockerfile, got %s", artifacts[0].Path)
	}
	if !strings.Contains(string(artifacts[0].Content), "EXPOSE 8080") {
		t.Errorf("Dockerfile should expose port 8080")
	}
}

func TestActixWebAdapter_GenerateCI(t *testing.T) {
	a := ActixWebAdapter{}
	artifacts := a.GenerateCI("myproj")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Path != ".github/workflows/ci.yml" {
		t.Errorf("expected path .github/workflows/ci.yml, got %s", artifacts[0].Path)
	}
	if !strings.Contains(string(artifacts[0].Content), "cargo test") {
		t.Errorf("CI should run cargo test")
	}
}

func TestActixWebAdapter_GenerateDockerCompose(t *testing.T) {
	a := ActixWebAdapter{}
	artifacts := a.GenerateDockerCompose("myproj")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Path != "docker-compose.yml" {
		t.Errorf("expected path docker-compose.yml, got %s", artifacts[0].Path)
	}
	if !strings.Contains(string(artifacts[0].Content), "8080:8080") {
		t.Errorf("docker-compose should map port 8080")
	}
}

func TestActixWebAdapter_GenerateArchitectureDoc(t *testing.T) {
	a := ActixWebAdapter{}
	artifacts := a.GenerateArchitectureDoc("myproj", "event-driven")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Path != "docs/architecture.md" {
		t.Errorf("expected path docs/architecture.md, got %s", artifacts[0].Path)
	}
	if !strings.Contains(string(artifacts[0].Content), "event-driven") {
		t.Errorf("architecture doc should contain pattern")
	}
}
