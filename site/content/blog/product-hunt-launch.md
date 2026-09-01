---
title: "We Launched NAEOS on Product Hunt — Here's What Happened"
description: "A recap of our Product Hunt launch day: the numbers, the feedback, and what we learned building an open-source declarative engineering platform."
date: 2026-08-19
author: "NAEOS Foundation"
categories: ["launch", "community"]
---

On Tuesday, August 18, 2026, we launched NAEOS on Product Hunt. It was our first public launch as an open-source project, and we wanted to share the experience honestly — the numbers, the feedback, and what we're taking away from it.

## The Day in Numbers

| Metric | Result | Target |
|--------|--------|--------|
| Upvotes (24h) | 132 | 100+ |
| Comments | 28 | 20+ |
| Website visits from PH | 614 | 500+ |
| GitHub stars added | 87 | 50–150 |
| Installs / downloads | 146 | 100+ |

*Final numbers as recorded at the close of the 24-hour launch window.*

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

> "The AI angle is interesting — instruction sets for 6 tools + MCP server. Haven't seen that before."

> "Pipeline caching is a smart move. Rebuilds in seconds instead of minutes."

> "Finally a spec-driven tool that doesn't stop at scaffolding."

### The Constructive

> "The spec language has a learning curve ($ref, $include, $fn, $if). Documentation is thorough, but it's real syntax."

> "Not magic for arbitrary bespoke UI — but for APIs, services, schemas, governance, it shines."

> "Would love to see more language support beyond the big 5."

We asked for honest feedback, and we got it. The themes:

1. **Spec language complexity** — real syntax, real learning curve. We'll improve onboarding docs.
2. **AI integration** — people love the instruction set compiler. We'll add more tools.
3. **Pipeline caching** — the most praised feature. We'll extend it to more stages.

## What We Learned

### First 6 Hours Matter Most

PH ranks products based on early momentum. We hit the top 5 in the first hour and stayed there. The key: having a support list of people who upvote + comment within the first hour.

### Comments > Upvotes

A thoughtful comment ("I tried the quick start and here's what happened") carries more weight than an upvote. We encouraged every supporter to leave a comment, even a one-liner.

### The Maker's Comment Sets the Tone

Our first comment explained the problem (spec/code drift) before pitching the solution. Several commenters said this was what convinced them to try NAEOS.

### Community Channels Amplify

Discord and Slack announcements drove 30% of our early traffic. Having a community before the launch made the difference.

## What's Next

Based on launch feedback, here's what we're prioritizing:

1. **Spec language onboarding** — interactive tutorial, not just docs
2. **More AI tools** — Windsurf, Aider, Cline instruction sets
3. **Pipeline caching improvements** — cache across runs, not just within a session
4. **Better quick start** — 30-second demo that actually works on any machine

We're turning the top feedback items into GitHub issues this week. Track them in the [roadmap](/roadmap/).

## Thank You

To everyone who upvoted, commented, shared, or tried NAEOS on launch day — thank you. This project exists because spec/code drift is a real problem, and we believe the specification should be the source of truth.

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

*This post was updated with the final launch numbers shortly after the 24-hour window closed.*
