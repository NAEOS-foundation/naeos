package pipeline

import (
	"fmt"
	"time"
	"os"
	"path/filepath"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/builder"
	"github.com/NAEOS-foundation/naeos/internal/neir/validator"
	"github.com/NAEOS-foundation/naeos/internal/planner/scheduler"
	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
	cfgpkg "github.com/NAEOS-foundation/naeos/pkg/config"
	"github.com/NAEOS-foundation/naeos/pkg/kernel"
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
	Kernel     *kernel.Kernel
}

// Pipeline coordinates the main NAEOS processing flow.
type Pipeline struct {
	parser     parser.Parser
	normalizer normalizer.Normalizer
	resolver   resolver.Resolver
	builder    builder.Builder
	validator  validator.Validator
	scheduler  scheduler.Scheduler
	generator  engine.GeneratorEngine
	kernel     *kernel.Kernel
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
func New(cfg Config) (*Pipeline, error) {
	p := &Pipeline{
		parser:     cfg.Parser,
		normalizer: cfg.Normalizer,
		resolver:   cfg.Resolver,
		builder:    cfg.Builder,
		validator:  cfg.Validator,
		scheduler:  cfg.Scheduler,
		generator:  cfg.Generator,
		kernel:     cfg.Kernel,
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
	if p.kernel == nil {
		p.kernel = kernel.NewKernel()
	}
	if err := p.registerKernelServices(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Pipeline) registerKernelServices() error {
	services := map[string]any{
		"parser":     p.parser,
		"normalizer": p.normalizer,
		"resolver":   p.resolver,
		"builder":    p.builder,
		"validator":  p.validator,
		"scheduler":  p.scheduler,
		"generator":  p.generator,
		"pipeline":   p,
	}

	for name, service := range services {
		if err := p.kernel.Register(name, service); err != nil {
			return err
		}
	}
	return nil
}

// Run executes the specification-to-artifact pipeline.
func (p *Pipeline) executeWithKernel(fn func() (*Result, error)) (*Result, error) {
	if err := p.kernel.Start(); err != nil {
		return nil, err
	}
	if err := p.emitKernelEvent("kernel.start", map[string]any{"services": p.kernel.RegisteredServices()}); err != nil {
		return nil, err
	}
	defer func() {
		if err := p.kernel.EmitTelemetry(kernel.TelemetryEvent{
			Name:      "kernel.stop",
			Timestamp: time.Now().UnixMilli(),
			Payload:   map[string]any{"services": p.kernel.RegisteredServices()},
		}); err != nil {
			_ = err
		}
		_ = p.kernel.Stop()
	}()

	return fn()
}

func (p *Pipeline) emitKernelEvent(name string, payload map[string]any) error {
	if p.kernel == nil {
		return nil
	}
	return p.kernel.EmitTelemetry(kernel.TelemetryEvent{
		Name:      name,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	})
}

func (p *Pipeline) validateWithoutKernel(input string) (*Result, error) {
=======
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

	result := &Result{
		Source: parsed.Raw,
		NEIR:   neir,
	}
	_ = p.emitKernelEvent("pipeline.validate", map[string]any{"source_len": len(result.Source)})
	return result, nil
}

func (p *Pipeline) Validate(input string) (*Result, error) {
	return p.executeWithKernel(func() (*Result, error) {
		return p.validateWithoutKernel(input)
	})
}

func (p *Pipeline) Run(input string) (*Result, error) {
	return p.executeWithKernel(func() (*Result, error) {
		result, err := p.validateWithoutKernel(input)
		if err != nil {
			return nil, err
		}

		tasks, err := p.scheduler.Schedule(result.NEIR)
		if err != nil {
			return nil, err
		}

		artifacts, err := p.generator.Generate(result.NEIR)
		if err != nil {
			return nil, err
		}

		result.Tasks = tasks
		result.Artifacts = artifacts
		_ = p.emitKernelEvent("pipeline.run", map[string]any{"artifacts": len(artifacts), "tasks": len(tasks)})
		return result, nil
	})
}

func (p *Pipeline) RegisteredKernelServices() []string {
	if p.kernel == nil {
		return nil
	}
	return p.kernel.RegisteredServices()
}

func (p *Pipeline) KernelMetrics() kernel.Metrics {
	if p.kernel == nil {
		return kernel.Metrics{}
	}
	return p.kernel.Metrics()
}

func (p *Pipeline) KernelTopics() []string {
	if p.kernel == nil {
		return nil
	}
	return p.kernel.Topics()
}

func (p *Pipeline) Publish(topic string, payload any) error {
	if p.kernel == nil {
		return fmt.Errorf("kernel not initialized")
	}
	p.kernel.Publish(topic, payload)
	return nil
}

func (p *Pipeline) Subscribe(topic string, handler func(any)) error {
	if p.kernel == nil {
		return fmt.Errorf("kernel not initialized")
	}
	return p.kernel.Subscribe(topic, handler)
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
	}, nill
}
