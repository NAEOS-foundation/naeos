# NAEOS — Product Hunt Listing

Everything you need to fill out the Product Hunt submission form. Copy-paste directly into the "Create a launch" page at producthunt.com.

---

## 1. Basic Info

| Field | Value |
|-------|-------|
| **Product name** | NAEOS |
| **Website URL** | https://naeos.dev |
| **Tagline** | See options below (PH limit: 60 characters) |
| **Topics** | Developer Tools, Artificial Intelligence, Open Source, Software Engineering |
| **Logo** | `brand/logo.svg` (GitHub avatar lockup; render as PNG at 240×240px on dark background) |
| **Gallery** | See "Gallery images" section below |
| **Maker** | Bayu Priatno — [@bayupriatno007](https://x.com/bayupriatno007) (see `maker-bio.md`) |
| **First comment** | See `first-comment.md` |

### Tagline options (pick one)

1. **Specify once. Build anywhere.** (28 chars) — official brand tagline; short and memorable.
2. **The engineering OS that builds software from your specs** (55 chars) — emphasizes the runtime, not just a generator.
3. **Turn specifications into production software — for humans and AI** (60 chars) — highlights the AI angle; use if targeting AI-forward audiences.
4. **Write the spec once. NAEOS builds, validates, and evolves your system** (60 chars) — most descriptive.

Recommended: **Option 1** if you want brand consistency, **Option 3** if you want to maximize AI-curious clicks.

---

## 2. Description (publish-ready)

Paste the following into the description field. Trim the "Features" bullets if PH truncates the preview — the first three paragraphs are the hook.

---

**NAEOS is a declarative engineering platform that transforms specifications into high-quality software systems through a consistent, validated, and extensible pipeline.**

Most teams start projects the same way: a spec document that slowly drifts out of sync with the code. NAEOS inverts this. You describe your system once in a YAML/JSON specification, and NAEOS builds an internal engineering model (NEIR), validates it, schedules a DAG of tasks, and generates production-quality code across Go, TypeScript, Python, Java, and Rust — with full traceability from intent to implementation.

And the spec stays the source of truth for the entire lifecycle. Watch mode re-runs the pipeline on every change, the diff engine shows what drifted, the migration engine evolves schemas across versions, and the artifact store keeps every generated output auditable.

**NAEOS is built for the AI era.** Its compiler transforms the engineering model into instruction sets for GitHub Copilot, Claude Code, Cursor, Gemini CLI, Codex, and OpenCode — so any AI assistant works from an accurate model of your system, not a guess. A built-in MCP server exposes the model to AI agents, and context bundles give LLMs the project summary they need in a single file.

### Features

- **Spec Language v2** — variable interpolation (`${var}`), environment resolution (`$env{VAR}`), cross-references (`$ref{path}`), multi-file composition (`$include{file}`), custom functions (`$fn{name(args)}`), and conditional sections (`$if` / `$endif`)
- **Multi-language generation** — Go, TypeScript, Python, Java, Rust from one specification
- **AI compiler** — instruction sets for 6 AI tools (Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode), MCP server, LLM-optimized context bundles
- **35+ CLI commands** — run, validate, compile, context, test, docgen, mcp, marketplace, diff, watch, migrate, and more
- **NEIR-aware LSP server** — autocomplete, real-time diagnostics, hover docs, and go-to-definition for `.naeos.yaml` files
- **Profile & plugin marketplaces** — 5 built-in industry profiles (SaaS, AI Agent, FinTech, Healthcare, Government), WASM plugin SDK with official example plugins
- **Governance built-in** — policy evaluator, artifact review, and full audit trail
- **Enterprise plumbing** — PostgreSQL/MySQL/SQLite, WebSocket, event sourcing, distributed task execution, stage caching, pipeline profiling
- **Open source** — Apache 2.0, Go, single static binary, Docker-ready

### Getting started

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos create          # interactive wizard — enter your project name
cd my-app
naeos run --input-file spec.yaml
```

Docs: https://docs.naeos.dev — GitHub: https://github.com/NAEOS-foundation/naeos
Whitepaper: https://naeos.dev/whitepaper

---

## 3. Gallery Images

PH requires at least 3 gallery images; minimum size 750×750px, recommended 1600×900px (16:9) or square. Dark backgrounds match the brand (`#05050a`).

All five gallery images + the logo are already generated in `assets/` (PNG, 1600×900, brand-accurate). Regenerate anytime with `node launch/producthunt/assets/generate.js` (requires `sharp`: `npm i sharp`).

| # | Image | File |
|---|-------|------|
| 1 | **Hero cover** — logo, tagline, real `naeos run` output in a terminal card | `assets/01-hero-cover.png` |
| 2 | **Pipeline diagram** — Input → Core Layer → Generation → Output with traceability bar | `assets/02-pipeline-diagram.png` |
| 3 | **CLI screenshot** — real `naeos run` pipeline output on a dark terminal | `assets/03-cli-pipeline.png` |
| 4 | **AI compiler output** — `naeos context` bundle + 6 compiled AI instruction sets | `assets/04-ai-compiler.png` |
| 5 | **LSP / VS Code** — `.naeos.yaml` with hover tooltip, diagnostics, and status bar | `assets/05-lsp-editor.png` |
| — | **Logo** (upload in the logo field, 240×240px) | `assets/logo-240.png` |

Do not put links in gallery images — PH strips them and it looks unprofessional.

---

## 4. Links

- Website: https://naeos.dev
- GitHub: https://github.com/NAEOS-foundation/naeos
- Documentation: https://docs.naeos.dev
- Whitepaper (EN): https://naeos.dev/whitepaper
- Release notes (v3.0.0): https://github.com/NAEOS-foundation/naeos/releases
- Community (Discord): https://discord.gg/WnUWmm7XMv