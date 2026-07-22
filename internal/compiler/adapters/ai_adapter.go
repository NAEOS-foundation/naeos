package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/ai"
	"github.com/NAEOS-foundation/naeos/internal/compiler"
	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
)

type AICompilerAdapter struct {
	target compiler.Target
	llm    *ai.LLMService
}

func NewAICompilerAdapter(target compiler.Target, llm *ai.LLMService) compiler.Adapter {
	return &AICompilerAdapter{
		target: target,
		llm:    llm,
	}
}

func (a *AICompilerAdapter) Target() compiler.Target {
	return a.target
}

func (a *AICompilerAdapter) Compile(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	return a.CompileContext(ctx, neir)
}

// CompileContext compiles NEIR into compiler output with context support.
func (a *AICompilerAdapter) CompileContext(ctx context.Context, neir *model.NEIR) (*compiler.CompiledOutput, error) {
	neirContext := buildNEIRContext(neir)

	var buf structuredBuffer
	if err := a.llm.StreamCompileSpec(ctx, string(a.target), neirContext, &buf); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "ai compile %s", a.target)
	}

	files, err := parseCompiledFiles(buf.String())
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse ai output for %s", a.target)
	}

	projectName := "unknown"
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	return &compiler.CompiledOutput{
		Target:     a.target,
		Files:      files,
		Summary:    fmt.Sprintf("AI-generated %d files for %s (%s)", len(files), a.target, projectName),
		CompiledAt: time.Now(),
	}, nil
}

type structuredBuffer struct {
	data []byte
}

func (b *structuredBuffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *structuredBuffer) String() string {
	return string(b.data)
}

func parseCompiledFiles(output string) ([]compiler.OutputFile, error) {
	cleaned := ai.CleanJSON(output)
	var files []compiler.OutputFile
	if err := json.Unmarshal([]byte(cleaned), &files); err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, naeoserr.New(naeoserr.ErrInternal, "no files in AI output")
	}
	return files, nil
}

func buildNEIRContext(neir *model.NEIR) string {
	var sb strings.Builder

	if neir.Project != nil {
		fmt.Fprintf(&sb, "Project: %s\n", neir.Project.Name)
		if neir.Project.Description != "" {
			fmt.Fprintf(&sb, "Description: %s\n", neir.Project.Description)
		}
		if neir.Project.Version != "" {
			fmt.Fprintf(&sb, "Version: %s\n", neir.Project.Version)
		}
	}

	if neir.Architecture != nil {
		fmt.Fprintf(&sb, "\nArchitecture: %s\n", neir.Architecture.Pattern)
		for _, p := range neir.Architecture.Principles {
			fmt.Fprintf(&sb, "- %s\n", p)
		}
	}

	if len(neir.Modules) > 0 {
		sb.WriteString("\nModules:\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "  %s (%s)", m.Name, m.Path)
			if m.Description != "" {
				fmt.Fprintf(&sb, ": %s", m.Description)
			}
			if len(m.Dependencies) > 0 {
				fmt.Fprintf(&sb, " [depends: %s]", strings.Join(m.Dependencies, ", "))
			}
			sb.WriteString("\n")
		}
	}

	if len(neir.Services) > 0 {
		sb.WriteString("\nServices:\n")
		for _, s := range neir.Services {
			fmt.Fprintf(&sb, "  %s [%s] port:%d\n", s.Name, s.Kind, s.Port)
			for _, ep := range s.Endpoints {
				fmt.Fprintf(&sb, "    %s %s -> %s\n", ep.Method, ep.Path, ep.Action)
			}
		}
	}

	if len(neir.Components) > 0 {
		sb.WriteString("\nComponents:\n")
		for _, c := range neir.Components {
			fmt.Fprintf(&sb, "  %s [%s] in %s\n", c.Name, c.Kind, c.Module)
		}
	}

	if len(neir.APIs) > 0 {
		sb.WriteString("\nAPIs:\n")
		for _, api := range neir.APIs {
			fmt.Fprintf(&sb, "  %s v%s (%s)\n", api.Name, api.Version, api.Protocol)
			for _, ep := range api.Endpoints {
				fmt.Fprintf(&sb, "    %s %s: %s\n", ep.Method, ep.Path, ep.Summary)
			}
		}
	}

	if neir.Security != nil {
		sb.WriteString("\nSecurity:\n")
		if neir.Security.Authentication != nil {
			fmt.Fprintf(&sb, "  Auth: %s via %s\n", neir.Security.Authentication.Method, neir.Security.Authentication.Provider)
		}
		if neir.Security.Authorization != nil {
			fmt.Fprintf(&sb, "  AuthZ: %s (roles: %s)\n", neir.Security.Authorization.Model, strings.Join(neir.Security.Authorization.Roles, ", "))
		}
	}

	if neir.Deployment != nil {
		fmt.Fprintf(&sb, "\nDeployment: %s\n", neir.Deployment.Strategy)
	}

	if neir.Testing != nil {
		fmt.Fprintf(&sb, "\nTesting: %s\n", neir.Testing.Strategy)
	}

	return sb.String()
}
