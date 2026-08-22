package mcp

import (
	"fmt"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// PromptArgument describes a single argument of an MCP prompt template.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// Prompt describes a reusable prompt template exposed via prompts/list.
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptMessage is a single message in a prompts/get response.
type PromptMessage struct {
	Role    string       `json:"role"`
	Content ContentBlock `json:"content"`
}

// GetPromptResult is the payload returned by prompts/get.
type GetPromptResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// builtinPrompts returns all prompt templates provided by the server,
// in deterministic order.
func builtinPrompts() []Prompt {
	return []Prompt{
		{
			Name:        "review-spec",
			Description: "Review a NAEOS specification for issues, gaps, and best-practice violations",
			Arguments: []PromptArgument{
				{Name: "spec", Description: "The YAML/JSON NAEOS specification content", Required: true},
			},
		},
		{
			Name:        "enrich-spec",
			Description: "Enrich a NAEOS specification with missing sections and best practices",
			Arguments: []PromptArgument{
				{Name: "spec", Description: "The YAML/JSON NAEOS specification content", Required: true},
			},
		},
		{
			Name:        "explain-architecture",
			Description: "Explain an architecture pattern in the context of a specification",
			Arguments: []PromptArgument{
				{Name: "spec", Description: "The YAML/JSON NAEOS specification content", Required: true},
				{Name: "architecture", Description: "Architecture pattern to explain (e.g. microservices, serverless)", Required: true},
			},
		},
		{
			Name:        "generate-spec",
			Description: "Draft a new NAEOS specification from a project description",
			Arguments: []PromptArgument{
				{Name: "description", Description: "Natural-language description of the project to build", Required: true},
			},
		},
	}
}

// getPrompt renders a prompt template with the supplied arguments and returns
// the resulting messages.
func getPrompt(name string, args map[string]any) (*GetPromptResult, error) {
	prompts := builtinPrompts()
	var selected *Prompt
	for i := range prompts {
		if prompts[i].Name == name {
			selected = &prompts[i]
			break
		}
	}
	if selected == nil {
		names := make([]string, 0, len(prompts))
		for _, p := range prompts {
			names = append(names, p.Name)
		}
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("unknown prompt %q; available prompts: %s", name, strings.Join(names, ", ")))
	}

	values := make(map[string]string, len(selected.Arguments))
	for _, arg := range selected.Arguments {
		v, _ := args[arg.Name].(string)
		if v == "" && arg.Required {
			return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("prompt %q: the '%s' argument is required", name, arg.Name))
		}
		values[arg.Name] = v
	}

	text := renderPromptTemplate(name, values)
	return &GetPromptResult{
		Description: selected.Description,
		Messages: []PromptMessage{
			{Role: "user", Content: ContentBlock{Type: "text", Text: text}},
		},
	}, nil
}

// renderPromptTemplate substitutes {{argument}} placeholders in the builtin
// template for the given prompt. Values are inserted verbatim; only literal
// "{{name}}" tokens are replaced.
func renderPromptTemplate(name string, values map[string]string) string {
	var tpl string
	switch name {
	case "review-spec":
		tpl = `Review this NAEOS specification as a platform engineering expert.
Identify issues, missing sections, and best-practice violations.
Order findings by severity (high, medium, low) and suggest a concrete fix for each.

Specification:
{{spec}}`
	case "enrich-spec":
		tpl = `Analyze this NAEOS specification and enrich it with best practices.
Add any missing sections that would improve the specification.
Keep the existing content intact and output only the enriched YAML specification.

Specification:
{{spec}}`
	case "explain-architecture":
		tpl = `Explain the architecture pattern "{{architecture}}" in the context of this specification.
Describe how the modules and services map to the pattern, its trade-offs,
and what changes (if any) would improve the fit.

Specification:
{{spec}}`
	case "generate-spec":
		tpl = `Draft a complete NAEOS specification (YAML) for the following project description.
Include project metadata, modules, services, architecture, deployment, and testing sections.
Output only the YAML specification.

Project description:
{{description}}`
	default:
		return ""
	}

	for key, value := range values {
		tpl = strings.ReplaceAll(tpl, "{{"+key+"}}", value)
	}
	return tpl
}
