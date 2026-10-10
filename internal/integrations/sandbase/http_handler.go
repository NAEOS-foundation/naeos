// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sandbase

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	ExternalAuthorizationSchema       = "sandbase.authz/v1"
	ExternalAuthorizationDigestSchema = "sandbase.digest/v1"
	maxAuthorizationBodyBytes         = 64 * 1024
)

// SandBaseAuthorizationRequest is the exact versioned envelope emitted by SandBase Harness.
type SandBaseAuthorizationRequest struct {
	Schema              string `json:"schema"`
	SessionID           string `json:"session_id"`
	InvocationID        string `json:"invocation_id"`
	Capability          string `json:"capability"`
	Target              string `json:"target"`
	ArgumentsDigest     string `json:"arguments_digest"`
	PolicyContextDigest string `json:"policy_context_digest"`
	DigestSchema        string `json:"digest_schema"`
}

// SandBaseAuthorizationResponse is the decision shape consumed by SandBase's veto-only hook.
type SandBaseAuthorizationResponse struct {
	Decision      string `json:"decision"`
	Reason        string `json:"reason,omitempty"`
	PolicyVersion string `json:"policy_version,omitempty"`
	DecisionID    string `json:"decision_id,omitempty"`
	ContextDigest string `json:"context_digest,omitempty"`
}

// NewHTTPHandler exposes Adapter using SandBase's sandbase.authz/v1 HTTP contract.
// The adapter's configured Grant supplies the trusted NAEOS agent identity; callers
// cannot select an agent by sending an untrusted field in the SandBase envelope.
func NewHTTPHandler(adapter *Adapter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, `{"decision":"deny","reason":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if adapter == nil || adapter.Gateway == nil || adapter.Policy == nil || adapter.Grant == nil {
			writeSandBaseDecision(w, http.StatusServiceUnavailable, SandBaseAuthorizationResponse{
				Decision: "deny", Reason: "authorization_unavailable",
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxAuthorizationBodyBytes)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request SandBaseAuthorizationRequest
		if err := decoder.Decode(&request); err != nil {
			writeSandBaseDecision(w, http.StatusBadRequest, SandBaseAuthorizationResponse{
				Decision: "deny", Reason: "malformed_request",
			})
			return
		}
		if err := ensureSingleJSONValue(decoder); err != nil {
			writeSandBaseDecision(w, http.StatusBadRequest, SandBaseAuthorizationResponse{
				Decision: "deny", Reason: "malformed_request",
			})
			return
		}
		if err := validateSandBaseRequest(request); err != nil {
			writeSandBaseDecision(w, http.StatusBadRequest, SandBaseAuthorizationResponse{
				Decision: "deny", Reason: "invalid_request",
			})
			return
		}

		// Bind both SandBase digests into the NAEOS action. The arguments digest is
		// also the artifact hash used by the NAEOS decision ledger.
		authorization, err := adapter.Authorize(AuthorizationRequest{
			SessionID: request.SessionID,
			RequestID: request.InvocationID,
			AgentID: adapter.Grant.AgentID,
			Capability: request.Capability,
			Target: request.Target,
			ArtifactHash: "sha256:" + strings.ToLower(request.ArgumentsDigest),
			Context: map[string]string{
				"schema": request.Schema,
				"digest_schema": request.DigestSchema,
				"arguments_digest": strings.ToLower(request.ArgumentsDigest),
				"policy_context_digest": strings.ToLower(request.PolicyContextDigest),
			},
		})
		if err != nil {
			writeSandBaseDecision(w, http.StatusServiceUnavailable, SandBaseAuthorizationResponse{
				Decision: "deny", Reason: "authorization_unavailable",
			})
			return
		}

		decision := strings.ToLower(authorization.Decision)
		switch decision {
		case "allow":
			// Keep ALLOW only for an explicit NAEOS allow. SandBase's hook is
			// veto-only; every other NAEOS status must refuse execution.
		case "deny":
			if authorization.Reason == "stale_policy" || authorization.Reason == "grant_version_mismatch" {
				decision = "reauthorize"
			}
		default:
			decision = "deny"
		}
		writeSandBaseDecision(w, http.StatusOK, SandBaseAuthorizationResponse{
			Decision: decision,
			Reason: authorization.Reason,
			PolicyVersion: strconv.Itoa(authorization.PolicyVersion),
			DecisionID: authorization.DecisionID,
			ContextDigest: strings.ToLower(request.PolicyContextDigest),
		})
	})
}

func validateSandBaseRequest(request SandBaseAuthorizationRequest) error {
	if request.Schema != ExternalAuthorizationSchema {
		return fmt.Errorf("unsupported schema")
	}
	if request.DigestSchema != ExternalAuthorizationDigestSchema {
		return fmt.Errorf("unsupported digest schema")
	}
	if strings.TrimSpace(request.SessionID) == "" ||
		strings.TrimSpace(request.InvocationID) == "" ||
		strings.TrimSpace(request.Capability) == "" ||
		strings.TrimSpace(request.Target) == "" {
		return fmt.Errorf("required field is empty")
	}
	if !isSHA256Hex(request.ArgumentsDigest) || !isSHA256Hex(request.PolicyContextDigest) {
		return fmt.Errorf("digest must be a 64-character SHA-256 hex string")
	}
	return nil
}

func isSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return fmt.Errorf("multiple JSON values")
	}
	return err
}

func writeSandBaseDecision(w http.ResponseWriter, status int, response SandBaseAuthorizationResponse) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}
