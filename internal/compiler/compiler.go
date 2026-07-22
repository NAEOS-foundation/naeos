package compiler

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/promptlib"
)

type Target string

const (
	TargetCopilot  Target = "copilot"
	TargetClaude   Target = "claude"
	TargetCursor   Target = "cursor"
	TargetGemini   Target = "gemini"
	TargetCodex    Target = "codex"
	TargetOpenCode Target = "opencode"
	TargetWindsurf Target = "windsurf"
)

type CompiledOutput struct {
	Target      Target       `json:"target"`
	Files       []OutputFile `json:"files"`
	Summary     string       `json:"summary"`
	CompiledAt  time.Time    `json:"compiled_at"`
	NEIRVersion string       `json:"neir_version,omitempty"`
}

type OutputFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Kind    string `json:"kind"`
}

type Adapter interface {
	Target() Target
	Compile(neir *model.NEIR) (*CompiledOutput, error)
}

type Compiler struct {
	adapters map[Target]Adapter
	library  *promptlib.Library
}

func New() *Compiler {
	return &Compiler{
		adapters: make(map[Target]Adapter),
	}
}

// NewWithLibrary creates a Compiler with a prompt library for template-based compilation.
func NewWithLibrary(lib *promptlib.Library) *Compiler {
	return &Compiler{
		adapters: make(map[Target]Adapter),
		library:  lib,
	}
}

// Library returns the prompt library associated with this compiler, or nil.
func (c *Compiler) Library() *promptlib.Library {
	return c.library
}

func (c *Compiler) Register(a Adapter) {
	c.adapters[a.Target()] = a
}

func (c *Compiler) Compile(neir *model.NEIR, target Target) (*CompiledOutput, error) {
	a, ok := c.adapters[target]
	if !ok {
		slog.Error("unknown compile target", "target", target)
		return nil, fmt.Errorf("unknown target: %s", target)
	}
	return a.Compile(neir)
}

func (c *Compiler) CompileAll(neir *model.NEIR) map[Target]*CompiledOutput {
	results := make(map[Target]*CompiledOutput)
	for target, a := range c.adapters {
		out, err := a.Compile(neir)
		if err != nil {
			results[target] = &CompiledOutput{
				Target:  target,
				Summary: fmt.Sprintf("error: %v", err),
			}
		} else {
			results[target] = out
		}
	}
	return results
}

func (c *Compiler) Targets() []Target {
	targets := make([]Target, 0, len(c.adapters))
	for t := range c.adapters {
		targets = append(targets, t)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i] < targets[j] })
	return targets
}

func buildProjectContext(neir *model.NEIR) string {
	var sb strings.Builder

	if neir.Project != nil {
		fmt.Fprintf(&sb, "# Project: %s\n", neir.Project.Name)
		if neir.Project.Description != "" {
			fmt.Fprintf(&sb, "# %s\n", neir.Project.Description)
		}
		if neir.Project.Version != "" {
			fmt.Fprintf(&sb, "# Version: %s\n", neir.Project.Version)
		}
	}

	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "\n## Architecture: %s\n", neir.Architecture.Pattern)
		for _, p := range neir.Architecture.Principles {
			fmt.Fprintf(&sb, "- %s\n", p)
		}
	}

	if len(neir.Modules) > 0 {
		sb.WriteString("\n## Modules\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "- %s (%s)\n", m.Name, m.Path)
			if m.Description != "" {
				fmt.Fprintf(&sb, "  %s\n", m.Description)
			}
		}
	}

	if len(neir.Services) > 0 {
		sb.WriteString("\n## Services\n")
		for _, s := range neir.Services {
			fmt.Fprintf(&sb, "- %s [%s] port:%d\n", s.Name, s.Kind, s.Port)
			for _, ep := range s.Endpoints {
				fmt.Fprintf(&sb, "  %s %s -> %s\n", ep.Method, ep.Path, ep.Action)
			}
		}
	}

	if len(neir.Components) > 0 {
		sb.WriteString("\n## Components\n")
		for _, comp := range neir.Components {
			fmt.Fprintf(&sb, "- %s [%s] in module %s\n", comp.Name, comp.Kind, comp.Module)
		}
	}

	if len(neir.APIs) > 0 {
		sb.WriteString("\n## APIs\n")
		for _, api := range neir.APIs {
			fmt.Fprintf(&sb, "- %s v%s (%s)\n", api.Name, api.Version, api.Protocol)
			for _, ep := range api.Endpoints {
				fmt.Fprintf(&sb, "  %s %s: %s\n", ep.Method, ep.Path, ep.Summary)
			}
		}
	}

	if neir.Security != nil {
		sb.WriteString("\n## Security\n")
		if neir.Security.Authentication != nil {
			fmt.Fprintf(&sb, "- Auth: %s via %s\n", neir.Security.Authentication.Method, neir.Security.Authentication.Provider)
		}
		if neir.Security.Authorization != nil {
			fmt.Fprintf(&sb, "- Authorization: %s (roles: %s)\n", neir.Security.Authorization.Model, strings.Join(neir.Security.Authorization.Roles, ", "))
		}
	}

	if neir.Deployment != nil {
		fmt.Fprintf(&sb, "\n## Deployment: %s strategy\n", neir.Deployment.Strategy)
	}

	if neir.Testing != nil {
		fmt.Fprintf(&sb, "\n## Testing: %s strategy\n", neir.Testing.Strategy)
		if neir.Testing.Coverage != nil {
			fmt.Fprintf(&sb, "- Coverage target: %.0f%%\n", neir.Testing.Coverage.MinPercent)
		}
	}

	return sb.String()
}

func resolveLanguages(neir *model.NEIR) []language.Language {
	if neir.Generation != nil && len(neir.Generation.Languages) > 0 {
		return neir.Generation.Languages
	}
	return []language.Language{language.LanguageGo}
}
