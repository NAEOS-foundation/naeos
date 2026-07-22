package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type State string

const (
	StatePending   State = "pending"
	StateRunning   State = "running"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
	StateCancelled State = "canceled"
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

type WorkflowEvent string

const (
	EventStepStart    WorkflowEvent = "step_start"
	EventStepComplete WorkflowEvent = "step_complete"
	EventStepFailed   WorkflowEvent = "step_failed"
	EventStepRetry    WorkflowEvent = "step_retry"
)

type WorkflowEventHandler func(event WorkflowEvent, step string, ctx *WorkflowContext, err error)

type RetryConfig struct {
	MaxRetries int
	Backoff    time.Duration
}

type ParallelStepGroup struct {
	Steps      []*WorkflowStep
	WaitForAll bool
}

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
	Name         string
	Steps        []*WorkflowStep
	Machine      *StateMachine
	Context      *WorkflowContext
	EventHandler WorkflowEventHandler
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
	return w.ExecuteWithContext(context.Background())
}

func (w *Workflow) Cancel() error {
	return w.Machine.Trigger("cancel")
}

func (w *Workflow) Status() State {
	return w.Machine.Current()
}

func (w *Workflow) ExecuteWithContext(ctx context.Context) error {
	if err := w.Machine.Trigger("start"); err != nil {
		return err
	}

	for _, step := range w.Steps {
		if ctx.Err() != nil {
			_ = w.Machine.Trigger("cancel")
			return ctx.Err()
		}

		w.Context.Current = step.Name
		w.Context.Steps = append(w.Context.Steps, step.Name)

		w.emitEvent(EventStepStart, step.Name, nil)

		var err error
		if step.Timeout > 0 {
			err = w.executeStepWithTimeout(ctx, step)
		} else {
			err = step.Action(w.Context)
		}

		if err != nil {
			w.Context.Error = err
			w.emitEvent(EventStepFailed, step.Name, err)
			_ = w.Machine.Trigger("error")
			return fmt.Errorf("step %q failed: %w", step.Name, err)
		}

		w.emitEvent(EventStepComplete, step.Name, nil)
		_ = w.Machine.Trigger("next")
	}

	_ = w.Machine.Trigger("complete")
	return nil
}

func (w *Workflow) executeStepWithTimeout(parentCtx context.Context, step *WorkflowStep) error {
	ctx, cancel := context.WithTimeout(parentCtx, step.Timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- step.Action(w.Context)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("step %q timed out after %v", step.Name, step.Timeout)
	case err := <-done:
		return err
	}
}

func (w *Workflow) ExecuteWithRetry(ctx context.Context, config RetryConfig) error {
	if err := w.Machine.Trigger("start"); err != nil {
		return err
	}

	for _, step := range w.Steps {
		if ctx.Err() != nil {
			_ = w.Machine.Trigger("cancel")
			return ctx.Err()
		}

		w.Context.Current = step.Name
		w.Context.Steps = append(w.Context.Steps, step.Name)

		var lastErr error
		for attempt := 0; attempt <= config.MaxRetries; attempt++ {
			if attempt > 0 {
				w.emitEvent(EventStepRetry, step.Name, fmt.Errorf("retry %d/%d", attempt, config.MaxRetries))
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(config.Backoff):
				}
			}

			w.emitEvent(EventStepStart, step.Name, nil)

			var err error
			if step.Timeout > 0 {
				err = w.executeStepWithTimeout(ctx, step)
			} else {
				err = step.Action(w.Context)
			}

			if err == nil {
				lastErr = nil
				break
			}
			lastErr = err
		}

		if lastErr != nil {
			w.Context.Error = lastErr
			w.emitEvent(EventStepFailed, step.Name, lastErr)
			_ = w.Machine.Trigger("error")
			return fmt.Errorf("step %q failed after %d retries: %w", step.Name, config.MaxRetries, lastErr)
		}

		w.emitEvent(EventStepComplete, step.Name, nil)
		_ = w.Machine.Trigger("next")
	}

	_ = w.Machine.Trigger("complete")
	return nil
}

func (w *Workflow) ExecuteParallelGroup(ctx context.Context, groups []*ParallelStepGroup) error {
	if err := w.Machine.Trigger("start"); err != nil {
		return err
	}

	for _, group := range groups {
		if ctx.Err() != nil {
			_ = w.Machine.Trigger("cancel")
			return ctx.Err()
		}

		if len(group.Steps) == 1 {
			step := group.Steps[0]
			w.Context.Current = step.Name
			w.Context.Steps = append(w.Context.Steps, step.Name)
			w.emitEvent(EventStepStart, step.Name, nil)

			if err := step.Action(w.Context); err != nil {
				w.Context.Error = err
				w.emitEvent(EventStepFailed, step.Name, err)
				_ = w.Machine.Trigger("error")
				return fmt.Errorf("step %q failed: %w", step.Name, err)
			}

			w.emitEvent(EventStepComplete, step.Name, nil)
			_ = w.Machine.Trigger("next")
			continue
		}

		var wg sync.WaitGroup
		errCh := make(chan error, len(group.Steps))

		for _, step := range group.Steps {
			wg.Add(1)
			go func(s *WorkflowStep) {
				defer wg.Done()
				w.emitEvent(EventStepStart, s.Name, nil)

				if err := s.Action(w.Context); err != nil {
					w.emitEvent(EventStepFailed, s.Name, err)
					errCh <- fmt.Errorf("step %q failed: %w", s.Name, err)
					return
				}

				w.emitEvent(EventStepComplete, s.Name, nil)
			}(step)
		}

		wg.Wait()
		close(errCh)

		var errs []error
		for err := range errCh {
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			w.Context.Error = errors.Join(errs...)
			_ = w.Machine.Trigger("error")
			return fmt.Errorf("parallel steps failed: %w", errors.Join(errs...))
		}

		_ = w.Machine.Trigger("next")
	}

	_ = w.Machine.Trigger("complete")
	return nil
}

func (w *Workflow) emitEvent(event WorkflowEvent, step string, err error) {
	if w.EventHandler != nil {
		w.EventHandler(event, step, w.Context, err)
	}
}

type WorkflowSnapshot struct {
	Name      string           `json:"name"`
	State     State            `json:"state"`
	Context   *WorkflowContext `json:"context"`
	CreatedAt time.Time        `json:"created_at"`
}

func (w *Workflow) SaveSnapshot(path string) error {
	snapshot := WorkflowSnapshot{
		Name:      w.Name,
		State:     w.Machine.Current(),
		Context:   w.Context,
		CreatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}

	return nil
}

func LoadSnapshot(path string) (*WorkflowSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}

	var snapshot WorkflowSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}

	return &snapshot, nil
}

func (w *Workflow) RestoreFromSnapshot(snapshot *WorkflowSnapshot) {
	w.Context = snapshot.Context
	if w.Context == nil {
		w.Context = &WorkflowContext{
			Data:  make(map[string]any),
			Steps: []string{},
		}
	}
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

type Manager struct {
	workflows map[string]*Workflow
	storePath string
	mu        sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		workflows: make(map[string]*Workflow),
	}
}

// NewManagerWithPath creates a Manager that persists workflows to the given directory.
func NewManagerWithPath(storePath string) *Manager {
	m := &Manager{
		workflows: make(map[string]*Workflow),
		storePath: storePath,
	}
	if err := m.load(); err != nil {
		fmt.Fprintf(os.Stderr, "workflow: load error: %v\n", err)
	}
	return m
}

func (m *Manager) Register(name string, workflow *Workflow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflows[name] = workflow
	return m.save()
}

func (m *Manager) Get(name string) (*Workflow, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	workflow, ok := m.workflows[name]
	return workflow, ok
}

func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.workflows, name)
	return m.save()
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
		slog.Error("workflow not found", "name", name)
		return fmt.Errorf("workflow not found: %s", name)
	}
	slog.Info("workflow executing", "name", name)
	return workflow.Execute()
}

func (m *Manager) ExecuteWithContext(ctx context.Context, name string) error {
	workflow, ok := m.Get(name)
	if !ok {
		slog.Error("workflow not found", "name", name)
		return fmt.Errorf("workflow not found: %s", name)
	}
	slog.Info("workflow executing with context", "name", name)
	return workflow.ExecuteWithContext(ctx)
}

type workflowEntry struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

func (m *Manager) save() error {
	if m.storePath == "" {
		return nil
	}
	if err := os.MkdirAll(m.storePath, 0o755); err != nil {
		return fmt.Errorf("create workflow store: %w", err)
	}
	var entries []workflowEntry
	for name, w := range m.workflows {
		var stepNames []string
		for _, s := range w.Steps {
			stepNames = append(stepNames, s.Name)
		}
		entries = append(entries, workflowEntry{Name: name, Steps: stepNames})
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal workflows: %w", err)
	}
	if err := os.WriteFile(filepath.Join(m.storePath, "workflows.json"), data, 0o600); err != nil {
		return fmt.Errorf("write workflows: %w", err)
	}
	return nil
}

func (m *Manager) load() error {
	if m.storePath == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(m.storePath, "workflows.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read workflows: %w", err)
	}
	var entries []workflowEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("unmarshal workflows: %w", err)
	}
	for _, e := range entries {
		var steps []*WorkflowStep
		for _, name := range e.Steps {
			steps = append(steps, &WorkflowStep{
				Name:     name,
				Action:   func(ctx *WorkflowContext) error { return nil },
				Timeout:  300,
				Required: true,
			})
		}
		m.workflows[e.Name] = NewWorkflow(e.Name, steps)
	}
	return nil
}
