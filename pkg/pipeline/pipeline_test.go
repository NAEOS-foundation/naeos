package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type stubParser struct{}

func (stubParser) Parse(input string) (*parser.SpecDocument, error) {
	return &parser.SpecDocument{Raw: "injected:" + input}, nil
}

func TestPipelineRunProducesResult(t *testing.T) {
	p := New(Config{})

	result, err := p.Run("sample specification")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected a non-nil result")
	}
	if result.NEIR == nil {
		t.Fatal("expected NEIR to be built")
	}
	if len(result.Artifacts) == 0 {
		t.Fatal("expected at least one artifact")
	}
	if len(result.Tasks) == 0 {
		t.Fatal("expected at least one planned task")
	}
}

func TestPipelineUsesInjectedParser(t *testing.T) {
	p := New(Config{Parser: stubParser{}})
	result, err := p.Run("sample")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Source != "injected:sample" {
		t.Fatalf("expected injected source, got %q", result.Source)
	}
}

func TestPipelineWritesArtifactsToOutputDir(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "out")
	p := New(Config{OutputDir: outputDir})

	_, err := p.Run("sample specification")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected output dir to contain generated files")
	}
}

func TestConfigFromFileJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"pipeline":{"name":"demo","mode":"development","verbose":true,"output_dir":"./out"}}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := ConfigFromFile(path)
	if err != nil {
		t.Fatalf("ConfigFromFile returned error: %v", err)
	}
	if cfg.Name != "demo" {
		t.Fatalf("expected config name demo, got %q", cfg.Name)
	}
	if cfg.Mode != "development" {
		t.Fatalf("expected config mode development, got %q", cfg.Mode)
	}
	if !cfg.Verbose {
		t.Fatal("expected verbose to be true")
	}
	if cfg.OutputDir != "./out" {
		t.Fatalf("expected output dir ./out, got %q", cfg.OutputDir)
	}
}

func TestConfigFromFileYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := ConfigFromFile(path)
	if err != nil {
		t.Fatalf("ConfigFromFile returned error: %v", err)
	}
	if cfg.Name != "demo" {
		t.Fatalf("expected config name demo, got %q", cfg.Name)
	}
	if cfg.Mode != "development" {
		t.Fatalf("expected config mode development, got %q", cfg.Mode)
	}
	if !cfg.Verbose {
		t.Fatal("expected verbose to be true")
	}
	if cfg.OutputDir != "./out" {
		t.Fatalf("expected output dir ./out, got %q", cfg.OutputDir)
	}
}
