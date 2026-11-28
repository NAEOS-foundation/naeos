// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package authboundary

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequirePrincipalFailClosed(t *testing.T) {
	called := false
	protected := RequirePrincipal(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatal("protected handler executed without principal")
	}
}

func TestRequirePrincipalAllowsTrustedContext(t *testing.T) {
	protected := RequirePrincipal(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok || principal.Subject != "test-subject" {
			t.Errorf("unexpected principal: %+v, ok=%v", principal, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(WithPrincipal(req.Context(), Principal{Subject: "test-subject"}))
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
