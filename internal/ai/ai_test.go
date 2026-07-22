package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuggestEmpty(t *testing.T) {
	svc := NewService()
	_, err := svc.Suggest("")
	if err == nil {
		t.Fatal("expected error for empty spec")
	}
}

func TestSuggestMissingArchitecture(t *testing.T) {
	svc := NewService()
	suggestions, err := svc.Suggest("project: test\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, s := range suggestions {
		if s.Category == "architecture" {
			found = true
		}
	}
	if !found {
		t.Error("expected architecture suggestion")
	}
}

func TestSuggestCompleteSpec(t *testing.T) {
	svc := NewService()
	spec := "project: test\narchitecture:\n  pattern: hexagonal\ndeployment:\n  strategy: rolling\ntesting:\n  strategy: unit\ndescription: A test project\n"
	suggestions, err := svc.Suggest(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) == 0 {
		t.Error("expected at least one suggestion")
	}
}

func TestExplainPipeline(t *testing.T) {
	svc := NewService()
	exp, err := svc.Explain("pipeline", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp.Content == "" {
		t.Error("expected non-empty content")
	}
	if len(exp.Details) == 0 {
		t.Error("expected non-empty details")
	}
}

func TestExplainNEIR(t *testing.T) {
	svc := NewService()
	exp, err := svc.Explain("neir", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestExplainEmptyTopic(t *testing.T) {
	svc := NewService()
	_, err := svc.Explain("", "")
	if err == nil {
		t.Fatal("expected error for empty topic")
	}
}

func TestExplainUnknownTopic(t *testing.T) {
	svc := NewService()
	exp, err := svc.Explain("unknown-topic", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp.Content == "" {
		t.Error("expected non-empty content for unknown topic")
	}
}

func TestSuggestWithLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suggestions := []Suggestion{
			{Category: "security", Title: "Add auth middleware", Description: "Add JWT authentication", Priority: "high"},
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

	llm := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})
	svc := NewServiceWithLLM(llm)

	suggestions, err := svc.Suggest("project: myapp\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 LLM suggestion, got %d", len(suggestions))
	}
	if suggestions[0].Title != "Add auth middleware" {
		t.Errorf("expected LLM suggestion title, got %s", suggestions[0].Title)
	}
}

func TestSuggestWithLLMFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	llm := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})
	svc := NewServiceWithLLM(llm)

	suggestions, err := svc.Suggest("project: test\n")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range suggestions {
		if s.Category == "architecture" {
			found = true
		}
	}
	if !found {
		t.Error("expected fallback rule-based architecture suggestion")
	}
}

func TestExplainWithLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{
				{Text: "Hexagonal architecture separates core logic from external concerns via ports and adapters."},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	llm := NewLLMService(LLMConfig{
		Provider: ProviderAnthropic,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})
	svc := NewServiceWithLLM(llm)

	exp, err := svc.Explain("architecture", "project: myapp\narchitecture:\n  pattern: hexagonal")
	if err != nil {
		t.Fatal(err)
	}
	if exp.Content == "" {
		t.Error("expected non-empty LLM content")
	}
}

func TestExplainWithLLMFallback(t *testing.T) {
	svc := NewServiceWithLLM(nil)

	exp, err := svc.Explain("pipeline", "")
	if err != nil {
		t.Fatal(err)
	}
	if exp.Content == "" {
		t.Error("expected non-empty fallback content")
	}
	if len(exp.Details) == 0 {
		t.Error("expected non-empty fallback details")
	}
}

func TestNewServiceWithLLMNil(t *testing.T) {
	svc := NewServiceWithLLM(nil)
	if svc.llm != nil {
		t.Error("expected nil llm")
	}
	suggestions, err := svc.Suggest("project: test\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) == 0 {
		t.Error("expected rule-based suggestions")
	}
}
