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

type openCodeAdapter struct {
	library *promptlib.Library
}

func NewOpenCodeAdapter(lib *promptlib.Library) compiler.Adapter {
	return &openCodeAdapter{library: lib}
}

func (a *openCodeAdapter) Target() compiler.Target {
	return compiler.TargetOpenCode
}

func (a *openCodeAdapter) Compile(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	if neir == nil {
		return nil, naeoserr.New(naeoserr.ErrInternal, "nil NEIR")
	}

	if a.library != nil {
		return a.compileFromLibrary(neir)
	}

	return a.compileLegacy(neir)
}

func (a *openCodeAdapter) compileFromLibrary(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	rendered, err := a.library.RenderCompiler("opencode", neir)
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
		Target:     compiler.TargetOpenCode,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for OpenCode (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *openCodeAdapter) compileLegacy(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	var files []compiler.OutputFile

	instructions := a.buildInstructions(neir)
	files = append(files, compiler.OutputFile{
		Path:    "AGENTS.md",
		Content: instructions,
		Kind:    "instructions",
	})

	contextFile := a.buildContextFile(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".opencode/context.md",
		Content: contextFile,
		Kind:    "context",
	})

	rulesFile := a.buildRulesFile(neir)
	files = append(files, compiler.OutputFile{
		Path:    ".opencode/rules.md",
		Content: rulesFile,
		Kind:    "rules",
	})

	projectName := "unknown"
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	return &compiler.CompiledOutput{
		Target:     compiler.TargetOpenCode,
		Files:      files,
		Summary:    fmt.Sprintf("Generated %d files for OpenCode (%s)", len(files), projectName),
		CompiledAt: time.Now(),
	}, nil
}

func (a *openCodeAdapter) buildInstructions(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# AGENTS.md\n\n")
	sb.WriteString("Instructions for OpenCode agents working on this project.\n\n")

	if neir.Project != nil {
		fmt.Fprintf(&sb, "## Project: %s\n\n", neir.Project.Name)
		if neir.Project.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", neir.Project.Description)
		}
		if neir.Project.Version != "" {
			fmt.Fprintf(&sb, "Version: %s\n\n", neir.Project.Version)
		}
	}

	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "## Architecture: %s\n\n", neir.Architecture.Pattern)
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

	if len(neir.APIs) > 0 {
		sb.WriteString("## APIs\n\n")
		for _, api := range neir.APIs {
			fmt.Fprintf(&sb, "### %s v%s (%s)\n", api.Name, api.Version, api.Protocol)
			for _, ep := range api.Endpoints {
				fmt.Fprintf(&sb, "- %s %s: %s\n", ep.Method, ep.Path, ep.Summary)
			}
			sb.WriteString("\n")
		}
	}

	if neir.Security != nil {
		sb.WriteString("## Security\n\n")
		if neir.Security.Authentication != nil {
			fmt.Fprintf(&sb, "- Auth: %s via %s\n", neir.Security.Authentication.Method, neir.Security.Authentication.Provider)
		}
		if neir.Security.Authorization != nil {
			fmt.Fprintf(&sb, "- Authorization: %s\n", neir.Security.Authorization.Model)
		}
		if neir.Security.Encryption != nil {
			fmt.Fprintf(&sb, "- Encryption: in_transit=%v, at_rest=%v\n", neir.Security.Encryption.InTransit, neir.Security.Encryption.AtRest)
		}
		sb.WriteString("\n")
	}

	if neir.Deployment != nil {
		fmt.Fprintf(&sb, "## Deployment: %s\n\n", neir.Deployment.Strategy)
	}

	if neir.Testing != nil {
		fmt.Fprintf(&sb, "## Testing: %s\n\n", neir.Testing.Strategy)
	}

	sb.WriteString("## Guidelines\n\n")
	sb.WriteString("1. Follow the established architecture pattern\n")
	sb.WriteString("2. Write clean, idiomatic code\n")
	sb.WriteString("3. Handle all error paths explicitly\n")
	sb.WriteString("4. Write tests for new functionality\n")
	sb.WriteString("5. Keep functions focused and under 50 lines\n")
	sb.WriteString("6. Use meaningful variable and function names\n")
	sb.WriteString("7. Follow the module dependency structure\n")

	return sb.String()
}

func (a *openCodeAdapter) buildContextFile(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# OpenCode Context\n\n")

	if len(neir.Modules) > 0 {
		sb.WriteString("## Dependency Graph\n\n")
		for _, m := range neir.Modules {
			if len(m.Dependencies) > 0 {
				fmt.Fprintf(&sb, "%s → %s\n", m.Name, strings.Join(m.Dependencies, ", "))
			} else {
				fmt.Fprintf(&sb, "%s (root)\n", m.Name)
			}
		}
		sb.WriteString("\n")
	}

	if len(neir.Storage) > 0 {
		sb.WriteString("## Storage\n\n")
		for _, st := range neir.Storage {
			fmt.Fprintf(&sb, "- %s (%s, %s)\n", st.Name, st.Type, st.Provider)
		}
		sb.WriteString("\n")
	}

	if neir.Infrastructure != nil {
		fmt.Fprintf(&sb, "## Infrastructure: %s (%s)\n\n", neir.Infrastructure.Provider, neir.Infrastructure.Region)
		for _, r := range neir.Infrastructure.Resources {
			fmt.Fprintf(&sb, "- %s (%s)\n", r.Name, r.Kind)
		}
		sb.WriteString("\n")
	}

	if neir.AI != nil && len(neir.AI.Models) > 0 {
		sb.WriteString("## AI Models\n\n")
		for _, m := range neir.AI.Models {
			fmt.Fprintf(&sb, "- %s (%s v%s)\n", m.Name, m.Kind, m.Version)
		}
		sb.WriteString("\n")
	}

	if neir.Documentation != nil && (len(neir.Documentation.ADRs) > 0 || len(neir.Documentation.RFCs) > 0) {
		sb.WriteString("## Design Documents\n\n")
		for _, doc := range neir.Documentation.ADRs {
			fmt.Fprintf(&sb, "- ADR: %s\n", doc.Title)
		}
		for _, doc := range neir.Documentation.RFCs {
			fmt.Fprintf(&sb, "- RFC: %s\n", doc.Title)
		}
	}

	return sb.String()
}

func (a *openCodeAdapter) buildRulesFile(neir *model.NEIR) string {
	var sb strings.Builder
	sb.WriteString("# OpenCode Rules\n\n")

	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "Architecture: %s\n\n", neir.Architecture.Pattern)
	}

	sb.WriteString("## Code Rules\n\n")
	sb.WriteString("1. Never use `panic` for error handling\n")
	sb.WriteString("2. Always check error returns\n")
	sb.WriteString("3. Use table-driven tests\n")
	sb.WriteString("4. Keep package imports clean (no circular dependencies)\n")
	sb.WriteString("5. Follow existing code patterns in the project\n")
	sb.WriteString("6. Use dependency injection for testability\n")

	if len(neir.Modules) > 0 {
		sb.WriteString("\n## Module Boundaries\n\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "- `%s` should only depend on: %s\n", m.Name, strings.Join(m.Dependencies, ", "))
		}
	}

	if neir.Security != nil {
		sb.WriteString("\n## Security Rules\n\n")
		sb.WriteString("- Never hardcode secrets\n")
		sb.WriteString("- Use environment variables for configuration\n")
		sb.WriteString("- Validate all external input\n")
		if neir.Security.Encryption != nil && neir.Security.Encryption.InTransit {
			sb.WriteString("- All network communication must use TLS\n")
		}
	}

	return sb.String()
}
