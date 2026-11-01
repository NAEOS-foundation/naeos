// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package control

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

// Decision mirrors policy.Decision for ergonomic use by callers of the
// control plane.
type Decision = policy.Decision

const (
	DecisionDeny            = policy.DecisionDeny
	DecisionAllow           = policy.DecisionAllow
	DecisionRequireApproval = policy.DecisionRequireApproval
)

// Request describes a single authorization decision request issued by an
// agent or process.
type Request struct {
	// Capability identifies the exact capability being requested. It is part
	// of the authorization identity so a decision for one capability cannot
	// be reused for a broader capability.
	Capability  string
	Resource    string
	Action      string
	Environment string
	Actor       string
	Context     map[string]any
}

// EvaluationFailure represents an internal inability to determine a policy outcome.
// It is not an authorization decision and must never be treated as DENY or ALLOW.
type EvaluationFailure struct {
	PolicyID      string
	PolicyVersion string
	RuleID        string
	Stage         string
	Code          string
	Timestamp     time.Time
	cause         error
}

func (e *EvaluationFailure) Error() string {
	if e == nil {
		return "policy evaluation failed"
	}
	return "policy evaluation failed"
}

// Unwrap exposes the evaluator cause to trusted internal callers. Callers must
// avoid returning the cause directly to agents or external API clients.
func (e *EvaluationFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// DecisionRecord is durable evidence of a single authorization decision.
type DecisionRecord struct {
	Request       Request
	Decision      Decision
	PolicyID      string
	PolicyVersion string
	RuleID        string
	Reasons       []string
	Timestamp     time.Time
	Deterministic bool
	// ExpiresAt bounds how long the issued authority remains usable. A zero
	// value means the control plane has no configured grant TTL.
	ExpiresAt time.Time
	// GrantDigest binds the exact authorization identity to this decision.
	GrantDigest string
}

// ControlPlane evaluates authorization requests against registered policies
// and issues deterministic decisions. It is the architectural heart of NAEOS
// governance: given the same policy, action, and context it always produces
// the same decision, with no LLM inference required.
type ControlPlane struct {
	registry  *policy.Registry
	evaluator policy.Evaluator
	grantTTL  time.Duration

	mu            sync.Mutex
	decisions     []DecisionRecord
	revokedGrants map[string]struct{}
}

// Option configures a ControlPlane.
type Option func(*ControlPlane)

// New creates a ControlPlane over the given policy registry.
func New(reg *policy.Registry, opts ...Option) *ControlPlane {
	c := &ControlPlane{
		registry:      reg,
		evaluator:     policy.NewEvaluator(),
		revokedGrants: make(map[string]struct{}),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// WithGrantTTL configures a finite lifetime for newly issued authorization
// grants. A non-positive TTL leaves grants without an expiry.
func WithGrantTTL(ttl time.Duration) Option {
	return func(c *ControlPlane) {
		if ttl > 0 {
			c.grantTTL = ttl
		}
	}
}

// Evaluate issues a deterministic decision for a request. It returns the
// decision record and, on error, a non-decision error (e.g. invalid request).
func (c *ControlPlane) Evaluate(req Request) (DecisionRecord, error) {
	if strings.TrimSpace(req.Resource) == "" || strings.TrimSpace(req.Action) == "" {
		return DecisionRecord{}, naeoserr.New(naeoserr.ErrValidation, "request resource and action are required")
	}
	if req.Context == nil {
		req.Context = map[string]any{}
	}

	policies := c.registry.ActiveFor(req.Resource, req.Action, req.Environment)

	if len(policies) == 0 {
		rec := DecisionRecord{
			Request:       req,
			Decision:      DecisionDeny,
			Reasons:       []string{"no matching policy"},
			Timestamp:     time.Now().UTC(),
			Deterministic: true,
		}
		if c.grantTTL > 0 {
			rec.ExpiresAt = rec.Timestamp.Add(c.grantTTL)
		}
		if err := bindGrantDigest(&rec); err != nil {
			return DecisionRecord{}, naeoserr.Wrapf(err, naeoserr.ErrInternal, "failed to bind authorization grant")
		}
		c.record(rec)
		return rec, nil
	}

	// Aggregate over all matching policies. Deny always wins; approval outranks
	// allow. Ties are broken by strictest decision regardless of policy order.
	now := time.Now().UTC()
	worstRec := DecisionRecord{Request: req, Deterministic: true, Timestamp: now}
	if c.grantTTL > 0 {
		worstRec.ExpiresAt = now.Add(c.grantTTL)
	}
	for _, pol := range policies {
		outcome, rec, evalErr := c.evaluatePolicy(pol, req)
		if evalErr != nil {
			// An evaluator failure is an internal indeterminate state, not a
			// policy decision. Do not mint a grant digest or append a decision.
			return DecisionRecord{}, &EvaluationFailure{
				PolicyID:      pol.ID,
				PolicyVersion: pol.Version,
				RuleID:        rec.RuleID,
				Stage:         "rule_evaluation",
				Code:          "policy_evaluator_error",
				Timestamp:     time.Now().UTC(),
				cause:         evalErr,
			}
		}
		rec.Request = req
		rec.Timestamp = time.Now().UTC()
		rec.Deterministic = true
		worstRec = stricter(worstRec, rec, outcome)
	}
	// The strictest policy record may replace worstRec, so restore the
	// configured grant lifetime before binding the final authorization identity.
	if c.grantTTL > 0 {
		worstRec.ExpiresAt = now.Add(c.grantTTL)
	}
	if err := bindGrantDigest(&worstRec); err != nil {
		return DecisionRecord{}, naeoserr.Wrapf(err, naeoserr.ErrInternal, "failed to bind authorization grant")
	}

	c.record(worstRec)
	return worstRec, nil
}

// evaluatePolicy evaluates every rule in deterministic priority order. Priority
// controls evaluation order and failure attribution only; it never overrides
// decision severity. A failed rule remains a fail-closed DENY. When all rules
// pass, the strictest of the policy default and all rule decisions wins:
// DENY > REQUIRE_APPROVAL > ALLOW.
func (c *ControlPlane) evaluatePolicy(pol *policy.Policy, req Request) (policy.Decision, DecisionRecord, error) {
	rec := DecisionRecord{
		PolicyID:      pol.ID,
		PolicyVersion: pol.Version,
	}

	ctx := make(map[string]any, len(req.Context)+4)
	for k, v := range req.Context {
		ctx[k] = v
	}
	ctx["resource"] = req.Resource
	ctx["action"] = req.Action
	ctx["environment"] = req.Environment
	ctx["actor"] = req.Actor

	rules := append([]policy.PolicyRule(nil), pol.Rules...)
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority < rules[j].Priority
		}
		if rules[i].RuleID != rules[j].RuleID {
			return rules[i].RuleID < rules[j].RuleID
		}
		if rules[i].Condition != rules[j].Condition {
			return rules[i].Condition < rules[j].Condition
		}
		return rules[i].Decision < rules[j].Decision
	})

	type scored struct {
		ruleID  string
		dec     policy.Decision
		passed  bool
		message string
	}

	var results []scored
	for _, r := range rules {
		er, err := c.evaluator.EvaluateRules([]policy.Rule{{
			RuleID:    r.RuleID,
			Condition: r.Condition,
			Priority:  r.Priority,
			Action:    string(r.Decision),
			Enabled:   true,
		}}, ctx)
		if err != nil {
			rec.RuleID = r.RuleID
			rec.Reasons = []string{fmt.Sprintf("policy %s v%s evaluation failed", pol.ID, pol.Version)}
			return "", rec, err
		}
		if len(er) == 0 {
			continue
		}
		results = append(results, scored{
			ruleID:  r.RuleID,
			dec:     r.Decision,
			passed:  er[0].Passed,
			message: er[0].Message,
		})
	}

	rec.Reasons = []string{fmt.Sprintf("policy %s v%s matched", pol.ID, pol.Version)}

	// A failed rule denies regardless of its configured action. The first
	// failure is stable because rules were evaluated by priority then RuleID.
	for _, result := range results {
		if !result.passed {
			rec.Decision = DecisionDeny
			rec.RuleID = result.ruleID
			rec.Reasons = append(rec.Reasons, fmt.Sprintf("rule %s failed: %s", result.ruleID, result.message))
			return DecisionDeny, rec, nil
		}
	}

	// Priority does not resolve conflicts. Aggregate all passing rule outcomes
	// with the policy default using the global decision-severity ordering.
	selected := pol.Default
	selectedRuleID := ""
	for _, result := range results {
		if rank(result.dec) > rank(selected) {
			selected = result.dec
			selectedRuleID = result.ruleID
		}
	}
	rec.Decision = selected
	rec.RuleID = selectedRuleID
	if selectedRuleID == "" {
		rec.Reasons = append(rec.Reasons, fmt.Sprintf("policy default %s is the strictest outcome", selected))
	} else {
		rec.Reasons = append(rec.Reasons, fmt.Sprintf("rule %s contributes strictest outcome %s", selectedRuleID, selected))
	}
	return selected, rec, nil
}

// stricter returns the higher-severity decision. Equal-severity outcomes use
// a stable policy/version/rule identity tie-break, independent of registry
// iteration or policy registration order. Order: DENY > REQUIRE_APPROVAL > ALLOW.
func stricter(a, b DecisionRecord, bDec policy.Decision) DecisionRecord {
	aRank := rank(a.Decision)
	bRank := rank(bDec)
	if bRank > aRank || (bRank == aRank && decisionSourceLess(b, a)) {
		b.Decision = bDec
		return b
	}
	return a
}

func decisionSourceLess(a, b DecisionRecord) bool {
	if a.PolicyID != b.PolicyID {
		return a.PolicyID < b.PolicyID
	}
	if a.PolicyVersion != b.PolicyVersion {
		return a.PolicyVersion < b.PolicyVersion
	}
	return a.RuleID < b.RuleID
}

func rank(d Decision) int {
	switch d {
	case DecisionDeny:
		return 3
	case DecisionRequireApproval:
		return 2
	case DecisionAllow:
		return 1
	default:
		return 0
	}
}

func (c *ControlPlane) record(rec DecisionRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.decisions = append(c.decisions, rec)
}

// RevokeGrant permanently invalidates a previously issued grant.
func (c *ControlPlane) RevokeGrant(grantDigest string) error {
	if strings.TrimSpace(grantDigest) == "" {
		return naeoserr.New(naeoserr.ErrValidation, "grant digest is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.revokedGrants[grantDigest] = struct{}{}
	return nil
}

func (c *ControlPlane) grantRevoked(grantDigest string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, revoked := c.revokedGrants[grantDigest]
	return revoked
}

// ValidateDecision re-evaluates a decision immediately before execution and
// rejects it if the grant is expired/revoked or the active policy identity,
// version, rule, or decision changed.
func (c *ControlPlane) ValidateDecision(req Request, issued DecisionRecord) (DecisionRecord, error) {
	if issued.GrantDigest == "" || !grantDigestMatches(issued) {
		return issued, naeoserr.New(naeoserr.ErrConflict, "authorization grant integrity check failed")
	}
	if c.grantRevoked(issued.GrantDigest) {
		return issued, naeoserr.New(naeoserr.ErrConflict, "authorization grant revoked")
	}
	if !issued.ExpiresAt.IsZero() && !time.Now().UTC().Before(issued.ExpiresAt) {
		return issued, naeoserr.New(naeoserr.ErrConflict, "authorization grant expired")
	}

	current, err := c.Evaluate(req)
	if err != nil {
		return DecisionRecord{}, err
	}
	// Reuse the issued expiry for the current policy identity. A fresh
	// Evaluate call must not mint a new lifetime during revalidation.
	current.ExpiresAt = issued.ExpiresAt
	if err := bindGrantDigest(&current); err != nil {
		return DecisionRecord{}, naeoserr.Wrapf(err, naeoserr.ErrInternal, "failed to rebind authorization grant")
	}
	if current.GrantDigest != issued.GrantDigest ||
		current.Request.Capability != issued.Request.Capability ||
		current.Request.Resource != issued.Request.Resource ||
		current.Request.Action != issued.Request.Action ||
		current.Request.Environment != issued.Request.Environment ||
		current.Request.Actor != issued.Request.Actor ||
		current.PolicyID != issued.PolicyID ||
		current.PolicyVersion != issued.PolicyVersion ||
		current.RuleID != issued.RuleID ||
		current.Decision != issued.Decision {
		return current, naeoserr.New(naeoserr.ErrConflict, "authorization invalidated by policy mutation")
	}
	return current, nil
}

// ListDecisions returns the decision records issued by this plane so far.
func (c *ControlPlane) ListDecisions() []DecisionRecord {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]DecisionRecord, len(c.decisions))
	copy(out, c.decisions)
	return out
}
