package container

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/testutil"
)

func TestGenerateDockerfileGo(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
		},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{"go"},
		},
	}

	g := NewGenerator()
	artifacts := g.Generate(neir)

	if len(artifacts) < 1 {
		t.Fatalf("expected at least 1 artifact, got %d", len(artifacts))
	}

	found := false
	for _, a := range artifacts {
		if a.Path == "Dockerfile" {
			found = true
			content := string(a.Content)
			if !testutil.Contains(content, "golang") {
				t.Error("expected Go Dockerfile")
			}
		}
	}
	if !found {
		t.Error("expected Dockerfile artifact")
	}
}

func TestGenerateDockerCompose(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
			{Name: "worker", Port: 9090},
		},
	}

	g := NewGenerator()
	artifacts := g.Generate(neir)

	found := false
	for _, a := range artifacts {
		if a.Path == "docker-compose.yaml" {
			found = true
			content := string(a.Content)
			if !testutil.Contains(content, "api:") || !testutil.Contains(content, "worker:") {
				t.Error("expected both services in docker-compose")
			}
		}
	}
	if !found {
		t.Error("expected docker-compose.yaml artifact")
	}
}

func TestGenerateK8s(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
		},
	}

	g := NewGenerator()
	artifacts := g.Generate(neir)

	needed := map[string]bool{
		"k8s/namespace.yaml":      false,
		"k8s/deployment-api.yaml": false,
		"k8s/service-api.yaml":    false,
	}
	for _, a := range artifacts {
		if _, ok := needed[a.Path]; ok {
			needed[a.Path] = true
		}
	}
	for path, found := range needed {
		if !found {
			t.Errorf("expected artifact %s", path)
		}
	}
}

func TestGeneratePython(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8000},
		},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{"python"},
		},
	}

	g := NewGenerator()
	artifacts := g.Generate(neir)

	for _, a := range artifacts {
		if a.Path == "Dockerfile" {
			content := string(a.Content)
			if !testutil.Contains(content, "python") {
				t.Error("expected Python Dockerfile")
			}
		}
	}
}

func TestGenerateNode(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "web", Port: 3000},
		},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{"typescript"},
		},
	}

	g := NewGenerator()
	artifacts := g.Generate(neir)

	for _, a := range artifacts {
		if a.Path == "Dockerfile" {
			content := string(a.Content)
			if !testutil.Contains(content, "node") {
				t.Error("expected Node Dockerfile")
			}
		}
	}
}


