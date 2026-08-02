package lsp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	mu         sync.Mutex
	reader     io.Reader
	writer     io.Writer
	documents  *DocumentManager
	completion *CompletionProvider
	diagnostic *DiagnosticProvider
	handler    *Handler
	closed     bool
}

func NewServer() *Server {
	s := &Server{
		reader:     os.Stdin,
		writer:     os.Stdout,
		documents:  NewDocumentManager(),
		completion: NewCompletionProvider(),
		diagnostic: NewDiagnosticProvider(),
	}
	s.handler = NewHandler(s)
	return s
}

func (s *Server) Documents() *DocumentManager     { return s.documents }
func (s *Server) Completion() *CompletionProvider { return s.completion }
func (s *Server) Diagnostic() *DiagnosticProvider { return s.diagnostic }

func (s *Server) Run() error {
	br := bufio.NewReader(s.reader)
	for {
		raw, err := readMessage(br)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read message: %w", err)
		}
		s.handleMessage(string(raw))
	}
}

func readMessage(r *bufio.Reader) ([]byte, error) {
	var length int
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(line[len("Content-Length:"):])
			length, err = strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %s", val)
			}
		}
	}

	body := make([]byte, length)
	_, err := io.ReadFull(r, body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}

func (s *Server) handleMessage(raw string) {
	var request Request
	if err := json.Unmarshal([]byte(raw), &request); err != nil {
		slog.Error("failed to parse request", "error", err)
		return
	}

	switch request.Method {
	case MethodInitialize:
		var params InitializeParams
		if request.Params != nil {
			data, _ := json.Marshal(request.Params)
			_ = json.Unmarshal(data, &params)
		}
		result := s.handler.Initialize(params)
		s.sendResponse(request.ID, result)

	case MethodShutdown:
		s.sendResponse(request.ID, nil)

	case MethodExit:
		s.closed = true

	case MethodDidOpen:
		var params DidOpenTextDocumentParams
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		s.handler.DidOpen(params)

	case MethodDidChange:
		var params DidChangeTextDocumentParams
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		s.handler.DidChange(params)

	case MethodDidClose:
		var params DidCloseTextDocumentParams
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		s.handler.DidClose(params)

	case MethodCompletion:
		var params CompletionParams
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		result := s.handler.Completion(params)
		s.sendResponse(request.ID, result)

	case MethodHover:
		var params HoverParams
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		result := s.handler.Hover(params)
		s.sendResponse(request.ID, result)

	case MethodDefinition:
		var params DefinitionParams
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		result := s.handler.Definition(params)
		s.sendResponse(request.ID, result)

	case MethodDocumentSymbol:
		var params DocumentSymbolParams
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		result := s.handler.DocumentSymbol(params)
		s.sendResponse(request.ID, result)

	case MethodCodeAction:
		var params CodeActionParams
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		result := s.handler.CodeAction(params)
		s.sendResponse(request.ID, result)

	default:
		if request.ID != nil {
			s.sendResponse(request.ID, nil)
		}
	}
}

func (s *Server) sendResponse(id any, result any) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.writeMessage(resp)
}

func (s *Server) SendNotification(method Method, params any) {
	notif := Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	s.writeMessage(notif)
}

func (s *Server) writeMessage(msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal message", "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := fmt.Fprint(s.writer, header); err != nil {
		slog.Error("failed to write header", "error", err)
		return
	}
	if _, err := s.writer.Write(data); err != nil {
		slog.Error("failed to write body", "error", err)
		return
	}
	if f, ok := s.writer.(*os.File); ok && f != os.Stdout {
		if err := f.Sync(); err != nil {
			slog.Error("failed to sync file", "error", err)
		}
	}
}
