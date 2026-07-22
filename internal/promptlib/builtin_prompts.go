package promptlib

// BuiltinLLMPrompts contains all built-in LLM prompt templates as YAML strings.
var builtinLLMPrompts = map[string]string{
	"enrich-spec": `name: enrich-spec
kind: llm
version: "1.0.0"
description: "Enrich a NAEOS specification with best practices"
provider: any
system: |
  You are a platform engineering expert specializing in NAEOS
  (Nusantara Engineering & Architecture Operating System).
user: |
  Analyze this NAEOS specification and enrich it with best practices.
  Add any missing sections that would improve the specification.
  Keep the existing content intact.
  Only output the enriched YAML specification, no explanations.

  Specification:
  {{.SpecContent}}
variables:
  - name: SpecContent
    type: string
    required: true
    description: "The raw YAML specification content"
constraints:
  max_tokens: 2048
  temperature: 0.3
`,
	"generate-suggestions": `name: generate-suggestions
kind: llm
version: "1.0.0"
description: "Generate improvement suggestions for a NAEOS specification"
provider: any
user: |
  Analyze this NAEOS specification and return a JSON array of suggestions.
  Each suggestion should have: category, title, description, priority (high/medium/low).
  Return ONLY the JSON array, no other text.

  Specification:
  {{.SpecContent}}
variables:
  - name: SpecContent
    type: string
    required: true
    description: "The raw YAML specification content"
constraints:
  max_tokens: 2048
  temperature: 0.3
`,
	"explain-architecture": `name: explain-architecture
kind: llm
version: "1.0.0"
description: "Explain an architecture pattern in the context of a specification"
provider: any
user: |
  Explain the architecture pattern "{{.Architecture}}" in the context of this specification.
  Provide a clear, concise explanation suitable for a developer.

  Specification:
  {{.SpecContent}}

  Architecture explanation:
variables:
  - name: Architecture
    type: string
    required: true
    description: "The architecture pattern name"
  - name: SpecContent
    type: string
    required: true
    description: "The raw YAML specification content"
constraints:
  max_tokens: 1024
  temperature: 0.3
`,
}

// BuiltinCompilerTemplates contains all built-in compiler templates as YAML strings.
var builtinCompilerTemplates = map[string]string{
	"copilot": `name: copilot
kind: compiler
version: "1.0.0"
target: copilot
files:
  - path: ".github/copilot-instructions.md"
    kind: instructions
    template: |
      # GitHub Copilot Instructions

      This file contains project-specific instructions for GitHub Copilot.

      {{if .Project.Name}}## Project: {{.Project.Name}}

      {{if .Project.Description}}{{.Project.Description}}{{end}}
      {{end}}
      {{if .Architecture.Pattern}}## Architecture Pattern: {{.Architecture.Pattern}}

      Follow these architectural principles:
      {{range .Architecture.Principles}}- {{.}}
      {{end}}{{end}}
      {{if .Modules}}## Module Structure

      {{range .Modules}}- **{{.Name}}**: {{.Path}}
      {{if .Description}}  {{.Description}}{{end}}
      {{end}}{{end}}
      {{if .Components}}## Components

      {{range .Components}}- {{code .Name}} ({{.Kind}}) in module {{code .Module}}
      {{end}}{{end}}
      {{if .Services}}## Services

      {{range .Services}}### {{.Name}} ({{.Kind}}, port {{.Port}})
      {{range .Endpoints}}- {{code (printf "%s %s" .Method .Path)}} -> {{code .Action}}
      {{end}}{{end}}{{end}}
      ## Coding Guidelines

      - Write clean, idiomatic code following the project's architecture pattern
      - Include proper error handling
      - Add comments for public APIs
      - Follow the module dependency structure

  - path: ".github/copilot-context.md"
    kind: context
    template: |
      # Project Context for Copilot

      Use this file as additional context when generating code.

      ` + "```yaml" + `
      project: {{.Project.Name}}
      {{if .Architecture.Pattern}}architecture: {{.Architecture.Pattern}}{{end}}
      {{if .Modules}}modules:
      {{range .Modules}}  - name: {{.Name}}
        path: {{.Path}}
      {{end}}{{end}}` + "```" + `

  - path: ".github/copilot-rules.md"
    kind: rules
    template: |
      # Copilot Rules

      ## File Organization

      {{range .Modules}}- Files in {{code .Path}} belong to the {{code .Name}} module
      {{end}}

      ## Code Style

      - Use early returns for error handling
      - Prefer composition over inheritance
      - Keep functions small and focused
      - Use meaningful variable and function names
variables:
  - name: Guidelines
    type: "[]string"
    default:
      - "Write clean, idiomatic code following the project's architecture pattern"
      - "Include proper error handling"
      - "Add comments for public APIs"
      - "Follow the module dependency structure"
`,

	"claude": `name: claude
kind: compiler
version: "1.0.0"
target: claude
files:
  - path: "CLAUDE.md"
    kind: instructions
    template: |
      # CLAUDE.md

      This file provides context for Claude Code when working on this project.

      {{if .Project.Name}}## Project: {{.Project.Name}}

      {{if .Project.Description}}{{.Project.Description}}{{end}}
      {{end}}
      {{if .Architecture.Pattern}}## Architecture

      Pattern: **{{.Architecture.Pattern}}**

      {{if .Architecture.Principles}}Principles:
      {{range .Architecture.Principles}}- {{.}}
      {{end}}{{end}}{{end}}
      {{if .Modules}}## Modules

      {{range .Modules}}### {{.Name}}
      Path: {{code .Path}}
      {{if .Description}}Description: {{.Description}}{{end}}
      {{if .Dependencies}}Dependencies: {{join .Dependencies ", "}}{{end}}

      {{end}}{{end}}
      {{if .Services}}## Services

      {{range .Services}}### {{.Name}}
      Type: {{.Kind}}, Port: {{.Port}}

      {{if .Endpoints}}Endpoints:
      {{range .Endpoints}}- {{code (printf "%s %s" .Method .Path)}} -> {{.Action}}
      {{end}}{{end}}
      {{end}}{{end}}
      {{if and .Security .Security.Authentication}}## Security

      Authentication: {{.Security.Authentication.Method}} via {{.Security.Authentication.Provider}}
      {{end}}

      ## Guidelines

      - Follow the architecture pattern strictly
      - Write clean, well-tested code
      - Handle errors explicitly
      - Keep functions focused and small
      - Document public APIs with godoc/JSDoc/docstrings

  - path: ".claude/context.md"
    kind: context
    template: |
      # Context Bundle for Claude Code

      Project structure and key design decisions.

      {{if .Modules}}
      ## Dependency Graph

      {{range .Modules}}{{if .Dependencies}}{{.Name}} -> {{join .Dependencies ", "}}{{else}}{{.Name}} (no dependencies){{end}}
      {{end}}{{end}}

      {{if .Components}}
      ## Component Map

      {{range .Components}}- **{{.Name}}** ({{.Kind}}) in {{code .Module}}
      {{end}}{{end}}

  - path: ".claude/rules.md"
    kind: rules
    template: |
      # Claude Code Rules

      {{if .Architecture.Pattern}}Architecture pattern: {{.Architecture.Pattern}}{{end}}

      ## Code Rules

      1. Always use explicit error returns, never panic
      2. Prefer table-driven tests
      3. Keep package boundaries clean
      4. Use dependency injection for testability
      5. Follow naming conventions of the target language

      {{if .Modules}}
      ## Module Rules

      {{range .Modules}}- {{code .Name}} should not import from unrelated modules
      {{end}}{{end}}
variables:
  - name: Guidelines
    type: "[]string"
    default:
      - "Follow the architecture pattern strictly"
      - "Write clean, well-tested code"
      - "Handle errors explicitly"
      - "Keep functions focused and small"
      - "Document public APIs with godoc/JSDoc/docstrings"

  - name: CodeRules
    type: "[]string"
    default:
      - "Always use explicit error returns, never panic"
      - "Prefer table-driven tests"
      - "Keep package boundaries clean"
      - "Use dependency injection for testability"
      - "Follow naming conventions of the target language"
`,

	"cursor": `name: cursor
kind: compiler
version: "1.0.0"
target: cursor
files:
  - path: ".cursorrules"
    kind: rules
    template: |
      # Cursor Rules

      {{if .Project.Name}}project_name: {{.Project.Name}}{{end}}
      {{if .Architecture.Pattern}}architecture: {{.Architecture.Pattern}}{{end}}

      ## Instructions

      You are working on a project with the following structure:

      {{if .Modules}}
      ### Modules

      {{range .Modules}}- {{code .Name}} at {{code .Path}}
      {{if .Description}}  > {{.Description}}{{end}}
      {{end}}{{end}}
      {{if .Services}}
      ### Services

      {{range .Services}}- **{{.Name}}** ({{.Kind}}, port {{.Port}})
      {{range .Endpoints}}  - {{.Method}} {{.Path}} -> {{.Action}}
      {{end}}{{end}}{{end}}
      {{if .Components}}
      ### Components

      {{range .Components}}- {{code .Name}} [{{.Kind}}] module: {{code .Module}}
      {{end}}{{end}}

      ## Style Rules

      - Use early returns
      - Prefer const/readonly where possible
      - Write self-documenting code with clear names
      - Handle all error paths
      - Keep functions under 50 lines

  - path: ".cursor/context.md"
    kind: context
    template: |
      # Cursor Context

      Additional project context for AI-assisted coding.

      {{if .Project.Name}}Project: {{.Project.Name}} {{.Project.Version}}{{end}}

      {{if .Modules}}
      ## Module Dependency Map

      {{range .Modules}}{{if .Dependencies}}{{.Name}} depends on: {{join .Dependencies ", "}}{{end}}
      {{end}}{{end}}
      {{if .APIs}}
      ## API Endpoints

      {{range .APIs}}### {{.Name}} v{{.Version}} ({{.Protocol}})
      {{range .Endpoints}}- {{.Method}} {{.Path}}: {{.Summary}}
      {{end}}{{end}}{{end}}
variables:
  - name: StyleRules
    type: "[]string"
    default:
      - "Use early returns"
      - "Prefer const/readonly where possible"
      - "Write self-documenting code with clear names"
      - "Handle all error paths"
      - "Keep functions under 50 lines"
`,

	"gemini": `name: gemini
kind: compiler
version: "1.0.0"
target: gemini
files:
  - path: ".gemini/CONFIG.md"
    kind: instructions
    template: |
      # Gemini CLI Configuration

      {{if .Project.Name}}Project: {{.Project.Name}}
      {{if .Project.Description}}Description: {{.Project.Description}}{{end}}
      {{end}}
      {{if .Architecture.Pattern}}

      Architecture: {{.Architecture.Pattern}}{{end}}

      ## Project Structure

      {{range .Modules}}- {{code .Name}} -> {{code .Path}}
      {{if .Description}}  {{.Description}}{{end}}
      {{end}}
      {{if .Services}}
      ## Services

      {{range .Services}}- {{.Name}} ({{.Kind}}, port {{.Port}})
      {{end}}{{end}}

      ## Guidelines

      - Follow established patterns in the codebase
      - Write clean, maintainable code
      - Include proper error handling
      - Add tests for new functionality

  - path: ".gemini/context.md"
    kind: context
    template: |
      # Gemini Context

      {{if .Components}}
      ## Components

      {{range .Components}}- {{.Name}} ({{.Kind}}) in module {{.Module}}
      {{end}}{{end}}
      {{if .APIs}}
      ## APIs

      {{range .APIs}}### {{.Name}} ({{.Protocol}})
      {{range .Endpoints}}- {{.Method}} {{.Path}}: {{.Summary}}
      {{end}}{{end}}{{end}}
      {{if .Testing}}

      ## Testing: {{.Testing.Strategy}} strategy
      {{if .Testing.Frameworks}}Frameworks: {{join .Testing.Frameworks ", "}}{{end}}
      {{end}}
variables:
  - name: Guidelines
    type: "[]string"
    default:
      - "Follow established patterns in the codebase"
      - "Write clean, maintainable code"
      - "Include proper error handling"
      - "Add tests for new functionality"
`,

	"codex": `name: codex
kind: compiler
version: "1.0.0"
target: codex
files:
  - path: "AGENTS.md"
    kind: instructions
    template: |
      # AGENTS.md

      Instructions for AI agents working on this project.

      {{if .Project.Name}}## Project: {{.Project.Name}}

      {{if .Project.Description}}{{.Project.Description}}{{end}}
      {{end}}
      {{if .Architecture.Pattern}}## Architecture: {{.Architecture.Pattern}}

      {{range .Architecture.Principles}}- {{.}}
      {{end}}{{end}}
      ## Module Structure

      {{range .Modules}}### {{.Name}}
      Path: {{code .Path}}
      {{if .Description}}{{.Description}}{{end}}
      {{if .Dependencies}}Dependencies: {{join .Dependencies ", "}}{{end}}

      {{end}}
      {{if .Services}}## Services

      {{range .Services}}### {{.Name}} ({{.Kind}}, port {{.Port}})
      {{range .Endpoints}}- {{.Method}} {{.Path}} -> {{.Action}}
      {{end}}

      {{end}}{{end}}
      {{if .Components}}## Components

      {{range .Components}}- {{code .Name}} [{{.Kind}}] in {{code .Module}}
      {{end}}

      {{end}}
      {{if .Deployment}}## Deployment: {{.Deployment.Strategy}}

      {{range .Deployment.Environments}}- {{.Name}} ({{.Kind}})
      {{end}}

      {{end}}
      ## Agent Guidelines

      1. Follow the architecture pattern
      2. Write clean, idiomatic code for the target language
      3. Handle errors explicitly
      4. Write tests for new code
      5. Keep functions focused and small
      6. Document public APIs

  - path: ".codex/context.md"
    kind: context
    template: |
      # Codex Context

      {{if .Storage}}
      ## Storage

      {{range .Storage}}- {{.Name}} ({{.Type}}) via {{.Provider}}
      {{range .Collections}}  - {{.Name}}
      {{end}}{{end}}{{end}}
      {{if .Infrastructure}}
      ## Infrastructure: {{.Infrastructure.Provider}}

      {{range .Infrastructure.Resources}}- {{.Name}} ({{.Kind}})
      {{end}}{{end}}
      {{if and .AI .AI.Models}}

      ## AI Models

      {{range .AI.Models}}- {{.Name}} ({{.Kind}}) v{{.Version}}
      {{end}}{{end}}
variables:
  - name: AgentGuidelines
    type: "[]string"
    default:
      - "Follow the architecture pattern"
      - "Write clean, idiomatic code for the target language"
      - "Handle errors explicitly"
      - "Write tests for new code"
      - "Keep functions focused and small"
      - "Document public APIs"
`,

	"windsurf": `name: windsurf
kind: compiler
version: "1.0.0"
target: windsurf
files:
  - path: ".windsurfrules"
    kind: rules
    template: |
      # Windsurf Rules

      {{if .Project.Name}}project_name: {{.Project.Name}}{{end}}
      {{if .Architecture.Pattern}}architecture: {{.Architecture.Pattern}}{{end}}

      ## Instructions

      You are working on a project with the following structure:

      {{if .Modules}}
      ### Modules

      {{range .Modules}}- {{code .Name}} at {{code .Path}}
      {{if .Description}}  > {{.Description}}{{end}}
      {{end}}{{end}}
      {{if .Services}}
      ### Services

      {{range .Services}}- **{{.Name}}** ({{.Kind}}, port {{.Port}})
      {{range .Endpoints}}  - {{.Method}} {{.Path}} -> {{.Action}}
      {{end}}{{end}}{{end}}
      {{if .Components}}
      ### Components

      {{range .Components}}- {{code .Name}} [{{.Kind}}] module: {{code .Module}}
      {{end}}{{end}}

      ## Style Rules

      - Use early returns
      - Prefer const/readonly where possible
      - Write self-documenting code with clear names
      - Handle all error paths
      - Keep functions under 50 lines

  - path: ".windsurf/context.md"
    kind: context
    template: |
      # Windsurf Context

      Additional project context for Windsurf AI.

      {{if .Project.Name}}Project: {{.Project.Name}} {{.Project.Version}}{{end}}

      {{if .Modules}}
      ## Module Dependency Map

      {{range .Modules}}{{if .Dependencies}}{{.Name}} depends on: {{join .Dependencies ", "}}{{end}}
      {{end}}{{end}}
      {{if .APIs}}
      ## API Endpoints

      {{range .APIs}}### {{.Name}} v{{.Version}} ({{.Protocol}})
      {{range .Endpoints}}- {{.Method}} {{.Path}}: {{.Summary}}
      {{end}}{{end}}{{end}}
variables:
  - name: StyleRules
    type: "[]string"
    default:
      - "Use early returns"
      - "Prefer const/readonly where possible"
      - "Write self-documenting code with clear names"
      - "Handle all error paths"
      - "Keep functions under 50 lines"
`,
	"opencode": `name: opencode
kind: compiler
version: "1.0.0"
target: opencode
files:
  - path: "AGENTS.md"
    kind: instructions
    template: |
      # AGENTS.md

      Instructions for OpenCode agents working on this project.

      {{if .Project.Name}}## Project: {{.Project.Name}}

      {{if .Project.Description}}{{.Project.Description}}{{end}}
      {{if .Project.Version}}Version: {{.Project.Version}}{{end}}
      {{end}}
      {{if .Architecture.Pattern}}## Architecture: {{.Architecture.Pattern}}

      {{if .Architecture.Principles}}Principles:
      {{range .Architecture.Principles}}- {{.}}
      {{end}}{{end}}{{end}}
      {{if .Modules}}## Modules

      {{range .Modules}}### {{.Name}}
      Path: {{code .Path}}
      {{if .Description}}{{.Description}}{{end}}
      {{if .Dependencies}}Dependencies: {{join .Dependencies ", "}}{{end}}

      {{end}}{{end}}
      {{if .Services}}## Services

      {{range .Services}}### {{.Name}} ({{.Kind}}, port {{.Port}})
      {{range .Endpoints}}- {{.Method}} {{.Path}} -> {{.Action}}
      {{end}}

      {{end}}{{end}}
      {{if .Components}}## Components

      {{range .Components}}- {{code .Name}} [{{.Kind}}] in {{code .Module}}
      {{end}}

      {{end}}
      {{if .APIs}}## APIs

      {{range .APIs}}### {{.Name}} v{{.Version}} ({{.Protocol}})
      {{range .Endpoints}}- {{.Method}} {{.Path}}: {{.Summary}}
      {{end}}{{end}}{{end}}
      {{if .Security}}## Security

      {{if .Security.Authentication}}- Auth: {{.Security.Authentication.Method}} via {{.Security.Authentication.Provider}}{{end}}
      {{if .Security.Authorization}}- Authorization: {{.Security.Authorization.Model}}{{end}}
      {{if .Security.Encryption}}- Encryption: in_transit={{.Security.Encryption.InTransit}}, at_rest={{.Security.Encryption.AtRest}}{{end}}

      {{end}}
      {{if .Deployment}}## Deployment: {{.Deployment.Strategy}}

      {{end}}
      {{if .Testing}}## Testing: {{.Testing.Strategy}}

      {{end}}
      ## Guidelines

      1. Follow the established architecture pattern
      2. Write clean, idiomatic code
      3. Handle all error paths explicitly
      4. Write tests for new functionality
      5. Keep functions focused and under 50 lines
      6. Use meaningful variable and function names
      7. Follow the module dependency structure

  - path: ".opencode/context.md"
    kind: context
    template: |
      # OpenCode Context

      {{if .Modules}}
      ## Dependency Graph

      {{range .Modules}}{{if .Dependencies}}{{.Name}} -> {{join .Dependencies ", "}}{{else}}{{.Name}} (root){{end}}
      {{end}}{{end}}
      {{if .Storage}}
      ## Storage

      {{range .Storage}}- {{.Name}} ({{.Type}}, {{.Provider}})
      {{end}}{{end}}
      {{if .Infrastructure}}
      ## Infrastructure: {{.Infrastructure.Provider}} ({{.Infrastructure.Region}})

      {{range .Infrastructure.Resources}}- {{.Name}} ({{.Kind}})
      {{end}}{{end}}
      {{if and .AI .AI.Models}}
      ## AI Models

      {{range .AI.Models}}- {{.Name}} ({{.Kind}} v{{.Version}})
      {{end}}{{end}}
      {{if .Documentation}}
      ## Design Documents

      {{range .Documentation.ADRs}}- ADR: {{.Title}}
      {{end}}{{range .Documentation.RFCs}}- RFC: {{.Title}}
      {{end}}{{end}}

  - path: ".opencode/rules.md"
    kind: rules
    template: |
      # OpenCode Rules

      {{if .Architecture.Pattern}}Architecture: {{.Architecture.Pattern}}{{end}}

      ## Code Rules

      1. Never use panic for error handling
      2. Always check error returns
      3. Use table-driven tests
      4. Keep package imports clean (no circular dependencies)
      5. Follow existing code patterns in the project
      6. Use dependency injection for testability

      {{if .Modules}}
      ## Module Boundaries

      {{range .Modules}}- {{code .Name}} should only depend on: {{join .Dependencies ", "}}
      {{end}}{{end}}
      {{if .Security}}
      ## Security Rules

      - Never hardcode secrets
      - Use environment variables for configuration
      - Validate all external input
      {{if and .Security.Encryption .Security.Encryption.InTransit}}- All network communication must use TLS
      {{end}}{{end}}
variables:
  - name: Guidelines
    type: "[]string"
    default:
      - "Follow the established architecture pattern"
      - "Write clean, idiomatic code"
      - "Handle all error paths explicitly"
      - "Write tests for new functionality"
      - "Keep functions focused and under 50 lines"
      - "Use meaningful variable and function names"
      - "Follow the module dependency structure"

  - name: CodeRules
    type: "[]string"
    default:
      - "Never use panic for error handling"
      - "Always check error returns"
      - "Use table-driven tests"
      - "Keep package imports clean (no circular dependencies)"
      - "Follow existing code patterns in the project"
      - "Use dependency injection for testability"
`,
}
