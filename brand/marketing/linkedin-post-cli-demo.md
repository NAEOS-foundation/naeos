# LinkedIn Post — CLI demo day (Specification-Driven Engineering you can run)

Today's objective: convert Awareness -> First successful run (funnel stage).
Audience: Engineering Leaders + Architectural / Platform engineers.
Channel: LinkedIn (native, structure per Mode 24).

## Hook
Most teams that adopt AI-assisted development can generate code fast.
What they struggle with is keeping the system coherent around it.

## Problem
Architecture drifts across docs, tickets, prompts, and PRs. Policies get
skipped. AI context goes stale. And no single layer validates the whole.

## Insight
The bottleneck is not generation speed. It is the engineering model that
generation is supposed to serve.

## NAEOS approach
NAEOS treats the specification as the source of truth:
- describe the system once (YAML spec)
- normalize, resolve, validate
- generate artifacts and AI context from the same model

## Technical detail
We made this runnable locally. Two commands:

```bash
 go build -o naeos ./cmd/naeos
 ./examples/demo-cli/run-demo.sh
```

The demo validates `spec.yaml`, writes an AI context bundle, runs the
pipeline, and generates a Go + TypeScript project with a smoke test.
Everything lands in `.run/summary.md` for inspection.

## Lesson
A specification-driven pipeline makes "what you intended" and "what you
built" comparable — and verifiable.

## CTA
Try it yourself, then tell us where you most need spec-driven engineering.
Repo: https://github.com/NAEOS-foundation/naeos