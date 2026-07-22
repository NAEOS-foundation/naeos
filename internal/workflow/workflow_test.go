package workflow

import (
	"fmt"
	"testing"
	"time"
)

func TestStateMachine(t *testing.T) {
	sm := NewStateMachine(StatePending)

	sm.AddTransition(StatePending, StateRunning, "start")
	sm.AddTransition(StateRunning, StateCompleted, "complete")

	if sm.Current() != StatePending {
		t.Errorf("expected pending, got %s", sm.Current())
	}

	err := sm.Trigger("start")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.Current() != StateRunning {
		t.Errorf("expected running, got %s", sm.Current())
	}

	err = sm.Trigger("complete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.Current() != StateCompleted {
		t.Errorf("expected completed, got %s", sm.Current())
	}
}

func TestStateMachineInvalidTransition(t *testing.T) {
	sm := NewStateMachine(StatePending)
	sm.AddTransition(StatePending, StateRunning, "start")

	err := sm.Trigger("invalid")
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestStateMachineCanTransition(t *testing.T) {
	sm := NewStateMachine(StatePending)
	sm.AddTransition(StatePending, StateRunning, "start")

	if !sm.CanTransition("start") {
		t.Error("expected can transition")
	}

	if sm.CanTransition("invalid") {
		t.Error("expected cannot transition")
	}
}

func TestStateMachineHistory(t *testing.T) {
	sm := NewStateMachine(StatePending)
	sm.AddTransition(StatePending, StateRunning, "start")
	sm.AddTransition(StateRunning, StateCompleted, "complete")

	sm.Trigger("start")
	sm.Trigger("complete")

	history := sm.History()
	if len(history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(history))
	}
}

func TestWorkflow(t *testing.T) {
	var steps []string

	wf := NewWorkflow("test", []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			steps = append(steps, "step1")
			return nil
		}},
		{Name: "step2", Action: func(ctx *WorkflowContext) error {
			steps = append(steps, "step2")
			return nil
		}},
		{Name: "step3", Action: func(ctx *WorkflowContext) error {
			steps = append(steps, "step3")
			return nil
		}},
	})

	err := wf.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf.Status() != StateCompleted {
		t.Errorf("expected completed, got %s", wf.Status())
	}

	if len(steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(steps))
	}
}

func TestWorkflowError(t *testing.T) {
	wf := NewWorkflow("test", []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			return nil
		}},
		{Name: "step2", Action: func(ctx *WorkflowContext) error {
			return fmt.Errorf("step2 failed")
		}},
		{Name: "step3", Action: func(ctx *WorkflowContext) error {
			return nil
		}},
	})

	err := wf.Execute()
	if err == nil {
		t.Error("expected error")
	}

	if wf.Status() != StateFailed {
		t.Errorf("expected failed, got %s", wf.Status())
	}
}

func TestWorkflowCancel(t *testing.T) {
	wf := NewWorkflow("test", []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			return nil
		}},
	})

	wf.Machine.Trigger("start")
	wf.Cancel()

	if wf.Status() != StateCancelled {
		t.Errorf("expected canceled, got %s", wf.Status())
	}
}

func TestWorkflowContext(t *testing.T) {
	wf := NewWorkflow("test", []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			ctx.Data["key"] = "value"
			return nil
		}},
	})

	wf.Execute()

	if wf.Context.Data["key"] != "value" {
		t.Error("expected data to be set")
	}
}

func TestApprovalWorkflow(t *testing.T) {
	aw := NewApprovalWorkflow()

	req := aw.CreateRequest("req1", "deploy", "user1")
	if req.Status != "pending" {
		t.Errorf("expected pending, got %s", req.Status)
	}

	err := aw.Approve("req1", "admin", "Looks good")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req, ok := aw.GetRequest("req1")
	if !ok {
		t.Fatal("expected request to be found")
	}
	if req.Status != "approved" {
		t.Errorf("expected approved, got %s", req.Status)
	}
	if req.Approver != "admin" {
		t.Errorf("expected approver 'admin', got %s", req.Approver)
	}
}

func TestApprovalWorkflowReject(t *testing.T) {
	aw := NewApprovalWorkflow()

	aw.CreateRequest("req1", "deploy", "user1")
	aw.Reject("req1", "admin", "Not ready")

	req, _ := aw.GetRequest("req1")
	if req.Status != "rejected" {
		t.Errorf("expected rejected, got %s", req.Status)
	}
}

func TestApprovalWorkflowNotFound(t *testing.T) {
	aw := NewApprovalWorkflow()

	err := aw.Approve("nonexistent", "admin", "")
	if err == nil {
		t.Error("expected error for nonexistent request")
	}
}

func TestApprovalWorkflowListByStatus(t *testing.T) {
	aw := NewApprovalWorkflow()

	aw.CreateRequest("req1", "deploy", "user1")
	aw.CreateRequest("req2", "deploy", "user2")
	aw.CreateRequest("req3", "deploy", "user3")

	aw.Approve("req1", "admin", "")

	pending := aw.ListByStatus("pending")
	if len(pending) != 2 {
		t.Errorf("expected 2 pending, got %d", len(pending))
	}

	approved := aw.ListByStatus("approved")
	if len(approved) != 1 {
		t.Errorf("expected 1 approved, got %d", len(approved))
	}
}

func TestManager(t *testing.T) {
	m := NewManager()

	wf := NewWorkflow("test", []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			return nil
		}},
	})

	_ = m.Register("test", wf)

	got, ok := m.Get("test")
	if !ok {
		t.Fatal("expected workflow to be found")
	}
	if got.Name != "test" {
		t.Errorf("expected name 'test', got %s", got.Name)
	}

	names := m.List()
	if len(names) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(names))
	}

	_ = m.Remove("test")
	_, ok = m.Get("test")
	if ok {
		t.Error("expected workflow to be removed")
	}
}

func TestManagerExecute(t *testing.T) {
	m := NewManager()

	wf := NewWorkflow("test", []*WorkflowStep{
		{Name: "step1", Action: func(ctx *WorkflowContext) error {
			return nil
		}},
	})

	_ = m.Register("test", wf)

	err := m.Execute("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerExecuteNotFound(t *testing.T) {
	m := NewManager()

	err := m.Execute("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent workflow")
	}
}

func TestWorkflowStepTimeout(t *testing.T) {
	wf := NewWorkflow("test", []*WorkflowStep{
		{Name: "step1", Timeout: time.Second, Action: func(ctx *WorkflowContext) error {
			return nil
		}},
	})

	err := wf.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
