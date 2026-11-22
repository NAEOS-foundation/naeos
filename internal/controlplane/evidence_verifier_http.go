// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	maxEvidenceVerifierRequestBytes  = 16 << 10
	maxEvidenceVerifierResponseBytes = 16 << 10
)

var sha256HexPattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

// HTTPExecutionEvidenceVerifier delegates verification to a separately operated
// NAEOS Evidence Verification Service. The service must independently observe the
// provider authority; this client only authenticates and validates its response.
type HTTPExecutionEvidenceVerifier struct {
	endpoint string
	token    string
	client   *http.Client
}

type evidenceVerificationRequest struct {
	RequestID   string           `json:"request_id"`
	ExecutionID string           `json:"execution_id"`
	Outcome     ExecutionOutcome `json:"expected_outcome"`
	EvidenceRef string           `json:"evidence_ref,omitempty"`
}

type evidenceVerificationResponse struct {
	RequestID                 string           `json:"request_id"`
	ExecutionID               string           `json:"execution_id"`
	Outcome                   ExecutionOutcome `json:"outcome"`
	Status                    string           `json:"status"`
	EvidenceRef               string           `json:"evidence_ref"`
	Authority                 string           `json:"authority"`
	EvidenceSHA256            string           `json:"evidence_sha256"`
	ObservedAt                time.Time        `json:"observed_at"`
	ValidUntil                time.Time        `json:"valid_until"`
	VerificationPolicyVersion string           `json:"verification_policy_version"`
}

// NewHTTPExecutionEvidenceVerifier constructs a provider-neutral verifier client.
// The endpoint must be an HTTPS URL and token must be a dedicated service credential.
func NewHTTPExecutionEvidenceVerifier(endpoint, token string) (*HTTPExecutionEvidenceVerifier, error) {
	return newHTTPExecutionEvidenceVerifier(endpoint, token, &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})
}

func newHTTPExecutionEvidenceVerifier(endpoint, token string, client *http.Client) (*HTTPExecutionEvidenceVerifier, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("evidence verifier endpoint must be an HTTPS URL without userinfo, query, or fragment")
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("evidence verifier service credential is required")
	}
	if client == nil {
		return nil, errors.New("evidence verifier HTTP client is required")
	}
	return &HTTPExecutionEvidenceVerifier{
		endpoint: parsed.String(),
		token:    token,
		client:   client,
	}, nil
}

// VerifyExecutionEvidence accepts only a fresh, request-bound VERIFIED receipt
// from the configured verifier service. The supplied evidence reference is an
// untrusted lookup hint, never proof by itself.
func (v *HTTPExecutionEvidenceVerifier) VerifyExecutionEvidence(ctx context.Context, executionID string, outcome ExecutionOutcome, submittedEvidenceRef string) (string, error) {
	if v == nil || v.client == nil || strings.TrimSpace(v.endpoint) == "" || strings.TrimSpace(v.token) == "" {
		return "", errors.New("evidence verifier is not configured")
	}
	executionID = strings.TrimSpace(executionID)
	if ctx == nil || executionID == "" || (outcome != ExecutionOutcomeCommitted && outcome != ExecutionOutcomeNotCommitted) {
		return "", errors.New("invalid evidence verification request")
	}
	requestIDBytes := make([]byte, 16)
	if _, err := rand.Read(requestIDBytes); err != nil {
		return "", errors.New("could not generate evidence verification request ID")
	}
	requestID := hex.EncodeToString(requestIDBytes)
	payload, err := json.Marshal(evidenceVerificationRequest{
		RequestID: requestID, ExecutionID: executionID, Outcome: outcome,
		EvidenceRef: strings.TrimSpace(submittedEvidenceRef),
	})
	if err != nil || len(payload) > maxEvidenceVerifierRequestBytes {
		return "", errors.New("evidence verification request is invalid")
	}

	// #nosec G704 -- endpoint is validated as HTTPS configuration; redirects are explicitly blocked by the client.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", errors.New("could not create evidence verification request")
	}
	req.Header.Set("Authorization", "Bearer "+v.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-store")

	// #nosec G704 -- request destination is the validated configured verifier endpoint; redirect policy blocks forwarding.
	resp, err := v.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("evidence verification service unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("evidence verification service returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxEvidenceVerifierResponseBytes+1))
	if err != nil || len(body) > maxEvidenceVerifierResponseBytes {
		return "", errors.New("evidence verification response is unreadable or too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var result evidenceVerificationResponse
	if err := decoder.Decode(&result); err != nil {
		return "", errors.New("evidence verification response does not match the contract")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return "", errors.New("evidence verification response must contain exactly one JSON object")
	}

	now := time.Now().UTC()
	if result.RequestID != requestID || result.ExecutionID != executionID || result.Outcome != outcome ||
		result.Status != "VERIFIED" || strings.TrimSpace(result.Authority) == "" ||
		strings.TrimSpace(result.VerificationPolicyVersion) == "" ||
		!sha256HexPattern.MatchString(result.EvidenceSHA256) ||
		!validEvidenceReference(result.EvidenceRef) ||
		result.ObservedAt.IsZero() || result.ValidUntil.IsZero() ||
		result.ObservedAt.After(now.Add(30*time.Second)) || !result.ValidUntil.After(now) ||
		result.ValidUntil.Before(result.ObservedAt) {
		return "", errors.New("evidence verification receipt is unbound, incomplete, or not currently valid")
	}
	return result.EvidenceRef, nil
}

func validEvidenceReference(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && parsed.IsAbs() && parsed.Scheme != "" && parsed.Host != "" &&
		parsed.User == nil && parsed.Fragment == ""
}
