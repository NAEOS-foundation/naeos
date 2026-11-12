// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresPolicyRepository persists immutable policy snapshots and their active
// version pointer. A transaction-scoped advisory lock serializes writers per policy.
type PostgresPolicyRepository struct{ db *sql.DB }

func NewPostgresPolicyRepository(db *sql.DB) *PostgresPolicyRepository {
	return &PostgresPolicyRepository{db: db}
}

func (r *PostgresPolicyRepository) PutVersion(ctx context.Context, policy *Policy) error {
	if err := ctx.Err(); err != nil { return err }
	if r == nil || r.db == nil { return ErrRepositoryUnavailable }
	if policy == nil || policy.ID == "" || policy.Version < 1 { return ErrRepositoryConflict }
	snapshot := clonePolicy(policy)
	payload, err := json.Marshal(snapshot)
	if err != nil { return fmt.Errorf("%w: encode policy: %v", ErrRepositoryIntegrity, err) }

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return postgresRepositoryError(err) }
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", snapshot.ID); err != nil {
		return postgresRepositoryError(err)
	}
	var activeVersion int
	err = tx.QueryRowContext(ctx, "SELECT version FROM naeos_control_plane.active_policies WHERE policy_id = $1 FOR UPDATE", snapshot.ID).Scan(&activeVersion)
	if err != nil && !errors.Is(err, sql.ErrNoRows) { return postgresRepositoryError(err) }
	hasActive := err == nil
	if hasActive && snapshot.Version < activeVersion { return ErrRepositoryConflict }

	_, err = tx.ExecContext(ctx, `INSERT INTO naeos_control_plane.policy_versions (policy_id, version, policy)
		VALUES ($1, $2, $3::jsonb) ON CONFLICT (policy_id, version) DO NOTHING`, snapshot.ID, snapshot.Version, string(payload))
	if err != nil { return postgresRepositoryError(err) }

	var existingJSON []byte
	if err = tx.QueryRowContext(ctx, "SELECT policy FROM naeos_control_plane.policy_versions WHERE policy_id = $1 AND version = $2", snapshot.ID, snapshot.Version).Scan(&existingJSON); err != nil {
		return postgresRepositoryError(err)
	}
	var existing Policy
	if err = json.Unmarshal(existingJSON, &existing); err != nil { return fmt.Errorf("%w: decode stored policy: %v", ErrRepositoryIntegrity, err) }
	existingCanonical, marshalErr := json.Marshal(&existing)
	if marshalErr != nil { return fmt.Errorf("%w: re-encode stored policy: %v", ErrRepositoryIntegrity, marshalErr) }
	if !bytes.Equal(existingCanonical, payload) { return ErrRepositoryConflict }

	if !hasActive || snapshot.Version > activeVersion {
		_, err = tx.ExecContext(ctx, `INSERT INTO naeos_control_plane.active_policies (policy_id, version, updated_at)
			VALUES ($1, $2, now()) ON CONFLICT (policy_id) DO UPDATE
			SET version = EXCLUDED.version, updated_at = EXCLUDED.updated_at
			WHERE naeos_control_plane.active_policies.version < EXCLUDED.version`, snapshot.ID, snapshot.Version)
		if err != nil { return postgresRepositoryError(err) }
	}
	if err = tx.Commit(); err != nil { return postgresRepositoryError(err) }
	return nil
}

func (r *PostgresPolicyRepository) Active(ctx context.Context, policyID string) (*Policy, error) {
	if err := ctx.Err(); err != nil { return nil, err }
	if r == nil || r.db == nil { return nil, ErrRepositoryUnavailable }
	var payload []byte
	err := r.db.QueryRowContext(ctx, `SELECT pv.policy FROM naeos_control_plane.active_policies ap
		JOIN naeos_control_plane.policy_versions pv ON pv.policy_id = ap.policy_id AND pv.version = ap.version
		WHERE ap.policy_id = $1`, policyID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) { return nil, ErrRepositoryNotFound }
	if err != nil { return nil, postgresRepositoryError(err) }
	var policy Policy
	if err := json.Unmarshal(payload, &policy); err != nil { return nil, fmt.Errorf("%w: decode active policy: %v", ErrRepositoryIntegrity, err) }
	if policy.Status != "active" { return nil, ErrRepositoryConflict }
	return clonePolicy(&policy), nil
}

func (r *PostgresPolicyRepository) Version(ctx context.Context, policyID string, version int) (*Policy, error) {
	if err := ctx.Err(); err != nil { return nil, err }
	if r == nil || r.db == nil { return nil, ErrRepositoryUnavailable }
	var payload []byte
	err := r.db.QueryRowContext(ctx, "SELECT policy FROM naeos_control_plane.policy_versions WHERE policy_id = $1 AND version = $2", policyID, version).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) { return nil, ErrRepositoryNotFound }
	if err != nil { return nil, postgresRepositoryError(err) }
	var policy Policy
	if err := json.Unmarshal(payload, &policy); err != nil { return nil, fmt.Errorf("%w: decode policy version: %v", ErrRepositoryIntegrity, err) }
	if policy.ID != policyID || policy.Version != version { return nil, ErrRepositoryIntegrity }
	return clonePolicy(&policy), nil
}

func (r *PostgresPolicyRepository) CheckActive(ctx context.Context, policyID string, version int) error {
	policy, err := r.Active(ctx, policyID)
	if err != nil { return err }
	if policy.Version != version { return ErrRepositoryConflict }
	return nil
}

// PostgresGrantRepository stores the canonical Grant JSON payload alongside
// indexed lifecycle columns used for agent lookup and revocation.
type PostgresGrantRepository struct{ db *sql.DB }

func NewPostgresGrantRepository(db *sql.DB) *PostgresGrantRepository {
	return &PostgresGrantRepository{db: db}
}

func (r *PostgresGrantRepository) Create(ctx context.Context, grant Grant) error {
	if err := ctx.Err(); err != nil { return err }
	if r == nil || r.db == nil { return ErrRepositoryUnavailable }
	if grant.GrantID == "" || grant.AgentID == "" || grant.PolicyID == "" || grant.PolicyVersion < 1 ||
		len(grant.Capabilities) == 0 || grant.Status != "active" || grant.Revoked ||
		(!grant.CreatedAt.IsZero() && !grant.ExpiresAt.IsZero() && !grant.CreatedAt.Before(grant.ExpiresAt)) {
		return ErrRepositoryConflict
	}
	if grant.CreatedAt.IsZero() { grant.CreatedAt = time.Now().UTC() }
	payload, err := json.Marshal(cloneGrant(grant))
	if err != nil { return fmt.Errorf("%w: encode grant: %v", ErrRepositoryIntegrity, err) }
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return postgresRepositoryError(err) }
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", grant.AgentID); err != nil {
		return postgresRepositoryError(err)
	}
	var existingID string
	err = tx.QueryRowContext(ctx, "SELECT grant_id FROM naeos_control_plane.grants WHERE agent_id = $1", grant.AgentID).Scan(&existingID)
	if err == nil { return ErrRepositoryConflict }
	if !errors.Is(err, sql.ErrNoRows) { return postgresRepositoryError(err) }
	_, err = tx.ExecContext(ctx, `INSERT INTO naeos_control_plane.grants
		(grant_id, agent_id, policy_id, policy_version, status, expires_at, revoked_at, grant_payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULL, $7::jsonb, $8)`,
		grant.GrantID, grant.AgentID, grant.PolicyID, grant.PolicyVersion, grant.Status,
		grant.ExpiresAt, string(payload), grant.CreatedAt)
	if err != nil { return postgresRepositoryError(err) }
	if err := tx.Commit(); err != nil { return postgresRepositoryError(err) }
	return nil
}

func (r *PostgresGrantRepository) Get(ctx context.Context, grantID string) (Grant, error) {
	if err := ctx.Err(); err != nil { return Grant{}, err }
	if r == nil || r.db == nil { return Grant{}, ErrRepositoryUnavailable }
	return r.get(ctx, "grant_id", grantID)
}

func (r *PostgresGrantRepository) GetByAgent(ctx context.Context, agentID string) (Grant, error) {
	if err := ctx.Err(); err != nil { return Grant{}, err }
	if r == nil || r.db == nil { return Grant{}, ErrRepositoryUnavailable }
	return r.get(ctx, "agent_id", agentID)
}

func (r *PostgresGrantRepository) get(ctx context.Context, column, value string) (Grant, error) {
	// column is selected only from the two fixed call-site constants above.
	var payload []byte
	err := r.db.QueryRowContext(ctx, "SELECT grant_payload FROM naeos_control_plane.grants WHERE "+column+" = $1", value).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) { return Grant{}, ErrRepositoryNotFound }
	if err != nil { return Grant{}, postgresRepositoryError(err) }
	var grant Grant
	if err := json.Unmarshal(payload, &grant); err != nil { return Grant{}, fmt.Errorf("%w: decode grant: %v", ErrRepositoryIntegrity, err) }
	if grant.GrantID == "" || grant.AgentID == "" { return Grant{}, ErrRepositoryIntegrity }
	return cloneGrant(grant), nil
}

func (r *PostgresGrantRepository) Revoke(ctx context.Context, grantID, reason string, at time.Time) error {
	if err := ctx.Err(); err != nil { return err }
	if r == nil || r.db == nil { return ErrRepositoryUnavailable }
	if reason == "" { return ErrRepositoryConflict }
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return postgresRepositoryError(err) }
	defer func() { _ = tx.Rollback() }()
	var payload []byte
	err = tx.QueryRowContext(ctx, "SELECT grant_payload FROM naeos_control_plane.grants WHERE grant_id = $1 FOR UPDATE", grantID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) { return ErrRepositoryNotFound }
	if err != nil { return postgresRepositoryError(err) }
	var grant Grant
	if err := json.Unmarshal(payload, &grant); err != nil { return fmt.Errorf("%w: decode grant: %v", ErrRepositoryIntegrity, err) }
	if grant.Revoked || grant.Status != "active" { return ErrRepositoryConflict }
	revokedAt := at.UTC()
	grant.Revoked = true
	grant.Status = "revoked"
	grant.RevokeReason = reason
	grant.RevokedAt = &revokedAt
	updated, err := json.Marshal(grant)
	if err != nil { return fmt.Errorf("%w: encode revoked grant: %v", ErrRepositoryIntegrity, err) }
	_, err = tx.ExecContext(ctx, `UPDATE naeos_control_plane.grants
		SET status = 'revoked', revoked_at = $2, grant_payload = $3::jsonb
		WHERE grant_id = $1 AND status = 'active'`, grantID, revokedAt, string(updated))
	if err != nil { return postgresRepositoryError(err) }
	if err := tx.Commit(); err != nil { return postgresRepositoryError(err) }
	return nil
}

// PostgresApprovalRepository implements durable approval creation and
// compare-and-set lifecycle transitions.
type PostgresApprovalRepository struct{ db *sql.DB }

func NewPostgresApprovalRepository(db *sql.DB) *PostgresApprovalRepository {
	return &PostgresApprovalRepository{db: db}
}

func (r *PostgresApprovalRepository) Create(ctx context.Context, approval Approval) error {
	if err := ctx.Err(); err != nil { return err }
	if r == nil || r.db == nil { return ErrRepositoryUnavailable }
	if approval.ID == "" || approval.DecisionID == "" || approval.Approver == "" || approval.Status != "pending" ||
		(!approval.ExpiresAt.IsZero() && !approval.CreatedAt.IsZero() && !approval.CreatedAt.Before(approval.ExpiresAt)) {
		return ErrRepositoryConflict
	}
	if approval.CreatedAt.IsZero() { approval.CreatedAt = time.Now().UTC() }
	var expires any
	if !approval.ExpiresAt.IsZero() { expires = approval.ExpiresAt }
	_, err := r.db.ExecContext(ctx, `INSERT INTO naeos_control_plane.approvals
		(approval_id, decision_id, agent_id, capability, policy_id, policy_version, approver, status,
		 created_at, expires_at, approved_at, artifact_hash, reason)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'pending',$8,$9,NULL,$10,$11)`,
		approval.ID, approval.DecisionID, approval.AgentID, string(approval.Capability),
		approval.PolicyID, approval.PolicyVersion, approval.Approver, approval.CreatedAt,
		expires, approval.ArtifactHash, approval.Reason)
	if err != nil { return postgresRepositoryError(err) }
	return nil
}

func (r *PostgresApprovalRepository) Get(ctx context.Context, id string) (Approval, error) {
	if err := ctx.Err(); err != nil { return Approval{}, err }
	if r == nil || r.db == nil { return Approval{}, ErrRepositoryUnavailable }
	return r.get(ctx, id)
}

func (r *PostgresApprovalRepository) get(ctx context.Context, id string) (Approval, error) {
	var a Approval
	var expires, approved sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT approval_id, decision_id, agent_id, capability, policy_id,
		policy_version, approver, status, created_at, expires_at, approved_at, artifact_hash, reason
		FROM naeos_control_plane.approvals WHERE approval_id = $1`, id).Scan(
		&a.ID, &a.DecisionID, &a.AgentID, &a.Capability, &a.PolicyID, &a.PolicyVersion,
		&a.Approver, &a.Status, &a.CreatedAt, &expires, &approved, &a.ArtifactHash, &a.Reason)
	if errors.Is(err, sql.ErrNoRows) { return Approval{}, ErrRepositoryNotFound }
	if err != nil { return Approval{}, postgresRepositoryError(err) }
	if expires.Valid { a.ExpiresAt = expires.Time }
	if approved.Valid { a.ApprovedAt = approved.Time }
	return a, nil
}

func (r *PostgresApprovalRepository) Transition(ctx context.Context, id, expectedStatus, nextStatus string, at time.Time) (Approval, error) {
	if err := ctx.Err(); err != nil { return Approval{}, err }
	if r == nil || r.db == nil { return Approval{}, ErrRepositoryUnavailable }
	if !validApprovalTransition(expectedStatus, nextStatus) { return Approval{}, ErrRepositoryConflict }
	at = at.UTC()

	if expectedStatus == "pending" && nextStatus != "expired" {
		var expires sql.NullTime
		err := r.db.QueryRowContext(ctx, "SELECT expires_at FROM naeos_control_plane.approvals WHERE approval_id = $1", id).Scan(&expires)
		if errors.Is(err, sql.ErrNoRows) { return Approval{}, ErrRepositoryNotFound }
		if err != nil { return Approval{}, postgresRepositoryError(err) }
		if expires.Valid && !at.Before(expires.Time) {
			_, updateErr := r.db.ExecContext(ctx, "UPDATE naeos_control_plane.approvals SET status = 'expired' WHERE approval_id = $1 AND status = 'pending'", id)
			if updateErr != nil { return Approval{}, postgresRepositoryError(updateErr) }
			return Approval{}, ErrRepositoryConflict
		}
	}
	if nextStatus == "expired" {
		var expires sql.NullTime
		err := r.db.QueryRowContext(ctx, "SELECT expires_at FROM naeos_control_plane.approvals WHERE approval_id = $1", id).Scan(&expires)
		if errors.Is(err, sql.ErrNoRows) { return Approval{}, ErrRepositoryNotFound }
		if err != nil { return Approval{}, postgresRepositoryError(err) }
		if !expires.Valid || at.Before(expires.Time) { return Approval{}, ErrRepositoryConflict }
	}
	var approvedAt any
	if nextStatus == "approved" { approvedAt = at }
	var consumedAt any
	if nextStatus == "consumed" { consumedAt = at }
	result, err := r.db.ExecContext(ctx, `UPDATE naeos_control_plane.approvals
		SET status = $3, approved_at = COALESCE($4, approved_at), consumed_at = COALESCE($5, consumed_at)
		WHERE approval_id = $1 AND status = $2
		  AND NOT ($2 = 'approved' AND $3 = 'consumed' AND expires_at IS NOT NULL AND expires_at <= $6)
		  AND NOT ($2 = 'pending' AND $3 IN ('approved','rejected') AND expires_at IS NOT NULL AND expires_at <= $6)`,
		id, expectedStatus, nextStatus, approvedAt, consumedAt, at)
	if err != nil { return Approval{}, postgresRepositoryError(err) }
	affected, err := result.RowsAffected()
	if err != nil { return Approval{}, postgresRepositoryError(err) }
	if affected == 0 {
		if _, getErr := r.get(ctx, id); errors.Is(getErr, ErrRepositoryNotFound) { return Approval{}, ErrRepositoryNotFound } else if getErr != nil { return Approval{}, getErr }
		return Approval{}, ErrRepositoryConflict
	}
	return r.get(ctx, id)
}

func postgresRepositoryError(err error) error {
	if err == nil { return nil }
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) { return err }
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": return fmt.Errorf("%w: %v", ErrRepositoryConflict, err)
		case "23503", "23514", "22P02": return fmt.Errorf("%w: %v", ErrRepositoryIntegrity, err)
		}
	}
	return fmt.Errorf("%w: %v", ErrRepositoryUnavailable, err)
}

var (
	_ PolicyRepository = (*PostgresPolicyRepository)(nil)
	_ GrantRepository = (*PostgresGrantRepository)(nil)
	_ ApprovalRepository = (*PostgresApprovalRepository)(nil)
)
