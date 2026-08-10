package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/audit"
)

const testAuditLine = `{"id":"e1","timestamp":"2025-01-01T00:00:00Z","user_id":"u1","action":"read","resource":"spec","status":"success"}`

func TestComplianceExportJSON(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")
	out := filepath.Join(dir, "export.json")
	writeTestFile(t, dir, "audit.log", testAuditLine)

	root := NewRootCommand()
	output, err := executeCommand(root, "compliance", "export", "--format", "json", "--output", out, "--audit-file", logFile)
	if err != nil {
		t.Fatalf("compliance export failed: %v", err)
	}
	if !strings.Contains(output, "Compliance report exported") {
		t.Fatalf("expected export message, got %q", output)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if !strings.Contains(string(data), "e1") {
		t.Fatalf("expected event in export, got %q", string(data))
	}
}

func TestComplianceExportCSV(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")
	out := filepath.Join(dir, "export.csv")
	writeTestFile(t, dir, "audit.log", testAuditLine)

	root := NewRootCommand()
	output, err := executeCommand(root, "compliance", "export", "--format", "csv", "--output", out, "--audit-file", logFile)
	if err != nil {
		t.Fatalf("compliance export csv failed: %v", err)
	}
	if !strings.Contains(output, "exported to") {
		t.Fatalf("expected export message, got %q", output)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected csv export file: %v", err)
	}
}

func TestComplianceExportUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "export.txt")

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "export", "--format", "xml", "--output", out, "--audit-file", filepath.Join(dir, "audit.log"))
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestComplianceExportMissingOutput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "export")
	if err == nil {
		t.Fatal("expected error when --output is missing")
	}
}

func TestComplianceExportMissingAuditFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "export.json")
	missing := filepath.Join(dir, "does-not-exist.log")

	root := NewRootCommand()
	if _, err := executeCommand(root, "compliance", "export", "--output", out, "--audit-file", missing); err != nil {
		t.Fatalf("export with missing audit file should succeed: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected export file: %v", err)
	}
}

func TestComplianceExportMalformedLine(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")
	writeTestFile(t, dir, "audit.log", "this is not json")

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "export", "--output", filepath.Join(dir, "out.json"), "--audit-file", logFile)
	if err == nil {
		t.Fatal("expected error for malformed audit line")
	}
}

func TestComplianceReportSOC2(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")
	writeTestFile(t, dir, "audit.log", testAuditLine)

	root := NewRootCommand()
	output, err := executeCommand(root, "compliance", "report", "--framework", "soc2", "--audit-file", logFile)
	if err != nil {
		t.Fatalf("compliance report failed: %v", err)
	}
	if !strings.Contains(output, "Compliance Report: SOC 2 Type II") {
		t.Fatalf("expected SOC 2 report header, got %q", output)
	}
}

func TestComplianceReportUnsupportedFramework(t *testing.T) {
	dir := t.TempDir()

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "report", "--framework", "iso27001", "--audit-file", filepath.Join(dir, "audit.log"))
	if err == nil {
		t.Fatal("expected error for unsupported framework")
	}
}

func TestComplianceReportOutputFile(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")
	out := filepath.Join(dir, "report.json")
	writeTestFile(t, dir, "audit.log", testAuditLine)

	root := NewRootCommand()
	output, err := executeCommand(root, "compliance", "report", "--framework", "gdpr", "--output", out, "--audit-file", logFile)
	if err != nil {
		t.Fatalf("compliance report failed: %v", err)
	}
	if !strings.Contains(output, "Report saved to") {
		t.Fatalf("expected 'Report saved to', got %q", output)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected report file: %v", err)
	}
}

func TestComplianceListFrameworks(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "compliance", "list-frameworks")
	if err != nil {
		t.Fatalf("compliance list-frameworks failed: %v", err)
	}
	for _, want := range []string{"soc2", "hipaa", "gdpr"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected framework %q in output, got %q", want, output)
		}
	}
}

func TestComplianceVerifyValidChain(t *testing.T) {
	dir := t.TempDir()
	mem := audit.NewMemoryAuditor()
	h := audit.NewHashedAuditor(mem)
	if err := h.Log(audit.AuditEvent{ID: "e1", Action: "login", Status: "success"}); err != nil {
		t.Fatalf("log event: %v", err)
	}
	if err := h.Log(audit.AuditEvent{ID: "e2", Action: "read", Status: "success"}); err != nil {
		t.Fatalf("log event: %v", err)
	}
	var sb strings.Builder
	for _, e := range mem.Events() {
		data, _ := json.Marshal(e)
		sb.Write(data)
		sb.WriteByte('\n')
	}
	logFile := filepath.Join(dir, "audit.log")
	writeTestFile(t, dir, "audit.log", sb.String())

	root := NewRootCommand()
	output, err := executeCommand(root, "compliance", "verify", "--audit-file", logFile)
	if err != nil {
		t.Fatalf("compliance verify failed: %v", err)
	}
	if !strings.Contains(output, "no violations found") {
		t.Fatalf("expected 'no violations found', got %q", output)
	}
}

func TestComplianceVerifyTampered(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "audit.log",
		`{"id":"e1","timestamp":"2025-01-01T00:00:00Z","action":"read","status":"success","previous_hash":"","hash":"deadbeef"}`)

	root := NewRootCommand()
	output, err := executeCommand(root, "compliance", "verify", "--audit-file", filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Fatalf("compliance verify failed: %v", err)
	}
	if !strings.Contains(output, "violation") {
		t.Fatalf("expected hash chain violation, got %q", output)
	}
}

func TestComplianceVerifyMissingFile(t *testing.T) {
	dir := t.TempDir()

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "verify", "--audit-file", filepath.Join(dir, "nope.log"))
	if err == nil {
		t.Fatal("expected error when audit file is missing")
	}
}

func TestComplianceCloudExportUnsupportedProvider(t *testing.T) {
	dir := t.TempDir()

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "cloud-export", "--provider", "dropbox", "--bucket", "b", "--audit-file", filepath.Join(dir, "audit.log"))
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestComplianceCloudExportAzureMissingAccount(t *testing.T) {
	dir := t.TempDir()

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "cloud-export", "--provider", "azure", "--bucket", "container", "--audit-file", filepath.Join(dir, "audit.log"))
	if err == nil {
		t.Fatal("expected error when Azure account name is missing")
	}
}

func TestComplianceCloudExportAzureMissingKey(t *testing.T) {
	dir := t.TempDir()

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "cloud-export", "--provider", "azure", "--bucket", "container", "--account-name", "acct", "--audit-file", filepath.Join(dir, "audit.log"))
	if err == nil {
		t.Fatal("expected error when Azure account key is missing")
	}
}

func TestComplianceCloudExportAzureMissingBucket(t *testing.T) {
	dir := t.TempDir()

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "cloud-export", "--provider", "azure", "--account-name", "acct", "--account-key", "key", "--audit-file", filepath.Join(dir, "audit.log"))
	if err == nil {
		t.Fatal("expected error when Azure container is missing")
	}
}

func TestComplianceCloudExportMalformedAudit(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")
	writeTestFile(t, dir, "audit.log", "garbage")

	root := NewRootCommand()
	_, err := executeCommand(root, "compliance", "cloud-export", "--provider", "s3", "--bucket", "b", "--audit-file", logFile)
	if err == nil {
		t.Fatal("expected error for malformed audit file")
	}
}
