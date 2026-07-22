package lint

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var specProjectStartRe = regexp.MustCompile(`^[a-zA-Z0-9]`)
var projectNameFormatRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityError    Severity = "error"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

type LintIssue struct {
	Line     int
	Column   int
	Severity Severity
	Rule     string
	Message  string
}

type LintResult struct {
	Path   string
	Issues []LintIssue
}

type Linter struct {
	rules []LintRule
}

type LintRule struct {
	ID       string
	Severity Severity
	Check    func(content string) []LintIssue
}

func NewLinter() *Linter {
	l := &Linter{}
	l.rules = append(l.rules, defaultRules()...)
	return l
}

func (l *Linter) AddRule(rule LintRule) {
	l.rules = append(l.rules, rule)
}

func (l *Linter) Lint(path, content string) *LintResult {
	result := &LintResult{Path: path}
	for _, rule := range l.rules {
		issues := rule.Check(content)
		for i := range issues {
			issues[i].Rule = rule.ID
			if issues[i].Severity == "" {
				issues[i].Severity = rule.Severity
			}
		}
		result.Issues = append(result.Issues, issues...)
	}
	return result
}

func (l *Linter) LintFile(path string) (*LintResult, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return l.Lint(path, string(data)), nil
}

func defaultRules() []LintRule {
	return append(basicYAMLRules(), specValidationRules()...)
}

func basicYAMLRules() []LintRule {
	return []LintRule{
		{
			ID:       "yaml-tabs",
			Severity: SeverityError,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					if strings.Contains(line, "\t") && (strings.HasSuffix(strings.TrimSpace(line), ":") || strings.Contains(line, "  ")) {
						issues = append(issues, LintIssue{
							Line:     i + 1,
							Severity: SeverityError,
							Message:  "use spaces instead of tabs in YAML",
						})
					}
				}
				return issues
			},
		},
		{
			ID:       "yaml-trailing-space",
			Severity: SeverityWarning,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
						issues = append(issues, LintIssue{
							Line:     i + 1,
							Severity: SeverityWarning,
							Message:  "trailing whitespace",
						})
					}
				}
				return issues
			},
		},
		{
			ID:       "project-name-format",
			Severity: SeverityWarning,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				inProject := false
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "project:") {
						inProject = true
						parts := strings.SplitN(trimmed, ":", 2)
						if len(parts) == 2 {
							name := strings.TrimSpace(parts[1])
							if name == "" {
								issues = append(issues, LintIssue{
									Severity: SeverityWarning,
									Rule:     "project-name-format",
									Message:  "project name is empty",
								})
							} else if matched := projectNameFormatRe.MatchString(name); !matched {
								issues = append(issues, LintIssue{
									Severity: SeverityWarning,
									Rule:     "project-name-format",
									Message:  fmt.Sprintf("project name %q should be lowercase alphanumeric with hyphens", name),
								})
							}
						}
					}
					if inProject && !strings.HasPrefix(trimmed, " ") && strings.Contains(trimmed, ":") {
						inProject = false
					}
				}
				return issues
			},
		},
		{
			ID:       "module-path-format",
			Severity: SeverityInfo,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "path:") {
						parts := strings.SplitN(trimmed, ":", 2)
						if len(parts) == 2 {
							path := strings.TrimSpace(parts[1])
							if path != "" && !strings.HasPrefix(path, "./") {
								issues = append(issues, LintIssue{
									Severity: SeverityInfo,
									Rule:     "module-path-format",
									Message:  fmt.Sprintf("module path %q should start with ./", path),
								})
							}
						}
					}
				}
				return issues
			},
		},
		{
			ID:       "port-range",
			Severity: SeverityError,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "port:") {
						parts := strings.SplitN(trimmed, ":", 2)
						if len(parts) == 2 {
							portStr := strings.TrimSpace(parts[1])
							var port int
							if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
								if port < 1 || port > 65535 {
									issues = append(issues, LintIssue{
										Line:     i + 1,
										Severity: SeverityError,
										Rule:     "port-range",
										Message:  fmt.Sprintf("port %d is out of valid range (1-65535)", port),
									})
								}
							}
						}
					}
				}
				return issues
			},
		},
		{
			ID:       "duplicate-keys",
			Severity: SeverityWarning,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				seen := make(map[string]int)
				for i, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.HasSuffix(trimmed, ":") {
						key := strings.TrimSuffix(trimmed, ":")
						key = strings.TrimSpace(key)
						if prev, ok := seen[key]; ok {
							issues = append(issues, LintIssue{
								Line:     i + 1,
								Severity: SeverityWarning,
								Rule:     "duplicate-keys",
								Message:  fmt.Sprintf("duplicate key %q (first seen at line %d)", key, prev),
							})
						}
						seen[key] = i + 1
					}
				}
				return issues
			},
		},
		{
			ID:       "empty-document",
			Severity: SeverityError,
			Check: func(content string) []LintIssue {
				if strings.TrimSpace(content) == "" {
					return []LintIssue{{
						Severity: SeverityError,
						Rule:     "empty-document",
						Message:  "specification document is empty",
					}}
				}
				return nil
			},
		},
		{
			ID:       "yaml-line-length",
			Severity: SeverityWarning,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				const maxLineLength = 200
				for i, line := range lines {
					if len(line) > maxLineLength {
						issues = append(issues, LintIssue{
							Line:     i + 1,
							Severity: SeverityWarning,
							Rule:     "yaml-line-length",
							Message:  fmt.Sprintf("line exceeds %d characters (%d)", maxLineLength, len(line)),
						})
					}
				}
				return issues
			},
		},
		{
			ID:       "yaml-nested-duplicate-keys",
			Severity: SeverityWarning,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				type keyEntry struct {
					key   string
					line  int
					depth int
				}
				var stack []keyEntry

				for i, line := range lines {
					trimmed := strings.TrimLeft(line, " \t")
					if strings.TrimSpace(trimmed) == "" {
						continue
					}

					indent := len(line) - len(trimmed)
					depth := indent / 2

					colonIdx := strings.Index(trimmed, ":")
					if colonIdx < 0 {
						continue
					}
					key := strings.TrimSpace(trimmed[:colonIdx])
					if key == "" {
						continue
					}

					for j := len(stack) - 1; j >= 0; j-- {
						entry := stack[j]
						if entry.depth > depth {
							continue
						}
						if entry.depth < depth {
							break
						}
						if entry.depth == depth && entry.key == key {
							issues = append(issues, LintIssue{
								Line:     i + 1,
								Severity: SeverityWarning,
								Rule:     "yaml-nested-duplicate-keys",
								Message:  fmt.Sprintf("duplicate key %q at same indentation level (first at line %d)", key, entry.line),
							})
							break
						}
					}

					for len(stack) > 0 && stack[len(stack)-1].depth >= depth {
						stack = stack[:len(stack)-1]
					}
					stack = append(stack, keyEntry{key: key, line: i + 1, depth: depth})
				}
				return issues
			},
		},
		{
			ID:       "yaml-bool-string-mix",
			Severity: SeverityWarning,
			Check: func(content string) []LintIssue {
				var issues []LintIssue
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.Contains(trimmed, ":") {
						parts := strings.SplitN(trimmed, ":", 2)
						if len(parts) == 2 {
							val := strings.TrimSpace(parts[1])
							if val == "\"true\"" || val == "\"false\"" || val == "'true'" || val == "'false'" {
								issues = append(issues, LintIssue{
									Line:     i + 1,
									Severity: SeverityWarning,
									Rule:     "yaml-bool-string-mix",
									Message:  "boolean value is quoted — use unquoted true/false",
								})
							}
						}
					}
				}
				return issues
			},
		},
	}
}

func specValidationRules() []LintRule {
	return []LintRule{}
}

func Fix(content string) string {
	lines := strings.Split(content, "\n")
	var fixed []string
	for _, line := range lines {
		fixed = append(fixed, strings.TrimRight(line, " \t"))
	}
	return strings.Join(fixed, "\n")
}

func readFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	return data, nil
}

func ValidateSpec(content string) []LintIssue {
	var issues []LintIssue

	if strings.TrimSpace(content) == "" {
		return []LintIssue{{Severity: SeverityError, Rule: "spec-empty", Message: "specification is empty"}}
	}

	var raw map[string]any
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return []LintIssue{{Severity: SeverityError, Rule: "spec-yaml", Message: fmt.Sprintf("invalid YAML: %v", err)}}
	}

	if raw == nil {
		return []LintIssue{{Severity: SeverityError, Rule: "spec-empty", Message: "specification parsed to nil"}}
	}

	if _, ok := raw["project"]; !ok {
		issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-required-project", Message: "missing required field: project"})
	} else if name, ok := raw["project"].(string); ok {
		if strings.TrimSpace(name) == "" {
			issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-project-empty", Message: "project name is empty"})
		} else if matched := specProjectStartRe.MatchString(name); !matched {
			issues = append(issues, LintIssue{Severity: SeverityWarning, Rule: "spec-project-format", Message: fmt.Sprintf("project name %q should start with alphanumeric character", name)})
		}
	}

	if modules, ok := raw["modules"]; ok {
		if moduleList, ok := modules.([]any); ok {
			if len(moduleList) == 0 {
				issues = append(issues, LintIssue{Severity: SeverityWarning, Rule: "spec-modules-empty", Message: "modules list is empty"})
			}
			seen := make(map[string]int)
			for i, m := range moduleList {
				if mod, ok := m.(map[string]any); ok {
					name, _ := mod["name"].(string)
					if name == "" {
						issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-module-name", Message: fmt.Sprintf("module[%d] missing required field: name", i)})
					} else {
						if prev, dup := seen[name]; dup {
							issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-module-duplicate", Message: fmt.Sprintf("duplicate module name %q at indices %d and %d", name, prev, i)})
						}
						seen[name] = i
					}
					if path, ok := mod["path"].(string); ok && path == "" {
						issues = append(issues, LintIssue{Severity: SeverityWarning, Rule: "spec-module-path", Message: fmt.Sprintf("module %q has empty path", name)})
					}
				} else {
					issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-module-type", Message: fmt.Sprintf("module[%d] must be a mapping, got %T", i, m)})
				}
			}
		} else {
			issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-modules-type", Message: "field 'modules' must be a list"})
		}
	}

	if services, ok := raw["services"]; ok {
		if serviceList, ok := services.([]any); ok {
			seenPorts := make(map[int]string)
			for i, s := range serviceList {
				if svc, ok := s.(map[string]any); ok {
					name, _ := svc["name"].(string)
					if name == "" {
						issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-service-name", Message: fmt.Sprintf("service[%d] missing required field: name", i)})
					}
					if port, ok := svc["port"].(int); ok {
						if port < 1 || port > 65535 {
							issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-service-port", Message: fmt.Sprintf("service %q port %d is out of range (1-65535)", name, port)})
						}
						if prev, exists := seenPorts[port]; exists {
							issues = append(issues, LintIssue{Severity: SeverityWarning, Rule: "spec-service-port-dup", Message: fmt.Sprintf("services %q and %q share port %d", prev, name, port)})
						}
						seenPorts[port] = name
					}
				} else {
					issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-service-type", Message: fmt.Sprintf("service[%d] must be a mapping, got %T", i, s)})
				}
			}
		} else {
			issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-services-type", Message: "field 'services' must be a list"})
		}
	}

	if gen, ok := raw["generation"]; ok {
		if genMap, ok := gen.(map[string]any); ok {
			if langs, ok := genMap["languages"]; ok {
				if langList, ok := langs.([]any); ok {
					validLangs := map[string]bool{"go": true, "typescript": true, "python": true, "java": true, "rust": true}
					for _, l := range langList {
						if lang, ok := l.(string); ok {
							if !validLangs[lang] {
								issues = append(issues, LintIssue{Severity: SeverityError, Rule: "spec-language-invalid", Message: fmt.Sprintf("unsupported language %q — supported: go, typescript, python, java, rust", lang)})
							}
						}
					}
				}
			}
		}
	}

	return issues
}
