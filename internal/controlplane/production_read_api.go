// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const controlPlaneReadScope = "control-plane:read"

type controlPlaneOverviewResponse struct {
	Environment string    `json:"environment"`
	Status      string    `json:"status"`
	ObservedAt  time.Time `json:"observed_at"`
	LastSequence int64    `json:"last_sequence"`
	DecisionCount int     `json:"decision_count"`
	AllowCount int        `json:"allow_count"`
	DenyCount int         `json:"deny_count"`
	PendingCount int      `json:"pending_count"`
}

type controlPlaneDecision struct {
	Sequence   int64           `json:"sequence"`
	ID         string          `json:"id"`
	Timestamp  time.Time       `json:"timestamp"`
	RequestID  string          `json:"request_id,omitempty"`
	DecisionID string          `json:"decision_id,omitempty"`
	AgentID    string          `json:"agent_id"`
	Capability Capability      `json:"capability,omitempty"`
	EventType  string          `json:"event_type"`
	Decision   DecisionStatus  `json:"decision,omitempty"`
	Reason     DecisionReason  `json:"reason,omitempty"`
}

type controlPlaneDecisionPage struct {
	Environment string                   `json:"environment"`
	ObservedAt  time.Time                `json:"observed_at"`
	Items       []controlPlaneDecision   `json:"items"`
	NextCursor  string                   `json:"next_cursor,omitempty"`
}

// readAuthorizationHandler authenticates the caller and requires an explicit
// Control Plane read scope. CRM roles or UI visibility are not accepted here.
func (r *ProductionRuntime) readAuthorizationHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r == nil || r.identity == nil || r.ledgerReader == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "control plane read runtime unavailable"})
			return
		}
		token, ok := bearerToken(req.Header.Get("Authorization"))
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "bearer token required"})
			return
		}
		claims, err := r.identity.Validate(req.Context(), token)
		if err != nil || claims == nil || strings.TrimSpace(claims.Sub) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid bearer token"})
			return
		}
		if !hasControlPlaneReadScope(claims.Scope, claims.Scopes) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "control-plane:read scope required"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, req)
	})
}

func hasControlPlaneReadScope(scope string, scopes []string) bool {
	for _, candidate := range strings.Fields(scope) {
		if candidate == controlPlaneReadScope {
			return true
		}
	}
	for _, candidate := range scopes {
		if strings.TrimSpace(candidate) == controlPlaneReadScope {
			return true
		}
	}
	return false
}

func (r *ProductionRuntime) overviewHandler(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	events, err := r.ledgerReader.Load(ctx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "verified control plane ledger is unavailable"})
		return
	}
	result := controlPlaneOverviewResponse{
		Environment: r.environment,
		Status: "verified",
		ObservedAt: time.Now().UTC(),
		LastSequence: int64(len(events)),
	}
	for _, event := range events {
		if event.EventType != "AUTHORIZATION_DECISION" {
			continue
		}
		result.DecisionCount++
		switch event.Decision {
		case DecisionAllow:
			result.AllowCount++
		case DecisionDeny:
			result.DenyCount++
		case DecisionPending:
			result.PendingCount++
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (r *ProductionRuntime) decisionsHandler(w http.ResponseWriter, req *http.Request) {
	limit := 50
	if raw := req.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be an integer between 1 and 100"})
			return
		}
		limit = parsed
	}
	var after int64
	if raw := req.URL.Query().Get("after"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "after must be a non-negative ledger sequence"})
			return
		}
		after = parsed
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	events, err := r.ledgerReader.Load(ctx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "verified control plane ledger is unavailable"})
		return
	}

	items := make([]controlPlaneDecision, 0, limit)
	var nextCursor string
	for index, event := range events {
		sequence := int64(index + 1)
		if sequence <= after || event.EventType != "AUTHORIZATION_DECISION" {
			continue
		}
		if len(items) == limit {
			nextCursor = strconv.FormatInt(sequence-1, 10)
			break
		}
		items = append(items, controlPlaneDecision{
			Sequence: sequence, ID: event.ID, Timestamp: event.Timestamp,
			RequestID: event.RequestID, DecisionID: event.DecisionID,
			AgentID: event.AgentID, Capability: event.Capability,
			EventType: event.EventType, Decision: event.Decision, Reason: event.Reason,
		})
	}
	writeJSON(w, http.StatusOK, controlPlaneDecisionPage{
		Environment: r.environment, ObservedAt: time.Now().UTC(),
		Items: items, NextCursor: nextCursor,
	})
}

