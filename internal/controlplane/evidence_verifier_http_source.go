// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPAuthoritativeEvidenceSource queries a configured read-only observation API.
type HTTPAuthoritativeEvidenceSource struct {
	endpoint *url.URL
	token    string
	client   *http.Client
}

type authoritativeObservationResponse struct {
	ExecutionID    string           `json:"execution_id"`
	Outcome        ExecutionOutcome `json:"outcome"`
	EvidenceRef    string           `json:"evidence_ref"`
	Authority      string           `json:"authority"`
	EvidenceSHA256 string           `json:"evidence_sha256"`
	ObservedAt     time.Time        `json:"observed_at"`
	ValidUntil     time.Time        `json:"valid_until"`
}

func NewHTTPAuthoritativeEvidenceSource(endpoint, token string) (*HTTPAuthoritativeEvidenceSource, error) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return nil, errors.New("evidence source endpoint must be HTTPS without userinfo or fragment")
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("evidence source credential required")
	}
	if u.RawQuery != "" {
		return nil, errors.New("evidence source endpoint must not contain query")
	}
	return &HTTPAuthoritativeEvidenceSource{
		endpoint: u,
		token:    token,
		client: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func (s *HTTPAuthoritativeEvidenceSource) ObserveExecution(ctx context.Context, executionID string, outcome ExecutionOutcome, hint string) (AuthoritativeExecutionObservation, error) {
	if s == nil || s.endpoint == nil || s.client == nil || s.token == "" {
		return AuthoritativeExecutionObservation{}, errors.New("evidence source not configured")
	}
	u := *s.endpoint
	q := u.Query()
	q.Set("execution_id", executionID)
	q.Set("expected_outcome", string(outcome))
	if strings.TrimSpace(hint) != "" {
		q.Set("evidence_ref_hint", hint)
	}
	u.RawQuery = q.Encode()
	// #nosec G704 -- destination derives from the validated HTTPS source URL; redirects are explicitly blocked.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return AuthoritativeExecutionObservation{}, err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-store")
	// #nosec G704 -- request is sent only to the configured validated HTTPS authority, with redirects disabled.
	resp, err := s.client.Do(req)
	if err != nil {
		return AuthoritativeExecutionObservation{}, fmt.Errorf("authoritative source request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusUnprocessableEntity {
		return AuthoritativeExecutionObservation{}, ErrAuthoritativeEvidenceNotVerified
	}
	if resp.StatusCode != http.StatusOK {
		return AuthoritativeExecutionObservation{}, fmt.Errorf("authoritative source returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024+1))
	if err != nil {
		return AuthoritativeExecutionObservation{}, err
	}
	if len(body) > 64*1024 {
		return AuthoritativeExecutionObservation{}, errors.New("source response too large")
	}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	var o authoritativeObservationResponse
	if err := dec.Decode(&o); err != nil {
		return AuthoritativeExecutionObservation{}, fmt.Errorf("decode source observation: %w", err)
	}
	if dec.Decode(new(any)) != io.EOF {
		return AuthoritativeExecutionObservation{}, errors.New("source response must be one JSON object")
	}
	return AuthoritativeExecutionObservation{
		ExecutionID: o.ExecutionID, Outcome: o.Outcome, EvidenceRef: o.EvidenceRef,
		Authority: o.Authority, EvidenceSHA256: o.EvidenceSHA256,
		ObservedAt: o.ObservedAt, ValidUntil: o.ValidUntil,
	}, nil
}
