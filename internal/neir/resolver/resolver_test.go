package resolver

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
)

func TestResolveNilModel(t *testing.T) {
	r := NewProfileResolver()
	_, err := r.Resolve(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil model")
	}
}

func TestResolveNoConditions(t *testing.T) {
	r := NewProfileResolver()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{
			{Name: "mod-a", Path: "./internal/a"},
			{Name: "mod-b", Path: "./internal/b"},
		},
	}
	result, err := r.Resolve(neir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.RemovedModules) != 0 {
		t.Errorf("expected 0 removed, got %v", result.RemovedModules)
	}
	if len(neir.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(neir.Modules))
	}
}

func TestResolveConditionMatch(t *testing.T) {
	r := NewProfileResolver()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
			{Name: "analytics", Path: "./internal/analytics", Condition: "env == prod"},
		},
	}
	result, err := r.Resolve(neir, map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.RemovedModules) != 0 {
		t.Errorf("expected 0 removed, got %v", result.RemovedModules)
	}
	if len(neir.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(neir.Modules))
	}
}

func TestResolveConditionNoMatch(t *testing.T) {
	r := NewProfileResolver()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
			{Name: "analytics", Path: "./internal/analytics", Condition: "env == prod"},
		},
	}
	result, err := r.Resolve(neir, map[string]string{"env": "dev"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.RemovedModules) != 1 {
		t.Errorf("expected 1 removed, got %v", result.RemovedModules)
	}
	if result.RemovedModules[0] != "analytics" {
		t.Errorf("expected analytics removed, got %v", result.RemovedModules)
	}
	if len(neir.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(neir.Modules))
	}
}

func TestResolveServiceCondition(t *testing.T) {
	r := NewProfileResolver()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
		},
		Services: []service.Service{
			{Name: "api", Port: 8080, Condition: "feature == beta"},
			{Name: "admin", Port: 9090},
		},
	}
	result, err := r.Resolve(neir, map[string]string{"feature": "stable"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.RemovedServices) != 1 {
		t.Errorf("expected 1 service removed, got %v", result.RemovedServices)
	}
	if len(neir.Services) != 1 || neir.Services[0].Name != "admin" {
		t.Errorf("expected only admin service, got %+v", neir.Services)
	}
}

func TestResolveActiveProfile(t *testing.T) {
	r := NewProfileResolver()
	neir := &model.NEIR{
		Project:       &project.Project{Name: "test"},
		ActiveProfile: "staging",
		Deployment: &deployment.Deployment{
			Environments: []deployment.Environment{
				{Name: "staging", Variables: map[string]string{"env": "staging", "log_level": "debug"}},
				{Name: "production", Variables: map[string]string{"env": "prod", "log_level": "info"}},
			},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
			{Name: "debug-tools", Path: "./internal/debug", Condition: "log_level == debug"},
		},
	}
	result, err := r.Resolve(neir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.RemovedModules) != 0 {
		t.Errorf("expected debug-tools to be kept, got removed: %v", result.RemovedModules)
	}
	if len(neir.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(neir.Modules))
	}
}

func TestResolveDependencyCleanup(t *testing.T) {
	r := NewProfileResolver()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
			{Name: "experimental", Path: "./internal/exp", Condition: "!stable"},
			{Name: "legacy", Path: "./internal/legacy", Dependencies: []string{"experimental"}},
		},
	}
	result, err := r.Resolve(neir, map[string]string{"stable": "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.RemovedModules) != 1 {
		t.Errorf("expected 1 removed, got %v", result.RemovedModules)
	}
	if len(neir.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(neir.Modules))
	}
	for _, m := range neir.Modules {
		if m.Name == "legacy" {
			if len(m.Dependencies) != 0 {
				t.Errorf("expected legacy deps cleaned, got %v", m.Dependencies)
			}
		}
	}
}

func TestResolveProfileNotFound(t *testing.T) {
	r := NewProfileResolver()
	neir := &model.NEIR{
		Project:       &project.Project{Name: "test"},
		ActiveProfile: "nonexistent",
		Deployment: &deployment.Deployment{
			Environments: []deployment.Environment{
				{Name: "dev"},
				{Name: "prod"},
			},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
		},
	}
	result, err := r.Resolve(neir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var found bool
	for _, w := range result.Warnings {
		if w == "active profile \"nonexistent\" not found in deployment environments" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about missing profile, got %v", result.Warnings)
	}
}
