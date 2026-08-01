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

	p.keywords["domain"] = []NEIRSuggestion{
		{Label: "events", Kind: CompletionProperty, Detail: "[]DomainEvent", Documentation: "Domain events"},
		{Label: "bounded_contexts", Kind: CompletionProperty, Detail: "[]BoundedContext", Documentation: "Bounded contexts"},
		{Label: "ubiquitous_language", Kind: CompletionField, Detail: "string", Documentation: "Ubiquitous language glossary"},
		{Label: "aggregates", Kind: CompletionProperty, Detail: "[]Aggregate", Documentation: "Domain aggregates"},
		{Label: "value_objects", Kind: CompletionProperty, Detail: "[]ValueObject", Documentation: "Value objects"},
	}

	p.keywords["security"] = []NEIRSuggestion{
		{Label: "authentication", Kind: CompletionProperty, Detail: "AuthConfig", Documentation: "Authentication configuration"},
		{Label: "authorization", Kind: CompletionProperty, Detail: "AuthzConfig", Documentation: "Authorization configuration"},
		{Label: "encryption", Kind: CompletionProperty, Detail: "EncryptionConfig", Documentation: "Encryption settings"},
		{Label: "cors", Kind: CompletionProperty, Detail: "CORSConfig", Documentation: "CORS configuration"},
		{Label: "rate_limiting", Kind: CompletionProperty, Detail: "RateLimitConfig", Documentation: "Rate limiting settings"},
	}

	p.keywords["infrastructure"] = []NEIRSuggestion{
		{Label: "provider", Kind: CompletionEnum, Detail: "string", Documentation: "Cloud provider: aws, gcp, azure, local"},
		{Label: "region", Kind: CompletionField, Detail: "string", Documentation: "Cloud region"},
		{Label: "resources", Kind: CompletionProperty, Detail: "[]Resource", Documentation: "Infrastructure resources"},
		{Label: "kubernetes", Kind: CompletionProperty, Detail: "K8sConfig", Documentation: "Kubernetes configuration"},
		{Label: "networking", Kind: CompletionProperty, Detail: "NetworkConfig", Documentation: "Network configuration"},
	}

	p.keywords["storage"] = []NEIRSuggestion{
		{Label: "type", Kind: CompletionEnum, Detail: "string", Documentation: "Storage type: postgres, mysql, redis, s3, mongodb"},
		{Label: "connection", Kind: CompletionField, Detail: "string", Documentation: "Connection string"},
		{Label: "migrations", Kind: CompletionProperty, Detail: "[]Migration", Documentation: "Database migrations"},
		{Label: "backup", Kind: CompletionProperty, Detail: "BackupConfig", Documentation: "Backup configuration"},
		{Label: "pooling", Kind: CompletionProperty, Detail: "PoolConfig", Documentation: "Connection pooling"},
	}

	p.keywords["components"] = []NEIRSuggestion{
		{Label: "type", Kind: CompletionEnum, Detail: "string", Documentation: "Component type: service, worker, cron, event-handler"},
		{Label: "name", Kind: CompletionField, Detail: "string", Documentation: "Component name"},
		{Label: "path", Kind: CompletionField, Detail: "string", Documentation: "Component path"},
		{Label: "dependencies", Kind: CompletionProperty, Detail: "[]string", Documentation: "Component dependencies"},
	}

	p.keywords["api"] = []NEIRSuggestion{
		{Label: "version", Kind: CompletionField, Detail: "string", Documentation: "API version"},
		{Label: "endpoints", Kind: CompletionProperty, Detail: "[]Endpoint", Documentation: "API endpoints"},
		{Label: "format", Kind: CompletionEnum, Detail: "string", Documentation: "API format: rest, graphql, grpc"},
		{Label: "documentation", Kind: CompletionField, Detail: "string", Documentation: "API documentation path"},
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

	p.suggestions = append(p.suggestions, []NEIRSuggestion{
		{Label: "$if{condition}", Kind: CompletionSnippet, Detail: "Conditional block", InsertText: "$if{ $1 }:\n  $0\n$endif", Documentation: "Conditional section evaluated against env vars"},
		{Label: "$import", Kind: CompletionSnippet, Detail: "Import file", InsertText: "$import{$1}", Documentation: "Import another spec file"},
		{Label: "$include", Kind: CompletionSnippet, Detail: "Include raw file", InsertText: "$include{$1}", Documentation: "Include raw file content"},
		{Label: "$fn{upper}", Kind: CompletionSnippet, Detail: "Uppercase function", InsertText: "$fn{upper($1)}", Documentation: "Convert to uppercase"},
		{Label: "$fn{lower}", Kind: CompletionSnippet, Detail: "Lowercase function", InsertText: "$fn{lower($1)}", Documentation: "Convert to lowercase"},
		{Label: "$fn{slug}", Kind: CompletionSnippet, Detail: "Slugify function", InsertText: "$fn{slug($1)}", Documentation: "Convert to URL-safe slug"},
		{Label: "$fn{default}", Kind: CompletionSnippet, Detail: "Default value", InsertText: "$fn{default($1, $2)}", Documentation: "Return first non-empty value"},
		{Label: "$env{VAR}", Kind: CompletionSnippet, Detail: "Environment variable", InsertText: "$env{$1}", Documentation: "Reference environment variable"},
		{Label: "$ref{path}", Kind: CompletionSnippet, Detail: "Reference path", InsertText: "$ref{$1}", Documentation: "Reference another spec path"},
	}...)
}

func (p *CompletionProvider) Provide(text string, line, character int) *CompletionList {
	lineText := p.lineAt(text, line)

	contextKey := p.detectContext(text, line, lineText, "")
	valuePrefix, fieldPrefix := p.extractPrefixes(lineText, character)

	var items []CompletionItem

	switch contextKey {
	case "kind":
		items = p.toItemsFilter(p.keywords["kind_values"], valuePrefix)
	case "method":
		items = p.toItemsFilter(p.keywords["method_values"], valuePrefix)
	case "pattern":
		items = p.toItemsFilter(p.keywords["pattern_values"], valuePrefix)
	case "strategy":
		items = p.toItemsFilter(p.keywords["strategy_values"], valuePrefix)
	case "language":
		items = p.toItemsFilter(p.keywords["language_values"], valuePrefix)
	default:
		if contextKey != "" {
			if suggestions, ok := p.keywords[contextKey]; ok {
				items = p.toItemsFilter(suggestions, fieldPrefix)
			}
		}
		if len(items) == 0 {
			items = p.toItemsFilter(p.suggestions, fieldPrefix)
		}
	}

	return &CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

func (p *CompletionProvider) extractPrefixes(lineText string, character int) (valuePrefix, fieldPrefix string) {
	if character > len(lineText) {
		character = len(lineText)
	}
	before := lineText[:character]
	trimmed := strings.TrimLeft(before, " \t-")

	colonIdx := strings.Index(trimmed, ":")
	if colonIdx >= 0 {
		afterColon := strings.TrimSpace(trimmed[colonIdx+1:])
		return afterColon, ""
	}

	fieldTrimmed := strings.TrimSpace(trimmed)
	return "", fieldTrimmed
}

func (p *CompletionProvider) toItemsFilter(suggestions []NEIRSuggestion, prefix string) []CompletionItem {
	items := p.toItems(suggestions)
	if prefix == "" {
		return items
	}
	lower := strings.ToLower(prefix)
	var filtered []CompletionItem
	for _, item := range items {
		if strings.HasPrefix(strings.ToLower(item.Label), lower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
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
		case "domain":
			return "domain"
		case "security":
			return "security"
		case "infrastructure":
			return "infrastructure"
		case "storage":
			return "storage"
		case "components":
			return "components"
		case "api":
			return "api"
		}
	}

	if strings.HasPrefix(lineText, "  - ") || strings.HasPrefix(lineText, "    ") {
		for i := line - 1; i >= 0; i-- {
			l := p.lineAt(text, i)
			trim := strings.TrimSpace(l)
			if strings.HasPrefix(trim, "- name:") {
				keyOnly := strings.SplitN(trim, ":", 2)[0]
				keyOnly = strings.TrimPrefix(keyOnly, "- ")
				if keyOnly == "name" {
					for j := i - 1; j >= 0; j-- {
						l2 := p.lineAt(text, j)
						trim2 := strings.TrimSpace(l2)
						parentKey := strings.SplitN(trim2, ":", 2)[0]
						parentKey = strings.TrimPrefix(parentKey, "- ")
						switch parentKey {
						case "services":
							return "service"
						case "endpoints":
							return "endpoint"
						case "modules":
							return "module"
						case "domain":
							return "domain"
						case "security":
							return "security"
						case "infrastructure":
							return "infrastructure"
						case "storage":
							return "storage"
						case "components":
							return "components"
						case "api":
							return "api"
						}
					}
				}
			}
		}
	}

	return ""
}
