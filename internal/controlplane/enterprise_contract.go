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

// ResourceHierarchy binds enterprise resource IDs to their parent boundaries.
type ResourceHierarchy interface {
	TenantBelongsToOrganization(tenantID, organizationID string) bool
	ProjectBelongsToTenant(projectID, tenantID string) bool
	EnvironmentBelongsToProject(environmentID, projectID string) bool
}

func (s EnterpriseScope) ValidateHierarchy(h ResourceHierarchy) error {
	if err := s.Validate(); err != nil { return err }
	if h == nil { return fmt.Errorf("resource hierarchy verifier is required") }
	if !h.TenantBelongsToOrganization(s.TenantID, s.OrganizationID) { return fmt.Errorf("tenant %q does not belong to organization %q", s.TenantID, s.OrganizationID) }
	if !h.ProjectBelongsToTenant(s.ProjectID, s.TenantID) { return fmt.Errorf("project %q does not belong to tenant %q", s.ProjectID, s.TenantID) }
	if !h.EnvironmentBelongsToProject(s.EnvironmentID, s.ProjectID) { return fmt.Errorf("environment %q does not belong to project %q", s.EnvironmentID, s.ProjectID) }
	return nil
}

type ActorIdentity struct {
	Subject string `json:"subject"`
	Type string `json:"type"`
	Issuer string `json:"issuer,omitempty"`
}

type AuthenticatedPrincipal struct {
	Actor ActorIdentity
	Claims map[string]string
}

type EnterpriseAuthorizationRequest struct {
	RequestID string `json:"request_id"`
	Scope EnterpriseScope `json:"scope"`
	Action Action `json:"action"`
	PolicyID string `json:"policy_id"`
	PolicyVersion int `json:"policy_version"`
}

func ResolveAuthorizationRequest(req EnterpriseAuthorizationRequest, principal AuthenticatedPrincipal) (EnterpriseAuthorizationRequest, ActorIdentity, error) {
	if principal.Actor.Subject == "" { return EnterpriseAuthorizationRequest{}, ActorIdentity{}, fmt.Errorf("authenticated principal subject is required") }
	if principal.Actor.Type != "human" && principal.Actor.Type != "service" && principal.Actor.Type != "agent" { return EnterpriseAuthorizationRequest{}, ActorIdentity{}, fmt.Errorf("unsupported authenticated principal type %q", principal.Actor.Type) }
	if err := req.Scope.Validate(); err != nil { return EnterpriseAuthorizationRequest{}, ActorIdentity{}, err }
	return req, principal.Actor, nil
}

type EnterpriseAuthorizationResponse struct {
	RequestID string `json:"request_id"`
	Decision DecisionResult `json:"decision"`
	Scope EnterpriseScope `json:"scope"`
	PolicyID string `json:"policy_id"`
	PolicyVersion int `json:"policy_version"`
}

type ResourceRef struct {
	Type string `json:"type"`
	ID string `json:"id"`
}

type AuditContext struct {
	RequestID string `json:"request_id"`
	Actor ActorIdentity `json:"actor"`
	Reason string `json:"reason,omitempty"`
}
