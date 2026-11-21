// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/api"
)

type reconciliationTestIdentity struct {
	claims *api.JWTClaims
	err    error
}

func (i reconciliationTestIdentity) Validate(context.Context, string) (*api.JWTClaims, error) {
	return i.claims, i.err
}

type reconciliationTestEvidenceVerifier struct {
	ref string
	err error
}

func (v reconciliationTestEvidenceVerifier) VerifyExecutionEvidence(context.Context, string, ExecutionOutcome, string) (string, error) {
	return v.ref, v.err
}

type reconciliationTestRepository struct {
	calls       int
	executionID string
	outcome     ExecutionOutcome
	evidenceRef string
	reconciler  string
	err         error
}

func (r *reconciliationTestRepository) ReconcileExecution(_ context.Context, executionID string, outcome ExecutionOutcome, evidenceRef, reconciler string, _ time.Time) (LedgerEvent, int64, error) {
	r.calls++
	r.executionID, r.outcome, r.evidenceRef, r.reconciler = executionID, outcome, evidenceRef, reconciler
	return LedgerEvent{ID: "event-reconciled", ExecutionID: executionID, EventType: "EXECUTION_RECONCILED"}, 42, r.err
}

func TestProductionReconciliationRequiresDedicatedScope(t *testing.T) {
	repository := &reconciliationTestRepository{}
	runtime := &ProductionRuntime{
		identity:          reconciliationTestIdentity{claims: &api.JWTClaims{Sub: "operator-1", Scope: "naeos:read"}},
		reconciler:        repository,
		evidenceVerifier: reconciliationTestEvidenceVerifier{ref: "evidence://verified/1"},
	}
	handler := runtime.reconciliationHandler()
	request := httptest.NewRequest(http.MethodPost, "/api/control-plane/executions/exec-1/reconcile", strings.NewReader(`{"outcome":"COMMITTED","evidence_ref":"evidence://submitted/1"}`))
	request.Header.Set("Authorization", "Bearer valid")
	request.SetPathValue("executionID", "exec-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d; want %d", response.Code, http.StatusForbidden)
	}
	if repository.calls != 0 {
		t.Fatal("repository must not be called without reconciliation scope")
	}
}

func TestProductionReconciliationUsesVerifiedPrincipalAndEvidence(t *testing.T) {
	repository := &reconciliationTestRepository{}
	runtime := &ProductionRuntime{
		identity:          reconciliationTestIdentity{claims: &api.JWTClaims{Sub: "operator-from-jwt", Scopes: []string{"naeos:execution:reconcile"}}},
		reconciler:        repository,
		evidenceVerifier: reconciliationTestEvidenceVerifier{ref: "evidence://immutable/sha256:abc"},
	}
	handler := runtime.reconciliationHandler()
	request := httptest.NewRequest(http.MethodPost, "/api/control-plane/executions/exec-9/reconcile", strings.NewReader(`{"outcome":"NOT_COMMITTED","evidence_ref":"evidence://submitted/ref"}`))
	request.Header.Set("Authorization", "Bearer valid")
	request.SetPathValue("executionID", "exec-9")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s; want %d", response.Code, response.Body.String(), http.StatusCreated)
	}
	if repository.calls != 1 || repository.executionID != "exec-9" ||
		repository.outcome != ExecutionOutcomeNotCommitted ||
		repository.evidenceRef != "evidence://immutable/sha256:abc" ||
		repository.reconciler != "operator-from-jwt" {
		t.Fatalf("repository did not receive verified principal/evidence: %+v", repository)
	}
}

func TestProductionReconciliationFailsClosedWhenEvidenceCannotBeVerified(t *testing.T) {
	repository := &reconciliationTestRepository{}
	runtime := &ProductionRuntime{
		identity:          reconciliationTestIdentity{claims: &api.JWTClaims{Sub: "operator-1", Scope: "naeos:execution:reconcile"}},
		reconciler:        repository,
		evidenceVerifier: reconciliationTestEvidenceVerifier{err: errors.New("evidence unavailable")},
	}
	handler := runtime.reconciliationHandler()
	request := httptest.NewRequest(http.MethodPost, "/api/control-plane/executions/exec-1/reconcile", strings.NewReader(`{"outcome":"COMMITTED","evidence_ref":"evidence://submitted/1"}`))
	request.Header.Set("Authorization", "Bearer valid")
	request.SetPathValue("executionID", "exec-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d; want %d", response.Code, http.StatusUnprocessableEntity)
	}
	if repository.calls != 0 {
		t.Fatal("repository must not be called when evidence verification fails")
	}
}

func TestProductionReconciliationRejectsUnknownRequestFields(t *testing.T) {
	repository := &reconciliationTestRepository{}
	runtime := &ProductionRuntime{
		identity:          reconciliationTestIdentity{claims: &api.JWTClaims{Sub: "operator-1", Scope: "naeos:execution:reconcile"}},
		reconciler:        repository,
		evidenceVerifier: reconciliationTestEvidenceVerifier{ref: "evidence://verified/1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/control-plane/executions/exec-1/reconcile", strings.NewReader(`{"outcome":"COMMITTED","evidence_ref":"evidence://submitted/1","reconciler_id":"spoofed"}`))
	request.Header.Set("Authorization", "Bearer valid")
	request.SetPathValue("executionID", "exec-1")
	response := httptest.NewRecorder()
	runtime.reconciliationHandler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d; want %d", response.Code, http.StatusBadRequest)
	}
	if repository.calls != 0 {
		t.Fatal("repository must not be called for a request with unknown fields")
	}
}
