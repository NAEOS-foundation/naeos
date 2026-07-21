package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPipelineGraph(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	g := p.Graph()
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
}

func TestPipelineRegistry(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	r := p.Registry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestPipelineHooks(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	h := p.Hooks()
	if h == nil {
		t.Fatal("expected non-nil hooks")
	}
}

func TestPipelineRegisteredKernelServices(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	services := p.RegisteredKernelServices()
	if len(services) == 0 {
		t.Error("expected at least one kernel service")
	}
}

func TestPipelineKernelMetrics(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	m := p.KernelMetrics()
	_ = m
}

func TestPipelineKernelTopics(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	topics := p.KernelTopics()
	_ = topics
}

func TestPipelinePublish(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = p.Publish("test.topic", "payload")
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
}

func TestPipelineSubscribe(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	received := false
	err = p.Subscribe("test.topic", func(v any) {
		received = true
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	p.Publish("test.topic", "data")
	if !received {
		t.Error("expected handler to be called")
	}
}

func TestPipelineValidate(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := p.Validate("project: test")
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPipelineValidateContext(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := p.ValidateContext(context.Background(), "project: test")
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPipelineValidateEmptyInput(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Validate("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestWithCache(t *testing.T) {
	cfg := Config{}
	opt := WithCache(nil)
	opt(&cfg)
}

func TestConfigFromFileNotFound(t *testing.T) {
	_, err := ConfigFromFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error")
	}
}

func TestConfigFromFileInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte("not json"), 0o644)
	_, err := ConfigFromFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestConfigFromFileInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(": invalid yaml :["), 0o644)
	_, err := ConfigFromFile(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestNameDefault(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	name := p.Name()
	if name != "unnamed" {
		t.Errorf("expected 'unnamed', got %q", name)
	}
}

func TestReader(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	_ = p
}

type captureObserver struct {
	started  string
	complete string
	failed   string
	artifact string
}

func (o *captureObserver) OnPipelineStart(pipelineID string)      { o.started = pipelineID }
func (o *captureObserver) OnPipelineComplete(pid string, artifacts int, dur string) { o.complete = pid }
func (o *captureObserver) OnPipelineFailed(pid, errMsg string)           { o.failed = pid }
func (o *captureObserver) OnArtifactGenerated(name, path string)         { o.artifact = name }

func TestChainObservers(t *testing.T) {
	o1 := &captureObserver{}
	o2 := &captureObserver{}
	chain := ChainObservers(o1, o2)
	if chain == nil {
		t.Fatal("expected non-nil observer")
	}

	chain.OnPipelineStart("test-1")
	if o1.started != "test-1" || o2.started != "test-1" {
		t.Error("OnPipelineStart should fan-out")
	}

	chain.OnPipelineComplete("test-2", 3, "100ms")
	if o1.complete != "test-2" {
		t.Error("OnPipelineComplete should fan-out")
	}

	chain.OnPipelineFailed("test-3", "error")
	if o1.failed != "test-3" {
		t.Error("OnPipelineFailed should fan-out")
	}

	chain.OnArtifactGenerated("main.go", "/out/main.go")
	if o1.artifact != "main.go" {
		t.Error("OnArtifactGenerated should fan-out")
	}
}

func TestChainObserversNil(t *testing.T) {
	chain := ChainObservers()
	chain.OnPipelineStart("noop")
}

func TestPipelineRunWithObserver(t *testing.T) {
	o := &captureObserver{}
	p, err := New(Config{Observer: o})
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Run("project: test\nmodules:\n  - name: core\n    path: ./core")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if o.started == "" {
		t.Error("expected OnPipelineStart to be called")
	}
	if o.complete == "" {
		t.Error("expected OnPipelineComplete to be called")
	}
}

func TestPipelineRunFailsWithObserver(t *testing.T) {
	o := &captureObserver{}
	p, err := New(Config{Observer: o})
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Run("")
	if err == nil {
		t.Fatal("expected error")
	}
	if o.failed == "" {
		t.Error("expected OnPipelineFailed to be called")
	}
}
