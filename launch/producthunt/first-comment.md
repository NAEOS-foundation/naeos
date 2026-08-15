# First Comment — Maker's Kickoff

Paste this as your first comment immediately after the product goes live (within the first hour — it anchors the discussion and answers the "why" behind the launch).

---

Hi Product Hunt! [Maker Name] here, the maker of NAEOS. I'm genuinely excited (and a little nervous) to share this one.

**The problem I kept hitting:** every project starts with a great spec document, and six months later the spec and the code are strangers. Nobody maintains the spec because the code "is the source of truth." But then the code base grows, the docs rot, and every new engineer or AI assistant has to reverse-engineer the whole system.

**The idea:** flip it. Make the specification the source of truth, and make the tooling so good that keeping the spec accurate is the path of least resistance. That's NAEOS — you describe your system once, and it parses, validates, builds an internal engineering model (NEIR), and generates real code in Go, TypeScript, Python, Java, and Rust. Not a scaffold — a consistent, validated, auditable pipeline.

**What's in v3.0.0 (the version we're launching with):**
- Pipeline profiling, memory analysis, and stage caching — rebuilds reuse cached results
- A NEIR-aware LSP server with autocomplete, diagnostics, and go-to-definition for `.naeos.yaml`
- Distributed builds with `naeos build --distributed --workers N`
- VS Code extension generation with `naeos dx vscode-gen`
- Official example plugins and a microservices Go starter template

**The AI angle:** we ship a compiler that turns the engineering model into instruction sets for 6 AI tools (Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode), plus an MCP server — so AI assistants work from an accurate model of your system instead of guessing.

Try it in 30 seconds:

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos create my-app
naeos run --input-file spec.yaml
```

It's open source (Apache 2.0) — the whole roadmap is public. I'd love feedback on what's missing, what's over-engineered, and what spec-driven workflow you'd want us to tackle next. Ask me anything — I'll be here all day.