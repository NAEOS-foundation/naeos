package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestLLMServiceOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("expected Bearer token")
		}

		resp := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "enriched spec content"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	result, err := svc.callOpenAI(t.Context(), "test prompt")
	if err != nil {
		t.Fatal(err)
	}
	if result != "enriched spec content" {
		t.Errorf("unexpected response: %s", result)
	}
}

func TestLLMServiceAnthropic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("expected x-api-key header")
		}

		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{
				{Text: "architectural explanation"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderAnthropic,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	result, err := svc.callAnthropic(t.Context(), "explain this")
	if err != nil {
		t.Fatal(err)
	}
	if result != "architectural explanation" {
		t.Errorf("unexpected response: %s", result)
	}
}

func TestLLMServiceUnsupportedProvider(t *testing.T) {
	svc := NewLLMService(LLMConfig{
		Provider: "unsupported",
		APIKey:   "key",
	})
	_, err := svc.callLLM(t.Context(), "test")
	if err == nil {
		t.Error("expected error for unsupported provider")
	}
}

func TestSSEDecoder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []struct{ event, data string }
		wantErr bool
	}{
		{
			name:  "single event",
			input: "event: chunk\ndata: {\"text\":\"hello\"}\n\n",
			want:  []struct{ event, data string }{{event: "chunk", data: `{"text":"hello"}`}},
		},
		{
			name:  "multi-line data",
			input: "event: chunk\ndata: {\"text\":\"line1\"}\ndata: {\"text\":\"line2\"}\n\n",
			want:  []struct{ event, data string }{{event: "chunk", data: "{\"text\":\"line1\"}\n{\"text\":\"line2\"}"}},
		},
		{
			name:  "multiple events",
			input: "event: start\ndata: {}\n\nevent: chunk\ndata: {\"text\":\"hello\"}\n\nevent: done\ndata: {\"status\":\"completed\"}\n\n",
			want: []struct{ event, data string }{
				{event: "start", data: `{}`},
				{event: "chunk", data: `{"text":"hello"}`},
				{event: "done", data: `{"status":"completed"}`},
			},
		},
		{
			name:  "comment line ignored",
			input: ":comment\nevent: chunk\ndata: x\n\n",
			want:  []struct{ event, data string }{{event: "chunk", data: "x"}},
		},
		{
			name:  "DONE marker",
			input: "event: data\ndata: [DONE]\n\n",
			want:  []struct{ event, data string }{{event: "data", data: "[DONE]"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewSSEDecoder(strings.NewReader(tt.input))
			var got []struct{ event, data string }
			for {
				event, data, err := decoder.Decode()
				if err != nil {
					break
				}
				got = append(got, struct{ event, data string }{event, string(data)})
			}

			if len(got) != len(tt.want) {
				t.Errorf("got %d events, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].event != tt.want[i].event || got[i].data != tt.want[i].data {
					t.Errorf("event %d: got (%q, %q), want (%q, %q)", i, got[i].event, got[i].data, tt.want[i].event, tt.want[i].data)
				}
			}
		})
	}
}

func TestStreamOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("expected Bearer token")
		}

		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if !req.Stream {
			t.Error("expected stream=true")
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flush")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "event: data\ndata: {\"choices\":[{\"delta\":{\"content\":\"Hello\"},\"finish_reason\":null}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: data\ndata: {\"choices\":[{\"delta\":{\"content\":\" World\"},\"finish_reason\":null}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: data\ndata: {\"choices\":[{\"delta\":{\"content\":\"\"},\"finish_reason\":\"stop\"}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: data\ndata: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf strings.Builder
	err := svc.streamOpenAI(t.Context(), "test prompt", &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, `"Hello"`) {
		t.Errorf("expected Hello in output, got: %s", output)
	}
	if !strings.Contains(output, `" World"`) {
		t.Errorf("expected ' World' in output, got: %s", output)
	}
	if !strings.Contains(output, `start`) {
		t.Errorf("expected start event, got: %s", output)
	}
	if !strings.Contains(output, `done`) {
		t.Errorf("expected done event, got: %s", output)
	}
}

func TestStreamOpenAIHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, `{"error":"overloaded"}`)
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf strings.Builder
	err := svc.streamOpenAI(t.Context(), "test", &buf)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStreamOpenAIError(t *testing.T) {
	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  "http://localhost:19999",
	})
	svc.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"text/event-stream"}},
			Body: io.NopCloser(strings.NewReader(
				"event: data\ndata: {\"error\":{\"message\":\"rate limit exceeded\"}}\n\n",
			)),
		}, nil
	})

	var buf strings.Builder
	err := svc.streamOpenAI(t.Context(), "test", &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStreamAnthropic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("expected x-api-key header")
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flush")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "event: content_block_delta\ndata: {\"delta\":{\"text\":\"Hello\"}}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: content_block_delta\ndata: {\"delta\":{\"text\":\" World\"}}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: message_stop\ndata: {}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderAnthropic,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf strings.Builder
	err := svc.streamAnthropic(t.Context(), "test prompt", &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, `"Hello"`) {
		t.Errorf("expected Hello in output, got: %s", output)
	}
	if !strings.Contains(output, `" World"`) {
		t.Errorf("expected ' World' in output, got: %s", output)
	}
	if !strings.Contains(output, `start`) {
		t.Errorf("expected start event, got: %s", output)
	}
	if !strings.Contains(output, `done`) {
		t.Errorf("expected done event, got: %s", output)
	}
}

func TestStreamEnrichSpec(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flush")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "event: data\ndata: {\"choices\":[{\"delta\":{\"content\":\"enriched\"},\"finish_reason\":null}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: data\ndata: {\"choices\":[{\"delta\":{\"content\":\" spec\"},\"finish_reason\":\"stop\"}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: data\ndata: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf strings.Builder
	err := svc.StreamEnrichSpec(t.Context(), "test spec", &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "enriched") || !strings.Contains(output, "spec") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestStreamExplainArchitecture(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flush")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "event: data\ndata: {\"choices\":[{\"delta\":{\"content\":\"arch\"},\"finish_reason\":null}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: data\ndata: {\"choices\":[{\"delta\":{\"content\":\" explanation\"},\"finish_reason\":\"stop\"}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: data\ndata: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf strings.Builder
	err := svc.StreamExplainArchitecture(t.Context(), "spec", "microservices", &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "arch") || !strings.Contains(output, "explanation") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestStreamCompileSpec(t *testing.T) {
	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  "http://localhost:19999",
	})
	svc.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"text/event-stream"}},
			Body: io.NopCloser(strings.NewReader(
				"event: data\ndata: {\"choices\":[{\"delta\":{\"content\":\"output\"},\"finish_reason\":\"stop\"}]}\n\nevent: data\ndata: [DONE]\n\n",
			)),
		}, nil
	})

	var buf strings.Builder
	err := svc.StreamCompileSpec(t.Context(), "opencode", "some spec content", &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "output") {
		t.Errorf("expected output in response, got: %s", output)
	}
}

func TestStreamLLMUnsupportedProvider(t *testing.T) {
	svc := NewLLMService(LLMConfig{
		Provider: "unsupported",
		APIKey:   "key",
	})

	var buf strings.Builder
	err := svc.streamLLM(t.Context(), "test", &buf)
	if err == nil {
		t.Error("expected error for unsupported provider")
	}
}

func TestStreamOpenAIContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		<-r.Context().Done()
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	var buf strings.Builder
	err := svc.streamOpenAI(ctx, "test", &buf)
	if err == nil {
		t.Error("expected error after context cancellation")
	}
}

func TestCleanJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"```json\n{\"a\":1}\n```", "{\"a\":1}"},
		{"```\n{\"a\":1}\n```", "{\"a\":1}"},
		{"{\"a\":1}", "{\"a\":1}"},
		{"  ```json\n[{\"b\":2}]\n```  ", "[{\"b\":2}]"},
	}

	for _, tt := range tests {
		result := CleanJSON(tt.input)
		if result != tt.expected {
			t.Errorf("CleanJSON(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestLLMServiceDefaultModel(t *testing.T) {
	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "key",
	})
	if svc.config.Model != "gpt-4o-mini" {
		t.Errorf("expected default model gpt-4o-mini, got %s", svc.config.Model)
	}

	svc2 := NewLLMService(LLMConfig{
		Provider: ProviderAnthropic,
		APIKey:   "key",
	})
	if svc2.config.Model != "claude-3-haiku-20240307" {
		t.Errorf("expected default model claude-3-haiku, got %s", svc2.config.Model)
	}

	svc3 := NewLLMService(LLMConfig{
		Provider: ProviderOllama,
	})
	if svc3.config.Model != "llama3.2" {
		t.Errorf("expected default model llama3.2, got %s", svc3.config.Model)
	}
	if svc3.config.BaseURL != "http://localhost:11434" {
		t.Errorf("expected default base URL http://localhost:11434, got %s", svc3.config.BaseURL)
	}
}

func TestLLMServiceOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("expected no Authorization header for Ollama")
		}

		resp := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "ollama response"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderOllama,
		Model:    "llama3.2",
		BaseURL:  server.URL,
	})

	result, err := svc.callOpenAI(t.Context(), "test prompt")
	if err != nil {
		t.Fatal(err)
	}
	if result != "ollama response" {
		t.Errorf("unexpected response: %s", result)
	}
}

func TestGenerateSuggestionsFromLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suggestions := []Suggestion{
			{Category: "security", Title: "Add auth", Description: "Add authentication", Priority: "high"},
			{Category: "performance", Title: "Add caching", Description: "Add Redis caching", Priority: "medium"},
		}
		resp := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: mustJSON(suggestions)}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	result, err := svc.GenerateSuggestions("project: myapp\nservices:\n  - name: api")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(result))
	}
	if !strings.Contains(result[0].Title, "auth") {
		t.Errorf("expected auth suggestion, got %s", result[0].Title)
	}
}

func mustJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
