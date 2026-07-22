# NES-054: Prompt Library

**Status:** Draft v0.1
**Created:** 2026-07-19
**Package:** `internal/promptlib`

## Summary

The Prompt Library provides a centralized, template-based system for managing AI prompts used by the LLM service and compiler adapters. It replaces hardcoded prompt strings with configurable YAML templates that can be customized without modifying Go source code.

## Motivation

Previously, prompts were hardcoded in two locations:

1. **LLM Service** (`internal/ai/llm.go`): 3 prompts using `fmt.Sprintf` for spec enrichment, suggestions, and architecture explanation.
2. **Compiler Adapters** (`internal/compiler/adapters/`): 6 adapters each building instruction files via inline `strings.Builder` code.

This fragmentation meant:
- No way to customize prompts without changing Go code
- No shared defaults between adapters
- Inconsistent output formats across adapters
- No mechanism for user-provided prompt overrides

## Architecture

```text
┌─────────────────────────────────────────────────┐
│              Prompt Library                      │
├─────────────────┬───────────────────────────────┤
│   LLM Prompts   │    Compiler Templates          │
│   (3 builtins)  │    (6 builtins)                │
├─────────────────┼───────────────────────────────┤
│   Library.go    │    Template.go                  │
│   (load/render) │    (NEIR context builder)      │
├─────────────────┴───────────────────────────────┤
│   User Overrides (.naeos/prompts/)               │
└─────────────────────────────────────────────────┘
```

## Components

### Library (`internal/promptlib/library.go`)

The core registry that manages prompt templates.

```go
lib := promptlib.NewWithDefaults()
// or with overrides directory:
lib, err := promptlib.New(promptlib.WithOverridesDir(".naeos/prompts"))
```

### Template Types

#### LLM Prompt

```yaml
name: enrich-spec
kind: llm
version: "1.0.0"
description: "Enrich a NAEOS specification with best practices"
system: |
  You are a platform engineering expert...
user: |
  Analyze this NAEOS specification...
  {{.SpecContent}}
variables:
  - name: SpecContent
    type: string
    required: true
constraints:
  max_tokens: 2048
  temperature: 0.3
```

#### Compiler Template

```yaml
name: claude
kind: compiler
version: "1.0.0"
target: claude
files:
  - path: "CLAUDE.md"
    kind: instructions
    template: |
      # CLAUDE.md
      {{if .Project.Name}}## Project: {{.Project.Name}}{{end}}
      {{if .Architecture.Pattern}}## Architecture: {{.Architecture.Pattern}}{{end}}
variables:
  - name: Guidelines
    type: "[]string"
    default:
      - "Follow the architecture pattern"
```

### Template Functions

Available in compiler templates:

| Function | Usage | Description |
|----------|-------|-------------|
| `join` | `{{join .Dependencies ", "}}` | Join string slice |
| `upper` | `{{upper .Name}}` | Uppercase string |
| `lower` | `{{lower .Name}}` | Lowercase string |
| `code` | `{{code .Name}}` | Wrap in backticks |
| `bt` | `{{bt}}` | Literal backtick |
| `json` | `{{json .}}` | JSON marshal |
| `yaml` | `{{yaml .}}` | YAML marshal |
| `default` | `{{default "fallback" .Val}}` | Default value |
| `len` | `{{len .Items}}` | Length of slice/string |
| `contains` | `{{contains .Str "sub"}}` | String contains |
| `hasPrefix` | `{{hasPrefix .Path "./"}}` | String prefix check |
| `replace` | `{{replace .Str "old" "new"}}` | String replace |

### NEIR Context Variables

Compiler templates receive a `NEIRContext` with these view models:

| Variable | Type | Fields |
|----------|------|--------|
| `.Project` | projectView | Name, Description, Version |
| `.Architecture` | architectureView | Pattern, Principles |
| `.Modules` | []moduleView | Name, Path, Description, Dependencies |
| `.Components` | []componentView | Name, Kind, Module |
| `.Services` | []serviceView | Name, Kind, Port, Endpoints |
| `.APIs` | []apiView | Name, Version, Protocol, Endpoints |
| `.Security` | *securityView | Authentication, Authorization, Encryption |
| `.Deployment` | *deploymentView | Strategy, Environments |
| `.Testing` | *testingView | Strategy, Frameworks |
| `.Storage` | []storageView | Name, Type, Provider |
| `.Infrastructure` | *infrastructureView | Provider, Region, Resources |
| `.AI` | *aiView | Models |
| `.Documentation` | *documentationView | ADRs, RFCs |

## CLI Commands

```bash
# List all templates (code + prompt-llm + prompt-compiler)
naeos template list

# Filter by kind
naeos template list --kind prompt-llm
naeos template list --kind prompt-compiler
naeos template list --kind code

# Show prompt details
naeos template show enrich-spec
naeos template show claude
```

## Built-in Prompts

### LLM Prompts (3)

| Name | Description | MaxTokens | Temperature |
|------|-------------|-----------|-------------|
| `enrich-spec` | Enrich spec with best practices | 2048 | 0.3 |
| `generate-suggestions` | Generate improvement suggestions | 2048 | 0.3 |
| `explain-architecture` | Explain architecture pattern | 1024 | 0.3 |

### Compiler Templates (6)

| Name | Target | Output Files |
|------|--------|-------------|
| `copilot` | GitHub Copilot | `.github/copilot-instructions.md`, `copilot-context.md`, `copilot-rules.md` |
| `claude` | Claude Code | `CLAUDE.md`, `.claude/context.md`, `.claude/rules.md` |
| `cursor` | Cursor | `.cursorrules`, `.cursor/context.md` |
| `gemini` | Gemini CLI | `.gemini/CONFIG.md`, `.gemini/context.md` |
| `codex` | Codex | `AGENTS.md`, `.codex/context.md` |
| `opencode` | OpenCode | `AGENTS.md`, `.opencode/context.md`, `.opencode/rules.md` |

## User Overrides

Create YAML files in `.naeos/prompts/` to override built-in prompts:

```
.naeos/prompts/
├── llm/
│   └── enrich-spec.yaml     # Overrides built-in enrich-spec
└── compiler/
    └── claude.yaml           # Overrides built-in claude template
```

Override files use the same YAML format as built-in templates. They take precedence over built-in versions.

## Integration Points

### LLM Service

The `LLMService` accepts an optional library:

```go
lib := promptlib.NewWithDefaults()
llm := ai.NewLLMService(config, lib)
```

When a library is provided, prompts are rendered from templates. When nil, the original hardcoded prompts are used as fallback.

### Compiler

The `Compiler` accepts a library for template-based compilation:

```go
lib := promptlib.NewWithDefaults()
c := compiler.NewWithLibrary(lib)
c.Register(adapters.NewClaudeAdapter(lib))
```

When adapters have a library, they use template rendering. When nil, the original string builder code is used as legacy fallback.

## Design Decisions

1. **Backward Compatible**: All existing APIs accept nil library and fall back to original behavior
2. **Override Directory**: User prompts in `.naeos/prompts/` override built-ins
3. **YAML Format**: Human-readable, consistent with NAEOS specification format
4. **NEIR Context Builder**: Converts raw NEIR models to template-friendly view models
5. **Builtin Embedding**: Built-in prompts stored as Go string constants (no filesystem dependency)

## Future Work

- [ ] `naeos template override <name>` CLI command for editing overrides
- [ ] `naeos template validate` to check YAML template syntax
- [ ] Hot-reload of overrides directory via watch mode
- [ ] Template inheritance (base + extends pattern)
- [ ] Version-aware template selection
