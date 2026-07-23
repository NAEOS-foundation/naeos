package pipeline

import (
	"context"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

type stubCache struct {
	store map[string]*Result
}

func (c *stubCache) Get(hash string) (*Result, bool) {
	r, ok := c.store[hash]
	return r, ok
}

func (c *stubCache) Set(hash string, result *Result) {
	c.store[hash] = result
}

func (c *stubCache) HashSpec(spec string) string {
	return "hash-" + spec
}

func TestPipelineCacheHit(t *testing.T) {
	t.Parallel()
	cache := &stubCache{store: make(map[string]*Result)}
	p, err := New(Config{Cache: cache})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	spec := "project: cache-test"
	result1, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	result2, err := p.Run(spec)
	if err != nil {
		t.Fatalf("Run (cached): %v", err)
	}
	if result2.Source != result1.Source {
		t.Errorf("cached result source differs")
	}
}

func TestPipelineCacheMissDifferentSpec(t *testing.T) {
	t.Parallel()
	cache := &stubCache{store: make(map[string]*Result)}
	p, err := New(Config{Cache: cache})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = p.Run("project: spec-a")
	if err != nil {
		t.Fatalf("Run a: %v", err)
	}
	result2, err := p.Run("project: spec-b")
	if err != nil {
		t.Fatalf("Run b: %v", err)
	}
	if result2.NEIR.Project.Name != "spec-b" {
		t.Errorf("expected spec-b, got %q", result2.NEIR.Project.Name)
	}
}

func TestPipelineCacheValidate(t *testing.T) {
	t.Parallel()
	cache := &stubCache{store: make(map[string]*Result)}
	p, err := New(Config{Cache: cache})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	spec := "project: cache-validate"
	_, err = p.Validate(spec)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	// Validate should populate cache
	hash := cache.HashSpec(spec)
	if _, ok := cache.Get(hash); !ok {
		t.Error("expected cache to be populated after Validate")
	}
}

type hookTestError struct{ msg string }

func (e *hookTestError) Error() string { return e.msg }

func TestPipelineHookErrorPropagation(t *testing.T) {
	t.Parallel()
	p, err := New(Config{
		Hooks: &Hooks{
			AfterParse: []HookFunc{func(ctx *HookContext) error {
				return &hookTestError{msg: "after-parse failed"}
			}},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = p.Run("project: hook-fail")
	if err == nil {
		t.Error("expected error from failing hook")
	}
}

func TestPipelineBeforeGenerateHook(t *testing.T) {
	t.Parallel()
	var called bool
	p, err := New(Config{
		Hooks: &Hooks{
			BeforeGenerate: []HookFunc{func(ctx *HookContext) error {
				called = true
				return nil
			}},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = p.Run("project: gen-hook-test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Error("expected BeforeGenerate hook to be called")
	}
}

func TestPipelineAfterGenerateHook(t *testing.T) {
	t.Parallel()
	var called bool
	p, err := New(Config{
		Hooks: &Hooks{
			AfterGenerate: []HookFunc{func(ctx *HookContext) error {
				called = true
				return nil
			}},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = p.Run("project: after-gen-hook")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Error("expected AfterGenerate hook to be called")
	}
}

func TestPipelineAfterParseHook(t *testing.T) {
	t.Parallel()
	var called bool
	p, err := New(Config{
		Hooks: &Hooks{
			AfterParse: []HookFunc{func(ctx *HookContext) error {
				called = true
				return nil
			}},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = p.Run("project: after-parse-hook")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Error("expected AfterParse hook to be called")
	}
}

func TestPipelinePolicyEvaluation(t *testing.T) {
	t.Parallel()
	rule := policy.Rule{
		RuleID:    "test-rule",
		Condition: "exists:project",
		Enabled:   true,
	}
	p, err := New(Config{
		Policies: []policy.Rule{rule},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run("project: policy-test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestPipelinePolicyEvaluationFailingCondition(t *testing.T) {
	t.Parallel()
	rule := policy.Rule{
		RuleID:    "failing-rule",
		Condition: "exists:nonexistent_key",
		Enabled:   true,
	}
	p, err := New(Config{
		Policies: []policy.Rule{rule},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run("project: policy-fail")
	if err != nil {
		t.Fatalf("Run should not fail on policy condition failure: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestPipelineDisabledPolicySkipped(t *testing.T) {
	t.Parallel()
	rule := policy.Rule{
		RuleID:    "disabled-rule",
		Condition: "exists:nonexistent_key",
		Enabled:   false,
	}
	p, err := New(Config{
		Policies: []policy.Rule{rule},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := p.Run("project: disabled-policy")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestPipelineRunContextCanceled(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.RunContext(ctx, "project: canceled")
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestPipelineValidateContextCanceled(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.ValidateContext(ctx, "project: canceled")
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestPipelineRenderer(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	r := p.Renderer()
	if r == nil {
		t.Error("expected non-nil renderer")
	}
}

func TestPipelineGraphAccessor(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	g := p.Graph()
	if g == nil {
		t.Error("expected non-nil graph")
	}
}

func TestPipelineRegistryAccessor(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	r := p.Registry()
	if r == nil {
		t.Error("expected non-nil registry")
	}
}

func TestPipelineNameAccessor(t *testing.T) {
	t.Parallel()
	p, err := New(Config{Name: "test-pipeline"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != "test-pipeline" {
		t.Errorf("Name() = %q, want %q", p.Name(), "test-pipeline")
	}
}

func TestPipelineNameDefaultAccessor(t *testing.T) {
	t.Parallel()
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != "unnamed" {
		t.Errorf("Name() = %q, want %q", p.Name(), "unnamed")
	}
}
