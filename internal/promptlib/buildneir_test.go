package promptlib

import (
	"testing"

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
	"github.com/NAEOS-foundation/naeos/internal/neir/model/storage"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
)

func fullNEIR() *model.NEIR {
	return &model.NEIR{
		Project: &project.Project{
			Name:        "full-project",
			Description: "A full project",
			Version:     "2.0.0",
		},
		Architecture: &architecture.Architecture{
			Pattern:    "clean",
			Principles: []string{"SRP", "OCP"},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./core", Description: "Core", Dependencies: nil},
		},
		Components: []component.Component{
			{Name: "handler", Kind: "interface", Module: "core"},
		},
		Services: []service.Service{
			{
				Name: "api", Kind: "http", Port: 8080,
				Endpoints: []service.Endpoint{
					{Method: "GET", Path: "/users", Action: "list"},
				},
			},
		},
		APIs: []api.API{
			{
				Name: "public", Version: "v1", Protocol: "rest",
				Endpoints: []api.APIEndpoint{
					{Method: "POST", Path: "/users", Summary: "Create user"},
				},
			},
		},
		Security: &security.Security{
			Authentication: &security.Authentication{
				Method:   "jwt",
				Provider: "auth0",
			},
			Authorization: &security.Authorization{
				Model: "rbac",
				Roles: []string{"admin", "user"},
			},
			Encryption: &security.Encryption{
				InTransit: true,
				AtRest:    true,
			},
		},
		Deployment: &deployment.Deployment{
			Strategy: "rolling",
			Environments: []deployment.Environment{
				{Name: "staging", Kind: "kubernetes"},
				{Name: "production", Kind: "kubernetes"},
			},
		},
		Testing: &testingmodel.Testing{
			Strategy:   "unit",
			Frameworks: []string{"jest", "playwright"},
		},
		Storage: []storage.Storage{
			{
				Name: "users-db", Type: "postgresql", Provider: "neon",
				Collections: []storage.Collection{
					{Name: "users"},
					{Name: "profiles"},
				},
			},
		},
		Infrastructure: &infrastructure.Infrastructure{
			Provider: "aws",
			Region:   "us-east-1",
			Resources: []infrastructure.Resource{
				{Name: "vpc", Kind: "network"},
			},
		},
		AI: &ai.AI{
			Models: []ai.Model{
				{Name: "gpt-4", Kind: "llm", Version: "latest"},
			},
		},
		Documentation: &docs.Documentation{
			ADRs: []docs.Doc{
				{Title: "ADR-001: Use hexagonal architecture"},
			},
			RFCs: []docs.Doc{
				{Title: "RFC-001: API design"},
			},
		},
	}
}

func TestBuildNEIRContextFull(t *testing.T) {
	neir := fullNEIR()
	ctx := buildNEIRContext(neir)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if ctx.Project.Name != "full-project" {
		t.Errorf("expected full-project, got %s", ctx.Project.Name)
	}
	if ctx.Architecture.Pattern != "clean" {
		t.Errorf("expected clean, got %s", ctx.Architecture.Pattern)
	}
	if len(ctx.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(ctx.Modules))
	}
	if len(ctx.Components) != 1 {
		t.Errorf("expected 1 component, got %d", len(ctx.Components))
	}
	if len(ctx.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(ctx.Services))
	}
	if len(ctx.Services[0].Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(ctx.Services[0].Endpoints))
	}
	if len(ctx.APIs) != 1 {
		t.Errorf("expected 1 api, got %d", len(ctx.APIs))
	}
	if len(ctx.APIs[0].Endpoints) != 1 {
		t.Errorf("expected 1 api endpoint, got %d", len(ctx.APIs[0].Endpoints))
	}
	if ctx.Security == nil {
		t.Fatal("expected non-nil security")
	}
	if ctx.Security.Authentication.Method != "jwt" {
		t.Errorf("expected jwt, got %s", ctx.Security.Authentication.Method)
	}
	if ctx.Security.Authorization.Model != "rbac" {
		t.Errorf("expected rbac, got %s", ctx.Security.Authorization.Model)
	}
	if !ctx.Security.Encryption.InTransit {
		t.Error("expected in-transit encryption")
	}
	if ctx.Deployment.Strategy != "rolling" {
		t.Errorf("expected rolling, got %s", ctx.Deployment.Strategy)
	}
	if len(ctx.Deployment.Environments) != 2 {
		t.Errorf("expected 2 environments, got %d", len(ctx.Deployment.Environments))
	}
	if ctx.Testing.Strategy != "unit" {
		t.Errorf("expected unit, got %s", ctx.Testing.Strategy)
	}
	if len(ctx.Storage) != 1 {
		t.Errorf("expected 1 storage, got %d", len(ctx.Storage))
	}
	if len(ctx.Storage[0].Collections) != 2 {
		t.Errorf("expected 2 collections, got %d", len(ctx.Storage[0].Collections))
	}
	if ctx.Infrastructure.Provider != "aws" {
		t.Errorf("expected aws, got %s", ctx.Infrastructure.Provider)
	}
	if len(ctx.Infrastructure.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(ctx.Infrastructure.Resources))
	}
	if len(ctx.AI.Models) != 1 {
		t.Errorf("expected 1 ai model, got %d", len(ctx.AI.Models))
	}
	if len(ctx.Documentation.ADRs) != 1 {
		t.Errorf("expected 1 adr, got %d", len(ctx.Documentation.ADRs))
	}
	if len(ctx.Documentation.RFCs) != 1 {
		t.Errorf("expected 1 rfc, got %d", len(ctx.Documentation.RFCs))
	}
}

func TestBuildNEIRContextNilSections(t *testing.T) {
	neir := &model.NEIR{}
	ctx := buildNEIRContext(neir)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if ctx.Project.Name != "" {
		t.Error("expected empty project")
	}
	if ctx.Architecture.Pattern != "" {
		t.Error("expected empty architecture")
	}
	if len(ctx.Modules) != 0 {
		t.Error("expected empty modules")
	}
	if len(ctx.Components) != 0 {
		t.Error("expected empty components")
	}
	if len(ctx.Services) != 0 {
		t.Error("expected empty services")
	}
	if ctx.Security != nil && ctx.Security.Authentication != nil {
		t.Error("expected nil security")
	}
	if ctx.Deployment != nil && ctx.Deployment.Strategy != "" {
		t.Error("expected nil deployment")
	}
	if ctx.Testing != nil && ctx.Testing.Strategy != "" {
		t.Error("expected nil testing")
	}
	if len(ctx.Storage) != 0 {
		t.Error("expected empty storage")
	}
	if ctx.Infrastructure != nil && ctx.Infrastructure.Provider != "" {
		t.Error("expected nil infrastructure")
	}
	if ctx.AI != nil && len(ctx.AI.Models) > 0 {
		t.Error("expected nil ai")
	}
	if ctx.Documentation != nil && len(ctx.Documentation.ADRs) > 0 {
		t.Error("expected nil documentation")
	}
}
