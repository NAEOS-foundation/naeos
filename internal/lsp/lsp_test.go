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
