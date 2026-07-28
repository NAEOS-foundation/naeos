package builder

import (
	"sync"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/infrastructure"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
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

func TestExtractModule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input map[string]any
		want module.Module
	}{
		{
			name: "all fields",
			input: map[string]any{
				"name": "auth", "path": "./auth", "description": "Auth module",
				"condition": "prod", "dependencies": []any{"db", "cache"},
			},
			want: module.Module{
				Name: "auth", Path: "./auth", Description: "Auth module",
				Condition: "prod", Dependencies: []string{"db", "cache"},
			},
		},
		{
			name:  "empty map",
			input: map[string]any{},
			want:  module.Module{},
		},
		{
			name: "non-string dependency skipped",
			input: map[string]any{
				"name": "x", "dependencies": []any{"good", 42},
			},
			want: module.Module{Name: "x", Dependencies: []string{"good"}},
		},
		{
			name: "wrong types ignored",
			input: map[string]any{
				"name": "m", "path": 42, "description": true, "condition": nil,
			},
			want: module.Module{Name: "m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModule(tt.input)
			if got.Name != tt.want.Name || got.Path != tt.want.Path || got.Description != tt.want.Description || got.Condition != tt.want.Condition {
				t.Fatalf("extractModule() = %+v, want %+v", got, tt.want)
			}
			if len(got.Dependencies) != len(tt.want.Dependencies) {
				t.Fatalf("got %d dependencies, want %d", len(got.Dependencies), len(tt.want.Dependencies))
			}
			for i := range got.Dependencies {
				if got.Dependencies[i] != tt.want.Dependencies[i] {
					t.Fatalf("dependency[%d] = %s, want %s", i, got.Dependencies[i], tt.want.Dependencies[i])
				}
			}
		})
	}
}

func TestExtractService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input map[string]any
		want service.Service
	}{
		{
			name: "all fields with endpoints",
			input: map[string]any{
				"name": "api", "kind": "http", "port": 8080,
				"description": "API service", "condition": "!dev",
				"endpoints": []any{
					map[string]any{"method": "GET", "path": "/users", "action": "list"},
				},
			},
			want: service.Service{
				Name: "api", Kind: "http", Port: 8080,
				Description: "API service", Condition: "!dev",
				Endpoints: []service.Endpoint{{Method: "GET", Path: "/users", Action: "list"}},
			},
		},
		{
			name:  "empty map",
			input: map[string]any{},
			want:  service.Service{},
		},
		{
			name: "wrong types ignored",
			input: map[string]any{
				"name": "s", "kind": 42, "port": "bad", "description": true,
			},
			want: service.Service{Name: "s"},
		},
		{
			name: "endpoints with non-map items skipped",
			input: map[string]any{
				"name": "s",
				"endpoints": []any{
					map[string]any{"method": "GET"},
					"not a map",
				},
			},
			want: service.Service{Name: "s", Endpoints: []service.Endpoint{{Method: "GET"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractService(tt.input)
			if got.Name != tt.want.Name || got.Kind != tt.want.Kind || got.Port != tt.want.Port {
				t.Fatalf("extractService() = %+v, want %+v", got, tt.want)
			}
			if got.Description != tt.want.Description || got.Condition != tt.want.Condition {
				t.Fatalf("extractService() = %+v, want %+v", got, tt.want)
			}
			if len(got.Endpoints) != len(tt.want.Endpoints) {
				t.Fatalf("got %d endpoints, want %d", len(got.Endpoints), len(tt.want.Endpoints))
			}
			for i := range got.Endpoints {
				if got.Endpoints[i] != tt.want.Endpoints[i] {
					t.Fatalf("endpoint[%d] = %+v, want %+v", i, got.Endpoints[i], tt.want.Endpoints[i])
				}
			}
		})
	}
}

func TestExtractArchitecture(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input map[string]any
		want *architecture.Architecture
	}{
		{
			name:  "full architecture",
			input: map[string]any{"pattern": "clean", "description": "Clean arch"},
			want:  &architecture.Architecture{Pattern: "clean", Description: "Clean arch"},
		},
		{
			name:  "empty map",
			input: map[string]any{},
			want:  &architecture.Architecture{},
		},
		{
			name:  "wrong types ignored",
			input: map[string]any{"pattern": 42, "description": true},
			want:  &architecture.Architecture{},
		},
		{
			name:  "description only",
			input: map[string]any{"description": "desc-only"},
			want:  &architecture.Architecture{Description: "desc-only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractArchitecture(tt.input)
			if got.Pattern != tt.want.Pattern || got.Description != tt.want.Description {
				t.Fatalf("extractArchitecture() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestExtractGeneration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input map[string]any
		want *generation.GenerationConfig
	}{
		{
			name: "full with []any languages",
			input: map[string]any{
				"languages": []any{"go", "python"}, "output_dir": "./out", "module_dir": "./mod",
			},
			want: &generation.GenerationConfig{
				Languages: []language.Language{"go", "python"}, OutputDir: "./out", ModuleDir: "./mod",
			},
		},
		{
			name: "empty map",
			input: map[string]any{},
			want: &generation.GenerationConfig{},
		},
		{
			name: "wrong types ignored",
			input: map[string]any{
				"languages": "not a slice", "output_dir": 42, "module_dir": true,
			},
			want: &generation.GenerationConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractGeneration(tt.input)
			if got.OutputDir != tt.want.OutputDir || got.ModuleDir != tt.want.ModuleDir {
				t.Fatalf("extractGeneration() = %+v, want %+v", got, tt.want)
			}
			if len(got.Languages) != len(tt.want.Languages) {
				t.Fatalf("got %d languages, want %d", len(got.Languages), len(tt.want.Languages))
			}
			for i := range got.Languages {
				if got.Languages[i] != tt.want.Languages[i] {
					t.Fatalf("language[%d] = %s, want %s", i, got.Languages[i], tt.want.Languages[i])
				}
			}
		})
	}
}

func TestExtractGenerationStringLanguages(t *testing.T) {
	got := extractGeneration(map[string]any{
		"languages": []string{"rust", "zig"},
	})
	if len(got.Languages) != 2 {
		t.Fatalf("expected 2 languages, got %d", len(got.Languages))
	}
	if got.Languages[0] != language.Language("rust") {
		t.Fatalf("expected rust, got %s", got.Languages[0])
	}
	if got.Languages[1] != language.Language("zig") {
		t.Fatalf("expected zig, got %s", got.Languages[1])
	}
}

func TestExtractDeployment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input map[string]any
		want *deployment.Deployment
	}{
		{
			name: "full with environments and variables",
			input: map[string]any{
				"strategy": "canary",
				"environments": []any{
					map[string]any{
						"name": "staging", "kind": "k8s",
						"variables": map[string]any{"ENV": "staging", "DEBUG": "1"},
					},
					"production",
				},
			},
			want: &deployment.Deployment{
				Strategy: "canary",
				Environments: []deployment.Environment{
					{Name: "staging", Kind: "k8s", Variables: map[string]string{"ENV": "staging", "DEBUG": "1"}},
					{Name: "production"},
				},
			},
		},
		{
			name:  "empty map",
			input: map[string]any{},
			want:  &deployment.Deployment{},
		},
		{
			name: "wrong types ignored",
			input: map[string]any{
				"strategy": 42, "environments": "not a slice",
			},
			want: &deployment.Deployment{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDeployment(tt.input)
			if got.Strategy != tt.want.Strategy {
				t.Fatalf("got strategy %s, want %s", got.Strategy, tt.want.Strategy)
			}
			if len(got.Environments) != len(tt.want.Environments) {
				t.Fatalf("got %d environments, want %d", len(got.Environments), len(tt.want.Environments))
			}
			for i := range got.Environments {
				if got.Environments[i].Name != tt.want.Environments[i].Name {
					t.Fatalf("env[%d].name = %s, want %s", i, got.Environments[i].Name, tt.want.Environments[i].Name)
				}
				if got.Environments[i].Kind != tt.want.Environments[i].Kind {
					t.Fatalf("env[%d].kind = %s, want %s", i, got.Environments[i].Kind, tt.want.Environments[i].Kind)
				}
				for k, v := range tt.want.Environments[i].Variables {
					if got.Environments[i].Variables[k] != v {
						t.Fatalf("env[%d].variables[%s] = %s, want %s", i, k, got.Environments[i].Variables[k], v)
					}
				}
			}
		})
	}
}

func TestExtractTesting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]any
		wantStrat testingmodel.TestingStrategy
		wantCov  *testingmodel.Coverage
	}{
		{
			name:     "full",
			input:    map[string]any{"strategy": "unit", "coverage": "high"},
			wantStrat: "unit",
			wantCov:  &testingmodel.Coverage{MinPercent: 80.0},
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			wantStrat: "",
			wantCov:  nil,
		},
		{
			name:     "strategy only",
			input:    map[string]any{"strategy": "integration"},
			wantStrat: "integration",
			wantCov:  nil,
		},
		{
			name:     "unknown coverage level",
			input:    map[string]any{"coverage": "critical"},
			wantStrat: "",
			wantCov:  &testingmodel.Coverage{MinPercent: 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTesting(tt.input)
			if got.Strategy != tt.wantStrat {
				t.Fatalf("got strategy %s, want %s", got.Strategy, tt.wantStrat)
			}
			if tt.wantCov == nil {
				if got.Coverage != nil {
					t.Fatal("expected nil coverage")
				}
				return
			}
			if got.Coverage == nil {
				t.Fatal("expected coverage to be set")
			}
			if got.Coverage.MinPercent != tt.wantCov.MinPercent {
				t.Fatalf("got coverage %f, want %f", got.Coverage.MinPercent, tt.wantCov.MinPercent)
			}
		})
	}
}

func TestExtractCloud(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input map[string]any
		want *infrastructure.Infrastructure
	}{
		{
			name: "full cloud",
			input: map[string]any{
				"provider": "aws", "region": "us-east-1", "project": "myapp", "environment": "prod",
				"resources": []any{
					map[string]any{
						"name": "db", "kind": "rds", "type": "postgres",
						"spec": map[string]any{"size": "large"},
					},
				},
				"attributes": map[string]any{"team": "platform"},
			},
			want: &infrastructure.Infrastructure{
				Provider: "aws", Region: "us-east-1", Project: "myapp", Environment: "prod",
				Resources: []infrastructure.Resource{
					{Name: "db", Kind: "rds", Type: "postgres", Spec: map[string]string{"size": "large"}},
				},
				Attributes: map[string]string{"team": "platform"},
			},
		},
		{
			name:  "empty map",
			input: map[string]any{},
			want:  &infrastructure.Infrastructure{},
		},
		{
			name: "wrong types ignored",
			input: map[string]any{
				"provider": 42, "region": true, "resources": "not a slice",
				"attributes": "not a map",
			},
			want: &infrastructure.Infrastructure{},
		},
		{
			name: "resource without spec",
			input: map[string]any{
				"resources": []any{
					map[string]any{"name": "queue", "kind": "sqs"},
				},
			},
			want: &infrastructure.Infrastructure{
				Resources: []infrastructure.Resource{
					{Name: "queue", Kind: "sqs"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCloud(tt.input)
			if got.Provider != tt.want.Provider || got.Region != tt.want.Region || got.Project != tt.want.Project || got.Environment != tt.want.Environment {
				t.Fatalf("extractCloud() = %+v, want %+v", got, tt.want)
			}
			if len(got.Resources) != len(tt.want.Resources) {
				t.Fatalf("got %d resources, want %d", len(got.Resources), len(tt.want.Resources))
			}
			for i := range got.Resources {
				r := got.Resources[i]
				wr := tt.want.Resources[i]
				if r.Name != wr.Name || r.Kind != wr.Kind || r.Type != wr.Type {
					t.Fatalf("resource[%d] = %+v, want %+v", i, r, wr)
				}
				for k, v := range wr.Spec {
					if r.Spec[k] != v {
						t.Fatalf("resource[%d].spec[%s] = %s, want %s", i, k, r.Spec[k], v)
					}
				}
			}
			if len(got.Attributes) != len(tt.want.Attributes) {
				t.Fatalf("got %d attributes, want %d", len(got.Attributes), len(tt.want.Attributes))
			}
			for k, v := range tt.want.Attributes {
				if got.Attributes[k] != v {
					t.Fatalf("attribute[%s] = %s, want %s", k, got.Attributes[k], v)
				}
			}
		})
	}
}

func TestBuilderExtractsActiveProfileAndInherits(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project":        "myapp",
		"active_profile": "staging",
		"inherits":       "base",
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.ActiveProfile != "staging" {
		t.Fatalf("expected active_profile staging, got %s", neir.ActiveProfile)
	}
	if neir.Inherits != "base" {
		t.Fatalf("expected inherits base, got %s", neir.Inherits)
	}
}

func TestBuilderActiveProfileWrongType(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"active_profile": 42,
	}}
	b := NewBuilder()
	neir, err := b.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.ActiveProfile != "" {
		t.Fatalf("expected empty active_profile, got %s", neir.ActiveProfile)
	}
}

func TestModuleLoader(t *testing.T) {
	t.Parallel()

	t.Run("LoadModules", func(t *testing.T) {
		loader := NewModuleLoader()

		t.Run("nil input", func(t *testing.T) {
			mods, err := loader.LoadModules(nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mods != nil {
				t.Fatal("expected nil modules")
			}
		})

		t.Run("wrong type", func(t *testing.T) {
			mods, err := loader.LoadModules("not a spec")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mods != nil {
				t.Fatal("expected nil modules")
			}
		})

		t.Run("no modules key", func(t *testing.T) {
			mods, err := loader.LoadModules(&resolver.ResolvedSpec{Context: map[string]any{}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mods != nil {
				t.Fatal("expected nil modules")
			}
		})

		t.Run("modules as []map[string]any", func(t *testing.T) {
			mods, err := loader.LoadModules(&resolver.ResolvedSpec{Context: map[string]any{
				"modules": []map[string]any{{"name": "auth", "path": "./auth"}},
			}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(mods) != 1 || mods[0].Name != "auth" {
				t.Fatalf("expected 1 module auth, got %+v", mods)
			}
		})

		t.Run("modules as []any", func(t *testing.T) {
			mods, err := loader.LoadModules(&resolver.ResolvedSpec{Context: map[string]any{
				"modules": []any{
					map[string]any{"name": "core", "path": "./core"},
					"not a map",
				},
			}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(mods) != 1 || mods[0].Name != "core" {
				t.Fatalf("expected 1 module core, got %+v", mods)
			}
		})
	})

	t.Run("LoadServices", func(t *testing.T) {
		loader := NewModuleLoader()

		t.Run("nil input", func(t *testing.T) {
			svcs, err := loader.LoadServices(nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if svcs != nil {
				t.Fatal("expected nil services")
			}
		})

		t.Run("wrong type", func(t *testing.T) {
			svcs, err := loader.LoadServices("not a spec")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if svcs != nil {
				t.Fatal("expected nil services")
			}
		})

		t.Run("no services key", func(t *testing.T) {
			svcs, err := loader.LoadServices(&resolver.ResolvedSpec{Context: map[string]any{}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if svcs != nil {
				t.Fatal("expected nil services")
			}
		})

		t.Run("services as []map[string]any", func(t *testing.T) {
			svcs, err := loader.LoadServices(&resolver.ResolvedSpec{Context: map[string]any{
				"services": []map[string]any{{"name": "api", "kind": "http", "port": 8080}},
			}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(svcs) != 1 || svcs[0].Name != "api" || svcs[0].Port != 8080 {
				t.Fatalf("expected 1 service api:8080, got %+v", svcs)
			}
		})

		t.Run("services as []any", func(t *testing.T) {
			svcs, err := loader.LoadServices(&resolver.ResolvedSpec{Context: map[string]any{
				"services": []any{
					map[string]any{"name": "worker", "kind": "worker"},
					"not a map",
				},
			}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(svcs) != 1 || svcs[0].Name != "worker" {
				t.Fatalf("expected 1 service worker, got %+v", svcs)
			}
		})
	})
}

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
	_, ok := b.(DefaultBuilder)
	if !ok {
		t.Fatal("expected DefaultBuilder type")
	}
}

func TestNewModuleLoader(t *testing.T) {
	loader := NewModuleLoader()
	if loader == nil {
		t.Fatal("expected non-nil loader")
	}
	_, ok := loader.(DefaultModuleLoader)
	if !ok {
		t.Fatal("expected DefaultModuleLoader type")
	}
}

func TestNewLazyBuilder(t *testing.T) {
	t.Parallel()

	t.Run("with nil inner and nil loader", func(t *testing.T) {
		lb := NewLazyBuilder(nil, nil)
		if lb == nil {
			t.Fatal("expected non-nil LazyBuilder")
		}
		if lb.built {
			t.Fatal("expected built=false initially")
		}
	})

	t.Run("with provided inner and loader", func(t *testing.T) {
		lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
		if lb == nil {
			t.Fatal("expected non-nil LazyBuilder")
		}
	})
}

func TestLazyBuilderBuild(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "lazy-test",
	}}

	t.Run("first build succeeds", func(t *testing.T) {
		lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
		neir, err := lb.Build(resolved)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if neir.Project == nil || neir.Project.Name != "lazy-test" {
			t.Fatalf("expected project lazy-test, got %+v", neir.Project)
		}
		if !lb.built {
			t.Fatal("expected built=true after first build")
		}
	})

	t.Run("second build returns cached result", func(t *testing.T) {
		lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
		first, err := lb.Build(resolved)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		second, err := lb.Build(resolved)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if first != second {
			t.Fatal("expected same pointer for cached result")
		}
	})

	t.Run("build error propagated", func(t *testing.T) {
		lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
		_, err := lb.Build(nil)
		if err == nil {
			t.Fatal("expected error for nil input")
		}
	})

	t.Run("LazyBuilder uses loader modules", func(t *testing.T) {
		lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
		resolved := &resolver.ResolvedSpec{Context: map[string]any{
			"project": "lazy-loader",
			"modules": []map[string]any{{"name": "from-loader", "path": "./loader"}},
		}}
		neir, err := lb.Build(resolved)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(neir.Modules) != 1 || neir.Modules[0].Name != "from-loader" {
			t.Fatalf("expected module from-loader, got %+v", neir.Modules)
		}
	})

	t.Run("LazyBuilder uses loader services", func(t *testing.T) {
		lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
		resolved := &resolver.ResolvedSpec{Context: map[string]any{
			"project":  "lazy-svc",
			"services": []map[string]any{{"name": "svc-loader", "kind": "http"}},
		}}
		neir, err := lb.Build(resolved)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(neir.Services) != 1 || neir.Services[0].Name != "svc-loader" {
			t.Fatalf("expected service svc-loader, got %+v", neir.Services)
		}
	})
}

func TestLazyBuilderReset(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "reset-test",
	}}

	lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
	_, err := lb.Build(resolved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !lb.IsBuilt() {
		t.Fatal("expected IsBuilt()=true before reset")
	}

	lb.Reset()
	if lb.IsBuilt() {
		t.Fatal("expected IsBuilt()=false after reset")
	}
	if lb.neir != nil {
		t.Fatal("expected nil neir after reset")
	}

	neir, err := lb.Build(resolved)
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if neir.Project == nil || neir.Project.Name != "reset-test" {
		t.Fatalf("expected project reset-test after rebuild, got %+v", neir.Project)
	}
}

func TestLazyBuilderIsBuilt(t *testing.T) {
	t.Parallel()

	t.Run("initially false", func(t *testing.T) {
		lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
		if lb.IsBuilt() {
			t.Fatal("expected false initially")
		}
	})

	t.Run("true after build", func(t *testing.T) {
		lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
		lb.Build(&resolver.ResolvedSpec{Context: map[string]any{}})
		if !lb.IsBuilt() {
			t.Fatal("expected true after build")
		}
	})
}

func TestLazyBuilderConcurrency(t *testing.T) {
	lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "concurrent",
	}}

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			neir, err := lb.Build(resolved)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if neir.Project == nil || neir.Project.Name != "concurrent" {
				t.Errorf("unexpected project: %+v", neir.Project)
			}
		}()
	}
	wg.Wait()

	if !lb.IsBuilt() {
		t.Fatal("expected built after concurrent access")
	}
}

func TestLazyBuilderResetConcurrency(t *testing.T) {
	lb := NewLazyBuilder(DefaultBuilder{}, DefaultModuleLoader{})
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "reset-concurrent",
	}}

	lb.Build(resolved)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lb.Reset()
		}()
	}
	wg.Wait()

	if lb.IsBuilt() {
		t.Fatal("expected not built after reset")
	}
}
