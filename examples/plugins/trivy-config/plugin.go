// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

// Plugin runs Trivy configuration checks and returns structured findings.
// It is intentionally a native plugin because it invokes an external binary.
type Plugin struct {
	pluginhost.BasePlugin
}

func New() *Plugin {
	return &Plugin{BasePlugin: pluginhost.BasePlugin{
		NameVal:        "trivy-config",
		VersionVal:     "0.1.0",
		DescriptionVal: "Run Trivy configuration scans and return structured findings.",
	}}
}

var _ pluginhost.Plugin = (*Plugin)(nil)

type commandRunner func(context.Context, string, ...string) ([]byte, error)

var runTrivy commandRunner = func(ctx context.Context, target string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "trivy", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("trivy: %w: %s", err, stderr.String())
		}
		return nil, fmt.Errorf("trivy: %w", err)
	}
	return output, nil
}

type misconfiguration struct {
	Severity string `json:"Severity"`
}

type scanResult struct {
	Target            string             `json:"Target"`
	Type              string             `json:"Type"`
	Misconfigurations []misconfiguration `json:"Misconfigurations"`
}

type trivyReport struct {
	Results []scanResult `json:"Results"`
}

var blockingSeverities = map[string]bool{
	"CRITICAL": true,
	"HIGH":     true,
}

// Execute dispatches an action with the given parameters.
func (p *Plugin) Execute(action string, params map[string]any) (any, error) {
	switch action {
	case "ping":
		return map[string]string{"status": "ok"}, nil
	case "describe":
		return map[string]any{
			"name":        p.NameVal,
			"version":     p.VersionVal,
			"description": p.DescriptionVal,
			"runtime":     "native",
		}, nil
	case "scan":
		target, _ := params["target"].(string)
		if target == "" {
			return nil, fmt.Errorf("missing required param: target")
		}
		return scan(context.Background(), target)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func scan(ctx context.Context, target string) (map[string]any, error) {
	output, err := runTrivy(ctx, target, "config", "--format", "json", "--quiet", target)
	if err != nil {
		return nil, err
	}

	var report trivyReport
	if err := json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("parse Trivy JSON: %w", err)
	}

	severityCounts := map[string]int{}
	findingCount := 0
	blockingCount := 0
	for _, result := range report.Results {
		for _, finding := range result.Misconfigurations {
			findingCount++
			severityCounts[finding.Severity]++
			if blockingSeverities[finding.Severity] {
				blockingCount++
			}
		}
	}

	severities := make([]string, 0, len(severityCounts))
	for severity := range severityCounts {
		severities = append(severities, severity)
	}
	sort.Strings(severities)
	orderedCounts := make(map[string]int, len(severityCounts))
	for _, severity := range severities {
		orderedCounts[severity] = severityCounts[severity]
	}

	return map[string]any{
		"ok":              blockingCount == 0,
		"target":          target,
		"finding_count":   findingCount,
		"blocking_count":  blockingCount,
		"severity_counts": orderedCounts,
		"report":          report,
	}, nil
}
