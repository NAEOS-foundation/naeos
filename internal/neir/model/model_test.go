package model

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model/ai"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/api"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/component"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/docs"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/domain"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/infrastructure"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/metadata"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/storage"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
)

func TestNEIR_ZeroValue(t *testing.T) {
	var n NEIR
	if n.Project != nil {
		t.Error("expected nil Project")
	}
	if n.Architecture != nil {
		t.Error("expected nil Architecture")
	}
	if n.Modules != nil {
		t.Error("expected nil Modules")
	}
	if n.Components != nil {
		t.Error("expected nil Components")
	}
	if n.Services != nil {
		t.Error("expected nil Services")
	}
	if n.APIs != nil {
		t.Error("expected nil APIs")
	}
	if n.Storage != nil {
		t.Error("expected nil Storage")
	}
	if n.Infrastructure != nil {
		t.Error("expected nil Infrastructure")
	}
	if n.Security != nil {
		t.Error("expected nil Security")
	}
	if n.AI != nil {
		t.Error("expected nil AI")
	}
	if n.Documentation != nil {
		t.Error("expected nil Documentation")
	}
	if n.Deployment != nil {
		t.Error("expected nil Deployment")
	}
	if n.Testing != nil {
		t.Error("expected nil Testing")
	}
	if n.Metadata != nil {
		t.Error("expected nil Metadata")
	}
	if n.Generation != nil {
		t.Error("expected nil Generation")
	}
}

func TestNEIR_Full(t *testing.T) {
	n := NEIR{
		Project:        &project.Project{Name: "naeos"},
		Architecture:   &architecture.Architecture{Pattern: architecture.PatternHexagonal},
		Domain:         &domain.Domain{Name: "platform"},
		Modules:        []module.Module{{Name: "core", Path: "./core"}},
		Components:     []component.Component{{Name: "handler", Kind: component.KindHandler}},
		Services:       []service.Service{{Name: "api", Kind: service.KindHTTP}},
		APIs:           []api.API{{Name: "public", Protocol: api.ProtocolHTTP}},
		Storage:        []storage.Storage{{Name: "db", Type: storage.TypeSQL}},
		Infrastructure: &infrastructure.Infrastructure{Provider: infrastructure.ProviderAWS},
		Security:       &security.Security{Authentication: &security.Authentication{Method: "oauth2"}},
		AI:             &ai.AI{Models: []ai.Model{{Name: "gpt4"}}},
		Documentation:  &docs.Documentation{Guides: []docs.Doc{{Title: "Getting Started"}}},
		Deployment:     &deployment.Deployment{Strategy: deployment.StrategyRolling},
		Testing:        &testingmodel.Testing{Strategy: testingmodel.StrategyUnit},
		Metadata:       &metadata.Metadata{NEIRVersion: "2.0"},
		Generation:     &generation.GenerationConfig{OutputDir: "./out"},
	}
	if n.Project.Name != "naeos" {
		t.Errorf("expected naeos, got %s", n.Project.Name)
	}
	if n.Architecture.Pattern != architecture.PatternHexagonal {
		t.Errorf("expected hexagonal, got %s", n.Architecture.Pattern)
	}
	if len(n.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(n.Modules))
	}
	if len(n.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(n.Services))
	}
	if n.Infrastructure.Provider != infrastructure.ProviderAWS {
		t.Errorf("expected aws, got %s", n.Infrastructure.Provider)
	}
	if n.Security.Authentication.Method != "oauth2" {
		t.Errorf("expected oauth2, got %s", n.Security.Authentication.Method)
	}
	if n.AI.Models[0].Name != "gpt4" {
		t.Errorf("expected gpt4, got %s", n.AI.Models[0].Name)
	}
	if n.Documentation.Guides[0].Title != "Getting Started" {
		t.Errorf("expected Getting Started, got %s", n.Documentation.Guides[0].Title)
	}
	if n.Deployment.Strategy != deployment.StrategyRolling {
		t.Errorf("expected rolling, got %s", n.Deployment.Strategy)
	}
	if n.Testing.Strategy != testingmodel.StrategyUnit {
		t.Errorf("expected unit, got %s", n.Testing.Strategy)
	}
	if n.Metadata.NEIRVersion != "2.0" {
		t.Errorf("expected 2.0, got %s", n.Metadata.NEIRVersion)
	}
	if n.Generation.OutputDir != "./out" {
		t.Errorf("expected ./out, got %s", n.Generation.OutputDir)
	}
}
