package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildEnrichPrompt(t *testing.T) {
	s := &LLMService{}
	prompt := s.buildEnrichPrompt("spec: test")
	if !strings.Contains(prompt, "spec: test") {
		t.Error("expected prompt to contain spec content")
	}
}

func TestTruncatePromptShort(t *testing.T) {
	s := &LLMService{}
	input := "short prompt"
	got := s.truncatePrompt(input)
	if got != input {
		t.Errorf("expected unchanged short prompt")
	}
}

func TestTruncatePromptLong(t *testing.T) {
	s := &LLMService{
		config: LLMConfig{
			Model:     "unknown",
			MaxTokens: 1,
		},
	}
	input := strings.Repeat("x ", 50000)
	got := s.truncatePrompt(input)
	if len(got) >= len(input) {
		t.Error("expected truncated output")
	}
}

func TestSplitSSELineEdgeCases(t *testing.T) {
	parts := splitSSELine("")
	if parts != nil {
		t.Error("expected nil for empty line")
	}
	parts = splitSSELine("no colon")
	if parts != nil {
		t.Error("expected nil for line without colon")
	}
	parts = splitSSELine("key:val")
	if len(parts) != 2 || parts[0] != "key" || parts[1] != "val" {
		t.Errorf("expected [key val], got %v", parts)
	}
	parts = splitSSELine("data: hello world")
	if len(parts) != 2 || parts[0] != "data" || parts[1] != "hello world" {
		t.Errorf("expected [data hello world], got %v", parts)
	}
}

func TestBuildSuggestionsPrompt(t *testing.T) {
	s := &LLMService{}
	prompt := s.buildSuggestionsPrompt("spec content here")
	if !strings.Contains(prompt, "spec content") {
		t.Error("expected prompt to contain spec content")
	}
}

func TestBuildExplainPrompt(t *testing.T) {
	s := &LLMService{}
	prompt := s.buildExplainPrompt("concept", "arch")
	if !strings.Contains(prompt, "concept") || !strings.Contains(prompt, "arch") {
		t.Error("expected prompt to contain concept and arch")
	}
}

func TestBuildCompilerPrompt(t *testing.T) {
	s := &LLMService{}
	prompt := s.buildCompilerPrompt("target", "spec")
	if !strings.Contains(prompt, "target") || !strings.Contains(prompt, "spec") {
		t.Error("expected prompt to contain target and spec")
	}
}

func TestNewLLMServiceDefaults(t *testing.T) {
	s := NewLLMService(LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "",
	})
	if s.config.Model != "gpt-4o-mini" {
		t.Errorf("expected default model gpt-4o-mini, got %s", s.config.Model)
	}
	if s.config.MaxTokens != 1024 {
		t.Errorf("expected default MaxTokens 1024, got %d", s.config.MaxTokens)
	}
	if s.config.Timeout == 0 {
		t.Error("expected default timeout")
	}
}

func TestNewLLMServiceOllamaDefaults(t *testing.T) {
	s := NewLLMService(LLMConfig{
		Provider: ProviderOllama,
		APIKey:   "",
	})
	if s.config.Model != "llama3.2" {
		t.Errorf("expected llama3.2, got %s", s.config.Model)
	}
	if s.config.BaseURL != "http://localhost:11434" {
		t.Errorf("expected default ollama url, got %s", s.config.BaseURL)
	}
}

func TestNewLLMServiceAnthropicDefaults(t *testing.T) {
	s := NewLLMService(LLMConfig{
		Provider: ProviderAnthropic,
		APIKey:   "key",
	})
	if s.config.Model != "claude-3-haiku-20240307" {
		t.Errorf("expected claude-3-haiku, got %s", s.config.Model)
	}
}

func TestEnrichSpecEmpty(t *testing.T) {
	s := &LLMService{}
	_, err := s.EnrichSpecContext(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty spec")
	}
}

func TestEnrichSpecWithLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "enriched content"}},
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

	result, err := svc.EnrichSpec("project: test\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "enriched content" {
		t.Errorf("expected enriched content, got %s", result)
	}
}

func TestExplainRulesArchitecture(t *testing.T) {
	svc := NewService()
	exp, err := svc.Explain("architecture", "")
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

func TestExplainRulesKernel(t *testing.T) {
	svc := NewService()
	exp, err := svc.Explain("kernel", "")
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

func TestSuggestPrivilegedPort(t *testing.T) {
	svc := NewService()
	spec := "project: test\ndeployment: rolling\nport: 80\n"
	suggestions, err := svc.Suggest(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, s := range suggestions {
		if s.Category == "security" && strings.Contains(s.Description, "80") {
			found = true
		}
	}
	if !found {
		t.Error("expected security suggestion for privileged port 80")
	}
}

func TestSuggestLooksGood(t *testing.T) {
	svc := NewService()
	spec := "architecture: hexagonal\ndeployment: rolling\ntesting: unit\ndescription: test\nport:\n  main: 8080\nname:\n  - a\n  - b\n  - c\n  - d\n  - e\n"
	suggestions, err := svc.Suggest(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, s := range suggestions {
		if s.Category == "general" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'looks good' suggestion when nothing is missing")
	}
}

func TestSuggestMultipleNames(t *testing.T) {
	svc := NewService()
	spec := "name: a\nname: b\nname: c\nname: d\nname: e\nname: f\n"
	suggestions, err := svc.Suggest(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, s := range suggestions {
		if s.Category == "structure" && strings.Contains(s.Description, "logical modules") {
			found = true
		}
	}
	if !found {
		t.Error("expected structure suggestion for >5 names")
	}
}
