// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type RuleConfig struct {
	Rules []CustomRule `yaml:"rules"`
}

type CustomRule struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Severity    string   `yaml:"severity"`
	Patterns    []string `yaml:"patterns"`
	Extensions  []string `yaml:"extensions"`
}

type SeverityFilter struct {
	MinSeverity Severity
}

func (f *SeverityFilter) Matches(s Severity) bool {
	levels := map[Severity]int{
		SeverityInfo:     0,
		SeverityLow:      1,
		SeverityMedium:   2,
		SeverityHigh:     3,
		SeverityCritical: 4,
	}
	return levels[s] >= levels[f.MinSeverity]
}

func FilterBySeverity(findings []Finding, minSeverity Severity) []Finding {
	filter := &SeverityFilter{MinSeverity: minSeverity}
	var filtered []Finding
	for _, f := range findings {
		if filter.Matches(f.Severity) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func LoadCustomRules(path string) ([]AuditRule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	var rules []AuditRule
	for _, cr := range config.Rules {
		rule := AuditRule{
			ID:       cr.ID,
			Severity: Severity(cr.Severity),
			Check:    buildPatternCheck(cr),
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func buildPatternCheck(cr CustomRule) func(string, string) []Finding {
	extSet := make(map[string]bool)
	for _, ext := range cr.Extensions {
		extSet[ext] = true
	}

	return func(filename, content string) []Finding {
		if len(extSet) > 0 {
			matched := false
			for ext := range extSet {
				if strings.HasSuffix(filename, ext) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		var findings []Finding
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			for _, pattern := range cr.Patterns {
				if strings.Contains(line, pattern) {
					findings = append(findings, Finding{
						Title:       cr.Title,
						Description: cr.Description,
						Severity:    Severity(cr.Severity),
						File:        filename,
						Line:        i + 1,
						Remediation: "Review and fix according to security policy",
					})
					break
				}
			}
		}
		return findings
	}
}
