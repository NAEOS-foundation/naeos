package pipeline

import (
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
