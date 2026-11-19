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

type productionReadLedgerStub struct {
	events []LedgerEvent
	err error
}

func (l productionReadLedgerStub) Load(ctx context.Context) ([]LedgerEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if l.err != nil {
		return nil, l.err
	}
	return append([]LedgerEvent(nil), l.events...), nil
}

func readRequest(handler http.Handler, token, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func TestControlPlaneReadAPIRequiresExplicitScope(t *testing.T) {
	runtime := &ProductionRuntime{
		identity: productionTokenValidatorStub{claims: &api.JWTClaims{Sub: "operator", Scope: "crm:read"}},
		ledgerReader: productionReadLedgerStub{},
	}
	response := readRequest(runtime.ProductionHandler(), "valid", "/api/control-plane/overview")
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want 403; body=%s", response.Code, response.Body.String())
	}
}

func TestControlPlaneReadAPIRequiresAuthentication(t *testing.T) {
	runtime := &ProductionRuntime{
		identity: productionTokenValidatorStub{claims: &api.JWTClaims{Sub: "operator", Scopes: []string{controlPlaneReadScope}}},
		ledgerReader: productionReadLedgerStub{},
	}
	response := readRequest(runtime.ProductionHandler(), "", "/api/control-plane/overview")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", response.Code)
	}
}

func TestControlPlaneOverviewUsesVerifiedLedgerAndCountsDecisions(t *testing.T) {
	runtime := &ProductionRuntime{
		environment: "staging",
		identity: productionTokenValidatorStub{claims: &api.JWTClaims{Sub: "operator", Scopes: []string{controlPlaneReadScope}}},
		ledgerReader: productionReadLedgerStub{events: []LedgerEvent{
			{ID: "event-1", EventType: "AUTHORIZATION_DECISION", Decision: DecisionAllow},
			{ID: "event-2", EventType: "POLICY_EVALUATION_FAILED"},
			{ID: "event-3", EventType: "AUTHORIZATION_DECISION", Decision: DecisionDeny},
			{ID: "event-4", EventType: "AUTHORIZATION_DECISION", Decision: DecisionPending},
		}},
	}
	response := readRequest(runtime.ProductionHandler(), "valid", "/api/control-plane/overview")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200; body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`"environment":"staging"`, `"status":"healthy"`, `"last_sequence":4`, `"decision_count":3`, `"allow_count":1`, `"deny_count":1`, `"pending_count":1`} {
		if !contains(body, expected) {
			t.Errorf("response %q missing %q", body, expected)
		}
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q; want no-store", got)
	}
}

func TestControlPlaneDecisionsAreBoundedAndDoNotExposeMetadata(t *testing.T) {
	runtime := &ProductionRuntime{
		environment: "staging",
		identity: productionTokenValidatorStub{claims: &api.JWTClaims{Sub: "operator", Scope: controlPlaneReadScope}},
		ledgerReader: productionReadLedgerStub{events: []LedgerEvent{
			{ID: "event-1", EventType: "AUTHORIZATION_DECISION", DecisionID: "decision-1", AgentID: "agent-1", Metadata: map[string]string{"secret": "must-not-leak"}},
			{ID: "event-2", EventType: "POLICY_EVALUATION_FAILED"},
			{ID: "event-3", EventType: "AUTHORIZATION_DECISION", DecisionID: "decision-2", AgentID: "agent-2"},
		}},
	}
	response := readRequest(runtime.ProductionHandler(), "valid", "/api/control-plane/decisions?limit=1")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200; body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !contains(body, `"decision_id":"decision-1"`) || contains(body, "must-not-leak") || !contains(body, `"next_cursor":"1"`) {
		t.Fatalf("unexpected decision page: %s", body)
	}
	next := readRequest(runtime.ProductionHandler(), "valid", "/api/control-plane/decisions?limit=1&after=1")
	if next.Code != http.StatusOK || !contains(next.Body.String(), `"decision_id":"decision-2"`) {
		t.Fatalf("next page status=%d body=%s", next.Code, next.Body.String())
	}
}

func TestControlPlaneDecisionsRejectInvalidLimit(t *testing.T) {
	runtime := &ProductionRuntime{
		identity: productionTokenValidatorStub{claims: &api.JWTClaims{Sub: "operator", Scope: controlPlaneReadScope}},
		ledgerReader: productionReadLedgerStub{},
	}
	for _, query := range []string{"limit=0", "limit=101", "limit=abc", "after=-1", "after=bad"} {
		response := readRequest(runtime.ProductionHandler(), "valid", "/api/control-plane/decisions?"+query)
		if response.Code != http.StatusBadRequest {
			t.Errorf("query %q status=%d; want 400", query, response.Code)
		}
	}
}

func TestControlPlaneReadAPIFailsClosedWhenLedgerUnavailable(t *testing.T) {
	runtime := &ProductionRuntime{
		identity: productionTokenValidatorStub{claims: &api.JWTClaims{Sub: "operator", Scope: controlPlaneReadScope}},
		ledgerReader: productionReadLedgerStub{err: errors.New("database unavailable")},
	}
	response := readRequest(runtime.ProductionHandler(), "valid", "/api/control-plane/overview")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d; want 503", response.Code)
	}
	if response.Code == http.StatusOK && contains(response.Body.String(), `"decision_count":0`) {
		t.Fatal("ledger failure must not be presented as an empty successful dataset")
	}
}

func TestControlPlaneReadAPIResponseHasObservationTime(t *testing.T) {
	runtime := &ProductionRuntime{
		identity: productionTokenValidatorStub{claims: &api.JWTClaims{Sub: "operator", Scope: controlPlaneReadScope}},
		ledgerReader: productionReadLedgerStub{},
	}
	response := readRequest(runtime.ProductionHandler(), "valid", "/api/control-plane/overview")
	if response.Code != http.StatusOK || !contains(response.Body.String(), `"observed_at":"`) {
		t.Fatalf("status=%d body=%s; expected timestamped response", response.Code, response.Body.String())
	}
	if !time.Now().UTC().After(time.Time{}) {
		t.Fatal("unexpected clock state")
	}
}

func contains(value, fragment string) bool {
	return strings.Contains(value, fragment)
}
