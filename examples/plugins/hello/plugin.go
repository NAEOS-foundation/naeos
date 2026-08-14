package main

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

// Plugin is the canonical minimal NAEOS plugin. Embed pluginhost.BasePlugin
// and override only the methods you need.
type Plugin struct {
	pluginhost.BasePlugin
}

// New constructs the plugin with its metadata.
func New() *Plugin {
	return &Plugin{
		BasePlugin: pluginhost.BasePlugin{
			NameVal:        "hello",
			VersionVal:     "0.1.0",
			DescriptionVal: "Canonical hello-world plugin: ping, describe, and greet.",
		},
	}
}

var (
	pluginName        = "hello"
	pluginVersion     = "0.1.0"
	pluginDescription = "Canonical hello-world plugin: ping, describe, and greet."
	pluginAuthor      = "NAEOS Foundation"

	PluginName        = &pluginName
	PluginVersion     = &pluginVersion
	PluginDescription = &pluginDescription
	PluginAuthor      = &pluginAuthor
)

var _ pluginhost.Plugin = (*Plugin)(nil)

var NaeosPlugin pluginhost.Plugin = New()

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

	case "greet":
		name, _ := params["name"].(string)
		if name == "" {
			name = "world"
		}
		return map[string]string{"message": fmt.Sprintf("Hello, %s!", name)}, nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
