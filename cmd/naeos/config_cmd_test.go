package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigEncryptToFile(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "config.yaml")
	out := filepath.Join(dir, "config.enc")
	writeTestFile(t, dir, "config.yaml", "name: myproject\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "encrypt", "--input", in, "--passphrase", "secret", "--output", out)
	if err != nil {
		t.Fatalf("config encrypt failed: %v", err)
	}
	if !strings.Contains(output, "Encrypted config written to") {
		t.Fatalf("expected write message, got %q", output)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read encrypted file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty encrypted output")
	}
}

func TestConfigEncryptStdout(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "name: myproject\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "encrypt", "--input", in, "--passphrase", "secret")
	if err != nil {
		t.Fatalf("config encrypt failed: %v", err)
	}
	if len(strings.TrimSpace(output)) == 0 {
		t.Fatal("expected encrypted content on stdout")
	}
}

func TestConfigEncryptMissingInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "encrypt", "--passphrase", "secret")
	if err == nil {
		t.Fatal("expected error when --input is missing")
	}
}

func TestConfigEncryptMissingPassphrase(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "name: myproject\n")

	root := NewRootCommand()
	_, err := executeCommand(root, "config", "encrypt", "--input", in)
	if err == nil {
		t.Fatal("expected error when --passphrase is missing")
	}
}

func TestConfigEncryptNonexistentInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "encrypt", "--input", "/nonexistent/config.yaml", "--passphrase", "secret")
	if err == nil {
		t.Fatal("expected error for nonexistent input file")
	}
}

func TestConfigDecryptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "config.yaml")
	enc := filepath.Join(dir, "config.enc")
	out := filepath.Join(dir, "decrypted.yaml")
	writeTestFile(t, dir, "config.yaml", "name: myproject\n")

	root := NewRootCommand()
	if _, err := executeCommand(root, "config", "encrypt", "--input", in, "--passphrase", "secret", "--output", enc); err != nil {
		t.Fatalf("config encrypt failed: %v", err)
	}

	output, err := executeCommand(root, "config", "decrypt", "--input", enc, "--passphrase", "secret", "--output", out)
	if err != nil {
		t.Fatalf("config decrypt failed: %v", err)
	}
	if !strings.Contains(output, "Decrypted config written to") {
		t.Fatalf("expected write message, got %q", output)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read decrypted file: %v", err)
	}
	if string(data) != "name: myproject\n" {
		t.Fatalf("expected original content, got %q", string(data))
	}
}

func TestConfigDecryptStdout(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "config.yaml")
	enc := filepath.Join(dir, "config.enc")
	writeTestFile(t, dir, "config.yaml", "name: myproject\n")

	root := NewRootCommand()
	if _, err := executeCommand(root, "config", "encrypt", "--input", in, "--passphrase", "secret", "--output", enc); err != nil {
		t.Fatalf("config encrypt failed: %v", err)
	}

	output, err := executeCommand(root, "config", "decrypt", "--input", enc, "--passphrase", "secret")
	if err != nil {
		t.Fatalf("config decrypt failed: %v", err)
	}
	if !strings.Contains(output, "name: myproject") {
		t.Fatalf("expected decrypted content on stdout, got %q", output)
	}
}

func TestConfigDecryptWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "config.yaml")
	enc := filepath.Join(dir, "config.enc")
	writeTestFile(t, dir, "config.yaml", "name: myproject\n")

	root := NewRootCommand()
	if _, err := executeCommand(root, "config", "encrypt", "--input", in, "--passphrase", "secret", "--output", enc); err != nil {
		t.Fatalf("config encrypt failed: %v", err)
	}

	_, err := executeCommand(root, "config", "decrypt", "--input", enc, "--passphrase", "wrong")
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}

func TestConfigDecryptMissingInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "decrypt", "--passphrase", "secret")
	if err == nil {
		t.Fatal("expected error when --input is missing")
	}
}

func TestConfigValidateValid(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "naeos.yaml")
	writeTestFile(t, dir, "naeos.yaml", "name: myproject\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "validate", "--input", cfg)
	if err != nil {
		t.Fatalf("config validate failed: %v", err)
	}
	if !strings.Contains(output, "Config is valid") {
		t.Fatalf("expected 'Config is valid', got %q", output)
	}
}

func TestConfigValidateInvalid(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "bad.yaml")
	writeTestFile(t, dir, "bad.yaml", "port: banana\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "validate", "--input", cfg)
	if err != nil {
		t.Fatalf("config validate failed: %v", err)
	}
	if !strings.Contains(output, "validation error") {
		t.Fatalf("expected validation errors, got %q", output)
	}
}

func TestConfigValidateJSON(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	writeTestFile(t, dir, "config.json", `{"name":"myproject"}`)

	root := NewRootCommand()
	output, err := executeCommand(root, "config", "validate", "--input", cfg)
	if err != nil {
		t.Fatalf("config validate json failed: %v", err)
	}
	if !strings.Contains(output, "Config is valid") {
		t.Fatalf("expected 'Config is valid', got %q", output)
	}
}

func TestConfigValidateMissingInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "validate")
	if err == nil {
		t.Fatal("expected error when --input is missing")
	}
}

func TestConfigValidateNonexistentFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "config", "validate", "--input", "/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent config file")
	}
}

func TestConfigShow(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "config", "show")
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	if !strings.Contains(output, "NAEOS Configuration Schema") || !strings.Contains(output, "name") {
		t.Fatalf("expected schema output, got %q", output)
	}
}
