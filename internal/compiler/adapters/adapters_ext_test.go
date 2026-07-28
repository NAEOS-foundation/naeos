package adapters

import (
	"context"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/ai"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/api"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/component"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/docs"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/infrastructure"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/storage"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
	"github.com/NAEOS-foundation/naeos/internal/testutil"
)

func fullNEIR() *model.NEIR {
	return &model.NEIR{
		Project: &project.Project{
			Name:        "test-project",
			Version:     "1.0.0",
			Description: "A test project for coverage",
			License:     "Apache-2.0",
			Authors:     []string{"dev1", "dev2"},
			Repository:  "https://github.com/test/repo",
		},
		Architecture: &architecture.Architecture{
			Pattern:    architecture.PatternLayered,
			Principles: []string{"separation-of-concerns", "single-responsibility"},
			Layers: []architecture.Layer{
				{Name: "api", Description: "API layer", Modules: []string{"http"}},
				{Name: "domain", Description: "Domain layer", Modules: []string{"core"}},
			},
		},
		Modules: []module.Module{
			{Name: "core", Path: "internal/core", Description: "Core logic", Dependencies: []string{}},
			{Name: "http", Path: "internal/http", Description: "HTTP handlers", Dependencies: []string{"core"}},
		},
		Services: []service.Service{
			{
				Name: "api-gateway",
				Kind: service.KindHTTP,
				Port: 8080,
				Endpoints: []service.Endpoint{
					{Method: "GET", Path: "/health", Action: "healthcheck"},
					{Method: "POST", Path: "/api/v1/users", Action: "createUser"},
				},
			},
		},
		Components: []component.Component{
			{Name: "UserService", Kind: component.KindService, Module: "core", Description: "User management"},
			{Name: "UserRepo", Kind: component.KindRepository, Module: "core"},
		},
		APIs: []api.API{
			{
				Name:     "users-api",
				Version:  "v1",
				Protocol: api.ProtocolHTTP,
				Endpoints: []api.APIEndpoint{
					{Method: "GET", Path: "/users", Summary: "List users"},
					{Method: "POST", Path: "/users", Summary: "Create user"},
				},
			},
		},
		Storage: []storage.Storage{
			{Name: "users-db", Type: storage.TypeSQL, Provider: "postgres", Connection: "postgres://localhost:5432/users"},
		},
		Infrastructure: &infrastructure.Infrastructure{
			Provider: infrastructure.ProviderAWS,
			Region:   "us-east-1",
			Resources: []infrastructure.Resource{
				{Name: "main-vpc", Kind: "vpc", Type: "aws_vpc"},
			},
		},
		Security: &security.Security{
			Authentication: &security.Authentication{Method: "jwt", Provider: "auth0"},
			Authorization:  &security.Authorization{Model: "rbac", Roles: []string{"admin", "user"}},
			Encryption:     &security.Encryption{InTransit: true, AtRest: true, Algorithm: "aes-256"},
		},
		AI: &ai.AI{
			Models: []ai.Model{
				{Name: "code-assistant", Kind: "llm", Version: "1.0"},
			},
		},
		Documentation: &docs.Documentation{
			ADRs: []docs.Doc{{Title: "ADR-001: Use layered architecture", Kind: docs.KindADR}},
			RFCs: []docs.Doc{{Title: "RFC-001: API versioning", Kind: docs.KindRFC}},
		},
		Deployment: &deployment.Deployment{
			Strategy: deployment.StrategyRolling,
			Environments: []deployment.Environment{
				{Name: "staging", Kind: "staging"},
			},
		},
		Testing: &testingmodel.Testing{
			Strategy:   testingmodel.StrategyUnit,
			Frameworks: []string{"go-test", "testify"},
		},
	}
}

func TestAIAdapterCompileContextError(t *testing.T) {
	t.Parallel()
	llm := newMockLLMService()
	a := NewAICompilerAdapter(compiler.TargetClaude, llm)
	aiA, ok := a.(*AICompilerAdapter)
	if !ok {
		t.Fatal("expected AICompilerAdapter")
	}
	_, err := aiA.CompileContext(context.Background(), &model.NEIR{})
	if err == nil {
		t.Error("expected error from mock LLM")
	}
}

func TestAIAdapterCompileTimeout(t *testing.T) {
	t.Parallel()
	llm := newMockLLMService()
	a := NewAICompilerAdapter(compiler.TargetClaude, llm)
	_, err := a.Compile(&model.NEIR{})
	if err == nil {
		t.Error("expected error from Compile (timeout or LLM failure)")
	}
}

func TestParseCompiledFilesEmptyArray(t *testing.T) {
	t.Parallel()
	_, err := parseCompiledFiles(`[]`)
	if err == nil {
		t.Error("expected error for empty array")
	}
}

func TestBuildNEIRContextRich(t *testing.T) {
	t.Parallel()
	neir := fullNEIR()
	ctx := buildNEIRContext(neir)
	checks := []string{
		"test-project", "1.0.0", "layered", "core", "api-gateway",
		"UserService", "users-api", "jwt", "rbac",
		"rolling", "unit",
	}
	for _, c := range checks {
		if !testutil.Contains(ctx, c) {
			t.Errorf("expected %q in context", c)
		}
	}
}

func TestBuildNEIRContextSecurityNoAuth(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Security: &security.Security{
			Encryption: &security.Encryption{InTransit: true},
		},
	}
	ctx := buildNEIRContext(neir)
	if !testutil.Contains(ctx, "Security") {
		t.Error("expected Security section")
	}
}

func TestClaudeAdapterFullNEIR(t *testing.T) {
	t.Parallel()
	a := NewClaudeAdapter(nil)
	out, err := a.Compile(fullNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) < 1 {
		t.Fatal("expected at least 1 file")
	}
	if out.Target != compiler.TargetClaude {
		t.Errorf("expected claude target, got %s", out.Target)
	}
}

func TestClaudeAdapterFull(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "minimal"},
	}
	a := NewClaudeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) < 1 {
		t.Fatal("expected files")
	}
}

func TestClaudeAdapterWithSecurity(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Security: &security.Security{
			Authentication: &security.Authentication{Method: "oauth2", Provider: "google"},
		},
	}
	a := NewClaudeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if f.Path == "CLAUDE.md" && testutil.Contains(f.Content, "oauth2") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected oauth2 in CLAUDE.md")
	}
}

func TestClaudeAdapterWithArchitecture(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Architecture: &architecture.Architecture{
			Pattern:    architecture.PatternClean,
			Principles: []string{"SRP", "OCP"},
		},
	}
	a := NewClaudeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rulesFile := ""
	for _, f := range out.Files {
		if f.Path == ".claude/rules.md" {
			rulesFile = f.Content
		}
	}
	if rulesFile == "" {
		t.Fatal("expected .claude/rules.md")
	}
	if !testutil.Contains(rulesFile, "clean") {
		t.Error("expected architecture pattern in rules")
	}
}

func TestClaudeAdapterLegacyRulesNoArchitecture(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Modules: []module.Module{
			{Name: "m1", Path: "p1"},
		},
	}
	a := NewClaudeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range out.Files {
		if f.Path == ".claude/rules.md" {
			if !testutil.Contains(f.Content, "m1") {
				t.Error("expected module name in rules")
			}
		}
	}
}

func TestClaudeAdapterContextWithComponents(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Components: []component.Component{
			{Name: "Handler", Kind: component.KindHandler, Module: "web"},
		},
	}
	a := NewClaudeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if testutil.Contains(f.Content, "Handler") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected component in context")
	}
}

func TestOpenCodeAdapterFullNEIR(t *testing.T) {
	t.Parallel()
	a := NewOpenCodeAdapter(nil)
	out, err := a.Compile(fullNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Target != compiler.TargetOpenCode {
		t.Errorf("expected opencode target, got %s", out.Target)
	}
}

func TestOpenCodeAdapterSecurityWithEncryption(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Security: &security.Security{
			Authentication: &security.Authentication{Method: "api-key"},
			Authorization:  &security.Authorization{Model: "abac"},
			Encryption:     &security.Encryption{InTransit: true, AtRest: false},
		},
	}
	a := NewOpenCodeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if f.Path == "AGENTS.md" && testutil.Contains(f.Content, "api-key") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth info in AGENTS.md")
	}
}

func TestOpenCodeAdapterWithInfrastructure(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Infrastructure: &infrastructure.Infrastructure{
			Provider: infrastructure.ProviderGCP,
			Region:   "us-central1",
			Resources: []infrastructure.Resource{
				{Name: "gke-cluster", Kind: "gke"},
			},
		},
	}
	a := NewOpenCodeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if testutil.Contains(f.Content, "gcp") {
			found = true
		}
	}
	if !found {
		t.Error("expected infrastructure info in context")
	}
}

func TestOpenCodeAdapterWithAIModels(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		AI: &ai.AI{
			Models: []ai.Model{{Name: "gpt4", Kind: "llm", Version: "4.0"}},
		},
	}
	a := NewOpenCodeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if testutil.Contains(f.Content, "gpt4") {
			found = true
		}
	}
	if !found {
		t.Error("expected AI model info in context")
	}
}

func TestOpenCodeAdapterWithDocs(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Documentation: &docs.Documentation{
			ADRs: []docs.Doc{{Title: "ADR-1: Tech Stack"}},
			RFCs: []docs.Doc{{Title: "RFC-1: Auth Flow"}},
		},
	}
	a := NewOpenCodeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if testutil.Contains(f.Content, "ADR-1") {
			found = true
		}
	}
	if !found {
		t.Error("expected ADR in context")
	}
}

func TestOpenCodeAdapterWithStorage(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Storage: []storage.Storage{
			{Name: "cache", Type: storage.TypeCache, Provider: "redis"},
		},
	}
	a := NewOpenCodeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if testutil.Contains(f.Content, "cache") {
			found = true
		}
	}
	if !found {
		t.Error("expected storage info in context")
	}
}

func TestOpenCodeAdapterWithDeploymentAndTesting(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Deployment: &deployment.Deployment{Strategy: deployment.StrategyCanary},
		Testing:    &testingmodel.Testing{Strategy: testingmodel.StrategyE2E},
	}
	a := NewOpenCodeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if testutil.Contains(f.Content, "canary") {
			found = true
		}
	}
	if !found {
		t.Error("expected deployment strategy in output")
	}
}

func TestOpenCodeAdapterRulesNoArchitecture(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Modules: []module.Module{{Name: "m1", Path: "p1"}},
	}
	a := NewOpenCodeAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range out.Files {
		if f.Path == ".opencode/rules.md" {
			if !testutil.Contains(f.Content, "m1") {
				t.Error("expected module name in rules")
			}
		}
	}
}

func TestGeminiAdapterFullNEIR(t *testing.T) {
	t.Parallel()
	a := NewGeminiAdapter(nil)
	out, err := a.Compile(fullNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Target != compiler.TargetGemini {
		t.Errorf("expected gemini target, got %s", out.Target)
	}
}

func TestGeminiAdapterContextWithComponentsAndAPIs(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Components: []component.Component{
			{Name: "Ctrl", Kind: component.KindHandler, Module: "web"},
		},
		APIs: []api.API{
			{
				Name:     "rest-api",
				Protocol: api.ProtocolHTTP,
				Endpoints: []api.APIEndpoint{
					{Method: "GET", Path: "/items", Summary: "List items"},
				},
			},
		},
	}
	a := NewGeminiAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if testutil.Contains(f.Content, "Ctrl") {
			found = true
		}
	}
	if !found {
		t.Error("expected component in context")
	}
}

func TestGeminiAdapterNoComponentsNoAPIs(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "empty"},
	}
	a := NewGeminiAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) < 1 {
		t.Fatal("expected files")
	}
}

func TestCodexAdapterFullNEIR(t *testing.T) {
	t.Parallel()
	a := NewCodexAdapter(nil)
	out, err := a.Compile(fullNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Target != compiler.TargetCodex {
		t.Errorf("expected codex target, got %s", out.Target)
	}
}

func TestCodexAdapterWithStorageAndAI(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Storage: []storage.Storage{
			{Name: "db", Type: storage.TypeSQL, Provider: "mysql"},
		},
		AI: &ai.AI{
			Models: []ai.Model{{Name: "embed", Kind: "embedding", Version: "2.0"}},
		},
		Documentation: &docs.Documentation{
			ADRs: []docs.Doc{{Title: "ADR-42"}},
		},
	}
	a := NewCodexAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range out.Files {
		if f.Path == ".codex/context.md" && testutil.Contains(f.Content, "db") {
			found = true
		}
	}
	if !found {
		t.Error("expected storage in context")
	}
}

func TestCodexAdapterContextEmpty(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "empty"},
	}
	a := NewCodexAdapter(nil)
	out, err := a.Compile(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range out.Files {
		if f.Path == ".codex/context.md" {
			if !testutil.Contains(f.Content, "Codex Context") {
				t.Error("expected context header")
			}
		}
	}
}

func TestCopilotAdapterFullNEIR(t *testing.T) {
	t.Parallel()
	a := NewCopilotAdapter(nil)
	out, err := a.Compile(fullNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Target != compiler.TargetCopilot {
		t.Errorf("expected copilot target, got %s", out.Target)
	}
}

func TestCursorAdapterFullNEIR(t *testing.T) {
	t.Parallel()
	a := NewCursorAdapter(nil)
	out, err := a.Compile(fullNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Target != compiler.TargetCursor {
		t.Errorf("expected cursor target, got %s", out.Target)
	}
}

func TestWindsurfAdapterFullNEIR(t *testing.T) {
	t.Parallel()
	a := NewWindsurfAdapter(nil)
	out, err := a.Compile(fullNEIR())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Target != compiler.TargetWindsurf {
		t.Errorf("expected windsurf target, got %s", out.Target)
	}
}
