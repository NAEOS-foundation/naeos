//go:build integration

package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func TestE2EPipelineFullRun(t *testing.T) {
	spec := `project:
  name: testapp
  version: "1.0.0"
services:
  - name: api
    port: 8080
`
	outputDir := t.TempDir()
	cfg := pipeline.Config{
		Name:      "testapp",
		OutputDir: outputDir,
		Languages: []string{"go"},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatalf("construct pipeline: %v", err)
	}

	result, err := p.Run(strings.TrimSpace(spec))
	if err != nil {
		t.Fatalf("pipeline run: %v", err)
	}

	if len(result.Artifacts) == 0 {
		t.Fatal("expected at least 1 artifact")
	}

	if result.NEIR == nil {
		t.Fatal("expected NEIR in result")
	}

	if result.NEIR.Project == nil {
		t.Fatal("expected project in NEIR")
	}
}

func TestE2EPipelineArtifactContent(t *testing.T) {
	spec := `project:
  name: testapp
services:
  - name: api
    port: 8080
`
	outputDir := t.TempDir()
	cfg := pipeline.Config{
		Name:      "testapp",
		OutputDir: outputDir,
		Languages: []string{"go"},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.Run(strings.TrimSpace(spec))
	if err != nil {
		t.Fatal(err)
	}

	foundGo := false
	for _, a := range result.Artifacts {
		if strings.HasSuffix(a.Path, ".go") {
			foundGo = true
			content := string(a.Content)
			if !strings.Contains(content, "package") {
				t.Errorf("expected Go package declaration in %s", a.Path)
			}
		}
	}
	if !foundGo {
		t.Error("expected at least one .go artifact")
	}
}

func TestE2EPipelineWritesFiles(t *testing.T) {
	spec := `project:
  name: testapp
services:
  - name: api
    port: 8080
`
	outputDir := t.TempDir()
	cfg := pipeline.Config{
		Name:      "testapp",
		OutputDir: outputDir,
		Languages: []string{"go"},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.Run(strings.TrimSpace(spec))
	if err != nil {
		t.Fatal(err)
	}

	for _, a := range result.Artifacts {
		path := filepath.Join(outputDir, a.Path)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("artifact %s not written: %v", a.Path, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("artifact %s is empty", a.Path)
		}
	}
}

func TestE2EPipelineTypeScript(t *testing.T) {
	spec := `project:
  name: tsapp
services:
  - name: web
    port: 3000
`
	outputDir := t.TempDir()
	cfg := pipeline.Config{
		Name:      "tsapp",
		OutputDir: outputDir,
		Languages: []string{"typescript"},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.Run(strings.TrimSpace(spec))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Artifacts) == 0 {
		t.Fatal("expected at least 1 artifact for TypeScript")
	}
}

func TestE2EPipelineMinimalSpec(t *testing.T) {
	spec := `project:
  name: minimal
`
	outputDir := t.TempDir()
	cfg := pipeline.Config{
		Name:      "minimal",
		OutputDir: outputDir,
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.Run(strings.TrimSpace(spec))
	if err != nil {
		t.Fatalf("minimal spec pipeline failed: %v", err)
	}

	if result.NEIR == nil {
		t.Fatal("expected NEIR from minimal spec")
	}
}

func TestE2EPipelineResultNEIRServices(t *testing.T) {
	spec := `project:
  name: porttest
services:
  - name: grpc
    port: 50051
  - name: http
    port: 3000
  - name: metrics
    port: 9090
`
	outputDir := t.TempDir()
	cfg := pipeline.Config{
		Name:      "porttest",
		OutputDir: outputDir,
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.Run(strings.TrimSpace(spec))
	if err != nil {
		t.Fatal(err)
	}

	if result.NEIR == nil {
		t.Fatal("expected NEIR")
	}

	if result.NEIR.Project == nil {
		t.Fatal("expected project in NEIR")
	}
}

func TestE2EPipelineArtifactCount(t *testing.T) {
	spec := `project:
  name: counttest
services:
  - name: api
    port: 8080
  - name: worker
    port: 9090
`
	outputDir := t.TempDir()
	cfg := pipeline.Config{
		Name:      "counttest",
		OutputDir: outputDir,
		Languages: []string{"go"},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.Run(strings.TrimSpace(spec))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Artifacts) < 2 {
		t.Errorf("expected at least 2 artifacts for 2 services, got %d", len(result.Artifacts))
	}

	if len(result.Tasks) == 0 {
		t.Error("expected at least 1 task in scheduler result")
	}
}

func TestE2EPipelineReview(t *testing.T) {
	spec := `project:
  name: reviewtest
services:
  - name: api
    port: 8080
`
	outputDir := t.TempDir()
	cfg := pipeline.Config{
		Name:      "reviewtest",
		OutputDir: outputDir,
		Languages: []string{"go"},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.Run(strings.TrimSpace(spec))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Reviews) == 0 {
		t.Error("expected at least 1 review result")
	}
}
