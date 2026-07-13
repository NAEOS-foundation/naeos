package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	contextbundle "github.com/NAEOS-foundation/naeos/internal/context/bundle"
	"github.com/NAEOS-foundation/naeos/internal/version"
)

func newTestServer() *Server {
	c := compiler.New()
	bg := contextbundle.NewGenerator(c)
	return NewServer(c, bg)
}

func TestNewServer(t *testing.T) {
	s := newTestServer()
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	if s.mux == nil {
		t.Error("expected non-nil mux")
	}
}

func TestHealthEndpoint(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", resp["status"])
	}
}

func TestHandleMCPMethodNotAllowed(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleMCPInvalidJSON(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestInitialize(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:      1,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocol 2024-11-05, got %v", result["protocolVersion"])
	}
	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("expected serverInfo")
	}
	if serverInfo["version"] != version.String() {
		t.Errorf("expected version %s, got %v", version.String(), serverInfo["version"])
	}
}

func TestToolsList(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      2,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("expected tools array")
	}
	if len(tools) != 5 {
		t.Errorf("expected 5 tools, got %d", len(tools))
	}
}

func TestUnknownMethod(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "unknown/method",
		ID:      3,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected code -32601, got %d", resp.Error.Code)
	}
}

func TestCallToolParseSpec(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "parse_spec",
			"arguments": map[string]any{"spec": "project: test\nmodules:\n  - name: core\n    path: ./core\n"},
		}),
		ID: 4,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
}

func TestCallToolValidateSpec(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "validate_spec",
			"arguments": map[string]any{"spec": "project: test\n"},
		}),
		ID: 5,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
}

func TestCallToolGenerateContext(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "generate_context",
			"arguments": map[string]any{"spec": "project: test\n"},
		}),
		ID: 6,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
}

func TestCallToolCompileSpec(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "compile_spec",
			"arguments": map[string]any{"spec": "project: test\n", "target": "claude"},
		}),
		ID: 7,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
}

func TestCallToolExplainConcept(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "explain_concept",
			"arguments": map[string]any{"concept": "pipeline"},
		}),
		ID: 8,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
}

func TestCallToolUnknownTool(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "nonexistent",
			"arguments": map[string]any{},
		}),
		ID: 9,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestCallToolInvalidParams(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name": 123}`),
		ID:      10,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for invalid params")
	}
}

func TestExplainConceptUnknown(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "explain_concept",
			"arguments": map[string]any{"concept": "nonexistent"},
		}),
		ID: 11,
	})
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
}

func TestListTools(t *testing.T) {
	s := newTestServer()
	tools := s.listTools()
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	expected := []string{"parse_spec", "validate_spec", "generate_context", "compile_spec", "explain_concept"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected tool '%s' to be listed", name)
		}
	}
}

func TestHandler(t *testing.T) {
	s := newTestServer()
	h := s.Handler()
	if h == nil {
		t.Error("expected non-nil handler")
	}
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
