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
	if s.OrganizationID == "" {
		return fmt.Errorf("organization_id is required")
	}
	if s.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if s.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if s.EnvironmentID == "" {
		return fmt.Errorf("environment_id is required")
	}
	return nil
}

// ResourceHierarchy binds enterprise resource IDs to their parent boundaries.
// The authorization layer must verify these relationships against its resource store.
type ResourceHierarchy interface {
	TenantBelongsToOrganization(tenantID, organizationID string) bool
	ProjectBelongsToTenant(projectID, tenantID string) bool
	EnvironmentBelongsToProject(environmentID, projectID string) bool
}

func (s EnterpriseScope) ValidateHierarchy(h ResourceHierarchy) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if h == nil {
		return fmt.Errorf("resource hierarchy verifier is required")
	}
	if !h.TenantBelongsToOrganization(s.TenantID, s.OrganizationID) {
		return fmt.Errorf("tenant %q does not belong to organization %q", s.TenantID, s.OrganizationID)
	}
	if !h.ProjectBelongsToTenant(s.ProjectID, s.TenantID) {
		return fmt.Errorf("project %q does not belong to tenant %q", s.ProjectID, s.TenantID)
	}
	if !h.EnvironmentBelongsToProject(s.EnvironmentID, s.ProjectID) {
		return fmt.Errorf("environment %q does not belong to project %q", s.EnvironmentID, s.ProjectID)
	}
	return nil
}

// ActorIdentity is a verified principal resolved by the server from authenticated credentials.
// It must not be trusted solely because the same fields were supplied by a client.
type ActorIdentity struct {
	Subject string `json:"subject"`
	Type    string `json:"type"` // human, service, agent
	Issuer  string `json:"issuer,omitempty"`
}

// AuthenticatedPrincipal is the server-side identity context used for authorization.
type AuthenticatedPrincipal struct {
	Actor  ActorIdentity
	Claims map[string]string
}

// EnterpriseAuthorizationRequest is the stable semantic input for authorization.
// Actor is intentionally not client-supplied; it is resolved from AuthenticatedPrincipal.
type EnterpriseAuthorizationRequest struct {
	RequestID     string          `json:"request_id"`
	Scope         EnterpriseScope `json:"scope"`
	Action        Action          `json:"action"`
	PolicyID      string          `json:"policy_id"`
	PolicyVersion int             `json:"policy_version"`
}

// ResolveAuthorizationRequest binds a request to a verified principal.
func ResolveAuthorizationRequest(req EnterpriseAuthorizationRequest, principal AuthenticatedPrincipal) (EnterpriseAuthorizationRequest, ActorIdentity, error) {
	if principal.Actor.Subject == "" {
		return EnterpriseAuthorizationRequest{}, ActorIdentity{}, fmt.Errorf("authenticated principal subject is required")
	}
	if principal.Actor.Type != "human" && principal.Actor.Type != "service" && principal.Actor.Type != "agent" {
		return EnterpriseAuthorizationRequest{}, ActorIdentity{}, fmt.Errorf("unsupported authenticated principal type %q", principal.Actor.Type)
	}
	if err := req.Scope.Validate(); err != nil {
		return EnterpriseAuthorizationRequest{}, ActorIdentity{}, err
	}
	return req, principal.Actor, nil
}

// EnterpriseAuthorizationResponse preserves deterministic authorization semantics.
type EnterpriseAuthorizationResponse struct {
	RequestID     string          `json:"request_id"`
	Decision      DecisionResult  `json:"decision"`
	Scope         EnterpriseScope `json:"scope"`
	PolicyID      string          `json:"policy_id"`
	PolicyVersion int             `json:"policy_version"`
}

// ResourceRef is a stable cross-system resource identity.
type ResourceRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// AuditContext binds administrative mutations to a verified enterprise principal and request.
type AuditContext struct {
	RequestID string       `json:"request_id"`
	Actor     ActorIdentity `json:"actor"`
	Reason    string       `json:"reason,omitempty"`
}
