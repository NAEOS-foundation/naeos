package investordemo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestAPI(t *testing.T) (*APIServer, *DemoSetup) {
	t.Helper()
	setup := SetupDemoEnvironment()
	return NewAPIServer(setup), setup
}

func doJSON(t *testing.T, as *APIServer, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	as.ServeHTTP(rec, req)
	return rec
}

func TestAPIHealth(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodGet, "/api/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "healthy" {
		t.Errorf("expected healthy status, got %s", resp["status"])
	}
}

func TestAPIOptionsPreflight(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodOptions, "/api/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS allow origin *")
	}
}

func TestAPIAuthorize(t *testing.T) {
	as, _ := newTestAPI(t)

	rec := doJSON(t, as, http.MethodPost, "/api/authorize", `{"agent_id":"agent-payment-01","capability":"repository.read"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var d AuthorizationDecision
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if !d.Authorized {
		t.Errorf("expected repository.read authorized for agent-payment-01: %s", d.Reason)
	}

	rec = doJSON(t, as, http.MethodPost, "/api/authorize", `{"agent_id":"agent-payment-01","capability":"credential.rotate"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d.Authorized {
		t.Errorf("expected credential.rotate to be denied")
	}
}

func TestAPIAuthorizeErrors(t *testing.T) {
	as, _ := newTestAPI(t)

	rec := doJSON(t, as, http.MethodGet, "/api/authorize", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET, got %d", rec.Code)
	}

	rec = doJSON(t, as, http.MethodPost, "/api/authorize", `{invalid`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}

func TestAPIExecute(t *testing.T) {
	as, setup := newTestAPI(t)

	rec := doJSON(t, as, http.MethodPost, "/api/execute", `{"agent_id":"agent-payment-01","capability":"repository.write","payload":{"file":"a.go"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var res ExecutionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if !res.Authorized || !res.Executed {
		t.Errorf("expected authorized execution, got %+v", res)
	}
	if len(setup.AuditLedger.GetEvents()) == 0 {
		t.Error("expected audit events after execution")
	}

	rec = doJSON(t, as, http.MethodPost, "/api/execute", `{"agent_id":"agent-payment-01","capability":"production.deploy","payload":{}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Authorized {
		t.Error("expected production.deploy to be blocked")
	}
}

func TestAPIExecuteErrors(t *testing.T) {
	as, _ := newTestAPI(t)

	rec := doJSON(t, as, http.MethodPut, "/api/execute", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	rec = doJSON(t, as, http.MethodPost, "/api/execute", `{bad`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAPIAuditEvents(t *testing.T) {
	as, setup := newTestAPI(t)
	_ = setup.AuditLedger.RecordEvent(&AuditEvent{
		EventID: "AUD-X", Timestamp: time.Now(), EventType: "TEST", AgentID: "a-1",
	})

	rec := doJSON(t, as, http.MethodGet, "/api/audit", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Events []AuditEvent `json:"events"`
		Total  int          `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total < 1 {
		t.Errorf("expected at least 1 audit event, got %d", resp.Total)
	}
}

func TestAPIRunScenarios(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/scenarios", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Scenarios []ScenarioResult `json:"scenarios"`
		Total     int              `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total != 6 {
		t.Errorf("expected 6 scenarios, got %d", resp.Total)
	}

	rec = doJSON(t, as, http.MethodGet, "/api/scenarios", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestAPIGetPolicy(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodGet, "/api/policy", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var p Policy
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.PolicyID != "POLICY-017" {
		t.Errorf("expected POLICY-017, got %s", p.PolicyID)
	}
	if p.Version != 17 {
		t.Errorf("expected version 17, got %d", p.Version)
	}
}

func TestAPIGetGrants(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodGet, "/api/grants", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Grants []*CapabilityGrant `json:"grants"`
		Total  int                `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total < 1 {
		t.Errorf("expected at least 1 grant, got %d", resp.Total)
	}
}

func TestAPIVerification(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/verification", `{"agent_id":"agent-payment-01"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var v VerificationResult
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Result != "PASS" {
		t.Errorf("expected PASS for clean session, got %s", v.Result)
	}

	rec = doJSON(t, as, http.MethodGet, "/api/verification", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	rec = doJSON(t, as, http.MethodPost, "/api/verification", `{no`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAPIReset(t *testing.T) {
	as, _ := newTestAPI(t)

	// Dirty the state via a blocked execution + full demo run.
	doJSON(t, as, http.MethodPost, "/api/execute", `{"agent_id":"agent-payment-01","capability":"credential.rotate","payload":{}}`)
	doJSON(t, as, http.MethodPost, "/api/investor-demo", "")

	rec := doJSON(t, as, http.MethodPost, "/api/reset", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "reset" {
		t.Errorf("expected reset status, got %s", resp["status"])
	}

	// Reset restores policy v17 on a fresh environment.
	p, err := as.setup.PolicyStore.GetPolicy("POLICY-017")
	if err != nil {
		t.Fatal(err)
	}
	if p.Version != 17 {
		t.Errorf("expected reset to policy v17, got %d", p.Version)
	}
	if len(as.setup.AuditLedger.GetEvents()) != 0 {
		t.Errorf("expected fresh audit ledger after reset")
	}

	rec = doJSON(t, as, http.MethodGet, "/api/reset", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestAPIInvestorDemo(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/investor-demo", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Demo    string                 `json:"demo"`
		Steps   []DemoStep             `json:"steps"`
		Total   int                    `json:"total"`
		Posture map[string]interface{} `json:"posture"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total != 10 {
		t.Errorf("expected 10 demo steps, got %d", resp.Total)
	}
	if strings.Contains(resp.Demo, "NAEOS") == false {
		t.Errorf("expected NAEOS in demo title")
	}
	if resp.Posture["replay_attempts"].(float64) < 1 {
		t.Errorf("expected at least 1 replay attempt recorded")
	}

	rec = doJSON(t, as, http.MethodGet, "/api/investor-demo", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestAPIUnknownRoute(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodGet, "/api/does-not-exist", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
