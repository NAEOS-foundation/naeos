// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "testing"

func TestEnterpriseScopeRequiresAllBoundaries(t *testing.T) {
	valid := EnterpriseScope{OrganizationID:"org-1", TenantID:"tenant-1", ProjectID:"project-1", EnvironmentID:"prod"}
	if err := valid.Validate(); err != nil { t.Fatalf("valid scope rejected: %v", err) }

	cases := []EnterpriseScope{
		{TenantID:"tenant-1", ProjectID:"project-1", EnvironmentID:"prod"},
		{OrganizationID:"org-1", ProjectID:"project-1", EnvironmentID:"prod"},
		{OrganizationID:"org-1", TenantID:"tenant-1", EnvironmentID:"prod"},
		{OrganizationID:"org-1", TenantID:"tenant-1", ProjectID:"project-1"},
	}
	for i, scope := range cases {
		if err := scope.Validate(); err == nil { t.Fatalf("case %d unexpectedly valid", i) }
	}
}

func TestEnterpriseAuthorizationRequestCarriesPolicyVersion(t *testing.T) {
	req := EnterpriseAuthorizationRequest{
		RequestID:"req-1",
		Scope:EnterpriseScope{OrganizationID:"org-1", TenantID:"tenant-1", ProjectID:"project-1", EnvironmentID:"prod"},
		Actor:ActorIdentity{Subject:"agent-1", Type:"agent"},
		Action:Action{AgentID:"agent-1", Capability:"repo.write"},
		PolicyID:"policy-main",
		PolicyVersion:7,
	}
	if req.PolicyID == "" || req.PolicyVersion != 7 { t.Fatal("policy identity/version not preserved") }
	if err := req.Scope.Validate(); err != nil { t.Fatal(err) }
}
