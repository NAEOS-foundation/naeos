package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
	"github.com/NAEOS-foundation/naeos/internal/compiler"
	contextbundle "github.com/NAEOS-foundation/naeos/internal/context/bundle"
	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/NAEOS-foundation/naeos/internal/version"
)

type PipelineJob struct {
	ID        string         `json:"id"`
	Status    string         `json:"status"`
	StartedAt time.Time      `json:"started_at"`
	EndedAt   *time.Time     `json:"ended_at,omitempty"`
	Artifacts int            `json:"artifacts"`
	Error     string         `json:"error,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type Server struct {
	mux          *http.ServeMux
	compiler     *compiler.Compiler
	bundle       *contextbundle.Generator
	store        *artifacts.Store
	pluginMgr    *pluginhost.Manager
	pipelineJobs map[string]*PipelineJob
	mu           sync.RWMutex
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      any           `json:"id"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type CallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func NewServer(c *compiler.Compiler, bg *contextbundle.Generator) *Server {
	s := &Server{
		mux:          http.NewServeMux(),
		compiler:     c,
		bundle:       bg,
		pipelineJobs: make(map[string]*PipelineJob),
	}
	s.mux.HandleFunc("/mcp", s.handleMCP)
	s.mux.HandleFunc("/health", s.handleHealth)
	return s
}

func (s *Server) SetArtifactStore(store *artifacts.Store) {
	s.store = store
}

func (s *Server) SetPluginManager(mgr *pluginhost.Manager) {
	s.pluginMgr = mgr
}

func (s *Server) TrackPipelineJob(job *PipelineJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipelineJobs[job.ID] = job
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) writeJSONRPCError(w http.ResponseWriter, id any, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		Error:   &JSONRPCError{Code: code, Message: message},
		ID:      id,
	})
}

func (s *Server) HandleMCP(w http.ResponseWriter, r *http.Request) {
	s.handleMCP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSONRPCError(w, nil, -32600, "method not allowed")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		s.writeJSONRPCError(w, nil, -32700, "read body failed")
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeJSONRPCError(w, nil, -32700, "parse error: invalid JSON")
		return
	}

	var resp JSONRPCResponse
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "naeos-mcp",
				"version": version.String(),
			},
		}
	case "tools/list":
		resp.Result = map[string]any{
			"tools": s.listTools(),
		}
	case "tools/call":
		var params struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &JSONRPCError{Code: -32602, Message: "invalid params"}
		} else {
			result, err := s.callTool(params.Name, params.Arguments)
			if err != nil {
				resp.Error = &JSONRPCError{Code: -32000, Message: err.Error()}
			} else {
				resp.Result = result
			}
		}
	default:
		resp.Error = &JSONRPCError{Code: -32601, Message: "method not found"}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) listTools() []Tool {
	return []Tool{
		{
			Name:        "parse_spec",
			Description: "Parse a NAEOS specification and return structured data",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"spec": map[string]any{
						"type":        "string",
						"description": "YAML/JSON specification content",
					},
				},
				"required": []string{"spec"},
			},
		},
		{
			Name:        "validate_spec",
			Description: "Validate a NAEOS specification for errors",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"spec": map[string]any{
						"type":        "string",
						"description": "YAML/JSON specification content",
					},
				},
				"required": []string{"spec"},
			},
		},
		{
			Name:        "generate_context",
			Description: "Generate an AI context bundle from a specification",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"spec": map[string]any{
						"type":        "string",
						"description": "YAML/JSON specification content",
					},
					"format": map[string]any{
						"type":        "string",
						"description": "Output format: markdown, plain",
						"default":     "markdown",
					},
				},
				"required": []string{"spec"},
			},
		},
		{
			Name:        "compile_spec",
			Description: "Compile a specification into AI instruction sets",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"spec": map[string]any{
						"type":        "string",
						"description": "YAML/JSON specification content",
					},
					"target": map[string]any{
						"type":        "string",
						"description": "Target tool: copilot, claude, cursor, gemini, codex, opencode",
						"default":     "copilot",
					},
				},
				"required": []string{"spec"},
			},
		},
		{
			Name:        "explain_concept",
			Description: "Explain a NAEOS concept (pipeline, neir, spec, profile, kernel, etc.)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"concept": map[string]any{
						"type":        "string",
						"description": "Concept to explain",
					},
				},
				"required": []string{"concept"},
			},
		},
		{
			Name:        "list_artifacts",
			Description: "List all generated artifacts from the artifact store",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "get_pipeline_status",
			Description: "Get the status of a pipeline job by ID",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"job_id": map[string]any{
						"type":        "string",
						"description": "Pipeline job ID",
					},
				},
				"required": []string{"job_id"},
			},
		},
		{
			Name:        "export_terraform",
			Description: "Export Terraform HCL configuration for a given specification",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"spec": map[string]any{
						"type":        "string",
						"description": "YAML/JSON specification content",
					},
				},
				"required": []string{"spec"},
			},
		},
		{
			Name:        "list_plugins",
			Description: "List all installed plugins and their status",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}

func (s *Server) callTool(name string, args map[string]any) (*CallResult, error) {
	spec, _ := args["spec"].(string)

	switch name {
	case "parse_spec":
		if spec == "" {
			return nil, fmt.Errorf("spec is required")
		}
		p := parser.NewParser()
		doc, err := p.Parse(spec)
		if err != nil {
			return &CallResult{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Parse error: %v", err)}},
				IsError: true,
			}, nil
		}
		result := fmt.Sprintf("Project: %s\nModules: %d\nServices: %d",
			doc.Project, len(doc.Modules), len(doc.Services))
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: result}},
		}, nil

	case "validate_spec":
		if spec == "" {
			return nil, fmt.Errorf("spec is required")
		}
		p := parser.NewParser()
		doc, err := p.Parse(spec)
		if err != nil {
			return &CallResult{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Validation error: %v", err)}},
				IsError: true,
			}, nil
		}
		v := parser.NewSpecValidator()
		result := v.Validate(doc.Data)
		status := "PASS"
		if !result.Valid {
			status = "FAIL"
		}
		text := fmt.Sprintf("Status: %s\nIssues: %d\nWarnings: %d",
			status, len(result.Issues), len(result.Warnings))
		for _, issue := range result.Issues {
			text += fmt.Sprintf("\n  [%s] %s: %s", issue.Severity, issue.Rule, issue.Message)
		}
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: text}},
		}, nil

	case "generate_context":
		if spec == "" {
			return nil, fmt.Errorf("spec is required")
		}
		p := parser.NewParser()
		doc, err := p.Parse(spec)
		if err != nil {
			return nil, fmt.Errorf("parse failed: %w", err)
		}
		format, _ := args["format"].(string)
		if format == "" {
			format = "markdown"
		}
		gen := s.bundle
		b := gen.GenerateFromSpec(doc)
		var text string
		if format == "plain" {
			text = b.ToPlainText()
		} else {
			text = b.ToMarkdown()
		}
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: text}},
		}, nil

	case "compile_spec":
		if spec == "" {
			return nil, fmt.Errorf("spec is required")
		}
		p := parser.NewParser()
		doc, err := p.Parse(spec)
		if err != nil {
			return nil, fmt.Errorf("parse failed: %w", err)
		}
		gen := s.bundle
		b := gen.GenerateFromSpec(doc)
		targetStr, _ := args["target"].(string)
		if targetStr == "" {
			targetStr = "copilot"
		}
		text := fmt.Sprintf("Compiled for target: %s\n\n%s", targetStr, b.ToMarkdown())
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: text}},
		}, nil

	case "explain_concept":
		concept, _ := args["concept"].(string)
		explanation := s.explainConcept(concept)
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: explanation}},
		}, nil

	case "list_artifacts":
		return s.handleListArtifacts()

	case "get_pipeline_status":
		jobID, _ := args["job_id"].(string)
		if jobID == "" {
			return nil, fmt.Errorf("job_id is required")
		}
		return s.handleGetPipelineStatus(jobID)

	case "export_terraform":
		if spec == "" {
			return nil, fmt.Errorf("spec is required")
		}
		return s.handleExportTerraform(spec)

	case "list_plugins":
		return s.handleListPlugins()

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) handleListArtifacts() (*CallResult, error) {
	if s.store == nil {
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: "No artifact store configured"}},
		}, nil
	}
	arts := s.store.List()
	if len(arts) == 0 {
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: "No artifacts found"}},
		}, nil
	}
	type artifactSummary struct {
		ID   string `json:"id"`
		Path string `json:"path"`
		Kind string `json:"kind"`
		Size int64  `json:"size"`
	}
	summaries := make([]artifactSummary, len(arts))
	for i, a := range arts {
		summaries[i] = artifactSummary{
			ID:   a.ID,
			Path: a.Path,
			Kind: string(a.Kind),
			Size: a.Size,
		}
	}
	data, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal artifacts: %w", err)
	}
	return &CallResult{
		Content: []ContentBlock{{Type: "text", Text: string(data)}},
	}, nil
}

func (s *Server) handleGetPipelineStatus(jobID string) (*CallResult, error) {
	s.mu.RLock()
	job, ok := s.pipelineJobs[jobID]
	s.mu.RUnlock()
	if !ok {
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Job %s not found", jobID)}},
			IsError: true,
		}, nil
	}
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal job: %w", err)
	}
	return &CallResult{
		Content: []ContentBlock{{Type: "text", Text: string(data)}},
	}, nil
}

func (s *Server) handleExportTerraform(spec string) (*CallResult, error) {
	p := parser.NewParser()
	doc, err := p.Parse(spec)
	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}

	var b strings.Builder
	b.WriteString("# Auto-generated by NAEOS\n")
	b.WriteString("# Specification: ")
	if doc.Project != "" {
		b.WriteString(doc.Project)
	} else {
		b.WriteString("unnamed")
	}
	b.WriteString("\n\n")

	b.WriteString("terraform {\n")
	b.WriteString("  required_version = \">= 1.0\"\n")
	b.WriteString("}\n\n")

	for _, svc := range doc.Services {
		fmt.Fprintf(&b, "resource \"null_resource\" \"%s\" {\n", svc.Name)
		b.WriteString("  triggers = {\n")
		fmt.Fprintf(&b, "    name    = \"%s\"\n", svc.Name)
		if svc.Kind != "" {
			fmt.Fprintf(&b, "    kind    = \"%s\"\n", svc.Kind)
		}
		b.WriteString("  }\n")
		b.WriteString("}\n\n")
	}

	if len(doc.Services) == 0 {
		b.WriteString("# No services defined in specification\n")
	}

	return &CallResult{
		Content: []ContentBlock{{Type: "text", Text: b.String()}},
	}, nil
}

func (s *Server) handleListPlugins() (*CallResult, error) {
	if s.pluginMgr == nil {
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: "No plugin manager configured"}},
		}, nil
	}
	plugins := s.pluginMgr.List()
	if len(plugins) == 0 {
		return &CallResult{
			Content: []ContentBlock{{Type: "text", Text: "No plugins installed"}},
		}, nil
	}
	type pluginSummary struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		State       string `json:"state"`
	}
	summaries := make([]pluginSummary, len(plugins))
	for i, p := range plugins {
		summaries[i] = pluginSummary{
			Name:        p.Name,
			Version:     p.Version,
			Description: p.Description,
			Enabled:     p.Enabled,
			State:       string(p.State),
		}
	}
	data, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal plugins: %w", err)
	}
	return &CallResult{
		Content: []ContentBlock{{Type: "text", Text: string(data)}},
	}, nil
}

func (s *Server) explainConcept(concept string) string {
	concept = strings.ToLower(strings.TrimSpace(concept))

	concepts := map[string]string{
		"pipeline": "NAEOS Pipeline processes specifications through stages:\n1. Parse YAML/JSON\n2. Normalize data\n3. Resolve cross-references\n4. Build NEIR model\n5. Validate\n6. Schedule tasks\n7. Generate artifacts\n8. Review with governance rules",
		"neir":     "NEIR (Nusantara Engineering Intermediate Representation) is the unified model that represents a project. It contains: Project, Modules, Services, APIs, Architecture, Security, Deployment, Testing, and Generation config.",
		"spec":     "A NAEOS Specification is a YAML/JSON document that defines your project. It includes: project name, modules, services, architecture, deployment, testing, and generation settings.",
		"kernel":   "The NAEOS Kernel manages service registry, event bus (pub/sub), and telemetry collection. It's the core runtime that all pipeline components connect to.",
		"policy":   "Policy rules validate specifications against governance requirements. Operators: exists, not_empty, contains, gt, lt, in. Actions: block, warn.",
		"profile":  "Industry profiles (SaaS, FinTech, Healthcare, etc.) provide pre-configured templates with modules, services, architecture patterns, and security rules.",
		"compiler": "The Compiler transforms NEIR into AI instruction sets for tools like Copilot, Claude Code, Cursor, Gemini CLI, Codex, and OpenCode.",
		"context":  "Context Bundles generate LLM-optimized summaries of your project in markdown or plain text format.",
		"module":   "A Module is a unit of code in your project with a name, path, description, and dependencies on other modules.",
		"service":  "A Service is a runtime component (http, grpc, worker, cli, job) with endpoints, port, and middleware configuration.",
	}

	if explanation, ok := concepts[concept]; ok {
		return explanation
	}

	return fmt.Sprintf("Unknown concept: %s\nAvailable concepts: pipeline, neir, spec, kernel, policy, profile, compiler, context, module, service", concept)
}
