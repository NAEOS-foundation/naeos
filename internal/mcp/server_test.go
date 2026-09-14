// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/health", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/mcp", nil)
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected JSON-RPC error")
	}
	if resp.Error.Code != -32600 {
		t.Errorf("expected error code -32600, got %d", resp.Error.Code)
	}
}

func TestHandleMCPInvalidJSON(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected JSON-RPC error")
	}
	if resp.Error.Code != -32700 {
		t.Errorf("expected error code -32700, got %d", resp.Error.Code)
	}
}

func TestInitialize(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:      1,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	if len(tools) != 9 {
		t.Errorf("expected 9 tools, got %d", len(tools))
	}
}

func TestUnknownMethod(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "unknown/method",
		ID:      3,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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

func TestCallToolListArtifactsNoStore(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "list_artifacts",
			"arguments": map[string]any{},
		}),
		ID: 12,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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

func TestCallToolGetPipelineStatusMissingJob(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "get_pipeline_status",
			"arguments": map[string]any{"job_id": "nonexistent"},
		}),
		ID: 13,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected JSONRPC error: %v", resp.Error.Message)
	}
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}
	if isError, ok := resultMap["isError"].(bool); !ok || !isError {
		t.Error("expected isError=true for missing job")
	}
}

func TestCallToolGetPipelineStatusFound(t *testing.T) {
	s := newTestServer()
	now := time.Now()
	s.TrackPipelineJob(&PipelineJob{
		ID:        "job-1",
		Status:    "completed",
		StartedAt: now,
		Artifacts: 5,
	})
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "get_pipeline_status",
			"arguments": map[string]any{"job_id": "job-1"},
		}),
		ID: 14,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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

func TestCallToolGetPipelineStatusNoJobID(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "get_pipeline_status",
			"arguments": map[string]any{},
		}),
		ID: 15,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for missing job_id")
	}
}

func TestCallToolExportTerraform(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "export_terraform",
			"arguments": map[string]any{"spec": "project: test\nservices:\n  - name: api\n    kind: http\n"},
		}),
		ID: 16,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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

func TestCallToolExportTerraformNoSpec(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "export_terraform",
			"arguments": map[string]any{},
		}),
		ID: 17,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handleMCP(w, req)

	var resp JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for missing spec")
	}
}

func TestCallToolListPluginsNoManager(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mustMarshal(map[string]any{
			"name":      "list_plugins",
			"arguments": map[string]any{},
		}),
		ID: 18,
	})
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", bytes.NewReader(body))
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
	expected := []string{"parse_spec", "validate_spec", "generate_context", "compile_spec", "explain_concept", "list_artifacts", "get_pipeline_status", "export_terraform", "list_plugins"}
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

func FuzzHandleMCP(f *testing.F) {
	f.Add(`{"jsonrpc":"2.0","method":"initialize","id":1}`)
	f.Add(`{"jsonrpc":"2.0","method":"tools/list","id":2}`)
	f.Add(`{"jsonrpc":"2.0","method":"tools/call","params":{"name":"parse_spec","arguments":{"spec":"project: test"}},"id":3}`)
	f.Add(`not json`)
	f.Add(`{}`)
	f.Add(`{"jsonrpc":"2.0","method":"unknown","id":4}`)
	f.Add(`{"jsonrpc":"2.0","method":"tools/call","params":{"name":"parse_spec","arguments":{}},"id":5}`)

	f.Fuzz(func(t *testing.T, input string) {
		s := newTestServer()
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/mcp", strings.NewReader(input))
		w := httptest.NewRecorder()

		s.handleMCP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp JSONRPCResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.JSONRPC != "2.0" {
			t.Errorf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
		}
	})
}
