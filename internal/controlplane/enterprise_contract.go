// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "fmt"

// EnterpriseScope is the minimum scope attached to enterprise control-plane operations.
type EnterpriseScope struct {
	OrganizationID string `json:"organization_id"`
	TenantID       string `json:"tenant_id"`
	ProjectID      string `json:"project_id"`
	EnvironmentID  string `json:"environment_id"`
}

func (s EnterpriseScope) Validate() error {
	if s.OrganizationID == "" { return fmt.Errorf("organization_id is required") }
	if s.TenantID == "" { return fmt.Errorf("tenant_id is required") }
	if s.ProjectID == "" { return fmt.Errorf("project_id is required") }
	if s.EnvironmentID == "" { return fmt.Errorf("environment_id is required") }
	return nil
}

// ActorIdentity identifies the principal making or delegating an operation.
type ActorIdentity struct {
	Subject string `json:"subject"`
	Type    string `json:"type"` // human, service, agent
	Issuer  string `json:"issuer,omitempty"`
}

// EnterpriseAuthorizationRequest is the stable semantic input for authorization.
type EnterpriseAuthorizationRequest struct {
	RequestID string `json:"request_id"`
	Scope     EnterpriseScope `json:"scope"`
	Actor     ActorIdentity `json:"actor"`
	Action    Action `json:"action"`
	PolicyID  string `json:"policy_id"`
	PolicyVersion int `json:"policy_version"`
}

// EnterpriseAuthorizationResponse preserves deterministic authorization semantics.
type EnterpriseAuthorizationResponse struct {
	RequestID string `json:"request_id"`
	Decision  DecisionResult `json:"decision"`
	Scope     EnterpriseScope `json:"scope"`
	PolicyID  string `json:"policy_id"`
	PolicyVersion int `json:"policy_version"`
}

// ResourceRef is a stable cross-system resource identity.
type ResourceRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// AuditContext binds administrative mutations to an enterprise principal and request.
type AuditContext struct {
	RequestID string `json:"request_id"`
	Actor     ActorIdentity `json:"actor"`
	Reason    string `json:"reason,omitempty"`
}
