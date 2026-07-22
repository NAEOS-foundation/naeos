# ADR-007: Prompt Library

> Status: Accepted
> Date: 2026-07-20

## Context

NAEOS generates AI instruction sets for multiple coding assistants (Copilot, Claude, Cursor, Gemini, Codex, OpenCode) from NEIR specifications. Each assistant expects a specific prompt format and style. As the compiler adapter set grows, maintaining prompt templates directly in Go code becomes brittle: every prompt change requires a recompile, and non-developer team members cannot contribute prompt improvements. Key requirements:

- Centralized storage of prompt templates in a human-readable format
- Template composition (shared snippets, variable substitution)
- Backward compatibility: existing hardcoded prompts must continue to work
- Support for custom template functions (JSON/YAML formatting, backtick escaping)
- Ability to add new prompt templates without modifying Go code

## Decision

Introduce a **Prompt Library** system: YAML-based prompt templates stored in `prompts/builtin/`, with a Go manifest registry, and a `text/template` rendering engine extended with custom functions.

Each template is a YAML file with `id`, `kind`, `description`, and a `template` field containing Go template syntax. Templates support: `{{ .NEIR }}`, `{{ .Context }}`, `{{ .Target }}`, and custom functions like `{{ json . }}`, `{{ bt }}` (backtick), `{{ code "lang" "body" }}`.

The `promptlib.Library` type loads templates at startup, caches parsed templates, and renders them on demand. If no library is configured, the system falls back to the previous hardcoded prompt builder functions.

## Consequences

### Positive

- Prompts can be edited, reviewed, and version-controlled independently of Go code
- Non-developers (technical writers, AI prompt engineers) can contribute prompt improvements
- New compiler adapter targets can be added by dropping in a new YAML file
- Custom template functions keep prompt syntax clean without leaking Go logic into templates

### Negative

- Template rendering has a runtime cost (parsing, variable expansion) vs. compiled string constants
- Malformed YAML templates cause startup failures; validation is needed
- Template function set must be maintained and documented for template authors

### Mitigations

- Templates are parsed once at startup (not on every render), minimizing runtime overhead
- A `naeos template list --kind compiler` command validates all templates on load
- `naeos template show <id>` renders a template with sample data, allowing authors to preview changes
- Template functions are documented inline and listed via `naeos template functions`

## Notes

- Implementation: `internal/promptlib/` — library, manifest, template functions, builtin loaders
- Builtin prompts directory: `prompts/builtin/` (11 YAML files + manifest.yaml)
- Related NES document: NES-054 (Prompt Library)
- Backward compatibility: `nil` library delegates to `buildLLMPrompt()` in `internal/ai/llm.go`
