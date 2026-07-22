package builder

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
)

func TestBuilderCreatesNEIRFromResolvedSpec(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project":  "acme-api",
		"modules":  []map[string]any{{"name": "auth", "path": "./internal/auth"}},
		"services": []map[string]any{{"name": "gateway", "kind": "http", "port": 8080}},
	}}

	builder := NewBuilder()
	neir, err := builder.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Project == nil || neir.Project.Name != "acme-api" {
		t.Fatalf("expected project acme-api, got %v", neir.Project)
	}
	if len(neir.Modules) != 1 {
		t.Fatalf("expected one module, got %d", len(neir.Modules))
	}
	if len(neir.Services) != 1 {
		t.Fatalf("expected one service, got %d", len(neir.Services))
	}
	if neir.Services[0].Name != "gateway" {
		t.Fatalf("expected service gateway, got %s", neir.Services[0].Name)
	}
	if neir.Services[0].Port != 8080 {
		t.Fatalf("expected service port 8080, got %d", neir.Services[0].Port)
	}
	if neir.Architecture != nil {
		t.Fatalf("expected nil architecture, got %v", neir.Architecture)
	}
}

func TestBuilderExtractsArchitecture(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project":      "acme-api",
		"modules":      []map[string]any{{"name": "core", "path": "./internal/core"}},
		"architecture": map[string]any{"pattern": "clean", "description": "Clean architecture"},
	}}

	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Architecture == nil {
		t.Fatal("expected architecture to be set")
	}
	if neir.Architecture.Pattern != "clean" {
		t.Fatalf("expected pattern clean, got %s", neir.Architecture.Pattern)
	}
	if neir.Architecture.Description != "Clean architecture" {
		t.Fatalf("expected description, got %s", neir.Architecture.Description)
	}
}

func TestBuilderWithNilInput(t *testing.T) {
	b := NewBuilder()
	_, err := b.Build(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestBuilderExtractsDeployment(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "acme-api",
		"deployment": map[string]any{
			"strategy": "canary",
			"environments": []any{
				map[string]any{"name": "staging"},
				"production",
			},
		},
	}}

	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Deployment == nil {
		t.Fatal("expected deployment to be set")
	}
	if neir.Deployment.Strategy != "canary" {
		t.Fatalf("expected strategy canary, got %s", neir.Deployment.Strategy)
	}
	if len(neir.Deployment.Environments) != 2 {
		t.Fatalf("expected 2 environments, got %d", len(neir.Deployment.Environments))
	}
	if neir.Deployment.Environments[0].Name != "staging" {
		t.Fatalf("expected first env staging, got %s", neir.Deployment.Environments[0].Name)
	}
	if neir.Deployment.Environments[1].Name != "production" {
		t.Fatalf("expected second env production, got %s", neir.Deployment.Environments[1].Name)
	}
}

func TestBuilderExtractsTesting(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "acme-api",
		"testing": map[string]any{
			"strategy": "unit",
			"coverage": "high",
		},
	}}

	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Testing == nil {
		t.Fatal("expected testing to be set")
	}
	if neir.Testing.Strategy != "unit" {
		t.Fatalf("expected strategy unit, got %s", neir.Testing.Strategy)
	}
	if neir.Testing.Coverage == nil {
		t.Fatal("expected coverage to be set")
	}
	if neir.Testing.Coverage.MinPercent != 80.0 {
		t.Fatalf("expected coverage 80.0, got %f", neir.Testing.Coverage.MinPercent)
	}
}

func TestBuilderExtractsTestingMediumCoverage(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"testing": map[string]any{"coverage": "medium"},
	}}

	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Testing.Coverage.MinPercent != 60.0 {
		t.Fatalf("expected coverage 60.0, got %f", neir.Testing.Coverage.MinPercent)
	}
}

func TestBuilderExtractsTestingLowCoverage(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"testing": map[string]any{"coverage": "low"},
	}}

	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Testing.Coverage.MinPercent != 40.0 {
		t.Fatalf("expected coverage 40.0, got %f", neir.Testing.Coverage.MinPercent)
	}
}

func TestBuilderRejectsWrongType(t *testing.T) {
	b := NewBuilder()
	_, err := b.Build("not a resolved spec")
	if err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestBuilderWithEmptyContext(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Project != nil {
		t.Fatal("expected nil project for empty context")
	}
	if len(neir.Modules) != 0 {
		t.Fatal("expected zero modules for empty context")
	}
}

func TestBuilderExtractsModuleDependencies(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "myapp",
		"modules": []map[string]any{
			{"name": "api", "path": "./api", "dependencies": []any{"auth", "db"}},
		},
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(neir.Modules[0].Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(neir.Modules[0].Dependencies))
	}
	if neir.Modules[0].Dependencies[0] != "auth" {
		t.Fatalf("expected dependency auth, got %s", neir.Modules[0].Dependencies[0])
	}
}

func TestBuilderExtractsServiceEndpoints(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "myapp",
		"modules": []map[string]any{{"name": "core", "path": "./core"}},
		"services": []map[string]any{
			{
				"name": "api", "kind": "http", "port": 8080,
				"endpoints": []any{
					map[string]any{"method": "GET", "path": "/users"},
					map[string]any{"method": "POST", "path": "/users", "action": "create"},
				},
			},
		},
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(neir.Services[0].Endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(neir.Services[0].Endpoints))
	}
	if neir.Services[0].Endpoints[0].Method != "GET" {
		t.Fatalf("expected GET, got %s", neir.Services[0].Endpoints[0].Method)
	}
}

func TestBuilderExtractsGeneration(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "myapp",
		"modules": []map[string]any{{"name": "core", "path": "./core"}},
		"generation": map[string]any{
			"languages":  []any{"go", "typescript"},
			"output_dir": "./dist",
			"module_dir": "./modules",
		},
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Generation == nil {
		t.Fatal("expected generation to be set")
	}
	if len(neir.Generation.Languages) != 2 {
		t.Fatalf("expected 2 languages, got %d", len(neir.Generation.Languages))
	}
	if neir.Generation.OutputDir != "./dist" {
		t.Fatalf("expected output_dir ./dist, got %s", neir.Generation.OutputDir)
	}
	if neir.Generation.ModuleDir != "./modules" {
		t.Fatalf("expected module_dir ./modules, got %s", neir.Generation.ModuleDir)
	}
}

func TestBuilderExtractsGenerationStringLanguages(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "myapp",
		"modules": []map[string]any{{"name": "core", "path": "./core"}},
		"generation": map[string]any{
			"languages": []string{"go", "python"},
		},
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(neir.Generation.Languages) != 2 {
		t.Fatalf("expected 2 languages, got %d", len(neir.Generation.Languages))
	}
}

func TestBuilderExtractsCloudInfrastructure(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "myapp",
		"modules": []map[string]any{{"name": "core", "path": "./core"}},
		"cloud": map[string]any{
			"provider":    "aws",
			"region":      "us-east-1",
			"project":     "myapp-prod",
			"environment": "production",
			"resources": []any{
				map[string]any{
					"name": "db",
					"kind": "rds",
					"type": "postgres",
					"spec": map[string]any{"size": "large", "storage": "100GB"},
				},
			},
			"attributes": map[string]any{"env": "prod", "team": "platform"},
		},
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Infrastructure == nil {
		t.Fatal("expected infrastructure to be set")
	}
	if neir.Infrastructure.Provider != "aws" {
		t.Fatalf("expected provider aws, got %s", neir.Infrastructure.Provider)
	}
	if neir.Infrastructure.Region != "us-east-1" {
		t.Fatalf("expected region us-east-1, got %s", neir.Infrastructure.Region)
	}
	if neir.Infrastructure.Project != "myapp-prod" {
		t.Fatalf("expected project myapp-prod, got %s", neir.Infrastructure.Project)
	}
	if len(neir.Infrastructure.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(neir.Infrastructure.Resources))
	}
	if neir.Infrastructure.Resources[0].Name != "db" {
		t.Fatalf("expected resource name db, got %s", neir.Infrastructure.Resources[0].Name)
	}
	if neir.Infrastructure.Attributes["env"] != "prod" {
		t.Fatalf("expected attribute env=prod, got %s", neir.Infrastructure.Attributes["env"])
	}
}

func TestBuilderExtractsModulesAsAnySlice(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "myapp",
		"modules": []any{
			map[string]any{"name": "auth", "path": "./auth"},
		},
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(neir.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(neir.Modules))
	}
	if neir.Modules[0].Name != "auth" {
		t.Fatalf("expected module auth, got %s", neir.Modules[0].Name)
	}
}

func TestBuilderExtractsServicesAsAnySlice(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "myapp",
		"modules": []map[string]any{{"name": "core", "path": "./core"}},
		"services": []any{
			map[string]any{"name": "api", "kind": "http", "port": 8080},
		},
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(neir.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(neir.Services))
	}
	if neir.Services[0].Name != "api" {
		t.Fatalf("expected service api, got %s", neir.Services[0].Name)
	}
}

func TestBuilderExtractsMaximalConfig(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project":      "maximal",
		"modules":      []map[string]any{{"name": "core", "path": "./core"}},
		"services":     []map[string]any{{"name": "api", "kind": "http", "port": 8080}},
		"architecture": map[string]any{"pattern": "hexagonal"},
		"generation":   map[string]any{"languages": []any{"go"}, "output_dir": "./out"},
		"deployment": map[string]any{
			"strategy":     "rolling",
			"environments": []any{map[string]any{"name": "prod"}},
		},
		"testing": map[string]any{"strategy": "unit", "coverage": "high"},
		"cloud":   map[string]any{"provider": "gcp", "region": "us-central1"},
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Project.Name != "maximal" {
		t.Fatalf("expected project maximal, got %s", neir.Project.Name)
	}
	if neir.Architecture == nil || neir.Architecture.Pattern != "hexagonal" {
		t.Fatal("expected hexagonal architecture")
	}
	if neir.Generation == nil || len(neir.Generation.Languages) != 1 {
		t.Fatal("expected generation with 1 language")
	}
	if neir.Deployment == nil || neir.Deployment.Strategy != "rolling" {
		t.Fatal("expected rolling deployment")
	}
	if neir.Testing == nil || neir.Testing.Coverage.MinPercent != 80.0 {
		t.Fatal("expected testing with 80% coverage")
	}
	if neir.Infrastructure == nil || neir.Infrastructure.Provider != "gcp" {
		t.Fatal("expected gcp infrastructure")
	}
}
