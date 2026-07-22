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

type windsurfAdapter struct {
	library *promptlib.Library
}

func NewWindsurfAdapter(lib *promptlib.Library) compiler.Adapter {
	return &windsurfAdapter{library: lib}
}

func (a *windsurfAdapter) Target() compiler.Target {
	return compiler.TargetWindsurf
}

func (a *windsurfAdapter) Compile(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	if neir == nil {
		return nil, naeoserr.New(naeoserr.ErrInternal, "nil NEIR")
	}

	if a.library != nil {
		return a.compileFromLibrary(neir)
	}

	return a.compileLegacy(neir)
}

func (a *windsurfAdapter) compileFromLibrary(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	rendered, err := a.library.RenderCompiler("windsurf", neir)
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
		Target:     compiler.TargetWindsurf,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for Windsurf (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *windsurfAdapter) compileLegacy(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	var files []compiler.OutputFile

	rulesFile := a.buildRulesFile(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".windsurfrules",
		Content: rulesFile,
		Kind:    "rules",
	})

	contextFile := a.buildContextFile(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".windsurf/context.md",
		Content: contextFile,
		Kind:    "context",
	})

	projectName := "unknown"
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	return &compiler.CompiledOutput{
		Target:     compiler.TargetWindsurf,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for Windsurf (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *windsurfAdapter) buildRulesFile(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# Windsurf Rules\n\n")

	if neir.Project != nil {
		fmt.Fprintf(&sb, "project_name: %s\n", neir.Project.Name)
	}
	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "architecture: %s\n", neir.Architecture.Pattern)
	}

	sb.WriteString("\n## Instructions\n\n")
	sb.WriteString("You are working on a project with the following structure:\n\n")

	if len(neir.Modules) > 0 {
		sb.WriteString("### Modules\n\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "- `%s` at `%s`\n", m.Name, m.Path)
			if m.Description != "" {
				fmt.Fprintf(&sb, "  > %s\n", m.Description)
			}
		}
		sb.WriteString("\n")
	}

	if len(neir.Services) > 0 {
		sb.WriteString("### Services\n\n")
		for _, s := range neir.Services {
			fmt.Fprintf(&sb, "- **%s** (%s, port %d)\n", s.Name, s.Kind, s.Port)
			for _, ep := range s.Endpoints {
				fmt.Fprintf(&sb, "  - %s %s -> %s\n", ep.Method, ep.Path, ep.Action)
			}
		}
		sb.WriteString("\n")
	}

	if len(neir.Components) > 0 {
		sb.WriteString("### Components\n\n")
		for _, c := range neir.Components {
			fmt.Fprintf(&sb, "- `%s` [%s] module: `%s`\n", c.Name, c.Kind, c.Module)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Style Rules\n\n")
	sb.WriteString("- Use early returns\n")
	sb.WriteString("- Prefer const/readonly where possible\n")
	sb.WriteString("- Write self-documenting code with clear names\n")
	sb.WriteString("- Handle all error paths\n")
	sb.WriteString("- Keep functions under 50 lines\n")

	return sb.String()
}

func (a *windsurfAdapter) buildContextFile(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# Windsurf Context\n\n")
	sb.WriteString("Additional project context for Windsurf AI.\n\n")

	if neir.Project != nil {
		fmt.Fprintf(&sb, "Project: %s v%s\n", neir.Project.Name, neir.Project.Version)
	}

	if len(neir.Modules) > 0 {
		sb.WriteString("\n## Module Dependency Map\n\n")
		for _, m := range neir.Modules {
			if len(m.Dependencies) > 0 {
				fmt.Fprintf(&sb, "%s depends on: %s\n", m.Name, strings.Join(m.Dependencies, ", "))
			}
		}
	}

	if len(neir.APIs) > 0 {
		sb.WriteString("\n## API Endpoints\n\n")
		for _, api := range neir.APIs {
			fmt.Fprintf(&sb, "### %s v%s (%s)\n", api.Name, api.Version, api.Protocol)
			for _, ep := range api.Endpoints {
				fmt.Fprintf(&sb, "- %s %s: %s\n", ep.Method, ep.Path, ep.Summary)
			}
		}
	}

	return sb.String()
}
