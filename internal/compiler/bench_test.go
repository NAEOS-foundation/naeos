package compiler

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

func benchmarkNEIR() *model.NEIR {
	return &model.NEIR{
		Project: &project.Project{
			Name:        "bench-project",
			Description: "A benchmark project with multiple modules and services",
			Version:     "2.0.0",
		},
		Modules: []module.Module{
			{Name: "core", Path: "./core"},
			{Name: "api", Path: "./api", Dependencies: []string{"core"}},
			{Name: "worker", Path: "./worker", Dependencies: []string{"core"}},
		},
		Services: []service.Service{
			{Name: "http-api", Kind: "http", Port: 8080},
			{Name: "grpc-worker", Kind: "grpc", Port: 9090},
		},
	}
}

func BenchmarkCompileAll(b *testing.B) {
	c := New()
	c.Register(&stubAdapter{target: TargetCopilot})
	c.Register(&stubAdapter{target: TargetClaude})
	c.Register(&stubAdapter{target: TargetCursor})
	c.Register(&stubAdapter{target: TargetGemini})
	c.Register(&stubAdapter{target: TargetCodex})
	c.Register(&stubAdapter{target: TargetOpenCode})
	c.Register(&stubAdapter{target: TargetWindsurf})

	neir := benchmarkNEIR()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := c.CompileAll(neir)
		if len(results) != 7 {
			b.Fatalf("expected 7 results, got %d", len(results))
		}
	}
}

func BenchmarkCompileSingleAdapter(b *testing.B) {
	c := New()
	c.Register(&stubAdapter{target: TargetCopilot})

	neir := benchmarkNEIR()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := c.CompileAll(neir)
		if len(results) != 1 {
			b.Fatalf("expected 1 result, got %d", len(results))
		}
	}
}

func BenchmarkBuildProjectContext(b *testing.B) {
	neir := benchmarkNEIR()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := buildProjectContext(neir)
		if ctx == "" {
			b.Fatal("expected non-empty context")
		}
	}
}
