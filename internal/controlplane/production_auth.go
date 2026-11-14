// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/api"
)

const maxAuthorizationRequestBytes = 1 << 20

// ProductionIdentityConfig binds the production runtime to the configured identity
// provider. Issuer, audience, and JWKS endpoint are explicit trust anchors.
type ProductionIdentityConfig struct {
	Issuer   string
	Audience string
	JWKSURL  string
}

// ConfigureIdentity installs JWKS authentication and repository-backed
// authorization. Readiness remains false until this succeeds.
func (r *ProductionRuntime) ConfigureIdentity(cfg ProductionIdentityConfig) error {
	if r == nil || r.Policies == nil || r.Grants == nil || r.Approvals == nil || r.Ledger == nil {
		return errors.New("control plane persistence dependencies are unavailable")
	}
	validator, err := api.NewJWKSValidator(cfg.Issuer, cfg.Audience, cfg.JWKSURL)
	if err != nil {
		return errors.New("control plane identity trust configuration is invalid")
	}
	r.identity = validator
	r.gateway = NewRepositoryDecisionGateway(r.Policies, r.Grants, r.Approvals, r.Ledger)
	return nil
}

// ProductionHandler exposes the health probes and one authenticated
// authorization route. No demo, reset, scenario, or execution routes are mounted.
func (r *ProductionRuntime) ProductionHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "live"})
	})
	mux.Handle("GET /health/ready", r.readyHandler())
	mux.Handle("POST /api/control-plane/decision", r.authorizationHandler())
	return mux
}

func (r *ProductionRuntime) readyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r == nil || r.pool == nil || r.sqlDB == nil || r.identity == nil || r.gateway == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		if err := r.pool.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		if err := r.sqlDB.PingContext(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		if err := verifyControlPlaneSchema(ctx, r.pool); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
}

func (r *ProductionRuntime) authorizationHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r == nil || r.identity == nil || r.gateway == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "authorization runtime unavailable"})
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
		req.Body = http.MaxBytesReader(w, req.Body, maxAuthorizationRequestBytes)
		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields()
		var authorization AuthorizeRequest
		if err := decoder.Decode(&authorization); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid authorization request"})
			return
		}
		if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request body must contain exactly one JSON object"})
			return
		}

		principal := strings.TrimSpace(claims.Sub)
		if authorization.AgentID != "" && authorization.AgentID != principal {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "token identity does not match requested agent"})
			return
		}
		if authorization.Action.AgentID != "" && authorization.Action.AgentID != principal {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "token identity does not match action agent"})
			return
		}
		// Identity comes only from the verified token. Client-supplied policy,
		// grant, and authorization-context objects are not sources of authority.
		authorization.AgentID = principal
		authorization.Action.AgentID = principal
		authorization.Grant = nil
		authorization.Policy = nil
		authorization.Context = AuthorizationContext{}
		if authorization.Timestamp.IsZero() {
			authorization.Timestamp = time.Now().UTC()
		}

		result, err := r.gateway.Authorize(req.Context(), authorization)
		if err != nil && result.Status == DecisionIndeterminate {
			writeJSON(w, http.StatusServiceUnavailable, result)
			return
		}
		switch result.Status {
		case DecisionAllow:
			writeJSON(w, http.StatusOK, result)
		case DecisionPending:
			writeJSON(w, http.StatusAccepted, result)
		case DecisionDeny:
			writeJSON(w, http.StatusForbidden, result)
		default:
			writeJSON(w, http.StatusServiceUnavailable, result)
		}
	})
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
