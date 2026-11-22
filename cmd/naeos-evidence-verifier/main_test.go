// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

type readinessSource struct{}

func (readinessSource) ObserveExecution(context.Context, string, controlplane.ExecutionOutcome, string) (controlplane.AuthoritativeExecutionObservation, error) {
	return controlplane.AuthoritativeExecutionObservation{}, controlplane.ErrAuthoritativeEvidenceNotVerified
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	svc, err := controlplane.NewProviderNeutralEvidenceVerificationService(readinessSource{}, policy{version: "test-v1"}, "service-token")
	if err != nil {
		t.Fatal(err)
	}
	return newHandler(svc)
}

func TestHealthRoutesDoNotClaimDependencyReadiness(t *testing.T) {
	handler := newTestHandler(t)

	for _, tc := range []struct {
		name, path string
		want       int
	}{
		{name: "liveness", path: "/healthz", want: http.StatusOK},
		{name: "configuration", path: "/configz", want: http.StatusOK},
		{name: "dependency readiness remains unverified", path: "/readyz", want: http.StatusServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != tc.want {
				t.Fatalf("%s returned %d, want %d", tc.path, recorder.Code, tc.want)
			}
		})
	}
}

func TestVerifyRouteRemainsMountedAndRequiresPost(t *testing.T) {
	handler := newTestHandler(t)

	t.Run("GET is rejected with POST allowance", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/verify", nil))

		if recorder.Code != http.StatusMethodNotAllowed {
			t.Fatalf("GET /v1/verify returned %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
		}
		if got := recorder.Header().Get("Allow"); got != http.MethodPost {
			t.Fatalf("Allow header = %q, want %q", got, http.MethodPost)
		}
	})

	t.Run("POST reaches authentication guard", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/verify", nil))

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("unauthenticated POST /v1/verify returned %d, want %d", recorder.Code, http.StatusUnauthorized)
		}
	})
}
