package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

	result, err := svc.callOpenAI("test prompt")
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

	result, err := svc.callAnthropic("explain this")
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
	_, err := svc.callLLM("test")
	if err == nil {
		t.Error("expected error for unsupported provider")
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
		result := cleanJSON(tt.input)
		if result != tt.expected {
			t.Errorf("cleanJSON(%q) = %q, want %q", tt.input, result, tt.expected)
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
