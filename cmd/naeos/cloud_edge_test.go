// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chmodExec(path string) error {
	return os.Chmod(path, 0o755)
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
