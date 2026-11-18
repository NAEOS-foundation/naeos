// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductionHealthHandlerIsNarrowAndFailClosed(t *testing.T) {
	var runtime *ProductionRuntime
	handler := runtime.ProductionHealthHandler()

	live := httptest.NewRecorder()
	handler.ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("liveness status = %d, want %d", live.Code, http.StatusOK)
	}

	ready := httptest.NewRecorder()
	handler.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil runtime readiness status = %d, want %d", ready.Code, http.StatusServiceUnavailable)
	}

	for _, path := range []string{
		"/api/control-plane/decision",
		"/api/investor-demo",
		"/api/scenarios",
		"/api/reset",
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("production health handler exposed %q: status %d, want 404", path, response.Code)
		}
	}
}

func TestProductionHealthHandlerRejectsWrongMethods(t *testing.T) {
	var runtime *ProductionRuntime
	handler := runtime.ProductionHealthHandler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/health/live", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST liveness status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}
