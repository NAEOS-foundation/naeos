package policy

import (
	"fmt"
	"strings"
	"sync"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Decision is the authorization result of a control plane evaluation.
type Decision string

const (
	DecisionDeny            Decision = "DENY"
	DecisionAllow           Decision = "ALLOW"
	DecisionRequireApproval Decision = "REQUIRE_APPROVAL"
)

// Valid reports whether d is a recognized decision value.
func (d Decision) Valid() bool {
	switch d {
	case DecisionDeny, DecisionAllow, DecisionRequireApproval:
		return true
	}
	return false
}

// Scope declares which authorization request a policy applies to. An empty
// field acts as a wildcard (matches any value).
type Scope struct {
	Resource    string
	Action      string
	Environment string
}

// PolicyRule is a deterministic guard attached to a policy. When the rule's
// condition passes it contributes its Decision to the policy outcome.
type PolicyRule struct {
	RuleID    string
	Condition string   // evaluated using the evaluator condition syntax
	Decision  Decision // ALLOW / DENY / REQUIRE_APPROVAL
	Priority  int
}

// Policy is a versioned, scope-bounded authorization policy.
type Policy struct {
	ID          string
	Name        string
	Description string
	Version     string
	Category    string
	Scope       Scope
	Default     Decision
	Rules       []PolicyRule
	Active      bool
}

// Registry stores versioned, active policies and answers scope lookups. It is
// safe for concurrent use.
type Registry struct {
	mu       sync.RWMutex
	policies map[string]*Policy   // active policy by ID
	versions map[string][]*Policy // full history by ID (newest first)
}

// NewRegistry returns an empty policy registry.
func NewRegistry() *Registry {
	return &Registry{
		policies: make(map[string]*Policy),
		versions: make(map[string][]*Policy),
	}
}

// Validate performs structural validation on a policy before it is registered.
func Validate(p *Policy) error {
	if p == nil {
		return naeoserr.New(naeoserr.ErrValidation, "policy is nil")
	}
	if strings.TrimSpace(p.ID) == "" {
		return naeoserr.New(naeoserr.ErrValidation, "policy ID is required")
	}
	if strings.TrimSpace(p.Version) == "" {
		return naeoserr.New(naeoserr.ErrValidation, "policy version is required")
	}
	if !p.Default.Valid() {
		return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("invalid default decision %q", p.Default))
	}
	for _, r := range p.Rules {
		if strings.TrimSpace(r.RuleID) == "" {
			return naeoserr.New(naeoserr.ErrValidation, "policy rule ID is required")
		}
		if !r.Decision.Valid() {
			return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("invalid rule decision %q on rule %s", r.Decision, r.RuleID))
		}
	}
	return nil
}

// Register stores a policy. Registering a version that already exists for the
// same ID returns a conflict. The newly registered version becomes active by
// default.
func (r *Registry) Register(p *Policy) error {
	if err := Validate(p); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.versions[p.ID] {
		if existing.Version == p.Version {
			return naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("policy %s version %s already registered", p.ID, p.Version))
		}
	}

	p.Active = true
	r.versions[p.ID] = append([]*Policy{p}, r.versions[p.ID]...)

	if old, ok := r.policies[p.ID]; ok {
		old.Active = false
	}
	r.policies[p.ID] = p
	return nil
}

// GetActive returns the active version of a policy, or nil if none is set.
func (r *Registry) GetActive(id string) (*Policy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.policies[id]
	return p, ok
}

// Get returns a specific version of a policy, or nil if not found.
func (r *Registry) Get(id, version string) (*Policy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.versions[id] {
		if p.Version == version {
			return p, true
		}
	}
	return nil, false
}

// SetActive switches the active version of a policy. It returns an error if
// the requested version does not exist.
func (r *Registry) SetActive(id, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var target *Policy
	for _, p := range r.versions[id] {
		if p.Version == version {
			target = p
			break
		}
	}
	if target == nil {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("policy %s version %s not found", id, version))
	}

	if old, ok := r.policies[id]; ok {
		old.Active = false
	}
	target.Active = true
	r.policies[id] = target
	return nil
}

// Delete removes all versions of a policy.
func (r *Registry) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.versions[id]; !ok {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("policy %s not found", id))
	}
	delete(r.versions, id)
	delete(r.policies, id)
	return nil
}

// List returns the active version of every registered policy (unsorted).
func (r *Registry) List() []*Policy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Policy, 0, len(r.policies))
	for _, p := range r.policies {
		out = append(out, p)
	}
	return out
}

// Versions returns the full version history for a policy (newest first).
func (r *Registry) Versions(id string) []*Policy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.versions[id]
}

// ActiveFor returns the active policies whose scope matches the request
// dimensions. An empty scope field on a policy is treated as a wildcard.
func (r *Registry) ActiveFor(resource, action, environment string) []*Policy {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []*Policy
	for _, p := range r.policies {
		if !p.Active {
			continue
		}
		if p.Scope.Resource != "" && p.Scope.Resource != resource {
			continue
		}
		if p.Scope.Action != "" && p.Scope.Action != action {
			continue
		}
		if p.Scope.Environment != "" && p.Scope.Environment != environment {
			continue
		}
		out = append(out, p)
	}
	return out
}
