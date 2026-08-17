# Launch Announcement Newsletter

Ready-to-send newsletter content for the NAEOS Product Hunt launch.
Copy into your newsletter platform (Substack, Mailchimp, Buttondown, etc.).

---

## Version 1: Short & Punchy (Dev Community)

**Subject:** NAEOS is live on Product Hunt — specify once, build anywhere

**Preview text:** Open-source engineering OS that turns one spec into Go, TS, Python, Java, Rust code + AI instruction sets.

---

Hey,

We just launched **NAEOS** on Product Hunt — an open-source declarative engineering platform that makes the specification the source of truth.

**The problem:** every project starts with a great spec document, and six months later the spec and the code are strangers. Nobody maintains the spec because the code "is the source of truth." But then the codebase grows, the docs rot, and every new engineer or AI assistant has to reverse-engineer the whole system.

**The solution:** NAEOS flips it. You describe your system once in YAML/JSON, and it:

- Parses, validates, and builds an internal engineering model (NEIR)
- Generates production code in **Go, TypeScript, Python, Java, and Rust**
- Compiles the model into instruction sets for **6 AI tools** (Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode)
- Exposes an **MCP server** for AI agent integration
- Enforces governance, audit trails, and compliance (SOC 2, HIPAA, GDPR)

**What's new in v3.1.0:**

- Pipeline caching — rebuilds reuse cached results for faster iteration
- Run-level profiling — per-stage timing + pprof for performance analysis
- Architecture patterns — monolithic, microservices, serverless as first-class concepts
- WASM plugin hardening — sandboxed plugin execution with metadata validation

**Try it in 30 seconds:**

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos create
cd my-app
naeos run --input-file spec.yaml
```

**Upvote on Product Hunt if it resonates:**
https://www.producthunt.com/posts/naeos

Open source, Apache 2.0, single Go binary. The whole roadmap is public.

— Bayu, NAEOS founder

---

## Version 2: Longer Form (Newsletter/Blog)

**Subject:** Why we built NAEOS — and why we're launching it today

**Preview text:** The spec-driven engineering platform that treats your specification as the source of truth.

---

Hi everyone,

Today we're launching **NAEOS** on Product Hunt, and I wanted to share the story behind it.

### The Problem Nobody Talks About

Every engineering team has a spec document. Maybe it's a Google Doc, a Confluence page, or a YAML file in the repo. It captures the intent — the modules, the services, the architecture decisions.

But here's the dirty secret: **the spec and the code drift apart within weeks.**

Nobody updates the spec because "the code is the source of truth." But then:
- New engineers have to reverse-engineer the system from code
- AI assistants generate code that doesn't match your architecture
- Governance and compliance become afterthoughts
- Documentation rots silently

### The Idea: Make the Spec the Source of Truth

NAEOS inverts this. Instead of treating code as the source of truth, we make the specification the single source of truth — and build tooling so good that keeping the spec accurate is the path of least resistance.

You describe your system once in YAML/JSON:

```yaml
project: my-app
modules:
  - name: api
    path: ./api
  - name: auth
    path: ./auth
services:
  - name: rest-api
    kind: http
    port: 8080
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]
```

Then NAEOS:

1. **Parses** the spec with variable interpolation, cross-references, and multi-file composition
2. **Normalizes** it to canonical form with defaults and type validation
3. **Resolves** dependencies and detects circular references
4. **Builds NEIR** — a complete internal engineering model of your system
5. **Validates** structural correctness and governance policies
6. **Schedules** parallel execution based on the dependency graph
7. **Generates** production code in Go, TypeScript, Python, Java, and Rust
8. **Reviews** generated artifacts against quality rules
9. **Writes** everything to disk with a machine-readable manifest

### The AI Angle

Here's where it gets interesting for the AI era. The NEIR model captures your entire architecture — modules, services, dependencies, patterns, constraints. NAEOS compiles this into instruction sets for 6 AI tools:

- GitHub Copilot → `copilot-instructions.md`
- Claude Code → `CLAUDE.md`
- Cursor → `.cursorrules`
- Gemini CLI → `.gemini/CONFIG.md`
- Codex → `AGENTS.md`
- OpenCode → `AGENTS.md`

Your AI assistants don't just see individual files — they see the entire architecture, dependency graph, and design intent. Plus an MCP server exposes the model to AI agents programmatically.

### What's in v3.1.0

This release focuses on developer experience and platform hardening:

- **Pipeline caching** — `naeos run --cache-dir` caches stage results keyed by NEIR hash. Rebuilds with unchanged inputs skip regeneration.
- **Run-level profiling** — `--profile` captures per-stage timing, `--pprof` starts a live pprof server for heap and CPU inspection.
- **Architecture patterns** — monolithic, microservices, and serverless are now first-class in the NEIR model.
- **WASM plugin hardening** — plugins run in isolated memory with metadata validation.

### The Numbers

- **65+ CLI commands** — run, validate, compile, context, test, docgen, mcp, marketplace, diff, watch, migrate, and more
- **5 languages** — Go, TypeScript, Python, Java, Rust
- **6 AI tools** — Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode
- **86.9% test coverage** with race detector, fuzz testing, and benchmark gates
- **424 test files** across the codebase
- **10 GitHub Actions workflows** for CI/CD

### Try It

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos create
cd my-app
naeos run --input-file spec.yaml
```

### Links

- Website: https://naeos.dev
- Docs: https://docs.naeos.dev
- GitHub: https://github.com/NAEOS-foundation/naeos
- Product Hunt: https://www.producthunt.com/posts/naeos
- Discord: https://discord.gg/WnUWmm7XMv
- Whitepaper: https://naeos.dev/whitepaper

### Help Us Out

If this resonates, an upvote on Product Hunt helps enormously:
https://www.producthunt.com/posts/naeos

And if you've tried the quick start, a comment about your experience (good or bad) helps more than the upvote.

Thanks for reading — Bayu

---

## Version 3: Indonesian (Bahasa Indonesia)

**Subject:** NAEOS launching di Product Hunt — specify once, build anywhere

**Preview text:** Platform engineering open-source yang menjadikan spesifikasi sebagai sumber kebenaran.

---

Halo,

Hari ini kita launching **NAEOS** di Product Hunt — platform engineering deklaratif open-source yang menjadikan spesifikasi sebagai sumber kebenaran.

**Masalahnya:** setiap project dimulai dengan dokumen spesifikasi yang bagus, dan enam bulan kemudian spesifikasi dan kode menjadi asing. Tidak ada yang memelihara spesifikasi karena "kode adalah sumber kebenaran." Tapi kemudian kodebase membesar, dokumen membusuk, dan setiap engineer baru atau AI assistant harus mereverse-engineer seluruh sistem.

**Solusinya:** NAEOS membalik ini. Anda mendeskripsikan sistem Anda sekali dalam YAML/JSON, dan NAEOS:

- Parse, validasi, dan bangun model engineering internal (NEIR)
- Generate kode produksi dalam **Go, TypeScript, Python, Java, dan Rust**
- Compile model menjadi instruction set untuk **6 AI tools** (Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode)
- Expose **MCP server** untuk integrasi AI agent
- Enforce governance, audit trail, dan compliance (SOC 2, HIPAA, GDPR)

**Yang baru di v3.1.0:**

- Pipeline caching — rebuild menggunakan hasil cache untuk iterasi lebih cepat
- Run-level profiling — timing per stage + pprof untuk analisis performa
- Architecture patterns — monolithic, microservices, serverless sebagai konsep first-class
- WASM plugin hardening — eksekusi plugin terisolasi dengan validasi metadata

**Coba dalam 30 detik:**

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos create
cd my-app
naeos run --input-file spec.yaml
```

**Upvote di Product Hunt jika resonate:**
https://www.producthunt.com/posts/naeos

Open source, Apache 2.0, single Go binary. Seluruh roadmap publik.

— Bayu, NAEOS founder

---

## Sending Checklist

- [ ] Pilih versi (short/long/Indonesian)
- [ ] Replace `[Name]` dengan nama subscriber jika personal
- [ ] Verify semua links aktif
- [ ] Schedule kirim 14:35 WIB (00:05 PT) — bersamaan dengan PH post
- [ ] Track open rate dan click rate
- [ ] Follow up D+1 dengan thank you + stats
