package adapters

import (
	"strings"
	"testing"
)

func TestRustAdapter_GenerateProject(t *testing.T) {
	a := RustAdapter{}
	artifacts := a.GenerateProject("MyProject")
	if len(artifacts) < 4 {
		t.Fatalf("expected at least 4 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{"README.md", "Cargo.toml", "src/main.rs", "src/lib.rs"} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	for _, a := range artifacts {
		if a.Path == "Cargo.toml" {
			content := string(a.Content)
			if !strings.Contains(content, "edition = \"2021\"") {
				t.Errorf("Cargo.toml should contain edition 2021")
			}
			if !strings.Contains(content, "tokio") {
				t.Errorf("Cargo.toml should contain tokio")
			}
		}
	}
}

func TestRustAdapter_GenerateModule(t *testing.T) {
	a := RustAdapter{}
	artifacts := a.GenerateModule("users", "./internal/users", "MyProject")
	if len(artifacts) < 6 {
		t.Fatalf("expected at least 6 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{
		"src/users/mod.rs",
		"src/users/handler.rs",
		"src/users/service.rs",
		"src/users/repository.rs",
		"src/users/models.rs",
		"tests/users_test.rs",
	} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
}

func TestRustAdapter_GenerateService(t *testing.T) {
	a := RustAdapter{}
	artifacts := a.GenerateService("api-gateway", "http", 8080, "MyProject")
	if len(artifacts) < 1 {
		t.Fatalf("expected at least 1 artifact, got %d", len(artifacts))
	}
	found := false
	for _, art := range artifacts {
		if strings.Contains(art.Path, "server.rs") {
			found = true
			content := string(art.Content)
			if !strings.Contains(content, "axum") {
				t.Errorf("server.rs should use axum")
			}
		}
	}
	if !found {
		t.Error("expected server.rs for http service")
	}
}

func TestRustAdapter_GenerateServiceNonHTTP(t *testing.T) {
	a := RustAdapter{}
	artifacts := a.GenerateService("worker", "grpc", 9090, "MyProject")
	if len(artifacts) != 0 {
		t.Errorf("expected 0 artifacts for non-http service, got %d", len(artifacts))
	}
}

func TestRustAdapter_GenerateDockerfile(t *testing.T) {
	a := RustAdapter{}
	artifacts := a.GenerateDockerfile("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "rust:1.78-alpine") {
		t.Errorf("Dockerfile should use rust:1.78-alpine")
	}
	if !strings.Contains(content, "cargo build --release") {
		t.Errorf("Dockerfile should use cargo build --release")
	}
}

func TestRustAdapter_GenerateCI(t *testing.T) {
	a := RustAdapter{}
	artifacts := a.GenerateCI("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "dtolnay/rust-toolchain@stable") {
		t.Errorf("CI should use dtolnay/rust-toolchain@stable")
	}
	if !strings.Contains(content, "cargo test") {
		t.Errorf("CI should run cargo test")
	}
}

func TestRustAdapter_GenerateDockerCompose(t *testing.T) {
	a := RustAdapter{}
	artifacts := a.GenerateDockerCompose("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "8080:8080") {
		t.Errorf("docker-compose should map port 8080")
	}
}

func TestRustAdapter_GenerateArchitectureDoc(t *testing.T) {
	a := RustAdapter{}
	artifacts := a.GenerateArchitectureDoc("MyProject", "hexagonal")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "hexagonal") {
		t.Errorf("architecture doc should contain pattern")
	}
	if !strings.Contains(content, "Rust") {
		t.Errorf("architecture doc should contain Rust")
	}
}
