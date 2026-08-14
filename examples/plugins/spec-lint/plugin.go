package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

// Plugin lints NAEOS specification content for basic naming and structural
// conventions. It demonstrates how a plugin can process pipeline input
// (spec text passed via params) and return a structured report.
type Plugin struct {
	pluginhost.BasePlugin
}

// New constructs the plugin with its metadata.
func New() *Plugin {
	return &Plugin{
		BasePlugin: pluginhost.BasePlugin{
			NameVal:        "spec-lint",
			VersionVal:     "0.1.0",
			DescriptionVal: "Lint NAEOS specifications for naming and structural conventions.",
		},
	}
}

var (
	pluginName        = "spec-lint"
	pluginVersion     = "0.1.0"
	pluginDescription = "Lint NAEOS specifications for naming and structural conventions."
	pluginAuthor      = "NAEOS Foundation"

	PluginName        = &pluginName
	PluginVersion     = &pluginVersion
	PluginDescription = &pluginDescription
	PluginAuthor      = &pluginAuthor
)

var _ pluginhost.Plugin = (*Plugin)(nil)

var NaeosPlugin pluginhost.Plugin = New()

var (
	reModuleName = regexp.MustCompile(`^\s*-\s+name:\s*(\S+)`)
	rePortValue  = regexp.MustCompile(`port:\s*(\d+)`)
)

type violation struct {
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	Detail   string `json:"detail"`
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

	case "lint":
		spec, _ := params["spec"].(string)
		if strings.TrimSpace(spec) == "" {
			return nil, fmt.Errorf("missing required param: spec")
		}
		return lintSpec(spec), nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// lintSpec applies a small set of convention checks to the spec YAML text.
func lintSpec(spec string) map[string]any {
	violations := []violation{}
	lines := strings.Split(spec, "\n")

	for i, line := range lines {
		if m := reModuleName.FindStringSubmatch(line); m != nil {
			name := m[1]
			if name != strings.ToLower(name) {
				violations = append(violations, violation{
					Severity: "warning",
					Rule:     "module-name-case",
					Detail:   fmt.Sprintf("line %d: module name %q must be lowercase", i+1, name),
				})
			}
			if strings.Contains(name, " ") || strings.Contains(name, "_") {
				violations = append(violations, violation{
					Severity: "warning",
					Rule:     "module-name-format",
					Detail:   fmt.Sprintf("line %d: module name %q must use kebab-case", i+1, name),
				})
			}
		}

		if m := rePortValue.FindStringSubmatch(line); m != nil {
			port := m[1]
			if port != strings.TrimLeft(port, "0") {
				violations = append(violations, violation{
					Severity: "warning",
					Rule:     "port-leading-zero",
					Detail:   fmt.Sprintf("line %d: port %q has a leading zero", i+1, port),
				})
			}
		}
	}

	ok := len(violations) == 0
	return map[string]any{
		"ok":          ok,
		"violations":  violations,
		"checks":      len(lines),
		"description": "Naming and structural convention checks",
	}
}
