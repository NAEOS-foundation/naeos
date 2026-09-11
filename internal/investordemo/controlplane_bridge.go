package investordemo

import (
	"fmt"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

type controlPlaneAuditObserver struct {
	ledger *controlplane.Ledger
}

func (o *controlPlaneAuditObserver) OnRecorded(event *AuditEvent) {
	if o == nil || o.ledger == nil || event == nil {
		return
	}
	decision := controlplane.DecisionStatus("")
	if event.Decision == "ALLOW" {
		decision = controlplane.DecisionAllow
	}
	if event.Decision == "BLOCK" {
		decision = controlplane.DecisionDeny
	}
	o.ledger.Append(controlplane.LedgerEvent{
		Timestamp:  event.Timestamp,
		AgentID:    event.AgentID,
		Capability: controlplane.Capability(event.RequestedCapability),
		EventType:  "LEGACY_" + event.EventType,
		Decision:   decision,
		Reason:     controlplane.DecisionReason(event.Reason),
		Metadata: map[string]string{
			"policy_id":       event.PolicyID,
			"policy_version":  fmt.Sprintf("%d", event.PolicyVersion),
			"source":          "legacy-compatibility",
			"projection":      "true",
			"legacy_event_id": event.EventID,
		},
	})
}

func newControlPlaneGateway(policyStore *PolicyStore, grantStore *GrantStore) *controlplane.DecisionGateway {
	cpStore := controlplane.NewPolicyStore()
	for _, policy := range policyStore.policies {
		_ = cpStore.Set(toControlPlanePolicy(policy))
	}
	for _, grant := range grantStore.grants {
		if policy, err := policyStore.GetPolicy(grant.PolicyID); err == nil {
			_ = cpStore.Set(toControlPlanePolicy(policy))
		}
	}
	return controlplane.NewDecisionGateway(controlplane.NewEvaluator(cpStore), controlplane.NewLedger())
}

func toControlPlanePolicy(policy *Policy) *controlplane.Policy {
	if policy == nil {
		return nil
	}
	return &controlplane.Policy{
		ID:                    policy.PolicyID,
		Version:               policy.Version,
		Status:                policy.Status,
		CreatedAt:             policy.CreatedAt,
		UpdatedAt:             policy.UpdatedAt,
		AllowedCapabilities:   toControlPlaneCapabilities(policy.AllowedCapabilities),
		DeniedCapabilities:    toControlPlaneCapabilities(policy.DeniedCapabilities),
		ProtectedCapabilities: toControlPlaneCapabilities(policy.ProtectedCapabilities),
		ApprovalRequired:      toControlPlaneCapabilities(policy.ApprovalRequired),
		RequiresExplicitAuth:  policy.RequiresExplicitAuth,
	}
}

func toControlPlaneGrant(grant *CapabilityGrant) *controlplane.Grant {
	if grant == nil {
		return nil
	}
	var revokedAt *time.Time
	if grant.RevokedAt != nil {
		revokedAt = grant.RevokedAt
	}
	return &controlplane.Grant{
		GrantID:       grant.GrantID,
		AgentID:       grant.AgentID,
		PolicyID:      grant.PolicyID,
		PolicyVersion: grant.PolicyVersion,
		Capabilities:  toControlPlaneCapabilities(grant.Capabilities),
		CreatedAt:     grant.CreatedAt,
		ExpiresAt:     grant.ExpiresAt,
		Status:        grant.Status,
		Revoked:       grant.Revoked,
		RevokedAt:     revokedAt,
		RevokeReason:  grant.RevokeReason,
	}
}

func toControlPlaneCapabilities(values []Capability) []controlplane.Capability {
	out := make([]controlplane.Capability, 0, len(values))
	for _, value := range values {
		out = append(out, controlplane.Capability(value))
	}
	return out
}

func toDemoCapabilities(values []controlplane.Capability) []Capability {
	out := make([]Capability, 0, len(values))
	for _, value := range values {
		out = append(out, Capability(value))
	}
	return out
}
