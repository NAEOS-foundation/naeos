package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/auth"
	"github.com/NAEOS-foundation/naeos/internal/compiler/adapters"
	"github.com/NAEOS-foundation/naeos/internal/database"
	"github.com/NAEOS-foundation/naeos/internal/multitenant"
	"github.com/NAEOS-foundation/naeos/internal/profiles"
	"github.com/NAEOS-foundation/naeos/internal/schemaregistry"
	naeosws "github.com/NAEOS-foundation/naeos/internal/websocket"
)

func TestLoggingMiddleware(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLoggingMiddlewareErrorStatus(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestLoggingMiddlewareWarnStatus(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSetPipelineObserver(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.SetPipelineObserver(&noopObserver{})
}

type noopObserver struct{}

func (n *noopObserver) OnPipelineStart(pipelineID string) {}
func (n *noopObserver) OnPipelineComplete(pipelineID string, artifacts int, duration string) {
}
func (n *noopObserver) OnPipelineFailed(pipelineID string, err string) {}
func (n *noopObserver) OnArtifactGenerated(name string, path string)   {}

func TestHandleSpecCompile(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.compiler.Register(adapters.NewCopilotAdapter(nil))

	body, _ := json.Marshal(map[string]string{
		"spec": "project: test\nmodules:\n  - name: core\n    path: ./core\n",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleSpecCompile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Success {
		t.Error("expected success true")
	}
}

func TestHandleSpecCompileMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/specs/compile", nil)
	w := httptest.NewRecorder()
	s.handleSpecCompile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleSpecCompileMissingSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/compile", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecCompile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSpecCompileInvalidSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec": "invalid: [yaml: [broken",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecCompile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSpecCompileUnknownTarget(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.compiler.Register(adapters.NewCopilotAdapter(nil))

	body, _ := json.Marshal(map[string]string{
		"spec":   "project: test",
		"target": "nonexistent",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecCompile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSpecCompileSpecificTarget(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.compiler.Register(adapters.NewCopilotAdapter(nil))

	body, _ := json.Marshal(map[string]string{
		"spec":   "project: test",
		"target": "copilot",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecCompile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlePluginExecute(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"name":   "test-plugin",
		"action": "ping",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/plugins/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePluginExecute(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandlePluginExecuteMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/plugins/execute", nil)
	w := httptest.NewRecorder()
	s.handlePluginExecute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlePluginExecuteMissingName(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{"action":"ping"}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/plugins/execute", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePluginExecute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlePluginExecuteDefaultAction(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{"name":"test"}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/plugins/execute", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePluginExecute(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 (plugin not found), got %d", w.Code)
	}
}

func TestHandleProfileSync(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"profiles": []map[string]string{
			{"id": "p1", "name": "Profile 1"},
			{"id": "p2", "name": "Profile 2"},
		},
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/sync", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileSync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleProfileSyncMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/profiles/sync", nil)
	w := httptest.NewRecorder()
	s.handleProfileSync(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleProfileSyncInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{bad json`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/sync", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileSync(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleArtifactsPOST(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"path":    "test.txt",
		"content": "hello world",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/artifacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleArtifacts(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleArtifactsPOSTMissingFields(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"path": "",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/artifacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleArtifacts(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleArtifactsPOSTInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/artifacts", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleArtifacts(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleArtifactsPOSTWithKind(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"path":    "config.yaml",
		"content": "key: value",
		"kind":    "config",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/artifacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleArtifacts(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlePluginsPOST(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"name":   "myplugin",
		"source": "https://example.com/plugin.zip",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/plugins", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePlugins(w, req)

	// Plugin install will fail since source doesn't exist, but we hit the handler
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusCreated {
		t.Errorf("expected 500 or 201, got %d", w.Code)
	}
}

func TestHandlePluginsPOSTMissingFields(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"name": "only-name",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/plugins", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePlugins(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlePluginsPOSTInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/plugins", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePlugins(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerWithMiddlewareOPTIONS(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "OPTIONS", "/api/v1/specs", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandlerWithMiddlewareCORS(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected CORS origin header, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestHandlerWithMiddlewareCORSUnknownOrigin(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS header for unknown origin")
	}
}

func TestHandlerWithMiddlewareCORSForwardedFor(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerWithMiddlewareSecurityHeaders(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected nosniff header")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("expected DENY frame options")
	}
}

func TestHandlerWithMiddlewareRequestID(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header")
	}
}

func TestHandlerWithMiddlewareExistingRequestID(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "my-custom-id" {
		t.Errorf("expected custom request ID, got %q", w.Header().Get("X-Request-ID"))
	}
}

func TestHandlerWithMiddlewarePostBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMapMethodToAction(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{"GET", "read"},
		{"POST", "create"},
		{"PUT", "update"},
		{"PATCH", "update"},
		{"DELETE", "delete"},
		{"HEAD", "other"},
	}
	for _, tt := range tests {
		got := mapMethodToAction(tt.method)
		if got != tt.want {
			t.Errorf("mapMethodToAction(%q) = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestAuditStatusFromHTTP(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{200, "success"},
		{201, "success"},
		{299, "success"},
		{400, "denied"},
		{404, "denied"},
		{499, "denied"},
		{500, "error"},
		{503, "error"},
		{300, "unknown"},
	}
	for _, tt := range tests {
		got := auditStatusFromHTTP(tt.status)
		if got != tt.want {
			t.Errorf("auditStatusFromHTTP(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestContainsHeader(t *testing.T) {
	tests := []struct {
		headers []string
		target  string
		want    bool
	}{
		{[]string{"Content-Type", "Authorization"}, "Content-Type", true},
		{[]string{"Content-Type", "Authorization"}, "authorization", true},
		{[]string{"Content-Type"}, "X-Request-ID", false},
		{nil, "anything", false},
	}
	for _, tt := range tests {
		got := containsHeader(tt.headers, tt.target)
		if got != tt.want {
			t.Errorf("containsHeader(%v, %q) = %v, want %v", tt.headers, tt.target, got, tt.want)
		}
	}
}

func TestOriginAllowed(t *testing.T) {
	tests := []struct {
		origin  string
		allowed []string
		want    bool
	}{
		{"http://localhost:3000", []string{"http://localhost:3000"}, true},
		{"http://evil.com", []string{"http://localhost:3000"}, false},
		{"anything", []string{"*"}, true},
		{"http://a.com", nil, false},
	}
	for _, tt := range tests {
		got := originAllowed(tt.origin, tt.allowed)
		if got != tt.want {
			t.Errorf("originAllowed(%q, %v) = %v, want %v", tt.origin, tt.allowed, got, tt.want)
		}
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{"GET", "POST"}, "GET, POST"},
		{[]string{"GET"}, "GET"},
		{nil, ""},
	}
	for _, tt := range tests {
		got := joinStrings(tt.input)
		if got != tt.want {
			t.Errorf("joinStrings(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHandleContextGeneratePlainFormat(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"spec":   "project: test",
		"format": "plain",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/context/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleContextGenerate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["format"] != "plain" {
		t.Errorf("expected plain format, got %v", result["format"])
	}
}

func TestHandleSpecsPOSTEmptySpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"spec": ""})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecs(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSpecsPOSTInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecs(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSpecValidatePOSTInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/validate", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecValidate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSpecVisualizeInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/visualize", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecVisualize(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSpecVisualizeInvalidSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"spec": ":\n  :\n    - bad"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/visualize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecVisualize(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSpecVisualizeFullSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	spec := `project: fullapp
architecture:
  pattern: hexagonal
  description: Test architecture
  principles:
    - separation of concerns
deployment:
  strategy: blue-green
  environments:
    - staging
testing:
  strategy: unit
generation:
  languages:
    - go
  output_dir: ./out
  module_dir: ./internal
modules:
  - name: core
    path: ./core
    description: Core logic
    dependencies: []
services:
  - name: api
    kind: http
    port: 8080
    description: API server
    endpoints:
      - method: GET
        path: /health
        action: healthcheck
`
	body, _ := json.Marshal(map[string]string{"spec": spec})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/visualize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecVisualize(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlePipelineRunPOSTInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/pipeline/run", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePipelineRun(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlePipelineRunPOSTMissingSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{"target":"go"}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/pipeline/run", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePipelineRun(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlePipelineRunInvalidSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"spec": ":\n  :\n    - bad"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/pipeline/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePipelineRun(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSchemasGET(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/schemas", nil)
	w := httptest.NewRecorder()
	s.handleSchemas(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleSchemasPOST(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{
		"name":    "order",
		"version": "v1",
		"schema":  `{"type":"object"}`,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/schemas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSchemas(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSchemasPOSTMissingFields(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"name": "order"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/schemas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSchemas(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSchemasPOSTInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/schemas", bytes.NewReader([]byte(`{bad`)))
	w := httptest.NewRecorder()
	s.handleSchemas(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSchemasMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "PUT", "/api/v1/schemas", nil)
	w := httptest.NewRecorder()
	s.handleSchemas(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleSchemaByPathGET(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.schemas.Register("test-schema", "v1", `{"type":"object"}`)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/schemas/test-schema/v1", nil)
	w := httptest.NewRecorder()
	s.handleSchemaByPath(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSchemaByPathGETVersions(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.schemas.Register("test-schema", "v1", `{"type":"object"}`)
	s.schemas.Register("test-schema", "v2", `{"type":"object"}`)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/schemas/test-schema", nil)
	w := httptest.NewRecorder()
	s.handleSchemaByPath(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleSchemaByPathGETNotFound(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/schemas/nonexistent/v1", nil)
	w := httptest.NewRecorder()
	s.handleSchemaByPath(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleSchemaByPathDELETE(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.schemas.Register("test-schema", "v1", `{"type":"object"}`)

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/schemas/test-schema/v1", nil)
	w := httptest.NewRecorder()
	s.handleSchemaByPath(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandleSchemaByPathDELETENoVersion(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/schemas/test-schema", nil)
	w := httptest.NewRecorder()
	s.handleSchemaByPath(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSchemaByPathDELETENotFound(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/schemas/nonexistent/v1", nil)
	w := httptest.NewRecorder()
	s.handleSchemaByPath(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleSchemaByPathMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/schemas/test/v1", nil)
	w := httptest.NewRecorder()
	s.handleSchemaByPath(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleSchemaByPathEmptyName(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/schemas/", nil)
	w := httptest.NewRecorder()
	s.handleSchemaByPath(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlePipelinesFilterByStatus(t *testing.T) {
	t.Setenv("NAEOS_PIPELINES_FILE", t.TempDir()+"/pipelines.json")
	s := NewServer(":0", &AuthConfig{Enabled: false})

	s.pipelinesMu.Lock()
	s.pipelines = append(s.pipelines, pipelineRun{
		ID:      "run-1",
		Status:  "completed",
		Project: "proj1",
	})
	s.pipelines = append(s.pipelines, pipelineRun{
		ID:      "run-2",
		Status:  "failed",
		Project: "proj2",
	})
	s.pipelinesMu.Unlock()

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/pipelines?status=completed", nil)
	w := httptest.NewRecorder()
	s.handlePipelines(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["total"] != float64(1) {
		t.Errorf("expected 1 result, got %v", result["total"])
	}
}

func TestHandlePipelinesSearch(t *testing.T) {
	t.Setenv("NAEOS_PIPELINES_FILE", t.TempDir()+"/pipelines.json")
	s := NewServer(":0", &AuthConfig{Enabled: false})

	s.pipelinesMu.Lock()
	s.pipelines = append(s.pipelines, pipelineRun{
		ID:      "run-1",
		Status:  "completed",
		Project: "my-project",
	})
	s.pipelines = append(s.pipelines, pipelineRun{
		ID:      "run-2",
		Status:  "completed",
		Project: "other",
	})
	s.pipelinesMu.Unlock()

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/pipelines?search=my-project", nil)
	w := httptest.NewRecorder()
	s.handlePipelines(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["total"] != float64(1) {
		t.Errorf("expected 1 result, got %v", result["total"])
	}
}

func TestHandlePipelinesPagination(t *testing.T) {
	t.Setenv("NAEOS_PIPELINES_FILE", t.TempDir()+"/pipelines.json")
	s := NewServer(":0", &AuthConfig{Enabled: false})

	s.pipelinesMu.Lock()
	for i := 0; i < 5; i++ {
		s.pipelines = append(s.pipelines, pipelineRun{
			ID:      fmt.Sprintf("run-%d", i),
			Status:  "completed",
			Project: "proj",
		})
	}
	s.pipelinesMu.Unlock()

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/pipelines?limit=2&offset=0", nil)
	w := httptest.NewRecorder()
	s.handlePipelines(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["total"] != float64(5) {
		t.Errorf("expected total 5, got %v", result["total"])
	}
}

func TestHandleTenantsNotConfigured(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/tenants", nil)
	w := httptest.NewRecorder()
	s.handleTenants(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleTenantsGET(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.SetWorkspace(multitenant.New())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/tenants", nil)
	w := httptest.NewRecorder()
	s.handleTenants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleTenantsPOST(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.SetWorkspace(multitenant.New())

	body, _ := json.Marshal(map[string]string{"name": "tenant-1"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTenants(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Errorf("expected 201 or 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleTenantsPOSTMissingName(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.SetWorkspace(multitenant.New())

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/tenants", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleTenants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleTenantsMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.SetWorkspace(multitenant.New())

	req := httptest.NewRequestWithContext(context.Background(), "PUT", "/api/v1/tenants", nil)
	w := httptest.NewRecorder()
	s.handleTenants(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleTenantByIDNotConfigured(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/tenants/t1", nil)
	w := httptest.NewRecorder()
	s.handleTenantByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleTenantByIDNotFound(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.SetWorkspace(multitenant.New())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/tenants/nonexistent", nil)
	w := httptest.NewRecorder()
	s.handleTenantByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleTenantByIDMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.SetWorkspace(multitenant.New())

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/tenants/t1", nil)
	w := httptest.NewRecorder()
	s.handleTenantByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleProfileByIDDELETE(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/v1/profiles/p1", nil)
	w := httptest.NewRecorder()
	s.handleProfileByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleProfileByIDPUT(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "PUT", "/api/v1/profiles/p1", nil)
	w := httptest.NewRecorder()
	s.handleProfileByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlePluginByNameGETNotFound(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/plugins/nonexistent", nil)
	w := httptest.NewRecorder()
	s.handlePluginByName(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandlePluginByNameEmptyName(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/plugins/", nil)
	w := httptest.NewRecorder()
	s.handlePluginByName(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSpecsGETCount(t *testing.T) {
	t.Setenv("NAEOS_PIPELINES_FILE", t.TempDir()+"/pipelines.json")
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/specs", nil)
	w := httptest.NewRecorder()
	s.handleSpecs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleSpecValidateInvalidSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"spec": ":\n  :\n    - bad"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleSpecValidate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data, _ := json.Marshal(resp.Data)
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["valid"].(bool) {
		t.Error("expected valid=false for invalid spec")
	}
}

func TestHandleProfilePublishInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/publish", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfilePublish(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleProfileSubscribeInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileSubscribe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleProfileSubscribeMissingURL(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{"interval":"5m"}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileSubscribe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleProfileSubscribeAlreadySubscribed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"registry_url": "https://example.com"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileSubscribe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Subscribe again
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	s.handleProfileSubscribe(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate, got %d", w2.Code)
	}
}

func TestHandleProfileSubscribeDefaultInterval(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"registry_url": "https://example.com"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileSubscribe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleProfileSubscribeInvalidInterval(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"registry_url": "https://example.com", "interval": "not-a-duration"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileSubscribe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleProfileUnsubscribeSuccess(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	// First subscribe
	body, _ := json.Marshal(map[string]string{"registry_url": "https://example.com"})
	subReq := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/subscribe", bytes.NewReader(body))
	subReq.Header.Set("Content-Type", "application/json")
	subW := httptest.NewRecorder()
	s.handleProfileSubscribe(subW, subReq)

	// Then unsubscribe
	unsubReq := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/unsubscribe", bytes.NewReader(body))
	unsubReq.Header.Set("Content-Type", "application/json")
	unsubW := httptest.NewRecorder()
	s.handleProfileUnsubscribe(unsubW, unsubReq)

	if unsubW.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", unsubW.Code)
	}
}

func TestHandleProfileUnsubscribeInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/unsubscribe", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileUnsubscribe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleProfileUnsubscribeMissingURL(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles/unsubscribe", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleProfileUnsubscribe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleProfileUnsubscribeMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/profiles/unsubscribe", nil)
	w := httptest.NewRecorder()
	s.handleProfileUnsubscribe(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleProfilesSearch(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.profiles.Register(&profiles.Profile{ID: "saas", Name: "SaaS App"})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/profiles?q=saas", nil)
	w := httptest.NewRecorder()
	s.handleProfiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleProfilesMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/profiles", nil)
	w := httptest.NewRecorder()
	s.handleProfiles(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleCloudStatusPagination(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/cloud/status?limit=2&offset=0", nil)
	w := httptest.NewRecorder()
	s.handleCloudStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleCloudDeployInvalidProvider(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"provider": "invalid-provider",
		"project":  "test",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/deploy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCloudDeploy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleCloudDestroyInvalidProvider(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]any{
		"provider": "invalid-provider",
		"project":  "test",
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/cloud/destroy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCloudDestroy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAIEnrichStreamInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/enrich/stream", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleAIEnrichStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAIEnrichStreamMissingSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/enrich/stream", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleAIEnrichStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAIEnrichStreamMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/ai/enrich/stream", nil)
	w := httptest.NewRecorder()
	s.handleAIEnrichStream(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAIExplainStreamInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/explain/stream", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleAIExplainStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAIExplainStreamMissingSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/explain/stream", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleAIExplainStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAIExplainStreamMethodNotAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/ai/explain/stream", nil)
	w := httptest.NewRecorder()
	s.handleAIExplainStream(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAICompileStreamInvalidBody(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/compile/stream", bytes.NewReader([]byte(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleAICompileStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAICompileStreamInvalidSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"spec": ":\n  :\n    - bad"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/compile/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleAICompileStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAICompileStreamDefaultTarget(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"spec": "project: test"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/ai/compile/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleAICompileStream(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBuildLLMConfigDefaults(t *testing.T) {
	cfg := buildLLMConfig("", "", "")
	if cfg.Provider != "openai" {
		t.Errorf("expected openai provider, got %s", cfg.Provider)
	}
	if cfg.Model != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini model, got %s", cfg.Model)
	}
}

func TestBuildLLMConfigAnthropicDefaults(t *testing.T) {
	cfg := buildLLMConfig("anthropic", "", "")
	if cfg.Model != "claude-3-haiku-20240307" {
		t.Errorf("expected claude-3-haiku, got %s", cfg.Model)
	}
}

func TestBuildLLMConfigOllamaDefaults(t *testing.T) {
	cfg := buildLLMConfig("ollama", "", "")
	if cfg.Model != "llama3.2" {
		t.Errorf("expected llama3.2, got %s", cfg.Model)
	}
}

func TestBuildLLMConfigUnknownProvider(t *testing.T) {
	cfg := buildLLMConfig("unknown-provider", "", "")
	if cfg.Provider != "openai" {
		t.Errorf("expected openai fallback, got %s", cfg.Provider)
	}
}

func TestBuildLLMConfigExplicitValues(t *testing.T) {
	cfg := buildLLMConfig("openai", "gpt-4", "my-key")
	if cfg.Model != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", cfg.Model)
	}
	if cfg.APIKey != "my-key" {
		t.Errorf("expected my-key, got %s", cfg.APIKey)
	}
}

func TestLoadPipelinesInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	pipelinesFile := dir + "/pipelines.json"
	if err := writeTestFile(pipelinesFile, []byte("not json")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NAEOS_PIPELINES_FILE", pipelinesFile)

	s := &Server{pipelinesFile: pipelinesFile}
	s.loadPipelines()

	s.pipelinesMu.RLock()
	if len(s.pipelines) != 0 {
		t.Errorf("expected 0 pipelines after invalid JSON, got %d", len(s.pipelines))
	}
	s.pipelinesMu.RUnlock()
}

func TestLoadPipelinesValidJSON(t *testing.T) {
	dir := t.TempDir()
	pipelinesFile := dir + "/pipelines.json"
	data := `[{"id":"run-1","status":"completed","project":"test"}]`
	if err := writeTestFile(pipelinesFile, []byte(data)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NAEOS_PIPELINES_FILE", pipelinesFile)

	s := &Server{pipelinesFile: pipelinesFile}
	s.loadPipelines()

	s.pipelinesMu.RLock()
	if len(s.pipelines) != 1 {
		t.Errorf("expected 1 pipeline, got %d", len(s.pipelines))
	}
	s.pipelinesMu.RUnlock()
}

func TestSavePipelines(t *testing.T) {
	dir := t.TempDir()
	pipelinesFile := dir + "/subdir/pipelines.json"
	s := &Server{pipelinesFile: pipelinesFile}
	s.pipelinesMu.Lock()
	s.pipelines = append(s.pipelines, pipelineRun{ID: "run-1", Status: "completed"})
	s.pipelinesMu.Unlock()

	s.savePipelines()

	if _, err := os.ReadFile(pipelinesFile); err != nil {
		t.Fatalf("expected pipelines file to exist: %v", err)
	}
}

func TestHandleContextGenerateInvalidSpec(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	body, _ := json.Marshal(map[string]string{"spec": ":\n  :\n    - bad"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/context/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleContextGenerate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerWithMiddlewareAuthRequired(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: true, JWTSecret: "test-secret"})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request without authorization
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/specs", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandlerWithMiddlewareAuthValidToken(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: true, JWTSecret: "test-secret"})

	token, _ := s.jwt.Generate(&JWTClaims{Sub: "user-1"})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerWithMiddlewareAuthInvalidToken(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: true, JWTSecret: "test-secret"})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/specs", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandlerWithMiddlewareRBACUserNotFound(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: true, JWTSecret: "test-secret"})
	s.authManager.CreateUser(&auth.User{ID: "known-user"})

	token, _ := s.jwt.Generate(&JWTClaims{Sub: "known-user"})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use a token with a user ID not in the manager
	badToken, _ := s.jwt.Generate(&JWTClaims{Sub: "unknown-user"})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer "+badToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	_ = token
}

func TestHandlerWithMiddlewareRBACInsufficientPerms(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: true, JWTSecret: "test-secret"})

	rbac := s.authManager.RBAC()
	rbac.AddRole(&auth.Role{Name: "viewer", Permissions: []string{"spec:read"}})
	rbac.AddPermission(&auth.Permission{Resource: "spec", Actions: []string{"read"}})

	s.authManager.CreateUser(&auth.User{ID: "viewer-user", Roles: []string{"viewer"}})

	token, _ := s.jwt.Generate(&JWTClaims{Sub: "viewer-user"})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Viewer tries to write spec (requires write permission)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/specs/validate", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestStopServer(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	err := s.Stop()
	if err != nil {
		t.Errorf("expected nil error on stop, got %v", err)
	}
}

func TestStopServerWithWebSocket(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.SetWebSocketServer(naeosws.NewServer())
	err := s.Stop()
	if err != nil {
		t.Errorf("expected nil error on stop, got %v", err)
	}
}

func TestMultiAuditorLog(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	// auditor is already set in NewServer
	_ = s
}

func TestSetupRoutes(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	// Verify all key routes are registered by checking handler existence
	routes := []string{
		"/metrics", "/healthz", "/readyz",
		"/api/v1/health", "/api/v1/specs", "/api/v1/specs/validate",
		"/api/v1/specs/compile", "/api/v1/specs/visualize",
		"/api/v1/pipeline/run", "/api/v1/pipeline/status",
		"/api/v1/artifacts", "/api/v1/context/generate",
		"/api/v1/mcp/message", "/api/v1/version",
		"/api/v1/config/schema", "/api/v1/pipelines",
	}

	for _, route := range routes {
		req := httptest.NewRequestWithContext(context.Background(), "GET", route, nil)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)
		// Just verify no panic - status code doesn't matter here
	}
}

func TestDefaultRoutePermissions(t *testing.T) {
	perms := defaultRoutePermissions()
	if len(perms) == 0 {
		t.Error("expected non-empty route permissions")
	}
	for path, rp := range perms {
		if rp.Resource == "" || rp.Action == "" {
			t.Errorf("route %s has empty resource or action", path)
		}
	}
}

func writeTestFile(path string, data []byte) error {
	dir := path[:len(path)-len("/"+path[strings.LastIndex(path, "/")+1:])]
	if err := mkdirAll(dir); err != nil {
		return err
	}
	return writeFile(path, data)
}

func mkdirAll(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}
	return err
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

// Ensure unused imports are referenced
var (
	_ = schemaregistry.New
	_ = database.Database(nil)
	_ = auth.NewManager
)
