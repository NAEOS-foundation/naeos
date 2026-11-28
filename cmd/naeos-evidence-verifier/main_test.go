// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

type routeTestSource struct{}

func (routeTestSource) ObserveExecution(context.Context, string, controlplane.ExecutionOutcome, string) (controlplane.AuthoritativeExecutionObservation, error) {
	return controlplane.AuthoritativeExecutionObservation{}, controlplane.ErrAuthoritativeEvidenceNotVerified
}

type routeTestPolicy struct{}

func (routeTestPolicy) Version() string { return "route-test-v1" }
func (routeTestPolicy) EvaluateExecutionEvidence(context.Context, controlplane.AuthoritativeExecutionObservation, controlplane.ExecutionOutcome) error {
	return nil
}

type routeTestGuard struct {
	ready bool
	readyErr error
}

func (routeTestGuard) Claim(context.Context, string, string) (bool, error) { return true, nil }
func (g routeTestGuard) ClaimReadiness(context.Context) (bool, error) { return g.ready, g.readyErr }

func newRouteTestHandler(t *testing.T, readyErr error) http.Handler {
	t.Helper()
	service, err := controlplane.NewProviderNeutralEvidenceVerificationService(routeTestSource{}, routeTestPolicy{}, routeTestGuard{}, "test-scope", "test-token")
	if err != nil {
		t.Fatal(err)
	}
	return newHandler(service, routeTestGuard{ready: readyErr == nil, readyErr: readyErr})
}

func TestVerifierHealthConfigReadinessRoutes(t *testing.T) {
	for _, tc := range []struct {
		path string
		want int
	}{
		{"/healthz", http.StatusOK},
		{"/configz", http.StatusOK},
		{"/readyz", http.StatusOK},
	} {
		t.Run(tc.path, func(t *testing.T) {
			handler := newRouteTestHandler(t, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != tc.want {
				t.Fatalf("%s returned %d, want %d", tc.path, recorder.Code, tc.want)
			}
		})
	}
}

func TestVerifierHealthConfigReadinessRejectNonGet(t *testing.T) {
	for _, path := range []string{"/healthz", "/configz", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			handler := newRouteTestHandler(t, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, nil))
			if recorder.Code != http.StatusMethodNotAllowed {
				t.Fatalf("POST %s returned %d, want 405", path, recorder.Code)
			}
			if got := recorder.Header().Get("Allow"); got != http.MethodGet {
				t.Fatalf("Allow header %q, want GET", got)
			}
		})
	}
}

func TestVerifierReadinessFailsClosed(t *testing.T) {
	handler := newRouteTestHandler(t, errors.New("database probe failed"))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness returned %d, want 503", recorder.Code)
	}
}

func TestVerifierReadinessFalseFailsClosed(t *testing.T) {
	service, err := controlplane.NewProviderNeutralEvidenceVerificationService(routeTestSource{}, routeTestPolicy{}, routeTestGuard{}, "test-scope", "test-token")
	if err != nil { t.Fatal(err) }
	handler := newHandler(service, routeTestGuard{ready: false})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if recorder.Code != http.StatusServiceUnavailable { t.Fatalf("readiness returned %d, want 503", recorder.Code) }
}

func TestVerifierNilServiceFailsClosed(t *testing.T) {
	handler := newHandler(nil, routeTestGuard{})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/verify", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil service returned %d, want 503", recorder.Code)
	}
}
