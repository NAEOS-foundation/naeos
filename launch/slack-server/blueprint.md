# NAEOS Community Slack — Workspace Blueprint

A professional Slack workspace mirroring the Discord community. Create the
workspace in ~20 minutes (browser-based), then copy the channel structure and
content below. Launch-adjacent copy lives in `launch/producthunt/` and the
Discord `pre-launch.md` — reuse the same messaging.

## Before you start

- Create the workspace at **https://slack.com/create** (free plan; your email).
- Workspace name: **NAEOS Community**
- Workspace URL (choose): `naeos-community`
- Icon: `brand/logo-mark.svg` → `launch/slack-server/assets/workspace-icon.png` (if generated)

## Automated setup (`tools/slack-setup`)

Once the workspace exists, provision it automatically:

1. Create a Slack app (`https://api.slack.com/apps` → From scratch → `NAEOS Bot`).
2. Add OAuth scopes: `channels:manage`, `channels:read`, `channels:join`, `chat:write`, `users:read`; Install to workspace.
3. Set `NAEOS_SLACK_TOKEN=<xoxb-...>` in `~/.env` (gitignored).
4. Run:
   ```bash
   set -a; source ~/.env; set +a
   go run ./tools/slack-setup -channels   # creates the blueprint channels
   go run ./tools/slack-setup -messages   # posts welcome/rules/pre-launch to #announcements
   ```
   Idempotent — safe to re-run.

---

## 1. Channel structure

Create channels in this order. `#announcements` and `#launch-upvotes` are the
only ones members must see first.

### Public channels
| Channel | Purpose |
|---------|---------|
| `#announcements` | Releases, Product Hunt launch, pre-launch posts (read-mostly) |
| `#general` | Main discussion |
| `#introductions` | New members say hi |
| `#showcase` | Share projects built with NAEOS |
| `#software-architecture` | Design discussions, patterns, NEIR model |
| `#ai-engineering` | AI-assisted engineering topics |
| `#ai-agent` | AI agent workflows, MCP server, instruction sets |
| `#spec-language` | Spec Language v2 (`$ref`, `$fn`, `$if`, migrations) |
| `#code-generation` | Multi-language generation (Go/TS/Py/Java/Rust) |
| `#plugins` | WASM plugin SDK, official example plugins |
| `#help` | Questions + answers |
| `#off-topic` | Casual chat |

### Launch-only channels (archive after launch week)
| Channel | Purpose |
|---------|---------|
| `#launch-upvotes` | PH link + how to support on launch day |
| `#launch-day` | Countdown, go-live watch, post-launch recap |

---

## 2. Roles (User Groups)

| Group | Purpose |
|-------|---------|
| `@admins` | Workspace admins (1–2 people) |
| `@moderators` | Enforce rules, manage members |
| `@contributors` | Anyone with merged code/docs (auto-invited) |
| `@launch-champions` | Launch-day helpers + top upvoters (temporary) |

---

## 3. Onboarding

1. **Welcome** — post the welcome message (below) in `#announcements` and pin it.
2. **Invite link** — Settings → Invite people → create an invite link; put it in
   `support-list.md` next to the Discord invite.
3. **Profile prompt** — post the `#introductions` prompt (below) in `#introductions`.
4. **Quick start** — same message as Discord: `curl -fsSL https://naeos.dev/install.sh | sh`.

---

## 4. Paste-ready messages

### Welcome — `#announcements` (pin)

> # Welcome to the NAEOS Community 👋
>
> **NAEOS** is a declarative engineering platform: describe your system once,
> and it builds, validates, and evolves real software — for humans and AI.
>
> **Specify once. Build anywhere.**
>
> ## Start here
> 1. Say hi in `#introductions`.
> 2. Try NAEOS in 30 seconds:
>    ```bash
>    curl -fsSL https://naeos.dev/install.sh | sh
>    naeos create          # interactive wizard
>    cd my-app
>    naeos run --input-file spec.yaml
>    ```
> 3. Get help in `#help`, share what you build in `#showcase`.
>
> ## Useful links
> - Docs: https://docs.naeos.dev
> - GitHub: https://github.com/NAEOS-foundation/naeos
> - Website: https://naeos.dev
> - Whitepaper: https://naeos.dev/whitepaper

### Rules — `#announcements` (pin below welcome)

> # Community Rules
> 1. **Be excellent to each other.** Disagree with ideas, not people.
> 2. **Stay on topic.** Engineering topics in the engineering channels; casual in `#off-topic`.
> 3. **Search before asking.** Check docs first, then ask in `#help` with your spec/CLI output.
> 4. **No spam.** Self-promotion belongs in `#showcase` and must add value.
>    No vote-brigading or paid promotion, ever.
> 5. **Respect the license.** NAEOS is Apache 2.0 — keep attribution where required.

### `#introductions` prompt

> New here? Tell us:
> - What you build / work on
> - What brought you to NAEOS
> - What you'd like to see next
>
> No need to be formal — a one-liner is perfect.

### Pre-launch (D-2 → D-0)

Reuse the exact messages from `launch/discord-server/pre-launch.md` — copy them
into `#announcements` (and `#launch-day`) on the same D-2 / D-1 / D-0 schedule.

---

## 5. Launch-day quick reference

| When (WIB) | Channel | Action |
|-----------|---------|--------|
| Tue 14:00 | `#announcements`, `#launch-upvotes`, `#launch-day` | Post go-live message (`pre-launch.md` §golive) with PH link |
| Tue 14:05 | `#launch-day` | First comment + thank-you thread |
| Tue 18:00+ | `#launch-day` | Peak traffic — answer questions, pin docs link |

---

## 6. File map

| File | Purpose |
|------|---------|
| `blueprint.md` | This guide |
| `assets/workspace-icon.png` | 512×512 icon (optional, reuse `brand/`) |
