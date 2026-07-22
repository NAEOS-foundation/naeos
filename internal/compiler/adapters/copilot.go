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

type copilotAdapter struct {
	library *promptlib.Library
}

func NewCopilotAdapter(lib *promptlib.Library) compiler.Adapter {
	return &copilotAdapter{library: lib}
}

func (a *copilotAdapter) Target() compiler.Target {
	return compiler.TargetCopilot
}

func (a *copilotAdapter) Compile(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	if neir == nil {
		return nil, naeoserr.New(naeoserr.ErrInternal, "nil NEIR")
	}

	if a.library != nil {
		return a.compileFromLibrary(neir)
	}

	return a.compileLegacy(neir)
}

func (a *copilotAdapter) compileFromLibrary(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	rendered, err := a.library.RenderCompiler("copilot", neir)
	if err != nil {
		return nil, fmt.Errorf("render from library: %w", err)
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
		Target:     compiler.TargetCopilot,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for GitHub Copilot (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *copilotAdapter) compileLegacy(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	var files []compiler.OutputFile

	instructions := a.buildInstructions(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".github/copilot-instructions.md",
		Content: instructions,
		Kind:    "instructions",
	})

	context := a.buildContextFile(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".github/copilot-context.md",
		Content: context,
		Kind:    "context",
	})

	rules := a.buildRulesFile(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".github/copilot-rules.md",
		Content: rules,
		Kind:    "rules",
	})

	projectName := "unknown"
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	return &compiler.CompiledOutput{
		Target:     compiler.TargetCopilot,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for GitHub Copilot (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *copilotAdapter) buildInstructions(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# GitHub Copilot Instructions\n\n")
	sb.WriteString("This file contains project-specific instructions for GitHub Copilot.\n\n")

	if neir.Project != nil {
		fmt.Fprintf(&sb, "## Project: %s\n\n", neir.Project.Name)
		if neir.Project.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", neir.Project.Description)
		}
	}

	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "## Architecture Pattern: %s\n\n", neir.Architecture.Pattern)
		sb.WriteString("Follow these architectural principles:\n")
		for _, p := range neir.Architecture.Principles {
			fmt.Fprintf(&sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}

	if len(neir.Modules) > 0 {
		sb.WriteString("## Module Structure\n\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "- **%s**: %s\n", m.Name, m.Path)
			if m.Description != "" {
				fmt.Fprintf(&sb, "  %s\n", m.Description)
			}
		}
		sb.WriteString("\n")
	}

	if len(neir.Components) > 0 {
		sb.WriteString("## Components\n\n")
		for _, c := range neir.Components {
			fmt.Fprintf(&sb, "- `%s` (%s) in module `%s`\n", c.Name, c.Kind, c.Module)
		}
		sb.WriteString("\n")
	}

	if len(neir.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, s := range neir.Services {
			fmt.Fprintf(&sb, "### %s (%s, port %d)\n", s.Name, s.Kind, s.Port)
			for _, ep := range s.Endpoints {
				fmt.Fprintf(&sb, "- `%s %s` -> `%s`\n", ep.Method, ep.Path, ep.Action)
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("## Coding Guidelines\n\n")
	sb.WriteString("- Write clean, idiomatic code following the project's architecture pattern\n")
	sb.WriteString("- Include proper error handling\n")
	sb.WriteString("- Add comments for public APIs\n")
	sb.WriteString("- Follow the module dependency structure\n")

	return sb.String()
}

func (a *copilotAdapter) buildContextFile(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# Project Context for Copilot\n\n")
	sb.WriteString("Use this file as additional context when generating code.\n\n")
	sb.WriteString("```yaml\n")
	fmt.Fprintf(&sb, "project: %s\n", neir.Project.Name)
	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "architecture: %s\n", neir.Architecture.Pattern)
	}
	if len(neir.Modules) > 0 {
		sb.WriteString("modules:\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "  - name: %s\n    path: %s\n", m.Name, m.Path)
		}
	}
	sb.WriteString("```\n")
	return sb.String()
}

func (a *copilotAdapter) buildRulesFile(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# Copilot Rules\n\n")
	sb.WriteString("## File Organization\n\n")

	if len(neir.Modules) > 0 {
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "- Files in `%s` belong to the `%s` module\n", m.Path, m.Name)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Code Style\n\n")
	sb.WriteString("- Use early returns for error handling\n")
	sb.WriteString("- Prefer composition over inheritance\n")
	sb.WriteString("- Keep functions small and focused\n")
	sb.WriteString("- Use meaningful variable and function names\n")

	return sb.String()
}
