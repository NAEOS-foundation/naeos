// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// This test runs only against the disposable PostgreSQL service provisioned by
// control-plane-postgres-integration.yml. It leaves records in that ephemeral DB.
func TestPostgresRepositoriesIntegrationLifecycleAndSingleUse(t *testing.T) {
	databaseURL := os.Getenv("NAEOS_CONTROLPLANE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set NAEOS_CONTROLPLANE_TEST_DATABASE_URL to an isolated migrated PostgreSQL test database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL database: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping PostgreSQL database: %v", err)
	}

	suffix := controlPlaneID("REPO-TEST")
	policyID := "policy-" + suffix
	agentID := "agent-" + suffix
	now := time.Now().UTC()
	policy := &Policy{
		ID: policyID, Version: 1, Status: "active",
		AllowedCapabilities: []Capability{"repo.read"},
	}
	policies := NewPostgresPolicyRepository(db)
	if err := policies.PutVersion(ctx, policy); err != nil {
		t.Fatalf("PutVersion: %v", err)
	}
	active, err := policies.Active(ctx, policyID)
	if err != nil || active.Version != 1 {
		t.Fatalf("Active = (%+v, %v)", active, err)
	}
	historical, err := policies.Version(ctx, policyID, 1)
	if err != nil || historical.ID != policyID {
		t.Fatalf("Version = (%+v, %v)", historical, err)
	}
	if err := policies.CheckActive(ctx, policyID, 1); err != nil {
		t.Fatalf("CheckActive current version: %v", err)
	}
	if err := policies.CheckActive(ctx, policyID, 2); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("CheckActive stale version error = %v, want conflict", err)
	}

	grant := Grant{
		GrantID: "grant-" + suffix, AgentID: agentID, PolicyID: policyID, PolicyVersion: 1,
		Capabilities: []Capability{"repo.read"}, CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	grants := NewPostgresGrantRepository(db)
	if err := grants.Create(ctx, grant); err != nil {
		t.Fatalf("Create grant: %v", err)
	}
	byID, err := grants.Get(ctx, grant.GrantID)
	if err != nil || byID.AgentID != agentID {
		t.Fatalf("Get grant = (%+v, %v)", byID, err)
	}
	byAgent, err := grants.GetByAgent(ctx, agentID)
	if err != nil || byAgent.GrantID != grant.GrantID {
		t.Fatalf("GetByAgent = (%+v, %v)", byAgent, err)
	}
	if err := grants.Revoke(ctx, grant.GrantID, "integration test revocation", now.Add(time.Minute)); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	revoked, err := grants.Get(ctx, grant.GrantID)
	if err != nil || !revoked.Revoked || revoked.Status != "revoked" {
		t.Fatalf("revoked grant = (%+v, %v)", revoked, err)
	}

	approval := Approval{
		ID: "approval-" + suffix, DecisionID: "decision-" + suffix, AgentID: agentID,
		Capability: "repo.read", PolicyID: policyID, PolicyVersion: 1,
		Approver: "integration-reviewer", Status: "pending", CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}
	approvals := NewPostgresApprovalRepository(db)
	if err := approvals.Create(ctx, approval); err != nil {
		t.Fatalf("Create approval: %v", err)
	}
	if _, err := approvals.Transition(ctx, approval.ID, "pending", "approved", now.Add(time.Second)); err != nil {
		t.Fatalf("approve transition: %v", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, transitionErr := approvals.Transition(ctx, approval.ID, "approved", "consumed", now.Add(2*time.Second))
			results <- transitionErr
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	successes, conflicts := 0, 0
	for transitionErr := range results {
		if transitionErr == nil {
			successes++
		} else if errors.Is(transitionErr, ErrRepositoryConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected consume transition error: %v", transitionErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("consume winners=%d conflicts=%d; expected exactly one atomic winner", successes, conflicts)
	}
	final, err := approvals.Get(ctx, approval.ID)
	if err != nil || final.Status != "consumed" {
		t.Fatalf("final approval = (%+v, %v)", final, err)
	}
}
