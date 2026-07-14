package workflow

import (
	"fmt"
	"sync"
	"time"
)

// State Machine

type State string

const (
	StatePending   State = "pending"
	StateRunning   State = "running"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

type Transition struct {
	From  State
	To    State
	Event string
}

type StateMachine struct {
	current     State
	transitions map[string]Transition
	history     []StateTransition
	mu          sync.RWMutex
}

type StateTransition struct {
	From      State
	To        State
	Event     string
	Timestamp time.Time
}

func NewStateMachine(initial State) *StateMachine {
	return &StateMachine{
		current:     initial,
		transitions: make(map[string]Transition),
		history:     []StateTransition{},
	}
}

func (sm *StateMachine) AddTransition(from, to State, event string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	key := fmt.Sprintf("%s->%s", from, event)
	sm.transitions[key] = Transition{From: from, To: to, Event: event}
}

func (sm *StateMachine) Trigger(event string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := fmt.Sprintf("%s->%s", sm.current, event)
	transition, ok := sm.transitions[key]
	if !ok {
		return fmt.Errorf("no transition from %s with event %s", sm.current, event)
	}

	sm.history = append(sm.history, StateTransition{
		From:      sm.current,
		To:        transition.To,
		Event:     event,
		Timestamp: time.Now(),
	})

	sm.current = transition.To
	return nil
}

func (sm *StateMachine) Current() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

func (sm *StateMachine) History() []StateTransition {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.history
}

func (sm *StateMachine) CanTransition(event string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	key := fmt.Sprintf("%s->%s", sm.current, event)
	_, ok := sm.transitions[key]
	return ok
}

// Workflow

type WorkflowStep struct {
	Name     string
	Action   func(ctx *WorkflowContext) error
	Timeout  time.Duration
	Required bool
}

type WorkflowContext struct {
	Data    map[string]any
	Steps   []string
	Current string
	Error   error
}

type Workflow struct {
	Name     string
	Steps    []*WorkflowStep
	Machine  *StateMachine
	Context  *WorkflowContext
	mu       sync.RWMutex
}

func NewWorkflow(name string, steps []*WorkflowStep) *Workflow {
	machine := NewStateMachine(StatePending)

	for i := range steps {
		if i == 0 {
			machine.AddTransition(StatePending, StateRunning, "start")
		}
		machine.AddTransition(StateRunning, StateRunning, "next")
		if i == len(steps)-1 {
			machine.AddTransition(StateRunning, StateCompleted, "complete")
		}
	}
	machine.AddTransition(StateRunning, StateFailed, "error")
	machine.AddTransition(StateRunning, StateCancelled, "cancel")

	return &Workflow{
		Name:    name,
		Steps:   steps,
		Machine: machine,
		Context: &WorkflowContext{
			Data:  make(map[string]any),
			Steps: []string{},
		},
	}
}

func (w *Workflow) Execute() error {
	w.Machine.Trigger("start")

	for _, step := range w.Steps {
		w.Context.Current = step.Name
		w.Context.Steps = append(w.Context.Steps, step.Name)

		if err := step.Action(w.Context); err != nil {
			w.Context.Error = err
			w.Machine.Trigger("error")
			return err
		}

		w.Machine.Trigger("next")
	}

	w.Machine.Trigger("complete")
	return nil
}

func (w *Workflow) Cancel() error {
	w.Machine.Trigger("cancel")
	return nil
}

func (w *Workflow) Status() State {
	return w.Machine.Current()
}

// Approval Workflow

type ApprovalRequest struct {
	ID        string
	Workflow  string
	Requester string
	Status    string
	Approver  string
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ApprovalWorkflow struct {
	requests map[string]*ApprovalRequest
	mu       sync.RWMutex
}

func NewApprovalWorkflow() *ApprovalWorkflow {
	return &ApprovalWorkflow{
		requests: make(map[string]*ApprovalRequest),
	}
}

func (a *ApprovalWorkflow) CreateRequest(id, workflow, requester string) *ApprovalRequest {
	a.mu.Lock()
	defer a.mu.Unlock()

	req := &ApprovalRequest{
		ID:        id,
		Workflow:  workflow,
		Requester: requester,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	a.requests[id] = req
	return req
}

func (a *ApprovalWorkflow) Approve(id, approver, comment string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	req, ok := a.requests[id]
	if !ok {
		return fmt.Errorf("request not found: %s", id)
	}

	req.Status = "approved"
	req.Approver = approver
	req.Comment = comment
	req.UpdatedAt = time.Now()
	return nil
}

func (a *ApprovalWorkflow) Reject(id, approver, comment string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	req, ok := a.requests[id]
	if !ok {
		return fmt.Errorf("request not found: %s", id)
	}

	req.Status = "rejected"
	req.Approver = approver
	req.Comment = comment
	req.UpdatedAt = time.Now()
	return nil
}

func (a *ApprovalWorkflow) GetRequest(id string) (*ApprovalRequest, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, ok := a.requests[id]
	return req, ok
}

func (a *ApprovalWorkflow) ListByStatus(status string) []*ApprovalRequest {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var reqs []*ApprovalRequest
	for _, req := range a.requests {
		if req.Status == status {
			reqs = append(reqs, req)
		}
	}
	return reqs
}

// Workflow Manager

type Manager struct {
	workflows map[string]*Workflow
	mu        sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		workflows: make(map[string]*Workflow),
	}
}

func (m *Manager) Register(name string, workflow *Workflow) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflows[name] = workflow
}

func (m *Manager) Get(name string) (*Workflow, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	workflow, ok := m.workflows[name]
	return workflow, ok
}

func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.workflows, name)
}

func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.workflows))
	for name := range m.workflows {
		names = append(names, name)
	}
	return names
}

func (m *Manager) Execute(name string) error {
	workflow, ok := m.Get(name)
	if !ok {
		return fmt.Errorf("workflow not found: %s", name)
	}
	return workflow.Execute()
}
