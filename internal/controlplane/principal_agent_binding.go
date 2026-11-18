// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// PrincipalAgentBindingResolver resolves a verified issuer-qualified principal
// to an explicitly provisioned NAEOS agent. A caller-supplied agent ID may only
// narrow the lookup; it never creates or expands authority.
type PrincipalAgentBindingResolver interface {
	ResolveAgent(ctx context.Context, issuer, principalKind, principalID, requestedAgentID string) (string, error)
}

// PostgresPrincipalAgentBindingResolver reads the immutable identity tuple and
// active status from the private Control Plane binding table.
type PostgresPrincipalAgentBindingResolver struct {
	db *sql.DB
}

func NewPostgresPrincipalAgentBindingResolver(db *sql.DB) *PostgresPrincipalAgentBindingResolver {
	return &PostgresPrincipalAgentBindingResolver{db: db}
}

func (r *PostgresPrincipalAgentBindingResolver) ResolveAgent(
	ctx context.Context, issuer, principalKind, principalID, requestedAgentID string,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if r == nil || r.db == nil {
		return "", ErrRepositoryUnavailable
	}
	if strings.TrimSpace(issuer) == "" || strings.TrimSpace(principalID) == "" ||
		(principalKind != "human" && principalKind != "service") ||
		(requestedAgentID != "" && strings.TrimSpace(requestedAgentID) == "") {
		return "", ErrRepositoryIntegrity
	}

	if requestedAgentID != "" {
		var agentID string
		err := r.db.QueryRowContext(ctx, `SELECT agent_id
			FROM naeos_control_plane.principal_agent_bindings
			WHERE issuer = $1 AND principal_kind = $2 AND principal_id = $3
			  AND agent_id = $4 AND status = 'active'`,
			issuer, principalKind, principalID, requestedAgentID).Scan(&agentID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRepositoryNotFound
		}
		if err != nil {
			return "", postgresRepositoryError(err)
		}
		if agentID != requestedAgentID {
			return "", ErrRepositoryIntegrity
		}
		return agentID, nil
	}

	rows, err := r.db.QueryContext(ctx, `SELECT agent_id
		FROM naeos_control_plane.principal_agent_bindings
		WHERE issuer = $1 AND principal_kind = $2 AND principal_id = $3
		  AND status = 'active'
		ORDER BY agent_id
		LIMIT 2`, issuer, principalKind, principalID)
	if err != nil {
		return "", postgresRepositoryError(err)
	}
	defer rows.Close()

	var agentID string
	count := 0
	for rows.Next() {
		var candidate string
		if err := rows.Scan(&candidate); err != nil {
			return "", postgresRepositoryError(err)
		}
		count++
		if count == 1 {
			agentID = candidate
		}
	}
	if err := rows.Err(); err != nil {
		return "", postgresRepositoryError(err)
	}
	switch count {
	case 0:
		return "", ErrRepositoryNotFound
	case 1:
		if agentID == "" {
			return "", ErrRepositoryIntegrity
		}
		return agentID, nil
	default:
		// Do not guess when one principal has multiple active agent bindings.
		// The request must explicitly select one of its authorized bindings.
		return "", ErrRepositoryConflict
	}
}

var _ PrincipalAgentBindingResolver = (*PostgresPrincipalAgentBindingResolver)(nil)
