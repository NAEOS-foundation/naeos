# Reddit Post — CLI demo day (founder journey)

Today's objective: convert Awareness -> First successful run (funnel stage).
Audience: AI engineers, builders, OSS developers.
Channel: Reddit (founder-journey format per Mode 22, no ads).

## Subreddit candidates
- r/ExperiencedDevs
- r/AI_Agents
- r/coding
- r/opensource
- r/programming

## Title (pick one)
- "I've been experimenting with a different approach to AI-assisted engineering"
- "I built a CLI that turns one spec into validated Go + TypeScript output"

## Body

I've been experimenting with a different approach to AI-assisted
engineering.

The problem I keep hitting: AI tools can generate code quickly, but the
system around the code — architecture, policies, AI context — tends to
drift. Docs, tickets, prompts, PRs each carry a slightly different version
of the truth.

So I tried making engineering intent machine-readable. One YAML spec.
Then: normalize, resolve, validate, generate. From the same model, produce
both artifacts and AI context.

I made this runnable locally so it is not just a website claim:

```bash
 go build -o naeos ./cmd/naeos
 ./examples/demo-cli/run-demo.sh
```

What it does:
1. Validates `examples/demo-cli/spec.yaml`
2. Writes an AI context bundle (`context.md`)
3. Runs the pipeline
4. Generates a Go + TypeScript project
5. Smoke-tests the expected output files

Everything lands in `.run/summary.md`.

The AI compiler maps the spec to context for GitHub Copilot, Cursor,
Claude Code, Gemini CLI, Codex, OpenCode.

Open questions I am wrestling with (honest ones):
- When does a team need this — before or after AI writes code?
- How do you keep AI-generated code traceable to the spec?
- What do you use today to keep architecture from drifting?

Repo: https://github.com/NAEOS-foundation/naeos \
Open to technical criticism — especially on the validation and traceability
model.

## CTA
Clone and run it, then tell me the gaps you see.