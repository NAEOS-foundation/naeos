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

type codexAdapter struct {
	library *promptlib.Library
}

func NewCodexAdapter(lib *promptlib.Library) compiler.Adapter {
	return &codexAdapter{library: lib}
}

func (a *codexAdapter) Target() compiler.Target {
	return compiler.TargetCodex
}

func (a *codexAdapter) Compile(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	if neir == nil {
		return nil, naeoserr.New(naeoserr.ErrInternal, "nil NEIR")
	}

	if a.library != nil {
		return a.compileFromLibrary(neir)
	}

	return a.compileLegacy(neir)
}

func (a *codexAdapter) compileFromLibrary(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	rendered, err := a.library.RenderCompiler("codex", neir)
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
		Target:     compiler.TargetCodex,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for Codex (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *codexAdapter) compileLegacy(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	var files []compiler.OutputFile

	instructions := a.buildInstructions(neir)
	files = append(files, compiler.OutputFile{
		Path:    "AGENTS.md",
		Content: instructions,
		Kind:    "instructions",
	})

	contextFile := a.buildContextFile(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".codex/context.md",
		Content: contextFile,
		Kind:    "context",
	})

	projectName := "unknown"
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	return &compiler.CompiledOutput{
		Target:     compiler.TargetCodex,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for Codex (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *codexAdapter) buildInstructions(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# AGENTS.md\n\n")
	sb.WriteString("Instructions for AI agents working on this project.\n\n")

	if neir.Project != nil {
		fmt.Fprintf(&sb, "## Project: %s\n\n", neir.Project.Name)
		if neir.Project.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", neir.Project.Description)
		}
	}

	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "## Architecture: %s\n\n", neir.Architecture.Pattern)
		for _, p := range neir.Architecture.Principles {
			fmt.Fprintf(&sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Module Structure\n\n")
	if len(neir.Modules) > 0 {
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "### %s\nPath: `%s`\n", m.Name, m.Path)
			if m.Description != "" {
				fmt.Fprintf(&sb, "%s\n", m.Description)
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
			fmt.Fprintf(&sb, "### %s (%s, port %d)\n", s.Name, s.Kind, s.Port)
			for _, ep := range s.Endpoints {
				fmt.Fprintf(&sb, "- %s %s → %s\n", ep.Method, ep.Path, ep.Action)
			}
			sb.WriteString("\n")
		}
	}

	if len(neir.Components) > 0 {
		sb.WriteString("## Components\n\n")
		for _, c := range neir.Components {
			fmt.Fprintf(&sb, "- `%s` [%s] in `%s`\n", c.Name, c.Kind, c.Module)
		}
		sb.WriteString("\n")
	}

	if neir.Deployment != nil {
		fmt.Fprintf(&sb, "## Deployment: %s\n\n", neir.Deployment.Strategy)
		if len(neir.Deployment.Environments) > 0 {
			for _, env := range neir.Deployment.Environments {
				fmt.Fprintf(&sb, "- %s (%s)\n", env.Name, env.Kind)
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Agent Guidelines\n\n")
	sb.WriteString("1. Follow the architecture pattern\n")
	sb.WriteString("2. Write clean, idiomatic code for the target language\n")
	sb.WriteString("3. Handle errors explicitly\n")
	sb.WriteString("4. Write tests for new code\n")
	sb.WriteString("5. Keep functions focused and small\n")
	sb.WriteString("6. Document public APIs\n")

	return sb.String()
}

func (a *codexAdapter) buildContextFile(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# Codex Context\n\n")

	if len(neir.Storage) > 0 {
		sb.WriteString("## Storage\n\n")
		for _, st := range neir.Storage {
			fmt.Fprintf(&sb, "- %s (%s) via %s\n", st.Name, st.Type, st.Provider)
			for _, col := range st.Collections {
				fmt.Fprintf(&sb, "  - %s\n", col.Name)
			}
		}
		sb.WriteString("\n")
	}

	if neir.Infrastructure != nil {
		fmt.Fprintf(&sb, "## Infrastructure: %s\n\n", neir.Infrastructure.Provider)
		for _, r := range neir.Infrastructure.Resources {
			fmt.Fprintf(&sb, "- %s (%s)\n", r.Name, r.Kind)
		}
	}

	if neir.AI != nil && len(neir.AI.Models) > 0 {
		sb.WriteString("\n## AI Models\n\n")
		for _, m := range neir.AI.Models {
			fmt.Fprintf(&sb, "- %s (%s) v%s\n", m.Name, m.Kind, m.Version)
		}
	}

	return sb.String()
}
