package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

// Plugin computes aggregate statistics over generated artifacts. It
// demonstrates processing batched inputs (file contents passed via params)
// and returning a JSON-friendly report.
type Plugin struct {
	pluginhost.BasePlugin
}

// New constructs the plugin with its metadata.
func New() *Plugin {
	return &Plugin{
		BasePlugin: pluginhost.BasePlugin{
			NameVal:        "artifact-stats",
			VersionVal:     "0.1.0",
			DescriptionVal: "Compute line/byte statistics over generated artifacts.",
		},
	}
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
		}, nil

	case "stats":
		raw, ok := params["files"].([]any)
		if !ok || len(raw) == 0 {
			return nil, fmt.Errorf("missing or empty required param: files")
		}
		return computeStats(raw), nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// computeStats aggregates line and byte counts per file extension.
func computeStats(raw []any) map[string]any {
	var totalFiles, totalLines, totalBytes int
	byExt := map[string]map[string]int{}

	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		path, _ := m["path"].(string)
		content, _ := m["content"].(string)

		totalFiles++
		lines := len(strings.Split(content, "\n"))
		totalLines += lines
		totalBytes += len(content)

		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = "(none)"
		}
		stats := byExt[ext]
		if stats == nil {
			stats = map[string]int{"files": 0, "lines": 0, "bytes": 0}
			byExt[ext] = stats
		}
		stats["files"]++
		stats["lines"] += lines
		stats["bytes"] += len(content)
	}

	return map[string]any{
		"files":  totalFiles,
		"lines":  totalLines,
		"bytes":  totalBytes,
		"by_ext": byExt,
	}
}
