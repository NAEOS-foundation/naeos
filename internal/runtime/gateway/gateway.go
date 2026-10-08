// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

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
	// RequestID identifies the logical request across adapter, authorization, execution, and evidence.
	RequestID string
	// InvocationID uniquely identifies one intended execution. When replay
	// protection is enabled, the gateway consumes it atomically before the
	// sandbox side effect and rejects reuse.
	InvocationID string
	// Capability is the normalized capability requested by the agent.
	// It is authorization-bound and must not be widened downstream.
	Capability  string
	Tool        string
	Action      string
	Resource    string
	Environment string
	Actor       string
	Payload     map[string]any
	Context     map[string]any
}

// ExecutionResult records the outcome of an authorized tool execution.
// Observation is an independently produced runtime record of the externally
// observable result of an execution. It deliberately does not trust the sandbox
// output as proof of a side effect.
type Observation struct {
	Status       string // "observed", "absent", "mismatch", "unavailable"
	Observed     bool
	ArtifactHash string
	ArtifactSize int64
	Metadata     map[string]string
	Timestamp    time.Time
}

// Observer verifies or records the externally observable effect of an execution.
// Implementations should inspect the target system rather than infer state from
// the sandbox's claimed output.
type Observer interface {
	Observe(req ToolRequest, result ExecutionResult) (Observation, error)
}

type ExecutionResult struct {
	RequestID    string
	InvocationID string
	Request      ToolRequest
	Decision     control.Decision
	PolicyID     string
	RuleID       string
	Status       string // "completed", "denied", "failed", "skipped"
	Output       string
	Hash         string // SHA-256 of output/payload
	Duration     time.Duration
	Timestamp    time.Time
	Reasons      []string
	Observation  *Observation
}

// AgentAdapter abstracts an external AI coding agent system. Each adapter
// translates between the agent's native tool invocation model and the
// normalized ToolRequest that the gateway understands.
// AdapterPolicy is the explicit privilege boundary for an adapter. An adapter
// may normalize any input, but it can reach the execution gateway only when its
// registered policy grants the requested tool/action/environment/capability.
type AdapterPolicy struct {
	AllowedTools        []string
	AllowedActions      []string
	AllowedEnvironments []string
	AllowedCapabilities []string
}

func (p AdapterPolicy) allows(req ToolRequest) bool {
	return adapterFieldAllowed(p.AllowedTools, req.Tool) &&
		adapterFieldAllowed(p.AllowedActions, req.Action) &&
		adapterFieldAllowed(p.AllowedEnvironments, req.Environment) &&
		adapterFieldAllowed(p.AllowedCapabilities, req.Capability)
}

func adapterFieldAllowed(patterns []string, value string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if matchPattern(pattern, value) {
			return true
		}
	}
	return false
}

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

// DecisionRevalidator is an optional second-phase authorization check.
// Production control planes should implement it so policy changes between
// authorization and execution invalidate the original decision.
type DecisionRevalidator interface {
	ValidateDecision(req control.Request, issued control.DecisionRecord) (control.DecisionRecord, error)
}

// Option configures an ExecutionGateway.
type Option func(*ExecutionGateway)

// FailClosed controls the behavior when the sandbox executor returns an
// error. When true (default) the gateway treats execution errors as denied.
func FailClosed(enabled bool) Option {
	return func(g *ExecutionGateway) { g.failClosed = enabled }
}

// WithObserver installs a first-class runtime observer. When configured,
// successful execution is not considered evidence-backed until the observer
// records the externally observable result.
func WithObserver(observer Observer) Option {
	return func(g *ExecutionGateway) { g.observer = observer }
}

// WithReplayProtection enables atomic invocation-id replay protection. When
// enabled, every execution request must provide a non-empty InvocationID.
func WithReplayProtection(enabled bool) Option {
	return func(g *ExecutionGateway) { g.replayProtection = enabled }
}

// WithInvocationStore installs the persistence boundary used to consume
// invocation identities.
func WithInvocationStore(store InvocationStore) Option {
	return func(g *ExecutionGateway) {
		if store != nil {
			g.invocationStore = store
		}
	}
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
	controlPlane    ControlPlane
	sandbox         Sandbox
	observer        Observer
	adapters        map[string]AgentAdapter
	adapterPolicies map[string]AdapterPolicy
	restrictions    []Restriction

	mu               sync.RWMutex
	history          []ExecutionResult
	invocationStore  InvocationStore
	replayProtection bool
	failClosed       bool
}

// New creates an ExecutionGateway with the given control plane and sandbox.
func New(cp ControlPlane, sb Sandbox, opts ...Option) *ExecutionGateway {
	g := &ExecutionGateway{
		controlPlane:    cp,
		sandbox:         sb,
		adapters:        make(map[string]AgentAdapter),
		adapterPolicies: make(map[string]AdapterPolicy),
		invocationStore: NewInMemoryInvocationStore(),
		failClosed:      true,
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

// RegisterAdapter registers an adapter without execution privilege. Callers
// must grant an explicit AdapterPolicy before AuthorizeFromAdapter can execute it.
func (g *ExecutionGateway) RegisterAdapter(name string, adapter AgentAdapter) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.adapters[name] = adapter
	delete(g.adapterPolicies, name)
}

// GrantAdapterPolicy gives an adapter an explicit, least-privilege execution grant.
func (g *ExecutionGateway) GrantAdapterPolicy(name string, policy AdapterPolicy) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.adapters[name]; !ok {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("adapter %q is not registered", name))
	}
	if len(policy.AllowedTools) == 0 || len(policy.AllowedActions) == 0 {
		return naeoserr.New(naeoserr.ErrValidation, "adapter policy must explicitly grant tools and actions")
	}
	g.adapterPolicies[name] = policy
	return nil
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
	if g.replayProtection && req.InvocationID == "" {
		return ExecutionResult{}, naeoserr.New(naeoserr.ErrValidation, "invocation id is required when replay protection is enabled")
	}

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
		Capability:  req.Capability,
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
		RequestID:    req.RequestID,
		InvocationID: req.InvocationID,
		Request:      req,
		Decision:     rec.Decision,
		PolicyID:     rec.PolicyID,
		RuleID:       rec.RuleID,
		Timestamp:    time.Now().UTC(),
		Reasons:      rec.Reasons,
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

	// Revalidate immediately before execution when the control plane supports
	// a second-phase check. This closes the policy-mutation window between the
	// initial authorization decision and the externally observable side effect.
	if revalidator, ok := g.controlPlane.(DecisionRevalidator); ok {
		current, err := revalidator.ValidateDecision(control.Request{
			Capability:  req.Capability,
			Resource:    resource,
			Action:      action,
			Environment: req.Environment,
			Actor:       req.Actor,
			Context:     req.Context,
		}, rec)
		if err != nil {
			result.Status = "denied"
			result.Output = "authorization invalidated before execution"
			result.Duration = time.Since(start)
			g.record(result)
			return result, nil
		}
		rec = current
		result.Decision = current.Decision
		result.PolicyID = current.PolicyID
		result.RuleID = current.RuleID
		result.Reasons = current.Reasons
		if current.Decision != control.DecisionAllow {
			result.Status = "denied"
			result.Output = "authorization invalidated before execution"
			result.Duration = time.Since(start)
			g.record(result)
			return result, nil
		}
	}

	// Atomically consume the invocation identity immediately before the side
	// effect. This is the replay boundary: concurrent or subsequent reuse of
	// the same invocation cannot reach the sandbox.
	if g.replayProtection {
		claimed, claimErr := g.claimInvocation(req.InvocationID)
		if claimErr != nil {
			result.Status = "denied"
			result.Output = "invocation replay state unavailable"
			result.Duration = time.Since(start)
			g.record(result)
			if g.failClosed {
				return result, naeoserr.Wrapf(claimErr, naeoserr.ErrPipeline, "invocation replay state unavailable")
			}
			return result, nil
		}
		if !claimed {
			result.Status = "denied"
			result.Output = "invocation already consumed"
			result.Duration = time.Since(start)
			g.record(result)
			return result, nil
		}
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

	// A sandbox completion is an execution claim, not proof that the requested
	// side effect became externally observable. If an observer is configured,
	// it becomes part of the runtime completion boundary.
	if g.observer != nil {
		observation, observeErr := g.observer.Observe(req, result)
		result.Observation = &observation
		if observeErr != nil {
			result.Status = "failed"
			result.Output = observeErr.Error()
			result.Hash = hashBytes([]byte(result.Output))
			g.record(result)
			if g.failClosed {
				return result, naeoserr.Wrapf(observeErr, naeoserr.ErrPipeline, "runtime observation failed")
			}
			return result, nil
		}
		if !observation.Observed {
			result.Status = "failed"
			result.Output = "execution completed but required side effect was not observed"
			result.Hash = hashBytes([]byte(result.Output))
			g.record(result)
			if g.failClosed {
				return result, naeoserr.New(naeoserr.ErrPipeline, result.Output)
			}
			return result, nil
		}
	}

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
	g.mu.RLock()
	adapterPolicy, granted := g.adapterPolicies[adapterName]
	g.mu.RUnlock()
	if !granted || !adapterPolicy.allows(req) {
		result := ExecutionResult{
			RequestID:    req.RequestID,
			InvocationID: req.InvocationID,
			Request:      req,
			Status:       "denied",
			Output:       "adapter privilege boundary denied request",
			Timestamp:    time.Now().UTC(),
		}
		g.record(result)
		return result, nil
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

func (g *ExecutionGateway) claimInvocation(invocationID string) (bool, error) {
	if g.invocationStore == nil {
		return false, fmt.Errorf("invocation replay store is not configured")
	}
	return g.invocationStore.Claim(invocationID)
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
