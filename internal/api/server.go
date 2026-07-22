package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/ai"
	"github.com/NAEOS-foundation/naeos/internal/artifacts"
	"github.com/NAEOS-foundation/naeos/internal/audit"
	"github.com/NAEOS-foundation/naeos/internal/auth"
	"github.com/NAEOS-foundation/naeos/internal/cloud"
	"github.com/NAEOS-foundation/naeos/internal/compiler"
	contextbundle "github.com/NAEOS-foundation/naeos/internal/context/bundle"
	"github.com/NAEOS-foundation/naeos/internal/database"
	"github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/mcp"
	"github.com/NAEOS-foundation/naeos/internal/monitoring"
	"github.com/NAEOS-foundation/naeos/internal/multitenant"
	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
	"github.com/NAEOS-foundation/naeos/internal/profiles"
	"github.com/NAEOS-foundation/naeos/internal/schemaregistry"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/version"
	naeosws "github.com/NAEOS-foundation/naeos/internal/websocket"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

// Server is the main HTTP API server for the NAEOS platform.
type Server struct {
	Addr             string
	Router           *http.ServeMux
	server           *http.Server
	Auth             *AuthConfig
	CORS             *CORSConfig
	MaxBodySize      int64
	Limiter          *RateLimiter
	TenantLimiter    *RateLimiter
	APIKeys          map[string]*RateLimiter
	apiKeysMu        sync.RWMutex
	jwt              *JWTValidator
	authManager      *auth.Manager
	routePerms       map[string]auth.RoutePermission
	parser           parser.Parser
	compiler         *compiler.Compiler
	bundle           *contextbundle.Generator
	store            *artifacts.Store
	pipelines        []pipelineRun
	pipelinesMu      sync.RWMutex
	pipelinesFile    string
	pipelineJobs     map[string]*pipelineJob
	jobsMu           sync.RWMutex
	deployments      []cloudDeployment
	deployMu         sync.RWMutex
	plugins          *pluginhost.Manager
	metrics          *monitoring.Metrics
	metricsRegistry  *monitoring.Registry
	auditor          audit.Auditor
	memoryAuditor    *audit.MemoryAuditor
	wsServer         *naeosws.Server
	mcpServer        *mcp.Server
	db               database.Database
	profiles         *profiles.Registry
	profileSubs      map[string]*profiles.Subscription
	schemas          *schemaregistry.Registry
	profileSubMu     sync.Mutex
	pipelineObserver pipeline.PipelineObserver
	tenantWorkspace  *multitenant.Workspace
}

type multiAuditor struct {
	file   *audit.FileAuditor
	memory *audit.MemoryAuditor
}

func (m *multiAuditor) Log(event audit.AuditEvent) error {
	if m.file != nil {
		if err := m.file.Log(event); err != nil {
			return err
		}
	}
	return m.memory.Log(event)
}

type pipelineRun struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	Project     string   `json:"project"`
	Target      string   `json:"target,omitempty"`
	Modules     int      `json:"modules"`
	Services    int      `json:"services"`
	ModuleNames []string `json:"module_names,omitempty"`
	CreatedAt   string   `json:"created_at"`
	CompletedAt string   `json:"completed_at,omitempty"`
	Duration    string   `json:"duration,omitempty"`
	Error       string   `json:"error,omitempty"`
	TenantID    string   `json:"tenant_id,omitempty"`
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
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ErrorResponse contains detailed error information in API responses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// NewServer creates a new API server with the given address and auth configuration.
func NewServer(addr string, authCfg *AuthConfig) *Server {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	store := artifacts.NewStore(".naeos/artifacts")
	if err := store.LoadFromDisk(); err != nil {
		slog.Warn("failed to load artifacts from disk", "error", err)
	}

	metrics := monitoring.NewMetrics()

	routePerms := defaultRoutePermissions()
	pipelinesFile := os.Getenv("NAEOS_PIPELINES_FILE")
	if pipelinesFile == "" {
		homeDir, _ := os.UserHomeDir()
		pipelinesFile = filepath.Join(homeDir, ".naeos", "pipelines.json")
	}

	s := &Server{
		Addr:       addr,
		Router:     http.NewServeMux(),
		Auth:       authCfg,
		routePerms: routePerms,
		CORS: &CORSConfig{
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://localhost:5173",
				"http://localhost:8080",
			},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
		MaxBodySize:   10 << 20,
		Limiter:       NewRateLimiter(100, time.Minute),
		TenantLimiter: NewRateLimiter(1000, time.Minute),
		APIKeys:       make(map[string]*RateLimiter),
		parser:        parser.NewParser("."),
		compiler:      compiler.New(),
		store:         store,
		pipelinesFile: pipelinesFile,
		pipelineJobs:  make(map[string]*pipelineJob),
		plugins:       pluginhost.NewManager(".naeos/plugins"),
		profiles:      profiles.NewRegistry(),
		profileSubs:   make(map[string]*profiles.Subscription),
		schemas:       schemaregistry.New(),
	}

	s.loadPipelines()

	if authCfg != nil && authCfg.JWTSecret != "" {
		s.jwt = NewJWTValidator(authCfg.JWTSecret)
	}
	s.bundle = contextbundle.NewGenerator(s.compiler)
	s.mcpServer = mcp.NewServer(s.compiler, s.bundle)
	s.mcpServer.SetArtifactStore(store)
	s.mcpServer.SetPluginManager(s.plugins)

	s.metrics = metrics
	s.metricsRegistry = metrics.Registry()
	s.pipelineObserver = monitoring.NewMetricsObserver(metrics)

	encKey := os.Getenv("NAEOS_ENCRYPTION_KEY")
	s.authManager = auth.NewManager(encKey)
	auth.SetupDefaultRoles(s.authManager.RBAC())

	homeDir, _ := os.UserHomeDir()
	fileAuditor, err := audit.NewFileAuditor(homeDir)
	if err != nil {
		slog.Warn("failed to create file auditor, audit log will be memory-only", "error", err)
	}
	memoryAuditor := audit.NewMemoryAuditor()
	if fileAuditor != nil {
		s.auditor = &multiAuditor{file: fileAuditor, memory: memoryAuditor}
	} else {
		s.auditor = memoryAuditor
	}
	s.memoryAuditor = memoryAuditor

	s.setupRoutes()
	return s
}

func (s *Server) loadPipelines() {
	data, err := os.ReadFile(s.pipelinesFile)
	if err != nil {
		return
	}
	var runs []pipelineRun
	if err := json.Unmarshal(data, &runs); err != nil {
		slog.Warn("failed to decode pipelines file", "error", err)
		return
	}
	s.pipelinesMu.Lock()
	s.pipelines = runs
	s.pipelinesMu.Unlock()
	slog.Info("loaded pipeline runs from disk", "count", len(runs))
}

func (s *Server) savePipelines() {
	s.pipelinesMu.RLock()
	data, err := json.MarshalIndent(s.pipelines, "", "  ")
	s.pipelinesMu.RUnlock()
	if err != nil {
		slog.Warn("failed to marshal pipelines", "error", err)
		return
	}
	dir := filepath.Dir(s.pipelinesFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Warn("failed to create pipelines directory", "error", err)
		return
	}
	if err := os.WriteFile(s.pipelinesFile, data, 0600); err != nil {
		slog.Warn("failed to write pipelines file", "error", err)
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
	s.Router.HandleFunc("/api/v1/specs/visualize", s.handleSpecVisualize)

	// Pipeline endpoints
	s.Router.HandleFunc("/api/v1/pipeline/run", s.handlePipelineRun)
	s.Router.HandleFunc("/api/v1/pipeline/status", s.handlePipelineStatus)

	// Artifact endpoints
	s.Router.HandleFunc("/api/v1/artifacts", s.handleArtifacts)

	// Context endpoints
	s.Router.HandleFunc("/api/v1/context/generate", s.handleContextGenerate)

	// Profile endpoints
	s.Router.HandleFunc("/api/v1/profiles", s.handleProfiles)
	s.Router.HandleFunc("/api/v1/profiles/publish", s.handleProfilePublish)
	s.Router.HandleFunc("/api/v1/profiles/sync", s.handleProfileSync)
	s.Router.HandleFunc("/api/v1/profiles/subscribe", s.handleProfileSubscribe)
	s.Router.HandleFunc("/api/v1/profiles/unsubscribe", s.handleProfileUnsubscribe)
	s.Router.HandleFunc("/api/v1/profiles/", s.handleProfileByID)

	// Schema registry endpoints
	s.Router.HandleFunc("/api/v1/schemas", s.handleSchemas)
	s.Router.HandleFunc("/api/v1/schemas/", s.handleSchemaByPath)

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

	// AI endpoints
	s.Router.HandleFunc("/api/v1/ai/enrich/stream", s.handleAIEnrichStream)
	s.Router.HandleFunc("/api/v1/ai/explain/stream", s.handleAIExplainStream)
	s.Router.HandleFunc("/api/v1/ai/compile/stream", s.handleAICompileStream)

	// Multi-tenant endpoints
	s.Router.HandleFunc("/api/v1/tenants", s.handleTenants)
	s.Router.HandleFunc("/api/v1/tenants/", s.handleTenantByID)

	// OIDC discovery
	s.Router.HandleFunc("/.well-known/openid-configuration", s.handleOIDCDiscovery)
}

func defaultRoutePermissions() map[string]auth.RoutePermission {
	return map[string]auth.RoutePermission{
		"/api/v1/specs":                {Resource: auth.ResourceSpec, Action: auth.ActionRead},
		"/api/v1/specs/validate":       {Resource: auth.ResourceSpec, Action: auth.ActionWrite},
		"/api/v1/specs/compile":        {Resource: auth.ResourceSpec, Action: auth.ActionWrite},
		"/api/v1/specs/visualize":      {Resource: auth.ResourceSpec, Action: auth.ActionRead},
		"/api/v1/pipeline/run":         {Resource: auth.ResourcePipeline, Action: auth.ActionWrite},
		"/api/v1/pipeline/status":      {Resource: auth.ResourcePipeline, Action: auth.ActionRead},
		"/api/v1/artifacts":            {Resource: auth.ResourceArtifact, Action: auth.ActionRead},
		"/api/v1/context/generate":     {Resource: auth.ResourceSpec, Action: auth.ActionWrite},
		"/api/v1/profiles":             {Resource: auth.ResourceProfile, Action: auth.ActionRead},
		"/api/v1/profiles/publish":     {Resource: auth.ResourceProfile, Action: auth.ActionWrite},
		"/api/v1/profiles/sync":        {Resource: auth.ResourceProfile, Action: auth.ActionWrite},
		"/api/v1/profiles/subscribe":   {Resource: auth.ResourceProfile, Action: auth.ActionWrite},
		"/api/v1/profiles/unsubscribe": {Resource: auth.ResourceProfile, Action: auth.ActionWrite},
		"/api/v1/cloud/plan":           {Resource: auth.ResourceCloud, Action: auth.ActionRead},
		"/api/v1/cloud/deploy":         {Resource: auth.ResourceCloud, Action: auth.ActionWrite},
		"/api/v1/cloud/destroy":        {Resource: auth.ResourceCloud, Action: auth.ActionDelete},
		"/api/v1/cloud/status":         {Resource: auth.ResourceCloud, Action: auth.ActionRead},
		"/api/v1/plugins":              {Resource: auth.ResourcePlugin, Action: auth.ActionRead},
		"/api/v1/plugins/execute":      {Resource: auth.ResourcePlugin, Action: auth.ActionWrite},
		"/api/v1/ai/enrich/stream":     {Resource: auth.ResourceAI, Action: auth.ActionWrite},
		"/api/v1/ai/explain/stream":    {Resource: auth.ResourceAI, Action: auth.ActionRead},
		"/api/v1/ai/compile/stream":    {Resource: auth.ResourceAI, Action: auth.ActionWrite},
		"/api/v1/config/schema":        {Resource: auth.ResourceConfig, Action: auth.ActionRead},
		"/api/v1/schemas":              {Resource: auth.ResourceConfig, Action: auth.ActionRead},
		"/api/v1/version":              {Resource: auth.ResourceAdmin, Action: auth.ActionRead},
		"/api/v1/health":               {Resource: auth.ResourceAdmin, Action: auth.ActionRead},
		"/api/v1/tenants":              {Resource: auth.ResourceAdmin, Action: auth.ActionRead},
		"/api/v1/tenants/":             {Resource: auth.ResourceAdmin, Action: auth.ActionRead},
	}
}

// SetWebSocketServer attaches a WebSocket server to the API server.
func (s *Server) SetWebSocketServer(ws *naeosws.Server) {
	s.wsServer = ws
}

// SetDatabase attaches a database to the API server for persistence.
func (s *Server) SetDatabase(db database.Database) {
	s.db = db
}

// SetAuthManager attaches an auth manager for RBAC enforcement.
func (s *Server) SetAuthManager(m *auth.Manager) {
	s.authManager = m
}

// SetWorkspace attaches a workspace manager for multi-tenant isolation.
func (s *Server) SetWorkspace(w *multitenant.Workspace) {
	s.tenantWorkspace = w
}

// SetRoutePermission sets the RBAC resource and action required for a route path.
func (s *Server) SetRoutePermission(path, resource, action string) {
	if s.routePerms == nil {
		s.routePerms = make(map[string]auth.RoutePermission)
	}
	s.routePerms[path] = auth.RoutePermission{Resource: resource, Action: action}
}

// SetPipelineObserver attaches a pipeline observer, chaining with metrics recording.
func (s *Server) SetPipelineObserver(obs pipeline.PipelineObserver) {
	metricsObs := monitoring.NewMetricsObserver(s.metrics)
	s.pipelineObserver = pipeline.ChainObservers(obs, metricsObs)
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

		var rateLimiter *RateLimiter
		var rateLimitID string

		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			s.apiKeysMu.RLock()
			limiter, exists := s.APIKeys[apiKey]
			s.apiKeysMu.RUnlock()
			if exists {
				rateLimiter = limiter
				rateLimitID = apiKey
			} else {
				rateLimiter = s.Limiter
				rateLimitID = clientID
			}
		} else {
			rateLimiter = s.Limiter
			rateLimitID = clientID
		}

		if !rateLimiter.Allow(rateLimitID) {
			s.writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimiter.Rate()))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rateLimiter.Remaining(rateLimitID)))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(rateLimiter.ResetTime(rateLimitID).Unix(), 10))

		// CORS
		origin := r.Header.Get("Origin")
		if s.CORS != nil && originAllowed(origin, s.CORS.AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if s.CORS.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
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

		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Auth
		var userID string
		if s.Auth.Enabled && r.URL.Path != "/api/v1/health" {
			token := r.Header.Get("Authorization")
			if token == "" {
				s.writeError(w, http.StatusUnauthorized, "authorization required")
				return
			}
			token = strings.TrimPrefix(token, "Bearer ")
			if s.jwt != nil {
				claims, err := s.jwt.Validate(token)
				if err != nil {
					s.writeError(w, http.StatusUnauthorized, "invalid token: "+err.Error())
					return
				}
				userID = claims.Sub
			}

			// RBAC check
			if s.authManager != nil {
				perm, ok := s.routePerms[r.URL.Path]
				if ok {
					var user *auth.User
					if userID != "" {
						var exists bool
						user, exists = s.authManager.GetUser(userID)
						if !exists {
							s.writeError(w, http.StatusForbidden, "user not found")
							return
						}
					}
					if user != nil && !s.authManager.RBAC().HasPermission(user, perm.Resource, perm.Action) {
						s.writeError(w, http.StatusForbidden, "insufficient permissions")
						return
					}
				}
			}
		}

		if userID != "" {
			r = r.WithContext(context.WithValue(r.Context(), UserContextKey, userID))

			// Resolve tenant from user metadata (X-Tenant-ID header or user attribute)
			if s.tenantWorkspace != nil {
				tenantID := r.Header.Get("X-Tenant-ID")
				if tenantID == "" {
					if user, exists := s.authManager.GetUser(userID); exists && len(user.Roles) > 0 {
						tenantID = user.Roles[0]
					}
				}
				if tenantID != "" {
					if _, err := s.tenantWorkspace.GetTenant(tenantID); err == nil {
						r = r.WithContext(context.WithValue(r.Context(), TenantContextKey, tenantID))
					}
				}
			}
		}

		// Tenant rate limit check
		if s.TenantLimiter != nil {
			tenantID := ""
			if v, ok := r.Context().Value(TenantContextKey).(string); ok {
				tenantID = v
			}
			if tenantID != "" {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(s.TenantLimiter.Rate()))
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(s.TenantLimiter.Remaining(tenantID)))
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(s.TenantLimiter.ResetTime(tenantID).Unix(), 10))
				if !s.TenantLimiter.Allow(tenantID) {
					s.writeError(w, http.StatusTooManyRequests, "tenant rate limit exceeded")
					return
				}
			}
		}

		slog.Info("request", "method", r.Method, "path", r.URL.Path, "request_id", requestID) //nolint:gosec

		// Capture status for audit logging
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		handler(rec, r)

		// Audit trail
		if s.auditor != nil {
			action := mapMethodToAction(r.Method)
			_ = s.auditor.Log(audit.AuditEvent{
				ID:        GenerateRequestID(),
				Timestamp: time.Now(),
				UserID:    userID,
				Action:    action,
				Resource:  r.URL.Path,
				Status:    auditStatusFromHTTP(rec.status),
				IP:        r.RemoteAddr,
				UserAgent: r.UserAgent(),
				Metadata: map[string]string{
					"method":     r.Method,
					"request_id": requestID,
				},
			})
		}
	}
}

func mapMethodToAction(method string) string {
	switch method {
	case http.MethodGet:
		return "read"
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "other"
	}
}

func auditStatusFromHTTP(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "success"
	case status >= 400 && status < 500:
		return "denied"
	case status >= 500:
		return "error"
	default:
		return "unknown"
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
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: status >= 200 && status < 300,
		Data:    data,
	})
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{
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

	allTargets := s.compiler.Targets()
	if len(allTargets) == 0 {
		s.writeError(w, http.StatusServiceUnavailable, "no compiler targets available; check compiler configuration")
		return
	}

	var targets []compiler.Target
	if req.Target != "" {
		requested := compiler.Target(req.Target)
		found := false
		for _, t := range allTargets {
			if t == requested {
				found = true
				break
			}
		}
		if !found {
			s.writeError(w, http.StatusBadRequest, "unknown target: "+req.Target)
			return
		}
		targets = []compiler.Target{requested}
	} else {
		targets = allTargets
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"compiled": true,
		"targets":  targets,
		"bundle":   b.ToMarkdown(),
		"project":  doc.Project,
		"modules":  len(doc.Modules),
		"services": len(doc.Services),
	})
}

type vizTreeNode struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Children []*vizTreeNode `json:"children,omitempty"`
	Props    map[string]any `json:"props,omitempty"`
}

type vizEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

func (s *Server) handleSpecVisualize(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var req struct {
		Spec string `json:"spec"`
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

	root := &vizTreeNode{
		Name: doc.Project,
		Type: "project",
		Props: map[string]any{
			"project": doc.Project,
		},
	}

	var edges []vizEdge

	// Modules
	if len(doc.Modules) > 0 {
		modNode := &vizTreeNode{Name: "Modules", Type: "group"}
		for _, m := range doc.Modules {
			child := &vizTreeNode{
				Name: m.Name,
				Type: "module",
				Props: map[string]any{
					"path":        m.Path,
					"description": m.Description,
				},
			}
			modNode.Children = append(modNode.Children, child)
			for _, dep := range m.Dependencies {
				edges = append(edges, vizEdge{
					From: m.Name,
					To:   dep,
					Type: "depends_on",
				})
			}
		}
		root.Children = append(root.Children, modNode)
	}

	// Services
	if len(doc.Services) > 0 {
		svcNode := &vizTreeNode{Name: "Services", Type: "group"}
		for _, svc := range doc.Services {
			child := &vizTreeNode{
				Name: svc.Name,
				Type: "service",
				Props: map[string]any{
					"kind":        svc.Kind,
					"port":        svc.Port,
					"description": svc.Description,
				},
			}
			if len(svc.Endpoints) > 0 {
				epNode := &vizTreeNode{Name: "Endpoints", Type: "group"}
				for _, ep := range svc.Endpoints {
					epNode.Children = append(epNode.Children, &vizTreeNode{
						Name: ep.Path,
						Type: "endpoint",
						Props: map[string]any{
							"method": ep.Method,
							"action": ep.Action,
						},
					})
				}
				child.Children = append(child.Children, epNode)
			}
			svcNode.Children = append(svcNode.Children, child)
		}
		root.Children = append(root.Children, svcNode)
	}

	// Architecture
	if doc.Architecture != nil {
		archNode := &vizTreeNode{
			Name: "Architecture",
			Type: "architecture",
			Props: map[string]any{
				"pattern":     doc.Architecture.Pattern,
				"description": doc.Architecture.Description,
			},
		}
		if len(doc.Architecture.Principles) > 0 {
			pNode := &vizTreeNode{Name: "Principles", Type: "group"}
			for _, p := range doc.Architecture.Principles {
				pNode.Children = append(pNode.Children, &vizTreeNode{
					Name: p, Type: "principle",
				})
			}
			archNode.Children = append(archNode.Children, pNode)
		}
		root.Children = append(root.Children, archNode)
	}

	// Deployment
	if doc.Deployment != nil {
		depNode := &vizTreeNode{
			Name: "Deployment",
			Type: "deployment",
			Props: map[string]any{
				"strategy": doc.Deployment.Strategy,
			},
		}
		if len(doc.Deployment.Environments) > 0 {
			envNode := &vizTreeNode{Name: "Environments", Type: "group"}
			for _, e := range doc.Deployment.Environments {
				envNode.Children = append(envNode.Children, &vizTreeNode{Name: e, Type: "environment"})
			}
			depNode.Children = append(depNode.Children, envNode)
		}
		root.Children = append(root.Children, depNode)
	}

	// Testing
	if doc.Testing != nil {
		root.Children = append(root.Children, &vizTreeNode{
			Name: "Testing",
			Type: "testing",
			Props: map[string]any{
				"strategy": doc.Testing.Strategy,
				"coverage": doc.Testing.Coverage,
			},
		})
	}

	// Generation
	if doc.Generation != nil {
		genNode := &vizTreeNode{
			Name: "Generation",
			Type: "generation",
			Props: map[string]any{
				"output_dir": doc.Generation.OutputDir,
				"module_dir": doc.Generation.ModuleDir,
			},
		}
		if len(doc.Generation.Languages) > 0 {
			langNode := &vizTreeNode{Name: "Languages", Type: "group"}
			for _, l := range doc.Generation.Languages {
				langNode.Children = append(langNode.Children, &vizTreeNode{Name: l, Type: "language"})
			}
			genNode.Children = append(genNode.Children, langNode)
		}
		root.Children = append(root.Children, genNode)
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"tree":  root,
		"edges": edges,
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
		startTime := time.Now()

		if s.pipelineObserver != nil {
			s.pipelineObserver.OnPipelineStart(jobID)
		}

		var modNames []string
		for _, m := range doc.Modules {
			modNames = append(modNames, m.Name)
		}

		tenantID, _ := r.Context().Value(TenantContextKey).(string)

		run := pipelineRun{
			ID:          jobID,
			Status:      "completed",
			Project:     doc.Project,
			Target:      req.Target,
			Modules:     len(doc.Modules),
			Services:    len(doc.Services),
			ModuleNames: modNames,
			CreatedAt:   startTime.Format(time.RFC3339),
			TenantID:    tenantID,
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					run.Status = "failed"
					run.Error = fmt.Sprintf("panic: %v", r)
				}
				run.CompletedAt = time.Now().Format(time.RFC3339)
				run.Duration = time.Since(startTime).Round(time.Millisecond).String()

				if s.pipelineObserver != nil {
					switch run.Status {
					case "completed":
						s.pipelineObserver.OnPipelineComplete(jobID, run.Modules+run.Services, run.Duration)
					case "failed":
						s.pipelineObserver.OnPipelineFailed(jobID, run.Error)
					}
				}

				s.pipelinesMu.Lock()
				s.pipelines = append(s.pipelines, run)
				s.pipelinesMu.Unlock()

				s.savePipelines()

				if s.db != nil {
					if _, err := s.db.Exec("INSERT INTO pipeline_runs (id, status, project, modules, services, created_at) VALUES (?, ?, ?, ?, ?, ?)",
						run.ID, run.Status, run.Project, run.Modules, run.Services, run.CreatedAt); err != nil {
						slog.Warn("failed to persist pipeline run", "error", err)
					}
				}

				s.jobsMu.Lock()
				job.Status = run.Status
				job.Error = run.Error
				s.jobsMu.Unlock()
			}()
			_ = s.bundle.GenerateFromSpec(doc)
		}()
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
	s.pipelinesMu.RLock()
	var lastRun *pipelineRun
	if len(s.pipelines) > 0 {
		last := s.pipelines[len(s.pipelines)-1]
		lastRun = &last
	}
	total := len(s.pipelines)
	s.pipelinesMu.RUnlock()
	status := "idle"
	if lastRun != nil && lastRun.Status == "running" {
		status = "running"
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status":   status,
		"total":    total,
		"last_run": lastRun,
	})
}

func (s *Server) handleArtifacts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		list := s.store.List()
		offset, limit := parsePagination(r)
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
		if err := s.store.WriteToDisk(); err != nil {
			slog.Error("failed to persist artifacts to disk", "error", err)
		}
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
		Provider  string           `json:"provider"`
		Project   string           `json:"project"`
		Region    string           `json:"region"`
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
		Provider:  cloud.CloudProvider(req.Provider),
		Region:    req.Region,
		Project:   req.Project,
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
		Provider  string           `json:"provider"`
		Project   string           `json:"project"`
		Region    string           `json:"region"`
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
		Provider:  cloud.CloudProvider(req.Provider),
		Region:    req.Region,
		Project:   req.Project,
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
		Provider  string           `json:"provider"`
		Project   string           `json:"project"`
		Region    string           `json:"region"`
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
		Provider:  cloud.CloudProvider(req.Provider),
		Region:    req.Region,
		Project:   req.Project,
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

	offset, limit := parsePagination(r)
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
		offset, limit := parsePagination(r)
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

	result, err := s.plugins.Execute(r.Context(), req.Name, req.Action, req.Params)
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

func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	query := r.URL.Query().Get("q")
	var result []profiles.Profile
	if query != "" {
		result = s.profiles.Search(query)
	} else {
		result = s.profiles.List()
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"profiles": result,
		"count":    len(result),
	})
}

func (s *Server) handleProfilePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var p profiles.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if p.ID == "" || p.Name == "" {
		s.writeError(w, http.StatusBadRequest, "id and name are required")
		return
	}
	s.profiles.Register(&p)
	s.writeJSON(w, http.StatusCreated, map[string]any{
		"message": "profile published",
		"id":      p.ID,
	})
}

func (s *Server) handleProfileSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Profiles []profiles.Profile `json:"profiles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	for _, p := range req.Profiles {
		s.profiles.Register(&p)
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"message": "profiles synced",
		"count":   len(req.Profiles),
	})
}

func (s *Server) handleProfileSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		RegistryURL string `json:"registry_url"`
		APIKey      string `json:"api_key,omitempty"`
		Interval    string `json:"interval,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RegistryURL == "" {
		s.writeError(w, http.StatusBadRequest, "registry_url is required")
		return
	}

	s.profileSubMu.Lock()
	defer s.profileSubMu.Unlock()

	if _, exists := s.profileSubs[req.RegistryURL]; exists {
		s.writeJSON(w, http.StatusConflict, map[string]any{
			"message":      "already subscribed to this registry",
			"registry_url": req.RegistryURL,
		})
		return
	}

	interval := 5 * time.Minute
	if req.Interval != "" {
		if d, err := time.ParseDuration(req.Interval); err == nil && d > 0 {
			interval = d
		}
	}

	reg := profiles.RemoteRegistry{
		URL:    req.RegistryURL,
		APIKey: req.APIKey,
	}
	sub := s.profiles.Subscribe(reg, interval)
	s.profileSubs[req.RegistryURL] = sub

	s.writeJSON(w, http.StatusOK, map[string]any{
		"message":      "subscription started",
		"registry_url": req.RegistryURL,
		"interval":     interval.String(),
	})
}

func (s *Server) handleProfileUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		RegistryURL string `json:"registry_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RegistryURL == "" {
		s.writeError(w, http.StatusBadRequest, "registry_url is required")
		return
	}

	s.profileSubMu.Lock()
	defer s.profileSubMu.Unlock()

	sub, exists := s.profileSubs[req.RegistryURL]
	if !exists {
		s.writeJSON(w, http.StatusNotFound, map[string]any{
			"message":      "no active subscription for this registry",
			"registry_url": req.RegistryURL,
		})
		return
	}

	sub.Stop()
	delete(s.profileSubs, req.RegistryURL)

	s.writeJSON(w, http.StatusOK, map[string]any{
		"message":      "subscription stopped",
		"registry_url": req.RegistryURL,
	})
}

func (s *Server) handleProfileByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/profiles/")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, errors.New(errors.ErrValidation, "profile id is required").Error())
		return
	}
	switch r.Method {
	case "GET":
		p, ok := s.profiles.Get(id)
		if !ok {
			s.writeError(w, http.StatusNotFound, fmt.Sprintf("profile %s not found", id))
			return
		}
		s.writeJSON(w, http.StatusOK, p)
	case "DELETE":
		s.writeError(w, http.StatusMethodNotAllowed, "delete not supported, use sync to update")
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleAIEnrichStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}

	var req struct {
		Spec     string         `json:"spec"`
		Model    string         `json:"model,omitempty"`
		Provider ai.LLMProvider `json:"provider,omitempty"`
		APIKey   string         `json:"api_key,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}
	if req.Spec == "" {
		s.writeError(w, http.StatusBadRequest, "spec is required")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	cfg := buildLLMConfig(req.Provider, req.Model, req.APIKey)

	svc := ai.NewLLMService(cfg)
	if err := svc.StreamEnrichSpec(r.Context(), req.Spec, w); err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\":%q}\n\n", err.Error())
	}
}

func (s *Server) handleAIExplainStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}

	var req struct {
		Spec         string         `json:"spec"`
		Architecture string         `json:"architecture"`
		Model        string         `json:"model,omitempty"`
		Provider     ai.LLMProvider `json:"provider,omitempty"`
		APIKey       string         `json:"api_key,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}
	if req.Spec == "" {
		s.writeError(w, http.StatusBadRequest, "spec is required")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	cfg := buildLLMConfig(req.Provider, req.Model, req.APIKey)

	svc := ai.NewLLMService(cfg)
	if err := svc.StreamExplainArchitecture(r.Context(), req.Spec, req.Architecture, w); err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\":%q}\n\n", err.Error())
	}
}

func (s *Server) handleAICompileStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}

	var req struct {
		Spec     string         `json:"spec"`
		Target   string         `json:"target"`
		Model    string         `json:"model,omitempty"`
		Provider ai.LLMProvider `json:"provider,omitempty"`
		APIKey   string         `json:"api_key,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}
	if req.Spec == "" {
		s.writeError(w, http.StatusBadRequest, "spec is required")
		return
	}
	if req.Target == "" {
		req.Target = "opencode"
	}

	doc, err := s.parser.Parse(req.Spec)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "parse error: "+err.Error())
		return
	}
	b := s.bundle.GenerateFromSpec(doc)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	cfg := buildLLMConfig(req.Provider, req.Model, req.APIKey)
	svc := ai.NewLLMService(cfg)
	if err := svc.StreamCompileSpec(r.Context(), req.Target, b.ToMarkdown(), w); err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\":%q}\n\n", err.Error())
	}
}

func buildLLMConfig(provider ai.LLMProvider, model, apiKey string) ai.LLMConfig {
	if provider == "" {
		provider = ai.ProviderOpenAI
	} else {
		switch provider {
		case ai.ProviderOpenAI, ai.ProviderAnthropic, ai.ProviderOllama:
		default:
			slog.Warn("unrecognized LLM provider, falling back to openai", "provider", provider)
			provider = ai.ProviderOpenAI
		}
	}
	if model == "" {
		switch provider {
		case ai.ProviderAnthropic:
			model = "claude-3-haiku-20240307"
		case ai.ProviderOllama:
			model = "llama3.2"
		default:
			model = "gpt-4o-mini"
		}
	}
	if apiKey == "" {
		if k := os.Getenv("NAEOS_LLM_API_KEY"); k != "" {
			apiKey = k
		}
	}
	if apiKey == "" {
		switch provider {
		case ai.ProviderAnthropic:
			if k := os.Getenv("ANTHROPIC_API_KEY"); k != "" {
				apiKey = k
			}
		case ai.ProviderOllama:
		default:
			if k := os.Getenv("OPENAI_API_KEY"); k != "" {
				apiKey = k
			}
		}
	}
	return ai.LLMConfig{
		Provider: provider,
		Model:    model,
		APIKey:   apiKey,
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
			"name":       map[string]string{"type": "string", "description": "project name"},
			"version":    map[string]string{"type": "string", "description": "project version"},
			"output_dir": map[string]string{"type": "string", "description": "output directory"},
			"mode":       map[string]string{"type": "string", "description": "pipeline mode"},
			"verbose":    map[string]string{"type": "boolean", "description": "verbose output"},
		},
		"required": []string{"name"},
	})
}

func (s *Server) handlePipelines(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	statusFilter := q.Get("status")
	projectSearch := strings.ToLower(q.Get("search"))
	offset, limit := parsePagination(r)
	if limit <= 0 {
		limit = 20
	}

	tenantFilter, _ := r.Context().Value(TenantContextKey).(string)

	s.pipelinesMu.RLock()
	filtered := make([]pipelineRun, 0, len(s.pipelines))
	for _, p := range s.pipelines {
		if statusFilter != "" && p.Status != statusFilter {
			continue
		}
		if projectSearch != "" && !strings.Contains(strings.ToLower(p.Project), projectSearch) {
			continue
		}
		if tenantFilter != "" && p.TenantID != tenantFilter {
			continue
		}
		filtered = append(filtered, p)
	}
	total := len(filtered)

	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	result := make([]pipelineRun, len(filtered[start:end]))
	copy(result, filtered[start:end])
	s.pipelinesMu.RUnlock()

	// Reverse so newest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"pipelines": result,
		"total":     total,
		"page":      offset/limit + 1,
		"limit":     limit,
	})
}

func (s *Server) handleSchemas(w http.ResponseWriter, r *http.Request) {
	tenantFilter, _ := r.Context().Value(TenantContextKey).(string)

	switch r.Method {
	case "GET":
		names := s.schemas.List()
		result := make([]map[string]any, 0, len(names))
		for _, name := range names {
			if tenantFilter != "" {
				entry, err := s.schemas.Get(name, "")
				if err != nil || entry.TenantID != tenantFilter {
					continue
				}
			}
			versions, err := s.schemas.Versions(name)
			if err != nil {
				continue
			}
			result = append(result, map[string]any{
				"name":     name,
				"versions": versions,
			})
		}
		s.writeJSON(w, http.StatusOK, map[string]any{
			"schemas": result,
			"count":   len(result),
		})
	case "POST":
		var req struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Schema  string `json:"schema"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" || req.Version == "" || req.Schema == "" {
			s.writeError(w, http.StatusBadRequest, "name, version, and schema are required")
			return
		}
		if err := s.schemas.Register(req.Name, req.Version, req.Schema); err != nil {
			s.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		entry, err := s.schemas.Get(req.Name, req.Version)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.writeJSON(w, http.StatusCreated, entry)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSchemaByPath(w http.ResponseWriter, r *http.Request) {
	tenantFilter, _ := r.Context().Value(TenantContextKey).(string)

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/schemas/")
	parts := strings.SplitN(path, "/", 2)
	name := parts[0]
	if name == "" {
		s.writeError(w, http.StatusBadRequest, "schema name is required")
		return
	}
	version := ""
	if len(parts) > 1 {
		version = parts[1]
	}
	switch r.Method {
	case "GET":
		entry, err := s.schemas.Get(name, version)
		if err != nil {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if tenantFilter != "" && entry.TenantID != tenantFilter {
			s.writeError(w, http.StatusNotFound, "schema not found")
			return
		}
		if version == "" {
			versions, err := s.schemas.Versions(name)
			if err != nil {
				s.writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			s.writeJSON(w, http.StatusOK, map[string]any{
				"schema":   entry,
				"versions": versions,
			})
		} else {
			s.writeJSON(w, http.StatusOK, entry)
		}
	case "DELETE":
		if version == "" {
			s.writeError(w, http.StatusBadRequest, "version is required for deletion")
			return
		}
		if err := s.schemas.Delete(name, version); err != nil {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
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
	_ = json.NewEncoder(w).Encode(doc)
}

func (s *Server) handleTenants(w http.ResponseWriter, r *http.Request) {
	if s.tenantWorkspace == nil {
		s.writeError(w, http.StatusNotFound, "tenant workspace not configured")
		return
	}
	switch r.Method {
	case "GET":
		tenants := s.tenantWorkspace.ListTenants()
		s.writeJSON(w, http.StatusOK, map[string]any{
			"tenants": tenants,
			"count":   len(tenants),
		})
	case "POST":
		var req struct {
			Name string `json:"name"`
			Plan string `json:"plan"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" {
			s.writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		id := fmt.Sprintf("tenant-%d", time.Now().UnixNano())
		tenant, err := s.tenantWorkspace.CreateTenant(id, req.Name, req.Plan)
		if err != nil {
			s.writeError(w, http.StatusConflict, err.Error())
			return
		}
		s.writeJSON(w, http.StatusCreated, tenant)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleTenantByID(w http.ResponseWriter, r *http.Request) {
	if s.tenantWorkspace == nil {
		s.writeError(w, http.StatusNotFound, "tenant workspace not configured")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/tenants/")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "tenant id is required")
		return
	}
	switch r.Method {
	case "GET":
		tenant, err := s.tenantWorkspace.GetTenant(id)
		if err != nil {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeJSON(w, http.StatusOK, tenant)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(s.metricsRegistry.FormatPrometheus()))
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

func parsePagination(r *http.Request) (offset, limit int) {
	limit = 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	offset = 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	} else if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			offset = (v - 1) * limit
		}
	}
	return
}

var startTime = time.Now()

// Start begins listening for HTTP requests and handles graceful shutdown.
func (s *Server) Start() error {
	mw := monitoring.MetricsMiddleware(s.metrics)
	wrappedHandler := mw(s.loggingMiddleware(s.handlerWithMiddleware(s.Router.ServeHTTP)))

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
		_ = s.server.Shutdown(ctx)
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
