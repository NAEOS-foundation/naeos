package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/generation/adapters"
	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/generation/renderers"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
	"github.com/NAEOS-foundation/naeos/internal/governance/review"
	"github.com/NAEOS-foundation/naeos/internal/neir/builder"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/validator"
	"github.com/NAEOS-foundation/naeos/internal/planner/graph"
	"github.com/NAEOS-foundation/naeos/internal/planner/scheduler"
	"github.com/NAEOS-foundation/naeos/internal/registry"
	naeoslog "github.com/NAEOS-foundation/naeos/internal/shared/log"
	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
	cfgpkg "github.com/NAEOS-foundation/naeos/pkg/config"
	"github.com/NAEOS-foundation/naeos/pkg/kernel"
)

type ParseCache interface {
	Get(specHash string) (*Result, bool)
	Set(specHash string, result *Result)
	HashSpec(spec string) string
}

type Config struct {
	Name       string
	Mode       string
	Verbose    bool
	DryRun     bool
	OutputDir  string
	Languages  []string
	Parallel   *bool
	Parser     parser.Parser
	Normalizer normalizer.Normalizer
	Resolver   resolver.Resolver
	Builder    builder.Builder
	Validator  validator.Validator
	Scheduler  scheduler.Scheduler
	Generator  engine.GeneratorEngine
	Renderer   renderers.Renderer
	Graph      *graph.PlannerGraph
	Registry   *registry.Registry
	Evaluator  policy.Evaluator
	Reviewer   review.Reviewer
	Kernel     *kernel.Kernel
	Policies   []policy.Rule
	Hooks      *Hooks
	Observer   PipelineObserver
	Cache      ParseCache
}

type HookFunc func(ctx *HookContext) error

type HookContext struct {
	Pipeline *Pipeline
	Stage    string
	Data     map[string]any
}

type Hooks struct {
	BeforeParse    []HookFunc
	AfterParse     []HookFunc
	BeforeRun      []HookFunc
	AfterRun       []HookFunc
	BeforeGenerate []HookFunc
	AfterGenerate  []HookFunc
}

type Pipeline struct {
	name           string
	parser         parser.Parser
	normalizer     normalizer.Normalizer
	resolver       resolver.Resolver
	builder        builder.Builder
	validator      validator.Validator
	scheduler      scheduler.Scheduler
	generator      engine.GeneratorEngine
	renderer       renderers.Renderer
	graph          *graph.PlannerGraph
	registry       *registry.Registry
	evaluator      policy.Evaluator
	reviewer       review.Reviewer
	kernel         *kernel.Kernel
	policies       []policy.Rule
	outputDirValue string
	languages      []string
	verbose        bool
	dryRun         bool
	parallel       bool
	hooks          *Hooks
	cache          ParseCache
}

type PipelineObserver interface {
	OnPipelineStart(pipelineID string)
	OnPipelineComplete(pipelineID string, artifacts int, duration string)
	OnPipelineFailed(pipelineID string, errMsg string)
	OnArtifactGenerated(name string, path string)
}

type Result struct {
	Source    string
	NEIR      *model.NEIR
	Artifacts []engine.Artifact
	Tasks     []scheduler.Task
	Graph     *graph.PlannerGraph
	Reviews   []*review.ReviewResult
}

func WithCache(cache ParseCache) func(*Config) {
	return func(cfg *Config) {
		cfg.Cache = cache
	}
}

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
		Languages: fileCfg.Pipeline.Language,
	}, nil
}

func New(cfg Config) (*Pipeline, error) { //nolint:gocritic // Public API, value semantics preferred
	parallel := true
	if cfg.Parallel != nil {
		parallel = *cfg.Parallel
	}
	p := &Pipeline{
		name:           cfg.Name,
		parser:         cfg.Parser,
		normalizer:     cfg.Normalizer,
		resolver:       cfg.Resolver,
		builder:        cfg.Builder,
		validator:      cfg.Validator,
		scheduler:      cfg.Scheduler,
		generator:      cfg.Generator,
		renderer:       cfg.Renderer,
		graph:          cfg.Graph,
		registry:       cfg.Registry,
		evaluator:      cfg.Evaluator,
		reviewer:       cfg.Reviewer,
		kernel:         cfg.Kernel,
		policies:       cfg.Policies,
		outputDirValue: cfg.OutputDir,
		languages:      cfg.Languages,
		verbose:        cfg.Verbose,
		dryRun:         cfg.DryRun,
		parallel:       parallel,
		hooks:          cfg.Hooks,
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
	if p.renderer == nil {
		p.renderer = renderers.NewRenderer()
	}
	if p.graph == nil {
		p.graph = graph.New()
	}
	if p.registry == nil {
		p.registry = registry.NewRegistry()
	}
	if p.evaluator == nil {
		p.evaluator = policy.NewEvaluator()
	}
	if p.reviewer == nil {
		p.reviewer = review.NewReviewer()
	}
	if p.kernel == nil {
		p.kernel = kernel.NewKernel()
	}
	p.cache = cfg.Cache
	if err := p.registerKernelServices(); err != nil {
		return nil, err
	}

	return p, nil
}

func reviewRulesForArtifact(path string) []string {
	if strings.HasSuffix(path, ".go") {
		return []string{"no-todo", "no-placeholder", "has-package-declaration", "has-license-header"}
	}
	return []string{"no-todo", "no-placeholder"}
}

func (p *Pipeline) Name() string {
	if p.name != "" {
		return p.name
	}
	return "unnamed"
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
		"renderer":   p.renderer,
		"graph":      p.graph,
		"registry":   p.registry,
		"evaluator":  p.evaluator,
		"reviewer":   p.reviewer,
		"pipeline":   p,
	}

	for name, service := range services {
		if err := p.kernel.Register(name, service); err != nil {
			return err
		}
	}
	return nil
}

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

func (p *Pipeline) logVerbose(format string, args ...any) {
	if p.verbose {
		naeoslog.Info(fmt.Sprintf(format, args...))
	}
}

func (p *Pipeline) executeHooks(hookFuncs []HookFunc, stage string) error {
	if p.hooks == nil || len(hookFuncs) == 0 {
		return nil
	}
	ctx := &HookContext{
		Pipeline: p,
		Stage:    stage,
		Data:     make(map[string]any),
	}
	for _, hook := range hookFuncs {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("hook %s failed: %w", stage, err)
		}
	}
	return nil
}

func (p *Pipeline) Hooks() *Hooks {
	if p.hooks == nil {
		return &Hooks{}
	}
	return p.hooks
}

func (p *Pipeline) buildExecutionGraph(neir *model.NEIR) *graph.PlannerGraph {
	g := graph.New()

	if neir.Project != nil {
		_ = g.AddNode(graph.Node{
			ID:   "project",
			Kind: graph.NodeKindModule,
			Name: neir.Project.Name,
		})
	}

	for i, mod := range neir.Modules {
		nodeID := fmt.Sprintf("module-%s", mod.Name)
		_ = g.AddNode(graph.Node{
			ID:   nodeID,
			Kind: graph.NodeKindModule,
			Name: mod.Name,
		})
		if i > 0 {
			prevID := fmt.Sprintf("module-%s", neir.Modules[i-1].Name)
			_ = g.AddEdge(graph.Edge{
				From: prevID,
				To:   nodeID,
				Kind: graph.EdgeKindDependency,
			})
		} else {
			_ = g.AddEdge(graph.Edge{
				From: "project",
				To:   nodeID,
				Kind: graph.EdgeKindDependency,
			})
		}
	}

	for _, svc := range neir.Services {
		nodeID := fmt.Sprintf("service-%s", svc.Name)
		_ = g.AddNode(graph.Node{
			ID:   nodeID,
			Kind: graph.NodeKindService,
			Name: svc.Name,
		})
		if len(neir.Modules) > 0 {
			lastModule := fmt.Sprintf("module-%s", neir.Modules[len(neir.Modules)-1].Name)
			_ = g.AddEdge(graph.Edge{
				From: lastModule,
				To:   nodeID,
				Kind: graph.EdgeKindDependency,
			})
		}
	}

	return g
}

func (p *Pipeline) validateWithoutKernel(input string) (*Result, error) {
	if input == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	if p.cache != nil {
		specHash := p.cache.HashSpec(input)
		if cached, ok := p.cache.Get(specHash); ok {
			p.logVerbose("cache hit for specification hash %s", specHash)
			return cached, nil
		}
	}

	if err := p.executeHooks(p.getHookFuncs().BeforeParse, "parse"); err != nil {
		return nil, err
	}

	p.logVerbose("parsing specification (%d bytes)", len(input))
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

	if err := p.executeHooks(p.getHookFuncs().AfterParse, "parse"); err != nil {
		return nil, err
	}

	p.logVerbose("normalizing specification")
	var normalized *normalizer.NormalizedSpec
	if p.parallel && parsed != nil && len(parsed.Modules) > 1 {
		normalized, err = p.normalizeParallel(parsed)
	} else {
		normalized, err = p.normalizer.Normalize(parsed)
	}
	if err != nil {
		return nil, err
	}

	p.logVerbose("resolving cross-references")
	resolved, err := p.resolver.Resolve(normalized)
	if err != nil {
		return nil, err
	}

	p.logVerbose("building NEIR model")
	neir, err := p.builder.Build(resolved)
	if err != nil {
		return nil, err
	}

	if len(p.languages) > 0 {
		if neir.Generation == nil {
			neir.Generation = &generation.GenerationConfig{}
		}
		neir.Generation.Languages = make([]language.Language, 0, len(p.languages))
		for _, l := range p.languages {
			neir.Generation.Languages = append(neir.Generation.Languages, language.Language(l))
		}
	}

	if err := p.validator.Validate(neir); err != nil {
		return nil, err
	}

	result := &Result{
		Source: parsed.Raw,
		NEIR:   neir,
	}
	_ = p.emitKernelEvent("pipeline.validate", map[string]any{"source_len": len(result.Source)})

	if p.cache != nil {
		specHash := p.cache.HashSpec(input)
		p.cache.Set(specHash, result)
	}

	return result, nil
}

func (p *Pipeline) normalizeParallel(parsed *parser.SpecDocument) (*normalizer.NormalizedSpec, error) {
	return p.normalizer.Normalize(parsed)
}

func normalizeServicesFromSpec(parsed *parser.SpecDocument) []map[string]any {
	result := make([]map[string]any, 0, len(parsed.Services))
	for _, s := range parsed.Services {
		entry := map[string]any{
			"name": s.Name,
			"kind": s.Kind,
			"port": s.Port,
		}
		if s.Description != "" {
			entry["description"] = s.Description
		}
		if len(s.Endpoints) > 0 {
			eps := make([]map[string]any, 0, len(s.Endpoints))
			for _, ep := range s.Endpoints {
				eps = append(eps, map[string]any{
					"method": ep.Method,
					"path":   ep.Path,
					"action": ep.Action,
				})
			}
			entry["endpoints"] = eps
		}
		result = append(result, entry)
	}
	return result
}

func normalizeArchitectureFromSpec(arch *parser.Architecture) map[string]any {
	result := map[string]any{
		"pattern":     arch.Pattern,
		"description": arch.Description,
	}
	if len(arch.Principles) > 0 {
		result["principles"] = arch.Principles
	}
	return result
}

func normalizeDeploymentFromSpec(deploy *parser.Deployment) map[string]any {
	result := map[string]any{
		"strategy": deploy.Strategy,
	}
	if len(deploy.Environments) > 0 {
		envs := make([]map[string]any, 0, len(deploy.Environments))
		for _, env := range deploy.Environments {
			envs = append(envs, map[string]any{"name": env})
		}
		result["environments"] = envs
	}
	return result
}

func normalizeGenerationFromSpec(gen *parser.Generation) map[string]any {
	result := map[string]any{}
	if len(gen.Languages) > 0 {
		result["languages"] = gen.Languages
	}
	if gen.OutputDir != "" {
		result["output_dir"] = gen.OutputDir
	}
	if gen.ModuleDir != "" {
		result["module_dir"] = gen.ModuleDir
	}
	return result
}

func normalizeTestingFromSpec(test *parser.Testing) map[string]any {
	result := map[string]any{
		"strategy": test.Strategy,
	}
	if test.Coverage != "" {
		result["coverage"] = test.Coverage
	}
	return result
}

func (p *Pipeline) Validate(input string) (*Result, error) {
	return p.ValidateContext(context.Background(), input)
}

func (p *Pipeline) ValidateContext(ctx context.Context, input string) (*Result, error) {
	return p.executeWithKernel(func() (*Result, error) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context canceled: %w", err)
		}
		return p.validateWithoutKernel(input)
	})
}

func (p *Pipeline) Run(input string) (*Result, error) {
	return p.RunContext(context.Background(), input)
}

func (p *Pipeline) RunContext(ctx context.Context, input string) (*Result, error) {
	return p.executeWithKernel(func() (*Result, error) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context canceled: %w", err)
		}

		if err := p.executeHooks(p.getHookFuncs().BeforeRun, "run"); err != nil {
			return nil, err
		}

		result, err := p.validateWithoutKernel(input)
		if err != nil {
			return nil, err
		}

		p.logVerbose("building execution graph")
		execGraph := p.buildExecutionGraph(result.NEIR)
		result.Graph = execGraph

		if len(p.policies) > 0 {
			p.logVerbose("evaluating %d policy rules", len(p.policies))
			ctx := map[string]any{
				"project":  result.NEIR.Project.Name,
				"modules":  len(result.NEIR.Modules),
				"services": len(result.NEIR.Services),
			}
			if _, err := p.evaluator.EvaluateRules(p.policies, ctx); err != nil {
				return nil, fmt.Errorf("policy evaluation failed: %w", err)
			}
		}

		p.logVerbose("scheduling %d tasks", len(result.NEIR.Modules)+len(result.NEIR.Services)+2)
		tasks, err := p.scheduler.Schedule(result.NEIR)
		if err != nil {
			return nil, err
		}

		if err := p.executeHooks(p.getHookFuncs().BeforeGenerate, "generate"); err != nil {
			return nil, err
		}

		p.logVerbose("generating artifacts")
		artifacts, err := p.generator.Generate(result.NEIR)
		if err != nil {
			return nil, err
		}

		p.logVerbose("running language adapters")
		adapterArtifacts, err := adapters.GenerateForNEIR(result.NEIR)
		if err != nil {
			return nil, fmt.Errorf("adapter generation failed: %w", err)
		}
		artifacts = append(artifacts, adapterArtifacts...)

		if err := p.executeHooks(p.getHookFuncs().AfterGenerate, "generate"); err != nil {
			return nil, err
		}

		p.logVerbose("reviewing %d artifacts", len(artifacts))
		var reviews []*review.ReviewResult
		for _, artifact := range artifacts {
			rules := reviewRulesForArtifact(artifact.Path)
			r, err := p.reviewer.ReviewArtifact(artifact.Path, string(artifact.Content), rules)
			if err == nil && r != nil {
				reviews = append(reviews, r)
			}
		}
		result.Reviews = reviews

		if outputDir := p.outputDirValue; outputDir != "" && !p.dryRun {
			p.logVerbose("writing %d artifacts to %s", len(artifacts), outputDir)
			for _, artifact := range artifacts {
				artifactPath := filepath.Join(outputDir, artifact.Path)
				if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
					return nil, fmt.Errorf("create artifact dir: %w", err)
				}
				if err := os.WriteFile(artifactPath, artifact.Content, 0o600); err != nil {
					return nil, fmt.Errorf("write artifact %s: %w", artifact.Path, err)
				}
			}
		} else if p.dryRun {
			p.logVerbose("dry-run: skipping write of %d artifacts", len(artifacts))
		}

		result.Tasks = tasks
		result.Artifacts = artifacts
		p.logVerbose("pipeline complete: %d artifacts, %d tasks, %d reviews", len(artifacts), len(tasks), len(reviews))
		_ = p.emitKernelEvent("pipeline.run", map[string]any{
			"artifacts":   len(artifacts),
			"tasks":       len(tasks),
			"reviews":     len(reviews),
			"graph_nodes": execGraph.NodeCount(),
			"graph_edges": execGraph.EdgeCount(),
		})

		if err := p.executeHooks(p.getHookFuncs().AfterRun, "run"); err != nil {
			return nil, err
		}

		return result, nil
	})
}

func (p *Pipeline) getHookFuncs() *Hooks {
	if p.hooks == nil {
		return &Hooks{}
	}
	return p.hooks
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
}

func (p *Pipeline) Registry() *registry.Registry {
	return p.registry
}

func (p *Pipeline) Graph() *graph.PlannerGraph {
	return p.graph
}

func (p *Pipeline) Renderer() renderers.Renderer {
	return p.renderer
}
