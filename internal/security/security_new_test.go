// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		input      string
		minEntropy float64
	}{
		{"aaaa", 0.0},
		{"abcdefghij", 2.0},
		{"sk-1234567890abcdef1234567890ab", 3.0},
		{"", 0.0},
	}

	for _, tt := range tests {
		entropy := shannonEntropy(tt.input)
		if entropy < tt.minEntropy {
			t.Errorf("shannonEntropy(%q) = %f, want >= %f", tt.input, entropy, tt.minEntropy)
		}
	}
}

func TestIsHighEntropySecret(t *testing.T) {
	tests := []struct {
		input     string
		threshold float64
		want      bool
	}{
		{"sk-1234567890abcdef1234567890ab", 3.0, true},
		{"aaaa", 3.0, false},
		{"short", 3.0, false},
		{"", 3.0, false},
	}

	for _, tt := range tests {
		got := isHighEntropySecret(tt.input, tt.threshold)
		if got != tt.want {
			t.Errorf("isHighEntropySecret(%q, %f) = %v, want %v", tt.input, tt.threshold, got, tt.want)
		}
	}
}

func TestSeverityFilter(t *testing.T) {
	filter := &SeverityFilter{MinSeverity: SeverityHigh}

	tests := []struct {
		severity Severity
		want     bool
	}{
		{SeverityCritical, true},
		{SeverityHigh, true},
		{SeverityMedium, false},
		{SeverityLow, false},
		{SeverityInfo, false},
	}

	for _, tt := range tests {
		got := filter.Matches(tt.severity)
		if got != tt.want {
			t.Errorf("SeverityFilter(%s).Matches(%s) = %v, want %v", filter.MinSeverity, tt.severity, got, tt.want)
		}
	}
}

func TestFilterBySeverity(t *testing.T) {
	findings := []Finding{
		{ID: "a", Severity: SeverityCritical},
		{ID: "b", Severity: SeverityHigh},
		{ID: "c", Severity: SeverityMedium},
		{ID: "d", Severity: SeverityLow},
	}

	filtered := FilterBySeverity(findings, SeverityHigh)
	if len(filtered) != 2 {
		t.Errorf("expected 2 findings, got %d", len(filtered))
	}
}

func TestGenerateSARIF(t *testing.T) {
	result := &AuditResult{
		Project: "test-project",
		Finding: []Finding{
			{
				ID:          "test-rule",
				Title:       "Test Finding",
				Description: "A test finding",
				Severity:    SeverityHigh,
				File:        "main.go",
				Line:        10,
			},
		},
		Summary: AuditSummary{Total: 1, High: 1},
	}

	sarifBytes, err := GenerateSARIF("test-project", result)
	if err != nil {
		t.Fatalf("GenerateSARIF failed: %v", err)
	}

	var sarif SARIFResult
	if err := json.Unmarshal(sarifBytes, &sarif); err != nil {
		t.Fatalf("failed to parse SARIF: %v", err)
	}

	if sarif.Version != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %s", sarif.Version)
	}

	if len(sarif.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(sarif.Runs))
	}

	if len(sarif.Runs[0].Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(sarif.Runs[0].Results))
	}

	if sarif.Runs[0].Results[0].RuleID != "test-rule" {
		t.Errorf("expected rule ID test-rule, got %s", sarif.Runs[0].Results[0].RuleID)
	}
}

func TestAuditXSS(t *testing.T) {
	auditor := NewAuditor()

	content := `const html = dangerouslySetInnerHTML={{ __html: userInput }}`
	findings := auditor.Audit("app.jsx", content)
	found := false
	for _, f := range findings {
		if f.ID == "xss-vulnerability" {
			found = true
		}
	}
	if !found {
		t.Error("expected XSS finding for dangerouslySetInnerHTML")
	}
}

func TestAuditInsecureHTTP(t *testing.T) {
	auditor := NewAuditor()

	content := `fetch("http://api.mysite.com/data")`
	findings := auditor.Audit("app.js", content)
	found := false
	for _, f := range findings {
		if f.ID == "insecure-http" {
			found = true
		}
	}
	if !found {
		t.Error("expected insecure HTTP finding")
	}
}

func TestAuditInsecureHTTPExempt(t *testing.T) {
	auditor := NewAuditor()

	content := `fetch("http://localhost:8080/api")`
	findings := auditor.Audit("app.js", content)
	for _, f := range findings {
		if f.ID == "insecure-http" {
			t.Error("should not flag localhost HTTP as insecure")
		}
	}
}

func TestAuditConnectionString(t *testing.T) {
	auditor := NewAuditor()

	content := `connStr := "postgresql://user:pass@localhost/db"`
	findings := auditor.Audit("db.go", content)
	found := false
	for _, f := range findings {
		if f.ID == "hardcoded-connection-string" {
			found = true
		}
	}
	if !found {
		t.Error("expected hardcoded connection string finding")
	}
}

func TestAuditConnectionStringEnvExempt(t *testing.T) {
	auditor := NewAuditor()

	content := `connStr := os.Getenv("DATABASE_URL")`
	findings := auditor.Audit("db.go", content)
	for _, f := range findings {
		if f.ID == "hardcoded-connection-string" {
			t.Error("should not flag env var usage as hardcoded connection string")
		}
	}
}

func TestAuditWeakCrypto(t *testing.T) {
	auditor := NewAuditor()

	content := `hash := md5.New()`
	findings := auditor.Audit("crypto.go", content)
	found := false
	for _, f := range findings {
		if f.ID == "weak-crypto" {
			found = true
		}
	}
	if !found {
		t.Error("expected weak crypto finding for MD5")
	}
}

func TestLoadCustomRules(t *testing.T) {
	rulesYAML := `
rules:
  - id: custom-rule
    title: Custom Check
    description: A custom security check
    severity: high
    patterns:
      - "dangerous_function"
    extensions:
      - .go
      - .py
`
	dir := t.TempDir()
	rulesPath := filepath.Join(dir, "rules.yaml")
	os.WriteFile(rulesPath, []byte(rulesYAML), 0o644)

	rules, err := LoadCustomRules(rulesPath)
	if err != nil {
		t.Fatalf("LoadCustomRules failed: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	if rules[0].ID != "custom-rule" {
		t.Errorf("expected rule ID custom-rule, got %s", rules[0].ID)
	}

	content := `result = dangerous_function(input)`
	findings := rules[0].Check("app.py", content)
	if len(findings) == 0 {
		t.Error("expected findings from custom rule")
	}
}

func TestLoadCustomRulesFileNotFound(t *testing.T) {
	_, err := LoadCustomRules("/nonexistent/rules.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestAuditWithFilter(t *testing.T) {
	auditor := NewAuditor()
	content := `password=secret123`

	criticalFindings := auditor.AuditWithFilter("config.go", content, SeverityCritical)
	highFindings := auditor.AuditWithFilter("config.go", content, SeverityHigh)

	if len(criticalFindings) == 0 {
		t.Error("expected critical findings")
	}
	if len(highFindings) == 0 {
		t.Error("expected high findings")
	}
}

func TestAuditFilesWithFilter(t *testing.T) {
	auditor := NewAuditor()
	files := map[string]string{
		"main.go":  `password=secret123`,
		"other.go": `http.ListenAndServe("0.0.0.0:8080", nil)`,
	}

	result := auditor.AuditFilesWithFilter(files, SeverityCritical)
	if result.Summary.Critical == 0 {
		t.Error("expected at least 1 critical finding")
	}
	if result.Summary.Medium > 0 {
		t.Error("expected no medium findings with critical filter")
	}
}

func TestScanDirLargeFile(t *testing.T) {
	dir := t.TempDir()

	largeContent := make([]byte, 2<<20)
	for i := range largeContent {
		largeContent[i] = 'a'
	}
	os.WriteFile(filepath.Join(dir, "large.go"), largeContent, 0o644)

	files, err := ScanDir(dir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if _, ok := files["large.go"]; ok {
		t.Error("large files should be skipped")
	}
}

func TestScanDirSubdirectories(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("test"), 0o644)
	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main"), 0o644)

	files, err := ScanDir(dir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if _, ok := files[".git/config"]; ok {
		t.Error(".git directory should be skipped")
	}
	if _, ok := files["src/main.go"]; !ok {
		t.Error("src/main.go should be scanned")
	}
}
