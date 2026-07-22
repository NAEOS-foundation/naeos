---
title: Context Bundles
description: Generate LLM-optimized project summaries for AI coding assistants.
weight: 12
---

Context Bundles generate project summaries in a format optimized for LLM consumption. These bundles provide AI assistants with a concise, structured understanding of your project.

## CLI Usage

```bash
# Generate markdown bundle
naeos context --input-file spec.yaml

# Generate plain text bundle
naeos context --input-file spec.yaml --output plain

# Generate JSON bundle
naeos context --input-file spec.yaml --output json

# Save to file
naeos context --input-file spec.yaml --output-file context.md

# Generate from existing NEIR model
naeos context --neir-file neir.json --output-file context.md
```

## Output Formats

### Markdown

The markdown format provides a structured, readable summary:

```markdown
# my-app — AI Context Bundle

## Summary
Project: my-app; Modules: auth, api; Services: 1

## Modules

- **auth** (`./auth`)
  Dependencies: core
- **api** (`./api`)
  Dependencies: auth

## Services

- **gateway** (kind=http, port=8080)
  - GET /users → listUsers
  - POST /users → createUser

## Languages
go, typescript
```

### Plain Text

The plain text format is compact and suitable for token-limited contexts:

```text
Project: my-app
Modules: 2, Services: 1
Languages: go, typescript
  Module: auth (./auth)
    deps: core
  Module: api (./api)
    deps: auth
  Service: gateway kind=http port=8080
```

### JSON

The JSON format includes full metadata and is machine-parseable:

```json
{
  "project": "my-app",
  "modules": [
    { "name": "auth", "path": "./auth", "dependencies": ["core"] },
    { "name": "api", "path": "./api", "dependencies": ["auth"] }
  ],
  "services": [
    { "name": "gateway", "kind": "http", "port": 8080 }
  ],
  "languages": ["go", "typescript"]
}
```

## Go API Usage

### From Specification

```go
import (
    "github.com/NAEOS-foundation/naeos/internal/context/bundle"
    "github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

p := parser.NewParser()
doc, err := p.Parse(specContent)
if err != nil {
    log.Fatal(err)
}

gen := bundle.NewGenerator(nil)
contextBundle := gen.GenerateFromSpec(doc)

// Output markdown
markdown := contextBundle.ToMarkdown()

// Output plain text
plainText := contextBundle.ToPlainText()

// Supported targets
targets := contextBundle.SupportedTargets()
// ["markdown", "plain"]
```

### From NEIR

```go
contextBundle := gen.GenerateFromNEIR(neir)

// Access bundle data
fmt.Println(contextBundle.Project)   // "my-app"
fmt.Println(contextBundle.Modules)   // []ModuleContext
fmt.Println(contextBundle.Services)  // []ServiceContext
```

## Bundle Structure

```go
type Bundle struct {
    Project   string
    Summary   string
    Modules   []ModuleContext
    Services  []ServiceContext
    Languages []string
    Targets   []string
    NEIR      string
    Raw       string
    Metadata  map[string]string
}

type ModuleContext struct {
    Name         string
    Path         string
    Description  string
    Dependencies []string
}

type ServiceContext struct {
    Name      string
    Kind      string
    Port      int
    Endpoints []EndpointContext
}

type EndpointContext struct {
    Method string
    Path   string
    Action string
}
```

## Metadata

Each bundle includes metadata for traceability:

```json
{
  "generated_by": "naeos-context-bundle",
  "module_count": "2",
  "service_count": "1"
}
```

## AI Tool Integration

Context bundles are designed to work with AI coding assistants. Generate a bundle and place it in the appropriate location for your tool:

```bash
# For GitHub Copilot
naeos context --input-file spec.yaml --output-file .github/copilot-context.md

# For Claude Code
naeos context --input-file spec.yaml --output-file CLAUDE.md

# For Cursor
naeos context --input-file spec.yaml --output-file .cursorrules
```

### Go Integration

```go
contextBundle := gen.GenerateFromSpec(doc)
copilotInstructions := contextBundle.ToMarkdown()

// Save to AI tool location
os.WriteFile(".github/copilot-context.md", []byte(copilotInstructions), 0644)
```

## Best Practices

- **Regenerate after spec changes** — Context bundles are snapshots; regenerate when your spec evolves.
- **Keep bundles concise** — Use plain text format for token-limited contexts.
- **Include in version control** — Commit generated bundles so AI assistants always have current context.
- **Use with `naeos compile`** — Context bundles complement AI compiler output for full AI integration.
