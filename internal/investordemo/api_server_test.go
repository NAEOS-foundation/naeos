package investordemo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
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

func TestAPIControlPlaneDecision(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/control-plane/decision", `{"agent_id":"agent-payment-01","capability":"repository.read"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "ALLOW" {
		t.Errorf("expected ALLOW, got %v", resp["status"])
	}
	if resp["allowed"] != true {
		t.Errorf("expected allowed=true, got %v", resp["allowed"])
	}

	rec = doJSON(t, as, http.MethodPost, "/api/control-plane/decision", `{"agent_id":"agent-payment-01","capability":"credential.rotate"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "DENY" {
		t.Errorf("expected DENY, got %v", resp["status"])
	}
}

func TestAPIControlPlaneSession(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/control-plane/session", `{"agent_id":"agent-payment-01"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["result"] == nil {
		t.Fatal("expected result field in session summary")
	}
}

func TestAPIControlPlaneEvidence(t *testing.T) {
	as, _ := newTestAPI(t)
	doJSON(t, as, http.MethodPost, "/api/control-plane/decision", `{"agent_id":"agent-payment-01","capability":"repository.read"}`)

	rec := doJSON(t, as, http.MethodGet, "/api/control-plane/evidence", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Events []map[string]interface{} `json:"events"`
		Total  int                      `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || len(resp.Events) != 1 {
		t.Fatalf("expected one control-plane evidence event, got total=%d events=%d", resp.Total, len(resp.Events))
	}

	rec = doJSON(t, as, http.MethodPost, "/api/control-plane/evidence", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}

	rec = doJSON(t, as, http.MethodGet, "/api/control-plane/evidence?agent_id=agent-payment-01&event_type=AUTHORIZATION_DECISION", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected filtered evidence 200, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || resp.Events[0]["agent_id"] != "agent-payment-01" {
		t.Fatalf("unexpected filtered evidence: %+v", resp)
	}
}

func TestAPIControlPlaneApprovalLifecycle(t *testing.T) {
	as, setup := newTestAPI(t)
	setup.ControlPlaneGateway.Ledger.Append(controlplane.LedgerEvent{
		DecisionID:   "DEC-API-1",
		AgentID:      "agent-payment-01",
		Capability:   "production.deploy",
		ArtifactHash: "sha256:test",
		EventType:    "AUTHORIZATION_DECISION",
		Decision:     controlplane.DecisionPending,
		Metadata:     map[string]string{"policy_id": "POLICY-017", "policy_version": "17"},
	})
	rec := doJSON(t, as, http.MethodPost, "/api/control-plane/approval",
		`{"decision_id":"DEC-API-1","approver":"reviewer-1","expires_in_seconds":60}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var approval struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &approval); err != nil {
		t.Fatal(err)
	}
	if approval.ID == "" {
		t.Fatal("expected approval ID")
	}
	rec = doJSON(t, as, http.MethodGet, "/api/control-plane/approval/"+approval.ID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected approval status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Status != "pending" {
		t.Fatalf("expected pending status, got %s", status.Status)
	}
	rec = doJSON(t, as, http.MethodPost, "/api/control-plane/approval/approve",
		`{"approval_id":"`+approval.ID+`","reason":"reviewed","artifact_hash":"sha256:test"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, as, http.MethodGet, "/api/control-plane/approval/"+approval.ID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected approved status 200, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Status != "approved" {
		t.Fatalf("expected approved status, got %s", status.Status)
	}

	as, setup = newTestAPI(t)
	setup.ControlPlaneGateway.Ledger.Append(controlplane.LedgerEvent{
		DecisionID: "DEC-API-REJECT", AgentID: "agent-payment-01",
		Capability: "production.deploy", EventType: "AUTHORIZATION_DECISION",
		Decision: controlplane.DecisionPending,
		Metadata: map[string]string{"policy_id": "POLICY-017", "policy_version": "17"},
	})
	rec = doJSON(t, as, http.MethodPost, "/api/control-plane/approval",
		`{"decision_id":"DEC-API-REJECT","approver":"reviewer-1","expires_in_seconds":60}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected rejection approval creation 200, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &approval); err != nil {
		t.Fatal(err)
	}
	rec = doJSON(t, as, http.MethodPost, "/api/control-plane/approval/reject",
		`{"approval_id":"`+approval.ID+`","reason":"risk review failed"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected rejection 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Status != "rejected" {
		t.Fatalf("expected rejected status, got %s", status.Status)
	}
}

func TestAPIControlPlaneEndToEndEvidenceFlow(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/control-plane/decision",
		`{"agent_id":"agent-payment-01","capability":"repository.read"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("decision returned %d", rec.Code)
	}

	var decision struct {
		Status     string `json:"status"`
		RequestID  string `json:"request_id"`
		DecisionID string `json:"decision_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.Status != "ALLOW" || decision.RequestID == "" || decision.DecisionID == "" {
		t.Fatalf("unexpected decision response: %+v", decision)
	}

	rec = doJSON(t, as, http.MethodGet, "/api/control-plane/evidence", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("evidence returned %d", rec.Code)
	}
	var evidence struct {
		Events []struct {
			RequestID  string `json:"request_id"`
			DecisionID string `json:"decision_id"`
		} `json:"events"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.Events) == 0 {
		t.Fatal("expected evidence event")
	}
	last := evidence.Events[len(evidence.Events)-1]
	if last.RequestID != decision.RequestID || last.DecisionID != decision.DecisionID {
		t.Fatalf("decision/evidence correlation mismatch: decision=%+v evidence=%+v", decision, last)
	}

	rec = doJSON(t, as, http.MethodPost, "/api/control-plane/session", `{"agent_id":"agent-payment-01"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("session returned %d", rec.Code)
	}
	var session struct {
		Result          string `json:"result"`
		PolicyCompliant bool   `json:"policy_compliant"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Result != "PASS" || !session.PolicyCompliant {
		t.Fatalf("unexpected session result: %+v", session)
	}
}

func TestAPIControlPlaneApprovalExecutionFlow(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/control-plane/decision",
		`{"agent_id":"agent-payment-01","capability":"approval.request","artifact_hash":"sha256:approval-demo"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("decision returned %d: %s", rec.Code, rec.Body.String())
	}
	var decision struct {
		Status string `json:"status"`
		ID     string `json:"decision_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.Status != "REQUIRE_APPROVAL" {
		t.Fatalf("expected approval-required decision, got %+v", decision)
	}
	rec = doJSON(t, as, http.MethodPost, "/api/control-plane/approval",
		`{"decision_id":"`+decision.ID+`","approver":"reviewer-1","expires_in_seconds":60}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("approval creation returned %d: %s", rec.Code, rec.Body.String())
	}
	var approval struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &approval); err != nil {
		t.Fatal(err)
	}
	rec = doJSON(t, as, http.MethodPost, "/api/control-plane/approval/approve",
		`{"approval_id":"`+approval.ID+`","reason":"approved","artifact_hash":"sha256:approval-demo"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("approval returned %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, as, http.MethodPost, "/api/execute",
		`{"agent_id":"agent-payment-01","capability":"approval.request","decision_id":"`+decision.ID+`","approval_id":"`+approval.ID+`","artifact_hash":"sha256:approval-demo","payload":{"change":"demo"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("execution returned %d: %s", rec.Code, rec.Body.String())
	}
	var execution ExecutionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &execution); err != nil {
		t.Fatal(err)
	}
	if !execution.Authorized || !execution.Executed {
		t.Fatalf("expected approved execution, got %+v", execution)
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

func TestAPIScenario(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/scenario", `{"name":"credential_rotation"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var s ScenarioResult
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatal(err)
	}
	if !s.Passed {
		t.Errorf("expected credential_rotation scenario to pass")
	}
	if !strings.Contains(s.ActualResult, "BLOCK") {
		t.Errorf("expected BLOCK result, got %s", s.ActualResult)
	}
}

func TestAPIScenarioIAM(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/scenario", `{"name":"iam_modify"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var s ScenarioResult
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatal(err)
	}
	if !s.Passed {
		t.Errorf("expected iam_modify scenario to pass")
	}
}

func TestAPIScenarioNotFound(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodPost, "/api/scenario", `{"name":"nonexistent"}`)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestAPIScenarioMethodNotAllowed(t *testing.T) {
	as, _ := newTestAPI(t)
	rec := doJSON(t, as, http.MethodGet, "/api/scenario", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
