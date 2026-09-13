// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/agent"
)

func TestInitCreatesConfigFile(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "config.yaml")

	err := run([]string{"init", "--output", output})
	if err != nil {
		t.Fatalf("run init returned error: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected generated config file to contain content")
	}
}

func TestRunOutputIncludesControlPlaneContext(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "run.json")

	spec := "project: demo-control-plane\nmodules:\n  - name: api\n    path: ./internal/api\nservices:\n  - name: api\n    kind: http\n    port: 8080\n"

	if err := run([]string{"run", "--input", spec, "--output", "json", "--output-file", outputPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read run output file: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal run output: %v", err)
	}

	if _, ok := payload["context"]; !ok {
		t.Fatal("expected run JSON output to include a context bundle")
	}
	if _, ok := payload["validation"]; !ok {
		t.Fatal("expected run JSON output to include validation status")
	}
	if _, ok := payload["policy"]; !ok {
		t.Fatal("expected run JSON output to include policy status")
	}
	if _, ok := payload["stages"]; !ok {
		t.Fatal("expected run JSON output to include ordered pipeline stages")
	}
	if _, ok := payload["run_id"]; !ok {
		t.Fatal("expected run JSON output to include an explicit run_id")
	}
	if _, ok := payload["specification_hash"]; !ok {
		t.Fatal("expected run JSON output to include a specification_hash")
	}
	if _, ok := payload["neir_hash"]; !ok {
		t.Fatal("expected run JSON output to include a neir_hash")
	}
}

func TestValidateUsesConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"validate", "--config", configPath, "--input", "sample specification"})
	if err != nil {
		t.Fatalf("run validate returned error: %v", err)
	}
}

func TestDoctorUsesConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"doctor", "--config", configPath})
	if err != nil {
		t.Fatalf("run doctor returned error: %v", err)
	}
}

func TestRepairCreatesOutputDirectory(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	outputDir := filepath.Join(dir, "out")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: "+outputDir+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"repair", "--config", configPath})
	if err != nil {
		t.Fatalf("run repair returned error: %v", err)
	}
	if _, err := os.Stat(outputDir); err != nil {
		t.Fatalf("expected output directory to exist: %v", err)
	}
}

func TestScaffoldCreatesStarterFiles(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "demo-app")

	err := run([]string{"scaffold", "--name", "demo-app", "--output", outputDir})
	if err != nil {
		t.Fatalf("run scaffold returned error: %v", err)
	}

	for _, name := range []string{"README.md", "spec.yaml", "Makefile", ".gitignore", "Dockerfile", ".github/workflows/ci.yml", "go.mod", "cmd/app/main.go"} {
		path := filepath.Join(outputDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
	for _, name := range []string{"internal/core/README.md", "internal/core/package.go", "internal/core/config.yaml", "internal/core/handler.go", "internal/core/repository.go", "internal/core/service.go", "internal/core/domain/model.go", "internal/core/http/handler.go", "internal/core/http/router.go", "internal/core/middleware/logging.go", "internal/core/config/config.go", "internal/core/config/load.go", "config.yaml", "config.json"} {
		path := filepath.Join(outputDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
	mainPath := filepath.Join(outputDir, "cmd", "app", "main.go")
	data, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("read scaffold main entrypoint: %v", err)
	}
	if !strings.Contains(string(data), "internal/core") || !strings.Contains(string(data), "NewHandler") || !strings.Contains(string(data), "config.Load") {
		t.Fatalf("expected scaffold main entrypoint to reference the generated module, got %q", string(data))
	}
	if !strings.Contains(string(data), "http.NewServeMux") || !strings.Contains(string(data), "ListenAndServe") || !strings.Contains(string(data), "/health") || !strings.Contains(string(data), "/api/v1") || !strings.Contains(string(data), "/api/v1/resources") {
		t.Fatalf("expected scaffold main entrypoint to start a runnable HTTP server with health and versioned resource endpoints, got %q", string(data))
	}
}

func TestAgentCreateAppendActionAndGetSession(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "agent-store.json")

	if err := run([]string{"agent", "create", "--store-path", storePath, "--agent-id", "demo-agent", "--project", "demo-project"}); err != nil {
		t.Fatalf("run agent create returned error: %v", err)
	}

	store, err := loadAgentStore(storePath)
	if err != nil {
		t.Fatalf("load agent store: %v", err)
	}
	if len(store.ListSessions()) != 1 {
		t.Fatal("expected exactly one session to be stored")
	}

	sessionID := store.ListSessions()[0].ID
	if err := run([]string{"agent", "append-action", "--store-path", storePath, "--session-id", sessionID, "--type", "tool", "--target", "shell", "--agent-id", "demo-agent", "--reason", "validate fix", "--decision", "allow", "--parameters", `{"cmd":"echo ok"}`}); err != nil {
		t.Fatalf("run agent append-action returned error: %v", err)
	}

	store, err = loadAgentStore(storePath)
	if err != nil {
		t.Fatalf("reload agent store: %v", err)
	}
	session, ok := store.GetSession(sessionID)
	if !ok {
		t.Fatal("expected session to persist after append action")
	}
	if len(session.Actions) != 1 {
		t.Fatalf("expected one stored action, got %d", len(session.Actions))
	}
	if session.Actions[0].Target != "shell" {
		t.Fatalf("expected appended action target to be shell, got %q", session.Actions[0].Target)
	}
	if session.Actions[0].Parameters["cmd"] != "echo ok" {
		t.Fatalf("expected appended action parameters to persist, got %#v", session.Actions[0].Parameters)
	}
}

func TestAgentStoreListsActionsAndDeletesSession(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "agent-store.json")
	store := agent.NewStore(storePath)

	if err := store.Load(); err != nil {
		t.Fatalf("load agent store: %v", err)
	}

	session, err := store.CreateSession(agent.Session{AgentID: "cleanup-agent", Project: "demo-project", Status: "active"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := store.AppendAction(session.ID, agent.Action{Type: "tool", Target: "shell", Decision: "ALLOW"}); err != nil {
		t.Fatalf("append action: %v", err)
	}
	if err := store.Save(); err != nil {
		t.Fatalf("save session store: %v", err)
	}

	actions := store.ListActions()
	if len(actions) != 1 {
		t.Fatalf("expected one action in store, got %d", len(actions))
	}
	if actions[0].Target != "shell" {
		t.Fatalf("expected listed action target shell, got %q", actions[0].Target)
	}

	if err := store.DeleteSession(session.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, ok := store.GetSession(session.ID); ok {
		t.Fatal("expected deleted session to be removed")
	}
	if len(store.ListSessions()) != 0 {
		t.Fatalf("expected zero sessions after deletion, got %d", len(store.ListSessions()))
	}
}

func TestRuntimeExecPersistsActionToSession(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "agent-store.json")

	store := agent.NewStore(storePath)
	if err := store.Load(); err != nil {
		t.Fatalf("load agent store before create: %v", err)
	}
	if _, err := store.CreateSession(agent.Session{AgentID: "runtime-agent", Project: "demo-project", Status: "active"}); err != nil {
		t.Fatalf("create session for runtime test: %v", err)
	}
	if err := store.Save(); err != nil {
		t.Fatalf("save session for runtime test: %v", err)
	}

	session, ok := store.GetSession(store.ListSessions()[0].ID)
	if !ok {
		t.Fatal("expected session to exist after create")
	}

	if err := run([]string{"runtime", "exec", "--tool", "shell", "--action", "run", "--resource", "scripts/deploy.sh", "--environment", "production", "--actor", "ci-bot", "--session-id", session.ID, "--store-path", storePath}); err != nil {
		t.Fatalf("run runtime exec returned error: %v", err)
	}

	store, err := loadAgentStore(storePath)
	if err != nil {
		t.Fatalf("reload agent store after runtime exec: %v", err)
	}
	session, ok = store.GetSession(session.ID)
	if !ok {
		t.Fatal("expected session to still exist after runtime exec")
	}
	if len(session.Actions) != 1 {
		t.Fatalf("expected one stored action, got %d", len(session.Actions))
	}
	if session.Actions[0].Type != "run" {
		t.Fatalf("expected action type run, got %q", session.Actions[0].Type)
	}
	if session.Actions[0].Target != "scripts/deploy.sh" {
		t.Fatalf("expected action target scripts/deploy.sh, got %q", session.Actions[0].Target)
	}
	if session.Actions[0].Decision != "DENY" {
		t.Fatalf("expected runtime exec to persist a DENY decision when no policy is registered, got %q", session.Actions[0].Decision)
	}
	if session.Actions[0].Parameters["tool"] != "shell" {
		t.Fatalf("expected persisted parameter tool=shell, got %#v", session.Actions[0].Parameters)
	}
}

func TestInspectUsesConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"inspect", "--config", configPath, "--input", "sample specification"})
	if err != nil {
		t.Fatalf("run inspect returned error: %v", err)
	}
}

func TestInspectReadsInputFromFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	outputPath := filepath.Join(dir, "inspect.txt")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("project: file-driven-project\nmodules:\n  - name: api\n    path: ./internal/api\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"inspect", "--config", configPath, "--input", specPath, "--output", "text", "--output-file", outputPath})
	if err != nil {
		t.Fatalf("run inspect returned error: %v", err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read inspect output file: %v", err)
	}
	if !strings.Contains(string(data), "project=file-driven-project") {
		t.Fatalf("expected inspect output to contain parsed project name, got %q", string(data))
	}
}

func TestExportWritesArtifactsToDirectory(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	outputDir := filepath.Join(dir, "generated")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: "+outputDir+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("project: export-demo\nmodules:\n  - name: api\n    path: ./internal/api\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"export", "--config", configPath, "--input", specPath})
	if err != nil {
		t.Fatalf("run export returned error: %v", err)
	}
	for _, name := range []string{"README.md", "Dockerfile", ".github/workflows/ci.yml", "go.mod", "cmd/app/main.go"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Fatalf("expected exported %s to exist: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(outputDir, "internal", "api")); err != nil {
		t.Fatalf("expected exported module directory to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "internal", "api", "config.yaml")); err != nil {
		t.Fatalf("expected exported service config file to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "internal", "api", "handler.go")); err != nil {
		t.Fatalf("expected exported handler skeleton to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "internal", "api", "repository.go")); err != nil {
		t.Fatalf("expected exported repository skeleton to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "internal", "api", "service.go")); err != nil {
		t.Fatalf("expected exported service skeleton to exist: %v", err)
	}
	for _, name := range []string{"internal/api/domain/model.go", "internal/api/http/handler.go", "internal/api/http/router.go", "internal/api/middleware/logging.go", "internal/api/config/config.go", "internal/api/config/load.go"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Fatalf("expected exported %s to exist: %v", name, err)
		}
	}

	for _, file := range []struct {
		path    string
		content string
	}{
		{path: filepath.Join(outputDir, "internal", "api", "handler.go"), content: "type Handler struct"},
		{path: filepath.Join(outputDir, "internal", "api", "repository.go"), content: "type Repository interface"},
		{path: filepath.Join(outputDir, "internal", "api", "service.go"), content: "type Service interface"},
	} {
		data, err := os.ReadFile(file.path)
		if err != nil {
			t.Fatalf("read %s: %v", file.path, err)
		}
		if !strings.Contains(string(data), file.content) {
			t.Fatalf("expected %s to contain %q, got %q", file.path, file.content, string(data))
		}
	}
}

func TestPreviewShowsGeneratedArtifacts(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("project: preview-demo\nmodules:\n  - name: api\n    path: ./internal/api\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"preview", "--config", configPath, "--input", specPath})
	if err != nil {
		t.Fatalf("run preview returned error: %v", err)
	}
}

func TestRunSupportsJSONOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input", "sample specification", "--output", "json"})
	if err != nil {
		t.Fatalf("run run returned error: %v", err)
	}
}

func TestRunSupportsYAMLOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input", "sample specification", "--output", "yaml"})
	if err != nil {
		t.Fatalf("run run returned error: %v", err)
	}
}

func TestRunWritesOutputToFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	outputPath := filepath.Join(dir, "result.json")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run([]string{"run", "--config", configPath, "--input", "sample specification", "--output", "json", "--output-file", outputPath})
	if err != nil {
		t.Fatalf("run run returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected output file to contain content")
	}
}

func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	root.SilenceErrors = true
	root.SilenceUsage = true
	_, err := root.ExecuteC()
	return buf.String(), err
}

func TestVerifyCommandUsesPipelineAndJSONOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root,
		"verify",
		"--config", configPath,
		"--input", "project: verify-demo\nmodules:\n  - name: api\n    path: ./internal/api\n",
		"--output-format", "json",
	)
	if err != nil {
		t.Fatalf("execute verify failed: %v", err)
	}

	if !strings.Contains(output, `"status": "valid"`) {
		t.Fatalf("expected verify json output with valid status, got %q", output)
	}
	if !strings.Contains(output, `"project": "verify-demo"`) {
		t.Fatalf("expected verify json output to include project name, got %q", output)
	}
}

func TestValidateCobraJSONOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "validate", "--config", configPath, "--input", "sample specification", "--output", "json")
	if err != nil {
		t.Fatalf("execute validate failed: %v", err)
	}

	if !strings.Contains(output, `"status": "valid"`) {
		t.Fatalf("expected json output, got %q", output)
	}
}

func TestValidateCobraYAMLOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "validate", "--config", configPath, "--input", "sample specification", "--output", "yaml")
	if err != nil {
		t.Fatalf("execute validate failed: %v", err)
	}

	if !strings.Contains(output, "status: valid") {
		t.Fatalf("expected yaml output, got %q", output)
	}
}

func TestValidateCobraAliasV(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	if _, err := executeCommand(root, "v", "--config", configPath, "--input", "sample specification"); err != nil {
		t.Fatalf("execute v failed: %v", err)
	}
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	os.Stdout = oldStdout
	return string(out)
}

func TestValidateSupportsJSONOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	output := captureOutput(t, func() {
		if err := run([]string{"validate", "--config", configPath, "--input", "sample specification", "--output", "json"}); err != nil {
			t.Fatalf("run validate returned error: %v", err)
		}
	})

	if !strings.Contains(output, `"status": "valid"`) {
		t.Fatalf("expected JSON output, got %q", output)
	}
}

func TestValidateSupportsYAMLOutput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	output := captureOutput(t, func() {
		if err := run([]string{"validate", "--config", configPath, "--input", "sample specification", "--output", "yaml"}); err != nil {
			t.Fatalf("run validate returned error: %v", err)
		}
	})

	if !strings.Contains(output, "status: valid") {
		t.Fatalf("expected YAML output, got %q", output)
	}
}

func TestValidateAliasV(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := run([]string{"v", "--config", configPath, "--input", "sample specification"}); err != nil {
		t.Fatalf("run v returned error: %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "version")
	if err != nil {
		t.Fatalf("execute version failed: %v", err)
	}
	if !strings.Contains(output, "naeos ") {
		t.Fatalf("expected version output, got %q", output)
	}
}

func TestKernelServicesCommand(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "kernel", "services", "--config", configPath, "--output", "text")
	if err != nil {
		t.Fatalf("execute kernel services failed: %v", err)
	}
	if !strings.Contains(output, "pipeline") || !strings.Contains(output, "parser") {
		t.Fatalf("expected kernel service list, got %q", output)
	}
}

func TestKernelMetricsCommand(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "kernel", "metrics", "--config", configPath, "--output", "text")
	if err != nil {
		t.Fatalf("execute kernel metrics failed: %v", err)
	}
	if !strings.Contains(output, "events=") {
		t.Fatalf("expected kernel metrics output, got %q", output)
	}
}

func TestKernelEventsCommand(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "kernel", "events", "--config", configPath, "--output", "text")
	if err != nil {
		t.Fatalf("execute kernel events failed: %v", err)
	}
	if strings.TrimSpace(output) != "" {
		t.Fatalf("expected no events output when no topics are registered, got %q", output)
	}
}

func TestKernelPublishSubscribeCommand(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	publishOutput, err := executeCommand(root, "kernel", "publish", "--config", configPath, "--topic", "test", "--payload", "hello", "--output", "text")
	if err != nil {
		t.Fatalf("execute kernel publish failed: %v", err)
	}
	if !strings.Contains(publishOutput, "published topic=test payload=hello") {
		t.Fatalf("expected publish output, got %q", publishOutput)
	}

	subscribeOutput, err := executeCommand(root, "kernel", "subscribe", "--config", configPath, "--topic", "test", "--payload", "hello", "--output", "text")
	if err != nil {
		t.Fatalf("execute kernel subscribe failed: %v", err)
	}
	if !strings.Contains(subscribeOutput, "topic=test") || !strings.Contains(subscribeOutput, "received=hello") {
		t.Fatalf("expected subscribe output, got %q", subscribeOutput)
	}
}

func TestRootVerboseFlag(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "--verbose", "validate", "--config", filepath.Join(t.TempDir(), "config.yaml"), "--input", "sample specification")
	if err == nil {
		t.Fatalf("expected validate to fail with missing config file, got output: %q", output)
	}
}
