package mcp

import (
	"fmt"
	"sort"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
)

// maxCompletionValues caps the number of completion values returned per
// request, as recommended by the MCP specification.
const maxCompletionValues = 100

// CompletionResult is the payload returned by completion/complete.
type CompletionResult struct {
	Values  []string `json:"values"`
	Total   int      `json:"total,omitempty"`
	HasMore bool     `json:"hasMore,omitempty"`
}

// completeRequestParams are the parameters of a completion/complete request.
type completeRequestParams struct {
	Ref struct {
		Type string `json:"type"`
		Name string `json:"name"`
		URI  string `json:"uri"`
	} `json:"ref"`
	Argument struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"argument"`
}

// architecturePatterns returns all NEIR architecture patterns accepted by the
// validator, in deterministic order.
func architecturePatterns() []string {
	return []string{
		string(architecture.PatternLayered),
		string(architecture.PatternClean),
		string(architecture.PatternHexagonal),
		string(architecture.PatternMicrokernel),
		string(architecture.PatternEventDriven),
		string(architecture.PatternCQRS),
		string(architecture.PatternMonolith),
		string(architecture.PatternMonolithic),
		string(architecture.PatternMicroservices),
		string(architecture.PatternServerless),
	}
}

// complete resolves completions for a prompt argument or resource URI
// referenced by a completion/complete request.
func (s *Server) complete(refType, refName, argName, value string) (*CompletionResult, error) {
	switch refType {
	case "ref/prompt":
		return s.completePromptArgument(refName, argName, value)
	case "ref/resource":
		return s.completeResourceURI(value), nil
	default:
		return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("unsupported ref type %q; expected \"ref/prompt\" or \"ref/resource\"", refType))
	}
}

// completePromptArgument provides candidate values for a named argument of a
// builtin prompt template.
func (s *Server) completePromptArgument(promptName, argName, value string) (*CompletionResult, error) {
	prompts := builtinPrompts()
	var selected *Prompt
	for i := range prompts {
		if prompts[i].Name == promptName {
			selected = &prompts[i]
			break
		}
	}
	if selected == nil {
		names := make([]string, 0, len(prompts))
		for _, p := range prompts {
			names = append(names, p.Name)
		}
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("unknown prompt %q; available prompts: %s", promptName, strings.Join(names, ", ")))
	}

	hasArg := false
	for _, arg := range selected.Arguments {
		if arg.Name == argName {
			hasArg = true
			break
		}
	}
	if !hasArg {
		argNames := make([]string, 0, len(selected.Arguments))
		for _, arg := range selected.Arguments {
			argNames = append(argNames, arg.Name)
		}
		return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("prompt %q has no argument %q; available arguments: %s", promptName, argName, strings.Join(argNames, ", ")))
	}

	candidates := promptArgumentCandidates(promptName, argName)
	return filterCompletions(candidates, value), nil
}

// promptArgumentCandidates returns known values for arguments backed by
// enumerated domains; free-text arguments yield no candidates.
func promptArgumentCandidates(promptName, argName string) []string {
	switch {
	case promptName == "explain-architecture" && argName == "architecture":
		return architecturePatterns()
	default:
		return nil
	}
}

// completeResourceURI completes resource URIs against the server's current
// resource list using a case-insensitive prefix match.
func (s *Server) completeResourceURI(value string) *CompletionResult {
	resources := s.listResources()
	uris := make([]string, 0, len(resources))
	for _, r := range resources {
		uris = append(uris, r.URI)
	}
	return filterCompletions(uris, value)
}

// filterCompletions filters candidates by case-insensitive prefix match,
// sorts them deterministically, and applies the MCP result cap.
func filterCompletions(candidates []string, value string) *CompletionResult {
	lower := strings.ToLower(value)
	matches := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if strings.HasPrefix(strings.ToLower(c), lower) {
			matches = append(matches, c)
		}
	}
	sort.Strings(matches)

	result := &CompletionResult{Values: matches, Total: len(matches)}
	if len(result.Values) > maxCompletionValues {
		result.Values = result.Values[:maxCompletionValues]
		result.HasMore = true
	}
	return result
}
