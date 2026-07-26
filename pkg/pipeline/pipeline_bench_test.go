package pipeline

import (
	"fmt"
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
	cfg := Config{
		Name:      "benchapp",
		OutputDir: b.TempDir(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(cfg)
	}
}
