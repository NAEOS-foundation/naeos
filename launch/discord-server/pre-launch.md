# NAEOS Pre-Launch Content — Attract & Grow the Community (D-2 → D-0)

Paste-ready content to grow the Discord community and rally members **before** the
Product Hunt launch (Tue, 18 Aug 2026). Goal: more members on the server = bigger
blast radius on launch day.

**3 attraction levers used here:**
1. **Exclusivity** — "join before launch" gives early access to v3.1.0 previews + roadmap voting
2. **Value** — each teaser teaches something real (a spec trick, a pipeline insight), not just hype
3. **Recognition** — `@Launch Champion` role = a real reason to join + help now

All times in **WIB (UTC+7)**. PT = WIB − 14h. Replace `[PH LINK]` once the post exists.

---

## 1. Social teasers — funnel people to the Discord

### D-2 (Sun) — Teaser #1: "building something on Tuesday"

**X post:**
> Tuesday we're shipping something we've been building for a while.
>
> NAEOS — a platform where you describe your system once, and it generates
> validated multi-language code + AI instruction sets for 6 tools.
>
> Spec-driven engineering, but real. 🚀
> Open source, Apache 2.0.
>
> Join the community for a first look before launch:
> https://discord.gg/WnUWmm7XMv

**LinkedIn post:**
> On Tuesday we're launching NAEOS — our open-source declarative engineering platform.
>
> One spec.yaml → validated Go/TS/Python/Java/Rust code, governance checks, docs,
> and AI context for Copilot, Claude Code, Cursor, and more.
>
> Spec stays the source of truth for the whole lifecycle. No more spec/code drift.
>
> If that resonates, come say hi before the launch:
> https://discord.gg/WnUWmm7XMv
>
> More details Tuesday. 🚀

### D-1 (Mon) — Teaser #2: v3.1.0 sneak peek

**X post:**
> Tomorrow: NAEOS v3.1.0 on Product Hunt.
>
> Preview of what ships:
> - Pipeline caching → re-runs are seconds, not minutes
> - Run-level profiling (`--profile`, `--pprof`)
> - Architecture patterns: monolithic / microservices / serverless
> - WASM plugin hardening
>
> Community gets the full walkthrough tonight:
> https://discord.gg/WnUWmm7XMv

**LinkedIn post:**
> Tomorrow NAEOS launches on Product Hunt 🚀
>
> v3.1.0 highlights:
> - Pipeline caching on `naeos run` — rebuild only what changed
> - Profiling built in (`--profile`, `--pprof`) so you can see where the pipeline spends time
> - First-class architecture patterns: monolithic, microservices, serverless
> - Hardened WASM plugin SDK
>
> Tonight the community gets the full feature walkthrough. Join:
> https://discord.gg/WnUWmm7XMv

---

## 2. Discord announcements — paste into `#announcements`

### D-2 (Sun) — Welcome + mission post (also pin)

> ## NAEOS is launching Tuesday 🚀
>
> On **Tuesday, 18 Aug** we launch NAEOS on Product Hunt.
>
> If you're new here: NAEOS is a declarative engineering platform. You describe
> your system **once** in YAML/JSON — it builds an internal engineering model
> (NEIR) and generates validated code in Go, TypeScript, Python, Java, Rust,
> plus AI instruction sets for 6 tools and an MCP server.
>
> **What's happening this week:**
> - 👀 **Tonight/Monday** — community walkthrough of v3.1.0 (caching, profiling, architecture patterns)
> - 🗳️ **Roadmap voting** — tell us what ships next (see `#roadmap` thread)
> - 🏆 **Become a Launch Champion** — help us launch, get the role (see below)
> - 🎉 **Launch day** — watch party + support thread in `#product-hunt`
>
> The server is quiet because we're heads-down shipping. Ask anything in
> `#general` — or try NAEOS right now:
> ```bash
> curl -fsSL https://naeos.dev/install.sh | sh
> naeos create
> cd my-app && naeos run --input-file spec.yaml
> ```
> Links: https://naeos.dev · https://docs.naeos.dev · https://github.com/NAEOS-foundation/naeos

### D-1 (Mon) — v3.1.0 walkthrough announcement

> ## Tonight: NAEOS v3.1.0 — community walkthrough 🎁
>
> Before tomorrow's launch, here's what the community gets first:
>
> **🚀 Pipeline caching** — `naeos run --cache-dir .naeos-cache` skips stages
> that haven't changed. Incremental builds in seconds.
>
> **📊 Run-level profiling** — `naeos run --profile --profile-out profile.json`
> shows exactly where the pipeline spends time. `--pprof` for the Go-nerd view.
>
> **🏗️ Architecture patterns** — set `architecture.pattern` in your spec
> (monolithic, microservices, or serverless) and the NEIR model + codegen
> adapts to it.
>
> **🧩 WASM plugins** — hardened plugin SDK + official examples.
>
> Try it, then vote on what we should build next in `#roadmap`. The top request
> becomes the first post-launch issue.
>
> Launch link goes live tomorrow → `[PH LINK]`

### D-0 (Tue morning, ~4–6h before go-live) — "today" post

> ## Today's the day 🎉
>
> NAEOS launches on Product Hunt at **14:00 WIB**.
>
> When it's live: upvote + comment here → `[PH LINK]` (a comment helps more than
> an upvote — even "what does this do?").
>
> `#launch-party` voice opens at launch. See you there.

---

## 3. Launch Champion recruitment — `#announcements` + `#build-in-public`

> ## Become a @Launch Champion 🏆
>
> We launch Tuesday and want 5–10 Launch Champions to help us make it loud.
>
> **What you'd do (pick any):**
> - Upvote + comment on the PH post within the first hour
> - Share NAEOS with one dev community you're part of
> - Test the quick start and report friction
> - Hang in `#launch-party` and help answer questions
>
> **What you get:**
> - 🏅 Exclusive `@Launch Champion` role (permanent, not just launch week)
> - 👀 Early access to roadmap + post-launch features
> - 🎁 A shout-out in the launch recap
>
> Reply to this message with **"I'm in"** + which part you'll help with.
> Everyone who participates keeps the role. No brigading, no spam — just an
> honest first-day boost from people who actually use NAEOS.

---

## 4. Sneak-peek mini-series (daily, `#announcements` or `#build-in-public`)

Post one per day as a short thread. Keeps the server active pre-launch without
launch-day noise.

- **Day 1 (D-2):** "One spec, six languages" — show a tiny spec.yaml + the
  generated Go AND Python files side by side. The "wait, that's it?" moment.
- **Day 2 (D-1):** "Your spec never lies" — `naeos validate` catches schema
  errors before codegen; show a broken spec → error → fix → green.
- **Day 3 (D-0):** "Your AI knows your architecture" — the AI compiler emits
  instruction sets for 6 tools + MCP server. Agents work from the model, not a guess.

---

## 5. Engagement hooks — keep members past the join

- **`#roadmap`** — pinned thread: "Vote on the next feature." Pre-seed with 3–5
  real candidates (from ROADMAP.md) so the first vote isn't empty.
- **`#introductions`** — pin the prompt from `templates.md` §6; the maker
  (Bayu) replies to every intro within the hour for the first week.
- **`#showcase`** — pin §7 prompt; highlight one build per day until launch.
- **`#help-requests`** — maker answers within 24h pre-launch (builds goodwill +
  thread depth that makes the server feel alive to new joiners).

---

## 6. Sequencing & timing (WIB)

| When | Channel | Action |
|------|---------|--------|
| D-2 (Sun) | X, LinkedIn | Teaser #1 |
| D-2 (Sun) | `#announcements` | Welcome + mission post (pin) |
| D-2 (Sun) | `#announcements`/`#build-in-public` | Sneak-peek #1 |
| D-2 (Sun) | `#announcements` | Launch Champion call (pin 24h) |
| D-1 (Mon) | X, LinkedIn | Teaser #2 (v3.1.0 preview) |
| D-1 (Mon) | `#announcements` | v3.1.0 walkthrough post |
| D-1 (Mon) | `#roadmap` | Open roadmap voting |
| D-1 (Mon) | `#announcements`/`#build-in-public` | Sneak-peek #2 |
| D-0 (Tue ~10:00) | `#announcements` | "Today's the day" post |
| D-0 (Tue 14:00) | `#announcements`, `#product-hunt`, `#launch-party` | Go-live (`templates.md` §4) |
