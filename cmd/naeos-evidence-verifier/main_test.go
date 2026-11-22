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

func TestHealthRoutesDoNotClaimDependencyReadiness(t *testing.T) {
	svc, err := controlplane.NewProviderNeutralEvidenceVerificationService(readinessSource{}, policy{version: "test-v1"}, "service-token")
	if err != nil { t.Fatal(err) }
	handler := newHandler(svc)

	for _, tc := range []struct{name, path string; want int}{
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
