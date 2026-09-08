# HN Post — Run a runnable NAEOS CLI demo (CLI demo day)

Today's objective: convert Awareness -> First successful run (funnel stage).
Audience: AI Engineers + Software Architects.
Channel: Hacker News (technical substance, no hype).

## Title options

- "Show HN: I made a 2-command CLI that turns one spec into Go + TypeScript output"
- "Show HN: Specification-Driven Engineering that you can run yourself"

## Post body

AI coding tools are fast. But I still want something I can verify,
re-run, and keep around for a long time. So I built a CLI that turns a
single YAML spec into a runnable build.

Try it yourself — about two commands:

```bash
git clone https://github.com/NAEOS-foundation/naeos
cd naeos
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

What happens:

1. Validates examples/demo-cli/spec.yaml
2. Generates an AI context bundle (context.md)
3. Runs the NAEOS pipeline
4. Generates a project with Go (go.mod) + TypeScript (package.json)
5. Runs a smoke test on the expected output files

Why this interests me (and maybe you):

- Pipeline behavior becomes observable, not just a claim
- AI context is derived from the spec, not prompt copy-paste
- Everything lands in .run/summary.md for inspection

The approach I chose (not claiming "best"):

- Specification-Driven Engineering: one source of intent
- Validation + a deterministic pipeline (steps 2 and 3)
- An AI compiler maps the spec to context for tools like
  GitHub Copilot, Cursor, Claude Code, Gemini CLI, Codex, OpenCode

Open questions for the community:

- When does a team need this — before or after AI writes code?
- How do you keep AI-generated code traceable back to the spec?
- What do you do today so architecture does not drift?

Open to technical criticism. Repo: https://github.com/NAEOS-foundation/naeos

## CTA

Clone, run `run-demo.sh`, then tell me: where do you most need
spec-driven engineering?

## Visual concept

CLI output of `run-demo.sh --summary` plus the .run/generated/ layout
(go.mod, package.json, context.md). Show real output, no hype.

## Supporting GitHub reference

- examples/demo-cli/README.md — how to run
- examples/demo-cli/run-demo.sh, examples/demo-cli/spec.yaml
- Commit 2f0c6db (feat: publish runnable CLI demo) and ecb4706 (printable guide)
- Release v3.4.0

## Expected outcome

- Several clones and run-demo.sh results (First successful run)
- Technical comments/feedback -> potential first Issue / Discussion
- Next experiment: NAEOS_DEMO_OUTPUT_DIR override to test isolated runs
