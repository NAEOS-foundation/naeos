package lsp

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	if s.documents == nil {
		t.Error("expected documents manager")
	}
	if s.completion == nil {
		t.Error("expected completion provider")
	}
	if s.diagnostic == nil {
		t.Error("expected diagnostic provider")
	}
	if s.handler == nil {
		t.Error("expected handler")
	}
}

func TestReadMessage(t *testing.T) {
	bodyStr := `{"jsonrpc":"2.0","id":1,"method":"test"}`
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(bodyStr), bodyStr)
	r := strings.NewReader(input)
	br := bufio.NewReader(r)
	body, err := readMessage(br)
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}
	if string(body) != bodyStr {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestReadMessageMultiple(t *testing.T) {
	first := `{"method":"first"}`
	second := `{"method":"second"}`
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%sContent-Length: %d\r\n\r\n%s", len(first), first, len(second), second)
	r := strings.NewReader(input)
	br := bufio.NewReader(r)

	body, err := readMessage(br)
	if err != nil {
		t.Fatalf("first read: %v", err)
	}
	if string(body) != first {
		t.Errorf("unexpected first body: %s", body)
	}

	body, err = readMessage(br)
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	if string(body) != second {
		t.Errorf("unexpected second body: %s", body)
	}
}

func TestCompletionProviderInit(t *testing.T) {
	p := NewCompletionProvider()
	if len(p.suggestions) == 0 {
		t.Error("expected non-empty suggestions")
	}
	if len(p.keywords) == 0 {
		t.Error("expected non-empty keywords")
	}
}

func TestCompletionProvide(t *testing.T) {
	p := NewCompletionProvider()
	text := "project: myapp\nservices:\n  - name: api\n    kind: http"

	list := p.Provide(text, 0, 0)
	if list == nil {
		t.Fatal("expected non-nil completion list")
	}
	if len(list.Items) == 0 {
		t.Error("expected non-empty items at top level")
	}
}

func TestCompletionFilterByPrefix(t *testing.T) {
	p := NewCompletionProvider()
	text := "pro"

	list := p.Provide(text, 0, 3)
	if list == nil {
		t.Fatal("expected non-nil completion list")
	}
	for _, item := range list.Items {
		if !strings.HasPrefix(item.Label, "pro") {
			t.Errorf("expected all items to start with 'pro', got %q", item.Label)
		}
	}
}

func TestCompletionKindValues(t *testing.T) {
	p := NewCompletionProvider()
	text := "    kind: "

	list := p.Provide(text, 0, len(text))
	if list == nil {
		t.Fatal("expected non-nil completion list")
	}
	if len(list.Items) == 0 {
		t.Fatal("expected kind value completions")
	}
	expected := []string{"http", "grpc", "worker", "cli", "job"}
	for _, exp := range expected {
		var found bool
		for _, item := range list.Items {
			if item.Label == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected kind value %q not found", exp)
		}
	}
}

func TestCompletionMethodValues(t *testing.T) {
	p := NewCompletionProvider()
	text := "      method: "

	list := p.Provide(text, 0, len(text))
	if list == nil {
		t.Fatal("expected non-nil completion list")
	}
	expected := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, exp := range expected {
		var found bool
		for _, item := range list.Items {
			if item.Label == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected method value %q not found", exp)
		}
	}
}

func TestCompletionPatternValues(t *testing.T) {
	p := NewCompletionProvider()
	text := "  pattern: "

	list := p.Provide(text, 0, len(text))
	if list == nil {
		t.Fatal("expected non-nil completion list")
	}
	if len(list.Items) == 0 {
		t.Fatal("expected pattern value completions")
	}
}

func TestCompletionSnippets(t *testing.T) {
	p := NewCompletionProvider()
	text := ""

	list := p.Provide(text, 0, 0)
	if list == nil {
		t.Fatal("expected non-nil completion list")
	}

	var foundIf bool
	for _, item := range list.Items {
		if item.Label == "$if{condition}" {
			foundIf = true
			if item.InsertText != "$if{ $1 }:\n  $0\n$endif" {
				t.Errorf("unexpected insert text for $if: %q", item.InsertText)
			}
			break
		}
	}
	if !foundIf {
		t.Error("expected $if{condition} snippet to be in suggestions")
	}
}

func TestCompletionContextService(t *testing.T) {
	p := NewCompletionProvider()
	text := "services:\n  - name: api\n    "

	list := p.Provide(text, 2, 4)
	if list == nil {
		t.Fatal("expected non-nil completion list")
	}
	expected := []string{"name", "kind", "port", "description", "endpoints"}
	for _, exp := range expected {
		var found bool
		for _, item := range list.Items {
			if item.Label == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected service field %q not found", exp)
		}
	}
}

func TestCompletionContextModule(t *testing.T) {
	p := NewCompletionProvider()
	text := "modules:\n  - name: auth\n    "

	list := p.Provide(text, 2, 4)
	expected := []string{"name", "path", "description", "dependencies"}
	for _, exp := range expected {
		var found bool
		for _, item := range list.Items {
			if item.Label == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected module field %q not found", exp)
		}
	}
}

func TestCompletionNoResultsAfterValue(t *testing.T) {
	p := NewCompletionProvider()
	text := "project: myapp\nservices:\n  - name: api\n    kind: http\n"

	list := p.Provide(text, 2, 13)
	if list == nil {
		t.Fatal("expected non-nil completion list")
	}
}

func TestDiagnosticProviderEmpty(t *testing.T) {
	dp := NewDiagnosticProvider()
	result := dp.Provide("test.yaml", "")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestDiagnosticValidSpec(t *testing.T) {
	dp := NewDiagnosticProvider()
	spec := `project: myapp
version: "1.0"
services:
  - name: api
    kind: http
    port: 8080
modules:
  - name: core
    path: ./core
`
	result := dp.Provide("test.yaml", spec)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	for _, d := range result.Diagnostics {
		if d.Severity == DiagError {
			t.Errorf("unexpected error diagnostic: %s", d.Message)
		}
	}
}

func TestDiagnosticMissingProject(t *testing.T) {
	dp := NewDiagnosticProvider()
	spec := `version: "1.0"`
	result := dp.Provide("test.yaml", spec)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	var found bool
	for _, d := range result.Diagnostics {
		if strings.Contains(d.Message, "project") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected diagnostic about missing project")
	}
}

func TestDiagnosticTabs(t *testing.T) {
	dp := NewDiagnosticProvider()
	spec := "project: myapp\n\tservices:"
	result := dp.Provide("test.yaml", spec)
	var found bool
	for _, d := range result.Diagnostics {
		if strings.Contains(d.Message, "Tabs") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected diagnostic about tabs")
	}
}

func TestDiagnosticInvalidPort(t *testing.T) {
	dp := NewDiagnosticProvider()
	spec := `project: myapp
services:
  - name: api
    port: 99999
`
	result := dp.Provide("test.yaml", spec)
	var found bool
	for _, d := range result.Diagnostics {
		if strings.Contains(d.Message, "Port") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected diagnostic about invalid port")
	}
}

func TestDiagnosticEmptyProject(t *testing.T) {
	dp := NewDiagnosticProvider()
	spec := `project: ""
version: "1.0"
services:
  - name: api
    kind: http
`
	result := dp.Provide("test.yaml", spec)
	var found bool
	for _, d := range result.Diagnostics {
		if d.Severity == DiagWarning && strings.Contains(d.Message, "empty") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected diagnostic about empty project value")
	}
}

func TestDiagnosticLineRange(t *testing.T) {
	spec := "project: myapp\nservices:\n  - name: api\n    port: 99999\n"
	r := lineRange(spec, "port: 99999")
	if r.Start.Line != 3 {
		t.Errorf("expected line 3, got %d", r.Start.Line)
	}
}

func TestHandlerInitialize(t *testing.T) {
	s := NewServer()
	h := s.handler
	result := h.Initialize(InitializeParams{})
	if !result.Capabilities.HoverProvider {
		t.Error("expected HoverProvider true")
	}
	if !result.Capabilities.DefinitionProvider {
		t.Error("expected DefinitionProvider true")
	}
	if !result.Capabilities.DocumentSymbolProvider {
		t.Error("expected DocumentSymbolProvider true")
	}
	if result.Capabilities.CompletionProvider == nil {
		t.Error("expected CompletionProvider")
	}
}

func TestHandlerDidOpen(t *testing.T) {
	s := NewServer()
	h := s.handler

	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "test.yaml",
			Text: "project: myapp\nservices:\n  - name: api\n    kind: http\n",
		},
	})

	doc, ok := s.documents.Get("test.yaml")
	if !ok {
		t.Fatal("expected document to be stored")
	}
	if doc.Text != "project: myapp\nservices:\n  - name: api\n    kind: http\n" {
		t.Errorf("unexpected document text: %s", doc.Text)
	}
}

func TestHandlerDidChange(t *testing.T) {
	s := NewServer()
	h := s.handler

	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "test.yaml",
			Text: "project: myapp",
		},
	})

	h.DidChange(DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			URI:     "test.yaml",
			Version: 2,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: "project: myapp\nservices:\n  - name: api"},
		},
	})

	doc, ok := s.documents.Get("test.yaml")
	if !ok {
		t.Fatal("expected document to exist")
	}
	if doc.Version != 2 {
		t.Errorf("expected version 2, got %d", doc.Version)
	}
	if doc.Text != "project: myapp\nservices:\n  - name: api" {
		t.Errorf("unexpected text: %s", doc.Text)
	}
}

func TestHandlerDidClose(t *testing.T) {
	s := NewServer()
	h := s.handler

	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "test.yaml",
			Text: "project: myapp",
		},
	})

	h.DidClose(DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
	})

	_, ok := s.documents.Get("test.yaml")
	if ok {
		t.Error("expected document to be removed")
	}
}

func TestHandlerHover(t *testing.T) {
	s := NewServer()
	h := s.handler

	spec := "project: myapp\nservices:\n  - name: api\n    kind: http\n"
	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: spec},
	})

	hover := h.Hover(HoverParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
		Position:     Position{Line: 1, Character: 2},
	})
	if hover == nil {
		t.Fatal("expected non-nil hover result")
	}
}

func TestHandlerDefinition(t *testing.T) {
	s := NewServer()
	h := s.handler

	spec := "project: myapp\napi:\n  name: user-service\nservices:\n  - name: api\n    kind: http\n"
	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: spec},
	})

	// "name" at line 4, char 7 (inside "name" value) → should find line 2 "name:" key
	loc := h.Definition(DefinitionParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
		Position:     Position{Line: 4, Character: 7},
	})
	if loc == nil {
		t.Fatal("expected non-nil definition result")
	}
}

func TestHandlerDocumentSymbol(t *testing.T) {
	s := NewServer()
	h := s.handler

	spec := "project: myapp\nservices:\n  - name: api\n    kind: http\n  - name: worker\n    kind: job\n"
	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: spec},
	})

	symbols := h.DocumentSymbol(DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
	})
	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol")
	}
	var foundProject bool
	for _, s := range symbols {
		if s.Name == "project" {
			foundProject = true
			break
		}
	}
	if !foundProject {
		t.Error("expected 'project' symbol")
	}
}

func TestDocumentManager(t *testing.T) {
	dm := NewDocumentManager()

	dm.Open("test.yaml", 1, "content1")
	doc, ok := dm.Get("test.yaml")
	if !ok {
		t.Fatal("expected document")
	}
	if doc.Text != "content1" {
		t.Errorf("unexpected text: %s", doc.Text)
	}
	if doc.Version != 1 {
		t.Errorf("expected version 1, got %d", doc.Version)
	}

	dm.Update("test.yaml", 2, "content2")
	doc, ok = dm.Get("test.yaml")
	if !ok {
		t.Fatal("expected document after update")
	}
	if doc.Text != "content2" {
		t.Errorf("unexpected text after update: %s", doc.Text)
	}

	dm.Close("test.yaml")
	_, ok = dm.Get("test.yaml")
	if ok {
		t.Error("expected nil after close")
	}
}

func TestDocumentManagerAll(t *testing.T) {
	dm := NewDocumentManager()
	dm.Open("a.yaml", 1, "a")
	dm.Open("b.yaml", 1, "b")

	all := dm.All()
	if len(all) != 2 {
		t.Errorf("expected 2 documents, got %d", len(all))
	}
}

func TestDocumentManagerGetNotFound(t *testing.T) {
	dm := NewDocumentManager()
	_, ok := dm.Get("nonexistent.yaml")
	if ok {
		t.Error("expected nil for nonexistent document")
	}
}

func TestWordAtPosition(t *testing.T) {
	s := NewServer()
	h := s.handler
	text := "project: myapp\nservices:\n  - name: api\n    kind: http\n"
	word := h.wordAtPosition(text, 2, 7)
	if word != "name" {
		t.Errorf("expected 'name', got %q", word)
	}

	word = h.wordAtPosition(text, 0, 4)
	if word != "project" {
		t.Errorf("expected 'project', got %q", word)
	}

	word = h.wordAtPosition(text, 2, 11)
	if word != "api" {
		t.Errorf("expected 'api', got %q", word)
	}
}

func TestLookupDocumentation(t *testing.T) {
	s := NewServer()
	h := s.handler
	doc := h.lookupDocumentation("services")
	if doc == "" {
		t.Error("expected documentation for 'services'")
	}

	doc = h.lookupDocumentation("nonexistent_keyword")
	if doc != "" {
		t.Errorf("expected empty for nonexistent, got %q", doc)
	}
}

func TestBuildSymbols(t *testing.T) {
	s := NewServer()
	h := s.handler
	text := "project: myapp\nservices:\n  - name: api\n    kind: http\n  - name: worker\n"
	symbols := h.buildSymbols(text)
	if len(symbols) == 0 {
		t.Fatal("expected symbols")
	}
}

func TestServerSendResponse(t *testing.T) {
	s := NewServer()
	s.sendResponse(1, "ok")
}

func TestServerSendNotification(t *testing.T) {
	s := NewServer()
	s.SendNotification(MethodPublishDiag, PublishDiagnosticsParams{
		URI: "test.yaml",
		Diagnostics: []Diagnostic{
			{Range: Range{Start: Position{}, End: Position{}}, Severity: DiagError, Message: "test"},
		},
	})
}

func TestHandleInitialize(t *testing.T) {
	s := NewServer()
	s.handleMessage(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
}

func TestHandleCompletion(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "project: myapp"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":2,"method":"textDocument/completion","params":{"textDocument":{"uri":"test.yaml"},"position":{"line":0,"character":0}}}`)
}

func TestHandleUnknownMethod(t *testing.T) {
	s := NewServer()
	s.handleMessage(`{"jsonrpc":"2.0","id":3,"method":"unknown/method"}`)
}

func TestCodeActionProvider(t *testing.T) {
	cap := NewCodeActionProvider()

	diags := []Diagnostic{
		{Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 1}}, Severity: DiagError, Message: "Required field 'project' is missing"},
	}
	actions := cap.Provide("test.yaml", diags)
	if len(actions) == 0 {
		t.Fatal("expected at least one code action")
	}

	var found bool
	for _, a := range actions {
		if a.Title == "Add missing 'project' field" {
			found = true
			if a.Kind != CodeActionQuickFix {
				t.Errorf("expected quickfix kind, got %s", a.Kind)
			}
			if a.Edit == nil {
				t.Fatal("expected edit")
			}
			edits, ok := a.Edit.Changes["test.yaml"]
			if !ok || len(edits) == 0 {
				t.Fatal("expected text edits for test.yaml")
			}
			if edits[0].NewText != "project: my-project\n" {
				t.Errorf("expected 'project: my-project\\n', got %q", edits[0].NewText)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'Add missing project field' code action")
	}
}

func TestCodeActionTrailingWhitespace(t *testing.T) {
	cap := NewCodeActionProvider()

	diags := []Diagnostic{
		{Range: Range{}, Severity: DiagHint, Message: "Trailing whitespace"},
	}
	actions := cap.Provide("test.yaml", diags)
	var found bool
	for _, a := range actions {
		if a.Title == "Fix trailing whitespace" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Fix trailing whitespace' code action")
	}

	fixed := cap.FixTrailingWhitespace("project: myapp  \nservices:\n  - name: api  \n")
	expected := "project: myapp\nservices:\n  - name: api"
	if fixed != expected {
		t.Errorf("expected %q, got %q", expected, fixed)
	}

	tabFixed := cap.FixTabs("project: myapp\n\tservices:")
	if tabFixed != "project: myapp\n  services:" {
		t.Errorf("expected tabs converted to spaces, got %q", tabFixed)
	}
}

func TestSignatureProvider(t *testing.T) {
	sp := NewSignatureProvider()
	text := "$"
	help := sp.Provide(text, 0, 1)
	if help == nil {
		t.Fatal("expected non-nil signature help")
	}
	if len(help.Signatures) == 0 {
		t.Fatal("expected at least one signature")
	}
}

func TestSignatureProviderIf(t *testing.T) {
	sp := NewSignatureProvider()
	text := "$if{"
	help := sp.Provide(text, 0, len(text))
	if help == nil {
		t.Fatal("expected non-nil signature help for $if{")
	}
	var found bool
	for _, s := range help.Signatures {
		if s.Label == "$if{condition: block}" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected '$if{condition: block}' signature")
	}
}

func TestSignatureProviderFn(t *testing.T) {
	sp := NewSignatureProvider()
	text := "$fn{upper("
	help := sp.Provide(text, 0, len(text))
	if help == nil {
		t.Fatal("expected non-nil signature help for $fn{upper(")
	}
	var found bool
	for _, s := range help.Signatures {
		if s.Label == "$fn{upper(value: string)}" {
			found = true
			if len(s.Parameters) != 1 {
				t.Errorf("expected 1 parameter, got %d", len(s.Parameters))
			}
			break
		}
	}
	if !found {
		t.Error("expected '$fn{upper(value: string)}' signature")
	}
}

func TestSignatureProviderDefault(t *testing.T) {
	sp := NewSignatureProvider()
	text := "$fn{default("
	help := sp.Provide(text, 0, len(text))
	if help == nil {
		t.Fatal("expected non-nil signature help for $fn{default(")
	}
	var found bool
	for _, s := range help.Signatures {
		if s.Label == "$fn{default(value: any, fallback: any)}" {
			found = true
			if len(s.Parameters) != 2 {
				t.Errorf("expected 2 parameters, got %d", len(s.Parameters))
			}
			break
		}
	}
	if !found {
		t.Error("expected '$fn{default(value: any, fallback: any)}' signature")
	}
}

func TestSignatureProviderEmptyLine(t *testing.T) {
	sp := NewSignatureProvider()
	help := sp.Provide("", 0, 0)
	if help == nil {
		t.Fatal("expected non-nil signature help for empty text")
	}
}

func TestFormatProvider(t *testing.T) {
	fp := NewFormatProvider()
	text := "services:\n  - name: api\nproject: myapp\n"
	edits, err := fp.Format(text)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if len(edits) == 0 {
		t.Fatal("expected at least one text edit")
	}
	formatted := edits[0].NewText
	lines := strings.Split(formatted, "\n")
	if len(lines) >= 2 {
		if !strings.HasPrefix(lines[0], "project") {
			t.Errorf("expected 'project' block first after sorting, got %q", lines[0])
		}
	}
}

func TestFormatProviderTrailingWhitespace(t *testing.T) {
	fp := NewFormatProvider()
	text := "project: myapp  \nversion: \"1.0\"  \n"
	edits, err := fp.Format(text)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	formatted := edits[0].NewText
	if strings.Contains(formatted, "  \n") {
		t.Error("expected no trailing whitespace")
	}
}

func TestFormatProviderTabs(t *testing.T) {
	fp := NewFormatProvider()
	text := "project: myapp\n\tservices:\n"
	edits, err := fp.Format(text)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	formatted := edits[0].NewText
	if strings.Contains(formatted, "\t") {
		t.Error("expected no tabs after formatting")
	}
}

func TestFormatProviderSortOrder(t *testing.T) {
	fp := NewFormatProvider()
	text := "services:\n  - name: api\nversion: \"1.0\"\nproject: myapp\n"
	edits, err := fp.Format(text)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	formatted := edits[0].NewText
	if !strings.HasPrefix(formatted, "project") {
		t.Errorf("expected 'project' first, got %q", formatted)
	}
}

func TestFormatRange(t *testing.T) {
	fp := NewFormatProvider()
	text := "project: myapp  \nservices:\n  - name: api\n"
	r := Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 1, Character: 0}}
	edits, err := fp.FormatRange(text, r)
	if err != nil {
		t.Fatalf("FormatRange: %v", err)
	}
	if len(edits) == 0 {
		t.Fatal("expected edits from FormatRange")
	}
}

func TestServerCodeActionAccessor(t *testing.T) {
	s := NewServer()
	if s.CodeAction() == nil {
		t.Error("expected non-nil CodeAction provider")
	}
	if s.Signature() == nil {
		t.Error("expected non-nil Signature provider")
	}
	if s.Format() == nil {
		t.Error("expected non-nil Format provider")
	}
}

func TestHandlerCodeAction(t *testing.T) {
	s := NewServer()
	h := s.handler

	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "test.yaml",
			Text: "services:\n  - name: api\n",
		},
	})

	actions := h.CodeAction(CodeActionParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
	})
	if len(actions) == 0 {
		t.Fatal("expected code actions for missing project")
	}
}

func TestHandlerSignatureHelp(t *testing.T) {
	s := NewServer()
	h := s.handler

	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "test.yaml",
			Text: "$",
		},
	})

	help := h.SignatureHelp(SignatureHelpParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
		Position:     Position{Line: 0, Character: 1},
	})
	if help == nil {
		t.Fatal("expected signature help")
	}
	if len(help.Signatures) == 0 {
		t.Error("expected at least one signature")
	}
}

func TestHandlerFormatting(t *testing.T) {
	s := NewServer()
	h := s.handler

	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "test.yaml",
			Text: "services:\n  - name: api\nproject: myapp\n",
		},
	})

	edits, err := h.Formatting(DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
	})
	if err != nil {
		t.Fatalf("Formatting: %v", err)
	}
	if len(edits) == 0 {
		t.Fatal("expected formatting edits")
	}
}

func TestHandlerRangeFormatting(t *testing.T) {
	s := NewServer()
	h := s.handler

	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "test.yaml",
			Text: "project: myapp  \nservices:\n  - name: api\n",
		},
	})

	edits, err := h.RangeFormatting(DocumentRangeFormattingParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
		Range:        Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 1, Character: 0}},
	})
	if err != nil {
		t.Fatalf("RangeFormatting: %v", err)
	}
	if len(edits) == 0 {
		t.Fatal("expected range formatting edits")
	}
}

func TestHandleCodeAction(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "services:\n  - name: api\n"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":1,"method":"textDocument/codeAction","params":{"textDocument":{"uri":"test.yaml"},"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":0}},"context":{"diagnostics":[]}}}`)
}

func TestHandleSignatureHelp(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "$"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":1,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"test.yaml"},"position":{"line":0,"character":1}}}`)
}

func TestHandleFormatting(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "services:\n  - name: api\nproject: myapp\n"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":1,"method":"textDocument/formatting","params":{"textDocument":{"uri":"test.yaml"},"options":{"tabSize":2,"insertSpaces":true}}}`)
}

func TestHandleRangeFormatting(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "project: myapp\nservices:\n  - name: api\n"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":1,"method":"textDocument/rangeFormatting","params":{"textDocument":{"uri":"test.yaml"},"range":{"start":{"line":0,"character":0},"end":{"line":1,"character":0}},"options":{"tabSize":2,"insertSpaces":true}}}`)
}

func TestHandlerCodeActionNoDocument(t *testing.T) {
	s := NewServer()
	h := s.handler
	actions := h.CodeAction(CodeActionParams{
		TextDocument: TextDocumentIdentifier{URI: "nonexistent.yaml"},
	})
	if actions != nil {
		t.Error("expected nil for nonexistent document")
	}
}

func TestHandlerHoverNoDocument(t *testing.T) {
	s := NewServer()
	h := s.handler
	hover := h.Hover(HoverParams{
		TextDocument: TextDocumentIdentifier{URI: "nonexistent.yaml"},
		Position:     Position{Line: 0, Character: 0},
	})
	if hover != nil {
		t.Error("expected nil hover for nonexistent document")
	}
}

func TestHandlerHoverUnknownWord(t *testing.T) {
	s := NewServer()
	h := s.handler
	spec := "project: myapp\n"
	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: spec},
	})

	hover := h.Hover(HoverParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
		Position:     Position{Line: 0, Character: 50},
	})
	if hover != nil {
		t.Error("expected nil hover for position beyond line length")
	}
}

func TestHandlerHoverExistingWord(t *testing.T) {
	s := NewServer()
	h := s.handler
	spec := "project: myapp\nservices:\n  - name: api\n    kind: http\n"
	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: spec},
	})

	hover := h.Hover(HoverParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
		Position:     Position{Line: 1, Character: 2},
	})
	if hover == nil {
		t.Fatal("expected hover result for 'services'")
	}
	if hover.Contents.Value == "" {
		t.Error("expected non-empty hover content")
	}
}

func TestHandlerDefinitionNotFound(t *testing.T) {
	s := NewServer()
	h := s.handler
	spec := "project: myapp\n"
	h.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: spec},
	})

	loc := h.Definition(DefinitionParams{
		TextDocument: TextDocumentIdentifier{URI: "test.yaml"},
		Position:     Position{Line: 99, Character: 0},
	})
	if loc != nil {
		t.Error("expected nil for out-of-range position")
	}
}

func TestHandleHoverMessage(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "project: myapp\nservices:\n  - name: api\n"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":4,"method":"textDocument/hover","params":{"textDocument":{"uri":"test.yaml"},"position":{"line":1,"character":2}}}`)
}

func TestHandleDefinitionMessage(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "project: myapp\nservices:\n  - name: api\n"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":5,"method":"textDocument/definition","params":{"textDocument":{"uri":"test.yaml"},"position":{"line":1,"character":2}}}`)
}

func TestHandleCodeActionMessage(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "project: myapp"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":6,"method":"textDocument/codeAction","params":{"textDocument":{"uri":"test.yaml"},"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":7}},"context":{"diagnostics":[{"code":"missing_project","range":{"start":{"line":0,"character":0},"end":{"line":0,"character":0}}}]}}}`)
}

func TestHandleDidChangeMessage(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "project: myapp"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"test.yaml","version":2},"contentChanges":[{"text":"project: myapp\nservices:\n  - name: api\n"}]}}`)
}

func TestHandleDidCloseMessage(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "project: myapp"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","method":"textDocument/didClose","params":{"textDocument":{"uri":"test.yaml"}}}`)
	_, ok := s.documents.Get("test.yaml")
	if ok {
		t.Error("expected document to be removed after close")
	}
}

func TestHandleDocumentSymbolMessage(t *testing.T) {
	s := NewServer()
	s.handler.DidOpen(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: "test.yaml", Text: "project: myapp\nservices:\n  - name: api\n"},
	})
	s.handleMessage(`{"jsonrpc":"2.0","id":7,"method":"textDocument/documentSymbol","params":{"textDocument":{"uri":"test.yaml"}}}`)
}
