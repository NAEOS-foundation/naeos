package adaptertest

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/ai"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/api"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/component"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/docs"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/infrastructure"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	storagemodel "github.com/NAEOS-foundation/naeos/internal/neir/model/storage"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
	"github.com/NAEOS-foundation/naeos/internal/testutil"
)

func BasicNEIR() *model.NEIR {
	return &model.NEIR{
		Project: &project.Project{
			Name:        "test-project",
			Description: "A test project",
			Version:     "1.0.0",
		},
		Architecture: &architecture.Architecture{
			Pattern:    "hexagonal",
			Principles: []string{"DI", "SRP"},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./core", Description: "Core module", Dependencies: []string{}},
			{Name: "api", Path: "./api", Description: "API module", Dependencies: []string{"core"}},
		},
		Services: []service.Service{
			{Name: "http-api", Kind: "http", Port: 8080},
		},
	}
}

func FullNEIR() *model.NEIR {
	return &model.NEIR{
		Project: &project.Project{
			Name:        "test-project",
			Version:     "1.0.0",
			Description: "A full test project for integration",
			License:     "Apache-2.0",
			Authors:     []string{"dev1", "dev2"},
			Repository:  "https://github.com/test/repo",
		},
		Architecture: &architecture.Architecture{
			Pattern:    architecture.PatternLayered,
			Principles: []string{"separation-of-concerns", "single-responsibility"},
			Layers: []architecture.Layer{
				{Name: "api", Description: "API layer", Modules: []string{"http"}},
				{Name: "domain", Description: "Domain layer", Modules: []string{"core"}},
			},
		},
		Modules: []module.Module{
			{Name: "core", Path: "internal/core", Description: "Core logic", Dependencies: []string{}},
			{Name: "http", Path: "internal/http", Description: "HTTP handlers", Dependencies: []string{"core"}},
		},
		Services: []service.Service{
			{
				Name: "api-gateway",
				Kind: service.KindHTTP,
				Port: 8080,
				Endpoints: []service.Endpoint{
					{Method: "GET", Path: "/health", Action: "healthcheck"},
					{Method: "POST", Path: "/api/v1/users", Action: "createUser"},
				},
			},
		},
		Components: []component.Component{
			{Name: "UserService", Kind: component.KindService, Module: "core", Description: "User management"},
			{Name: "UserRepo", Kind: component.KindRepository, Module: "core"},
		},
		APIs: []api.API{
			{
				Name:     "users-api",
				Version:  "v1",
				Protocol: api.ProtocolHTTP,
				Endpoints: []api.APIEndpoint{
					{Method: "GET", Path: "/users", Summary: "List users"},
					{Method: "POST", Path: "/users", Summary: "Create user"},
				},
			},
		},
		Storage: []storagemodel.Storage{
			{Name: "users-db", Type: storagemodel.TypeSQL, Provider: "postgres", Connection: "postgres://localhost:5432/users"},
		},
		Infrastructure: &infrastructure.Infrastructure{
			Provider: infrastructure.ProviderAWS,
			Region:   "us-east-1",
			Resources: []infrastructure.Resource{
				{Name: "main-vpc", Kind: "vpc", Type: "aws_vpc"},
			},
		},
		Security: &security.Security{
			Authentication: &security.Authentication{Method: "jwt", Provider: "auth0"},
			Authorization:  &security.Authorization{Model: "rbac", Roles: []string{"admin", "user"}},
			Encryption:     &security.Encryption{InTransit: true, AtRest: true, Algorithm: "aes-256"},
		},
		AI: &ai.AI{
			Models: []ai.Model{
				{Name: "code-assistant", Kind: "llm", Version: "1.0"},
			},
		},
		Documentation: &docs.Documentation{
			ADRs: []docs.Doc{{Title: "ADR-001: Use layered architecture", Kind: docs.KindADR}},
			RFCs: []docs.Doc{{Title: "RFC-001: API versioning", Kind: docs.KindRFC}},
		},
		Deployment: &deployment.Deployment{
			Strategy: deployment.StrategyRolling,
			Environments: []deployment.Environment{
				{Name: "staging", Kind: "staging"},
			},
		},
		Testing: &testingmodel.Testing{
			Strategy:   testingmodel.StrategyUnit,
			Frameworks: []string{"go-test", "jest"},
		},
	}
}

func Contains(s, substr string) bool {
	return testutil.Contains(s, substr)
}

func ValidateCompilerOutput(t *testing.T, output *compiler.CompiledOutput) {
	t.Helper()
	if output == nil {
		t.Fatal("expected non-nil compiled output")
	}
	if output.Target == "" {
		t.Error("expected non-empty target")
	}
	if output.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if len(output.Files) == 0 {
		t.Error("expected at least one output file")
	}
	for _, f := range output.Files {
		if f.Path == "" {
			t.Error("expected non-empty file path")
		}
		if f.Content == "" {
			t.Errorf("expected non-empty content for %s", f.Path)
		}
	}
}

func ValidateArtifacts(t *testing.T, artifacts []engine.Artifact) {
	t.Helper()
	if len(artifacts) == 0 {
		t.Fatal("expected at least one artifact")
	}
	for _, a := range artifacts {
		if a.Path == "" {
			t.Error("expected non-empty artifact path")
		}
		if len(a.Content) == 0 {
			t.Errorf("expected non-empty content for %s", a.Path)
		}
	}
}

func AssertFileContains(t *testing.T, artifacts []engine.Artifact, path, substr string) bool {
	t.Helper()
	for _, a := range artifacts {
		if a.Path == path {
			if strings.Contains(string(a.Content), substr) {
				return true
			}
			t.Errorf("file %s does not contain %q", path, substr)
			return false
		}
	}
	t.Errorf("file %s not found in artifacts", path)
	return false
}
