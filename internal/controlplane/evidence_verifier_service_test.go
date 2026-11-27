// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type evidenceSourceStub struct {
	observation AuthoritativeExecutionObservation
	err         error
	calls       int
	executionID string
	outcome     ExecutionOutcome
	hint        string
}

func (s *evidenceSourceStub) ObserveExecution(_ context.Context, executionID string, outcome ExecutionOutcome, hint string) (AuthoritativeExecutionObservation, error) {
	s.calls++
	s.executionID, s.outcome, s.hint = executionID, outcome, hint
	return s.observation, s.err
}

type evidencePolicyStub struct {
	err   error
	calls int
}

func (p *evidencePolicyStub) Version() string { return "service-policy-v1" }

func (p *evidencePolicyStub) EvaluateExecutionEvidence(context.Context, AuthoritativeExecutionObservation, ExecutionOutcome) error {
	p.calls++
	return p.err
}

func validEvidenceObservation(executionID string, outcome ExecutionOutcome) AuthoritativeExecutionObservation {
	now := time.Now().UTC()
	return AuthoritativeExecutionObservation{
		ExecutionID: executionID, Outcome: outcome, EvidenceRef: "evidence://provider/receipt/1",
		Authority: "provider-api", EvidenceSHA256: strings.Repeat("b", 64),
		ObservedAt: now.Add(-time.Minute), ValidUntil: now.Add(time.Minute),
	}
}

func TestProviderNeutralEvidenceVerificationServiceRequiresDependencies(t *testing.T) {
	policy := &evidencePolicyStub{}
	source := &evidenceSourceStub{}
	for _, test := range []struct {
		name   string
		source AuthoritativeExecutionEvidenceSource
		policy ExecutionEvidencePolicy
		token  string
		guard  EvidenceVerifierReplayGuard
		scope  string
	}{
		{name: "missing source", policy: policy, token: "secret", guard: &memoryReplayGuard{claims: map[string]bool{}}, scope: "test"},
		{name: "missing policy", source: source, token: "secret", guard: &memoryReplayGuard{claims: map[string]bool{}}, scope: "test"},
		{name: "missing token", source: source, policy: policy, guard: &memoryReplayGuard{claims: map[string]bool{}}, scope: "test"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewProviderNeutralEvidenceVerificationService(test.source, test.policy, test.guard, test.scope, test.token); err == nil {
				t.Fatal("expected missing dependency to fail closed")
			}
		})
	}
}

func TestProviderNeutralEvidenceVerificationServiceReturnsBoundReceipt(t *testing.T) {
	source := &evidenceSourceStub{observation: validEvidenceObservation("execution-1", ExecutionOutcomeCommitted)}
	policy := &evidencePolicyStub{}
	service, err := NewProviderNeutralEvidenceVerificationService(source, policy, &memoryReplayGuard{claims: map[string]bool{}}, "test", "service-secret")
	if err != nil {
		t.Fatal(err)
	}
	requestBody := evidenceVerificationRequest{
		RequestID: strings.Repeat("a", 32), ExecutionID: "execution-1",
		Outcome: ExecutionOutcomeCommitted, EvidenceRef: "untrusted://hint/1",
	}
	body, _ := json.Marshal(requestBody)
	request := httptest.NewRequest(http.MethodPost, "/v1/verify", strings.NewReader(string(body)))
	request.Header.Set("Authorization", "Bearer service-secret")
	response := httptest.NewRecorder()
	service.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var receipt evidenceVerificationResponse
	if err := json.Unmarshal(response.Body.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.RequestID != requestBody.RequestID || receipt.ExecutionID != requestBody.ExecutionID ||
		receipt.Outcome != requestBody.Outcome || receipt.Status != "VERIFIED" ||
		receipt.EvidenceRef != source.observation.EvidenceRef || receipt.Authority != source.observation.Authority {
		t.Fatalf("unexpected verification receipt: %+v", receipt)
	}
	if source.calls != 1 || source.hint != requestBody.EvidenceRef || policy.calls != 1 {
		t.Fatalf("source/policy invocation mismatch: source=%+v policy=%+v", source, policy)
	}
}

func TestProviderNeutralEvidenceVerificationServiceFailsClosed(t *testing.T) {
	tests := []struct {
		name        string
		observation AuthoritativeExecutionObservation
		sourceErr   error
		policyErr   error
		wantStatus  int
	}{
		{name: "provider unavailable", observation: validEvidenceObservation("execution-1", ExecutionOutcomeCommitted), sourceErr: errors.New("provider unavailable"), wantStatus: http.StatusServiceUnavailable},
		{name: "provider unknown", observation: validEvidenceObservation("execution-1", ExecutionOutcomeCommitted), sourceErr: ErrAuthoritativeEvidenceNotVerified, wantStatus: http.StatusUnprocessableEntity},
		{name: "wrong execution binding", observation: validEvidenceObservation("other-execution", ExecutionOutcomeCommitted), wantStatus: http.StatusUnprocessableEntity},
		{name: "wrong outcome", observation: validEvidenceObservation("execution-1", ExecutionOutcomeNotCommitted), wantStatus: http.StatusUnprocessableEntity},
		{name: "expired observation", observation: func() AuthoritativeExecutionObservation {
			o := validEvidenceObservation("execution-1", ExecutionOutcomeCommitted)
			o.ValidUntil = time.Now().UTC().Add(-time.Second)
			return o
		}(), wantStatus: http.StatusUnprocessableEntity},
		{name: "policy rejects", observation: validEvidenceObservation("execution-1", ExecutionOutcomeCommitted), policyErr: errors.New("stale or contradictory evidence"), wantStatus: http.StatusUnprocessableEntity},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := &evidenceSourceStub{observation: test.observation, err: test.sourceErr}
			policy := &evidencePolicyStub{err: test.policyErr}
			service, err := NewProviderNeutralEvidenceVerificationService(source, policy, &memoryReplayGuard{claims: map[string]bool{}}, "test", "service-secret")
			if err != nil {
				t.Fatal(err)
			}
			body, _ := json.Marshal(evidenceVerificationRequest{
				RequestID: strings.Repeat("c", 32), ExecutionID: "execution-1", Outcome: ExecutionOutcomeCommitted,
			})
			request := httptest.NewRequest(http.MethodPost, "/v1/verify", strings.NewReader(string(body)))
			request.Header.Set("Authorization", "Bearer service-secret")
			response := httptest.NewRecorder()
			service.Handler().ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d body=%s; want %d", response.Code, response.Body.String(), test.wantStatus)
			}
			if response.Code != http.StatusOK && strings.Contains(response.Body.String(), `"status":"VERIFIED"`) {
				t.Fatal("non-confirming observation must not produce VERIFIED")
			}
		})
	}
}

func TestProviderNeutralEvidenceVerificationServiceRejectsUnauthorizedAndMalformedRequests(t *testing.T) {
	source := &evidenceSourceStub{observation: validEvidenceObservation("execution-1", ExecutionOutcomeCommitted)}
	policy := &evidencePolicyStub{}
	service, err := NewProviderNeutralEvidenceVerificationService(source, policy, &memoryReplayGuard{claims: map[string]bool{}}, "test", "service-secret")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		body       string
		auth       string
		wantStatus int
	}{
		{name: "missing auth", body: `{}`, wantStatus: http.StatusUnauthorized},
		{name: "bad credential", body: `{}`, auth: "Bearer wrong", wantStatus: http.StatusUnauthorized},
		{name: "unknown field", body: `{"request_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","execution_id":"execution-1","expected_outcome":"COMMITTED","unexpected":true}`, auth: "Bearer service-secret", wantStatus: http.StatusBadRequest},
		{name: "invalid request id", body: `{"request_id":"bad","execution_id":"execution-1","expected_outcome":"COMMITTED"}`, auth: "Bearer service-secret", wantStatus: http.StatusBadRequest},
		{name: "invalid outcome", body: `{"request_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","execution_id":"execution-1","expected_outcome":"UNKNOWN"}`, auth: "Bearer service-secret", wantStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/v1/verify", strings.NewReader(test.body))
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			service.Handler().ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d body=%s; want %d", response.Code, response.Body.String(), test.wantStatus)
			}
		})
	}
	if source.calls != 0 {
		t.Fatal("source must not be queried for unauthorized or malformed requests")
	}
}
