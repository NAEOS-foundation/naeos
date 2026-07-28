package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
