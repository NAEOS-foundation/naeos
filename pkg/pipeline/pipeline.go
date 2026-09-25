// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
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
	"github.com/NAEOS-foundation/naeos/internal/profiling"
	"github.com/NAEOS-foundation/naeos/internal/registry"
	"github.com/NAEOS-foundation/naeos/internal/schemaregistry"
	"github.com/NAEOS-foundation/naeos/internal/securityext"
	naeoslog "github.com/NAEOS-foundation/naeos/internal/shared/log"
	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
	cfgpkg "github.com/NAEOS-foundation/naeos/pkg/config"
	"github.com/NAEOS-foundation/naeos/pkg/kernel"

	"gopkg.in/yaml.v3"
)

type ParseCache interface {
	Get(specHash string) (*Result, bool)
	Set(specHash string, result *Result)
	HashSpec(spec string) string
	ModuleKey(specHash, moduleName, stage string) string
	GetModuleStage(key string) ([]byte, bool)
	SetModuleStage(key string, data []byte)
	UnchangedModules(specHash string, moduleHashes map[string]string) []string
}

type Config struct {
	Name       string
	Mode       string
	Verbose    bool
	DryRun     bool
	OutputDir  string
	Languages  []string
	Parallel   *bool
	Profiling  bool
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
	// RequireGovernance makes an execution fail closed when no effective policy set is configured.
	// Policy-free execution remains available only when this explicit guard is disabled.
	RequireGovernance bool
	Hooks             *Hooks
	Observer          PipelineObserver
	Cache             ParseCache
	StageCache        *StageCache
	LazyBuild         bool
	Profile           *profiling.PipelineProfile
	MemProfile        *profiling.MemProfiler
	SchemaSource      string
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
	name              string
	parser            parser.Parser
	normalizer        normalizer.Normalizer
	resolver          resolver.Resolver
	builder           builder.Builder
	validator         validator.Validator
	scheduler         scheduler.Scheduler
	generator         engine.GeneratorEngine
	renderer          renderers.Renderer
	graph             *graph.PlannerGraph
	registry          *registry.Registry
	evaluator         policy.Evaluator
	reviewer          review.Reviewer
	kernel            *kernel.Kernel
	policies          []policy.Rule
	requireGovernance bool
	outputDirValue    string
	languages         []string
	verbose           bool
	dryRun            bool
	parallel          bool
	profiling         bool
	profile           *profiling.PipelineProfile
	hooks             *Hooks
	cache             ParseCache
	stageCache        *StageCache
	lazyBuild         bool
	observer          PipelineObserver
	neirResolved      any
	memProfile        *profiling.MemProfiler
	schemaSource      string
}

type PipelineObserver interface {
	OnPipelineStart(pipelineID string)
	OnPipelineComplete(pipelineID string, artifacts int, duration string)
	OnPipelineFailed(pipelineID string, errMsg string)
	OnArtifactGenerated(name string, path string)
}

type Result struct {
	RunID                string
	Source               string
	SpecificationHash    string
	NEIRHash             string
	NEIR                 *model.NEIR
	Artifacts            []engine.Artifact
	Tasks                []scheduler.Task
	Graph                *graph.PlannerGraph
	Reviews              []*review.ReviewResult
	PolicyResults        []policy.EvaluationResult
	GovernanceMode       string
	GovernanceStatus     string
	EffectivePolicyCount int
	DisabledPolicyRules  []string
	PolicyContextVersion string
	PolicyContextDigest  string
}

func WithCache(cache ParseCache) func(*Config) {
	return func(cfg *Config) {
		cfg.Cache = cache
	}
}

func WithStageCache(sc *StageCache) func(*Config) {
	return func(cfg *Config) {
		cfg.StageCache = sc
	}
}

func WithProfile(p *profiling.PipelineProfile) func(*Config) {
	return func(cfg *Config) {
		cfg.Profile = p
	}
}

func WithMemProfile(m *profiling.MemProfiler) func(*Config) {
	return func(cfg *Config) {
		cfg.MemProfile = m
	}
}

func ConfigFromFile(path string) (Config, error) {
	fileCfg, err := cfgpkg.LoadFile(path)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Name:              fileCfg.Pipeline.Name,
		Mode:              fileCfg.Pipeline.Mode,
		Verbose:           fileCfg.Pipeline.Verbose,
		OutputDir:         fileCfg.Pipeline.OutputDir,
		Languages:         fileCfg.Pipeline.Language,
		Policies:          fileCfg.Pipeline.Policies,
		RequireGovernance: strings.EqualFold(fileCfg.Pipeline.Mode, "governed"),
	}, nil
}

func New(cfg Config) (*Pipeline, error) { //nolint:gocritic // Public API, value semantics preferred
	parallel := true
	if cfg.Parallel != nil {
		parallel = *cfg.Parallel
	}
	profilingEnabled := cfg.Profiling
	var profile *profiling.PipelineProfile
	if profilingEnabled {
		profile = profiling.NewProfile()
	}
	p := &Pipeline{
		name:              cfg.Name,
		parser:            cfg.Parser,
		normalizer:        cfg.Normalizer,
		resolver:          cfg.Resolver,
		builder:           cfg.Builder,
		validator:         cfg.Validator,
		scheduler:         cfg.Scheduler,
		generator:         cfg.Generator,
		renderer:          cfg.Renderer,
		graph:             cfg.Graph,
		registry:          cfg.Registry,
		evaluator:         cfg.Evaluator,
		reviewer:          cfg.Reviewer,
		kernel:            cfg.Kernel,
		policies:          cfg.Policies,
		requireGovernance: cfg.RequireGovernance,
		outputDirValue:    cfg.OutputDir,
		languages:         cfg.Languages,
		verbose:           cfg.Verbose,
		dryRun:            cfg.DryRun,
		parallel:          parallel,
		profiling:         profilingEnabled,
		profile:           profile,
		hooks:             cfg.Hooks,
	}

	if p.parser == nil {
		p.parser = parser.NewParser(".")
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
	p.stageCache = cfg.StageCache
	p.lazyBuild = cfg.LazyBuild
	p.observer = cfg.Observer
	p.profile = cfg.Profile
	p.memProfile = cfg.MemProfile
	p.schemaSource = cfg.SchemaSource
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

func (p *Pipeline) ProfileResult() *profiling.PipelineProfile {
	return p.profile
}

func (p *Pipeline) MemProfileResult() *profiling.MemProfiler {
	return p.memProfile
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
			p.warn("failed to emit kernel.stop telemetry", "error", err)
		}
		if err := p.kernel.Stop(); err != nil {
			p.warn("failed to stop kernel", "error", err)
		}
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

func (p *Pipeline) Profile() *profiling.PipelineProfile {
	return p.profile
}

func (p *Pipeline) ProfileEnabled() bool {
	return p.profiling
}

func (p *Pipeline) ProfileSave(path string) error {
	if p.profile == nil {
		return fmt.Errorf("profiling not enabled")
	}
	return profiling.SaveProfile(path, p.profile)
}

func (p *Pipeline) warn(msg string, args ...any) {
	naeoslog.Warn(msg, args...)
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

	if p.schemaSource != "" && parsed != nil {
		if err := p.runSchemaValidate(input); err != nil {
			return nil, err
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

	if p.verbose {
		client := schemaregistry.NewNEIRClient(schemaregistry.DefaultNEIRSchemaURL)
		schema, fetchErr := client.FetchSchema()
		if fetchErr == nil && schema != nil {
			specPath := ""
			if parsed != nil && parsed.Raw != "" {
				tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("naeos-spec-%d.yaml", time.Now().UnixNano()))
				if writeErr := os.WriteFile(tmpFile, []byte(parsed.Raw), 0o600); writeErr == nil { //nolint:gosec // fixed TempDir+random suffix path
					specPath = tmpFile
					defer os.Remove(tmpFile)
				}
			}
			if specPath != "" {
				if vr, ve := schemaregistry.ValidateNEIRSpec(specPath, schema); ve == nil && !vr.Valid {
					naeoslog.Warn("schema validation: spec does not conform to NEIR JSON Schema")
					for _, e := range vr.Errors {
						naeoslog.Warn("  schema: %s: %s", e.Field, e.Message)
					}
				}
			}
		}
	}

	p.neirResolved = resolved
	neir.Modules = nil
	neir.Services = nil

	result := &Result{
		Source: parsed.Raw,
		NEIR:   neir,
	}
	if err := p.emitKernelEvent("pipeline.validate", map[string]any{"source_len": len(result.Source)}); err != nil {
		p.warn("failed to emit pipeline.validate event", "error", err)
	}

	if p.cache != nil {
		specHash := p.cache.HashSpec(input)
		p.cache.Set(specHash, result)
	}

	return result, nil
}

func (p *Pipeline) materializeNEIR(neir *model.NEIR) {
	if p.neirResolved == nil || len(neir.Modules) > 0 {
		return
	}
	loader := builder.NewModuleLoader()
	if modules, err := loader.LoadModules(p.neirResolved); err == nil && len(modules) > 0 {
		neir.Modules = modules
	}
	if services, err := loader.LoadServices(p.neirResolved); err == nil && len(services) > 0 {
		neir.Services = services
	}
}

func (p *Pipeline) normalizeParallel(parsed *parser.SpecDocument) (*normalizer.NormalizedSpec, error) {
	return p.normalizer.Normalize(parsed)
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
	pipelineID := fmt.Sprintf("pipe-%d", time.Now().UnixNano())
	durableLedgerPath, durableReceiptPath := durableRuntimePaths(p.outputDirValue, pipelineID)
	durableLedger, err := evidence.NewDurableRuntimeEventLedger(durableLedgerPath)
	if err != nil {
		return nil, fmt.Errorf("initialize durable runtime ledger: %w", err)
	}
	runEvidence := evidence.NewStore()
	runtimeObserver := evidence.NewIndependentRuntimeObserverWithDurableLedger(durableLedger)
	evidenceBuilder := evidence.NewRuntimeEvidenceBuilder(runEvidence, runtimeObserver)
	if err := appendRunEvidence(evidenceBuilder, runtimeObserver, pipelineID, "intent", 1, "run", "pipeline.start", evidence.ComputeArtifactHash([]byte(input))); err != nil {
		return nil, fmt.Errorf("record run intent evidence: %w", err)
	}
	if p.observer != nil {
		p.observer.OnPipelineStart(pipelineID)
	}

	if p.profile != nil {
		p.profile.Start()
	}

	p.memSnapshot("pipeline_start")

	startTime := time.Now()
	result, err := p.executeWithKernel(func() (*Result, error) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context canceled: %w", err)
		}

		p.profileStageStart("validate")
		result, err := p.runValidate(input)
		p.profileStageEnd("validate", err)
		p.memSnapshot("validate")
		if err != nil {
			return nil, err
		}
		p.materializeNEIR(result.NEIR)
		if p.profile != nil {
			p.profile.EndStage("validate")
		}

		p.profileStageStart("build_graph")
		execGraph := p.runBuildGraph(result)
		p.profileStageEnd("build_graph", nil)
		p.memSnapshot("build_graph")
		result.Graph = execGraph
		if p.profile != nil {
			p.profile.EndStage("build_graph")
		}

		p.profileStageStart("policy_eval")
		policyErr := p.runPolicyEval(result)
		if policyErr == nil {
			if err := appendRunEvidence(evidenceBuilder, runtimeObserver, pipelineID, "decision", 2, "policy_eval", "pipeline.policy_decision", result.PolicyContextDigest); err != nil {
				return nil, fmt.Errorf("record policy decision evidence: %w", err)
			}
		}
		p.profileStageEnd("policy_eval", policyErr)
		p.memSnapshot("policy_eval")
		if policyErr != nil {
			return nil, policyErr
		}

		p.profileStageStart("schedule")
		tasks, err := p.runSchedule(result)
		p.profileStageEnd("schedule", err)
		p.memSnapshot("schedule")
		if err != nil {
			return nil, err
		}
		if p.profile != nil {
			p.profile.EndStage("schedule")
		}

		p.profileStageStart("generate")
		var artifacts []engine.Artifact
		if p.cache != nil && len(result.Artifacts) > 0 {
			p.logVerbose("generation cache hit, reusing %d artifacts", len(result.Artifacts))
			artifacts = result.Artifacts
		} else {
			var genErr error
			artifacts, genErr = p.runGenerate(result)
			if genErr != nil {
				return nil, genErr
			}
			result.Artifacts = artifacts
			if p.cache != nil {
				p.cache.Set(p.cache.HashSpec(input), result)
			}
		}
		p.profileStageEnd("generate", nil)
		p.memSnapshot("generate")

		p.profileStageStart("review")
		reviews := p.runReview(artifacts)
		p.profileStageEnd("review", nil)
		p.memSnapshot("review")
		result.Reviews = reviews
		if p.profile != nil {
			p.profile.EndStage("review")
		}

		p.profileStageStart("write_artifacts")
		writeErr := p.runWriteArtifacts(artifacts)
		p.profileStageEnd("write_artifacts", writeErr)
		p.memSnapshot("write_artifacts")
		if writeErr != nil {
			return nil, writeErr
		}
		if err := appendRunEvidence(evidenceBuilder, runtimeObserver, pipelineID, "execution", 3, "write_artifacts", "pipeline.execution", artifactDigest(artifacts)); err != nil {
			return nil, fmt.Errorf("record execution evidence: %w", err)
		}

		result.Tasks = tasks
		result.Artifacts = artifacts
		if err := appendRunEvidence(evidenceBuilder, runtimeObserver, pipelineID, "observation", 4, "observation", "pipeline.observation", observationDigest(tasks, artifacts, reviews)); err != nil {
			return nil, fmt.Errorf("record observation evidence: %w", err)
		}
		if err := appendRunEvidence(evidenceBuilder, runtimeObserver, pipelineID, "verification", 5, "completion", "pipeline.verification", verificationDigest(tasks, artifacts, reviews)); err != nil {
			return nil, fmt.Errorf("record verification evidence: %w", err)
		}
		runtimeObserver.Seal()
		receipt, receiptErr := evidence.CreateDurableRuntimeReceipt(durableLedger, pipelineID, time.Now().UTC())
		if receiptErr != nil {
			return nil, fmt.Errorf("create durable runtime receipt: %w", receiptErr)
		}
		if receiptErr := evidence.WriteDurableRuntimeReceipt(durableReceiptPath, receipt); receiptErr != nil {
			return nil, fmt.Errorf("persist durable runtime receipt: %w", receiptErr)
		}
		loadedReceipt, receiptErr := evidence.LoadDurableRuntimeReceipt(durableReceiptPath)
		if receiptErr != nil {
			return nil, fmt.Errorf("reload durable runtime receipt: %w", receiptErr)
		}
		if err := validateRunCompletion(runEvidence, runtimeObserver.Ledger(), durableLedger, loadedReceipt, pipelineID); err != nil {
			return nil, err
		}
		p.logVerbose("pipeline complete: %d artifacts, %d tasks, %d reviews", len(artifacts), len(tasks), len(reviews))
		if err := p.emitKernelEvent("pipeline.run", map[string]any{
			"artifacts":   len(artifacts),
			"tasks":       len(tasks),
			"reviews":     len(reviews),
			"graph_nodes": execGraph.NodeCount(),
			"graph_edges": execGraph.EdgeCount(),
		}); err != nil {
			p.warn("failed to emit pipeline.run event", "error", err)
		}

		if err := p.executeHooks(p.getHookFuncs().AfterRun, "run"); err != nil {
			return nil, err
		}

		return result, nil
	})

	if p.profile != nil {
		p.profile.Finish()
	}

	if result != nil {
		result.RunID = pipelineID
		result.SpecificationHash = specHash(input)
		if result.NEIR != nil {
			result.NEIRHash = neirHash(result.NEIR)
		}
	}

	p.runNotify(pipelineID, startTime, result, err)
	return result, err
}

func (p *Pipeline) profileStageStart(name string) {
	if p.profile != nil {
		p.profile.StartStage(name)
	}
}

func (p *Pipeline) profileStageEnd(name string, _ error) {
	if p.profile != nil {
		p.profile.EndStage(name)
	}
}

func (p *Pipeline) memSnapshot(label string) {
	if p.memProfile != nil {
		p.memProfile.Snapshot(label)
	}
}

var requiredRunEvidenceKinds = []string{"intent", "decision", "execution", "observation", "verification"}

func appendRunEvidence(builder *evidence.RuntimeEvidenceBuilder, observer *evidence.IndependentRuntimeObserver, runID, kind string, sequence int, stage, event, payloadDigest string) error {
	if builder == nil || observer == nil {
		return fmt.Errorf("runtime evidence builder and producer are required")
	}
	runtimeEvent, err := observer.Observe(runID, event, payloadDigest, sequence)
	if err != nil {
		return err
	}
	return builder.Build(runID, kind, sequence, stage, event, runtimeEvent)
}

func artifactDigest(artifacts []engine.Artifact) string {
	data, _ := json.Marshal(artifacts)
	return evidence.ComputeArtifactHash(data)
}

func observationDigest(tasks []scheduler.Task, artifacts []engine.Artifact, reviews []*review.ReviewResult) string {
	data, _ := json.Marshal(map[string]int{"tasks": len(tasks), "artifacts": len(artifacts), "reviews": len(reviews)})
	return evidence.ComputeArtifactHash(data)
}

func verificationDigest(tasks []scheduler.Task, artifacts []engine.Artifact, reviews []*review.ReviewResult) string {
	data, _ := json.Marshal(map[string]any{
		"tasks": len(tasks), "artifacts": len(artifacts), "reviews": len(reviews), "verification": "completion-boundary",
	})
	return evidence.ComputeArtifactHash(data)
}

func durableRuntimePaths(outputDir, runID string) (string, string) {
	root := filepath.Join(outputDir, ".naeos", "runtime")
	return filepath.Join(root, runID+".jsonl"), filepath.Join(root, runID+".receipt.json")
}

func validateRunCompletion(store *evidence.EvidenceStore, runtimeLedger *evidence.RuntimeEventLedger, durableLedger *evidence.DurableRuntimeEventLedger, receipt evidence.DurableRuntimeReceipt, runID string) error {
	if durableLedger == nil {
		return fmt.Errorf("run completion blocked: durable runtime ledger is required")
	}
	if err := durableLedger.Verify(); err != nil {
		return fmt.Errorf("run completion blocked: durable runtime ledger verification failed: %w", err)
	}
	if err := evidence.VerifyDurableRuntimeReceipt(receipt, durableLedger, runID); err != nil {
		return fmt.Errorf("run completion blocked: durable runtime receipt verification failed: %w", err)
	}
	completion := evidence.ValidateCompletionWithRuntimeLedger(store, runtimeLedger, runID, requiredRunEvidenceKinds)
	if !completion.Complete {
		return fmt.Errorf("run completion blocked: incomplete evidence contract: %v", completion.Missing)
	}
	return nil
}

func (p *Pipeline) runValidate(input string) (*Result, error) {
	if err := p.executeHooks(p.getHookFuncs().BeforeRun, "run"); err != nil {
		return nil, err
	}
	return p.validateWithoutKernel(input)
}

func (p *Pipeline) runBuildGraph(result *Result) *graph.PlannerGraph {
	p.logVerbose("building execution graph")
	return p.buildExecutionGraph(result.NEIR)
}

func (p *Pipeline) runSchemaValidate(input string) error {
	schema, err := p.fetchSchema()
	if err != nil {
		return fmt.Errorf("schema fetch: %w", err)
	}

	var spec map[string]any
	if err := json.Unmarshal([]byte(input), &spec); err != nil {
		if err := yaml.Unmarshal([]byte(input), &spec); err != nil {
			return fmt.Errorf("schema validate: parse input: %w", err)
		}
	}

	result := schemaregistry.ValidateSpec(spec, schema)
	if !result.Valid {
		errs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			errs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
		}
		return fmt.Errorf("schema validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	p.logVerbose("spec conforms to schema %s", result.Version)
	return nil
}

func (p *Pipeline) fetchSchema() (map[string]any, error) {
	client := schemaregistry.NewNEIRClient(p.schemaSource)
	return client.FetchSchema()
}

func (p *Pipeline) runPolicyEval(result *Result) error {
	result.EffectivePolicyCount = 0
	result.DisabledPolicyRules = nil
	result.PolicyContextDigest = policyDecisionDigest("pending", 0, "", nil)
	for _, rule := range p.policies {
		if rule.Enabled {
			result.EffectivePolicyCount++
			continue
		}
		if rule.RuleID != "" {
			result.DisabledPolicyRules = append(result.DisabledPolicyRules, rule.RuleID)
		}
	}
	if len(result.DisabledPolicyRules) > 0 {
		if err := p.emitKernelEvent("governance.policy_disabled", map[string]any{
			"count":     len(result.DisabledPolicyRules),
			"rule_ids":  append([]string(nil), result.DisabledPolicyRules...),
			"status":    "observed",
			"run_scope": "policy_evaluation",
		}); err != nil {
			return fmt.Errorf("record disabled policy observation: %w", err)
		}
	}

	if p.requireGovernance {
		result.GovernanceMode = "governed"
		if result.EffectivePolicyCount == 0 {
			result.GovernanceStatus = "unconfigured"
			result.PolicyResults = nil
			result.PolicyContextDigest = policyDecisionDigest(result.GovernanceStatus, result.EffectivePolicyCount, result.GovernanceMode, nil)
			_ = p.emitKernelEvent("governance.unconfigured", map[string]any{
				"mode": "governed", "effective_policy_count": 0, "status": "blocked",
			})
			return fmt.Errorf("governance configuration required: no effective policies configured")
		}
	} else if result.EffectivePolicyCount == 0 {
		result.GovernanceMode = "ungoverned"
		result.GovernanceStatus = "intentionally-disabled"
		result.PolicyResults = nil
		result.PolicyContextDigest = policyDecisionDigest(result.GovernanceStatus, result.EffectivePolicyCount, result.GovernanceMode, nil)
		return nil
	} else {
		result.GovernanceMode = "governed"
		result.GovernanceStatus = "evaluating"
	}
	p.logVerbose("evaluating %d policy rules", result.EffectivePolicyCount)
	policyContext, err := policy.ContextFromNEIR(result.NEIR)
	if err != nil {
		result.GovernanceStatus = "invalid-context"
		return fmt.Errorf("policy context construction failed: %w", err)
	}
	if err := policyContext.Validate(); err != nil {
		result.GovernanceStatus = "invalid-context"
		return fmt.Errorf("policy context validation failed: %w", err)
	}
	contextDigest, err := policyContext.Digest()
	if err != nil {
		result.GovernanceStatus = "invalid-context"
		return fmt.Errorf("policy context digest failed: %w", err)
	}
	result.PolicyContextVersion = policyContext.Version
	result.PolicyContextDigest = contextDigest
	p.logVerbose("evaluating %d policy rules against policy context %s digest %s", result.EffectivePolicyCount, policyContext.Version, contextDigest)
	contextEvaluator, ok := p.evaluator.(policy.ContextEvaluator)
	if !ok {
		result.GovernanceStatus = "invalid-context"
		return fmt.Errorf("policy evaluator does not support versioned policy context")
	}
	results, err := contextEvaluator.EvaluateRulesContext(policyContext, p.policies)
	if err != nil {
		return fmt.Errorf("policy evaluation failed: %w", err)
	}
	result.PolicyResults = results
	result.GovernanceStatus = "evaluated"
	_ = p.emitKernelEvent("governance.policy_context", map[string]any{
		"version":                result.PolicyContextVersion,
		"digest":                 result.PolicyContextDigest,
		"effective_policy_count": result.EffectivePolicyCount,
	})
	for _, res := range results {
		if !res.Passed {
			return fmt.Errorf("policy evaluation failed: rule %s: %s", res.RuleID, res.Message)
		}
	}
	return nil
}

type taskList struct {
	Tasks []scheduler.Task `json:"tasks"`
}

type artifactList struct {
	Artifacts []engine.Artifact `json:"artifacts"`
}

func policyDecisionDigest(status string, effectivePolicyCount int, mode string, results []policy.EvaluationResult) string {
	data, _ := json.Marshal(map[string]any{
		"status":                 status,
		"effective_policy_count": effectivePolicyCount,
		"mode":                   mode,
		"results":                results,
	})
	return evidence.ComputeArtifactHash(data)
}

func specHash(input string) string {
	h := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", h[:8])
}

func neirHash(neir *model.NEIR) string {
	data, _ := json.Marshal(neir)
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}

func (p *Pipeline) runSchedule(result *Result) ([]scheduler.Task, error) {
	p.logVerbose("scheduling %d tasks", len(result.NEIR.Modules)+len(result.NEIR.Services)+2)
	if p.stageCache == nil || result.NEIR == nil {
		return p.scheduler.Schedule(result.NEIR)
	}
	key := neirHash(result.NEIR)
	if cached, ok := p.stageCache.Get("schedule", []byte(key)); ok {
		p.logVerbose("stage cache hit for schedule")
		var tasks taskList
		if err := json.Unmarshal(cached, &tasks); err == nil {
			return tasks.Tasks, nil
		}
	}
	tasks, err := p.scheduler.Schedule(result.NEIR)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(taskList{Tasks: tasks})
	if err != nil {
		return tasks, nil
	}
	p.stageCache.Set("schedule", []byte(key), data)
	return tasks, nil
}

func (p *Pipeline) runGenerate(result *Result) ([]engine.Artifact, error) {
	if err := p.executeHooks(p.getHookFuncs().BeforeGenerate, "generate"); err != nil {
		return nil, err
	}

	p.logVerbose("generating artifacts")

	if p.stageCache != nil && result.NEIR != nil {
		key := neirHash(result.NEIR)
		if cached, ok := p.stageCache.Get("generate", []byte(key)); ok {
			p.logVerbose("stage cache hit for generate")
			var arts artifactList
			if err := json.Unmarshal(cached, &arts); err == nil {
				if err := p.executeHooks(p.getHookFuncs().AfterGenerate, "generate"); err != nil {
					return nil, err
				}
				return arts.Artifacts, nil
			}
		}
	}

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

	if p.stageCache != nil && result.NEIR != nil {
		key := neirHash(result.NEIR)
		data, err := json.Marshal(artifactList{Artifacts: artifacts})
		if err == nil {
			p.stageCache.Set("generate", []byte(key), data)
		}
	}

	return artifacts, nil
}

func (p *Pipeline) runReview(artifacts []engine.Artifact) []*review.ReviewResult {
	p.logVerbose("reviewing %d artifacts", len(artifacts))
	var reviews []*review.ReviewResult
	for _, artifact := range artifacts {
		rules := reviewRulesForArtifact(artifact.Path)
		r, err := p.reviewer.ReviewArtifact(artifact.Path, string(artifact.Content), rules)
		if err == nil && r != nil {
			reviews = append(reviews, r)
		}
	}
	return reviews
}

func (p *Pipeline) runWriteArtifacts(artifacts []engine.Artifact) error {
	outputDir := p.outputDirValue
	if outputDir == "" || p.dryRun {
		if p.dryRun {
			p.logVerbose("dry-run: skipping write of %d artifacts", len(artifacts))
		}
		return nil
	}
	p.logVerbose("writing %d artifacts to %s", len(artifacts), outputDir)
	for _, artifact := range artifacts {
		if _, err := securityext.ValidateFilePath(filepath.Join(outputDir, artifact.Path), outputDir); err != nil {
			return fmt.Errorf("invalid artifact path: %w", err)
		}
		artifactPath := filepath.Join(outputDir, artifact.Path)
		if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
			return fmt.Errorf("create artifact dir: %w", err)
		}
		if err := os.WriteFile(artifactPath, artifact.Content, 0o600); err != nil {
			return fmt.Errorf("write artifact %s: %w", artifact.Path, err)
		}
		if p.observer != nil {
			p.observer.OnArtifactGenerated(artifact.Path, artifactPath)
		}
	}
	return nil
}

func (p *Pipeline) runNotify(pipelineID string, startTime time.Time, result *Result, err error) {
	if p.observer == nil {
		return
	}
	if err != nil {
		p.observer.OnPipelineFailed(pipelineID, err.Error())
		return
	}
	duration := time.Since(startTime).Round(time.Millisecond).String()
	artifactCount := 0
	if result != nil {
		artifactCount = len(result.Artifacts)
	}
	p.observer.OnPipelineComplete(pipelineID, artifactCount, duration)
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
