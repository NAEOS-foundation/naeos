package lsp

import (
	"strings"
	"unicode"
)

type Handler struct {
	server *Server
}

func NewHandler(s *Server) *Handler {
	return &Handler{server: s}
}

func (h *Handler) Initialize(params InitializeParams) InitializeResult {
	return InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:           1,
			CompletionProvider:         &CompletionOptions{TriggerCharacters: []string{":", " ", "-"}},
			HoverProvider:              true,
			DefinitionProvider:         true,
			DocumentSymbolProvider:     true,
			SignatureHelpProvider:      &SignatureHelpOptions{TriggerCharacters: []string{"$", "{", "("}},
			CodeActionProvider:         &CodeActionOptions{CodeActionKinds: []CodeActionKind{CodeActionQuickFix, CodeActionSourceOrganize}},
			DocumentFormattingProvider: &DocumentFormattingOptions{},
		},
	}
}

func (h *Handler) DidOpen(params DidOpenTextDocumentParams) {
	h.server.Documents().Open(params.TextDocument.URI, params.TextDocument.Version, params.TextDocument.Text)
	h.publishDiagnostics(params.TextDocument.URI, params.TextDocument.Text)
}

func (h *Handler) DidChange(params DidChangeTextDocumentParams) {
	if len(params.ContentChanges) > 0 {
		h.server.Documents().Update(params.TextDocument.URI, params.TextDocument.Version, params.ContentChanges[0].Text)
		if doc, ok := h.server.Documents().Get(params.TextDocument.URI); ok {
			h.publishDiagnostics(params.TextDocument.URI, doc.Text)
		}
	}
}

func (h *Handler) DidClose(params DidCloseTextDocumentParams) {
	h.server.Documents().Close(params.TextDocument.URI)
}

func (h *Handler) Completion(params CompletionParams) *CompletionList {
	doc, ok := h.server.Documents().Get(params.TextDocument.URI)
	if !ok {
		return &CompletionList{Items: []CompletionItem{}}
	}
	return h.server.Completion().Provide(doc.Text, params.Position.Line, params.Position.Character)
}

func (h *Handler) Hover(params HoverParams) *Hover {
	doc, ok := h.server.Documents().Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	word := h.wordAtPosition(doc.Text, params.Position.Line, params.Position.Character)
	if word == "" {
		return nil
	}

	info := h.lookupDocumentation(word)
	if info == "" {
		return nil
	}

	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: info,
		},
	}
}

func (h *Handler) Definition(params DefinitionParams) *Location {
	doc, ok := h.server.Documents().Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	word := h.wordAtPosition(doc.Text, params.Position.Line, params.Position.Character)
	if word == "" {
		return nil
	}

	lines := strings.Split(doc.Text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, word+":") && i != params.Position.Line {
			col := strings.Index(line, word)
			if col >= 0 {
				return &Location{
					URI: params.TextDocument.URI,
					Range: Range{
						Start: Position{Line: i, Character: col},
						End:   Position{Line: i, Character: col + len(word)},
					},
				}
			}
		}
	}

	return nil
}

func (h *Handler) DocumentSymbol(params DocumentSymbolParams) []DocumentSymbol {
	doc, ok := h.server.Documents().Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	return h.buildSymbols(doc.Text)
}

func (h *Handler) CodeAction(params CodeActionParams) []CodeAction {
	doc, ok := h.server.Documents().Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	diagResult := h.server.Diagnostic().Provide(params.TextDocument.URI, doc.Text)
	actions := h.server.CodeAction().Provide(params.TextDocument.URI, diagResult.Diagnostics)

	return actions
}

func (h *Handler) SignatureHelp(params SignatureHelpParams) *SignatureHelp {
	doc, ok := h.server.Documents().Get(params.TextDocument.URI)
	if !ok {
		return nil
	}
	return h.server.Signature().Provide(doc.Text, params.Position.Line, params.Position.Character)
}

func (h *Handler) Formatting(params DocumentFormattingParams) ([]TextEdit, error) {
	doc, ok := h.server.Documents().Get(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	return h.server.Format().Format(doc.Text)
}

func (h *Handler) RangeFormatting(params DocumentRangeFormattingParams) ([]TextEdit, error) {
	doc, ok := h.server.Documents().Get(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	return h.server.Format().FormatRange(doc.Text, params.Range)
}

func (h *Handler) publishDiagnostics(uri, text string) {
	params := h.server.Diagnostic().Provide(uri, text)
	h.server.SendNotification(MethodPublishDiag, params)
}

func (h *Handler) wordAtPosition(text string, line, character int) string {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}
	lineText := lines[line]
	if character < 0 || character > len(lineText) {
		return ""
	}

	start := character
	for start > 0 && (unicode.IsLetter(rune(lineText[start-1])) || lineText[start-1] == '-' || lineText[start-1] == '_') {
		start--
	}
	end := character
	for end < len(lineText) && (unicode.IsLetter(rune(lineText[end])) || lineText[end] == '-' || lineText[end] == '_') {
		end++
	}

	return lineText[start:end]
}

func (h *Handler) lookupDocumentation(word string) string {
	docs := map[string]string{
		"project":        "**project** (required)\nThe name of your project. Must be lowercase alphanumeric with hyphens.\n\nExample: `project: e-commerce-platform`",
		"version":        "**version**\nProject version in semver format.\n\nExample: `version: 1.0.0`",
		"description":    "**description**\nA short description of the project.",
		"license":        "**license**\nProject license identifier (e.g. MIT, Apache-2.0).",
		"repository":     "**repository**\nURL to the project source repository.",
		"tags":           "**tags**\nList of keyword tags for the project.",
		"modules":        "**modules**\nList of code modules in the project.\n\nEach module has: `name`, `path`, `description`, `dependencies`",
		"services":       "**services**\nList of runnable services.\n\nEach service has: `name`, `kind`, `port`, `description`, `endpoints`",
		"architecture":   "**architecture**\nArchitecture pattern and principles.\n\nPatterns: `layered`, `clean`, `hexagonal`, `microkernel`, `event-driven`, `cqrs`, `monolith`",
		"deployment":     "**deployment**\nDeployment configuration.\n\nStrategies: `rolling`, `blue-green`, `canary`, `recreate`",
		"testing":        "**testing**\nTesting strategy and coverage targets.\n\nStrategies: `unit`, `integration`, `e2e`, `contract`",
		"generation":     "**generation**\nCode generation configuration.\n\nLanguages: `go`, `typescript`, `python`, `java`, `rust`",
		"kind":           "**kind**\nService kind.\n\nValues: `http`, `grpc`, `worker`, `cli`, `job`",
		"port":           "**port**\nService port number (1-65535).",
		"endpoints":      "**endpoints**\nList of API endpoints.\n\nEach endpoint has: `method`, `path`, `action`",
		"method":         "**method**\nHTTP method.\n\nValues: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`",
		"path":           "**path**\nURL path for the endpoint.\n\nExample: `path: /auth/login`",
		"action":         "**action**\nHandler function name.\n\nExample: `action: login`",
		"name":           "**name**\nIdentifier name.",
		"pattern":        "**pattern**\nArchitecture pattern.\n\nValues: `layered`, `clean`, `hexagonal`, `microkernel`, `event-driven`, `cqrs`, `monolith`",
		"principles":     "**principles**\nList of architecture principles (e.g. DI, SRP, OCP).",
		"strategy":       "**strategy**\nDeployment or testing strategy.",
		"languages":      "**languages**\nTarget programming languages.\n\nValues: `go`, `typescript`, `python`, `java`, `rust`",
		"coverage":       "**coverage**\nCode coverage target.\n\nExample: `coverage: 85%`",
		"dependencies":   "**dependencies**\nList of module dependencies.",
		"domain":         "**domain**\nDomain-driven design configuration.\n\nSections:\n- `events` — Domain events\n- `bounded_contexts` — Bounded contexts\n- `ubiquitous_language` — Glossary of domain terms\n- `aggregates` — Domain aggregates\n- `value_objects` — Value objects\n\nExample:\n```yaml\ndomain:\n  bounded_contexts:\n    - name: billing\n      aggregates:\n        - Invoice\n```",
		"security":       "**security**\nSecurity configuration.\n\nSections:\n- `authentication` — Auth config (OAuth2, OIDC, SAML, LDAP)\n- `authorization` — RBAC/ABAC policies\n- `encryption` — Encryption at rest and in transit\n- `cors` — CORS settings\n- `rate_limiting` — Rate limiting\n\nExample:\n```yaml\nsecurity:\n  authentication:\n    provider: oidc\n  cors:\n    origins:\n      - https://app.example.com\n```",
		"infrastructure": "**infrastructure**\nCloud infrastructure configuration.\n\nSections:\n- `provider` — Cloud provider (aws, gcp, azure, local)\n- `region` — Deployment region\n- `resources` — Infrastructure resources\n- `kubernetes` — K8s cluster settings\n- `networking` — Network config (VPC, subnets)\n\nExample:\n```yaml\ninfrastructure:\n  provider: aws\n  region: us-east-1\n  kubernetes:\n    version: \"1.28\"\n```",
		"storage":        "**storage**\nData store configuration.\n\nSections:\n- `type` — Storage type (postgres, mysql, redis, s3, mongodb)\n- `connection` — Connection string\n- `migrations` — Database migrations\n- `backup` — Backup configuration\n- `pooling` — Connection pooling\n\nExample:\n```yaml\nstorage:\n  - name: primary\n    type: postgres\n    connection: postgresql://localhost:5432/db\n```",
		"components":     "**components**\nInternal component definitions.\n\nSections:\n- `type` — Component type (service, worker, cron, event-handler)\n- `name` — Component name\n- `path` — Component path\n- `dependencies` — Component dependencies\n\nExample:\n```yaml\ncomponents:\n  - name: auth-service\n    type: service\n    path: ./internal/auth\n```",
		"api":            "**api**\nAPI definition.\n\nSections:\n- `version` — API version\n- `endpoints` — API endpoints\n- `format` — API format (rest, graphql, grpc)\n- `documentation` — API documentation path\n\nExample:\n```yaml\napi:\n  version: v1\n  format: rest\n  endpoints:\n    - path: /users\n      method: GET\n```",
		"database":       "**database**\nDatabase configuration.\n\nSee `storage` for data store definitions.\n\nExample:\n```yaml\ndatabase:\n  type: postgres\n  migrations:\n    - name: init\n      path: ./migrations/001_init.sql\n```",
		"monitoring":     "**monitoring**\nMonitoring and observability configuration.\n\nSections:\n- `metrics` — Metrics collection (prometheus, datadog)\n- `logging` — Logging configuration\n- `tracing` — Distributed tracing\n\nExample:\n```yaml\nmonitoring:\n  metrics:\n    provider: prometheus\n  logging:\n    level: info\n```",
		"logging":        "**logging**\nLogging configuration.\n\nSections:\n- `level` — Log level (debug, info, warn, error)\n- `format` — Log format (json, text)\n- `output` — Log output destination\n\nExample:\n```yaml\nlogging:\n  level: info\n  format: json\n```",
		"environments":   "**environments**\nList of deployment environments (e.g. dev, staging, production).",
		"output_dir":     "**output_dir**\nOutput directory for generated code.",
		"provider":       "**provider**\nCloud provider.\n\nValues: `aws`, `gcp`, `azure`, `local`",
		"region":         "**region**\nCloud region identifier.",
		"protocol":       "**protocol**\nAPI protocol.\n\nValues: `http`, `grpc`, `graphql`, `websocket`",
	}

	if info, ok := docs[word]; ok {
		return info
	}
	return ""
}

func (h *Handler) buildSymbols(text string) []DocumentSymbol {
	lines := strings.Split(text, "\n")
	var symbols []DocumentSymbol

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent > 0 {
			continue
		}

		colonIdx := strings.Index(trimmed, ":")
		if colonIdx <= 0 {
			continue
		}

		key := trimmed[:colonIdx]
		var detail string
		rest := strings.TrimSpace(trimmed[colonIdx+1:])
		if rest != "" && rest[0] != '\n' {
			detail = rest
		}

		kind := SymbolProperty
		switch key {
		case "project":
			kind = SymbolConstant
		case "modules":
			kind = SymbolModule
		case "services":
			kind = SymbolClass
		case "architecture", "deployment", "testing", "generation", "security", "infrastructure":
			kind = SymbolNamespace
		}

		symbols = append(symbols, DocumentSymbol{
			Name:           key,
			Detail:         detail,
			Kind:           kind,
			Range:          Range{Start: Position{Line: i, Character: 0}, End: Position{Line: i, Character: len(line)}},
			SelectionRange: Range{Start: Position{Line: i, Character: 0}, End: Position{Line: i, Character: len(key)}},
			Children:       h.buildChildSymbols(lines, i+1, 2),
		})
	}

	return symbols
}

func (h *Handler) buildChildSymbols(lines []string, startIdx, minIndent int) []DocumentSymbol {
	var symbols []DocumentSymbol

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent < minIndent {
			break
		}

		if indent != minIndent {
			continue
		}

		colonIdx := strings.Index(trimmed, ":")
		if colonIdx <= 0 {
			continue
		}

		key := trimmed[:colonIdx]
		rest := strings.TrimSpace(trimmed[colonIdx+1:])

		if strings.HasPrefix(trimmed, "- ") {
			key = strings.TrimPrefix(trimmed, "- ")
			colonIdx = strings.Index(key, ":")
			if colonIdx > 0 {
				key = key[:colonIdx]
			}
		}

		var detail string
		if rest != "" && rest[0] != '\n' {
			detail = rest
		}

		symbols = append(symbols, DocumentSymbol{
			Name:           key,
			Detail:         detail,
			Kind:           SymbolField,
			Range:          Range{Start: Position{Line: i, Character: 0}, End: Position{Line: i, Character: len(line)}},
			SelectionRange: Range{Start: Position{Line: i, Character: indent}, End: Position{Line: i, Character: indent + len(key)}},
			Children:       h.buildChildSymbols(lines, i+1, minIndent+2),
		})
	}

	return symbols
}
