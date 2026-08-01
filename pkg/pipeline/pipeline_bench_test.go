package pipeline

// Benchmark target times:
// - small (BenchmarkPipelineRun): <1s
// - medium (BenchmarkPipelineRun_Medium): <5s
// - large (BenchmarkPipelineRun_Large): <30s

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

const benchSpec = `project:
  name: benchapp
  version: "1.0.0"
services:
  - name: api
    port: 8080
  - name: worker
    port: 9090
`

func genMediumSpec() string {
	var sb strings.Builder
	sb.WriteString("project:\n  name: mediumapp\n  version: \"1.0.0\"\n\nmodules:\n")
	for i := range 50 {
		sb.WriteString(fmt.Sprintf("  - name: module-%d\n    path: ./modules/module-%d\n", i, i))
	}
	sb.WriteString("\nservices:\n")
	for i := range 5 {
		sb.WriteString(fmt.Sprintf("  - name: svc-%d\n    port: %d\n    kind: http\n", i, 8000+i))
	}
	return sb.String()
}

func genLargeSpec() string {
	var sb strings.Builder
	sb.WriteString("project:\n  name: largeapp\n  version: \"1.0.0\"\n\nmodules:\n")
	for i := range 500 {
		sb.WriteString(fmt.Sprintf("  - name: module-%d\n    path: ./modules/module-%d\n", i, i))
	}
	sb.WriteString("\nservices:\n")
	for i := range 10 {
		sb.WriteString(fmt.Sprintf("  - name: svc-%d\n    port: %d\n    kind: http\n", i, 8000+i))
	}
	return sb.String()
}

func BenchmarkPipelineRun(b *testing.B) {
	b.ReportAllocs()
	cfg := Config{
		Name:      "benchapp",
		OutputDir: b.TempDir(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(strings.TrimSpace(benchSpec))
	}
}

func BenchmarkPipelineRun_Medium(b *testing.B) {
	b.ReportAllocs()
	spec := genMediumSpec()
	cfg := Config{
		Name:      "mediumapp",
		OutputDir: b.TempDir(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(spec)
	}
}

func BenchmarkPipelineRun_Large(b *testing.B) {
	b.ReportAllocs()
	spec := genLargeSpec()
	cfg := Config{
		Name:      "largeapp",
		OutputDir: b.TempDir(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(spec)
	}
}

func BenchmarkPipelineRun_ParallelSmall(b *testing.B) {
	b.ReportAllocs()
	parallel := true
	cfg := Config{
		Name:      "benchapp",
		OutputDir: b.TempDir(),
		Parallel:  &parallel,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(strings.TrimSpace(benchSpec))
	}
}

func BenchmarkPipelineRun_ParallelMedium(b *testing.B) {
	b.ReportAllocs()
	parallel := true
	spec := genMediumSpec()
	cfg := Config{
		Name:      "mediumapp",
		OutputDir: b.TempDir(),
		Parallel:  &parallel,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(spec)
	}
}

func BenchmarkPipelineRun_ParallelLarge(b *testing.B) {
	b.ReportAllocs()
	parallel := true
	spec := genLargeSpec()
	cfg := Config{
		Name:      "largeapp",
		OutputDir: b.TempDir(),
		Parallel:  &parallel,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(spec)
	}
}

func BenchmarkPipelineRun_Profiled(b *testing.B) {
	b.ReportAllocs()
	cfg := Config{
		Name:      "benchapp",
		OutputDir: b.TempDir(),
		Profiling: true,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(strings.TrimSpace(benchSpec))
	}
}

func BenchmarkPipelineRun_ProfiledParallel(b *testing.B) {
	b.ReportAllocs()
	parallel := true
	cfg := Config{
		Name:      "benchapp",
		OutputDir: b.TempDir(),
		Parallel:  &parallel,
		Profiling: true,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(strings.TrimSpace(benchSpec))
	}
}

func BenchmarkPipelineValidate(b *testing.B) {
	b.ReportAllocs()
	cfg := Config{
		Name:      "benchapp",
		OutputDir: b.TempDir(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Validate(strings.TrimSpace(benchSpec))
	}
}

func BenchmarkPipelineNew(b *testing.B) {
	b.ReportAllocs()
	cfg := Config{
		Name:      "benchapp",
		OutputDir: b.TempDir(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(cfg)
	}
}

func BenchmarkPipelineRun_Memory(b *testing.B) {
	b.ReportAllocs()
	cfg := Config{
		Name:      "benchapp",
		OutputDir: b.TempDir(),
	}

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := New(cfg)
		p.Run(strings.TrimSpace(benchSpec))
	}
	b.StopTimer()
	runtime.ReadMemStats(&after)
	b.ReportMetric(float64(after.TotalAlloc-before.TotalAlloc)/float64(b.N), "alloced/op")
	b.ReportMetric(float64(after.Mallocs-before.Mallocs)/float64(b.N), "mallocs/op")
}
