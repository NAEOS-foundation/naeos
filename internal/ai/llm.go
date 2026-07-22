package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	Stream      bool          `json:"stream,omitempty"`
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

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewLLMService creates an LLM service with the given configuration.
func NewLLMService(config LLMConfig, lib ...*promptlib.Library) *LLMService {
	if config.APIKey == "" && (config.Provider == ProviderOpenAI || config.Provider == ProviderAnthropic) {
		slog.Warn("LLM API key is empty; requests will likely fail", "provider", config.Provider)
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	return s.EnrichSpecContext(ctx, specContent)
}

// EnrichSpecContext enriches a specification with context support.
func (s *LLMService) EnrichSpecContext(ctx context.Context, specContent string) (string, error) {
	if specContent == "" {
		return "", fmt.Errorf("empty specification")
	}
	prompt := s.buildEnrichPrompt(specContent)
	return s.callLLM(ctx, prompt)
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
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	return s.GenerateSuggestionsContext(ctx, specContent)
}

// GenerateSuggestionsContext asks the LLM for improvement suggestions with context support.
func (s *LLMService) GenerateSuggestionsContext(ctx context.Context, specContent string) ([]Suggestion, error) {
	if specContent == "" {
		return nil, fmt.Errorf("empty specification")
	}
	prompt := s.buildSuggestionsPrompt(specContent)

	response, err := s.callLLM(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var suggestions []Suggestion
	if err := json.Unmarshal([]byte(CleanJSON(response)), &suggestions); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	return s.ExplainArchitectureContext(ctx, specContent, architecture)
}

// ExplainArchitectureContext explains an architecture pattern with context support.
func (s *LLMService) ExplainArchitectureContext(ctx context.Context, specContent, architecture string) (string, error) {
	if specContent == "" {
		return "", fmt.Errorf("empty specification")
	}
	prompt := s.buildExplainPrompt(specContent, architecture)
	return s.callLLM(ctx, prompt)
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

// modelContextWindows maps known model names to their context window sizes (in tokens).
var modelContextWindows = map[string]int{
	"gpt-4o":                     128000,
	"gpt-4o-mini":                128000,
	"gpt-4-turbo":                128000,
	"gpt-4":                      8192,
	"gpt-3.5-turbo":              16385,
	"claude-3-opus-20240229":     200000,
	"claude-3-sonnet-20240229":   200000,
	"claude-3-haiku-20240307":    200000,
	"claude-3-5-sonnet-20241022": 200000,
	"claude-3-5-haiku-20241022":  200000,
	"llama3.2":                   8192,
	"llama3.1":                   8192,
	"llama3":                     8192,
	"mistral":                    8192,
	"codellama":                  16384,
	"mixtral":                    32768,
}

// estimateTokens returns a rough estimate of the number of tokens in a string.
// Uses the common heuristic of ~4 characters per token for English text.
func estimateTokens(s string) int {
	return len(s) / 4
}

// truncatePrompt truncates the prompt to fit within the model's context window,
// reserving config.MaxTokens output tokens.
func (s *LLMService) truncatePrompt(prompt string) string {
	window, ok := modelContextWindows[s.config.Model]
	if !ok {
		window = 8192
	}

	available := window - s.config.MaxTokens
	if available < 256 {
		available = 256
	}

	estimated := estimateTokens(prompt)
	if estimated <= available {
		return prompt
	}

	maxChars := available * 4
	truncated := prompt[:maxChars]

	slog.Warn("prompt truncated",
		"model", s.config.Model,
		"estimated_tokens", estimated,
		"context_window", window,
		"available_tokens", available,
	)

	return truncated
}

func (s *LLMService) callLLM(ctx context.Context, prompt string) (string, error) {
	prompt = s.truncatePrompt(prompt)
	switch s.config.Provider {
	case ProviderOpenAI, ProviderOllama:
		return s.callOpenAI(ctx, prompt)
	case ProviderAnthropic:
		return s.callAnthropic(ctx, prompt)
	default:
		slog.Error("unsupported LLM provider", "provider", s.config.Provider)
		return "", fmt.Errorf("unsupported LLM provider: %s", s.config.Provider)
	}
}

func (s *LLMService) callOpenAI(ctx context.Context, prompt string) (string, error) {
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

	reqCtx := ctx
	if s.config.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		slog.Error("openai request failed", "error", err)
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("openai read response failed", "error", err)
		return "", err
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		slog.Error("openai parse response failed", "error", err)
		return "", err
	}

	if result.Error != nil {
		slog.Error("openai api error", "message", result.Error.Message)
		return "", fmt.Errorf("openai error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		slog.Error("openai no choices returned")
		return "", fmt.Errorf("openai: no choices returned")
	}

	slog.Info("openai call succeeded", "model", s.config.Model)
	return result.Choices[0].Message.Content, nil
}

func (s *LLMService) streamOpenAI(ctx context.Context, prompt string, w io.Writer) error {
	flusher, flushable := w.(http.Flusher)

	writeEvent := func(event, data string) error {
		_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
		if flushable {
			flusher.Flush()
		}
		return err
	}

	if err := writeEvent("start", "{}"); err != nil {
		return err
	}

	reqBody := openAIRequest{
		Model: s.config.Model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   s.config.MaxTokens,
		Temperature: 0.3,
		Stream:      true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	baseURL := s.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	url := baseURL + "/v1/chat/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		_ = writeEvent("error", fmt.Sprintf(`{"message":"%s"}`, err.Error()))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		_ = writeEvent("error", fmt.Sprintf(`{"message":"API error (status %d): %s"}`, resp.StatusCode, string(respBody)))
		return fmt.Errorf("openai streaming: status %d", resp.StatusCode)
	}

	decoder := NewSSEDecoder(resp.Body)
	for {
		event, data, err := decoder.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_ = writeEvent("error", fmt.Sprintf(`{"message":"%s"}`, err.Error()))
			return err
		}
		if event != "data" {
			continue
		}

		if string(data) == "[DONE]" {
			break
		}

		var chunk openAIStreamChunk
		if err := json.Unmarshal(data, &chunk); err != nil {
			slog.Warn("failed to parse OpenAI SSE data chunk", "error", err)
			continue
		}

		if chunk.Error != nil {
			_ = writeEvent("error", fmt.Sprintf(`{"message":"%s"}`, chunk.Error.Message))
			return fmt.Errorf("openai streaming error: %s", chunk.Error.Message)
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			escaped := strings.ReplaceAll(chunk.Choices[0].Delta.Content, "\n", "\\n")
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			if err := writeEvent("chunk", fmt.Sprintf(`{"text":"%s"}`, escaped)); err != nil {
				return err
			}
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != nil {
			break
		}
	}

	slog.Info("openai stream completed")
	return writeEvent("done", `{"status":"completed"}`)
}

func (s *LLMService) streamAnthropic(ctx context.Context, prompt string, w io.Writer) error {
	flusher, flushable := w.(http.Flusher)

	writeEvent := func(event, data string) error {
		_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
		if flushable {
			flusher.Flush()
		}
		return err
	}

	if err := writeEvent("start", "{}"); err != nil {
		return err
	}

	reqBody := anthropicRequest{
		Model:     s.config.Model,
		Messages:  []chatMessage{{Role: "user", Content: prompt}},
		MaxTokens: s.config.MaxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	baseURL := s.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	url := baseURL + "/v1/messages?stream=true"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		_ = writeEvent("error", fmt.Sprintf(`{"message":"%s"}`, err.Error()))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		_ = writeEvent("error", fmt.Sprintf(`{"message":"Anthropic error (status %d): %s"}`, resp.StatusCode, string(respBody)))
		return fmt.Errorf("anthropic streaming: status %d", resp.StatusCode)
	}

	decoder := NewSSEDecoder(resp.Body)
	for {
		event, data, err := decoder.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_ = writeEvent("error", fmt.Sprintf(`{"message":"%s"}`, err.Error()))
			return err
		}

		switch event {
		case "content_block_delta":
			var delta struct {
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal(data, &delta); err == nil && delta.Delta.Text != "" {
				escaped := strings.ReplaceAll(delta.Delta.Text, "\n", "\\n")
				escaped = strings.ReplaceAll(escaped, `"`, `\"`)
				if err := writeEvent("chunk", fmt.Sprintf(`{"text":"%s"}`, escaped)); err != nil {
					return err
				}
			} else if err != nil {
				slog.Debug("failed to parse anthropic delta", "error", err)
			}
		case "error":
			var errResp struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(data, &errResp); err == nil && errResp.Error.Message != "" {
				_ = writeEvent("error", fmt.Sprintf(`{"message":"%s"}`, errResp.Error.Message))
				return fmt.Errorf("anthropic streaming error: %s", errResp.Error.Message)
			}
		case "message_stop":
			slog.Info("anthropic stream completed")
			return writeEvent("done", `{"status":"completed"}`)
		}
	}

	slog.Info("anthropic stream completed")
	return writeEvent("done", `{"status":"completed"}`)
}

func (s *LLMService) callAnthropic(ctx context.Context, prompt string) (string, error) {
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

	reqCtx := ctx
	if s.config.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewReader(body))
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

func (s *LLMService) buildCompilerPrompt(target string, neirContext string) string {
	if s.library != nil {
		rendered, err := s.library.RenderLLM("compile-spec", map[string]any{
			"Target":      target,
			"NEIRContext": neirContext,
		})
		if err == nil && rendered.User != "" {
			return rendered.User
		}
	}

	return fmt.Sprintf(`You are a compiler that converts NAEOS architectural specifications into configuration files for %s AI coding assistants.
Generate the most effective rules, context, and instruction files that capture the project structure accurately.
Output format: JSON array of files, each with "path", "content", and "kind" fields.
Only output valid JSON, no explanations.

Project Context:
%s

Generate the %s compiler output:`, target, neirContext, target)
}

func (s *LLMService) StreamCompileSpec(ctx context.Context, target, specContent string, w io.Writer) error {
	prompt := s.buildCompilerPrompt(target, specContent)
	return s.streamLLM(ctx, prompt, w)
}

func (s *LLMService) StreamEnrichSpec(ctx context.Context, specContent string, w io.Writer) error {
	prompt := s.buildEnrichPrompt(specContent)
	return s.streamLLM(ctx, prompt, w)
}

func (s *LLMService) StreamExplainArchitecture(ctx context.Context, specContent, architecture string, w io.Writer) error {
	prompt := s.buildExplainPrompt(specContent, architecture)
	return s.streamLLM(ctx, prompt, w)
}

func (s *LLMService) streamLLM(ctx context.Context, prompt string, w io.Writer) error {
	prompt = s.truncatePrompt(prompt)

	switch s.config.Provider {
	case ProviderOpenAI, ProviderOllama:
		return s.streamOpenAI(ctx, prompt, w)
	case ProviderAnthropic:
		return s.streamAnthropic(ctx, prompt, w)
	default:
		return fmt.Errorf("unsupported LLM provider for streaming: %s", s.config.Provider)
	}
}

// SSEDecoder parses Server-Sent Events from a stream.
type SSEDecoder struct {
	r   io.Reader
	buf []byte
}

func NewSSEDecoder(r io.Reader) *SSEDecoder {
	return &SSEDecoder{r: r}
}

func (d *SSEDecoder) Decode() (event string, data []byte, err error) {
	d.buf = d.buf[:0]
	for {
		var line []byte
		line, err = d.readLine()
		if err != nil {
			return
		}
		if len(line) == 0 {
			if len(d.buf) > 0 {
				return event, d.buf, nil
			}
			continue
		}
		if line[0] == ':' {
			continue
		}
		parts := splitSSELine(string(line))
		if len(parts) < 2 {
			continue
		}
		switch parts[0] {
		case "event":
			event = parts[1]
		case "data":
			if len(d.buf) > 0 {
				d.buf = append(d.buf, '\n')
			}
			d.buf = append(d.buf, []byte(parts[1])...)
		}
	}
}

func (d *SSEDecoder) readLine() ([]byte, error) {
	var line []byte
	for {
		b := make([]byte, 1)
		n, err := d.r.Read(b)
		if err != nil {
			return line, err
		}
		if n == 0 {
			return line, io.ErrNoProgress
		}
		if b[0] == '\n' {
			return line, nil
		}
		if b[0] != '\r' {
			line = append(line, b[0])
		}
	}
}

func splitSSELine(line string) []string {
	idx := strings.IndexByte(line, ':')
	if idx < 0 {
		return nil
	}
	field := strings.TrimSpace(line[:idx])
	value := line[idx+1:]
	if len(value) > 0 && value[0] == ' ' {
		value = value[1:]
	}
	return []string{field, value}
}

func CleanJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
