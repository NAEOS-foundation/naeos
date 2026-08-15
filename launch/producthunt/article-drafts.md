# NAEOS — HackerNoon PoU & DEV.to Launch Articles

Paste-ready drafts for the HackerNoon "Proof of Usefulness" submission and a
DEV.to article. Fill the bracketed placeholders with final numbers on launch
day. Traction data current as of Aug 16, 2026: 391 commits, ~6 stars, 6 forks.

---

## 1. HackerNoon Proof of Usefulness — submission fields

Go to `proofofusefulness.com` → **Submit Project**:

| Field | Value |
|-------|-------|
| Project name | NAEOS |
| URL | https://github.com/NAEOS-foundation/naeos |
| Website | https://naeos.dev |
| Tech stack | Go, YAML/JSON, WASM plugins, LSP, MCP server |
| Traction proof | 391 commits, [N] stars, [N] forks, v3.0.0 released [date] |
| Short description | Declarative engineering platform: write a spec once, get validated multi-language code (Go/TS/Py/Java/Rust), governance, docs, and AI instruction sets — for humans and AI |

Then convert to a HackerNoon draft and polish with the article below.

---

## 2. Article — HackerNoon draft / DEV.to post (shared)

**Title options:**
- "NAEOS: The Open-Source Engineering OS That Turns Specs into Software"
- "Specify Once, Build Anywhere: What I Learned Building a Spec-Driven Engineering Runtime"
- "Why I Built an Engineering OS Instead of Another Code Generator"

**Body:**

> **TL;DR** — NAEOS is an open-source (Apache 2.0) declarative engineering
> platform written in Go. You describe your system once in a YAML/JSON spec;
> it validates, builds an internal engineering model (NEIR), and generates
> production-quality code in Go, TypeScript, Python, Java, and Rust — plus
> governance checks, docs, and AI context. The spec stays the source of truth
> for the whole lifecycle.
>
> `curl -fsSL https://naeos.dev/install.sh | sh`
>
> ## The problem
>
> Every project starts with a great spec document. Six months later, the spec
> and the codebase are strangers. Nobody maintains the spec because the code
> "is the source of truth." Then the code base grows, docs rot, and every new
> engineer — human or AI — reverse-engineers the whole system.
>
> ## The flip
>
> What if the specification were the source of truth, and the tooling made
> keeping it accurate the path of least resistance?
>
> NAEOS parses your spec, normalizes it, resolves references (`$ref`, `$include`,
> `$fn`, `$if`), validates it, and builds a NEIR model that captures project,
> architecture, domain, modules, services, APIs, storage, security, and AI config.
> A scheduler runs a DAG of generation tasks and produces real code — not a
> scaffold, but a consistent, validated, auditable pipeline.
>
> ## Built for the AI era
>
> NAEOS compiles the engineering model into instruction sets for 6 AI tools —
> Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode — and exposes an MCP
> server. AI assistants work from an accurate model of your system, not a guess.
>
> ## What's in v3.0.0
>
> - Pipeline profiling + stage caching (rebuilds reuse cached results)
> - NEIR-aware LSP server (autocomplete, diagnostics, go-to-definition for `.naeos.yaml`)
> - Distributed builds (`naeos build --distributed --workers N`)
> - VS Code extension generation (`naeos dx vscode-gen`)
> - Official example plugins + microservices Go starter template
>
> ## Try it
>
> ```bash
> curl -fsSL https://naeos.dev/install.sh | sh
> naeos create          # interactive wizard — enter your project name
> cd my-app
> naeos run --input-file spec.yaml
> ```
>
> Docs: https://docs.naeos.dev · GitHub: https://github.com/NAEOS-foundation/naeos
> Community: https://discord.gg/WnUWmm7XMv
>
> **Specify once. Build anywhere.**
>
> *NAEOS is open source under Apache 2.0. Feedback, issues, and PRs welcome.*

---

## 3. dev.to post formatting notes

- Add tags: `opensource`, `go`, `devtools`, `spec` (max 4)
- Set `canonical_url` to the HackerNoon version if cross-posting (dev.to rewards
  canonical links; posting first on dev.to avoids duplicate-content issues)
- Use a code block for the install command and keep headers for skimmability
- Embed 1–2 gallery images from `launch/producthunt/assets/` (hero, pipeline diagram)
