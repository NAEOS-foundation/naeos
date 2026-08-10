package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func writeKernelTestConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n")
	return configPath
}

func TestKernelServicesJSON(t *testing.T) {
	configPath := writeKernelTestConfig(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "kernel", "services", "--config", configPath, "--output", "json")
	if err != nil {
		t.Fatalf("kernel services json failed: %v", err)
	}
	if !strings.Contains(output, "parser") {
		t.Fatalf("expected parser in service list, got %q", output)
	}
}

func TestKernelServicesYAML(t *testing.T) {
	configPath := writeKernelTestConfig(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "kernel", "services", "--config", configPath, "--output", "yaml")
	if err != nil {
		t.Fatalf("kernel services yaml failed: %v", err)
	}
	if !strings.Contains(output, "pipeline") {
		t.Fatalf("expected pipeline in yaml output, got %q", output)
	}
}

func TestKernelMetricsJSON(t *testing.T) {
	configPath := writeKernelTestConfig(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "kernel", "metrics", "--config", configPath, "--output", "json")
	if err != nil {
		t.Fatalf("kernel metrics json failed: %v", err)
	}
	if !strings.Contains(output, `"Events"`) || !strings.Contains(output, `"LastEvent"`) {
		t.Fatalf("expected metrics json output, got %q", output)
	}
}

func TestKernelServicesMissingConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "kernel", "services", "--config", "/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent config")
	}
}

func TestKernelMetricsMissingConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "kernel", "metrics", "--config", "/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent config")
	}
}

func TestKernelPublishMissingTopic(t *testing.T) {
	configPath := writeKernelTestConfig(t)

	root := NewRootCommand()
	_, err := executeCommand(root, "kernel", "publish", "--config", configPath)
	if err == nil {
		t.Fatal("expected error when --topic is missing")
	}
}

func TestKernelSubscribeMissingTopic(t *testing.T) {
	configPath := writeKernelTestConfig(t)

	root := NewRootCommand()
	_, err := executeCommand(root, "kernel", "subscribe", "--config", configPath)
	if err == nil {
		t.Fatal("expected error when --topic is missing")
	}
}
