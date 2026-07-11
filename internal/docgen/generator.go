package docgen

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type DocGenerator struct{}

func NewDocGenerator() *DocGenerator {
	return &DocGenerator{}
}

func (g *DocGenerator) GenerateFromSpec(doc *parser.SpecDocument) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", doc.Project))
	sb.WriteString("Auto-generated documentation from specification.\n\n")

	if doc.Architecture != nil {
		sb.WriteString("## Architecture\n\n")
		sb.WriteString(fmt.Sprintf("- Pattern: %s\n", doc.Architecture.Pattern))
		if len(doc.Architecture.Principles) > 0 {
			sb.WriteString(fmt.Sprintf("- Principles: %s\n", strings.Join(doc.Architecture.Principles, ", ")))
		}
		sb.WriteString("\n")
	}

	if len(doc.Modules) > 0 {
		sb.WriteString("## Modules\n\n")
		for _, m := range doc.Modules {
			sb.WriteString(fmt.Sprintf("### %s\n\n", m.Name))
			if m.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", m.Description))
			}
			sb.WriteString(fmt.Sprintf("- Path: `%s`\n", m.Path))
			if len(m.Dependencies) > 0 {
				sb.WriteString(fmt.Sprintf("- Dependencies: %s\n", strings.Join(m.Dependencies, ", ")))
			}
			sb.WriteString("\n")
		}
	}

	if len(doc.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, s := range doc.Services {
			sb.WriteString(fmt.Sprintf("### %s\n\n", s.Name))
			sb.WriteString(fmt.Sprintf("- Kind: %s\n", s.Kind))
			if s.Port > 0 {
				sb.WriteString(fmt.Sprintf("- Port: %d\n", s.Port))
			}
			if len(s.Endpoints) > 0 {
				sb.WriteString("\n**Endpoints:**\n\n")
				for _, ep := range s.Endpoints {
					sb.WriteString(fmt.Sprintf("- `%s %s` → %s\n", ep.Method, ep.Path, ep.Action))
				}
			}
			sb.WriteString("\n")
		}
	}

	if doc.Deployment != nil {
		sb.WriteString("## Deployment\n\n")
		sb.WriteString(fmt.Sprintf("- Strategy: %s\n", doc.Deployment.Strategy))
		if len(doc.Deployment.Environments) > 0 {
			sb.WriteString(fmt.Sprintf("- Environments: %s\n", strings.Join(doc.Deployment.Environments, ", ")))
		}
		sb.WriteString("\n")
	}

	if doc.Testing != nil {
		sb.WriteString("## Testing\n\n")
		sb.WriteString(fmt.Sprintf("- Strategy: %s\n", doc.Testing.Strategy))
		if doc.Testing.Coverage != "" {
			sb.WriteString(fmt.Sprintf("- Coverage target: %s%%\n", doc.Testing.Coverage))
		}
		sb.WriteString("\n")
	}

	if doc.Generation != nil && len(doc.Generation.Languages) > 0 {
		sb.WriteString("## Generation\n\n")
		sb.WriteString(fmt.Sprintf("- Languages: %s\n", strings.Join(doc.Generation.Languages, ", ")))
		if doc.Generation.OutputDir != "" {
			sb.WriteString(fmt.Sprintf("- Output: %s\n", doc.Generation.OutputDir))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *DocGenerator) GenerateFromNEIR(neir *model.NEIR) string {
	var sb strings.Builder

	if neir.Project != nil {
		sb.WriteString(fmt.Sprintf("# %s\n\n", neir.Project.Name))
		if neir.Project.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", neir.Project.Description))
		}
		if neir.Project.Version != "" {
			sb.WriteString(fmt.Sprintf("Version: %s\n\n", neir.Project.Version))
		}
	}

	if len(neir.Modules) > 0 {
		sb.WriteString("## Modules\n\n")
		for _, m := range neir.Modules {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`)", m.Name, m.Path))
			if m.Description != "" {
				sb.WriteString(fmt.Sprintf(" — %s", m.Description))
			}
			sb.WriteString("\n")
			if len(m.Dependencies) > 0 {
				sb.WriteString(fmt.Sprintf("  Dependencies: %s\n", strings.Join(m.Dependencies, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	if len(neir.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, s := range neir.Services {
			sb.WriteString(fmt.Sprintf("- **%s** (kind=%s", s.Name, string(s.Kind)))
			if s.Port > 0 {
				sb.WriteString(fmt.Sprintf(", port=%d", s.Port))
			}
			sb.WriteString(")\n")
			for _, ep := range s.Endpoints {
				sb.WriteString(fmt.Sprintf("  - `%s %s` → %s\n", ep.Method, ep.Path, ep.Action))
			}
		}
		sb.WriteString("\n")
	}

	if neir.Generation != nil {
		sb.WriteString("## Generation Config\n\n")
		var langs []string
		for _, l := range neir.Generation.Languages {
			langs = append(langs, string(l))
		}
		if len(langs) > 0 {
			sb.WriteString(fmt.Sprintf("- Languages: %s\n", strings.Join(langs, ", ")))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *DocGenerator) GenerateAPIDoc(doc *parser.SpecDocument) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — API Reference\n\n", doc.Project))

	for _, svc := range doc.Services {
		if len(svc.Endpoints) > 0 {
			sb.WriteString(fmt.Sprintf("## %s\n\n", svc.Name))
			for _, ep := range svc.Endpoints {
				sb.WriteString(fmt.Sprintf("### `%s %s`\n\n", ep.Method, ep.Path))
				if ep.Action != "" {
					sb.WriteString(fmt.Sprintf("**Action:** %s\n\n", ep.Action))
				}
			}
		}
	}

	return sb.String()
}

func (g *DocGenerator) GenerateModuleDocs(doc *parser.SpecDocument) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — Module Documentation\n\n", doc.Project))

	for _, m := range doc.Modules {
		sb.WriteString(fmt.Sprintf("## %s\n\n", m.Name))
		if m.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", m.Description))
		}
		sb.WriteString(fmt.Sprintf("**Path:** `%s`\n\n", m.Path))
		if len(m.Dependencies) > 0 {
			sb.WriteString("**Dependencies:**\n\n")
			for _, dep := range m.Dependencies {
				sb.WriteString(fmt.Sprintf("- %s\n", dep))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (g *DocGenerator) SupportedLanguages() []language.Language {
	return []language.Language{
		language.LanguageGo,
		language.LanguageTypeScript,
		language.LanguagePython,
		language.LanguageJava,
		language.LanguageRust,
	}
}
