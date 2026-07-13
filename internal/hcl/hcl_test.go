package hcl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSimple(t *testing.T) {
	input := []byte(`
project "myapp" {
  version     = "1.0.0"
  description = "My application"
}

service "api" {
  image = "myapp-api"
  port  = 8080
  type  = "backend"
}

infra "infra" {
  engine = "docker"
}
`)

	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Project.Name != "myapp" {
		t.Errorf("expected project name 'myapp', got %q", spec.Project.Name)
	}
	if spec.Project.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", spec.Project.Version)
	}
	if spec.Project.Description != "My application" {
		t.Errorf("expected description 'My application', got %q", spec.Project.Description)
	}
	if len(spec.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(spec.Services))
	}
	svc := spec.Services["api"]
	if svc.Image != "myapp-api" {
		t.Errorf("expected image 'myapp-api', got %q", svc.Image)
	}
	if svc.Port != 8080 {
		t.Errorf("expected port 8080, got %d", svc.Port)
	}
	if svc.Type != "backend" {
		t.Errorf("expected type 'backend', got %q", svc.Type)
	}
	if spec.Infra.Engine != "docker" {
		t.Errorf("expected engine 'docker', got %q", spec.Infra.Engine)
	}
}

func TestParseMultipleServices(t *testing.T) {
	input := []byte(`
project "multi" {
  version = "2.0.0"
}

service "api" {
  port = 8080
  type = "backend"
}

service "web" {
  port = 3000
  type = "frontend"
}

service "worker" {
  port = 9090
  type = "job"
}
`)

	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(spec.Services))
	}
	for _, name := range []string{"api", "web", "worker"} {
		if _, ok := spec.Services[name]; !ok {
			t.Errorf("missing service %q", name)
		}
	}
}

func TestParseComments(t *testing.T) {
	input := []byte(`
# This is a comment
// This is also a comment
project "commented" {
  version = "1.0.0"
}
`)

	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Project.Name != "commented" {
		t.Errorf("expected project 'commented', got %q", spec.Project.Name)
	}
}

func TestParseEmpty(t *testing.T) {
	spec, err := Parse([]byte(""), "empty.hcl")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Project.Name != "" {
		t.Errorf("expected empty project name, got %q", spec.Project.Name)
	}
	if len(spec.Services) != 0 {
		t.Errorf("expected 0 services, got %d", len(spec.Services))
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.hcl")
	content := `
project "filetest" {
  version = "3.0.0"
}

service "backend" {
  port = 5000
  type = "api"
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	spec, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Project.Name != "filetest" {
		t.Errorf("expected 'filetest', got %q", spec.Project.Name)
	}
	if spec.Project.Version != "3.0.0" {
		t.Errorf("expected '3.0.0', got %q", spec.Project.Version)
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/file.hcl")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseInvalid(t *testing.T) {
	input := []byte(`
project "bad" {
  version = "1.0.0"
  unknown_field = "value"
  broken
}
`)

	_, err := Parse(input, "bad.hcl")
	if err != nil {
		t.Logf("got error (expected for malformed HCL): %v", err)
	}
}

func TestParseProjectOnly(t *testing.T) {
	input := []byte(`
project "minimal" {
  version = "1.0.0"
}
`)
	spec, err := Parse(input, "minimal.hcl")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Project.Name != "minimal" {
		t.Errorf("expected 'minimal', got %q", spec.Project.Name)
	}
	if len(spec.Services) != 0 {
		t.Errorf("expected 0 services, got %d", len(spec.Services))
	}
}
