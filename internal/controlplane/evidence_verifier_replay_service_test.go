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
)

func invokeReplayGuardTestHandler(t *testing.T, svc *ProviderNeutralEvidenceVerificationService, id string) int {
	t.Helper()
	payload, _ := json.Marshal(evidenceVerificationRequest{RequestID: id, ExecutionID: "exec-1", Outcome: ExecutionOutcomeCommitted})
	req := httptest.NewRequest(http.MethodPost, "/v1/verify", strings.NewReader(string(payload)))
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	svc.Handler().ServeHTTP(res, req)
	return res.Code
}

type replayGuardOutcome struct {
	claimed bool
	err     error
	calls   int
}

func (g *replayGuardOutcome) Claim(context.Context, string, string) (bool, error) {
	g.calls++
	return g.claimed, g.err
}
func TestEvidenceVerifierDuplicateClaimNeverObservesSource(t *testing.T) {
	src := &evidenceSourceStub{observation: validEvidenceObservation("exec-1", ExecutionOutcomeCommitted)}
	policy := &evidencePolicyStub{}
	guard := &replayGuardOutcome{claimed: false}
	svc, err := NewProviderNeutralEvidenceVerificationService(src, policy, guard, "scope", "test-token")
	if err != nil {
		t.Fatal(err)
	}
	status := invokeReplayGuardTestHandler(t, svc, strings.Repeat("A", 32))
	if status != http.StatusConflict {
		t.Fatalf("status=%d expected 409", status)
	}
	if src.calls != 0 || policy.calls != 0 {
		t.Fatalf("source/policy called on duplicate: source=%d policy=%d", src.calls, policy.calls)
	}
	if guard.calls != 1 {
		t.Fatalf("guard calls=%d", guard.calls)
	}
}
func TestEvidenceVerifierReplayStorageErrorFailClosedBeforeObservation(t *testing.T) {
	src := &evidenceSourceStub{observation: validEvidenceObservation("exec-1", ExecutionOutcomeCommitted)}
	guard := &replayGuardOutcome{err: errors.New("storage unavailable")}
	svc, err := NewProviderNeutralEvidenceVerificationService(src, &evidencePolicyStub{}, guard, "scope", "test-token")
	if err != nil {
		t.Fatal(err)
	}
	status := invokeReplayGuardTestHandler(t, svc, strings.Repeat("b", 32))
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status=%d expected 503", status)
	}
	if src.calls != 0 {
		t.Fatalf("source called %d times after storage failure", src.calls)
	}
}
