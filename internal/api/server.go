package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
	"github.com/NAEOS-foundation/naeos/internal/compiler"
	contextbundle "github.com/NAEOS-foundation/naeos/internal/context/bundle"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/version"
)

type Server struct {
	Addr       string
	Router     *http.ServeMux
	server     *http.Server
	Auth       *AuthConfig
	Limiter    *RateLimiter
	jwt        *JWTValidator
	parser     parser.Parser
	compiler   *compiler.Compiler
	bundle     *contextbundle.Generator
	store      *artifacts.Store
	pipelines  []pipelineRun
}

type pipelineRun struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Project   string `json:"project"`
	Modules   int    `json:"modules"`
	Services  int    `json:"services"`
	CreatedAt string `json:"created_at"`
}

type AuthConfig struct {
	JWTSecret string
	Enabled   bool
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewServer(addr string, auth *AuthConfig) *Server {
	store := artifacts.NewStore(".naeos/artifacts")
	_ = store.LoadFromDisk()

	s := &Server{
		Addr:     addr,
		Router:   http.NewServeMux(),
		Auth:     auth,
		Limiter:  NewRateLimiter(100, time.Minute),
		parser:   parser.NewParser(),
		compiler: compiler.New(),
		store:    store,
	}

	if auth != nil && auth.JWTSecret != "" {
		s.jwt = NewJWTValidator(auth.JWTSecret)
	}
	s.bundle = contextbundle.NewGenerator(s.compiler)

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Health
	s.Router.HandleFunc("/api/v1/health", s.handleHealth)

	// Spec endpoints
	s.Router.HandleFunc("/api/v1/specs", s.handleSpecs)
	s.Router.HandleFunc("/api/v1/specs/validate", s.handleSpecValidate)
	s.Router.HandleFunc("/api/v1/specs/compile", s.handleSpecCompile)

	// Pipeline endpoints
	s.Router.HandleFunc("/api/v1/pipeline/run", s.handlePipelineRun)
	s.Router.HandleFunc("/api/v1/pipeline/status", s.handlePipelineStatus)

	// Artifact endpoints
	s.Router.HandleFunc("/api/v1/artifacts", s.handleArtifacts)

	// Context endpoints
	s.Router.HandleFunc("/api/v1/context/generate", s.handleContextGenerate)

	// MCP endpoints
	s.Router.HandleFunc("/api/v1/mcp/message", s.handleMCPMessage)

	// System endpoints
	s.Router.HandleFunc("/api/v1/version", s.handleVersion)
	s.Router.HandleFunc("/api/v1/config/schema", s.handleConfigSchema)
	s.Router.HandleFunc("/api/v1/pipelines", s.handlePipelines)
}

func (s *Server) handlerWithMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Rate limit
		if !s.Limiter.Allow(r.RemoteAddr) {
			s.writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Auth
		if s.Auth.Enabled && r.URL.Path != "/api/v1/health" {
			token := r.Header.Get("Authorization")
			if token == "" {
				s.writeError(w, http.StatusUnauthorized, "authorization required")
				return
			}
			token = strings.TrimPrefix(token, "Bearer ")
			if s.jwt != nil {
				_, err := s.jwt.Validate(token)
				if err != nil {
					s.writeError(w, http.StatusUnauthorized, "invalid token: "+err.Error())
					return
				}
			}
		}

		handler(w, r)
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: status >= 200 && status < 300,
		Data:    data,
	})
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"version": version.String(),
		"uptime":  time.Since(startTime).String(),
	})
}

func (s *Server) handleSpecs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"count": len(s.pipelines),
		})
	case "POST":
		var req struct {
			Spec string `json:"spec"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Spec == "" {
			s.writeError(w, http.StatusBadRequest, "spec field required")
			return
		}
		doc, err := s.parser.Parse(req.Spec)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "parse error: "+err.Error())
			return
		}
		s.writeJSON(w, http.StatusCreated, map[string]interface{}{
			"message":  "spec received and parsed",
			"project":  doc.Project,
			"modules":  len(doc.Modules),
			"services": len(doc.Services),
		})
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSpecValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Spec string `json:"spec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Spec == "" {
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":    false,
			"errors":   []string{"spec field is required"},
			"warnings": []string{},
		})
		return
	}
	_, err := s.parser.Parse(req.Spec)
	if err != nil {
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":    false,
			"errors":   []string{err.Error()},
			"warnings": []string{},
		})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":    true,
		"errors":   []string{},
		"warnings": []string{},
	})
}

func (s *Server) handleSpecCompile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Spec   string `json:"spec"`
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Spec == "" {
		s.writeError(w, http.StatusBadRequest, "spec field required")
		return
	}
	doc, err := s.parser.Parse(req.Spec)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "parse error: "+err.Error())
		return
	}
	b := s.bundle.GenerateFromSpec(doc)
	targets := s.compiler.Targets()
	if len(targets) == 0 {
		targets = []compiler.Target{"copilot", "claude", "cursor", "gemini", "codex", "opencode"}
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"compiled":  true,
		"targets":   targets,
		"bundle":    b.ToMarkdown(),
		"project":   doc.Project,
		"modules":   len(doc.Modules),
		"services":  len(doc.Services),
	})
}

func (s *Server) handlePipelineRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Spec   string `json:"spec"`
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Spec == "" {
		s.writeError(w, http.StatusBadRequest, "spec field required")
		return
	}
	doc, err := s.parser.Parse(req.Spec)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "parse error: "+err.Error())
		return
	}
	b := s.bundle.GenerateFromSpec(doc)
	pipelineID := fmt.Sprintf("pipeline-%d", time.Now().UnixNano())
	run := pipelineRun{
		ID:        pipelineID,
		Status:    "completed",
		Project:   doc.Project,
		Modules:   len(doc.Modules),
		Services:  len(doc.Services),
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.pipelines = append(s.pipelines, run)
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"pipeline_id": pipelineID,
		"status":      "completed",
		"project":     doc.Project,
		"modules":     len(doc.Modules),
		"services":    len(doc.Services),
		"bundle":      b.ToMarkdown(),
	})
}

func (s *Server) handlePipelineStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var lastRun *pipelineRun
	if len(s.pipelines) > 0 {
		last := s.pipelines[len(s.pipelines)-1]
		lastRun = &last
	}
	status := "idle"
	if lastRun != nil && lastRun.Status == "running" {
		status = "running"
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   status,
		"total":    len(s.pipelines),
		"last_run": lastRun,
	})
}

func (s *Server) handleArtifacts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		list := s.store.List()
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"artifacts": list,
			"count":     len(list),
		})
	case "POST":
		var req struct {
			Path    string `json:"path"`
			Content string `json:"content"`
			Kind    string `json:"kind"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Path == "" || req.Content == "" {
			s.writeError(w, http.StatusBadRequest, "path and content required")
			return
		}
		kind := artifacts.DetectKind(req.Path)
		if req.Kind != "" {
			kind = artifacts.ArtifactKind(req.Kind)
		}
		artifact, err := s.store.Add(req.Path, []byte(req.Content), kind)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "failed to store artifact: "+err.Error())
			return
		}
		_ = s.store.WriteToDisk()
		s.writeJSON(w, http.StatusCreated, map[string]interface{}{
			"message":  "artifact stored",
			"artifact": artifact,
		})
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleContextGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Spec   string `json:"spec"`
		Format string `json:"format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Spec == "" {
		s.writeError(w, http.StatusBadRequest, "spec field required")
		return
	}
	doc, err := s.parser.Parse(req.Spec)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "parse error: "+err.Error())
		return
	}
	b := s.bundle.GenerateFromSpec(doc)
	format := req.Format
	if format == "" {
		format = "markdown"
	}
	var text string
	switch format {
	case "plain":
		text = b.ToPlainText()
	default:
		text = b.ToMarkdown()
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"context": text,
		"format":  format,
		"project": doc.Project,
	})
}

func (s *Server) handleMCPMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
		ID      any             `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid JSON-RPC request")
		return
	}

	type jsonRPCResponse struct {
		JSONRPC string `json:"jsonrpc"`
		Result  any    `json:"result,omitempty"`
		Error   any    `json:"error,omitempty"`
		ID      any    `json:"id"`
	}
	resp := jsonRPCResponse{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]any{"name": "naeos-api-mcp", "version": version.String()},
		}
	case "tools/list":
		resp.Result = map[string]any{
			"tools": []map[string]any{
				{"name": "parse_spec", "description": "Parse a NAEOS specification"},
				{"name": "validate_spec", "description": "Validate a specification"},
				{"name": "compile_spec", "description": "Compile specification to AI instructions"},
			},
		}
	case "tools/call":
		var params struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = map[string]any{"code": -32602, "message": "invalid params"}
		} else {
			spec, _ := params.Arguments["spec"].(string)
			switch params.Name {
			case "parse_spec":
				if spec == "" {
					resp.Error = map[string]any{"code": -32000, "message": "spec is required"}
				} else {
					doc, err := s.parser.Parse(spec)
					if err != nil {
						resp.Error = map[string]any{"code": -32000, "message": err.Error()}
					} else {
						resp.Result = map[string]any{
							"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("Project: %s\nModules: %d\nServices: %d", doc.Project, len(doc.Modules), len(doc.Services))}},
						}
					}
				}
			case "validate_spec":
				if spec == "" {
					resp.Error = map[string]any{"code": -32000, "message": "spec is required"}
				} else {
					doc, err := s.parser.Parse(spec)
					if err != nil {
						resp.Result = map[string]any{
							"isError": true,
							"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("Validation error: %v", err)}},
						}
					} else {
						status := "PASS"
						if len(doc.Modules) == 0 {
							status = "WARN"
						}
						resp.Result = map[string]any{
							"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("Status: %s\nProject: %s\nModules: %d\nServices: %d", status, doc.Project, len(doc.Modules), len(doc.Services))}},
						}
					}
				}
			case "compile_spec":
				if spec == "" {
					resp.Error = map[string]any{"code": -32000, "message": "spec is required"}
				} else {
					doc, err := s.parser.Parse(spec)
					if err != nil {
						resp.Error = map[string]any{"code": -32000, "message": err.Error()}
					} else {
						b := s.bundle.GenerateFromSpec(doc)
						target, _ := params.Arguments["target"].(string)
						if target == "" {
							target = "copilot"
						}
						resp.Result = map[string]any{
							"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("Compiled for target: %s\n\n%s", target, b.ToMarkdown())}},
						}
					}
				}
			default:
				resp.Error = map[string]any{"code": -32601, "message": "method not found"}
			}
		}
	default:
		resp.Error = map[string]any{"code": -32601, "message": "method not found"}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{
		"version":  version.String(),
		"go":       fmt.Sprintf("go%d", 25),
		"platform": "naeos-api",
	})
}

func (s *Server) handleConfigSchema(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":        map[string]string{"type": "string", "description": "project name"},
			"version":     map[string]string{"type": "string", "description": "project version"},
			"output_dir":  map[string]string{"type": "string", "description": "output directory"},
			"mode":        map[string]string{"type": "string", "description": "pipeline mode"},
			"verbose":     map[string]string{"type": "boolean", "description": "verbose output"},
		},
		"required": []string{"name"},
	})
}

func (s *Server) handlePipelines(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"pipelines": s.pipelines,
		"total":     len(s.pipelines),
	})
}

var startTime = time.Now()

func (s *Server) Start() error {
	wrappedHandler := s.handlerWithMiddleware(s.Router.ServeHTTP)

	s.server = &http.Server{
		Addr:         s.Addr,
		Handler:      wrappedHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}()

	log.Printf("Starting NAEOS API server on %s", s.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}
