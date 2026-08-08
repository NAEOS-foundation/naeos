package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeBinDir(t *testing.T, scripts map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range scripts {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := chmodExec(path); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func withPATH(t *testing.T, binDir string) {
	t.Helper()
	t.Setenv("PATH", binDir)
}

func TestCheckJavaNotFound(t *testing.T) {
	withPATH(t, t.TempDir())
	res := checkJava()
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s", res.Status)
	}
	if !strings.Contains(res.Detail, "not installed") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckJavaVersionFail(t *testing.T) {
	withPATH(t, fakeBinDir(t, map[string]string{
		"java": "echo bad java >&2\nexit 1\n",
	}))
	res := checkJava()
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s", res.Status)
	}
	if !strings.Contains(res.Detail, "version check failed") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckJavaPass(t *testing.T) {
	withPATH(t, fakeBinDir(t, map[string]string{
		"java":   "echo 'java version \"17.0.1\" 2021-10-19 LTS'\n",
		"mvn":    "echo maven\n",
		"gradle": "echo gradle\n",
	}))
	res := checkJava()
	if res.Status != "pass" {
		t.Fatalf("expected pass, got %s", res.Status)
	}
	if !strings.Contains(res.Detail, "17.0.1") {
		t.Errorf("expected version in detail, got %s", res.Detail)
	}
	if !strings.Contains(res.Detail, "maven") || !strings.Contains(res.Detail, "gradle") {
		t.Errorf("expected build tools in detail, got %s", res.Detail)
	}
}

func TestCheckRustNotFound(t *testing.T) {
	withPATH(t, t.TempDir())
	res := checkRust()
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s", res.Status)
	}
}

func TestCheckRustVersionFail(t *testing.T) {
	withPATH(t, fakeBinDir(t, map[string]string{
		"rustc": "exit 1\n",
	}))
	res := checkRust()
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s", res.Status)
	}
	if !strings.Contains(res.Detail, "version check failed") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckRustPass(t *testing.T) {
	withPATH(t, fakeBinDir(t, map[string]string{
		"rustc": "echo 'rustc 1.75.0 (0b8f1c56a 2023-12-27)'\n",
		"cargo": "echo cargo\n",
	}))
	res := checkRust()
	if res.Status != "pass" {
		t.Fatalf("expected pass, got %s", res.Status)
	}
	if !strings.Contains(res.Detail, "1.75.0") || !strings.Contains(res.Detail, "cargo available") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckSpecWithErrors(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "naeos.yaml")
	if err := os.WriteFile(spec, []byte("{{invalid yaml}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := checkSpec(spec)
	if res.Status != "fail" {
		t.Errorf("expected fail, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "error(s)") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckSpecWarningsOnly(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "naeos.yaml")
	if err := os.WriteFile(spec, []byte(`project: -bad-project-name
modules:
  - name: core
    path: ./internal/core
`), 0o644); err != nil {
		t.Fatal(err)
	}
	res := checkSpec(spec)
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "warning(s)") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckGoModuleMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	res := checkGoModule()
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s", res.Status)
	}
	if !strings.Contains(res.Detail, "go.mod not found") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckGoModuleVerifyFail(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module broken\nbad line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	res := checkGoModule()
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s", res.Status)
	}
}

func TestCheckConfigNoFile(t *testing.T) {
	t.Chdir(t.TempDir())
	res := checkConfig("")
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s", res.Status)
	}
}

func TestCheckConfigInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(": : : invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := checkConfig(path)
	if res.Status != "fail" {
		t.Errorf("expected fail, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "invalid config") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckConfigValid(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "out"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	content := "pipeline:\n  name: test\n  mode: local\n  output_dir: " + filepath.Join(dir, "out") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	res := checkConfig(path)
	if res.Status != "pass" {
		t.Errorf("expected pass, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "[exists]") {
		t.Errorf("expected existing output_dir, got %s", res.Detail)
	}
}

func TestCheckConfigOutputDirMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "pipeline:\n  name: test\n  mode: local\n  output_dir: " + filepath.Join(dir, "does-not-exist") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	res := checkConfig(path)
	if res.Status != "pass" {
		t.Errorf("expected pass, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "[missing]") {
		t.Errorf("expected missing output_dir, got %s", res.Detail)
	}
}

func TestCheckBrokerNoConnections(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	res := checkBroker()
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "no saved connections") {
		t.Errorf("unexpected detail: %s", res.Detail)
	}
}

func TestCheckBrokerInitFailed(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "naeos")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `[{"name":"unknown","driver":"does-not-exist","config":{}}]`
	if err := os.WriteFile(filepath.Join(dir, "brokers.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	res := checkBroker()
	if res.Status != "pass" {
		t.Errorf("expected pass, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "init failed") {
		t.Errorf("expected init failed detail, got %s", res.Detail)
	}
}

func TestCheckBrokerConnected(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "naeos")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `[{"name":"mem","driver":"memory","config":{"host":"127.0.0.1","port":0}}]`
	if err := os.WriteFile(filepath.Join(dir, "brokers.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	res := checkBroker()
	if res.Status != "pass" {
		t.Errorf("expected pass, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "connected") {
		t.Errorf("expected connected detail, got %s", res.Detail)
	}
}

func TestCheckDatabaseNoConnections(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	res := checkDatabase()
	if res.Status != "warn" {
		t.Errorf("expected warn, got %s (%s)", res.Status, res.Detail)
	}
}

func TestCheckDatabaseInitFailed(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".naeos", "db")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `[{"name":"unknown","driver":"does-not-exist","config":{}}]`
	if err := os.WriteFile(filepath.Join(dir, "connections.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	res := checkDatabase()
	if res.Status != "pass" {
		t.Errorf("expected pass, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "init failed") {
		t.Errorf("expected init failed detail, got %s", res.Detail)
	}
}

func TestCheckOutputWritable(t *testing.T) {
	t.Chdir(t.TempDir())
	res := checkOutputWritable()
	if res.Status != "pass" {
		t.Errorf("expected pass, got %s (%s)", res.Status, res.Detail)
	}
	if _, err := os.Stat("output"); !os.IsNotExist(err) {
		t.Errorf("output dir should be cleaned up")
	}
}
