package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffRequiresInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "diff")
	if err == nil {
		t.Fatal("expected error for missing --input/--input-file")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected 'missing' error, got: %v", err)
	}
}

func TestDiffWithInput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input", "project: test", "--output-dir", dir)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected Summary in output, got %q", output)
	}
}

func TestDiffVisual(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input", "project: visual-test", "--output-dir", dir, "--visual")
	if err != nil {
		t.Fatalf("diff --visual failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected Summary, got %q", output)
	}
}

func TestDiffFormatUnified(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "diff", "--config", configPath, "--input", "project: unified-test", "--output-dir", dir, "--format", "unified")
	if err != nil {
		t.Fatalf("diff --format unified failed: %v", err)
	}
	if !strings.Contains(output, "Summary:") {
		t.Fatalf("expected Summary, got %q", output)
	}
}

func TestBuildWithInput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "build", "--config", configPath, "--input", "project: build-test")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if !strings.Contains(output, "pipeline") {
		t.Fatalf("expected pipeline output, got %q", output)
	}
}

func TestBuildDryRun(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "build", "--config", configPath, "--input", "project: dryrun-test", "--dry-run")
	if err != nil {
		t.Fatalf("build --dry-run failed: %v", err)
	}
	if !strings.Contains(output, "pipeline") {
		t.Fatalf("expected pipeline output, got %q", output)
	}
}

func TestLSPCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "lsp", "--help")
	if err != nil {
		t.Fatalf("lsp --help failed: %v", err)
	}
	if !strings.Contains(output, "Language Server Protocol") {
		t.Fatalf("expected LSP help, got %q", output)
	}
}

func TestDXCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "dx", "--help")
	if err != nil {
		t.Fatalf("dx --help failed: %v", err)
	}
	if !strings.Contains(output, "VS Code") {
		t.Fatalf("expected DX help about VS Code, got %q", output)
	}
}

func TestDXVSCodeGen(t *testing.T) {
	dir := t.TempDir()

	root := NewRootCommand()
	output, err := executeCommand(root, "dx", "vscode-gen", "--output", dir)
	if err != nil {
		t.Fatalf("dx vscode-gen failed: %v", err)
	}
	if !strings.Contains(output, "VS Code extension generated") {
		t.Fatalf("expected success message, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "package.json")); os.IsNotExist(err) {
		t.Fatal("expected package.json to be generated")
	}
	if _, err := os.Stat(filepath.Join(dir, "extension.js")); os.IsNotExist(err) {
		t.Fatal("expected extension.js to be generated")
	}
}

func TestCLIVersionCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(output, "naeos") {
		t.Fatalf("expected version info, got %q", output)
	}
}

func TestConfigCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "config", "--help")
	if err != nil {
		t.Fatalf("config --help failed: %v", err)
	}
	if !strings.Contains(output, "config") {
		t.Fatalf("expected config help, got %q", output)
	}
}

func TestInitCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "init", "--help")
	if err != nil {
		t.Fatalf("init --help failed: %v", err)
	}
	if !strings.Contains(output, "init") {
		t.Fatalf("expected init help, got %q", output)
	}
}

func TestStatusCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "status", "--help")
	if err != nil {
		t.Fatalf("status --help failed: %v", err)
	}
	if !strings.Contains(output, "status") {
		t.Fatalf("expected status help, got %q", output)
	}
}

func TestHealthCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "health", "--help")
	if err != nil {
		t.Fatalf("health --help failed: %v", err)
	}
	if !strings.Contains(output, "health") {
		t.Fatalf("expected health help, got %q", output)
	}
}

func TestTemplateCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "template", "--help")
	if err != nil {
		t.Fatalf("template --help failed: %v", err)
	}
	if !strings.Contains(output, "template") {
		t.Fatalf("expected template help, got %q", output)
	}
}

func TestPluginCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "--help")
	if err != nil {
		t.Fatalf("plugin --help failed: %v", err)
	}
	if !strings.Contains(output, "plugin") {
		t.Fatalf("expected plugin help, got %q", output)
	}
}

func TestSchemaCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "schema", "--help")
	if err != nil {
		t.Fatalf("schema --help failed: %v", err)
	}
	if !strings.Contains(output, "schema") {
		t.Fatalf("expected schema help, got %q", output)
	}
}

func TestComplianceCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "compliance", "--help")
	if err != nil {
		t.Fatalf("compliance --help failed: %v", err)
	}
	if !strings.Contains(output, "compliance") {
		t.Fatalf("expected compliance help, got %q", output)
	}
}

func TestSearchCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "search", "--help")
	if err != nil {
		t.Fatalf("search --help failed: %v", err)
	}
	if !strings.Contains(output, "search") {
		t.Fatalf("expected search help, got %q", output)
	}
}

func TestAiCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "ai", "--help")
	if err != nil {
		t.Fatalf("ai --help failed: %v", err)
	}
	if !strings.Contains(output, "ai") {
		t.Fatalf("expected ai help, got %q", output)
	}
}

func TestSecurityCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "--help")
	if err != nil {
		t.Fatalf("security --help failed: %v", err)
	}
	if !strings.Contains(output, "security") {
		t.Fatalf("expected security help, got %q", output)
	}
}

func TestDocgenCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "docgen", "--help")
	if err != nil {
		t.Fatalf("docgen --help failed: %v", err)
	}
	if !strings.Contains(output, "doc") {
		t.Fatalf("expected docgen help, got %q", output)
	}
}

func TestAuthCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "--help")
	if err != nil {
		t.Fatalf("auth --help failed: %v", err)
	}
	if !strings.Contains(output, "auth") {
		t.Fatalf("expected auth help, got %q", output)
	}
}

func TestDashboardCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "dashboard", "--help")
	if err != nil {
		t.Fatalf("dashboard --help failed: %v", err)
	}
	if !strings.Contains(output, "dashboard") {
		t.Fatalf("expected dashboard help, got %q", output)
	}
}

func TestObservabilityCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "observability", "--help")
	if err != nil {
		t.Fatalf("observability --help failed: %v", err)
	}
	if !strings.Contains(output, "observability") {
		t.Fatalf("expected observability help, got %q", output)
	}
}

func TestEventsCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "events", "--help")
	if err != nil {
		t.Fatalf("events --help failed: %v", err)
	}
	if !strings.Contains(output, "events") {
		t.Fatalf("expected events help, got %q", output)
	}
}

func TestMCPCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "mcp", "--help")
	if err != nil {
		t.Fatalf("mcp --help failed: %v", err)
	}
	if !strings.Contains(output, "mcp") {
		t.Fatalf("expected mcp help, got %q", output)
	}
}

func TestValidateCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "validate", "--help")
	if err != nil {
		t.Fatalf("validate --help failed: %v", err)
	}
	if !strings.Contains(output, "validate") {
		t.Fatalf("expected validate help, got %q", output)
	}
}

func TestWorkflowCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "workflow", "--help")
	if err != nil {
		t.Fatalf("workflow --help failed: %v", err)
	}
	if !strings.Contains(output, "workflow") {
		t.Fatalf("expected workflow help, got %q", output)
	}
}

func TestSupabaseCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "--help")
	if err != nil {
		t.Fatalf("supabase --help failed: %v", err)
	}
	if !strings.Contains(output, "supabase") {
		t.Fatalf("expected supabase help, got %q", output)
	}
}

func TestMarketplaceCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "--help")
	if err != nil {
		t.Fatalf("marketplace --help failed: %v", err)
	}
	if !strings.Contains(output, "marketplace") {
		t.Fatalf("expected marketplace help, got %q", output)
	}
}

func TestWatchCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "watch", "--help")
	if err != nil {
		t.Fatalf("watch --help failed: %v", err)
	}
	if !strings.Contains(output, "watch") {
		t.Fatalf("expected watch help, got %q", output)
	}
}

func TestRollbackCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "rollback", "--help")
	if err != nil {
		t.Fatalf("rollback --help failed: %v", err)
	}
	if !strings.Contains(output, "rollback") {
		t.Fatalf("expected rollback help, got %q", output)
	}
}

func TestLockCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "lock", "--help")
	if err != nil {
		t.Fatalf("lock --help failed: %v", err)
	}
	if !strings.Contains(output, "lock") {
		t.Fatalf("expected lock help, got %q", output)
	}
}

func TestMigrateCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "--help")
	if err != nil {
		t.Fatalf("migrate --help failed: %v", err)
	}
	if !strings.Contains(output, "migrate") {
		t.Fatalf("expected migrate help, got %q", output)
	}
}

func TestDeployCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "deploy", "--help")
	if err != nil {
		t.Fatalf("deploy --help failed: %v", err)
	}
	if !strings.Contains(output, "deploy") {
		t.Fatalf("expected deploy help, got %q", output)
	}
}

func TestRunCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "run", "--help")
	if err != nil {
		t.Fatalf("run --help failed: %v", err)
	}
	if !strings.Contains(output, "run") {
		t.Fatalf("expected run help, got %q", output)
	}
}

func TestCreateCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "create", "--help")
	if err != nil {
		t.Fatalf("create --help failed: %v", err)
	}
	if !strings.Contains(output, "create") {
		t.Fatalf("expected create help, got %q", output)
	}
}

func TestLintCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "lint", "--help")
	if err != nil {
		t.Fatalf("lint --help failed: %v", err)
	}
	if !strings.Contains(output, "lint") {
		t.Fatalf("expected lint help, got %q", output)
	}
}

func TestCICDCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "cicd", "--help")
	if err != nil {
		t.Fatalf("cicd --help failed: %v", err)
	}
	if !strings.Contains(output, "cicd") {
		t.Fatalf("expected cicd help, got %q", output)
	}
}

func TestGatewayCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "gateway", "--help")
	if err != nil {
		t.Fatalf("gateway --help failed: %v", err)
	}
	if !strings.Contains(output, "gateway") {
		t.Fatalf("expected gateway help, got %q", output)
	}
}

func TestSSOCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "sso", "--help")
	if err != nil {
		t.Fatalf("auth sso --help failed: %v", err)
	}
	if !strings.Contains(output, "sso") {
		t.Fatalf("expected SSO help, got %q", output)
	}
}

func TestProfileCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "profile", "--help")
	if err != nil {
		t.Fatalf("profile --help failed: %v", err)
	}
	if !strings.Contains(output, "profile") {
		t.Fatalf("expected profile help, got %q", output)
	}
}

func TestArtifactsCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "artifacts", "--help")
	if err != nil {
		t.Fatalf("artifacts --help failed: %v", err)
	}
	if !strings.Contains(output, "artifact") {
		t.Fatalf("expected artifacts help, got %q", output)
	}
}
