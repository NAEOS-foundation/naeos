// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPExecutionEvidenceVerifierRequiresHTTPSAndCredential(t *testing.T) {
	for _, test := range []struct {
		name, endpoint, token string
	}{
		{name: "plain HTTP", endpoint: "http://verifier.example/v1/verify", token: "secret"},
		{name: "missing host", endpoint: "https:///v1/verify", token: "secret"},
		{name: "userinfo", endpoint: "https://user:pass@verifier.example/v1/verify", token: "secret"},
		{name: "query", endpoint: "https://verifier.example/v1/verify?debug=true", token: "secret"},
		{name: "missing token", endpoint: "https://verifier.example/v1/verify"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewHTTPExecutionEvidenceVerifier(test.endpoint, test.token); err == nil {
				t.Fatal("expected invalid verifier configuration to fail")
			}
		})
	}
}

func TestHTTPExecutionEvidenceVerifierAcceptsBoundFreshReceipt(t *testing.T) {
	var received evidenceVerificationRequest
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer test-service-token" ||
			r.Header.Get("Cache-Control") != "no-store" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		writeVerificationReceipt(t, w, received, "VERIFIED", "evidence://provider/receipt/123", time.Now().UTC().Add(-time.Minute), time.Now().UTC().Add(time.Minute))
	}))
	defer server.Close()

	verifier, err := newHTTPExecutionEvidenceVerifier(server.URL, "test-service-token", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	ref, err := verifier.VerifyExecutionEvidence(context.Background(), "execution-123", ExecutionOutcomeCommitted, "untrusted://submitted/hint")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "evidence://provider/receipt/123" {
		t.Fatalf("verified reference=%q", ref)
	}
	if received.ExecutionID != "execution-123" || received.Outcome != ExecutionOutcomeCommitted ||
		received.EvidenceRef != "untrusted://submitted/hint" || received.RequestID == "" {
		t.Fatalf("unexpected request contract: %+v", received)
	}
}

func TestHTTPExecutionEvidenceVerifierRejectsUnboundOrStaleReceipts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*evidenceVerificationResponse)
	}{
		{name: "wrong execution", mutate: func(r *evidenceVerificationResponse) { r.ExecutionID = "other-execution" }},
		{name: "wrong outcome", mutate: func(r *evidenceVerificationResponse) { r.Outcome = ExecutionOutcomeNotCommitted }},
		{name: "not verified", mutate: func(r *evidenceVerificationResponse) { r.Status = "UNKNOWN" }},
		{name: "missing authority", mutate: func(r *evidenceVerificationResponse) { r.Authority = "" }},
		{name: "invalid digest", mutate: func(r *evidenceVerificationResponse) { r.EvidenceSHA256 = "not-a-digest" }},
		{name: "invalid reference", mutate: func(r *evidenceVerificationResponse) { r.EvidenceRef = "relative/ref" }},
		{name: "stale receipt", mutate: func(r *evidenceVerificationResponse) { r.ValidUntil = time.Now().UTC().Add(-time.Second) }},
		{name: "future observation", mutate: func(r *evidenceVerificationResponse) { r.ObservedAt = time.Now().UTC().Add(time.Hour) }},
		{name: "missing policy version", mutate: func(r *evidenceVerificationResponse) { r.VerificationPolicyVersion = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request evidenceVerificationRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					http.Error(w, "invalid request", http.StatusBadRequest)
					return
				}
				now := time.Now().UTC()
				result := evidenceVerificationResponse{
					RequestID: request.RequestID, ExecutionID: request.ExecutionID, Outcome: request.Outcome,
					Status: "VERIFIED", EvidenceRef: "evidence://provider/receipt/123", Authority: "provider-api",
					EvidenceSHA256: strings.Repeat("a", 64), ObservedAt: now.Add(-time.Minute),
					ValidUntil: now.Add(time.Minute), VerificationPolicyVersion: "provider-policy-v1",
				}
				test.mutate(&result)
				_ = json.NewEncoder(w).Encode(result)
			}))
			defer server.Close()
			verifier, err := newHTTPExecutionEvidenceVerifier(server.URL, "test-token", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			if ref, err := verifier.VerifyExecutionEvidence(context.Background(), "execution-123", ExecutionOutcomeCommitted, ""); err == nil || ref != "" {
				t.Fatalf("expected fail-closed receipt rejection, ref=%q err=%v", ref, err)
			}
		})
	}
}

func TestHTTPExecutionEvidenceVerifierRejectsUnknownFieldsAndRedirects(t *testing.T) {
	t.Run("unknown response field", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"unexpected":true}`))
		}))
		defer server.Close()
		verifier, _ := newHTTPExecutionEvidenceVerifier(server.URL, "test-token", server.Client())
		if _, err := verifier.VerifyExecutionEvidence(context.Background(), "exec", ExecutionOutcomeCommitted, ""); err == nil {
			t.Fatal("expected strict response decoding to reject unknown fields")
		}
	})
	t.Run("redirect", func(t *testing.T) {
		targetCalled := false
		target := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { targetCalled = true }))
		defer target.Close()
		source := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target.URL, http.StatusFound)
		}))
		defer source.Close()
		client := source.Client()
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		verifier, _ := newHTTPExecutionEvidenceVerifier(source.URL, "test-token", client)
		if _, err := verifier.VerifyExecutionEvidence(context.Background(), "exec", ExecutionOutcomeCommitted, ""); err == nil {
			t.Fatal("expected redirect to be rejected")
		}
		if targetCalled {
			t.Fatal("verifier must not follow redirects")
		}
	})
}

func writeVerificationReceipt(t *testing.T, w http.ResponseWriter, request evidenceVerificationRequest, status, ref string, observedAt, validUntil time.Time) {
	t.Helper()
	digest := sha256.Sum256([]byte("provider-authoritative-evidence"))
	_ = json.NewEncoder(w).Encode(evidenceVerificationResponse{
		RequestID: request.RequestID, ExecutionID: request.ExecutionID, Outcome: request.Outcome,
		Status: status, EvidenceRef: ref, Authority: "provider-api",
		EvidenceSHA256: hex.EncodeToString(digest[:]), ObservedAt: observedAt,
		ValidUntil: validUntil, VerificationPolicyVersion: "provider-policy-v1",
	})
}
