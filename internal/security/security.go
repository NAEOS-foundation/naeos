package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var scanExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true,
	".java": true, ".rs": true, ".rb": true, ".php": true,
	".yaml": true, ".yml": true, ".json": true, ".toml": true,
	".env": true, ".sh": true, ".sql": true, ".html": true,
}

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
	".venv": true, "__pycache__": true, "target": true,
	".naeos": true,
}

func ScanDir(dir string) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if !scanExtensions[ext] {
			return nil
		}
		if info.Size() > 1<<20 {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		files[rel] = string(data)
		return nil
	})
	return files, err
}

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type Finding struct {
	ID          string
	Title       string
	Description string
	Severity    Severity
	File        string
	Line        int
	Remediation string
}

type AuditResult struct {
	Project string
	Finding []Finding
	Summary AuditSummary
}

type AuditSummary struct {
	Total    int
	Critical int
	High     int
	Medium   int
	Low      int
	Info     int
}

type Auditor struct {
	rules []AuditRule
}

type AuditRule struct {
	ID       string
	Severity Severity
	Check    func(filename, content string) []Finding
}

func NewAuditor() *Auditor {
	a := &Auditor{}
	a.rules = append(a.rules, defaultAuditRules()...)
	return a
}

func (a *Auditor) AddRule(rule AuditRule) {
	a.rules = append(a.rules, rule)
}

func (a *Auditor) Audit(filename, content string) []Finding {
	var findings []Finding
	for _, rule := range a.rules {
		f := rule.Check(filename, content)
		for i := range f {
			f[i].ID = rule.ID
			if f[i].Severity == "" {
				f[i].Severity = rule.Severity
			}
		}
		findings = append(findings, f...)
	}
	return findings
}

func (a *Auditor) AuditFiles(files map[string]string) *AuditResult {
	result := &AuditResult{}
	for name, content := range files {
		findings := a.Audit(name, content)
		result.Finding = append(result.Finding, findings...)
	}
	for _, f := range result.Finding {
		result.Summary.Total++
		switch f.Severity {
		case SeverityCritical:
			result.Summary.Critical++
		case SeverityHigh:
			result.Summary.High++
		case SeverityMedium:
			result.Summary.Medium++
		case SeverityLow:
			result.Summary.Low++
		case SeverityInfo:
			result.Summary.Info++
		}
	}
	return result
}

func defaultAuditRules() []AuditRule {
	return []AuditRule{
		{
			ID:       "hardcoded-secret",
			Severity: SeverityCritical,
			Check: func(filename, content string) []Finding {
				var findings []Finding
				secretPatterns := []string{
					"password=", "PASSWORD=", "password =",
					"api_key=", "API_KEY=", "api_key =",
					"secret=", "SECRET=", "secret =",
					"token=", "TOKEN=", "token =",
					"PRIVATE_KEY",
				}
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					for _, pattern := range secretPatterns {
						if strings.Contains(line, pattern) && !strings.Contains(line, "os.Getenv") && !strings.Contains(line, "process.env") && !strings.Contains(line, "os.environ") && !strings.Contains(line, "System.getenv") {
							findings = append(findings, Finding{
								Title:       "Potential hardcoded secret",
								Description: fmt.Sprintf("Line contains potential secret: %s", pattern),
								Severity:    SeverityCritical,
								File:        filename,
								Line:        i + 1,
								Remediation: "Use environment variables or a secrets manager instead of hardcoding secrets",
							})
						}
					}
				}
				return findings
			},
		},
		{
			ID:       "sql-injection",
			Severity: SeverityHigh,
			Check: func(filename, content string) []Finding {
				var findings []Finding
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					switch {
					case strings.HasSuffix(filename, ".go") && strings.Contains(line, "fmt.Sprintf") && strings.Contains(line, "SELECT"):
						findings = append(findings, Finding{
							Title:       "Potential SQL injection",
							Description: "String interpolation used in SQL query",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Use parameterized queries instead of string interpolation",
						})
					case strings.HasSuffix(filename, ".py") && strings.Contains(line, "f\"") && strings.Contains(strings.ToLower(line), "select"):
						findings = append(findings, Finding{
							Title:       "Potential SQL injection",
							Description: "f-string used in SQL query",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Use parameterized queries instead of f-strings",
						})
					case strings.HasSuffix(filename, ".ts") && strings.Contains(line, "`") && strings.Contains(strings.ToLower(line), "select"):
						findings = append(findings, Finding{
							Title:       "Potential SQL injection",
							Description: "Template literal used in SQL query",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Use parameterized queries instead of template literals",
						})
					case strings.HasSuffix(filename, ".java") && strings.Contains(line, "+") && strings.Contains(strings.ToUpper(line), "SELECT"):
						findings = append(findings, Finding{
							Title:       "Potential SQL injection",
							Description: "String concatenation used in SQL query",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Use PreparedStatement instead of string concatenation",
						})
					case strings.HasSuffix(filename, ".rs") && strings.Contains(line, "format!") && strings.Contains(strings.ToUpper(line), "SELECT"):
						findings = append(findings, Finding{
							Title:       "Potential SQL injection",
							Description: "format! macro used in SQL query",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Use parameterized queries instead of format! macro",
						})
					}
				}
				return findings
			},
		},
		{
			ID:       "insecure-listen",
			Severity: SeverityMedium,
			Check: func(filename, content string) []Finding {
				var findings []Finding
				if strings.Contains(content, "0.0.0.0") {
					findings = append(findings, Finding{
						Title:       "Binding to all interfaces",
						Description: "Server is configured to listen on all network interfaces",
						Severity:    SeverityMedium,
						File:        filename,
						Remediation: "Consider binding to a specific interface in production",
					})
				}
				return findings
			},
		},
		{
			ID:       "debug-mode",
			Severity: SeverityMedium,
			Check: func(filename, content string) []Finding {
				var findings []Finding
				debugPatterns := []string{
					"debug: true", "DEBUG=true", "debug=True",
					"DEBUG = true", "debugMode: true", "DEBUG_MODE=true",
				}
				for _, p := range debugPatterns {
					if strings.Contains(content, p) {
						findings = append(findings, Finding{
							Title:       "Debug mode enabled",
							Description: "Debug mode should not be enabled in production",
							Severity:    SeverityMedium,
							File:        filename,
							Remediation: "Disable debug mode in production deployments",
						})
						break
					}
				}
				return findings
			},
		},
		{
			ID:       "missing-health-check",
			Severity: SeverityLow,
			Check: func(filename, content string) []Finding {
				var findings []Finding
				hasServer := strings.Contains(content, "ListenAndServe") ||
					strings.Contains(content, "http.ListenAndServe") ||
					strings.Contains(content, "app.listen") ||
					strings.Contains(content, "@app.route") ||
					strings.Contains(content, "express()") ||
					strings.Contains(content, "HttpServer") ||
					strings.Contains(content, "actix_web::") ||
					strings.Contains(content, "hyper::")
				hasHealth := strings.Contains(content, "/health") || strings.Contains(content, "/ready") || strings.Contains(content, "healthz")
				if hasServer && !hasHealth {
					findings = append(findings, Finding{
						Title:       "Missing health check endpoint",
						Description: "HTTP server does not appear to have a health check endpoint",
						Severity:    SeverityLow,
						File:        filename,
						Remediation: "Add /health and /ready endpoints for container orchestration",
					})
				}
				return findings
			},
		},
		{
			ID:       "no-tls",
			Severity: SeverityInfo,
			Check: func(filename, content string) []Finding {
				var findings []Finding
				if strings.Contains(content, "ListenAndServe") && !strings.Contains(content, "ListenAndServeTLS") {
					findings = append(findings, Finding{
						Title:       "No TLS configuration",
						Description: "Server uses plain HTTP instead of HTTPS",
						Severity:    SeverityInfo,
						File:        filename,
						Remediation: "Consider using TLS in production or placing behind a reverse proxy",
					})
				}
				return findings
			},
		},
		{
			ID:       "eval-usage",
			Severity: SeverityHigh,
			Check: func(filename, content string) []Finding {
				var findings []Finding
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					trimmed := strings.TrimSpace(line)
					switch {
					case strings.HasSuffix(filename, ".py") && strings.Contains(trimmed, "eval("):
						findings = append(findings, Finding{
							Title:       "Use of eval()",
							Description: "eval() can execute arbitrary code",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Avoid eval(); use ast.literal_eval() for data parsing",
						})
					case (strings.HasSuffix(filename, ".js") || strings.HasSuffix(filename, ".ts")) && strings.Contains(trimmed, "eval(") && !strings.Contains(trimmed, "//"):
						findings = append(findings, Finding{
							Title:       "Use of eval()",
							Description: "eval() can execute arbitrary code",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Avoid eval(); use JSON.parse() or Function constructor with caution",
						})
					case strings.HasSuffix(filename, ".java") && strings.Contains(trimmed, "Runtime.getRuntime().exec("):
						findings = append(findings, Finding{
							Title:       "Command execution",
							Description: "Runtime.exec() can execute arbitrary system commands",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Validate and sanitize input before passing to exec()",
						})
					case strings.HasSuffix(filename, ".rs") && strings.Contains(trimmed, "std::process::Command"):
						findings = append(findings, Finding{
							Title:       "Command execution",
							Description: "Command::new() can execute arbitrary system commands",
							Severity:    SeverityHigh,
							File:        filename,
							Line:        i + 1,
							Remediation: "Validate and sanitize input before passing to Command",
						})
					}
				}
				return findings
			},
		},
		{
			ID:       "unsafe-deserialization",
			Severity: SeverityCritical,
			Check: func(filename, content string) []Finding {
				var findings []Finding
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					trimmed := strings.TrimSpace(line)
					switch {
					case strings.HasSuffix(filename, ".py") && strings.Contains(trimmed, "pickle.loads("):
						findings = append(findings, Finding{
							Title:       "Unsafe deserialization",
							Description: "pickle.loads() can execute arbitrary code during deserialization",
							Severity:    SeverityCritical,
							File:        filename,
							Line:        i + 1,
							Remediation: "Use json.loads() or a safe serialization format instead of pickle",
						})
					case strings.HasSuffix(filename, ".java") && strings.Contains(trimmed, "ObjectInputStream"):
						findings = append(findings, Finding{
							Title:       "Unsafe deserialization",
							Description: "ObjectInputStream can execute arbitrary code during deserialization",
							Severity:    SeverityCritical,
							File:        filename,
							Line:        i + 1,
							Remediation: "Use JSON or Protocol Buffers instead of Java serialization",
						})
					case strings.HasSuffix(filename, ".rs") && strings.Contains(trimmed, "bincode::deserialize"):
						findings = append(findings, Finding{
							Title:       "Unsafe deserialization risk",
							Description: "bincode::deserialize on untrusted data can be dangerous",
							Severity:    SeverityMedium,
							File:        filename,
							Line:        i + 1,
							Remediation: "Validate data source and consider using serde_json for untrusted input",
						})
					}
				}
				return findings
			},
		},
	}
}
