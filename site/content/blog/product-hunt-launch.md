---
title: "We Launched NAEOS on Product Hunt — Here's What Happened"
description: "A transparent Product Hunt launch retrospective: the campaign structure, lessons to validate, and next experiments."
date: 2026-08-19
author: "NAEOS Foundation"
categories: ["launch", "community"]
---

On Tuesday, August 18, 2026, we prepared a Product Hunt launch for NAEOS. This retrospective documents the campaign structure and the questions we still need to validate. No launch metrics are claimed here unless they are linked to a source of record.

## Measurement Plan

The campaign should measure Product Hunt referrals, GitHub visits, quick-start completion, first successful run, and new discussions. The repository does not currently verify final Product Hunt upvotes, comments, traffic, downloads, or attribution, so those numbers are intentionally not reported here.

## What We Posted

The launch went live at 00:01 PT (14:01 WIB). Here's the sequence:

1. **PH post published** — name, tagline ("Specify Once. Build Anywhere."), 5 gallery images, and a 300-word description
2. **First comment** — a personal note from Bayu explaining the problem (spec/code drift) and why we built NAEOS
3. **X thread** — 6-tweet thread covering the problem, the solution, AI integration, v3.1.0 highlights, and the quick start
4. **LinkedIn post** — longer-form version for the professional network
5. **Discord + Slack** — "we're live" messages in `#launch-upvotes` channels
6. **Support list** — personal messages to 4 core contributors

## What People Said

### The Positive

### Feedback to Collect

Ask participants to evaluate the AI compiler, pipeline caching, specification language, and language coverage. Publish direct quotes only after obtaining permission and linking to the original discussion.

1. **Spec language complexity** — real syntax, real learning curve. We'll improve onboarding docs.
2. **AI integration** — measure whether the instruction set compiler leads to a reproducible quick-start run.
3. **Pipeline caching** — measure whether caching changes repeat-run behavior before expanding its scope.

## What We Learned

### Early Activation

Measure whether early visitors move from the launch page to the quick start and complete a first run. Do not infer activation from impressions alone.

### Comments and Reproducibility

A quick-start report with the command, result, and next question is more useful than an unqualified reaction.

### The Maker's Comment Sets the Tone

The opening comment should explain the problem (spec/code drift) before presenting NAEOS.

### Community Channels Need Attribution

Use channel-specific links or campaign parameters before claiming which community channel contributed traffic.

## What's Next

Based on the campaign hypotheses, here's what we're testing next:

1. **Spec language onboarding** — interactive tutorial, not just docs
2. **More AI tools** — Windsurf, Aider, Cline instruction sets
3. **Pipeline caching improvements** — cache across runs, not just within a session
4. **Better quick start** — a short, reproducible demo with documented prerequisites

Record validated feedback as GitHub issues with evidence and acceptance criteria. Track project priorities in the [roadmap](/roadmap/).

## Thank You

To everyone who reviews, tests, or challenges the project — thank you. This project exists because spec/code drift is a real problem, and we believe the specification should be the source of truth.

If you haven't tried NAEOS yet:

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos create
cd my-app
naeos run --input-file spec.yaml
```

Open source, Apache 2.0, single Go binary. The whole roadmap is public.

See you in the [community](https://discord.gg/WnUWmm7XMv).

---

*Launch metrics are intentionally omitted until a source of record is available.*
