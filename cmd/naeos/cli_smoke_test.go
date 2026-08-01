package main

import (
	"strings"
	"testing"
)

func TestCompletionCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "completion", "--help")
	if err != nil {
		t.Fatalf("completion --help failed: %v", err)
	}
	if !strings.Contains(output, "completion") {
		t.Fatalf("expected completion help, got %q", output)
	}
}

func TestMonitorCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "monitor", "--help")
	if err != nil {
		t.Fatalf("monitor --help failed: %v", err)
	}
	if !strings.Contains(output, "monitor") {
		t.Fatalf("expected monitor help, got %q", output)
	}
}

func TestContextCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "context", "--help")
	if err != nil {
		t.Fatalf("context --help failed: %v", err)
	}
	if !strings.Contains(output, "context") {
		t.Fatalf("expected context help, got %q", output)
	}
}

func TestDocsCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "docs", "--help")
	if err != nil {
		t.Fatalf("docs --help failed: %v", err)
	}
	if !strings.Contains(output, "docs") {
		t.Fatalf("expected docs help, got %q", output)
	}
}

func TestImportCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "import", "--help")
	if err != nil {
		t.Fatalf("import --help failed: %v", err)
	}
	if !strings.Contains(output, "import") {
		t.Fatalf("expected import help, got %q", output)
	}
}

func TestAPICommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "api", "--help")
	if err != nil {
		t.Fatalf("api --help failed: %v", err)
	}
	if !strings.Contains(output, "api") {
		t.Fatalf("expected api help, got %q", output)
	}
}

func TestBenchmarkCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "benchmark", "--help")
	if err != nil {
		t.Fatalf("benchmark --help failed: %v", err)
	}
	if !strings.Contains(output, "benchmark") {
		t.Fatalf("expected benchmark help, got %q", output)
	}
}

func TestBrokerCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "broker", "--help")
	if err != nil {
		t.Fatalf("broker --help failed: %v", err)
	}
	if !strings.Contains(output, "broker") {
		t.Fatalf("expected broker help, got %q", output)
	}
}

func TestDistributedCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "distributed", "--help")
	if err != nil {
		t.Fatalf("distributed --help failed: %v", err)
	}
	if !strings.Contains(output, "distributed") {
		t.Fatalf("expected distributed help, got %q", output)
	}
}

func TestDocsgenCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "docsgen", "--help")
	if err != nil {
		t.Fatalf("docsgen --help failed: %v", err)
	}
	if !strings.Contains(output, "docsgen") {
		t.Fatalf("expected docsgen help, got %q", output)
	}
}

func TestExportCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "export", "--help")
	if err != nil {
		t.Fatalf("export --help failed: %v", err)
	}
	if !strings.Contains(output, "export") {
		t.Fatalf("expected export help, got %q", output)
	}
}

func TestExportComposeHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "export", "compose", "--help")
	if err != nil {
		t.Fatalf("export compose --help failed: %v", err)
	}
	if !strings.Contains(output, "compose") {
		t.Fatalf("expected export compose help, got %q", output)
	}
}

func TestGraphqlCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "graphql", "--help")
	if err != nil {
		t.Fatalf("graphql --help failed: %v", err)
	}
	if !strings.Contains(output, "graphql") {
		t.Fatalf("expected graphql help, got %q", output)
	}
}

func TestHistoryCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "history", "--help")
	if err != nil {
		t.Fatalf("history --help failed: %v", err)
	}
	if !strings.Contains(output, "history") {
		t.Fatalf("expected history help, got %q", output)
	}
}

func TestKernelCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "kernel", "--help")
	if err != nil {
		t.Fatalf("kernel --help failed: %v", err)
	}
	if !strings.Contains(output, "kernel") {
		t.Fatalf("expected kernel help, got %q", output)
	}
}

func TestPreviewCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "preview", "--help")
	if err != nil {
		t.Fatalf("preview --help failed: %v", err)
	}
	if !strings.Contains(output, "preview") {
		t.Fatalf("expected preview help, got %q", output)
	}
}

func TestRepairCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "repair", "--help")
	if err != nil {
		t.Fatalf("repair --help failed: %v", err)
	}
	if !strings.Contains(output, "repair") {
		t.Fatalf("expected repair help, got %q", output)
	}
}

func TestScaffoldCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "scaffold", "--help")
	if err != nil {
		t.Fatalf("scaffold --help failed: %v", err)
	}
	if !strings.Contains(output, "scaffold") {
		t.Fatalf("expected scaffold help, got %q", output)
	}
}

func TestTestCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "test", "--help")
	if err != nil {
		t.Fatalf("test --help failed: %v", err)
	}
	if !strings.Contains(output, "test") {
		t.Fatalf("expected test help, got %q", output)
	}
}

func TestTUICommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "tui", "--help")
	if err != nil {
		t.Fatalf("tui --help failed: %v", err)
	}
	if !strings.Contains(output, "tui") {
		t.Fatalf("expected tui help, got %q", output)
	}
}

func TestWorkspaceCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "workspace", "--help")
	if err != nil {
		t.Fatalf("workspace --help failed: %v", err)
	}
	if !strings.Contains(output, "workspace") {
		t.Fatalf("expected workspace help, got %q", output)
	}
}

func TestWebsocketCommandHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "ws", "--help")
	if err != nil {
		t.Fatalf("ws --help failed: %v", err)
	}
	if !strings.Contains(output, "ws") {
		t.Fatalf("expected ws help, got %q", output)
	}
}

func TestProfileRunHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "profile", "run", "--help")
	if err != nil {
		t.Fatalf("profile run --help failed: %v", err)
	}
	if !strings.Contains(output, "profile") && !strings.Contains(output, "run") {
		t.Fatalf("expected profile run help, got %q", output)
	}
}

func TestSecurityAuditHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "audit", "--help")
	if err != nil {
		t.Fatalf("security audit --help failed: %v", err)
	}
	if !strings.Contains(output, "audit") {
		t.Fatalf("expected security audit help, got %q", output)
	}
}

func TestGenerateSimpleID(t *testing.T) {
	id := generateSimpleID()
	if id == "" {
		t.Fatal("expected non-empty ID")
	}
	if len(id) != 14 {
		t.Errorf("expected 14 chars (YYYYMMDDHHMMSS), got %d: %s", len(id), id)
	}
}

func TestOrDefault(t *testing.T) {
	tests := []struct {
		s, def, want string
	}{
		{"hello", "default", "hello"},
		{"", "default", "default"},
		{"", "", ""},
	}
	for _, tt := range tests {
		got := orDefault(tt.s, tt.def)
		if got != tt.want {
			t.Errorf("orDefault(%q, %q) = %q, want %q", tt.s, tt.def, got, tt.want)
		}
	}
}

func TestParseSpecInput(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		spec, err := parseSpecInput(`{"name":"test"}`)
		if err != nil {
			t.Fatalf("parse JSON: %v", err)
		}
		if spec["name"] != "test" {
			t.Errorf("expected name=test, got %v", spec["name"])
		}
	})

	t.Run("yaml", func(t *testing.T) {
		spec, err := parseSpecInput("name: test\nversion: 1")
		if err != nil {
			t.Fatalf("parse YAML: %v", err)
		}
		if spec["name"] != "test" {
			t.Errorf("expected name=test, got %v", spec["name"])
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseSpecInput("{{{invalid}}}")
		if err == nil {
			t.Fatal("expected error for invalid input")
		}
	})
}

func TestKernelInfoHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "kernel", "info", "--help")
	if err != nil {
		t.Fatalf("kernel info --help failed: %v", err)
	}
	if !strings.Contains(output, "info") {
		t.Fatalf("expected kernel info help, got %q", output)
	}
}

func TestAuthWhoamiHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "whoami", "--help")
	if err != nil {
		t.Fatalf("auth whoami --help failed: %v", err)
	}
	if !strings.Contains(output, "whoami") {
		t.Fatalf("expected auth whoami help, got %q", output)
	}
}

func TestAuthCreateUserHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-user", "--help")
	if err != nil {
		t.Fatalf("auth create-user --help failed: %v", err)
	}
	if !strings.Contains(output, "create-user") {
		t.Fatalf("expected auth create-user help, got %q", output)
	}
}

func TestAuthCreateKeyHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-key", "--help")
	if err != nil {
		t.Fatalf("auth create-key --help failed: %v", err)
	}
	if !strings.Contains(output, "create-key") {
		t.Fatalf("expected auth create-key help, got %q", output)
	}
}

func TestAuthCreateRoleHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-role", "--help")
	if err != nil {
		t.Fatalf("auth create-role --help failed: %v", err)
	}
	if !strings.Contains(output, "create-role") {
		t.Fatalf("expected auth create-role help, got %q", output)
	}
}

func TestCloudDeployHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "deploy", "--help")
	if err != nil {
		t.Fatalf("cloud deploy --help failed: %v", err)
	}
	if !strings.Contains(output, "deploy") {
		t.Fatalf("expected cloud deploy help, got %q", output)
	}
}

func TestCloudPlanHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "plan", "--help")
	if err != nil {
		t.Fatalf("cloud plan --help failed: %v", err)
	}
	if !strings.Contains(output, "plan") {
		t.Fatalf("expected cloud plan help, got %q", output)
	}
}

func TestAISuggestHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "ai", "suggest", "--help")
	if err != nil {
		t.Fatalf("ai suggest --help failed: %v", err)
	}
	if !strings.Contains(output, "suggest") {
		t.Fatalf("expected ai suggest help, got %q", output)
	}
}

func TestAIExplainHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "ai", "explain", "--help")
	if err != nil {
		t.Fatalf("ai explain --help failed: %v", err)
	}
	if !strings.Contains(output, "explain") {
		t.Fatalf("expected ai explain help, got %q", output)
	}
}

func TestProfileListHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "profile", "list", "--help")
	if err != nil {
		t.Fatalf("profile list --help failed: %v", err)
	}
	if !strings.Contains(output, "list") {
		t.Fatalf("expected profile list help, got %q", output)
	}
}

func TestProfileShowHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "profile", "show", "--help")
	if err != nil {
		t.Fatalf("profile show --help failed: %v", err)
	}
	if !strings.Contains(output, "show") {
		t.Fatalf("expected profile show help, got %q", output)
	}
}

func TestBrokerConnectHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "broker", "connect", "--help")
	if err != nil {
		t.Fatalf("broker connect --help failed: %v", err)
	}
	if !strings.Contains(output, "connect") {
		t.Fatalf("expected broker connect help, got %q", output)
	}
}

func TestArtifactsInfoHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "artifacts", "info", "--help")
	if err != nil {
		t.Fatalf("artifacts info --help failed: %v", err)
	}
	if !strings.Contains(output, "info") {
		t.Fatalf("expected artifacts info help, got %q", output)
	}
}

func TestTUIDashboardHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "tui", "dashboard", "--help")
	if err != nil {
		t.Fatalf("tui dashboard --help failed: %v", err)
	}
	if !strings.Contains(output, "dashboard") {
		t.Fatalf("expected tui dashboard help, got %q", output)
	}
}

func TestTUIWizardHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "tui", "wizard", "--help")
	if err != nil {
		t.Fatalf("tui wizard --help failed: %v", err)
	}
	if !strings.Contains(output, "wizard") {
		t.Fatalf("expected tui wizard help, got %q", output)
	}
}
