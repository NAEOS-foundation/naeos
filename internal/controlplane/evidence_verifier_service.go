// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrAuthoritativeEvidenceNotVerified indicates that the provider observation
// is missing, contradictory, stale, or otherwise insufficient to confirm an outcome.
var ErrAuthoritativeEvidenceNotVerified = errors.New("authoritative execution evidence is not verified")

// AuthoritativeExecutionObservation is returned by a provider-specific, read-only
// adapter after independently observing the provider authority.
type AuthoritativeExecutionObservation struct {
	ExecutionID    string
	Outcome        ExecutionOutcome
	EvidenceRef    string
	Authority      string
	EvidenceSHA256 string
	ObservedAt     time.Time
	ValidUntil     time.Time
}

// AuthoritativeExecutionEvidenceSource must query an authority capable of observing
// the external effect. It must not trust the submitted reference as proof.
type AuthoritativeExecutionEvidenceSource interface {
	ObserveExecution(context.Context, string, ExecutionOutcome, string) (AuthoritativeExecutionObservation, error)
}

// ExecutionEvidencePolicy applies effect-specific verification and freshness rules.
// It must reject estimated, contradictory, stale, or authority-inadequate evidence.
type ExecutionEvidencePolicy interface {
	Version() string
	EvaluateExecutionEvidence(context.Context, AuthoritativeExecutionObservation, ExecutionOutcome) error
}

// ProviderNeutralEvidenceVerificationService is the provider-neutral HTTP service
// core. Provider adapters and policy are injected; neither is implemented by this
// package, so the service cannot confirm an outcome without both dependencies.
type ProviderNeutralEvidenceVerificationService struct {
	source AuthoritativeExecutionEvidenceSource
	policy ExecutionEvidencePolicy
	token  string
	now    func() time.Time
}

// NewProviderNeutralEvidenceVerificationService creates a fail-closed service core.
func NewProviderNeutralEvidenceVerificationService(source AuthoritativeExecutionEvidenceSource, policy ExecutionEvidencePolicy, token string) (*ProviderNeutralEvidenceVerificationService, error) {
	if source == nil || policy == nil || strings.TrimSpace(policy.Version()) == "" || strings.TrimSpace(token) == "" {
		return nil, errors.New("authoritative evidence source, verification policy, and service credential are required")
	}
	return &ProviderNeutralEvidenceVerificationService{
		source: source, policy: policy, token: token, now: time.Now,
	}, nil
}

// Handler implements the v1 POST verification contract. Mount it on a dedicated
// HTTPS listener or an authenticated internal service route.
func (s *ProviderNeutralEvidenceVerificationService) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s == nil || s.source == nil || s.policy == nil || strings.TrimSpace(s.token) == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "evidence verification service is not configured"})
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok || len(token) != len(s.token) || subtle.ConstantTimeCompare([]byte(token), []byte(s.token)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "service credential required"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxEvidenceVerifierRequestBytes)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request evidenceVerificationRequest
		if err := decoder.Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid evidence verification request"})
			return
		}
		if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request body must contain exactly one JSON object"})
			return
		}
		if len(request.RequestID) != 32 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request ID must be 128-bit hexadecimal"})
			return
		}
		if _, err := hex.DecodeString(request.RequestID); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request ID must be 128-bit hexadecimal"})
			return
		}
		request.ExecutionID = strings.TrimSpace(request.ExecutionID)
		request.EvidenceRef = strings.TrimSpace(request.EvidenceRef)
		if request.ExecutionID == "" || len(request.ExecutionID) > 256 ||
			(request.Outcome != ExecutionOutcomeCommitted && request.Outcome != ExecutionOutcomeNotCommitted) ||
			len(request.EvidenceRef) > 4096 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid execution binding or expected outcome"})
			return
		}

		observation, err := s.source.ObserveExecution(r.Context(), request.ExecutionID, request.Outcome, request.EvidenceRef)
		if err != nil {
			if errors.Is(err, ErrAuthoritativeEvidenceNotVerified) {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "authoritative execution evidence is not verified"})
				return
			}
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "authoritative evidence source unavailable"})
			return
		}
		now := s.now().UTC()
		if observation.ExecutionID != request.ExecutionID || observation.Outcome != request.Outcome ||
			!validEvidenceReference(observation.EvidenceRef) || strings.TrimSpace(observation.Authority) == "" ||
			!sha256HexPattern.MatchString(observation.EvidenceSHA256) ||
			observation.ObservedAt.IsZero() || observation.ValidUntil.IsZero() ||
			observation.ObservedAt.After(now.Add(30*time.Second)) || !observation.ValidUntil.After(now) ||
			observation.ValidUntil.Before(observation.ObservedAt) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "provider observation is unbound, incomplete, or expired"})
			return
		}
		if err := s.policy.EvaluateExecutionEvidence(r.Context(), observation, request.Outcome); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "evidence does not satisfy verification policy"})
			return
		}

		writeJSON(w, http.StatusOK, evidenceVerificationResponse{
			RequestID: request.RequestID, ExecutionID: observation.ExecutionID, Outcome: observation.Outcome,
			Status: "VERIFIED", EvidenceRef: observation.EvidenceRef, Authority: observation.Authority,
			EvidenceSHA256: observation.EvidenceSHA256, ObservedAt: observation.ObservedAt,
			ValidUntil: observation.ValidUntil, VerificationPolicyVersion: s.policy.Version(),
		})
	})
}
