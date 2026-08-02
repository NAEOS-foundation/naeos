package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/cloud"
)

func chmodExec(path string) error {
	return os.Chmod(path, 0o755)
}

func TestLoadCloudConfigFromSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "cloud.yaml")
	writeTestFile(t, dir, "cloud.yaml", `cloud:
  provider: aws
  region: us-east-1
  project: demo
  environment: dev
  resources:
    - name: app
      kind: compute
      spec:
        size: small
    - name: db
      type: database
`)

	cfg, err := loadCloudConfigFromSpec(specPath)
	if err != nil {
		t.Fatalf("loadCloudConfigFromSpec failed: %v", err)
	}
	if cfg.Provider != cloud.AWS {
		t.Errorf("expected aws provider, got %v", cfg.Provider)
	}
	if cfg.Region != "us-east-1" {
		t.Errorf("expected us-east-1, got %q", cfg.Region)
	}
	if len(cfg.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(cfg.Resources))
	}
	if cfg.Resources[0].Type != "compute" {
		t.Errorf("expected kind fallback for type, got %q", cfg.Resources[0].Type)
	}
	if cfg.Resources[1].Type != "database" {
		t.Errorf("expected explicit type, got %q", cfg.Resources[1].Type)
	}
	if cfg.Resources[0].Spec["size"] != "small" {
		t.Errorf("expected spec size small, got %v", cfg.Resources[0].Spec["size"])
	}
}

func TestLoadCloudConfigFromSpecMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := loadCloudConfigFromSpec(filepath.Join(dir, "missing.yaml"))
	if err == nil {
		t.Fatal("expected error for missing spec file")
	}
}

func TestLoadCloudConfigFromSpecInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	writeTestFile(t, dir, "bad.yaml", "{{{{invalid")
	_, err := loadCloudConfigFromSpec(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadCloudConfigFromSpecNoProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noprovider.yaml")
	writeTestFile(t, dir, "noprovider.yaml", "cloud:\n  region: us-east-1\n")
	_, err := loadCloudConfigFromSpec(path)
	if err == nil {
		t.Fatal("expected error when provider missing")
	}
}

func TestCloudDeployFromSpec(t *testing.T) {
	binDir := t.TempDir()
	writeTestFile(t, binDir, "terraform", "#!/bin/sh\nexit 0\n")
	if err := chmodExec(binDir + "/terraform"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	dir := t.TempDir()
	specPath := filepath.Join(dir, "cloud.yaml")
	writeTestFile(t, dir, "cloud.yaml", `cloud:
  provider: aws
  region: us-east-1
  project: demo
  environment: dev
  resources:
    - name: app
      kind: compute
`)

	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "deploy", "--input-file", specPath)
	if err != nil {
		t.Fatalf("cloud deploy failed: %v", err)
	}
	if !strings.Contains(output, "Deployed to") {
		t.Fatalf("expected deploy output, got %q", output)
	}
}

func TestCloudPlanFromSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "cloud.yaml")
	writeTestFile(t, dir, "cloud.yaml", `cloud:
  provider: aws
  region: us-east-1
  project: demo
  environment: dev
  resources:
    - name: app
      kind: compute
    - name: data
      kind: storage
`)

	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "plan", "--input-file", specPath)
	if err != nil {
		t.Fatalf("cloud plan failed: %v", err)
	}
	if !strings.Contains(output, "Plan: 2 resources") {
		t.Fatalf("expected plan output, got %q", output)
	}
	if !strings.Contains(output, "Generated HCL") {
		t.Fatalf("expected HCL output, got %q", output)
	}
}

func TestCloudExportFromSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "cloud.yaml")
	writeTestFile(t, dir, "cloud.yaml", `cloud:
  provider: aws
  region: us-east-1
  project: demo
  environment: dev
  resources:
    - name: app
      kind: compute
`)

	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "export", "--input-file", specPath)
	if err != nil {
		t.Fatalf("cloud export failed: %v", err)
	}
	if !strings.Contains(output, "terraform") && !strings.Contains(output, "resource") {
		t.Fatalf("expected terraform output, got %q", output)
	}
}

func TestCloudDeployInvalidProvider(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "cloud.yaml")
	writeTestFile(t, dir, "cloud.yaml", `cloud:
  provider: oracle
  region: us-east-1
  project: demo
  resources:
    - name: app
      kind: compute
`)

	root := NewRootCommand()
	_, err := executeCommand(root, "cloud", "deploy", "--input-file", specPath)
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
}

func TestCloudStatusWithState(t *testing.T) {
	home := t.TempDir()
	stateDir := home + "/.naeos/cloud"
	writeTestFile(t, stateDir, "aws-demo.json", `{
  "provider": "aws",
  "project": "demo",
  "region": "us-east-1",
  "status": "deployed",
  "timestamp": "2026-08-01T00:00:00Z",
  "resources": [{"name": "app", "type": "compute", "id": "arn:aws:1"}]
}`)

	t.Setenv("HOME", home)

	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "status", "--project", "demo")
	if err != nil {
		t.Fatalf("cloud status failed: %v", err)
	}
	if !strings.Contains(output, "Deployed resources") {
		t.Fatalf("expected deployed resources output, got %q", output)
	}
	if !strings.Contains(output, "arn:aws:1") {
		t.Fatalf("expected resource ID in output, got %q", output)
	}
}

func TestCloudStatusFiltered(t *testing.T) {
	home := t.TempDir()
	stateDir := home + "/.naeos/cloud"
	writeTestFile(t, stateDir, "aws-demo.json", `{
  "provider": "aws",
  "project": "other",
  "region": "us-east-1",
  "status": "deployed",
  "timestamp": "2026-08-01T00:00:00Z",
  "resources": []
}`)

	t.Setenv("HOME", home)

	root := NewRootCommand()
	output, err := executeCommand(root, "cloud", "status", "--project", "demo")
	if err != nil {
		t.Fatalf("cloud status failed: %v", err)
	}
	if !strings.Contains(output, "No deployments found matching") {
		t.Fatalf("expected filtered output, got %q", output)
	}
}
