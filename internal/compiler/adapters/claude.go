package adapters

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/promptlib"
)

type claudeAdapter struct {
	library *promptlib.Library
}

func NewClaudeAdapter(lib *promptlib.Library) compiler.Adapter {
	return &claudeAdapter{library: lib}
}

func (a *claudeAdapter) Target() compiler.Target {
	return compiler.TargetClaude
}

func (a *claudeAdapter) Compile(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	if neir == nil {
		return nil, naeoserr.New(naeoserr.ErrInternal, "nil NEIR")
	}

	if a.library != nil {
		return a.compileFromLibrary(neir)
	}
	return a.compileLegacy(neir)
}

func (a *claudeAdapter) compileFromLibrary(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	rendered, err := a.library.RenderCompiler("claude", neir)
	if err != nil {
		return nil, fmt.Errorf("render claude template: %w", err)
	}

	var files []compiler.OutputFile
	for _, f := range rendered {
		files = append(files, compiler.OutputFile{
			Path:    f.Path,
			Content: f.Content,
			Kind:    f.Kind,
		})
	}

	projectName := "unknown"
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	return &compiler.CompiledOutput{
		Target:     compiler.TargetClaude,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for Claude Code (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *claudeAdapter) compileLegacy(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	var files []compiler.OutputFile

	claudeMd := buildClaudeMdLegacy(neir)
	files = append(files, compiler.OutputFile{
		Path:    "CLAUDE.md",
		Content: claudeMd,
		Kind:    "instructions",
	})

	contextFile := buildClaudeContextLegacy(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".claude/context.md",
		Content: contextFile,
		Kind:    "context",
	})

	rulesFile := buildClaudeRulesLegacy(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".claude/rules.md",
		Content: rulesFile,
		Kind:    "rules",
	})

	projectName := "unknown"
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	return &compiler.CompiledOutput{
		Target:     compiler.TargetClaude,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for Claude Code (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func buildClaudeMdLegacy(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# CLAUDE.md\n\n")
	sb.WriteString("This file provides context for Claude Code when working on this project.\n\n")

	if neir.Project != nil {
		fmt.Fprintf(&sb, "## Project: %s\n\n", neir.Project.Name)
		if neir.Project.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", neir.Project.Description)
		}
	}

	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "## Architecture\n\nPattern: **%s**\n\n", neir.Architecture.Pattern)
		if len(neir.Architecture.Principles) > 0 {
			sb.WriteString("Principles:\n")
			for _, p := range neir.Architecture.Principles {
				fmt.Fprintf(&sb, "- %s\n", p)
			}
			sb.WriteString("\n")
		}
	}

	if len(neir.Modules) > 0 {
		sb.WriteString("## Modules\n\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "### %s\nPath: `%s`\n", m.Name, m.Path)
			if m.Description != "" {
				fmt.Fprintf(&sb, "Description: %s\n", m.Description)
			}
			if len(m.Dependencies) > 0 {
				fmt.Fprintf(&sb, "Dependencies: %s\n", strings.Join(m.Dependencies, ", "))
			}
			sb.WriteString("\n")
		}
	}

	if len(neir.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, s := range neir.Services {
			fmt.Fprintf(&sb, "### %s\nType: %s, Port: %d\n\n", s.Name, s.Kind, s.Port)
			if len(s.Endpoints) > 0 {
				sb.WriteString("Endpoints:\n")
				for _, ep := range s.Endpoints {
					fmt.Fprintf(&sb, "- `%s %s` → %s\n", ep.Method, ep.Path, ep.Action)
				}
				sb.WriteString("\n")
			}
		}
	}

	if neir.Security != nil && neir.Security.Authentication != nil {
		sb.WriteString("## Security\n\n")
		fmt.Fprintf(&sb, "Authentication: %s via %s\n\n", neir.Security.Authentication.Method, neir.Security.Authentication.Provider)
	}

	sb.WriteString("## Guidelines\n\n")
	sb.WriteString("- Follow the architecture pattern strictly\n")
	sb.WriteString("- Write clean, well-tested code\n")
	sb.WriteString("- Handle errors explicitly\n")
	sb.WriteString("- Keep functions focused and small\n")
	sb.WriteString("- Document public APIs with godoc/JSDoc/docstrings\n")

	return sb.String()
}

func buildClaudeContextLegacy(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# Context Bundle for Claude Code\n\n")
	sb.WriteString("Project structure and key design decisions.\n\n")

	if len(neir.Modules) > 0 {
		sb.WriteString("## Dependency Graph\n\n")
		for _, m := range neir.Modules {
			if len(m.Dependencies) > 0 {
				fmt.Fprintf(&sb, "%s → %s\n", m.Name, strings.Join(m.Dependencies, ", "))
			} else {
				fmt.Fprintf(&sb, "%s (no dependencies)\n", m.Name)
			}
		}
		sb.WriteString("\n")
	}

	if len(neir.Components) > 0 {
		sb.WriteString("## Component Map\n\n")
		for _, c := range neir.Components {
			fmt.Fprintf(&sb, "- **%s** (%s) in `%s`\n", c.Name, c.Kind, c.Module)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func buildClaudeRulesLegacy(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# Claude Code Rules\n\n")

	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "Architecture pattern: %s\n\n", neir.Architecture.Pattern)
	}

	sb.WriteString("## Code Rules\n\n")
	sb.WriteString("1. Always use explicit error returns, never panic\n")
	sb.WriteString("2. Prefer table-driven tests\n")
	sb.WriteString("3. Keep package boundaries clean\n")
	sb.WriteString("4. Use dependency injection for testability\n")
	sb.WriteString("5. Follow naming conventions of the target language\n")

	if len(neir.Modules) > 0 {
		sb.WriteString("\n## Module Rules\n\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "- `%s` should not import from unrelated modules\n", m.Name)
		}
	}

	return sb.String()
}
