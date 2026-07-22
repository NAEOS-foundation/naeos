package adapters

import (
	"github.com/NAEOS-foundation/naeos/internal/ai"
	"github.com/NAEOS-foundation/naeos/internal/promptlib"
)

// newMockLLMService creates a mock LLM service for testing.
// It doesn't make real HTTP calls but allows the adapter to be constructed.
func newMockLLMService() *ai.LLMService {
	lib, _ := promptlib.New()
	return ai.NewLLMService(ai.LLMConfig{
		Provider:  ai.ProviderOpenAI,
		APIKey:    "test-key",
		Model:     "gpt-4",
		MaxTokens: 1000,
	}, lib)
}
