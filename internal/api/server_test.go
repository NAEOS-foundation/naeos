package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/auth"
	"github.com/NAEOS-foundation/naeos/internal/database"
	"github.com/NAEOS-foundation/naeos/internal/multitenant"
	"github.com/NAEOS-foundation/naeos/internal/profiles"
	naeosws "github.com/NAEOS-foundation/naeos/internal/websocket"
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/specs", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/validate", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/validate", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/pipeline/run", bytes.NewReader(body))
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

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/specs", nil)
	w := httptest.NewRecorder()

	s.handleSpecs(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestOIDCDiscovery(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: true, JWTSecret: "test-secret"})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/.well-known/openid-configuration", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()

	s.handleOIDCDiscovery(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestPipelineStatusEndpoint(t *testing.T) {
	t.Setenv("NAEOS_PIPELINES_FILE", t.TempDir()+"/pipelines.json")
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/pipeline/status", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/pipeline/status", nil)
	w := httptest.NewRecorder()

	s.handlePipelineStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestArtifactsEndpointDELETE(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/artifacts", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/context/generate", bytes.NewReader(body))
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/context/generate", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/context/generate", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/mcp/message", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/mcp/message", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/mcp/message", bytes.NewReader(body))
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/mcp/message", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/plan", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/plan", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/plan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudPlan(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCloudPlanMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/cloud/plan", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/deploy", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/deploy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudDeploy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCloudDeployMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/cloud/deploy", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/destroy", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/destroy", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/destroy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCloudDestroy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCloudDestroyMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/cloud/destroy", nil)
	w := httptest.NewRecorder()

	s.handleCloudDestroy(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestCloudStatusEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/cloud/status", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/status", nil)
	w := httptest.NewRecorder()

	s.handleCloudStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestPluginsEndpointGET(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/plugins", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/plugins", nil)
	w := httptest.NewRecorder()

	s.handlePlugins(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestPluginByNameDELETE(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/plugins/test-plugin", nil)
	w := httptest.NewRecorder()

	s.handlePluginByName(w, req)

	// Plugin doesn't exist; Uninstall returns error -> 404
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestPluginByNameMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "PUT", "/api/v1/plugins/test-plugin", nil)
	w := httptest.NewRecorder()

	s.handlePluginByName(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestVersionEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/version", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/config/schema", nil)
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
	t.Setenv("NAEOS_PIPELINES_FILE", t.TempDir()+"/pipelines.json")
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/pipelines", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/metrics", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/healthz", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/readyz", nil)
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

func TestProfilesEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/profiles", nil)
	w := httptest.NewRecorder()

	s.handleProfiles(w, req)

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

	if _, ok := result["profiles"]; !ok {
		t.Error("expected profiles in response")
	}
	if _, ok := result["count"]; !ok {
		t.Error("expected count in response")
	}
}

func TestProfilePublishEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"id":   "test-profile",
		"name": "Test Profile",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/publish", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleProfilePublish(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestProfileSubscribeEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"registry_url": "https://example.com/registry",
		"api_key":      "test-key",
		"interval":     "10s",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleProfileSubscribe(w, req)

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

func TestProfileUnsubscribeEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"registry_url": "https://example.com/registry",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/unsubscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleProfileUnsubscribe(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Error("expected success to be false")
	}
}

func TestProfileByIDEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	s.profiles.Register(&profiles.Profile{
		ID:   "saas",
		Name: "SaaS App",
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/profiles/saas", nil)
	w := httptest.NewRecorder()

	s.handleProfileByID(w, req)

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

	req404 := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/profiles/nonexistent", nil)
	w404 := httptest.NewRecorder()

	s.handleProfileByID(w404, req404)

	if w404.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w404.Code)
	}
}

func TestAIEnrichStreamEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec": "test spec",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/enrich/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleAIEnrichStream(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected content-type text/event-stream, got %s", ct)
	}

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "event:") {
		t.Error("expected SSE event framing in response")
	}
	if !strings.Contains(bodyStr, "data:") {
		t.Error("expected SSE data in response")
	}
}

func TestAIExplainStreamEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec":         "test",
		"architecture": "microservices",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/explain/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleAIExplainStream(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected content-type text/event-stream, got %s", ct)
	}

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "event:") {
		t.Error("expected SSE event framing in response")
	}
	if !strings.Contains(bodyStr, "data:") {
		t.Error("expected SSE data in response")
	}
}

func TestProfilePublishEndpointMissingFields(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/publish", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleProfilePublish(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestProfileSubscribeEndpointMissingFields(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleProfileSubscribe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSpecVisualizeEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec": "project: test\nmodules:\n  - name: core\n    path: ./core\n",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/visualize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleSpecVisualize(w, req)

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

func TestAICompileStreamEndpoint(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{"spec":"project: test","target":"opencode"}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/compile/stream", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleAICompileStream(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected text/event-stream content-type, got %s", ct)
	}
}

func TestAICompileStreamEndpointMissingSpec(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/compile/stream", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleAICompileStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAICompileStreamEndpointMethodNotAllowed(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/ai/compile/stream", nil)
	w := httptest.NewRecorder()

	s.handleAICompileStream(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantOffset int
		wantLimit  int
	}{
		{name: "defaults", query: "", wantOffset: 0, wantLimit: 50},
		{name: "limit only", query: "limit=10", wantOffset: 0, wantLimit: 10},
		{name: "limit capped", query: "limit=999", wantOffset: 0, wantLimit: 50},
		{name: "offset zero", query: "offset=0&limit=20", wantOffset: 0, wantLimit: 20},
		{name: "offset positive", query: "offset=30&limit=15", wantOffset: 30, wantLimit: 15},
		{name: "page one", query: "page=1&limit=10", wantOffset: 0, wantLimit: 10},
		{name: "page three", query: "page=3&limit=20", wantOffset: 40, wantLimit: 20},
		{name: "offset takes precedence", query: "offset=5&page=3&limit=10", wantOffset: 5, wantLimit: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), "GET", "/?"+tt.query, nil)
			offset, limit := parsePagination(req)
			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}
		})
	}
}

func TestServer_SetterMethods(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})

	ws := naeosws.NewServer()
	s.SetWebSocketServer(ws)
	if s.wsServer != ws {
		t.Error("SetWebSocketServer did not store reference")
	}

	s.SetDatabase(database.Database(nil))
	// Just ensures no panic

	s.SetAuthManager(auth.NewManager())
	// Just ensures no panic

	s.SetWorkspace(&multitenant.Workspace{})
	// Just ensures no panic

	s.SetRoutePermission("/custom", "spec", "read")
	if s.routePerms["/custom"].Resource != "spec" {
		t.Error("SetRoutePermission did not store permission")
	}
	if s.routePerms["/custom"].Action != "read" {
		t.Error("SetRoutePermission did not store action")
	}

	// SetPipelineObserver just ensures no panic - we can't easily create a PipelineObserver here
	// due to import cycle constraints, so skip it in this test

	s.RegisterAPIKey("new-key", 50)
	rl, ok := s.APIKeys["new-key"]
	if !ok {
		t.Fatal("RegisterAPIKey did not store key")
	}
	if !rl.Allow("any-client") {
		t.Error("expected registered API key rate limiter to allow first request")
	}
}

func TestServer_SetRoutePermissionNilMap(t *testing.T) {
	s := &Server{}
	s.SetRoutePermission("/path", "resource", "action")
	if s.routePerms == nil {
		t.Fatal("expected routePerms to be initialized")
	}
	if s.routePerms["/path"].Resource != "resource" {
		t.Errorf("expected resource 'resource', got %q", s.routePerms["/path"].Resource)
	}
}

func TestServer_handleHealthWithDB(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})
	s.SetDatabase(&mockDB{})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if db, ok := resp.Data["database"]; !ok || db != "healthy" {
		t.Errorf("expected database healthy, got %v", resp.Data["database"])
	}
}

func TestServer_handleHealthWithUnhealthyDB(t *testing.T) {
	s := NewServer(":8080", &AuthConfig{Enabled: false})
	s.SetDatabase(&unhealthyDB{})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	dbStatus, ok := resp.Data["database"].(string)
	if !ok || !strings.Contains(dbStatus, "unhealthy") {
		t.Errorf("expected database unhealthy, got %v", resp.Data["database"])
	}
}

type mockDB struct{}

func (m *mockDB) Name() string                          { return "mock" }
func (m *mockDB) Connect(config *database.Config) error { return nil }
func (m *mockDB) Close() error                          { return nil }
func (m *mockDB) Ping() error                           { return nil }
func (m *mockDB) Exec(query string, args ...any) (database.Result, error) {
	return database.Result{}, nil
}
func (m *mockDB) ExecContext(ctx context.Context, query string, args ...any) (database.Result, error) {
	return database.Result{}, nil
}
func (m *mockDB) Query(query string, args ...any) ([]database.Row, error) { return nil, nil }
func (m *mockDB) QueryContext(ctx context.Context, query string, args ...any) ([]database.Row, error) {
	return nil, nil
}
func (m *mockDB) QueryRow(query string, args ...any) (database.Row, error) { return nil, nil }
func (m *mockDB) QueryRowContext(ctx context.Context, query string, args ...any) (database.Row, error) {
	return nil, nil
}
func (m *mockDB) Begin() (database.Transaction, error)                      { return nil, nil }
func (m *mockDB) BeginTx(ctx context.Context) (database.Transaction, error) { return nil, nil }
func (m *mockDB) Migrate(migrations []database.Migration) error             { return nil }
func (m *mockDB) MigrateContext(ctx context.Context, migrations []database.Migration) error {
	return nil
}
func (m *mockDB) Rollback(version int) error                             { return nil }
func (m *mockDB) RollbackContext(ctx context.Context, version int) error { return nil }
func (m *mockDB) HealthCheck() error                                     { return nil }

type unhealthyDB struct{}

func (m *unhealthyDB) Name() string                          { return "unhealthy" }
func (m *unhealthyDB) Connect(config *database.Config) error { return nil }
func (m *unhealthyDB) Close() error                          { return nil }
func (m *unhealthyDB) Ping() error                           { return nil }
func (m *unhealthyDB) Exec(query string, args ...any) (database.Result, error) {
	return database.Result{}, nil
}
func (m *unhealthyDB) ExecContext(ctx context.Context, query string, args ...any) (database.Result, error) {
	return database.Result{}, nil
}
func (m *unhealthyDB) Query(query string, args ...any) ([]database.Row, error) { return nil, nil }
func (m *unhealthyDB) QueryContext(ctx context.Context, query string, args ...any) ([]database.Row, error) {
	return nil, nil
}
func (m *unhealthyDB) QueryRow(query string, args ...any) (database.Row, error) { return nil, nil }
func (m *unhealthyDB) QueryRowContext(ctx context.Context, query string, args ...any) (database.Row, error) {
	return nil, nil
}
func (m *unhealthyDB) Begin() (database.Transaction, error)                      { return nil, nil }
func (m *unhealthyDB) BeginTx(ctx context.Context) (database.Transaction, error) { return nil, nil }
func (m *unhealthyDB) Migrate(migrations []database.Migration) error             { return nil }
func (m *unhealthyDB) MigrateContext(ctx context.Context, migrations []database.Migration) error {
	return nil
}
func (m *unhealthyDB) Rollback(version int) error                             { return nil }
func (m *unhealthyDB) RollbackContext(ctx context.Context, version int) error { return nil }
func (m *unhealthyDB) HealthCheck() error                                     { return fmt.Errorf("connection refused") }
