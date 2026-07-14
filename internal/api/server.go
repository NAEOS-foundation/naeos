package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
	"github.com/NAEOS-foundation/naeos/internal/audit"
	"github.com/NAEOS-foundation/naeos/internal/cloud"
	"github.com/NAEOS-foundation/naeos/internal/compiler"
	contextbundle "github.com/NAEOS-foundation/naeos/internal/context/bundle"
	"github.com/NAEOS-foundation/naeos/internal/database"
	"github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/mcp"
	"github.com/NAEOS-foundation/naeos/internal/monitoring"
	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/version"
	naeosws "github.com/NAEOS-foundation/naeos/internal/websocket"
)

// Server is the main HTTP API server for the NAEOS platform.
type Server struct {
	Addr        string
	Router      *http.ServeMux
	server      *http.Server
	Auth        *AuthConfig
	CORS        *CORSConfig
	MaxBodySize int64
	Limiter     *RateLimiter
	APIKeys     map[string]*RateLimiter
	apiKeysMu   sync.RWMutex
	jwt         *JWTValidator
	parser      parser.Parser
	compiler    *compiler.Compiler
	bundle      *contextbundle.Generator
	store       *artifacts.Store
	pipelines   []pipelineRun
	pipelineJobs map[string]*pipelineJob
	jobsMu      sync.RWMutex
	deployments []cloudDeployment
	deployMu       sync.RWMutex
	plugins        *pluginhost.Manager
	metricsRegistry *monitoring.Registry
	auditor        audit.Auditor
	wsServer       *naeosws.Server
	mcpServer      *mcp.Server
	db             database.Database
}

type pipelineRun struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Project   string `json:"project"`
	Modules   int    `json:"modules"`
	Services  int    `json:"services"`
	CreatedAt string `json:"created_at"`
}

type pipelineJob struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Project   string `json:"project"`
	CreatedAt string `json:"created_at"`
	Error     string `json:"error,omitempty"`
}

type cloudDeployment struct {
	ID        string              `json:"id"`
	Provider  cloud.CloudProvider `json:"provider"`
	Project   string              `json:"project"`
	Status    string              `json:"status"`
	CreatedAt string              `json:"created_at"`
	Error     string              `json:"error,omitempty"`
}

// AuthConfig holds authentication settings for the API server.
type AuthConfig struct {
	JWTSecret string
	Enabled   bool
}

// CORSConfig holds Cross-Origin Resource Sharing settings.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// APIResponse is the standard JSON envelope returned by API endpoints.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    any `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse contains detailed error information in API responses.
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details any `json:"details,omitempty"`
}

// NewServer creates a new API server with the given address and auth configuration.
func NewServer(addr string, auth *AuthConfig) *Server {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	store := artifacts.NewStore(".naeos/artifacts")
	_ = store.LoadFromDisk()

	metrics := monitoring.NewMetrics()

	s := &Server{
		Addr:  addr,
		Router: http.NewServeMux(),
		Auth:  auth,
		CORS: &CORSConfig{
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://localhost:5173",
				"http://localhost:8080",
			},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
		MaxBodySize: 10 << 20,
		Limiter:     NewRateLimiter(100, time.Minute),
		APIKeys:     make(map[string]*RateLimiter),
		parser:      parser.NewParser(),
		compiler:    compiler.New(),
		store:       store,
		pipelineJobs: make(map[string]*pipelineJob),
		plugins:     pluginhost.NewManager(".naeos/plugins"),
	}

	if auth != nil && auth.JWTSecret != "" {
		s.jwt = NewJWTValidator(auth.JWTSecret)
	}
	s.bundle = contextbundle.NewGenerator(s.compiler)
	s.mcpServer = mcp.NewServer(s.compiler, s.bundle)
	s.mcpServer.SetArtifactStore(store)
	s.mcpServer.SetPluginManager(s.plugins)

	s.metricsRegistry = metrics.Registry()
	s.auditor = audit.NewMemoryAuditor()
	s.setupRoutes()
	return s
}

func (s *Server) auditEvent(r *http.Request, action, resource, resourceID, status, details string) {
	if s.auditor == nil {
		return
	}
	userID := ""
	ip := ""
	ua := ""
	if r != nil {
		if uid := r.Header.Get("X-User-ID"); uid != "" {
			userID = uid
		}
		ip = r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}
		ua = r.UserAgent()
	}
	if err := s.auditor.Log(audit.AuditEvent{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IP:         ip,
		UserAgent:  ua,
		Status:     status,
		Details:    details,
	}); err != nil {
		slog.Error("failed to write audit event", "error", err,
			"action", action, "resource", resource)
	}
}

func (s *Server) setupRoutes() {
	// Monitoring endpoints
	s.Router.HandleFunc("/metrics", s.handleMetrics)
	s.Router.HandleFunc("/healthz", s.handleHealthz)
	s.Router.HandleFunc("/readyz", s.handleReadyz)

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

	// Cloud endpoints
	s.Router.HandleFunc("/api/v1/cloud/plan", s.handleCloudPlan)
	s.Router.HandleFunc("/api/v1/cloud/deploy", s.handleCloudDeploy)
	s.Router.HandleFunc("/api/v1/cloud/destroy", s.handleCloudDestroy)
	s.Router.HandleFunc("/api/v1/cloud/status", s.handleCloudStatus)

	// Plugin endpoints
	s.Router.HandleFunc("/api/v1/plugins", s.handlePlugins)
	s.Router.HandleFunc("/api/v1/plugins/execute", s.handlePluginExecute)
	s.Router.HandleFunc("/api/v1/plugins/", s.handlePluginByName)

	// System endpoints
	s.Router.HandleFunc("/api/v1/version", s.handleVersion)
	s.Router.HandleFunc("/api/v1/config/schema", s.handleConfigSchema)
	s.Router.HandleFunc("/api/v1/pipelines", s.handlePipelines)

	// OIDC discovery
	s.Router.HandleFunc("/.well-known/openid-configuration", s.handleOIDCDiscovery)
}

// SetWebSocketServer attaches a WebSocket server to the API server.
func (s *Server) SetWebSocketServer(ws *naeosws.Server) {
	s.wsServer = ws
}

// SetDatabase attaches a database to the API server for persistence.
func (s *Server) SetDatabase(db database.Database) {
	s.db = db
}

// RegisterAPIKey registers an API key with its associated rate limit.
func (s *Server) RegisterAPIKey(key string, requestsPerSecond int) {
	s.apiKeysMu.Lock()
	defer s.apiKeysMu.Unlock()
	s.APIKeys[key] = NewRateLimiter(requestsPerSecond, time.Second)
}

func (s *Server) handlerWithMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = GenerateRequestID()
		}
		r = r.WithContext(ContextWithRequestID(r.Context(), requestID))
		w.Header().Set("X-Request-ID", requestID)

		// Body size limit for methods that carry a payload
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if s.MaxBodySize > 0 {
				exceeded := false
				r.Body = &maxBytesBody{
					ReadCloser: http.MaxBytesReader(w, r.Body, s.MaxBodySize),
					exceeded:   &exceeded,
				}
				w = &maxBytesResponseWriter{
					ResponseWriter: w,
					exceeded:       &exceeded,
				}
			}
		}

		// Rate limit - check API key first, fall back to IP
		clientID := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientID = forwarded
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			s.apiKeysMu.RLock()
			limiter, exists := s.APIKeys[apiKey]
			s.apiKeysMu.RUnlock()
			if exists {
				if !limiter.Allow(apiKey) {
					s.writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
					return
				}
			} else {
				if !s.Limiter.Allow(clientID) {
					s.writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
					return
				}
			}
		} else {
			if !s.Limiter.Allow(clientID) {
				s.writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
		}

		// CORS
		origin := r.Header.Get("Origin")
		if s.CORS != nil && originAllowed(origin, s.CORS.AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if s.CORS.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		} else if s.CORS == nil {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		methods := "GET, POST, PUT, DELETE, OPTIONS"
		headers := "Content-Type, Authorization, X-Request-ID"
		if s.CORS != nil {
			if len(s.CORS.AllowedMethods) > 0 {
				methods = joinStrings(s.CORS.AllowedMethods)
			}
			if len(s.CORS.AllowedHeaders) > 0 {
				headers = joinStrings(s.CORS.AllowedHeaders)
				if !containsHeader(s.CORS.AllowedHeaders, "X-Request-ID") {
					headers += ", X-Request-ID"
				}
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", methods)
		w.Header().Set("Access-Control-Allow-Headers", headers)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
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

		slog.Info("request", "method", r.Method, "path", r.URL.Path, "request_id", requestID)
		handler(w, r)
	}
}

func containsHeader(headers []string, target string) bool {
	for _, h := range headers {
		if strings.EqualFold(h, target) {
			return true
		}
	}
	return false
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data any) {
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
		Data: ErrorResponse{
			Error:   http.StatusText(status),
			Message: message,
		},
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	dbStatus := "not_configured"
	if s.db != nil {
		if err := s.db.HealthCheck(); err != nil {
			dbStatus = "unhealthy: " + err.Error()
		} else {
			dbStatus = "healthy"
		}
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status":   "healthy",
		"version":  version.String(),
		"uptime":   time.Since(startTime).String(),
		"database": dbStatus,
	})
}

func (s *Server) handleSpecs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.writeJSON(w, http.StatusOK, map[string]any{
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
		s.writeJSON(w, http.StatusCreated, map[string]any{
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
		s.writeJSON(w, http.StatusOK, map[string]any{
			"valid":    false,
			"errors":   []string{"spec field is required"},
			"warnings": []string{},
		})
		return
	}
	_, err := s.parser.Parse(req.Spec)
	if err != nil {
		s.writeJSON(w, http.StatusOK, map[string]any{
			"valid":    false,
			"errors":   []string{err.Error()},
			"warnings": []string{},
		})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
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
		s.writeError(w, http.StatusServiceUnavailable, "no compiler targets available; check compiler configuration")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
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

	jobID := fmt.Sprintf("pipeline-%d", time.Now().UnixNano())
	job := &pipelineJob{
		ID:        jobID,
		Status:    "running",
		Project:   doc.Project,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.jobsMu.Lock()
	s.pipelineJobs[jobID] = job
	s.jobsMu.Unlock()

	go func() {
		b := s.bundle.GenerateFromSpec(doc)
		run := pipelineRun{
			ID:        jobID,
			Status:    "completed",
			Project:   doc.Project,
			Modules:   len(doc.Modules),
			Services:  len(doc.Services),
			CreatedAt: time.Now().Format(time.RFC3339),
		}
		s.pipelines = append(s.pipelines, run)

		if s.db != nil {
			s.db.Exec("INSERT INTO pipeline_runs (id, status, project, modules, services, created_at) VALUES (?, ?, ?, ?, ?, ?)",
				run.ID, run.Status, run.Project, run.Modules, run.Services, run.CreatedAt)
		}

		s.jobsMu.Lock()
		job.Status = "completed"
		s.jobsMu.Unlock()
		_ = b
	}()

	s.writeJSON(w, http.StatusAccepted, map[string]any{
		"job_id": jobID,
		"status": "running",
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
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status":   status,
		"total":    len(s.pipelines),
		"last_run": lastRun,
	})
}

func (s *Server) handleArtifacts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		list := s.store.List()
		offset, limit := parsePagination(r, 50, 200)
		total := len(list)
		start := offset
		if start > total {
			start = total
		}
		end := start + limit
		if end > total {
			end = total
		}
		s.writeJSON(w, http.StatusOK, map[string]any{
			"artifacts": list[start:end],
			"count":     total,
			"page":      offset/limit + 1,
			"limit":     limit,
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
		s.writeJSON(w, http.StatusCreated, map[string]any{
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
	s.writeJSON(w, http.StatusOK, map[string]any{
		"context": text,
		"format":  format,
		"project": doc.Project,
	})
}

func (s *Server) handleMCPMessage(w http.ResponseWriter, r *http.Request) {
	s.mcpServer.HandleMCP(w, r)
}

func (s *Server) handleCloudPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Provider  string                   `json:"provider"`
		Project   string                   `json:"project"`
		Region    string                   `json:"region"`
		Resources []map[string]any `json:"resources"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Provider == "" || req.Project == "" {
		s.writeError(w, http.StatusBadRequest, errors.New(errors.ErrValidation, "provider and project are required").Error())
		return
	}

	adapter, err := cloud.GetAdapter(cloud.CloudProvider(req.Provider))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, errors.Wrap(errors.ErrCloud, "invalid provider", err).Error())
		return
	}

	var resources []cloud.Resource
	for _, r := range req.Resources {
		resources = append(resources, cloud.Resource{
			Name: fmt.Sprintf("%v", r["name"]),
			Type: fmt.Sprintf("%v", r["type"]),
			Spec: r,
		})
	}

	config := &cloud.DeployConfig{
		Provider: cloud.CloudProvider(req.Provider),
		Region:   req.Region,
		Project:  req.Project,
		Resources: resources,
	}

	if err := adapter.Validate(config); err != nil {
		s.writeError(w, http.StatusBadRequest, errors.Wrap(errors.ErrCloud, "validation failed", err).Error())
		return
	}

	result, err := adapter.ExportTerraform(config)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, errors.Wrap(errors.ErrCloud, "plan generation failed", err).Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"provider": req.Provider,
		"project":  req.Project,
		"hcl":      result,
	})
}

func (s *Server) handleCloudDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Provider  string                   `json:"provider"`
		Project   string                   `json:"project"`
		Region    string                   `json:"region"`
		Resources []map[string]any `json:"resources"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Provider == "" || req.Project == "" {
		s.writeError(w, http.StatusBadRequest, errors.New(errors.ErrValidation, "provider and project are required").Error())
		return
	}

	adapter, err := cloud.GetAdapter(cloud.CloudProvider(req.Provider))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, errors.Wrap(errors.ErrCloud, "invalid provider", err).Error())
		return
	}

	var resources []cloud.Resource
	for _, r := range req.Resources {
		resources = append(resources, cloud.Resource{
			Name: fmt.Sprintf("%v", r["name"]),
			Type: fmt.Sprintf("%v", r["type"]),
			Spec: r,
		})
	}

	config := &cloud.DeployConfig{
		Provider: cloud.CloudProvider(req.Provider),
		Region:   req.Region,
		Project:  req.Project,
		Resources: resources,
	}

	result, err := adapter.Deploy(config)
	if err != nil {
		deploymentID := fmt.Sprintf("deploy-%d", time.Now().UnixNano())
		s.deployMu.Lock()
		s.deployments = append(s.deployments, cloudDeployment{
			ID:        deploymentID,
			Provider:  cloud.CloudProvider(req.Provider),
			Project:   req.Project,
			Status:    "failed",
			CreatedAt: time.Now().Format(time.RFC3339),
			Error:     err.Error(),
		})
		s.deployMu.Unlock()
		s.writeError(w, http.StatusInternalServerError, errors.Wrap(errors.ErrCloud, "deploy failed", err).Error())
		return
	}

	deploymentID := fmt.Sprintf("deploy-%d", time.Now().UnixNano())
	s.deployMu.Lock()
	s.deployments = append(s.deployments, cloudDeployment{
		ID:        deploymentID,
		Provider:  cloud.CloudProvider(req.Provider),
		Project:   req.Project,
		Status:    "completed",
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	s.deployMu.Unlock()

	s.writeJSON(w, http.StatusOK, map[string]any{
		"deployment_id": deploymentID,
		"provider":      req.Provider,
		"project":       req.Project,
		"status":        result.Status,
		"resources":     result.Resources,
	})
}

func (s *Server) handleCloudDestroy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Provider  string                   `json:"provider"`
		Project   string                   `json:"project"`
		Region    string                   `json:"region"`
		Resources []map[string]any `json:"resources"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Provider == "" || req.Project == "" {
		s.writeError(w, http.StatusBadRequest, errors.New(errors.ErrValidation, "provider and project are required").Error())
		return
	}

	adapter, err := cloud.GetAdapter(cloud.CloudProvider(req.Provider))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, errors.Wrap(errors.ErrCloud, "invalid provider", err).Error())
		return
	}

	var resources []cloud.Resource
	for _, r := range req.Resources {
		resources = append(resources, cloud.Resource{
			Name: fmt.Sprintf("%v", r["name"]),
			Type: fmt.Sprintf("%v", r["type"]),
			Spec: r,
		})
	}

	config := &cloud.DeployConfig{
		Provider: cloud.CloudProvider(req.Provider),
		Region:   req.Region,
		Project:  req.Project,
		Resources: resources,
	}

	if err := adapter.Destroy(config); err != nil {
		s.writeError(w, http.StatusInternalServerError, errors.Wrap(errors.ErrCloud, "destroy failed", err).Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"provider": req.Provider,
		"project":  req.Project,
		"status":   "destroyed",
	})
}

func (s *Server) handleCloudStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.deployMu.RLock()
	defer s.deployMu.RUnlock()

	offset, limit := parsePagination(r, 50, 200)
	total := len(s.deployments)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"deployments": s.deployments[start:end],
		"count":       total,
		"page":        offset/limit + 1,
		"limit":       limit,
	})
}

func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		plugins := s.plugins.List()
		offset, limit := parsePagination(r, 50, 200)
		total := len(plugins)
		start := offset
		if start > total {
			start = total
		}
		end := start + limit
		if end > total {
			end = total
		}
		s.writeJSON(w, http.StatusOK, map[string]any{
			"plugins": plugins[start:end],
			"count":   total,
			"page":    offset/limit + 1,
			"limit":   limit,
		})
	case "POST":
		var req struct {
			Name    string `json:"name"`
			Source  string `json:"source"`
			Version string `json:"version"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" || req.Source == "" {
			s.writeError(w, http.StatusBadRequest, "name and source are required")
			return
		}
		info, err := s.plugins.Install(req.Source)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "install failed: "+err.Error())
			return
		}
		s.writeJSON(w, http.StatusCreated, map[string]any{
			"name":         info.Name,
			"version":      info.Version,
			"description":  info.Description,
			"kind":         "native",
			"enabled":      info.Enabled,
			"installed_at": info.StartedAt.Format(time.RFC3339),
		})
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handlePluginExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Name   string         `json:"name"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		s.writeError(w, http.StatusBadRequest, errors.New(errors.ErrValidation, "plugin name is required").Error())
		return
	}
	if req.Action == "" {
		req.Action = "execute"
	}

	result, err := s.plugins.Execute(context.Background(), req.Name, req.Action, req.Params)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, errors.Wrap(errors.ErrPlugin, "execution failed", err).Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"plugin": req.Name,
		"action": req.Action,
		"result": result,
	})
}

func (s *Server) handlePluginByName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/plugins/")
	if name == "" {
		s.writeError(w, http.StatusBadRequest, errors.New(errors.ErrValidation, "plugin name is required").Error())
		return
	}

	switch r.Method {
	case "GET":
		info, ok := s.plugins.GetInfo(name)
		if !ok {
			s.writeError(w, http.StatusNotFound, fmt.Sprintf("plugin %s not found", name))
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]any{
			"name":         info.Name,
			"version":      info.Version,
			"description":  info.Description,
			"kind":         "native",
			"enabled":      info.Enabled,
			"installed_at": info.StartedAt.Format(time.RFC3339),
		})
	case "DELETE":
		if err := s.plugins.Uninstall(name); err != nil {
			s.writeError(w, http.StatusNotFound, fmt.Sprintf("plugin %s not found", name))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
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
	offset, limit := parsePagination(r, 50, 200)
	total := len(s.pipelines)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"pipelines": s.pipelines[start:end],
		"total":     total,
		"page":      offset/limit + 1,
		"limit":     limit,
	})
}

func (s *Server) issuerFromRequest(r *http.Request) string {
	issuer := fmt.Sprintf("http://%s", r.Host)
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
		issuer = fmt.Sprintf("http://%s", fwd)
	}
	return issuer
}

func (s *Server) handleOIDCDiscovery(w http.ResponseWriter, r *http.Request) {
	if s.jwt == nil {
		s.writeError(w, http.StatusNotFound, "OIDC not configured")
		return
	}
	issuer := s.issuerFromRequest(r)
	doc := s.jwt.OIDCDiscoveryDocument(issuer)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte(s.metricsRegistry.FormatPrometheus()))
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ready"}`)
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

func parsePagination(r *http.Request, defaultLimit, maxLimit int) (offset, limit int) {
	limit = defaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= maxLimit {
			limit = v
		}
	}
	offset = 0
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			offset = (v - 1) * limit
		}
	}
	return
}

var startTime = time.Now()

// Start begins listening for HTTP requests and handles graceful shutdown.
func (s *Server) Start() error {
	wrappedHandler := s.loggingMiddleware(s.handlerWithMiddleware(s.Router.ServeHTTP))

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

		slog.Warn("shutting down server", "component", "api-server")
		if s.wsServer != nil {
			s.wsServer.Stop()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}()

	slog.Info("starting NAEOS API server", "addr", s.Addr, "component", "api-server")
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server and any attached WebSocket server.
func (s *Server) Stop() error {
	if s.wsServer != nil {
		s.wsServer.Stop()
	}
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}
