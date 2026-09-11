package investordemo

import "time"

// This file defines storage seams for the investor demo. The demo uses
// in-memory implementations (GrantStore, AuditLedger, PolicyStore) that satisfy
// these interfaces, so a persistent or distributed implementation can be
// substituted later without changing the control-plane logic.
//
// The components accept the concrete types today for demo simplicity; the
// interfaces document and enforce the replaceable boundary.

// GrantRepository stores and retrieves capability grants.
type GrantRepository interface {
	StoreGrant(grant *CapabilityGrant) error
	GetGrant(grantID string) (*CapabilityGrant, error)
	GetGrantByAgent(agentID string) (*CapabilityGrant, error)
	RevokeGrant(grantID string, reason string, revokedAt *time.Time) error
	ListGrants() []*CapabilityGrant
}

// AuditEventStore is an append-only store for security audit events.
type AuditEventStore interface {
	RecordEvent(event *AuditEvent) error
	GetEvents() []*AuditEvent
	GetEventsByAgent(agentID string) []*AuditEvent
	GetEventsByType(eventType string) []*AuditEvent
	CountEventsByType(eventType string) int
}

// PolicyRepository stores versioned policies.
type PolicyRepository interface {
	StorePolicy(policy *Policy) error
	GetPolicy(policyID string) (*Policy, error)
	GetPolicyVersion(policyID string, version int) (*Policy, error)
	UpdatePolicy(policy *Policy) error
}

// Compile-time assertions that the in-memory implementations satisfy the seams.
var (
	_ GrantRepository  = (*GrantStore)(nil)
	_ AuditEventStore  = (*AuditLedger)(nil)
	_ PolicyRepository = (*PolicyStore)(nil)
)
