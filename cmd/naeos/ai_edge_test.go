package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/ai"
)

func newSSEMockServer(t *testing.T, openAIChunks ...string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		f := w.(http.Flusher)
		for _, c := range openAIChunks {
			fmt.Fprintf(w, "event: data\ndata: %s\n\n", c)
			f.Flush()
		}
		fmt.Fprint(w, "event: data\ndata: [DONE]\n\n")
		f.Flush()
	}))
}

func TestEnrichStreamChunks(t *testing.T) {
	server := newSSEMockServer(t,
		`{"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}`,
		`{"choices":[{"delta":{"content":" world"},"finish_reason":null}]}`,
	)
	defer server.Close()

	svc := ai.NewLLMService(ai.LLMConfig{
		Provider: ai.ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf bytes.Buffer
	if err := enrichStream(context.Background(), svc, "project: test", &buf); err != nil {
		t.Fatalf("enrichStream failed: %v", err)
	}
	if buf.String() != "Hello world\n" {
		t.Errorf("expected 'Hello world\\n', got %q", buf.String())
	}
}

func TestEnrichStreamErrorEvent(t *testing.T) {
	server := newSSEMockServer(t,
		`{"error":{"message":"rate limited"}}`,
	)
	defer server.Close()

	svc := ai.NewLLMService(ai.LLMConfig{
		Provider: ai.ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf bytes.Buffer
	err := enrichStream(context.Background(), svc, "project: test", &buf)
	if err == nil {
		t.Fatal("expected error from error event")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("expected rate limited message, got %v", err)
	}
}

func TestEnrichStreamErrorEventRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "plain error text")
	}))
	defer server.Close()

	svc := ai.NewLLMService(ai.LLMConfig{
		Provider: ai.ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf bytes.Buffer
	err := enrichStream(context.Background(), svc, "project: test", &buf)
	if err == nil {
		t.Fatal("expected error from raw error event")
	}
	if !strings.Contains(err.Error(), "plain error text") {
		t.Errorf("expected raw error text, got %v", err)
	}
}

func TestEnrichStreamEOF(t *testing.T) {
	server := newSSEMockServer(t,
		`{"choices":[{"delta":{"content":"partial"},"finish_reason":null}]}`,
	)
	defer server.Close()

	svc := ai.NewLLMService(ai.LLMConfig{
		Provider: ai.ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf bytes.Buffer
	if err := enrichStream(context.Background(), svc, "project: test", &buf); err != nil {
		t.Fatalf("enrichStream EOF failed: %v", err)
	}
	if buf.String() != "partial\n" {
		t.Errorf("expected 'partial\\n', got %q", buf.String())
	}
}

func TestEnrichStreamInvalidChunk(t *testing.T) {
	server := newSSEMockServer(t,
		`{not-json}`,
		`{"choices":[{"delta":{"content":"ok"},"finish_reason":null}]}`,
	)
	defer server.Close()

	svc := ai.NewLLMService(ai.LLMConfig{
		Provider: ai.ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	var buf bytes.Buffer
	if err := enrichStream(context.Background(), svc, "project: test", &buf); err != nil {
		t.Fatalf("enrichStream failed: %v", err)
	}
	if buf.String() != "ok\n" {
		t.Errorf("expected 'ok\\n', got %q", buf.String())
	}
}
