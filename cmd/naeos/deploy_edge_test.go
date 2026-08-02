package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunDeployLocal(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")
	writeTestFile(t, dir, "output/app.go", "package app\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "local", "development", false)
	if err != nil {
		t.Fatalf("runDeploy local failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Deploying") {
		t.Errorf("expected deploy banner, got %q", buf.String())
	}

	copied := filepath.Join(dir, "deploy-test-deploy", "app.go")
	if _, err := os.Stat(copied); err != nil {
		t.Errorf("expected copied file: %v", err)
	}
}

func TestRunDeployLocalDryRun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")
	writeTestFile(t, dir, "output/app.go", "package app\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "local", "development", true)
	if err != nil {
		t.Fatalf("runDeploy dry-run failed: %v", err)
	}
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run message, got %q", buf.String())
	}
}

func TestRunDeployDockerDryRun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")
	writeTestFile(t, dir, "output/app.go", "package app\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "docker", "staging", true)
	if err != nil {
		t.Fatalf("runDeploy docker dry-run failed: %v", err)
	}
	if !strings.Contains(buf.String(), "docker build") {
		t.Errorf("expected docker build command, got %q", buf.String())
	}
}

func TestRunDeployK8sDryRun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")
	writeTestFile(t, dir, "output/k8s/app.yaml", "apiVersion: v1\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "k8s", "prod", true)
	if err != nil {
		t.Fatalf("runDeploy k8s dry-run failed: %v", err)
	}
	if !strings.Contains(buf.String(), "kubectl apply") {
		t.Errorf("expected kubectl apply command, got %q", buf.String())
	}
}

func TestRunDeployComposeDryRun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")
	writeTestFile(t, dir, "output/app.go", "package app\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "compose", "dev", true)
	if err != nil {
		t.Fatalf("runDeploy compose dry-run failed: %v", err)
	}
	if !strings.Contains(buf.String(), "docker-compose up") {
		t.Errorf("expected docker-compose command, got %q", buf.String())
	}
}

func TestRunDeploySSHDryRun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")
	writeTestFile(t, dir, "output/app.go", "package app\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "ssh", "user@host", true)
	if err != nil {
		t.Fatalf("runDeploy ssh dry-run failed: %v", err)
	}
	if !strings.Contains(buf.String(), "rsync") {
		t.Errorf("expected rsync command, got %q", buf.String())
	}
}

func TestRunDeployRsyncDryRun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")
	writeTestFile(t, dir, "output/app.go", "package app\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "rsync", "user@host", true)
	if err != nil {
		t.Fatalf("runDeploy rsync dry-run failed: %v", err)
	}
	if !strings.Contains(buf.String(), "rsync -avz") {
		t.Errorf("expected rsync command, got %q", buf.String())
	}
}

func TestRunDeployMissingOutput(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "local", "development", false)
	if err == nil {
		t.Fatal("expected error when output dir missing")
	}
	if !strings.Contains(err.Error(), "Run 'naeos run' first") {
		t.Errorf("expected helpful message, got %v", err)
	}
}

func TestRunDeployUnknownTarget(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: deploy-test\n  mode: development\n  output_dir: ./output\n")
	writeTestFile(t, dir, "output/app.go", "package app\n")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, filepath.Join(dir, "config.yaml"), "helm", "development", false)
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestRunDeployMissingConfig(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeploy(cmd, "/nonexistent/config.yaml", "local", "development", false)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestCpDir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeTestFile(t, src, "a.txt", "hello")
	writeTestFile(t, src, "sub/b.txt", "world")

	dst := filepath.Join(dir, "dst")
	if err := cpDir(src, dst); err != nil {
		t.Fatal(err)
	}

	for _, f := range []string{"a.txt", "sub/b.txt"} {
		if _, err := os.Stat(filepath.Join(dst, f)); err != nil {
			t.Errorf("expected %s in dest: %v", f, err)
		}
	}
}
