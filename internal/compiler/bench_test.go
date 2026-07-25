package compiler

import (
	"fmt"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/api"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/component"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/infrastructure"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/storage"
)

type slowAdapter struct {
	target Target
	sleep  time.Duration
}

func (a *slowAdapter) Target() Target { return a.target }
func (a *slowAdapter) Compile(neir *model.NEIR) (*CompiledOutput, error) {
	time.Sleep(a.sleep)
	return &CompiledOutput{
		Target:     a.target,
		Files:      []OutputFile{{Path: "test.md", Content: "bench", Kind: "instructions"}},
		Summary:    "bench output",
		CompiledAt: time.Now(),
	}, nil
}

func smallNEIR() *model.NEIR {
	return &model.NEIR{
		Project: &project.Project{Name: "small", Version: "1.0.0"},
		Modules: []module.Module{
			{Name: "core", Path: "./core"},
			{Name: "api", Path: "./api", Dependencies: []string{"core"}},
		},
		Services: []service.Service{
			{Name: "http-api", Kind: service.KindHTTP, Port: 8080},
		},
		Architecture: &architecture.Architecture{
			Pattern: "layered",
		},
	}
}

func mediumNEIR() *model.NEIR {
	mods := make([]module.Module, 10)
	for i := range mods {
		mods[i] = module.Module{
			Name: fmt.Sprintf("mod-%d", i),
			Path: fmt.Sprintf("./internal/mod%d", i),
		}
		if i > 0 {
			mods[i].Dependencies = []string{fmt.Sprintf("mod-%d", i-1)}
		}
	}

	svcs := make([]service.Service, 5)
	for i := range svcs {
		kinds := []service.ServiceKind{service.KindHTTP, service.KindGRPC, service.KindWorker, service.KindCLI, service.KindJob}
		svcs[i] = service.Service{
			Name: fmt.Sprintf("svc-%d", i),
			Kind: kinds[i%5],
			Port: 8000 + i,
		}
	}

	return &model.NEIR{
		Project: &project.Project{Name: "medium", Version: "2.0.0"},
		Modules: mods,
		Services: svcs,
		Architecture: &architecture.Architecture{
			Pattern:    "clean",
			Principles: []string{"DI", "SRP", "OCP"},
		},
		Storage: []storage.Storage{
			{Name: "postgres", Type: "postgresql"},
			{Name: "redis", Type: "redis"},
		},
		Deployment: &deployment.Deployment{Strategy: "rolling"},
	}
}

func largeNEIR() *model.NEIR {
	mods := make([]module.Module, 50)
	for i := range mods {
		mods[i] = module.Module{
			Name: fmt.Sprintf("mod-%d", i),
			Path: fmt.Sprintf("./pkg/mod%d", i),
		}
		if i > 0 {
			mods[i].Dependencies = []string{fmt.Sprintf("mod-%d", (i-1)/2)}
		}
	}

	svcs := make([]service.Service, 20)
	for i := range svcs {
		kinds := []service.ServiceKind{service.KindHTTP, service.KindGRPC, service.KindWorker, service.KindCLI, service.KindJob}
		svcs[i] = service.Service{
			Name: fmt.Sprintf("svc-%d", i),
			Kind: kinds[i%5],
			Port: 8000 + i,
		}
	}

	comps := make([]component.Component, 10)
	for i := range comps {
		comps[i] = component.Component{
			Name: fmt.Sprintf("comp-%d", i),
			Kind: component.KindHandler,
		}
	}

	apis := make([]api.API, 10)
	for i := range apis {
		apis[i] = api.API{
			Name:     fmt.Sprintf("api-%d", i),
			Protocol: api.ProtocolHTTP,
		}
	}

	return &model.NEIR{
		Project: &project.Project{Name: "large", Version: "3.0.0", Description: "Large project with 50 modules"},
		Modules:  mods,
		Services: svcs,
		Components: comps,
		APIs: apis,
		Architecture: &architecture.Architecture{
			Pattern:    "hexagonal",
			Principles: []string{"DI", "SRP", "OCP", "LSP", "ISP", "DIP"},
		},
		Storage: []storage.Storage{
			{Name: "primary-db", Type: "postgresql"},
			{Name: "cache", Type: "redis"},
			{Name: "analytics", Type: "clickhouse"},
			{Name: "search", Type: "elasticsearch"},
			{Name: "queue", Type: "rabbitmq"},
		},
		Deployment: &deployment.Deployment{Strategy: "blue-green"},
		Security: &security.Security{
			Authentication: &security.Authentication{Method: "oauth2"},
		},
		Infrastructure: &infrastructure.Infrastructure{
			Provider: "aws",
			Region:   "us-east-1",
		},
	}
}

func runBenchmark(b *testing.B, neir *model.NEIR, numAdapters int) {
	c := New()
	targets := []Target{TargetCopilot, TargetClaude, TargetCursor, TargetGemini, TargetCodex, TargetOpenCode, TargetWindsurf}
	for i := 0; i < numAdapters && i < len(targets); i++ {
		c.Register(&slowAdapter{target: targets[i], sleep: 0})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := c.CompileAll(neir)
		if len(results) != numAdapters {
			b.Fatalf("expected %d results, got %d", numAdapters, len(results))
		}
	}
}

func BenchmarkCompileAll_Small(b *testing.B)  { runBenchmark(b, smallNEIR(), 3) }
func BenchmarkCompileAll_Medium(b *testing.B) { runBenchmark(b, mediumNEIR(), 5) }
func BenchmarkCompileAll_Large(b *testing.B)  { runBenchmark(b, largeNEIR(), 7) }

func BenchmarkCompileSingleAdapter_Small(b *testing.B) { runBenchmark(b, smallNEIR(), 1) }
func BenchmarkCompileSingleAdapter_Large(b *testing.B) { runBenchmark(b, largeNEIR(), 1) }

func BenchmarkCompileAll_Small_1msPerAdapter(b *testing.B) {
	neir := smallNEIR()
	c := New()
	targets := []Target{TargetCopilot, TargetClaude, TargetCursor}
	for _, tgt := range targets {
		c.Register(&slowAdapter{target: tgt, sleep: time.Millisecond})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := c.CompileAll(neir)
		if len(results) != 3 {
			b.Fatalf("expected 3 results, got %d", len(results))
		}
	}
}

func BenchmarkBuildProjectContext_Small(b *testing.B) {
	neir := smallNEIR()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if ctx := buildProjectContext(neir); ctx == "" {
			b.Fatal("expected non-empty context")
		}
	}
}

func BenchmarkBuildProjectContext_Large(b *testing.B) {
	neir := largeNEIR()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if ctx := buildProjectContext(neir); ctx == "" {
			b.Fatal("expected non-empty context")
		}
	}
}
