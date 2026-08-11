package builder

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
)

func TestLazyNEIRBasicFields(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"project":        "myapp",
			"active_profile": "production",
			"inherits":       "base",
		},
	}
	lazy := newLazyNEIR(raw)
	if lazy.Project().Name != "myapp" {
		t.Errorf("expected project 'myapp', got %q", lazy.Project().Name)
	}
	if lazy.ActiveProfile() != "production" {
		t.Errorf("expected active_profile 'production', got %q", lazy.ActiveProfile())
	}
	if lazy.Inherits() != "base" {
		t.Errorf("expected inherits 'base', got %q", lazy.Inherits())
	}
}

func TestLazyNEIRModulesLazy(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"project": "myapp",
			"modules": []any{
				map[string]any{"name": "core", "path": "./core"},
				map[string]any{"name": "api", "path": "./api"},
			},
		},
	}
	lazy := newLazyNEIR(raw)
	mods := lazy.Modules()
	if len(mods) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(mods))
	}
	if mods[0].Name != "core" {
		t.Errorf("expected module 'core', got %q", mods[0].Name)
	}
	if mods[1].Name != "api" {
		t.Errorf("expected module 'api', got %q", mods[1].Name)
	}
	toNEIR := lazy.ToNEIR()
	if len(toNEIR.Modules) != 2 {
		t.Errorf("expected 2 modules in full NEIR, got %d", len(toNEIR.Modules))
	}
}

func TestLazyNEIRServicesLazy(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"project": "myapp",
			"services": []any{
				map[string]any{"name": "api", "port": 8080},
			},
		},
	}
	lazy := newLazyNEIR(raw)
	svcs := lazy.Services()
	if len(svcs) != 1 {
		t.Fatalf("expected 1 service, got %d", len(svcs))
	}
	if svcs[0].Name != "api" {
		t.Errorf("expected service 'api', got %q", svcs[0].Name)
	}
	if svcs[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", svcs[0].Port)
	}
}

func TestLazyNEIRLoadAll(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"project": "myapp",
			"modules": []any{
				map[string]any{"name": "core", "path": "./core"},
			},
			"services": []any{
				map[string]any{"name": "api", "port": 8080, "kind": "http"},
			},
		},
	}
	lazy := newLazyNEIR(raw)
	full := lazy.ToNEIR()
	if full.Project == nil || full.Project.Name != "myapp" {
		t.Error("expected project in full NEIR")
	}
	if len(full.Modules) != 1 || full.Modules[0].Name != "core" {
		t.Error("expected modules in full NEIR")
	}
	if len(full.Services) != 1 || full.Services[0].Name != "api" {
		t.Error("expected services in full NEIR")
	}
}

func TestLazyNEIRLoadsOnlyOnce(t *testing.T) {
	callCount := 0
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"project": "myapp",
			"modules": []any{
				map[string]any{"name": "core", "path": "./core"},
			},
		},
	}
	lazy := newLazyNEIR(raw)
	_ = lazy.Modules()
	callCount++
	_ = lazy.Modules()
	callCount++
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
	if len(lazy.neir.Modules) != 1 {
		t.Error("expected modules to be loaded once")
	}
}

func TestLazyNEIRNilRaw(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{},
	}
	lazy := newLazyNEIR(raw)
	if lazy.Project() != nil {
		t.Error("expected nil project for empty context")
	}
	if len(lazy.Modules()) != 0 {
		t.Error("expected empty modules for empty context")
	}
	if len(lazy.Services()) != 0 {
		t.Error("expected empty services for empty context")
	}
}

func TestDefaultBuilderBuildLazy(t *testing.T) {
	b := DefaultBuilder{}
	resolved := &resolver.ResolvedSpec{
		Context: map[string]any{
			"project": "testapp",
			"modules": []any{
				map[string]any{"name": "mod1", "path": "./mod1"},
			},
		},
	}
	lazy, err := b.BuildLazy(resolved)
	if err != nil {
		t.Fatalf("BuildLazy returned error: %v", err)
	}
	if lazy == nil {
		t.Fatal("expected non-nil LazyNEIR")
	}
	if lazy.Project().Name != "testapp" {
		t.Errorf("expected project 'testapp', got %q", lazy.Project().Name)
	}
	if len(lazy.Modules()) != 1 {
		t.Errorf("expected 1 module, got %d", len(lazy.Modules()))
	}
}

func TestDefaultBuilderBuildLazyNil(t *testing.T) {
	b := DefaultBuilder{}
	_, err := b.BuildLazy(nil)
	if err == nil {
		t.Fatal("expected error for nil resolved")
	}
}

func TestDefaultBuilderBuildLazyWrongType(t *testing.T) {
	b := DefaultBuilder{}
	_, err := b.BuildLazy("not a resolved spec")
	if err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestLazyNEIRComponentsAPIsStorage(t *testing.T) {
	raw := &resolver.ResolvedSpec{Context: map[string]any{}}
	lazy := newLazyNEIR(raw)
	if lazy.Components() == nil {
		t.Error("expected non-nil Components")
	}
	if lazy.APIs() == nil {
		t.Error("expected non-nil APIs")
	}
	if lazy.Storage() == nil {
		t.Error("expected non-nil Storage")
	}
}

func TestLazyNEIRToNEIRIdempotent(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"project": "myapp",
		},
	}
	lazy := newLazyNEIR(raw)
	a := lazy.ToNEIR()
	b := lazy.ToNEIR()
	if a != b {
		t.Error("expected ToNEIR to return same instance")
	}
}

func TestLazyNEIRMetadata(t *testing.T) {
	raw := &resolver.ResolvedSpec{Context: map[string]any{}}
	lazy := newLazyNEIR(raw)
	meta := lazy.Metadata()
	if meta == nil {
		t.Fatal("expected non-nil Metadata")
	}
	if meta.NEIRVersion != "0.1.0" {
		t.Errorf("expected NEIRVersion '0.1.0', got %q", meta.NEIRVersion)
	}
}

func TestLazyNEIRArchitecture(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"architecture": map[string]any{
				"pattern": "clean",
			},
		},
	}
	lazy := newLazyNEIR(raw)
	arch := lazy.Architecture()
	if arch == nil {
		t.Fatal("expected non-nil Architecture")
	}
	if string(arch.Pattern) != "clean" {
		t.Errorf("expected pattern 'clean', got %q", string(arch.Pattern))
	}
}

func TestLazyNEIRGeneration(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"generation": map[string]any{
				"languages": []any{"go", "ts"},
			},
		},
	}
	lazy := newLazyNEIR(raw)
	gen := lazy.Generation()
	if gen == nil {
		t.Fatal("expected non-nil Generation")
	}
	if len(gen.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(gen.Languages))
	}
}

func TestLazyNEIRDeployment(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"deployment": map[string]any{
				"strategy": "blue-green",
			},
		},
	}
	lazy := newLazyNEIR(raw)
	deploy := lazy.Deployment()
	if deploy == nil {
		t.Fatal("expected non-nil Deployment")
	}
	if string(deploy.Strategy) != "blue-green" {
		t.Errorf("expected strategy 'blue-green', got %q", string(deploy.Strategy))
	}
}

func TestLazyNEIRTesting(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"testing": map[string]any{
				"strategy": "unit",
			},
		},
	}
	lazy := newLazyNEIR(raw)
	test := lazy.Testing()
	if test == nil {
		t.Fatal("expected non-nil Testing")
	}
	if string(test.Strategy) != "unit" {
		t.Errorf("expected strategy 'unit', got %q", string(test.Strategy))
	}
}

func TestLazyNEIRInfrastructure(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"cloud": map[string]any{
				"provider": "aws",
			},
		},
	}
	lazy := newLazyNEIR(raw)
	infra := lazy.Infrastructure()
	if infra == nil {
		t.Fatal("expected non-nil Infrastructure")
	}
	if string(infra.Provider) != "aws" {
		t.Errorf("expected provider 'aws', got %q", string(infra.Provider))
	}
}

func TestLazyNEIRBackwardCompat(t *testing.T) {
	raw := &resolver.ResolvedSpec{
		Context: map[string]any{
			"project": "myapp",
		},
	}
	lazy := newLazyNEIR(raw)
	full := lazy.ToNEIR()
	var _ = full
	_ = full.Project
	_ = full.Modules
	_ = full.Services
	_ = full.Architecture
}

func TestDefaultBuilderLazyBuilderInterface(t *testing.T) {
	b := DefaultBuilder{}
	lazyB, ok := any(b).(interface {
		BuildLazy(resolved any) (*LazyNEIR, error)
	})
	if !ok {
		t.Fatal("DefaultBuilder should implement LazyBuilder")
	}
	resolved := &resolver.ResolvedSpec{
		Context: map[string]any{"project": "test"},
	}
	lazy, err := lazyB.BuildLazy(resolved)
	if err != nil {
		t.Fatalf("BuildLazy: %v", err)
	}
	if lazy == nil {
		t.Fatal("expected non-nil lazy NEIR")
	}
}
