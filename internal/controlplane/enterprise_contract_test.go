// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "testing"

type testHierarchy struct { tenantOrg, projectTenant, environmentProject bool }

func (h testHierarchy) TenantBelongsToOrganization(_, _ string) bool { return h.tenantOrg }
func (h testHierarchy) ProjectBelongsToTenant(_, _ string) bool { return h.projectTenant }
func (h testHierarchy) EnvironmentBelongsToProject(_, _ string) bool { return h.environmentProject }

func enterpriseTestScope() EnterpriseScope {
	return EnterpriseScope{OrganizationID: "org-1", TenantID: "tenant-1", ProjectID: "project-1", EnvironmentID: "prod"}
}

func TestEnterpriseScopeRequiresAllBoundaries(t *testing.T) {
	valid := enterpriseTestScope()
	if err := valid.Validate(); err != nil { t.Fatalf("valid scope rejected: %v", err) }
	for i, scope := range []EnterpriseScope{
		{TenantID: "tenant-1", ProjectID: "project-1", EnvironmentID: "prod"},
		{OrganizationID: "org-1", ProjectID: "project-1", EnvironmentID: "prod"},
		{OrganizationID: "org-1", TenantID: "tenant-1", EnvironmentID: "prod"},
		{OrganizationID: "org-1", TenantID: "tenant-1", ProjectID: "project-1"},
	} {
		if err := scope.Validate(); err == nil { t.Fatalf("case %d unexpectedly valid", i) }
	}
}

func TestEnterpriseScopeRequiresVerifiedHierarchy(t *testing.T) {
	if err := enterpriseTestScope().ValidateHierarchy(testHierarchy{true, true, true}); err != nil { t.Fatalf("valid hierarchy rejected: %v", err) }
	for i, h := range []testHierarchy{{false, true, true}, {true, false, true}, {true, true, false}} {
		if err := enterpriseTestScope().ValidateHierarchy(h); err == nil { t.Fatalf("hierarchy case %d unexpectedly valid", i) }
	}
}

func TestEnterpriseAuthorizationUsesAuthenticatedPrincipal(t *testing.T) {
	req := EnterpriseAuthorizationRequest{RequestID: "req-1", Scope: enterpriseTestScope(), Action: Action{AgentID: "agent-1", Capability: "repo.write"}, PolicyID: "policy-main", PolicyVersion: 7}
	_, actor, err := ResolveAuthorizationRequest(req, AuthenticatedPrincipal{Actor: ActorIdentity{Subject: "verified-agent", Type: "agent", Issuer: "https://issuer.example"}})
	if err != nil { t.Fatalf("resolve request: %v", err) }
	if actor.Subject != "verified-agent" { t.Fatalf("unexpected resolved actor: %q", actor.Subject) }
}

func TestEnterpriseAuthorizationRejectsMissingPrincipal(t *testing.T) {
	req := EnterpriseAuthorizationRequest{RequestID: "req-1", Scope: enterpriseTestScope(), PolicyID: "p", PolicyVersion: 1}
	if _, _, err := ResolveAuthorizationRequest(req, AuthenticatedPrincipal{}); err == nil { t.Fatal("missing authenticated principal unexpectedly accepted") }
}
