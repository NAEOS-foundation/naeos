package ai

import (
	"fmt"
	"strings"
)

// AIService provides analysis and suggestions for NAEOS specifications.
// When an LLMService is configured, it uses LLM-powered analysis with rule-based fallback.
type AIService struct {
	context map[string]any
	llm     *LLMService
}

// NewService creates a new AI service instance with rule-based analysis.
func NewService() *AIService {
	return &AIService{
		context: make(map[string]any),
	}
}

// NewServiceWithLLM creates an AI service that uses LLM when available, with rule-based fallback.
func NewServiceWithLLM(llm *LLMService) *AIService {
	return &AIService{
		context: make(map[string]any),
		llm:     llm,
	}
}

// Suggestion represents an improvement recommendation for a specification.
type Suggestion struct {
	Category    string
	Title       string
	Description string
	Priority    string
}

// Explanation holds a structured explanation of a NAEOS concept.
type Explanation struct {
	Topic   string
	Content string
	Details []string
}

// Suggest analyses a specification and returns improvement suggestions.
// Uses LLM when available, falls back to rule-based analysis.
func (s *AIService) Suggest(specContent string) ([]Suggestion, error) {
	if specContent == "" {
		return nil, fmt.Errorf("empty specification")
	}

	if s.llm != nil {
		suggestions, err := s.llm.GenerateSuggestions(specContent)
		if err == nil && len(suggestions) > 0 {
			return suggestions, nil
		}
	}

	return s.suggestRules(specContent)
}

// suggestRules returns rule-based suggestions when LLM is unavailable.
func (s *AIService) suggestRules(specContent string) ([]Suggestion, error) {
	var suggestions []Suggestion

	if !strings.Contains(specContent, "architecture:") {
		suggestions = append(suggestions, Suggestion{
			Category:    "architecture",
			Title:       "Add architecture pattern",
			Description: "Consider adding an architecture section to define the structural pattern (hexagonal, clean, layered, etc.)",
			Priority:    "high",
		})
	}

	if !strings.Contains(specContent, "deployment:") {
		suggestions = append(suggestions, Suggestion{
			Category:    "deployment",
			Title:       "Add deployment configuration",
			Description: "Add a deployment section to specify the deployment strategy (rolling, blue-green, canary, etc.)",
			Priority:    "medium",
		})
	}

	if !strings.Contains(specContent, "testing:") {
		suggestions = append(suggestions, Suggestion{
			Category:    "testing",
			Title:       "Add testing configuration",
			Description: "Add a testing section to define test strategy and coverage goals",
			Priority:    "medium",
		})
	}

	if strings.Contains(specContent, "port:") {
		lines := strings.Split(specContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, "port:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					portStr := strings.TrimSpace(parts[1])
					var port int
					if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
						if port < 1024 {
							suggestions = append(suggestions, Suggestion{
								Category:    "security",
								Title:       "Consider using a non-privileged port",
								Description: fmt.Sprintf("Port %d requires root privileges. Consider using a port above 1024.", port),
								Priority:    "high",
							})
						}
					}
				}
			}
		}
	}

	if !strings.Contains(specContent, "description:") {
		suggestions = append(suggestions, Suggestion{
			Category:    "documentation",
			Title:       "Add project description",
			Description: "Add a description field to document the project purpose",
			Priority:    "low",
		})
	}

	if strings.Count(specContent, "name:") > 5 {
		suggestions = append(suggestions, Suggestion{
			Category:    "structure",
			Title:       "Consider splitting into modules",
			Description: "Your specification has many named entities. Consider organizing them into logical modules.",
			Priority:    "medium",
		})
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, Suggestion{
			Category:    "general",
			Title:       "Specification looks good",
			Description: "No major issues found. The specification covers the essential sections.",
			Priority:    "low",
		})
	}

	return suggestions, nil
}

// Explain returns a detailed explanation of the given topic.
// Uses LLM when available for architecture topics, falls back to built-in knowledge.
func (s *AIService) Explain(topic, specContent string) (*Explanation, error) {
	if topic == "" {
		return nil, fmt.Errorf("topic is required")
	}

	if s.llm != nil && specContent != "" && strings.ToLower(topic) == "architecture" {
		content, err := s.llm.ExplainArchitecture(specContent, topic)
		if err == nil && content != "" {
			return &Explanation{
				Topic:   topic,
				Content: content,
			}, nil
		}
	}

	return s.explainRules(topic)
}

// explainRules returns built-in explanations when LLM is unavailable.
func (s *AIService) explainRules(topic string) (*Explanation, error) {
	exp := &Explanation{
		Topic: topic,
	}

	switch strings.ToLower(topic) {
	case "pipeline":
		exp.Content = "The NAEOS pipeline transforms specifications into validated project artifacts."
		exp.Details = []string{
			"1. Parse: YAML/JSON specification is parsed into a structured document",
			"2. Normalize: Apply defaults and fill gaps in the specification",
			"3. Resolve: Cross-reference dependencies and validate references",
			"4. Build NEIR: Create the NAEOS Engineering Intermediate Representation",
			"5. Validate: Check the NEIR model for correctness",
			"6. Plan: Create an execution graph with dependency ordering",
			"7. Generate: Produce artifacts for the target language(s)",
			"8. Review: Check artifacts for governance compliance",
		}
	case "neir":
		exp.Content = "NEIR (NAEOS Engineering Intermediate Representation) is the internal model that captures all aspects of a project."
		exp.Details = []string{
			"Contains 16 domain-specific sub-models:",
			"- project, architecture, domain, module, component",
			"- service, api, storage, infrastructure, security",
			"- ai, docs, deployment, testing, metadata, generation",
		}
	case "architecture":
		exp.Content = "Architecture patterns define the structural organization of the generated project."
		exp.Details = []string{
			"Supported patterns:",
			"- hexagonal: Ports and adapters architecture",
			"- layered: Traditional N-tier architecture",
			"- clean: Clean architecture with dependency rule",
			"- event-driven: Event sourcing and CQRS patterns",
			"- monolith: Single deployable unit",
		}
	case "kernel":
		exp.Content = "The kernel provides runtime services for the pipeline execution."
		exp.Details = []string{
			"Services: parser, normalizer, resolver, builder, validator, scheduler, generator, renderer",
			"Features: service registry, lifecycle management, event bus, telemetry",
		}
	default:
		exp.Content = fmt.Sprintf("Topic '%s' is recognized but detailed documentation is not yet available.", topic)
		exp.Details = []string{
			"Available topics: pipeline, neir, architecture, kernel",
		}
	}

	return exp, nil
}
