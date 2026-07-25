package lsp

import (
	"strings"
)

type NEIRSuggestion struct {
	Label         string
	Kind          CompletionKind
	Detail        string
	Documentation string
	InsertText    string
}

type CompletionProvider struct {
	suggestions []NEIRSuggestion
	keywords    map[string][]NEIRSuggestion
}

func NewCompletionProvider() *CompletionProvider {
	p := &CompletionProvider{
		keywords: make(map[string][]NEIRSuggestion),
	}
	p.init()
	return p
}

func (p *CompletionProvider) init() {
	p.suggestions = []NEIRSuggestion{
		{Label: "project", Kind: CompletionField, Detail: "string", Documentation: "Project name (required)"},
		{Label: "version", Kind: CompletionField, Detail: "string", Documentation: "Project version (semver)"},
		{Label: "description", Kind: CompletionField, Detail: "string", Documentation: "Project description"},
		{Label: "modules", Kind: CompletionProperty, Detail: "[]Module", Documentation: "Code modules"},
		{Label: "services", Kind: CompletionProperty, Detail: "[]Service", Documentation: "Runnable services"},
		{Label: "architecture", Kind: CompletionProperty, Detail: "Architecture", Documentation: "Architecture pattern"},
		{Label: "deployment", Kind: CompletionProperty, Detail: "Deployment", Documentation: "Deployment configuration"},
		{Label: "testing", Kind: CompletionProperty, Detail: "Testing", Documentation: "Testing strategy"},
		{Label: "generation", Kind: CompletionProperty, Detail: "Generation", Documentation: "Code generation config"},
		{Label: "domain", Kind: CompletionProperty, Detail: "Domain", Documentation: "Domain-driven design"},
		{Label: "security", Kind: CompletionProperty, Detail: "Security", Documentation: "Security configuration"},
		{Label: "infrastructure", Kind: CompletionProperty, Detail: "Infrastructure", Documentation: "Cloud infrastructure"},
		{Label: "storage", Kind: CompletionProperty, Detail: "[]Storage", Documentation: "Data stores"},
		{Label: "components", Kind: CompletionProperty, Detail: "[]Component", Documentation: "Internal components"},
		{Label: "api", Kind: CompletionProperty, Detail: "[]API", Documentation: "API definitions"},
	}

	p.keywords["service"] = []NEIRSuggestion{
		{Label: "name", Kind: CompletionField, Detail: "string", Documentation: "Service name"},
		{Label: "kind", Kind: CompletionEnum, Detail: "string", Documentation: "Service kind: http, grpc, worker, cli, job"},
		{Label: "port", Kind: CompletionField, Detail: "integer", Documentation: "Service port (1-65535)"},
		{Label: "description", Kind: CompletionField, Detail: "string", Documentation: "Service description"},
		{Label: "endpoints", Kind: CompletionProperty, Detail: "[]Endpoint", Documentation: "Service endpoints"},
	}

	p.keywords["endpoint"] = []NEIRSuggestion{
		{Label: "method", Kind: CompletionEnum, Detail: "string", Documentation: "HTTP method: GET, POST, PUT, DELETE, PATCH"},
		{Label: "path", Kind: CompletionField, Detail: "string", Documentation: "Endpoint path"},
		{Label: "action", Kind: CompletionField, Detail: "string", Documentation: "Handler action name"},
		{Label: "summary", Kind: CompletionField, Detail: "string", Documentation: "Endpoint description"},
	}

	p.keywords["module"] = []NEIRSuggestion{
		{Label: "name", Kind: CompletionField, Detail: "string", Documentation: "Module name"},
		{Label: "path", Kind: CompletionField, Detail: "string", Documentation: "Module path (e.g. ./internal/auth)"},
		{Label: "description", Kind: CompletionField, Detail: "string", Documentation: "Module description"},
		{Label: "dependencies", Kind: CompletionProperty, Detail: "[]string", Documentation: "Module dependencies"},
	}

	p.keywords["architecture"] = []NEIRSuggestion{
		{Label: "pattern", Kind: CompletionEnum, Detail: "string", Documentation: "Pattern: layered, clean, hexagonal, microkernel, event-driven, cqrs, monolith"},
		{Label: "description", Kind: CompletionField, Detail: "string", Documentation: "Architecture description"},
		{Label: "principles", Kind: CompletionProperty, Detail: "[]string", Documentation: "Architecture principles"},
	}

	p.keywords["deployment"] = []NEIRSuggestion{
		{Label: "strategy", Kind: CompletionEnum, Detail: "string", Documentation: "Strategy: rolling, blue-green, canary, recreate"},
		{Label: "environments", Kind: CompletionProperty, Detail: "[]Environment", Documentation: "Deployment environments"},
	}

	p.keywords["testing"] = []NEIRSuggestion{
		{Label: "strategy", Kind: CompletionEnum, Detail: "string", Documentation: "Strategy: unit, integration, e2e, contract"},
		{Label: "coverage", Kind: CompletionField, Detail: "string", Documentation: "Coverage target (e.g. 80%)"},
	}

	p.keywords["generation"] = []NEIRSuggestion{
		{Label: "languages", Kind: CompletionProperty, Detail: "[]string", Documentation: "Target languages: go, typescript, python, java, rust"},
		{Label: "output_dir", Kind: CompletionField, Detail: "string", Documentation: "Output directory"},
	}

	p.keywords["kind_values"] = []NEIRSuggestion{
		{Label: "http", Kind: CompletionEnumMember, Detail: "Service kind", Documentation: "HTTP service"},
		{Label: "grpc", Kind: CompletionEnumMember, Detail: "Service kind", Documentation: "gRPC service"},
		{Label: "worker", Kind: CompletionEnumMember, Detail: "Service kind", Documentation: "Background worker"},
		{Label: "cli", Kind: CompletionEnumMember, Detail: "Service kind", Documentation: "CLI tool"},
		{Label: "job", Kind: CompletionEnumMember, Detail: "Service kind", Documentation: "Scheduled job"},
	}

	p.keywords["method_values"] = []NEIRSuggestion{
		{Label: "GET", Kind: CompletionEnumMember, Detail: "HTTP method"},
		{Label: "POST", Kind: CompletionEnumMember, Detail: "HTTP method"},
		{Label: "PUT", Kind: CompletionEnumMember, Detail: "HTTP method"},
		{Label: "DELETE", Kind: CompletionEnumMember, Detail: "HTTP method"},
		{Label: "PATCH", Kind: CompletionEnumMember, Detail: "HTTP method"},
	}

	p.keywords["pattern_values"] = []NEIRSuggestion{
		{Label: "layered", Kind: CompletionEnumMember, Detail: "Architecture pattern"},
		{Label: "clean", Kind: CompletionEnumMember, Detail: "Architecture pattern"},
		{Label: "hexagonal", Kind: CompletionEnumMember, Detail: "Architecture pattern"},
		{Label: "microkernel", Kind: CompletionEnumMember, Detail: "Architecture pattern"},
		{Label: "event-driven", Kind: CompletionEnumMember, Detail: "Architecture pattern"},
		{Label: "cqrs", Kind: CompletionEnumMember, Detail: "Architecture pattern"},
		{Label: "monolith", Kind: CompletionEnumMember, Detail: "Architecture pattern"},
	}

	p.keywords["strategy_values"] = []NEIRSuggestion{
		{Label: "rolling", Kind: CompletionEnumMember, Detail: "Deployment strategy"},
		{Label: "blue-green", Kind: CompletionEnumMember, Detail: "Deployment strategy"},
		{Label: "canary", Kind: CompletionEnumMember, Detail: "Deployment strategy"},
		{Label: "recreate", Kind: CompletionEnumMember, Detail: "Deployment strategy"},
	}

	p.keywords["language_values"] = []NEIRSuggestion{
		{Label: "go", Kind: CompletionEnumMember, Detail: "Go language"},
		{Label: "typescript", Kind: CompletionEnumMember, Detail: "TypeScript language"},
		{Label: "python", Kind: CompletionEnumMember, Detail: "Python language"},
		{Label: "java", Kind: CompletionEnumMember, Detail: "Java language"},
		{Label: "rust", Kind: CompletionEnumMember, Detail: "Rust language"},
	}
}

func (p *CompletionProvider) Provide(text string, line, character int) *CompletionList {
	lineText := p.lineAt(text, line)
	prefix := p.prefixBeforeChar(lineText, character)

	beforeColon := strings.TrimSpace(lineText[:character])
	if strings.Contains(beforeColon, ":") && !strings.HasSuffix(strings.TrimSpace(beforeColon), ":") {
		afterLastColon := beforeColon[strings.LastIndex(beforeColon, ":")+1:]
		if strings.TrimSpace(afterLastColon) != "" {
			return &CompletionList{Items: []CompletionItem{}}
		}
	}

	contextKey := p.detectContext(text, line, lineText, prefix)

	var items []CompletionItem

	switch contextKey {
	case "kind":
		items = p.toItems(p.keywords["kind_values"])
	case "method":
		items = p.toItems(p.keywords["method_values"])
	case "pattern":
		items = p.toItems(p.keywords["pattern_values"])
	case "strategy":
		items = p.toItems(p.keywords["strategy_values"])
	case "language":
		items = p.toItems(p.keywords["language_values"])
	default:
		if contextKey != "" {
			if suggestions, ok := p.keywords[contextKey]; ok {
				items = p.toItems(suggestions)
			}
		}
		if len(items) == 0 {
			items = p.toItems(p.suggestions)
		}
	}

	if prefix != "" {
		lower := strings.ToLower(prefix)
		var filtered []CompletionItem
		for _, item := range items {
			if strings.HasPrefix(strings.ToLower(item.Label), lower) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	return &CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

func (p *CompletionProvider) toItems(suggestions []NEIRSuggestion) []CompletionItem {
	items := make([]CompletionItem, len(suggestions))
	for i, s := range suggestions {
		insertText := s.InsertText
		if insertText == "" {
			insertText = s.Label
		}
		items[i] = CompletionItem{
			Label:         s.Label,
			Kind:          s.Kind,
			Detail:        s.Detail,
			Documentation: s.Documentation,
			InsertText:    insertText,
		}
	}
	return items
}

func (p *CompletionProvider) lineAt(text string, line int) string {
	lines := strings.Split(text, "\n")
	if line >= 0 && line < len(lines) {
		return lines[line]
	}
	return ""
}

func (p *CompletionProvider) prefixBeforeChar(line string, character int) string {
	if character > len(line) {
		character = len(line)
	}
	before := line[:character]
	trimmed := strings.TrimLeft(before, " \t-")
	colonIdx := strings.Index(trimmed, ":")
	if colonIdx >= 0 {
		trimmed = strings.TrimSpace(trimmed[:colonIdx])
	}
	return trimmed
}

func (p *CompletionProvider) detectContext(text string, line int, lineText, prefix string) string {
	ltrim := strings.TrimLeft(lineText, " \t-")

	if strings.HasPrefix(ltrim, "kind:") {
		return "kind"
	}
	if strings.HasPrefix(ltrim, "method:") {
		return "method"
	}
	if strings.HasPrefix(ltrim, "pattern:") {
		return "pattern"
	}
	if strings.HasPrefix(ltrim, "strategy:") {
		return "strategy"
	}
	if strings.HasPrefix(ltrim, "languages:") || strings.HasPrefix(ltrim, "- ") && prefix == "" {
		return "language"
	}

	for i := line; i >= 0; i-- {
		l := p.lineAt(text, i)
		trim := strings.TrimSpace(l)
		keyOnly := strings.SplitN(trim, ":", 2)[0]
		keyOnly = strings.TrimPrefix(keyOnly, "- ")
		switch keyOnly {
		case "services":
			return "service"
		case "endpoints":
			return "endpoint"
		case "modules":
			return "module"
		case "architecture":
			return "architecture"
		case "deployment":
			return "deployment"
		case "testing":
			return "testing"
		case "generation":
			return "generation"
		}
	}

	return ""
}
