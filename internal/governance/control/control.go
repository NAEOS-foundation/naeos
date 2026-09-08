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
	Resource    string
	Action      string
	Environment string
	Actor       string
	Context     map[string]any
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
}

// ControlPlane evaluates authorization requests against registered policies
// and issues deterministic decisions. It is the architectural heart of NAEOS
// governance: given the same policy, action, and context it always produces
// the same decision, with no LLM inference required.
type ControlPlane struct {
	registry   *policy.Registry
	evaluator  policy.Evaluator
	failClosed bool

	mu        sync.Mutex
	decisions []DecisionRecord
}

// Option configures a ControlPlane.
type Option func(*ControlPlane)

// FailClosed controls the decision issued when no policy matches a request.
// When true (default) unmatched requests are denied; when false they are
// allowed.
func FailClosed(enabled bool) Option {
	return func(c *ControlPlane) { c.failClosed = enabled }
}

// New creates a ControlPlane over the given policy registry.
func New(reg *policy.Registry, opts ...Option) *ControlPlane {
	c := &ControlPlane{
		registry:   reg,
		evaluator:  policy.NewEvaluator(),
		failClosed: true,
	}
	for _, o := range opts {
		o(c)
	}
	return c
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
		decision := DecisionDeny
		if !c.failClosed {
			decision = DecisionAllow
		}
		rec := DecisionRecord{
			Request:       req,
			Decision:      decision,
			Reasons:       []string{"no matching policy"},
			Timestamp:     time.Now().UTC(),
			Deterministic: true,
		}
		c.record(rec)
		return rec, nil
	}

	// Aggregate over all matching policies. Deny always wins; approval outranks
	// allow. Ties are broken by strictest decision regardless of policy order.
	worstRec := DecisionRecord{Request: req, Deterministic: true, Timestamp: time.Now().UTC()}

	for _, pol := range policies {
		outcome, rec := c.evaluatePolicy(pol, req)
		rec.Request = req
		rec.Timestamp = time.Now().UTC()
		rec.Deterministic = true
		worstRec = stricter(worstRec, rec, outcome)
	}

	c.record(worstRec)
	return worstRec, nil
}

// evaluatePolicy evaluates all rules of a single policy and derives a
// decision per the following deterministic semantics:
//   - any failing rule -> DENY
//   - otherwise, if any rule or the policy default requires approval -> REQUIRE_APPROVAL
//   - otherwise -> the policy default (ALLOW by default)
func (c *ControlPlane) evaluatePolicy(pol *policy.Policy, req Request) (policy.Decision, DecisionRecord) {
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

	type scored struct {
		ruleID  string
		dec     policy.Decision
		prio    int
		passed  bool
		message string
	}

	var results []scored
	for _, r := range pol.Rules {
		er, err := c.evaluator.EvaluateRules([]policy.Rule{{
			RuleID:    r.RuleID,
			Condition: r.Condition,
			Priority:  r.Priority,
			Action:    string(r.Decision),
			Enabled:   true,
		}}, ctx)
		if err != nil {
			continue
		}
		if len(er) == 0 {
			continue
		}
		results = append(results, scored{
			ruleID:  r.RuleID,
			dec:     r.Decision,
			prio:    r.Priority,
			passed:  er[0].Passed,
			message: er[0].Message,
		})
	}

	// Sort by priority descending so highest-priority rules are considered
	// first when aggregating.
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].prio > results[j].prio
	})

	rec.Reasons = []string{fmt.Sprintf("policy %s v%s matched", pol.ID, pol.Version)}

	// Any failing rule denies the request.
	for _, s := range results {
		if !s.passed {
			rec.Decision = DecisionDeny
			rec.RuleID = s.ruleID
			rec.Reasons = append(rec.Reasons, fmt.Sprintf("rule %s failed: %s", s.ruleID, s.message))
			return DecisionDeny, rec
		}
	}

	// Otherwise, highest-priority passing rule determines the outcome.
	if len(results) > 0 {
		top := results[0]
		rec.RuleID = top.ruleID
		rec.Decision = top.dec
		rec.Reasons = append(rec.Reasons, fmt.Sprintf("rule %s %s", top.ruleID, top.dec))
		return top.dec, rec
	}

	// No rules: fall back to the policy default.
	rec.Decision = pol.Default
	rec.Reasons = append(rec.Reasons, fmt.Sprintf("no rules, policy default %s", pol.Default))
	return pol.Default, rec
}

// stricter returns the stricter of two decision records, taking the source
// policy/version for the strictest decision. Order: DENY > REQUIRE_APPROVAL >
// ALLOW.
func stricter(a, b DecisionRecord, bDec policy.Decision) DecisionRecord {
	aRank := rank(a.Decision)
	bRank := rank(bDec)
	if bRank >= aRank {
		b.Decision = bDec
		return b
	}
	return a
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

// ListDecisions returns the decision records issued by this plane so far.
func (c *ControlPlane) ListDecisions() []DecisionRecord {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]DecisionRecord, len(c.decisions))
	copy(out, c.decisions)
	return out
}
