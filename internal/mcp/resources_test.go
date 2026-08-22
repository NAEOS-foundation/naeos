package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
)

func postMCP(t *testing.T, s *Server, method string, params any) map[string]any {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handleMCP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestInitializeAdvertisesResourcesAndPrompts(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	resp := postMCP(t, s, "initialize", map[string]any{})

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatal("expected result object")
	}
	caps, ok := result["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("expected capabilities object")
	}
	for _, capName := range []string{"tools", "resources", "prompts"} {
		if _, ok := caps[capName]; !ok {
			t.Errorf("expected capability %q to be advertised", capName)
		}
	}
}

func TestResourcesListIncludesDocs(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	resp := postMCP(t, s, "resources/list", nil)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatal("expected result object")
	}
	rawList, ok := result["resources"].([]any)
	if !ok {
		t.Fatal("expected resources array")
	}

	uris := make(map[string]bool, len(rawList))
	for _, raw := range rawList {
		r, ok := raw.(map[string]any)
		if !ok {
			t.Fatal("expected resource object")
		}
		uri, _ := r["uri"].(string)
		if uri == "" {
			t.Error("resource with empty uri")
		}
		uris[uri] = true
	}
	for _, concept := range conceptNames() {
		want := uriSchemeDocs + concept
		if !uris[want] {
			t.Errorf("expected docs resource %q in list", want)
		}
	}
}

func TestResourcesReadDoc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		uri      string
		wantText string
	}{
		{name: "pipeline", uri: "naeos://docs/pipeline", wantText: "NAEOS Pipeline"},
		{name: "neir", uri: "naeos://docs/neir", wantText: "NEIR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			resp := postMCP(t, s, "resources/read", map[string]any{"uri": tt.uri})

			result, ok := resp["result"].(map[string]any)
			if !ok {
				t.Fatal("expected result object")
			}
			contents, ok := result["contents"].([]any)
			if !ok || len(contents) == 0 {
				t.Fatal("expected non-empty contents array")
			}
			first, ok := contents[0].(map[string]any)
			if !ok {
				t.Fatal("expected content object")
			}
			if got := first["text"].(string); !strings.Contains(got, tt.wantText) {
				t.Errorf("text = %q, want substring %q", got, tt.wantText)
			}
			if first["uri"] != tt.uri {
				t.Errorf("uri = %v, want %q", first["uri"], tt.uri)
			}
		})
	}
}

func TestResourcesReadErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		params     map[string]any
		errContain string
	}{
		{
			name:       "missing uri",
			params:     map[string]any{},
			errContain: "invalid params",
		},
		{
			name:       "unknown concept",
			params:     map[string]any{"uri": "naeos://docs/nonexistent"},
			errContain: "not found",
		},
		{
			name:       "unsupported scheme",
			params:     map[string]any{"uri": "file:///etc/passwd"},
			errContain: "unsupported resource URI",
		},
		{
			name:       "unknown artifact",
			params:     map[string]any{"uri": "naeos://artifacts/does/not/exist.txt"},
			errContain: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			s.SetArtifactStore(artifacts.NewStore(t.TempDir()))
			resp := postMCP(t, s, "resources/read", tt.params)

			errObj, ok := resp["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected error object, got %v", resp)
			}
			msg, _ := errObj["message"].(string)
			if !strings.Contains(msg, tt.errContain) {
				t.Errorf("error message = %q, want substring %q", msg, tt.errContain)
			}
		})
	}
}

func TestResourcesArtifactsLifecycle(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	store := artifacts.NewStore(t.TempDir())
	_, err := store.Add("main.go", []byte("package main"), artifacts.KindCode)
	if err != nil {
		t.Fatalf("add artifact: %v", err)
	}
	s.SetArtifactStore(store)

	listResp := postMCP(t, s, "resources/list", nil)
	result := listResp["result"].(map[string]any)
	rawList := result["resources"].([]any)
	found := false
	for _, raw := range rawList {
		r := raw.(map[string]any)
		if r["uri"] == uriSchemeArtifacts+"main.go" {
			found = true
		}
	}
	if !found {
		t.Error("artifact resource missing from resources/list")
	}

	readResp := postMCP(t, s, "resources/read", map[string]any{"uri": uriSchemeArtifacts + "main.go"})
	contents := readResp["result"].(map[string]any)["contents"].([]any)
	first := contents[0].(map[string]any)
	if got := first["text"]; got != "package main" {
		t.Errorf("artifact text = %v, want %q", got, "package main")
	}
}

func TestResourcesJobsLifecycle(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	job := &PipelineJob{ID: "job-1", Status: "running"}
	s.TrackPipelineJob(job)

	readResp := postMCP(t, s, "resources/read", map[string]any{"uri": uriSchemeJobs + "job-1"})
	contents := readResp["result"].(map[string]any)["contents"].([]any)
	first := contents[0].(map[string]any)
	text, _ := first["text"].(string)
	if !strings.Contains(text, `"status": "running"`) {
		t.Errorf("job text = %q, want status running", text)
	}
}

func TestPromptsList(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	resp := postMCP(t, s, "prompts/list", nil)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatal("expected result object")
	}
	rawList, ok := result["prompts"].([]any)
	if !ok {
		t.Fatal("expected prompts array")
	}
	names := make(map[string]bool, len(rawList))
	for _, raw := range rawList {
		p := raw.(map[string]any)
		name, _ := p["name"].(string)
		names[name] = true
		args, ok := p["arguments"].([]any)
		if !ok || len(args) == 0 {
			t.Errorf("prompt %q has no arguments", name)
		}
	}
	for _, want := range []string{"review-spec", "enrich-spec", "explain-architecture", "generate-spec"} {
		if !names[want] {
			t.Errorf("expected prompt %q in list", want)
		}
	}
}

func TestPromptsGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		prompt       string
		args         map[string]any
		wantContains []string
	}{
		{
			name:   "review spec",
			prompt: "review-spec",
			args:   map[string]any{"spec": "project: demo"},
			wantContains: []string{
				"Review this NAEOS specification",
				"project: demo",
			},
		},
		{
			name:   "explain architecture",
			prompt: "explain-architecture",
			args:   map[string]any{"spec": "project: demo", "architecture": "microservices"},
			wantContains: []string{
				"microservices",
				"project: demo",
			},
		},
		{
			name:   "generate spec",
			prompt: "generate-spec",
			args:   map[string]any{"description": "a todo API"},
			wantContains: []string{
				"a todo API",
				"Draft a complete NAEOS specification",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			resp := postMCP(t, s, "prompts/get", map[string]any{"name": tt.prompt, "arguments": tt.args})

			result, ok := resp["result"].(map[string]any)
			if !ok {
				t.Fatalf("expected result object, got %v", resp)
			}
			messages, ok := result["messages"].([]any)
			if !ok || len(messages) != 1 {
				t.Fatal("expected exactly one message")
			}
			msg := messages[0].(map[string]any)
			if msg["role"] != "user" {
				t.Errorf("role = %v, want user", msg["role"])
			}
			content := msg["content"].(map[string]any)
			text, _ := content["text"].(string)
			for _, want := range tt.wantContains {
				if !strings.Contains(text, want) {
					t.Errorf("text does not contain %q:\n%s", want, text)
				}
			}
		})
	}
}

func TestPromptsGetErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		params     map[string]any
		errContain string
	}{
		{
			name:       "unknown prompt",
			params:     map[string]any{"name": "nonexistent"},
			errContain: "unknown prompt",
		},
		{
			name:       "missing required argument",
			params:     map[string]any{"name": "review-spec", "arguments": map[string]any{}},
			errContain: "'spec' argument is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			resp := postMCP(t, s, "prompts/get", tt.params)

			errObj, ok := resp["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected error object, got %v", resp)
			}
			msg, _ := errObj["message"].(string)
			if !strings.Contains(msg, tt.errContain) {
				t.Errorf("error message = %q, want substring %q", msg, tt.errContain)
			}
		})
	}
}

func TestRenderPromptTemplateVerbatimValues(t *testing.T) {
	t.Parallel()
	values := map[string]string{"spec": "spec with {{spec}} inside"}
	got := renderPromptTemplate("review-spec", values)
	if strings.Count(got, "{{spec}}") < 1 {
		t.Error("value should be inserted verbatim without recursive substitution")
	}
	if !strings.Contains(got, fmt.Sprintf("%s", values["spec"])) {
		t.Error("value text should appear in output")
	}
}
