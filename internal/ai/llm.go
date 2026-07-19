package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/promptlib"
)

// LLMProvider identifies a supported LLM provider.
type LLMProvider string

const (
	// ProviderOpenAI is the OpenAI LLM provider.
	ProviderOpenAI LLMProvider = "openai"
	// ProviderAnthropic is the Anthropic LLM provider.
	ProviderAnthropic LLMProvider = "anthropic"
	// ProviderOllama is the Ollama local LLM provider (OpenAI-compatible).
	ProviderOllama LLMProvider = "ollama"
)

// LLMConfig holds configuration for an LLM service.
type LLMConfig struct {
	Provider  LLMProvider
	APIKey    string
	Model     string
	MaxTokens int
	Timeout   time.Duration
	BaseURL   string
}

// LLMService communicates with external LLM APIs to enrich specifications.
type LLMService struct {
	config     LLMConfig
	httpClient *http.Client
	library    *promptlib.Library
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type anthropicRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	Messages  []chatMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewLLMService creates an LLM service with the given configuration.
func NewLLMService(config LLMConfig, lib ...*promptlib.Library) *LLMService {
	if config.Model == "" {
		switch config.Provider {
		case ProviderOpenAI:
			config.Model = "gpt-4o-mini"
		case ProviderAnthropic:
			config.Model = "claude-3-haiku-20240307"
		case ProviderOllama:
			config.Model = "llama3.2"
		}
	}
	if config.BaseURL == "" && config.Provider == ProviderOllama {
		config.BaseURL = "http://localhost:11434"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1024
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	s := &LLMService{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
	if len(lib) > 0 && lib[0] != nil {
		s.library = lib[0]
	}
	return s
}

// EnrichSpec sends a specification to the LLM for enhancement with best practices.
func (s *LLMService) EnrichSpec(specContent string) (string, error) {
	prompt := s.buildEnrichPrompt(specContent)
	return s.callLLM(prompt)
}

func (s *LLMService) buildEnrichPrompt(specContent string) string {
	if s.library != nil {
		rendered, err := s.library.RenderLLM("enrich-spec", map[string]any{
			"SpecContent": specContent,
		})
		if err == nil && rendered.User != "" {
			return rendered.User
		}
	}

	return fmt.Sprintf(`You are a platform engineering expert. Analyze this NAEOS specification and enrich it with best practices.
Add any missing sections that would improve the specification. Keep the existing content intact.
Only output the enriched YAML specification, no explanations.

Specification:
%s`, specContent)
}

// GenerateSuggestions asks the LLM to produce improvement suggestions for a specification.
func (s *LLMService) GenerateSuggestions(specContent string) ([]Suggestion, error) {
	prompt := s.buildSuggestionsPrompt(specContent)

	response, err := s.callLLM(prompt)
	if err != nil {
		return nil, err
	}

	var suggestions []Suggestion
	if err := json.Unmarshal([]byte(cleanJSON(response)), &suggestions); err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	return suggestions, nil
}

func (s *LLMService) buildSuggestionsPrompt(specContent string) string {
	if s.library != nil {
		rendered, err := s.library.RenderLLM("generate-suggestions", map[string]any{
			"SpecContent": specContent,
		})
		if err == nil && rendered.User != "" {
			return rendered.User
		}
	}

	return fmt.Sprintf(`Analyze this NAEOS specification and return a JSON array of suggestions.
Each suggestion should have: category, title, description, priority (high/medium/low).
Return ONLY the JSON array, no other text.

Specification:
%s`, specContent)
}

// ExplainArchitecture asks the LLM to explain an architecture pattern in the context of the specification.
func (s *LLMService) ExplainArchitecture(specContent, architecture string) (string, error) {
	prompt := s.buildExplainPrompt(specContent, architecture)
	return s.callLLM(prompt)
}

func (s *LLMService) buildExplainPrompt(specContent, arch string) string {
	if s.library != nil {
		rendered, err := s.library.RenderLLM("explain-architecture", map[string]any{
			"SpecContent":  specContent,
			"Architecture": arch,
		})
		if err == nil && rendered.User != "" {
			return rendered.User
		}
	}

	return fmt.Sprintf(`Explain the architecture pattern "%s" in the context of this specification.
Provide a clear, concise explanation suitable for a developer.

Specification:
%s

Architecture explanation:`, arch, specContent)
}

func (s *LLMService) callLLM(prompt string) (string, error) {
	switch s.config.Provider {
	case ProviderOpenAI, ProviderOllama:
		return s.callOpenAI(prompt)
	case ProviderAnthropic:
		return s.callAnthropic(prompt)
	default:
		return "", fmt.Errorf("unsupported LLM provider: %s", s.config.Provider)
	}
}

func (s *LLMService) callOpenAI(prompt string) (string, error) {
	reqBody := openAIRequest{
		Model: s.config.Model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   s.config.MaxTokens,
		Temperature: 0.3,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	baseURL := s.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	url := baseURL + "/v1/chat/completions"

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("openai error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices returned")
	}

	return result.Choices[0].Message.Content, nil
}

func (s *LLMService) callAnthropic(prompt string) (string, error) {
	reqBody := anthropicRequest{
		Model: s.config.Model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: s.config.MaxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	baseURL := s.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	url := baseURL + "/v1/messages"

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", result.Error.Message)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("anthropic: no content returned")
	}

	return result.Content[0].Text, nil
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
