---
title: Context Bundles
description: Hasilkan ringkasan proyek yang optimal untuk konsumsi LLM.
weight: 12
---

Context Bundles menghasilkan ringkasan proyek dalam format yang optimal untuk konsumsi LLM. Bundle ini menyediakan pemahaman yang terstruktur dan ringkas tentang proyek Anda kepada asisten AI.

## Penggunaan CLI

```bash
# Hasilkan bundle markdown
naeos context --input-file spec.yaml

# Hasilkan bundle plain text
naeos context --input-file spec.yaml --output plain

# Hasilkan bundle JSON
naeos context --input-file spec.yaml --output json

# Simpan ke file
naeos context --input-file spec.yaml --output-file context.md

# Hasilkan dari model NEIR yang sudah ada
naeos context --neir-file neir.json --output-file context.md
```

## Format Output

### Markdown

Format markdown menyediakan ringkasan yang terstruktur dan mudah dibaca:

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

Format plain text ringkas dan cocok untuk konteks dengan batas token:

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

Format JSON menyertakan metadata lengkap dan dapat di-parse oleh mesin:

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

## Penggunaan Go API

### Dari Spesifikasi

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

// Target yang didukung
targets := contextBundle.SupportedTargets()
// ["markdown", "plain"]
```

### Dari NEIR

```go
contextBundle := gen.GenerateFromNEIR(neir)

// Akses data bundle
fmt.Println(contextBundle.Project)   // "my-app"
fmt.Println(contextBundle.Modules)   // []ModuleContext
fmt.Println(contextBundle.Services)  // []ServiceContext
```

## Struktur Bundle

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

Setiap bundle menyertakan metadata untuk ketelusuran:

```json
{
  "generated_by": "naeos-context-bundle",
  "module_count": "2",
  "service_count": "1"
}
```

## Integrasi dengan AI Tools

Context bundle dirancang untuk bekerja dengan asisten coding AI. Hasilkan bundle dan tempatkan di lokasi yang sesuai untuk tool Anda:

```bash
# Untuk GitHub Copilot
naeos context --input-file spec.yaml --output-file .github/copilot-context.md

# Untuk Claude Code
naeos context --input-file spec.yaml --output-file CLAUDE.md

# Untuk Cursor
naeos context --input-file spec.yaml --output-file .cursorrules
```

### Integrasi Go

```go
contextBundle := gen.GenerateFromSpec(doc)
copilotInstructions := contextBundle.ToMarkdown()

// Simpan ke lokasi AI tool
os.WriteFile(".github/copilot-context.md", []byte(copilotInstructions), 0644)
```

## Best Practices

- **Regenerasi setelah perubahan spec** — Context bundle adalah snapshot; regenerasi saat spesifikasi berubah.
- **Pertahankan bundle ringkas** — Gunakan format plain text untuk konteks dengan batas token.
- **Masukkan ke version control** — Commit bundle yang dihasilkan agar asisten AI selalu memiliki konteks terkini.
- **Gunakan dengan `naeos compile`** — Context bundle melengkapi output AI compiler untuk integrasi AI yang lengkap.
