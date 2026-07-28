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
			TextDocumentSync:       1,
			CompletionProvider:     &CompletionOptions{TriggerCharacters: []string{":", " ", "-"}},
			HoverProvider:          true,
			DefinitionProvider:     true,
			DocumentSymbolProvider: true,
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
		"project":      "**project** (required)\nThe name of your project. Must be lowercase alphanumeric with hyphens.\n\nExample: `project: e-commerce-platform`",
		"version":      "**version**\nProject version in semver format.\n\nExample: `version: 1.0.0`",
		"description":  "**description**\nA short description of the project.",
		"modules":      "**modules**\nList of code modules in the project.\n\nEach module has: `name`, `path`, `description`, `dependencies`",
		"services":     "**services**\nList of runnable services.\n\nEach service has: `name`, `kind`, `port`, `description`, `endpoints`",
		"architecture": "**architecture**\nArchitecture pattern and principles.\n\nPatterns: `layered`, `clean`, `hexagonal`, `microkernel`, `event-driven`, `cqrs`, `monolith`",
		"deployment":   "**deployment**\nDeployment configuration.\n\nStrategies: `rolling`, `blue-green`, `canary`, `recreate`",
		"testing":      "**testing**\nTesting strategy and coverage targets.\n\nStrategies: `unit`, `integration`, `e2e`, `contract`",
		"generation":   "**generation**\nCode generation configuration.\n\nLanguages: `go`, `typescript`, `python`, `java`, `rust`",
		"kind":         "**kind**\nService kind.\n\nValues: `http`, `grpc`, `worker`, `cli`, `job`",
		"port":         "**port**\nService port number (1-65535).",
		"endpoints":    "**endpoints**\nList of API endpoints.\n\nEach endpoint has: `method`, `path`, `action`",
		"method":       "**method**\nHTTP method.\n\nValues: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`",
		"path":         "**path**\nURL path for the endpoint.\n\nExample: `path: /auth/login`",
		"action":       "**action**\nHandler function name.\n\nExample: `action: login`",
		"name":         "**name**\nIdentifier name.",
		"pattern":      "**pattern**\nArchitecture pattern.\n\nValues: `layered`, `clean`, `hexagonal`, `microkernel`, `event-driven`, `cqrs`, `monolith`",
		"strategy":     "**strategy**\nDeployment or testing strategy.",
		"languages":    "**languages**\nTarget programming languages.\n\nValues: `go`, `typescript`, `python`, `java`, `rust`",
		"coverage":     "**coverage**\nCode coverage target.\n\nExample: `coverage: 85%`",
		"dependencies": "**dependencies**\nList of module dependencies.",
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
