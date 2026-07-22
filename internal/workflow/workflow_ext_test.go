package workflow

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestManagerWithPathPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	m1 := NewManagerWithPath(path)
	wf1 := NewWorkflow("persist-test", []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error { return nil }},
	})
	m1.Register("wf1", wf1)

	m2 := NewManagerWithPath(path)
	wf2, ok := m2.Get("wf1")
	if !ok {
		t.Fatal("workflow not found after reload")
	}
	if wf2.Name != "wf1" {
		t.Errorf("expected 'persist-test', got %q", wf2.Name)
	}
}

func TestExecuteParallelGroupSingleStep(t *testing.T) {
	var ran bool
	groups := []*ParallelStepGroup{
		{Steps: []*WorkflowStep{{Name: "alone", Action: func(ctx *WorkflowContext) error { ran = true; return nil }}}},
	}
	err := NewWorkflow("test", []*WorkflowStep{
		{Name: "init", Action: func(ctx *WorkflowContext) error { return nil }},
	}).ExecuteParallelGroup(context.Background(), groups)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Error("expected step to execute")
	}
}

func TestExecuteParallelGroupSingleStepFailure(t *testing.T) {
	w := NewWorkflow("test", []*WorkflowStep{
		{Name: "init", Action: func(ctx *WorkflowContext) error { return nil }},
	})
	groups := []*ParallelStepGroup{
		{Steps: []*WorkflowStep{{Name: "fail", Action: func(ctx *WorkflowContext) error { return errors.New("step failed") }}}},
	}
	err := w.ExecuteParallelGroup(context.Background(), groups)
	if err == nil {
		t.Error("expected error from failed step")
	}
}

func TestExecuteWithRetryStartFailure(t *testing.T) {
	w := NewWorkflow("test", []*WorkflowStep{
		{Name: "s1", Action: func(ctx *WorkflowContext) error { return nil }},
	})
	w.Machine.Trigger("start")
	err := w.ExecuteWithRetry(context.Background(), RetryConfig{MaxRetries: 1, Backoff: time.Millisecond})
	if err == nil {
		t.Error("expected start failure error")
	}
}

func TestExecuteWithContextStartFailure(t *testing.T) {
	w := NewWorkflow("test", []*WorkflowStep{
		{Name: "s1", Action: func(ctx *WorkflowContext) error { return nil }},
	})
	w.Machine.Trigger("start")
	err := w.ExecuteWithContext(context.Background())
	if err == nil {
		t.Error("expected start failure error")
	}
}

func TestExecuteWithRetryContextCancel(t *testing.T) {
	w := NewWorkflow("test", []*WorkflowStep{
		{Name: "s1", Action: func(ctx *WorkflowContext) error {
			return errors.New("always fail")
		}},
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := w.ExecuteWithRetry(ctx, RetryConfig{MaxRetries: 3, Backoff: 100 * time.Millisecond})
	if err == nil {
		t.Error("expected error")
	}
}

func TestApprovalWorkflowRejectNotFound(t *testing.T) {
	aw := NewApprovalWorkflow()
	err := aw.Reject("nonexistent", "approver", "no thanks")
	if err == nil {
		t.Error("expected error for nonexistent request")
	}
}

func TestRestoreFromSnapshotNilContext(t *testing.T) {
	w := NewWorkflow("test", []*WorkflowStep{
		{Name: "s1", Action: func(ctx *WorkflowContext) error { return nil }},
	})
	snap := &WorkflowSnapshot{
		Name:      "test",
		State:     "pending",
		Context:   nil,
		CreatedAt: time.Now(),
	}
	w.RestoreFromSnapshot(snap)
	if w.Context == nil {
		t.Error("expected context to be initialized")
	}
}

func TestSaveLoadSnapshotErrors(t *testing.T) {
	w := NewWorkflow("test", []*WorkflowStep{
		{Name: "s1", Action: func(ctx *WorkflowContext) error { return nil }},
	})

	err := w.SaveSnapshot("/nonexistent/deep/dir/snap.json")
	if err == nil {
		t.Error("expected error for bad path")
	}

	_, err = LoadSnapshot("/nonexistent/file.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExecuteWithRetryStepTimeout(t *testing.T) {
	w := NewWorkflow("test", []*WorkflowStep{
		{
			Name:    "slow",
			Timeout: time.Millisecond,
			Action: func(ctx *WorkflowContext) error {
				time.Sleep(5 * time.Second)
				return nil
			},
		},
	})
	err := w.ExecuteWithRetry(context.Background(), RetryConfig{MaxRetries: 1, Backoff: time.Millisecond})
	if err == nil {
		t.Error("expected timeout error")
	}
}
