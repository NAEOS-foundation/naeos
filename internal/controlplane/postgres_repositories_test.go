// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresRepositoriesRejectUnavailableAndCanceledContexts(t *testing.T) {
	ctx := context.Background()
	if err := NewPostgresPolicyRepository(nil).PutVersion(ctx, &Policy{ID: "p", Version: 1}); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("policy unavailable error = %v", err)
	}
	if _, err := NewPostgresPolicyRepository(nil).Active(ctx, "p"); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("policy Active unavailable error = %v", err)
	}
	if err := NewPostgresGrantRepository(nil).Create(ctx, Grant{}); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("grant unavailable error = %v", err)
	}
	if _, err := NewPostgresApprovalRepository(nil).Get(ctx, "a"); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("approval unavailable error = %v", err)
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := NewPostgresPolicyRepository(db).PutVersion(canceled, &Policy{ID: "p", Version: 1}); !errors.Is(err, context.Canceled) {
		t.Fatalf("policy canceled error = %v", err)
	}
	if _, err := NewPostgresGrantRepository(db).Get(canceled, "g"); !errors.Is(err, context.Canceled) {
		t.Fatalf("grant canceled error = %v", err)
	}
	if _, err := NewPostgresApprovalRepository(db).Get(canceled, "a"); !errors.Is(err, context.Canceled) {
		t.Fatalf("approval canceled error = %v", err)
	}
}

func TestPostgresApprovalRepositoryTransitionUsesCompareAndSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)
	repo := NewPostgresApprovalRepository(db)

	mock.ExpectQuery("SELECT expires_at FROM naeos_control_plane.approvals WHERE approval_id = \\$1").
		WithArgs("approval-1").
		WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(expires))
	mock.ExpectExec("UPDATE naeos_control_plane.approvals").
		WithArgs("approval-1", "pending", "approved", now, nil, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT approval_id, decision_id, agent_id, capability, policy_id").
		WithArgs("approval-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"approval_id", "decision_id", "agent_id", "capability", "policy_id", "policy_version",
			"approver", "status", "created_at", "expires_at", "approved_at", "artifact_hash", "reason",
		}).AddRow("approval-1", "decision-1", "agent-1", "repo.read", "policy-1", 1,
			"reviewer", "approved", now.Add(-time.Minute), expires, now, "", ""))

	got, err := repo.Transition(context.Background(), "approval-1", "pending", "approved", now)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if got.Status != "approved" || !got.ApprovedAt.Equal(now) {
		t.Fatalf("Transition returned %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestPostgresGrantRepositoryRejectsExpiredGrantWithDefaultCreatedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	grant := Grant{
		GrantID: "grant-expired", AgentID: "agent-expired", PolicyID: "policy-1",
		PolicyVersion: 1, Capabilities: []Capability{"repo.read"}, Status: "active",
		ExpiresAt: time.Now().UTC().Add(-time.Minute),
	}
	if err := NewPostgresGrantRepository(db).Create(context.Background(), grant); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("Create expired grant error = %v; want ErrRepositoryConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expired grant unexpectedly reached the database: %v", err)
	}
}

func TestPostgresApprovalRepositoryRejectsExpiredApprovalWithDefaultCreatedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	approval := Approval{
		ID: "approval-expired", DecisionID: "decision-1", Approver: "reviewer",
		Status: "pending", ExpiresAt: time.Now().UTC().Add(-time.Minute),
	}
	if err := NewPostgresApprovalRepository(db).Create(context.Background(), approval); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("Create expired approval error = %v; want ErrRepositoryConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expired approval unexpectedly reached the database: %v", err)
	}
}
