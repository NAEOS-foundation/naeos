// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"encoding/json"
	"fmt"
	"time"
)

type SARIFResult struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool         `json:"tool"`
	Results []SARIFResultItem `json:"results"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

type SARIFRule struct {
	ID                   string           `json:"id"`
	Name                 string           `json:"name"`
	ShortDescription     SARIFDescription `json:"shortDescription"`
	FullDescription      SARIFDescription `json:"fullDescription"`
	HelpURI              string           `json:"helpUri"`
	DefaultConfiguration SARIFConfig      `json:"defaultConfiguration"`
}

type SARIFDescription struct {
	Text string `json:"text"`
}

type SARIFConfig struct {
	Level string `json:"level"`
}

type SARIFResultItem struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   SARIFMessage    `json:"message"`
	Locations []SARIFLocation `json:"locations"`
}

type SARIFMessage struct {
	Text string `json:"text"`
}

type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

type SARIFRegion struct {
	StartLine int `json:"startLine"`
}

func severityToSARIFLevel(s Severity) string {
	switch s {
	case SeverityCritical, SeverityHigh:
		return "error"
	case SeverityMedium:
		return "warning"
	default:
		return "note"
	}
}

func GenerateSARIF(project string, result *AuditResult) ([]byte, error) {
	sarif := SARIFResult{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "naeos-security",
						Version:        "1.4.0",
						InformationURI: "https://github.com/NAEOS-foundation/naeos",
						Rules:          buildSARIFRules(result),
					},
				},
				Results: buildSARIFResults(result),
			},
		},
	}

	return json.MarshalIndent(sarif, "", "  ")
}

func buildSARIFRules(result *AuditResult) []SARIFRule {
	seen := make(map[string]bool)
	var rules []SARIFRule

	for _, f := range result.Finding {
		if seen[f.ID] {
			continue
		}
		seen[f.ID] = true

		rules = append(rules, SARIFRule{
			ID:   f.ID,
			Name: f.Title,
			ShortDescription: SARIFDescription{
				Text: f.Title,
			},
			FullDescription: SARIFDescription{
				Text: f.Description,
			},
			HelpURI: fmt.Sprintf("https://github.com/NAEOS-foundation/naeos/docs/security/%s", f.ID),
			DefaultConfiguration: SARIFConfig{
				Level: severityToSARIFLevel(f.Severity),
			},
		})
	}
	return rules
}

func buildSARIFResults(result *AuditResult) []SARIFResultItem {
	var items []SARIFResultItem

	for _, f := range result.Finding {
		item := SARIFResultItem{
			RuleID: f.ID,
			Level:  severityToSARIFLevel(f.Severity),
			Message: SARIFMessage{
				Text: f.Description,
			},
		}

		if f.File != "" {
			loc := SARIFLocation{
				PhysicalLocation: SARIFPhysicalLocation{
					ArtifactLocation: SARIFArtifactLocation{
						URI: f.File,
					},
				},
			}
			if f.Line > 0 {
				loc.PhysicalLocation.Region.StartLine = f.Line
			}
			item.Locations = append(item.Locations, loc)
		}

		items = append(items, item)
	}
	return items
}

func FormatSARIFTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
