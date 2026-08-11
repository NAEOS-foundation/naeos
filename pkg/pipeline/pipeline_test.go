package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/review"
	"github.com/NAEOS-foundation/naeos/internal/profiling"
	"github.com/NAEOS-foundation/naeos/internal/schemaregistry"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func schemaForTesting() map[string]any {
	return map[string]any{
		"title":       "NEIR Spec",
		"description": "Test schema for NEIR spec validation",
		"type":        "object",
		"required":    []any{"project", "modules"},
		"properties": map[string]any{
			"project": map[string]any{
				"type": "string",
			},
			"modules": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":     "object",
					"required": []any{"name", "path"},
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"path": map[string]any{"type": "string"},
					},
				},
			},
			"services": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"kind": map[string]any{
							"type": "string",
							"enum": []any{"http", "grpc"},
						},
					},
				},
			},
		},
	}
}

type stubParser struct{}

func (stubParser) Parse(input string) (*parser.SpecDocument, error) {
	return &parser.SpecDocument{Raw: "injected:" + input}, nil
}

func TestPipelineRunProducesResult(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

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
	p, err := New(Config{Parser: stubParser{}})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}
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
	p, err := New(Config{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	_, err = p.Run("sample specification")
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

func TestPipelineRunWithLanguageOverride(t *testing.T) {
	p, err := New(Config{Languages: []string{"go"}})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	result, err := p.Run("project: test-proj\nmodules:\n  - name: core\n    path: ./internal/core\n")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.NEIR == nil {
		t.Fatal("expected NEIR")
	}
	if result.NEIR.Generation == nil {
		t.Fatal("expected Generation to be set")
	}
	if len(result.NEIR.Generation.Languages) != 1 {
		t.Fatalf("expected 1 language, got %d", len(result.NEIR.Generation.Languages))
	}
	if result.NEIR.Generation.Languages[0] != "go" {
		t.Fatalf("expected go, got %s", result.NEIR.Generation.Languages[0])
	}
}

func TestPipelineRunWithMultipleLanguages(t *testing.T) {
	p, err := New(Config{Languages: []string{"go", "typescript"}})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	result, err := p.Run("project: multi-proj\nmodules:\n  - name: core\n    path: ./internal/core\n")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.NEIR == nil || result.NEIR.Generation == nil {
		t.Fatal("expected NEIR with Generation")
	}
	if len(result.NEIR.Generation.Languages) != 2 {
		t.Fatalf("expected 2 languages, got %d", len(result.NEIR.Generation.Languages))
	}
	if len(result.Artifacts) == 0 {
		t.Fatal("expected artifacts from multi-language generation")
	}

	hasGoArtifact := false
	hasTSArtifact := false
	for _, a := range result.Artifacts {
		if strings.HasSuffix(a.Path, ".go") || a.Path == "go.mod" {
			hasGoArtifact = true
		}
		if strings.HasSuffix(a.Path, ".ts") || strings.HasSuffix(a.Path, ".tsx") || a.Path == "package.json" {
			hasTSArtifact = true
		}
	}
	if !hasGoArtifact {
		t.Fatal("expected at least one Go artifact")
	}
	if !hasTSArtifact {
		t.Fatal("expected at least one TypeScript artifact")
	}
}

func TestPipelineRunWithSpecFullExample(t *testing.T) {
	specData, err := os.ReadFile("../../examples/spec-full.yaml")
	if err != nil {
		t.Fatalf("read spec-full.yaml: %v", err)
	}

	p, err := New(Config{})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	result, err := p.Run(string(specData))
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.NEIR == nil {
		t.Fatal("expected NEIR")
	}
	if result.NEIR.Project == nil || result.NEIR.Project.Name != "e-commerce-platform" {
		t.Fatalf("expected project e-commerce-platform, got %v", result.NEIR.Project)
	}
	if len(result.NEIR.Modules) != 5 {
		t.Fatalf("expected 5 modules, got %d", len(result.NEIR.Modules))
	}
	if len(result.NEIR.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(result.NEIR.Services))
	}
	if result.NEIR.Generation == nil {
		t.Fatal("expected Generation from spec-full.yaml")
	}
	if len(result.NEIR.Generation.Languages) != 2 {
		t.Fatalf("expected 2 languages from spec-full.yaml, got %d", len(result.NEIR.Generation.Languages))
	}
	if len(result.Artifacts) == 0 {
		t.Fatal("expected artifacts")
	}

	goArtifacts := 0
	tsArtifacts := 0
	for _, a := range result.Artifacts {
		if strings.HasSuffix(a.Path, ".go") || a.Path == "go.mod" {
			goArtifacts++
		}
		if strings.HasSuffix(a.Path, ".ts") || strings.HasSuffix(a.Path, ".tsx") || a.Path == "package.json" {
			tsArtifacts++
		}
	}
	if goArtifacts == 0 {
		t.Fatal("expected Go artifacts from spec-full.yaml generation")
	}
	if tsArtifacts == 0 {
		t.Fatal("expected TypeScript artifacts from spec-full.yaml generation")
	}
}

func TestReviewRulesForArtifactGo(t *testing.T) {
	goRules := reviewRulesForArtifact("src/main.go")
	if len(goRules) != 4 {
		t.Fatalf("expected 4 review rules for Go artifact, got %d", len(goRules))
	}
	expected := []string{"no-todo", "no-placeholder", "has-package-declaration", "has-license-header"}
	for i, rule := range expected {
		if goRules[i] != rule {
			t.Fatalf("expected rule %q at index %d, got %q", rule, i, goRules[i])
		}
	}

	nonGoRules := reviewRulesForArtifact("README.md")
	if len(nonGoRules) != 2 {
		t.Fatalf("expected 2 review rules for non-Go artifact, got %d", len(nonGoRules))
	}
	if nonGoRules[0] != "no-todo" || nonGoRules[1] != "no-placeholder" {
		t.Fatalf("unexpected non-Go review rules: %#v", nonGoRules)
	}
}

func TestPipelineRunWithGoReviewArtifacts(t *testing.T) {
	p, err := New(Config{Languages: []string{"go"}})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	spec := `project: review-go
services:
  - name: api
    kind: http
    port: 8080
`

	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.Reviews) == 0 {
		t.Fatal("expected review results")
	}

	approved := 0
	for _, r := range result.Reviews {
		if r.Status == review.StatusApproved {
			approved++
		}
	}
	if approved == 0 {
		t.Fatal("expected at least one approved review")
	}
}

func TestPipelineVerboseMode(t *testing.T) {
	p, err := New(Config{Verbose: true})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	if !p.verbose {
		t.Fatal("expected verbose to be true")
	}

	result, err := p.Run("sample specification")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPipelineVerboseDisabled(t *testing.T) {
	p, err := New(Config{Verbose: false})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	if p.verbose {
		t.Fatal("expected verbose to be false")
	}
}

func TestPipelineRendererIntegration(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	renderer := p.Renderer()
	if renderer == nil {
		t.Fatal("expected renderer to be set")
	}
}

type mockCache struct {
	data map[string]*Result
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]*Result)}
}

func (c *mockCache) Get(hash string) (*Result, bool) {
	r, ok := c.data[hash]
	return r, ok
}

func (c *mockCache) Set(hash string, result *Result) {
	c.data[hash] = result
}

func (c *mockCache) HashSpec(spec string) string {
	return spec
}

func (c *mockCache) ModuleKey(specHash, moduleName, stage string) string {
	return specHash + ":" + moduleName + ":" + stage
}

func (c *mockCache) GetModuleStage(key string) ([]byte, bool) {
	return nil, false
}

func (c *mockCache) SetModuleStage(key string, data []byte) {
}

func (c *mockCache) UnchangedModules(specHash string, moduleHashes map[string]string) []string {
	return nil
}

func TestPipelineCacheHit(t *testing.T) {
	cache := newMockCache()
	p, err := New(Config{Cache: cache})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	result, err := p.Validate("project: cached")
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	_, exists := cache.data["project: cached"]
	if !exists {
		t.Fatal("expected result to be cached")
	}

	result2, err := p.Validate("project: cached")
	if err != nil {
		t.Fatalf("second Validate returned error: %v", err)
	}
	if result2 == nil {
		t.Fatal("expected non-nil result on cache hit")
	}
}

func TestPipelineCacheMiss(t *testing.T) {
	cache := newMockCache()
	p, err := New(Config{Cache: cache})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	_, err = p.Validate("project: alpha")
	if err != nil {
		t.Fatalf("Validate alpha returned error: %v", err)
	}

	_, err = p.Validate("project: beta")
	if err != nil {
		t.Fatalf("Validate beta returned error: %v", err)
	}

	if len(cache.data) != 2 {
		t.Fatalf("expected 2 cache entries, got %d", len(cache.data))
	}
}

func TestPipelineValidateCacheBypass(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	result, err := p.Validate("project: no-cache")
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

type artifactRecorder struct {
	artifacts []string
}

func (r *artifactRecorder) OnPipelineStart(pipelineID string)                        {}
func (r *artifactRecorder) OnPipelineComplete(pid string, artifacts int, dur string) {}
func (r *artifactRecorder) OnPipelineFailed(pid, errMsg string)                      {}
func (r *artifactRecorder) OnArtifactGenerated(name, path string) {
	r.artifacts = append(r.artifacts, name)
}

func TestPipelineObserverArtifactGenerated(t *testing.T) {
	rec := &artifactRecorder{}
	chain := ChainObservers(rec)
	chain.OnArtifactGenerated("main.go", "/out/main.go")
	if len(rec.artifacts) != 1 {
		t.Fatal("expected OnArtifactGenerated to fan-out")
	}
}

func TestPipelineObserverRunLifecycle(t *testing.T) {
	type lifecycle struct {
		started  bool
		complete bool
	}
	obs := &lifecycleRecorder{}
	p, err := New(Config{Observer: obs})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}
	_, err = p.Run("project: lifecycle\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !obs.started {
		t.Fatal("expected OnPipelineStart")
	}
	if !obs.complete {
		t.Fatal("expected OnPipelineComplete")
	}
}

func TestPipelineObserverArtifactGeneratedDuringRun(t *testing.T) {
	rec := &artifactRecorder{}
	p, err := New(Config{
		Observer:  rec,
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}
	_, err = p.Run("project: artifacts\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(rec.artifacts) == 0 {
		t.Fatal("expected OnArtifactGenerated to be called during Run")
	}
}

type lifecycleRecorder struct {
	started  bool
	complete bool
}

func (l *lifecycleRecorder) OnPipelineStart(pipelineID string)              { l.started = true }
func (l *lifecycleRecorder) OnPipelineComplete(pid string, a int, d string) { l.complete = true }
func (l *lifecycleRecorder) OnPipelineFailed(pid, errMsg string)            {}
func (l *lifecycleRecorder) OnArtifactGenerated(name, path string)          {}

func TestPipelineStageCache(t *testing.T) {
	sc := NewStageCache(t.TempDir())
	p, err := New(Config{StageCache: sc})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}
	result, err := p.Run("project: stage-cache-test\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("first Run returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result on first run")
	}
	stats := sc.Stats()
	if stats["schedule"] == 0 && stats["generate"] == 0 {
		t.Log("stage cache may not have been populated (no cache hits expected on first run)")
	}

	result2, err := p.Run("project: stage-cache-test\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("second Run returned error: %v", err)
	}
	if result2 == nil {
		t.Fatal("expected non-nil result on second run")
	}
	if len(result2.Artifacts) != len(result.Artifacts) {
		t.Errorf("expected same number of artifacts, got %d vs %d", len(result2.Artifacts), len(result.Artifacts))
	}
}

func TestPipelineWithProfileRecordsStages(t *testing.T) {
	prof := profiling.NewProfile()
	p, err := New(Config{Profile: prof})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	result, err := p.Run("project: profile-test\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	got := p.ProfileResult()
	if got == nil {
		t.Fatal("expected ProfileResult to return non-nil")
	}

	if got.TotalTime == 0 {
		t.Fatal("expected TotalTime to be set")
	}

	if len(got.Stages) == 0 {
		t.Fatal("expected at least one stage to be recorded")
	}

	stageNames := make(map[string]bool)
	for _, s := range got.Stages {
		stageNames[s.Name] = true
		if s.Duration == 0 {
			t.Fatalf("stage %q has zero duration", s.Name)
		}
	}

	expectedStages := []string{"validate", "build_graph", "policy_eval", "schedule", "generate", "review", "write_artifacts"}
	for _, name := range expectedStages {
		if !stageNames[name] {
			t.Fatalf("expected stage %q to be recorded", name)
		}
	}

	summary := got.Summary()
	if !strings.Contains(summary, "Pipeline Performance Profile") {
		t.Fatal("summary should contain header")
	}
}

func TestPipelineWithProfileAndSaveLoad(t *testing.T) {
	prof := profiling.NewProfile()
	p, err := New(Config{Profile: prof})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	_, err = p.Run("project: save-profile-test\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	profilePath := filepath.Join(t.TempDir(), "profile.json")
	if err := profiling.SaveProfile(profilePath, p.ProfileResult()); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	snap, err := profiling.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("LoadProfile failed: %v", err)
	}
	if snap.TotalTime == 0 {
		t.Fatal("expected non-zero total time in loaded profile")
	}
	if snap.StageCount == 0 {
		t.Fatal("expected non-zero stage count in loaded profile")
	}
}

func TestPipelineWithoutProfileReturnsNil(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	got := p.ProfileResult()
	if got != nil {
		t.Fatal("expected ProfileResult to return nil when no profile configured")
	}
}

func TestWithProfileOption(t *testing.T) {
	prof := profiling.NewProfile()
	cfg := Config{}
	WithProfile(prof)(&cfg)
	if cfg.Profile != prof {
		t.Fatal("WithProfile should set Profile on config")
	}
}

func TestPipelineProfileRecordsCorrectStageCount(t *testing.T) {
	prof := profiling.NewProfile()
	p, err := New(Config{Profile: prof})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	_, err = p.Run("project: stage-count\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(prof.Stages) != 7 {
		t.Fatalf("expected 7 recorded stages, got %d", len(prof.Stages))
	}
}

func TestValidateSpecWithSchemaValid(t *testing.T) {
	schema := schemaForTesting()
	spec := map[string]any{
		"project": "my-app",
		"modules": []any{
			map[string]any{"name": "core", "path": "./core"},
		},
	}

	result := schemaregistry.ValidateSpec(spec, schema)
	if !result.Valid {
		t.Fatalf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidateSpecWithSchemaMissingRequiredField(t *testing.T) {
	schema := schemaForTesting()
	spec := map[string]any{
		"project": "my-app",
	}

	result := schemaregistry.ValidateSpec(spec, schema)
	if result.Valid {
		t.Fatal("expected invalid spec due to missing 'modules'")
	}
	found := false
	for _, e := range result.Errors {
		if e.Field == "modules" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error about missing 'modules', got: %v", result.Errors)
	}
}

func TestValidateSpecWithSchemaInvalidEnum(t *testing.T) {
	schema := schemaForTesting()
	spec := map[string]any{
		"project": "my-app",
		"modules": []any{
			map[string]any{"name": "core", "path": "./core"},
		},
		"services": []any{
			map[string]any{"name": "api", "kind": "invalid-kind"},
		},
	}

	result := schemaregistry.ValidateSpec(spec, schema)
	if result.Valid {
		t.Fatal("expected invalid spec due to invalid enum value")
	}
}

func TestPipelineSchemaSourceValid(t *testing.T) {
	schemaFile := filepath.Join(t.TempDir(), "schema.json")
	schema := schemaForTesting()
	schemaData, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	if err := os.WriteFile(schemaFile, schemaData, 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	p, err := New(Config{SchemaSource: "file://" + schemaFile})
	if err != nil {
		t.Fatalf("create pipeline: %v", err)
	}

	result, err := p.Run("project: my-app\nmodules:\n  - name: core\n    path: ./core\n")
	if err != nil {
		t.Fatalf("Run with valid schema source failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPipelineSchemaSourceInvalidSpec(t *testing.T) {
	schemaFile := filepath.Join(t.TempDir(), "schema.json")
	schema := schemaForTesting()
	schemaData, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	if err := os.WriteFile(schemaFile, schemaData, 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	p, err := New(Config{SchemaSource: "file://" + schemaFile})
	if err != nil {
		t.Fatalf("create pipeline: %v", err)
	}

	_, err = p.Run("project: my-app\n")
	if err == nil {
		t.Fatal("expected error for spec missing required 'modules' field")
	}
	if !strings.Contains(err.Error(), "schema validation failed") {
		t.Fatalf("expected schema validation error, got: %v", err)
	}
}

func TestPipelineWithMemProfileRecordsSnapshots(t *testing.T) {
	mp := profiling.NewMemProfiler()
	p, err := New(Config{MemProfile: mp})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	_, err = p.Run("project: memprofile-test\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := p.MemProfileResult()
	if got == nil {
		t.Fatal("expected MemProfileResult to return non-nil")
	}

	snaps := got.Snapshots()
	if len(snaps) < 2 {
		t.Fatalf("expected at least 2 heap snapshots, got %d", len(snaps))
	}

	expectedLabels := []string{"pipeline_start", "validate", "build_graph", "policy_eval", "schedule", "generate", "review", "write_artifacts"}
	found := make(map[string]bool)
	for _, s := range snaps {
		found[s.Label] = true
	}
	for _, label := range expectedLabels {
		if !found[label] {
			t.Fatalf("expected snapshot with label %q", label)
		}
	}
}

func TestPipelineWithoutMemProfileReturnsNil(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	got := p.MemProfileResult()
	if got != nil {
		t.Fatal("expected MemProfileResult to return nil when no memprofile configured")
	}
}

func TestWithMemProfileOption(t *testing.T) {
	mp := profiling.NewMemProfiler()
	cfg := Config{}
	WithMemProfile(mp)(&cfg)
	if cfg.MemProfile != mp {
		t.Fatal("WithMemProfile should set MemProfile on config")
	}
}

func TestPipelineMemProfileSnapshotsCount(t *testing.T) {
	mp := profiling.NewMemProfiler()
	p, err := New(Config{MemProfile: mp})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	_, err = p.Run("project: memprofile-count\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(mp.Snapshots()) != 8 {
		t.Fatalf("expected 8 heap snapshots, got %d", len(mp.Snapshots()))
	}
}

func TestPipelineMemProfileDiffs(t *testing.T) {
	mp := profiling.NewMemProfiler()
	p, err := New(Config{MemProfile: mp})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	_, err = p.Run("project: memprofile-diffs\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	diffs := mp.Diffs()
	if len(diffs) == 0 {
		t.Fatal("expected at least one heap diff")
	}
}

func TestPipelineMemProfileAnalyze(t *testing.T) {
	mp := profiling.NewMemProfiler()
	p, err := New(Config{MemProfile: mp})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	_, err = p.Run("project: memprofile-analyze\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	report := mp.Analyze()
	if report.Diffs == nil {
		t.Fatal("expected non-nil diffs in analyze report")
	}
}

func TestPipelineSchemaSourceNotSet(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("create pipeline: %v", err)
	}

	result, err := p.Run("project: my-app\nmodules:\n  - name: core\n    path: ./core\n")
	if err != nil {
		t.Fatalf("Run without schema source should work: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
