package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})
	if s == nil {
		t.Fatal("expected server to be created")
	}
	if s.Addr != ":8080" {
		t.Errorf("expected addr ':8080', got %s", s.Addr)
	}
}

func TestHealthEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestSpecsEndpointGET(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/specs", nil)
	w := httptest.NewRecorder()

	s.handleSpecs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestSpecsEndpointPOST(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec": "project: test\nmodules:\n  - name: core\n    path: ./core\n",
	})
	req := httptest.NewRequest("POST", "/api/v1/specs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleSpecs(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestSpecValidateEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec": "project: test\nmodules:\n  - name: core\n    path: ./core\n",
	})
	req := httptest.NewRequest("POST", "/api/v1/specs/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleSpecValidate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestSpecValidateEndpointInvalid(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec": "",
	})
	req := httptest.NewRequest("POST", "/api/v1/specs/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleSpecValidate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["valid"].(bool) {
		t.Error("expected valid to be false for empty spec")
	}
}

func TestPipelineRunEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec": "project: test\nmodules:\n  - name: core\n    path: ./core\n",
	})
	req := httptest.NewRequest("POST", "/api/v1/pipeline/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handlePipelineRun(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("job_id")) {
		t.Errorf("expected job_id in response, got %s", w.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("DELETE", "/api/v1/specs", nil)
	w := httptest.NewRecorder()

	s.handleSpecs(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestOIDCDiscovery(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: true, JWTSecret: "test-secret"})

	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	req.Host = "localhost:8080"
	w := httptest.NewRecorder()

	s.handleOIDCDiscovery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var doc OIDCDiscovery
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("failed to decode OIDC discovery: %v", err)
	}

	if doc.Issuer != "http://localhost:8080" {
		t.Errorf("expected issuer 'http://localhost:8080', got %s", doc.Issuer)
	}

	if len(doc.IDTokenSigningAlgValuesSupported) != 1 || doc.IDTokenSigningAlgValuesSupported[0] != "HS256" {
		t.Errorf("expected HS256 signing alg, got %v", doc.IDTokenSigningAlgValuesSupported)
	}
}

func TestOIDCDiscoveryNotConfigured(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()

	s.handleOIDCDiscovery(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestPipelineStatusEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/pipeline/status", nil)
	w := httptest.NewRecorder()

	s.handlePipelineStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}

	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)

	if result["status"] != "idle" {
		t.Errorf("expected status 'idle', got %v", result["status"])
	}
	if result["total"] != float64(0) {
		t.Errorf("expected total 0, got %v", result["total"])
	}
}

func TestPipelineStatusMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("POST", "/api/v1/pipeline/status", nil)
	w := httptest.NewRecorder()

	s.handlePipelineStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestArtifactsEndpointDELETE(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("DELETE", "/api/v1/artifacts", nil)
	w := httptest.NewRecorder()

	s.handleArtifacts(w, req)

	// DELETE is not handled in the switch; falls through to default -> 405
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestContextGenerateEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec":   "project: test\nmodules:\n  - name: core\n    path: ./core\n",
		"format": "markdown",
	})
	req := httptest.NewRequest("POST", "/api/v1/context/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleContextGenerate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestContextGenerateMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/context/generate", nil)
	w := httptest.NewRecorder()

	s.handleContextGenerate(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestContextGenerateMissingSpec(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec": "",
	})
	req := httptest.NewRequest("POST", "/api/v1/context/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleContextGenerate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestMCPMessageEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialize",
		"id":      1,
	})
	req := httptest.NewRequest("POST", "/api/v1/mcp/message", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleMCPMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var rpcResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if rpcResp["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %v", rpcResp["jsonrpc"])
	}
	if rpcResp["result"] == nil {
		t.Error("expected result to be non-nil for initialize")
	}
}

func TestMCPToolsList(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "tools/list",
		"id":      2,
	})
	req := httptest.NewRequest("POST", "/api/v1/mcp/message", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleMCPMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var rpcResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if rpcResp["result"] == nil {
		t.Error("expected result to be non-nil for tools/list")
	}
}

func TestMCPMethodNotFound(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "nonexistent/method",
		"id":      3,
	})
	req := httptest.NewRequest("POST", "/api/v1/mcp/message", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleMCPMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 (error is in JSON-RPC body), got %d", w.Code)
	}

	var rpcResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if rpcResp["error"] == nil {
		t.Error("expected error in response for unknown method")
	}
}

func TestMCPMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/mcp/message", nil)
	w := httptest.NewRecorder()

	s.handleMCPMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var rpcResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if rpcResp["error"] == nil {
		t.Error("expected JSON-RPC error")
	}
}

func TestCloudPlanEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"provider": "aws",
		"project":  "test-project",
		"region":   "us-east-1",
		"resources": []map[string]any{
			{"name": "bucket1", "type": "storage"},
		},
	})
	req := httptest.NewRequest("POST", "/api/v1/cloud/plan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudPlan(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestCloudPlanMissingProvider(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"project": "test-project",
	})
	req := httptest.NewRequest("POST", "/api/v1/cloud/plan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudPlan(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCloudPlanInvalidProvider(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"provider": "invalid",
		"project":  "test-project",
	})
	req := httptest.NewRequest("POST", "/api/v1/cloud/plan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudPlan(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCloudPlanMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/cloud/plan", nil)
	w := httptest.NewRecorder()

	s.handleCloudPlan(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestCloudDeployEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"provider": "aws",
		"project":  "test-project",
		"region":   "us-east-1",
		"resources": []map[string]any{
			{"name": "bucket1", "type": "storage"},
		},
	})
	req := httptest.NewRequest("POST", "/api/v1/cloud/deploy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudDeploy(w, req)

	// Deploy requires terraform binary; expect either 200 or 500
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500, got %d", w.Code)
	}
}

func TestCloudDeployMissingProvider(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"project": "test-project",
	})
	req := httptest.NewRequest("POST", "/api/v1/cloud/deploy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudDeploy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCloudDeployMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/cloud/deploy", nil)
	w := httptest.NewRecorder()

	s.handleCloudDeploy(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestCloudDestroyEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"provider": "aws",
		"project":  "test-project",
		"region":   "us-east-1",
		"resources": []map[string]any{
			{"name": "bucket1", "type": "storage"},
		},
	})
	req := httptest.NewRequest("POST", "/api/v1/cloud/destroy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudDestroy(w, req)

	// Destroy requires terraform binary; expect either 200 or 500
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500, got %d", w.Code)
	}
}

func TestCloudDestroyMissingProvider(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"project": "test-project",
	})
	req := httptest.NewRequest("POST", "/api/v1/cloud/destroy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudDestroy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCloudDestroyInvalidProvider(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"provider": "bogus",
		"project":  "test-project",
	})
	req := httptest.NewRequest("POST", "/api/v1/cloud/destroy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudDestroy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCloudDestroyMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("DELETE", "/api/v1/cloud/destroy", nil)
	w := httptest.NewRecorder()

	s.handleCloudDestroy(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestCloudStatusEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/cloud/status", nil)
	w := httptest.NewRecorder()

	s.handleCloudStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestCloudStatusMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("POST", "/api/v1/cloud/status", nil)
	w := httptest.NewRecorder()

	s.handleCloudStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestPluginsEndpointGET(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/plugins", nil)
	w := httptest.NewRecorder()

	s.handlePlugins(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestPluginsEndpointMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("DELETE", "/api/v1/plugins", nil)
	w := httptest.NewRecorder()

	s.handlePlugins(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestPluginByNameDELETE(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("DELETE", "/api/v1/plugins/test-plugin", nil)
	w := httptest.NewRecorder()

	s.handlePluginByName(w, req)

	// Plugin doesn't exist; Uninstall returns error -> 404
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestPluginByNameMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("PUT", "/api/v1/plugins/test-plugin", nil)
	w := httptest.NewRecorder()

	s.handlePluginByName(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestVersionEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/version", nil)
	w := httptest.NewRecorder()

	s.handleVersion(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}

	data, _ := json.Marshal(resp.Data)
	var result map[string]string
	json.Unmarshal(data, &result)

	if result["platform"] != "naeos-api" {
		t.Errorf("expected platform 'naeos-api', got %s", result["platform"])
	}
}

func TestConfigSchemaEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/config/schema", nil)
	w := httptest.NewRecorder()

	s.handleConfigSchema(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}

	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)

	if result["type"] != "object" {
		t.Errorf("expected type 'object', got %v", result["type"])
	}
	if result["properties"] == nil {
		t.Error("expected properties to be non-nil")
	}
}

func TestPipelinesEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/api/v1/pipelines", nil)
	w := httptest.NewRecorder()

	s.handlePipelines(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}

	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)

	if result["total"] != float64(0) {
		t.Errorf("expected total 0, got %v", result["total"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	s.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain; version=0.0.4" {
		t.Errorf("expected content type 'text/plain; version=0.0.4', got %s", contentType)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty metrics body")
	}
}

func TestHealthzEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	s.handleHealthz(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content type 'application/json', got %s", contentType)
	}

	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("healthy")) {
		t.Errorf("expected 'healthy' in response body, got %s", body)
	}
}

func TestReadyzEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()

	s.handleReadyz(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content type 'application/json', got %s", contentType)
	}

	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("ready")) {
		t.Errorf("expected 'ready' in response body, got %s", body)
	}
}
