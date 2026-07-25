package lsp

type Method string

const (
	MethodInitialize       Method = "initialize"
	MethodShutdown         Method = "shutdown"
	MethodExit             Method = "exit"
	MethodDidOpen          Method = "textDocument/didOpen"
	MethodDidChange        Method = "textDocument/didChange"
	MethodDidClose         Method = "textDocument/didClose"
	MethodCompletion       Method = "textDocument/completion"
	MethodHover            Method = "textDocument/hover"
	MethodDefinition       Method = "textDocument/definition"
	MethodDocumentSymbol   Method = "textDocument/documentSymbol"
	MethodPublishDiag      Method = "textDocument/publishDiagnostics"
)

type Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  Method `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  Method `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type InitializeParams struct {
	ProcessID int    `json:"processId"`
	RootURI   string `json:"rootUri"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument"`
}

type TextDocumentClientCapabilities struct {
	Completion           *CompletionCapability    `json:"completion,omitempty"`
	Hover                *HoverCapability          `json:"hover,omitempty"`
	Definition           *DefinitionCapability      `json:"definition,omitempty"`
	DocumentSymbol       *DocumentSymbolCapability  `json:"documentSymbol,omitempty"`
	PublishDiagnostics   *DiagnosticCapability      `json:"publishDiagnostics,omitempty"`
}

type CompletionCapability struct{}
type HoverCapability struct{}
type DefinitionCapability struct{}
type DocumentSymbolCapability struct{}
type DiagnosticCapability struct{}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSync   int                `json:"textDocumentSync"`
	CompletionProvider *CompletionOptions `json:"completionProvider,omitempty"`
	HoverProvider      bool               `json:"hoverProvider,omitempty"`
	DefinitionProvider bool               `json:"definitionProvider,omitempty"`
	DocumentSymbolProvider bool           `json:"documentSymbolProvider,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position    Position                `json:"position"`
}

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position    Position                `json:"position"`
}

type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position    Position                `json:"position"`
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type CompletionItem struct {
	Label         string          `json:"label"`
	Kind          CompletionKind  `json:"kind,omitempty"`
	Detail        string          `json:"detail,omitempty"`
	Documentation string          `json:"documentation,omitempty"`
	InsertText    string          `json:"insertText,omitempty"`
}

type CompletionKind int

const (
	CompletionText        CompletionKind = 1
	CompletionMethod     CompletionKind = 2
	CompletionFunction   CompletionKind = 3
	CompletionConstructor CompletionKind = 4
	CompletionField      CompletionKind = 5
	CompletionVariable   CompletionKind = 6
	CompletionClass      CompletionKind = 7
	CompletionInterface  CompletionKind = 8
	CompletionModule     CompletionKind = 9
	CompletionProperty   CompletionKind = 10
	CompletionValue      CompletionKind = 12
	CompletionEnum       CompletionKind = 13
	CompletionKeyword    CompletionKind = 14
	CompletionSnippet    CompletionKind = 15
	CompletionColor      CompletionKind = 16
	CompletionFile       CompletionKind = 17
	CompletionReference  CompletionKind = 18
	CompletionFolder     CompletionKind = 19
	CompletionEnumMember CompletionKind = 20
	CompletionConstant   CompletionKind = 21
	CompletionStruct     CompletionKind = 22
	CompletionEvent      CompletionKind = 23
	CompletionOperator   CompletionKind = 24
	CompletionTypeParam  CompletionKind = 25
)

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type SymbolKind int

const (
	SymbolFile        SymbolKind = 1
	SymbolModule      SymbolKind = 2
	SymbolNamespace   SymbolKind = 3
	SymbolPackage     SymbolKind = 4
	SymbolClass       SymbolKind = 5
	SymbolMethod      SymbolKind = 6
	SymbolProperty    SymbolKind = 7
	SymbolField       SymbolKind = 8
	SymbolConstructor SymbolKind = 9
	SymbolEnum        SymbolKind = 10
	SymbolInterface   SymbolKind = 11
	SymbolFunction    SymbolKind = 12
	SymbolVariable    SymbolKind = 13
	SymbolConstant    SymbolKind = 14
	SymbolString      SymbolKind = 15
	SymbolNumber      SymbolKind = 16
	SymbolBoolean     SymbolKind = 17
	SymbolArray       SymbolKind = 18
	SymbolObject      SymbolKind = 19
	SymbolKey         SymbolKind = 20
	SymbolNull        SymbolKind = 21
	SymbolEnumMember  SymbolKind = 22
	SymbolStruct      SymbolKind = 23
	SymbolEvent       SymbolKind = 24
	SymbolOperator    SymbolKind = 25
	SymbolTypeParam   SymbolKind = 26
)

type DocumentSymbol struct {
	Name           string     `json:"name"`
	Detail         string     `json:"detail,omitempty"`
	Kind           SymbolKind `json:"kind"`
	Range          Range      `json:"range"`
	SelectionRange Range      `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range           `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Message  string          `json:"message"`
}

type DiagnosticSeverity int

const (
	DiagError       DiagnosticSeverity = 1
	DiagWarning     DiagnosticSeverity = 2
	DiagInformation DiagnosticSeverity = 3
	DiagHint        DiagnosticSeverity = 4
)
