package gateway

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

// ToolRequest describes an action that an agent intends to perform on a
// specific tool or resource. Every request must pass through the execution
// gateway before reaching the runtime.
type ToolRequest struct {
	Tool        string
	Action      string
	Resource    string
	Environment string
	Actor       string
	Payload     map[string]any
	Context     map[string]any
}

// ExecutionResult records the outcome of an authorized tool execution.
type ExecutionResult struct {
	Request    ToolRequest
	Decision   control.Decision
	PolicyID   string
	RuleID     string
	Status     string // "completed", "denied", "failed", "skipped"
	Output     string
	Hash       string // SHA-256 of output/payload
	Duration   time.Duration
	Timestamp  time.Time
	Reasons    []string
}

// AgentAdapter abstracts an external AI coding agent system. Each adapter
// translates between the agent's native tool invocation model and the
// normalized ToolRequest that the gateway understands.
type AgentAdapter interface {
	// Name returns the identifier of the agent (e.g. "claude", "copilot").
	Name() string

	// NormalizeTool converts a native agent tool invocation into a
	// ToolRequest. The raw argument is agent-specific.
	NormalizeTool(raw any) (ToolRequest, error)

	// OnDecision is called after the gateway produces a decision so the
	// adapter can relay the result back to the agent.
	OnDecision(result ExecutionResult) error
}

// Sandbox defines the execution boundary for authorized tool calls. It
// isolates the execution environment and captures output.
type Sandbox interface {
	// Execute runs the authorized tool request inside the sandbox and
	// returns the execution result.
	Execute(req ToolRequest) (string, error)
}

// ControlPlane is the subset of the governance control plane required by
// the gateway. Using a narrow interface keeps the gateway decoupled from
// the full control plane implementation.
type ControlPlane interface {
	Evaluate(req control.Request) (control.DecisionRecord, error)
}

// Option configures an ExecutionGateway.
type Option func(*ExecutionGateway)

// FailClosed controls the behaviour when the sandbox executor returns an
// error. When true (default) the gateway treats execution errors as denied.
func FailClosed(enabled bool) Option {
	return func(g *ExecutionGateway) { g.failClosed = enabled }
}

// ExecutionGateway is the single enforcement boundary between agent intent
// and tool execution. Every tool invocation must pass through this gateway,
// which evaluates the request against registered policies before allowing
// execution.
//
// Architecture:
//
//	Agent → Adapter.NormalizeTool() → Gateway.Authorize()
//	  → ControlPlane.Evaluate() → Decision
//	  → (ALLOW) Sandbox.Execute() → ExecutionResult
//	  → (DENY)  → ExecutionResult{Status: "denied"}
type ExecutionGateway struct {
	controlPlane ControlPlane
	sandbox      Sandbox
	adapters     map[string]AgentAdapter
	restrictions []Restriction

	mu      sync.RWMutex
	history []ExecutionResult
	failClosed bool
}

// New creates an ExecutionGateway with the given control plane and sandbox.
func New(cp ControlPlane, sb Sandbox, opts ...Option) *ExecutionGateway {
	g := &ExecutionGateway{
		controlPlane: cp,
		sandbox:      sb,
		adapters:     make(map[string]AgentAdapter),
		failClosed:   true,
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

// RegisterAdapter registers an agent adapter under the given name.
func (g *ExecutionGateway) RegisterAdapter(name string, adapter AgentAdapter) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.adapters[name] = adapter
}

// Adapter returns the registered adapter for the given name, or nil.
func (g *ExecutionGateway) Adapter(name string) AgentAdapter {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.adapters[name]
}

// AddRestriction appends a command restriction. Restrictions are evaluated
// in order after the control plane decision; a matching restriction always
// results in DENY regardless of the policy outcome.
func (g *ExecutionGateway) AddRestriction(r Restriction) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.restrictions = append(g.restrictions, r)
}

// Authorize evaluates a tool request through the control plane and, if
// allowed, executes it inside the sandbox. The result is recorded in the
// gateway history.
func (g *ExecutionGateway) Authorize(req ToolRequest) (ExecutionResult, error) {
	if req.Tool == "" {
		return ExecutionResult{}, naeoserr.New(naeoserr.ErrValidation, "tool name is required")
	}

	start := time.Now()

	// Derive resource and action from tool name when not explicitly set.
	resource := req.Resource
	action := req.Action
	if resource == "" {
		resource = req.Tool
	}
	if action == "" {
		action = "execute"
	}

	// Evaluate against the control plane.
	rec, err := g.controlPlane.Evaluate(control.Request{
		Resource:    resource,
		Action:      action,
		Environment: req.Environment,
		Actor:       req.Actor,
		Context:     req.Context,
	})
	if err != nil {
		return ExecutionResult{}, naeoserr.Wrapf(err, naeoserr.ErrInternal, "control plane evaluation failed")
	}

	result := ExecutionResult{
		Request:   req,
		Decision:  rec.Decision,
		PolicyID:  rec.PolicyID,
		RuleID:    rec.RuleID,
		Timestamp: time.Now().UTC(),
		Reasons:   rec.Reasons,
	}

	// Check command restrictions before proceeding. A matching restriction
	// overrides the policy decision and denies execution.
	if g.restrictionsMatch(req) {
		result.Status = "denied"
		result.Output = "command restricted by policy"
		result.Duration = time.Since(start)
		g.record(result)
		return result, nil
	}

	switch rec.Decision {
	case control.DecisionDeny:
		result.Status = "denied"
		result.Output = fmt.Sprintf("denied by policy %s: %s", rec.PolicyID, joinReasons(rec.Reasons))
		result.Duration = time.Since(start)
		g.record(result)
		return result, nil

	case control.DecisionRequireApproval:
		result.Status = "denied"
		result.Output = fmt.Sprintf("approval required by policy %s", rec.PolicyID)
		result.Duration = time.Since(start)
		g.record(result)
		return result, nil

	case control.DecisionAllow:
		// Proceed to sandbox execution.
	default:
		result.Status = "denied"
		result.Output = "unknown decision"
		result.Duration = time.Since(start)
		g.record(result)
		return result, nil
	}

	// Execute inside the sandbox.
	output, execErr := g.sandbox.Execute(req)
	result.Duration = time.Since(start)
	if execErr != nil {
		result.Status = "failed"
		result.Output = execErr.Error()
		result.Hash = hashBytes([]byte(result.Output))
		g.record(result)
		if g.failClosed {
			return result, naeoserr.Wrapf(execErr, naeoserr.ErrPipeline, "sandbox execution failed")
		}
		g.record(result)
		return result, nil
	}

	result.Status = "completed"
	result.Output = output
	result.Hash = hashBytes([]byte(output))
	g.record(result)
	return result, nil
}

// AuthorizeFromAdapter normalizes a raw agent invocation through the
// named adapter and then authorizes it.
func (g *ExecutionGateway) AuthorizeFromAdapter(adapterName string, raw any) (ExecutionResult, error) {
	adapter := g.Adapter(adapterName)
	if adapter == nil {
		return ExecutionResult{}, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("adapter %q not registered", adapterName))
	}
	req, err := adapter.NormalizeTool(raw)
	if err != nil {
		return ExecutionResult{}, naeoserr.Wrapf(err, naeoserr.ErrValidation, "adapter %s failed to normalize tool", adapterName)
	}
	result, err := g.Authorize(req)
	if err != nil {
		return result, err
	}
	if err := adapter.OnDecision(result); err != nil {
		return result, naeoserr.Wrapf(err, naeoserr.ErrInternal, "adapter %s failed to relay decision", adapterName)
	}
	return result, nil
}

// History returns the full execution history.
func (g *ExecutionGateway) History() []ExecutionResult {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]ExecutionResult, len(g.history))
	copy(out, g.history)
	return out
}

// Denials returns only denied execution results.
func (g *ExecutionGateway) Denials() []ExecutionResult {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out []ExecutionResult
	for _, r := range g.history {
		if r.Status == "denied" {
			out = append(out, r)
		}
	}
	return out
}

func (g *ExecutionGateway) record(r ExecutionResult) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.history = append(g.history, r)
}

func (g *ExecutionGateway) restrictionsMatch(req ToolRequest) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, r := range g.restrictions {
		if r.Matches(req) {
			return true
		}
	}
	return false
}

func hashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	s := reasons[0]
	for i := 1; i < len(reasons); i++ {
		s += "; " + reasons[i]
	}
	return s
}
