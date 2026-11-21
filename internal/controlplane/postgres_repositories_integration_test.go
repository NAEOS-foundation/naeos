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

	"github.com/jackc/pgx/v5/pgxpool"
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

	// Production startup must reject admin/BYPASSRLS credentials and require
	// the dedicated least-privilege login principal.
	runtimePool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create runtime role verification pool: %v", err)
	}
	if err := verifyControlPlaneRuntimeRole(ctx, runtimePool); err != nil {
		runtimePool.Close()
		t.Fatalf("dedicated runtime login was rejected: %v", err)
	}
	runtimePool.Close()

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

	// Verify the production execution claim consumes approval and appends the
	// durable intent in one transaction, including rollback when the ledger rejects it.
	claimPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create atomic execution claim pool: %v", err)
	}
	defer claimPool.Close()
	postgresLedger, err := NewPostgresLedger(claimPool)
	if err != nil {
		t.Fatalf("create atomic execution ledger: %v", err)
	}
	ledgerRepository, err := NewPostgresLedgerRepository(postgresLedger)
	if err != nil {
		t.Fatalf("create atomic execution ledger repository: %v", err)
	}

	atomicAgentID := "agent-atomic-" + suffix
	atomicGrant := Grant{
		GrantID: "grant-atomic-" + suffix, AgentID: atomicAgentID, PolicyID: policyID, PolicyVersion: 1,
		Capabilities: []Capability{"repo.read"}, CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	if err := grants.Create(ctx, atomicGrant); err != nil {
		t.Fatalf("create active atomic grant: %v", err)
	}
	atomicApproval := Approval{
		ID: "approval-atomic-" + suffix, DecisionID: "decision-atomic-" + suffix,
		AgentID: atomicAgentID, Capability: "repo.read", PolicyID: policyID, PolicyVersion: 1,
		Approver: "integration-reviewer", Status: "pending", CreatedAt: now,
		ExpiresAt: now.Add(time.Hour), ArtifactHash: "sha256:atomic-" + suffix,
	}
	if err := approvals.Create(ctx, atomicApproval); err != nil {
		t.Fatalf("create atomic approval: %v", err)
	}
	if _, err := approvals.Transition(ctx, atomicApproval.ID, "pending", "approved", now.Add(3*time.Second)); err != nil {
		t.Fatalf("approve atomic approval: %v", err)
	}
	startClaims := make(chan struct{})
	claimResults := make(chan error, 2)
	var claimWG sync.WaitGroup
	for i := 0; i < 2; i++ {
		claimWG.Add(1)
		go func(index int) {
			defer claimWG.Done()
			<-startClaims
			intent := LedgerEvent{
				Timestamp: now.Add(4 * time.Second), RequestID: "request-" + suffix,
				DecisionID: atomicApproval.DecisionID, ExecutionID: controlPlaneID("EXEC"),
				AgentID: atomicAgentID, Capability: "repo.read", ArtifactHash: atomicApproval.ArtifactHash,
				EventType: "EXECUTION_STARTED", Decision: DecisionAllow,
				Metadata: map[string]string{"approval_id": atomicApproval.ID, "policy_id": policyID, "policy_version": "1", "grant_id": atomicGrant.GrantID},
			}
			_, _, claimErr := ledgerRepository.ClaimExecution(ctx, intent, atomicApproval, now.Add(4*time.Second))
			claimResults <- claimErr
		}(i)
	}
	close(startClaims)
	claimWG.Wait()
	close(claimResults)
	claimSuccesses, claimConflicts := 0, 0
	for claimErr := range claimResults {
		if claimErr == nil {
			claimSuccesses++
		} else if errors.Is(claimErr, ErrRepositoryConflict) {
			claimConflicts++
		} else {
			t.Fatalf("unexpected atomic claim error: %v", claimErr)
		}
	}
	if claimSuccesses != 1 || claimConflicts != 1 {
		t.Fatalf("atomic claim winners=%d conflicts=%d; want exactly one of each", claimSuccesses, claimConflicts)
	}
	consumedApproval, err := approvals.Get(ctx, atomicApproval.ID)
	if err != nil || consumedApproval.Status != "consumed" {
		t.Fatalf("atomic approval state = (%+v, %v); want consumed", consumedApproval, err)
	}

	rollbackApproval := Approval{
		ID: "approval-rollback-" + suffix, DecisionID: atomicApproval.DecisionID,
		AgentID: atomicAgentID, Capability: "repo.read", PolicyID: policyID, PolicyVersion: 1,
		Approver: "integration-reviewer", Status: "pending", CreatedAt: now,
		ExpiresAt: now.Add(time.Hour), ArtifactHash: atomicApproval.ArtifactHash,
	}
	if err := approvals.Create(ctx, rollbackApproval); err != nil {
		t.Fatalf("create rollback approval: %v", err)
	}
	if _, err := approvals.Transition(ctx, rollbackApproval.ID, "pending", "approved", now.Add(5*time.Second)); err != nil {
		t.Fatalf("approve rollback approval: %v", err)
	}
	rollbackIntent := LedgerEvent{
		Timestamp: now.Add(6 * time.Second), RequestID: "request-" + suffix,
		DecisionID: atomicApproval.DecisionID, ExecutionID: controlPlaneID("EXEC"),
		AgentID: atomicAgentID, Capability: "repo.read", ArtifactHash: atomicApproval.ArtifactHash,
		EventType: "EXECUTION_STARTED", Decision: DecisionAllow,
		Metadata: map[string]string{"approval_id": rollbackApproval.ID, "policy_id": policyID, "policy_version": "1", "grant_id": atomicGrant.GrantID},
	}
	if _, _, err := ledgerRepository.ClaimExecution(ctx, rollbackIntent, rollbackApproval, now.Add(6*time.Second)); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("duplicate execution claim error = %v; want conflict", err)
	}
	unchangedApproval, err := approvals.Get(ctx, rollbackApproval.ID)
	if err != nil || unchangedApproval.Status != "approved" {
		t.Fatalf("approval after rolled-back claim = (%+v, %v); want still approved", unchangedApproval, err)
	}

	// A revocation that commits before the claim must fence execution.
	if err := grants.Revoke(ctx, atomicGrant.GrantID, "integration test revocation before claim", now.Add(7*time.Second)); err != nil {
		t.Fatalf("revoke atomic grant before claim: %v", err)
	}
	revokedApproval := Approval{
		ID: "approval-revoked-" + suffix, DecisionID: "decision-revoked-" + suffix,
		AgentID: atomicAgentID, Capability: "repo.read", PolicyID: policyID, PolicyVersion: 1,
		Approver: "integration-reviewer", Status: "pending", CreatedAt: now,
		ExpiresAt: now.Add(time.Hour), ArtifactHash: "sha256:revoked-" + suffix,
	}
	if err := approvals.Create(ctx, revokedApproval); err != nil {
		t.Fatalf("create revoked-grant approval: %v", err)
	}
	if _, err := approvals.Transition(ctx, revokedApproval.ID, "pending", "approved", now.Add(8*time.Second)); err != nil {
		t.Fatalf("approve revoked-grant approval: %v", err)
	}
	revokedIntent := LedgerEvent{
		Timestamp: now.Add(9 * time.Second), RequestID: "request-" + suffix,
		DecisionID: revokedApproval.DecisionID, ExecutionID: controlPlaneID("EXEC"),
		AgentID: atomicAgentID, Capability: "repo.read", ArtifactHash: revokedApproval.ArtifactHash,
		EventType: "EXECUTION_STARTED", Decision: DecisionAllow,
		Metadata: map[string]string{"approval_id": revokedApproval.ID, "policy_id": policyID, "policy_version": "1", "grant_id": atomicGrant.GrantID},
	}
	if _, _, err := ledgerRepository.ClaimExecution(ctx, revokedIntent, revokedApproval, now.Add(9*time.Second)); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("claim with revoked grant error = %v; want conflict", err)
	}
	stillApproved, err := approvals.Get(ctx, revokedApproval.ID)
	if err != nil || stillApproved.Status != "approved" {
		t.Fatalf("approval after revoked-grant claim = (%+v, %v); want approved", stillApproved, err)
	}

	// Advancing the active policy version before claim must reject the stale decision.
	staleAgentID := "agent-stale-policy-" + suffix
	staleGrant := Grant{
		GrantID: "grant-stale-policy-" + suffix, AgentID: staleAgentID, PolicyID: policyID, PolicyVersion: 1,
		Capabilities: []Capability{"repo.read"}, CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	if err := grants.Create(ctx, staleGrant); err != nil {
		t.Fatalf("create stale-policy grant: %v", err)
	}
	staleApproval := Approval{
		ID: "approval-stale-policy-" + suffix, DecisionID: "decision-stale-policy-" + suffix,
		AgentID: staleAgentID, Capability: "repo.read", PolicyID: policyID, PolicyVersion: 1,
		Approver: "integration-reviewer", Status: "pending", CreatedAt: now,
		ExpiresAt: now.Add(time.Hour), ArtifactHash: "sha256:stale-policy-" + suffix,
	}
	if err := approvals.Create(ctx, staleApproval); err != nil {
		t.Fatalf("create stale-policy approval: %v", err)
	}
	if _, err := approvals.Transition(ctx, staleApproval.ID, "pending", "approved", now.Add(10*time.Second)); err != nil {
		t.Fatalf("approve stale-policy approval: %v", err)
	}
	policyV2 := clonePolicy(policy)
	policyV2.Version = 2
	policyV2.CreatedAt = now.Add(11 * time.Second)
	policyV2.UpdatedAt = now.Add(11 * time.Second)
	if err := policies.PutVersion(ctx, policyV2); err != nil {
		t.Fatalf("activate policy version 2: %v", err)
	}
	staleIntent := LedgerEvent{
		Timestamp: now.Add(12 * time.Second), RequestID: "request-" + suffix,
		DecisionID: staleApproval.DecisionID, ExecutionID: controlPlaneID("EXEC"),
		AgentID: staleAgentID, Capability: "repo.read", ArtifactHash: staleApproval.ArtifactHash,
		EventType: "EXECUTION_STARTED", Decision: DecisionAllow,
		Metadata: map[string]string{"approval_id": staleApproval.ID, "policy_id": policyID, "policy_version": "1", "grant_id": staleGrant.GrantID},
	}
	if _, _, err := ledgerRepository.ClaimExecution(ctx, staleIntent, staleApproval, now.Add(12*time.Second)); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("claim with stale policy error = %v; want conflict", err)
	}
	staleStillApproved, err := approvals.Get(ctx, staleApproval.ID)
	if err != nil || staleStillApproved.Status != "approved" {
		t.Fatalf("approval after stale-policy claim = (%+v, %v); want approved", staleStillApproved, err)
	}

	adminURL := os.Getenv("NAEOS_CONTROLPLANE_ADMIN_TEST_DATABASE_URL")
	if adminURL == "" {
		t.Skip("set NAEOS_CONTROLPLANE_ADMIN_TEST_DATABASE_URL to verify admin-role rejection")
	}
	adminPool, err := pgxpool.New(ctx, adminURL)
	if err != nil {
		t.Fatalf("create isolated admin test pool: %v", err)
	}
	defer adminPool.Close()
	if err := adminPool.Ping(ctx); err != nil {
		t.Fatalf("ping isolated admin test database: %v", err)
	}
	if err := verifyControlPlaneRuntimeRole(ctx, adminPool); err == nil {
		t.Fatal("production runtime role verification accepted the admin database principal")
	}
}
