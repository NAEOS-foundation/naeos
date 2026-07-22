package workflow

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecuteWithContext(t *testing.T) {
	var executed bool
	steps := []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			executed = true
			return nil
		}},
	}

	w := NewWorkflow("test", steps)
	err := w.ExecuteWithContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("expected step to be executed")
	}
}

func TestExecuteWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	steps := []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			return nil
		}},
	}

	w := NewWorkflow("test", steps)
	err := w.ExecuteWithContext(ctx)
	if err == nil {
		t.Error("expected error from canceled context")
	}
}

func TestExecuteStepWithTimeout(t *testing.T) {
	steps := []*WorkflowStep{
		{
			Name:    "slow-step",
			Timeout: 50 * time.Millisecond,
			Action: func(ctx *WorkflowContext) error {
				time.Sleep(200 * time.Millisecond)
				return nil
			},
		},
	}

	w := NewWorkflow("test", steps)
	err := w.ExecuteWithContext(context.Background())
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestExecuteStepWithTimeoutSuccess(t *testing.T) {
	steps := []*WorkflowStep{
		{
			Name:    "fast-step",
			Timeout: 200 * time.Millisecond,
			Action: func(ctx *WorkflowContext) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		},
	}

	w := NewWorkflow("test", steps)
	err := w.ExecuteWithContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteWithRetry(t *testing.T) {
	var attempts atomic.Int32
	steps := []*WorkflowStep{
		{
			Name: "retry-step",
			Action: func(ctx *WorkflowContext) error {
				if attempts.Add(1) < 3 {
					return errors.New("transient error")
				}
				return nil
			},
		},
	}

	w := NewWorkflow("test", steps)
	config := RetryConfig{MaxRetries: 5, Backoff: 10 * time.Millisecond}
	err := w.ExecuteWithRetry(context.Background(), config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestExecuteWithRetryExhausted(t *testing.T) {
	steps := []*WorkflowStep{
		{
			Name: "fail-step",
			Action: func(ctx *WorkflowContext) error {
				return errors.New("permanent error")
			},
		},
	}

	w := NewWorkflow("test", steps)
	config := RetryConfig{MaxRetries: 2, Backoff: 10 * time.Millisecond}
	err := w.ExecuteWithRetry(context.Background(), config)
	if err == nil {
		t.Error("expected error after retries exhausted")
	}
}

func TestExecuteParallelGroup(t *testing.T) {
	var count atomic.Int32
	steps := []*WorkflowStep{
		{Name: "a", Action: func(ctx *WorkflowContext) error {
			count.Add(1)
			time.Sleep(10 * time.Millisecond)
			return nil
		}},
		{Name: "b", Action: func(ctx *WorkflowContext) error {
			count.Add(1)
			time.Sleep(10 * time.Millisecond)
			return nil
		}},
	}

	groups := []*ParallelStepGroup{
		{Steps: steps, WaitForAll: true},
	}

	w := NewWorkflow("parallel-test", steps)

	err := w.ExecuteParallelGroup(context.Background(), groups)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count.Load() != 2 {
		t.Errorf("expected 2 steps executed, got %d", count.Load())
	}
}

func TestExecuteParallelGroupFailure(t *testing.T) {
	steps := []*WorkflowStep{
		{Name: "fail", Action: func(ctx *WorkflowContext) error {
			return errors.New("step failed")
		}},
		{Name: "ok", Action: func(ctx *WorkflowContext) error {
			return nil
		}},
	}

	groups := []*ParallelStepGroup{
		{Steps: steps, WaitForAll: true},
	}

	w := NewWorkflow("parallel-fail", nil)
	w.Steps = steps

	err := w.ExecuteParallelGroup(context.Background(), groups)
	if err == nil {
		t.Error("expected error from parallel group")
	}
}

func TestEventHandler(t *testing.T) {
	var events []WorkflowEvent
	var mu sync.Mutex

	steps := []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			return nil
		}},
	}

	w := NewWorkflow("test", steps)
	w.EventHandler = func(event WorkflowEvent, step string, ctx *WorkflowContext, err error) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
	}

	w.Execute()

	mu.Lock()
	defer mu.Unlock()
	if len(events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(events))
	}
}

func TestSaveAndLoadSnapshot(t *testing.T) {
	steps := []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error { return nil }},
	}

	w := NewWorkflow("test", steps)
	w.Context.Data["key"] = "value"

	tmpDir := t.TempDir()
	snapshotPath := tmpDir + "/snapshot.json"

	if err := w.SaveSnapshot(snapshotPath); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	loaded, err := LoadSnapshot(snapshotPath)
	if err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}

	if loaded.Name != "test" {
		t.Errorf("expected name test, got %s", loaded.Name)
	}
}

func TestRestoreFromSnapshot(t *testing.T) {
	snapshot := &WorkflowSnapshot{
		Name:  "restored",
		State: StateRunning,
		Context: &WorkflowContext{
			Data:  map[string]any{"restored": true},
			Steps: []string{"step1"},
		},
	}

	w := NewWorkflow("original", nil)
	w.RestoreFromSnapshot(snapshot)

	if w.Context.Data["restored"] != true {
		t.Error("expected restored context data")
	}
}

func TestManagerExecuteWithContext(t *testing.T) {
	var executed bool
	steps := []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			executed = true
			return nil
		}},
	}

	m := NewManager()
	w := NewWorkflow("test", steps)
	_ = m.Register("test", w)

	err := m.ExecuteWithContext(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("expected step to be executed")
	}
}

func TestManagerExecuteWithContextNotFound(t *testing.T) {
	m := NewManager()
	err := m.ExecuteWithContext(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent workflow")
	}
}
