// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pluginhost

import (
	"time"
)

// Plugin is the unified interface that all NAEOS plugins must implement.
type Plugin interface {
	Name() string
	Version() string
	Description() string
	Initialize(ctx *PluginContext) error
	Execute(action string, params map[string]any) (any, error)
	Shutdown() error
}

// PluginContext is the context passed to plugins during initialization.
type PluginContext struct {
	ConfigDir string
	OutputDir string
	Verbose   bool
	Config    map[string]any
	Logger    Logger
	Metrics   MetricsCollector
	EventBus  EventEmitter
}

// PluginState represents the lifecycle state of a plugin.
type PluginState string

const (
	StateCreated     PluginState = "created"
	StateInitialized PluginState = "initialized"
	StateRunning     PluginState = "running"
	StateStopped     PluginState = "stopped"
	StateError       PluginState = "error"
)

// PluginInfo holds metadata about a registered plugin.
type PluginInfo struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Description string      `json:"description"`
	Author      string      `json:"author,omitempty"`
	Path        string      `json:"path,omitempty"`
	Enabled     bool        `json:"enabled"`
	Loaded      bool        `json:"loaded"`
	State       PluginState `json:"state"`
	StartedAt   time.Time   `json:"started_at,omitempty"`
	Error       error       `json:"error,omitempty"`
}

// Manifest describes a plugin's capabilities, actions, and configuration schema.
type Manifest struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author,omitempty"`
	License      string                 `json:"license,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Actions      []ActionManifest       `json:"actions,omitempty"`
	Config       map[string]ConfigField `json:"config,omitempty"`
}

// ActionManifest describes a single action a plugin can perform.
type ActionManifest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Params      map[string]string `json:"params,omitempty"`
	Returns     string            `json:"returns,omitempty"`
}

// ConfigField describes a configuration field for a plugin.
type ConfigField struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
}

// BasePlugin is an embeddable struct that provides default implementations
// for the Plugin interface. Plugin authors can embed this and override only
// the methods they need.
type BasePlugin struct {
	NameVal        string
	VersionVal     string
	DescriptionVal string
}

func (b *BasePlugin) Name() string                      { return b.NameVal }
func (b *BasePlugin) Version() string                   { return b.VersionVal }
func (b *BasePlugin) Description() string               { return b.DescriptionVal }
func (b *BasePlugin) Initialize(_ *PluginContext) error { return nil }
func (b *BasePlugin) Shutdown() error                   { return nil }
