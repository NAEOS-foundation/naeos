package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	cfgpkg "github.com/NAEOS-foundation/naeos/pkg/config"
	"github.com/NAEOS-foundation/naeos/internal/neir/builder"
	"github.com/NAEOS-foundation/naeos/internal/neir/validator"
	"github.com/NAEOS-foundation/naeos/internal/planner/scheduler"
	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
)

// Config provides optional dependencies and runtime settings for the pipeline.
type Config struct {
	Name       string
	Mode       string
	Verbose    bool
	OutputDir  string
	Parser     parser.Parser
	Normalizer normalizer.Normalizer
	Resolver   resolver.Resolver
	Builder    builder.Builder
	Validator  validator.Validator
	Scheduler  scheduler.Scheduler
	Generator  engine.GeneratorEngine
}

// Pipeline coordinates the main NAEOS processing flow.
type Pipeline struct {
	parser        parser.Parser
	normalizer    normalizer.Normalizer
	resolver      resolver.Resolver
	builder       builder.Builder
	validator     validator.Validator
	scheduler     scheduler.Scheduler
	generator     engine.GeneratorEngine
	outputDirValue string
}

// Result is the output produced by a pipeline run.
type Result struct {
	Source    string
	NEIR      *builder.NEIR
	Artifacts []engine.Artifact
	Tasks     []scheduler.Task
}

// ConfigFromFile loads pipeline configuration from a JSON or YAML file and returns a Config.
func ConfigFromFile(path string) (Config, error) {
	fileCfg, err := cfgpkg.LoadFile(path)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Name:      fileCfg.Pipeline.Name,
		Mode:      fileCfg.Pipeline.Mode,
		Verbose:   fileCfg.Pipeline.Verbose,
		OutputDir: fileCfg.Pipeline.OutputDir,
	}, nil
}

// New creates a default pipeline implementation with optional dependency injection.
func New(cfg Config) *Pipeline {
	p := &Pipeline{
		parser:         cfg.Parser,
		normalizer:     cfg.Normalizer,
		resolver:       cfg.Resolver,
		builder:        cfg.Builder,
		validator:      cfg.Validator,
		scheduler:      cfg.Scheduler,
		generator:      cfg.Generator,
		outputDirValue: cfg.OutputDir,
	}

	if p.parser == nil {
		p.parser = parser.NewParser()
	}
	if p.normalizer == nil {
		p.normalizer = normalizer.NewNormalizer()
	}
	if p.resolver == nil {
		p.resolver = resolver.NewResolver()
	}
	if p.builder == nil {
		p.builder = builder.NewBuilder()
	}
	if p.validator == nil {
		p.validator = validator.NewValidator()
	}
	if p.scheduler == nil {
		p.scheduler = scheduler.NewScheduler()
	}
	if p.generator == nil {
		p.generator = engine.NewEngine()
	}

	return p
}

// Run executes the specification-to-artifact pipeline.
func (p *Pipeline) outputDir() string {
	if p == nil {
		return ""
	}
	return p.outputDirValue
}

func (p *Pipeline) Run(input string) (*Result, error) {
	if input == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	parsed, err := p.parser.Parse(input)
	if err != nil {
		return nil, err
	}
	if parsed != nil {
		if parsed.Project == "" {
			parsed.Project = parser.DefaultProjectNameForInput(input)
		}
		if len(parsed.Modules) == 0 {
			parsed.Modules = []parser.Module{{Name: parser.DefaultModuleNameForProject(parsed.Project), Path: fmt.Sprintf("./%s", parser.Slugify(parsed.Project))}}
		}
	}

	normalized, err := p.normalizer.Normalize(parsed)
	if err != nil {
		return nil, err
	}

	resolved, err := p.resolver.Resolve(normalized)
	if err != nil {
		return nil, err
	}

	neir, err := p.builder.Build(resolved)
	if err != nil {
		return nil, err
	}

	if err := p.validator.Validate(neir); err != nil {
		return nil, err
	}

	tasks, err := p.scheduler.Schedule(neir)
	if err != nil {
		return nil, err
	}

	artifacts, err := p.generator.Generate(neir)
	if err != nil {
		return nil, err
	}

	outputDir := p.outputDir()
	if outputDir != "" {
		for _, artifact := range artifacts {
			artifactPath := filepath.Join(outputDir, artifact.Path)
			if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
				return nil, fmt.Errorf("create artifact dir: %w", err)
			}
			if err := os.WriteFile(artifactPath, artifact.Content, 0o644); err != nil {
				return nil, fmt.Errorf("write artifact %s: %w", artifact.Path, err)
			}
		}
	}

	return &Result{
		Source:    parsed.Raw,
		NEIR:      neir,
		Artifacts: artifacts,
		Tasks:     tasks,
	}, nil
}
