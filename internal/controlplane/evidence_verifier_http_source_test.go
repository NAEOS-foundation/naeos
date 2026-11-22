// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPAuthoritativeEvidenceSourceRejectsUnsafeConfig(t *testing.T) {
	for _, tc := range []struct{ url, token string }{
		{"http://host/observe", "secret"},
		{"https:///observe", "secret"},
		{"https://user:pass@host/observe", "secret"},
		{"https://host/observe?x=y", "secret"},
		{"https://host/observe", ""},
	} {
		if _, err := NewHTTPAuthoritativeEvidenceSource(tc.url, tc.token); err == nil {
			t.Errorf("expected rejection: %s", tc.url)
		}
	}
}

func TestHTTPAuthoritativeEvidenceSourceBoundReadOnlyRequest(t *testing.T) {
	now := time.Now().UTC()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.Header.Get("Authorization") != "Bearer source-token" || r.Header.Get("Cache-Control") != "no-store" {
			t.Errorf("unexpected headers/method")
		}
		if r.URL.Query().Get("execution_id") != "exec-1" || r.URL.Query().Get("expected_outcome") != "COMMITTED" || r.URL.Query().Get("evidence_ref_hint") != "hint://untrusted" {
			t.Errorf("bad query %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(authoritativeObservationResponse{
			ExecutionID: "exec-1", Outcome: ExecutionOutcomeCommitted, EvidenceRef: "evidence://provider/1",
			Authority: "provider-api", EvidenceSHA256: strings.Repeat("d", 64),
			ObservedAt: now.Add(-time.Minute), ValidUntil: now.Add(time.Minute),
		})
	}))
	defer srv.Close()
	src, err := NewHTTPAuthoritativeEvidenceSource(srv.URL+"/observe", "source-token")
	if err != nil {
		t.Fatal(err)
	}
	src.client = srv.Client()
	got, err := src.ObserveExecution(context.Background(), "exec-1", ExecutionOutcomeCommitted, "hint://untrusted")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecutionID != "exec-1" || got.Outcome != ExecutionOutcomeCommitted || got.Authority != "provider-api" {
		t.Fatalf("got %+v", got)
	}
}

func TestHTTPAuthoritativeEvidenceSourceRejectsBadResponses(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
	}{
		{404, ""},
		{503, ""},
		{200, `{"execution_id":"exec","unexpected":true}`},
		{200, `{}{}`},
	} {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		}))
		src, _ := NewHTTPAuthoritativeEvidenceSource(srv.URL, "token")
		src.client = srv.Client()
		_, err := src.ObserveExecution(context.Background(), "exec", ExecutionOutcomeCommitted, "")
		srv.Close()
		if err == nil {
			t.Errorf("status %d body %s expected fail closed", tc.status, tc.body)
		}
	}
}
