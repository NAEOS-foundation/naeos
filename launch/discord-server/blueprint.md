# NAEOS Community Discord — Server Blueprint

A professional Discord server for the NAEOS community. Follow this blueprint top-to-bottom to build the server in ~30 minutes, then run the launch playbook when the Product Hunt launch goes live.

## Contents

1. [Server identity](#1-server-identity)
2. [Channel structure](#2-channel-structure)
3. [Roles & permissions](#3-roles--permissions)
4. [Onboarding flow](#4-onboarding-flow)
5. [Moderation](#5-moderation)
6. [Bot integration](#6-bot-integration)
7. [Launch-day playbook](#7-launch-day-playbook)
8. [Recommended settings](#8-recommended-settings)

---

## 1. Server identity

| Item | Value |
|------|-------|
| **Name** | NAEOS Community |
| **Icon** | `assets/server-icon.png` (generated from `brand/logo.svg`) |
| **Banner** | `assets/server-banner.png` (requires a Level 2 boost) |
| **Invite splash** | `assets/invite-splash.png` |
| **Description** | Declarative engineering community. Specify once, build anywhere. |
| **Tags** | Developer Tools, Open Source, Go, AI, Engineering |

Create the server, then under **Server Settings → Overview** set the name and
description. Upload the icon and (once boosted) the banner. Set the invite splash
under **Server Settings → Invites**.

---

## 2. Channel structure

Create categories in this order, then channels inside them. Starred channels
(*) receive bot/webhook posts.

### 🟢 WELCOME
| Channel | Purpose |
|---------|---------|
| `#welcome` * | First channel people see. Post the welcome message (see templates) |
| `#rules` * | Read-only. The rules (see templates) |
| `#announcements` * | NAEOS bot posts releases + launch updates here (`/setup` in this channel) |
| `#product-hunt` | Product Hunt launch discussion + upvote link |

### 💬 COMMUNITY
| Channel | Purpose |
|---------|---------|
| `#general` | Main discussion |
| `#introductions` | New members say hi |
| `#showcase` | Share projects built with NAEOS |
| `#off-topic` | Casual chat |

### ⚙️ ENGINEERING (topics with forums where noted)
| Channel | Purpose |
|---------|---------|
| `#spec-language` | Spec Language v2 questions (`$ref`, `$fn`, `$if`, migrations) |
| `#code-generation` | Multi-language generation (Go/TS/Py/Java/Rust) |
| `#ai-integration` | AI compiler, MCP server, Copilot/Claude/Cursor adapters |
| `#go` | Go internals, kernel, LSP server, plugin SDK |
| `#plugins` | WASM plugin SDK, official example plugins |
| `#architecture` | NEIR model, patterns, governance & policy |
| `#roadmap` | Roadmap and RFC/ADR discussion |

### 🆘 HELP
| Channel | Purpose |
|---------|---------|
| `#help-requests` | Open a thread for your issue (use forum or threads) |
| `#troubleshooting` | Known issues, setup problems |
| `#faq` | Read-only FAQ (copy from `site/content/faq.md`) |

### 🔊 VOICE
| Channel | Purpose |
|---------|---------|
| `#general-vc` | General voice chat |
| `#dev-vc` | Code/streaming voice chat |
| `#launch-party` | Temporary voice channel for launch day |

### 🛡️ MODERATION (private, visible only to staff roles)
| Channel | Purpose |
|---------|---------|
| `#mod-chat` | Staff coordination |
| `#mod-log` | Bots log moderation actions here |

### Launch-only channels (delete after launch week)
- `#launch-day-announcements` — countdown + go-live post
- `#launch-upvotes` — the PH link + how to support
- `#launch-party` (voice)

---

## 3. Roles & permissions

Create roles in this order (highest first). Assign colors that match the brand.

| Role | Color | Purpose |
|------|-------|---------|
| `@Administrator` | `#9333ea` | Full server control (keep to 1–2 people) |
| `@Moderator` | `#7c4dff` | Enforce rules, manage threads |
| `@Core Contributor` | `#3a7cf8` | Trusted maintainers; ping for RFCs/ADR |
| `@Contributor` | `#08d6ff` | Anyone with merged code/docs |
| `@Launch Champion` | `#ffaa00` | Temporary: launch-day helpers + top upvoters |
| `@Member` | `#9999b0` | Default role for everyone |
| `@Bot` | `#60a5fa` | Assigned to bots |

### Permission templates

**`@Member`** (default):
- Text: View Channels, Read Message History, Send Messages, Add Reactions, Use Slash Commands, Create Public Threads
- Denied: Manage Messages, Mention Everyone, Use External Emoji
- Voice: Connect, Speak, Use Voice Activity

**`@Contributor`** = `@Member` + can create `#showcase` threads + `@Launch Champion`-adjacent perms during launch.

**`@Moderator`** = `@Member` + Manage Messages, Manage Channels, Kick Members, Manage Threads, Move Members (voice).

**Channel overrides** — critical ones:
- `#announcements`: `@Member` → Send Messages **denied** (only bots/staff post)
- `#rules`, `#faq`: `@Member` → Send Messages **denied**
- `#mod-chat`, `#mod-log`: only `@Administrator` + `@Moderator` see/send
- `#launch-party` voice: everyone, but only `@Launch Champion` + staff can mute/undeafen

### Server-wide settings
- Verification level: **Medium** (email verified) — good spam balance for a dev community
- Explicit content filter: **Scan media sent by members without a role** (or higher)
- Anti-spam: keep default + a moderation bot

---

## 4. Onboarding flow

1. New member joins → lands in `#welcome` (set as the default landing channel in
   Server Settings → Onboarding).
2. `#welcome` message explains what NAEOS is + points to `#rules`, `#introductions`,
   `#help-requests`.
3. Ask members to introduce themselves in `#introductions` (optional but warm).
4. Post the quick start in `#welcome` so people can try NAEOS in 30 seconds.
5. Use **Server Guide** (Onboarding) to pin: `#announcements`, `#help-requests`,
   `#showcase`, `#product-hunt`.

Welcome message, rules, and community guidelines templates live in `templates.md`.

---

## 5. Moderation

- **Moderation policy:** Be excellent. Disagreements about engineering are welcome;
  personal attacks are not. Mods act on reports + message logs.
- **Rules** (paste from `templates.md` into `#rules`).
- **Tools:**
  - Use a moderation/logging bot (e.g. **Carl-bot** or **MEE6**) writing to `#mod-log`.
  - Keep the NAEOS bot in `#announcements` for release + launch posts.
- **Mod actions:** warn → mute → kick → ban, escalating. Log all actions in `#mod-log`.
- **Link policy:** allow GitHub links freely; external self-promo requires a note in `#showcase`.

---

## 6. Bot integration

| Bot | Channel | Purpose |
|-----|---------|---------|
| **NAEOS bot** (`tools/discord-bot`) | `#announcements` | Releases, Product Hunt launch, `/status`, `/release`, `/docs`, `/doctor`, `/setup`, `/config` |
| **Carl-bot** (or MEE6) | `#mod-log` | Moderation, logging, auto-roles, timers |
| **GitHub webhook** (server → repo) | `#announcements` or `#dev-feed` | Push/release/issue events — use Discord's native "Add to Server → GitHub" integration |

### NAEOS bot setup
```bash
set -a; source .env; set +a

# Optional: build the whole channel/role structure in one shot
# (bot must be invited with Administrator permission):
go run ./tools/discord-bot provision

# Then run the bot and point announcements at #announcements:
go run ./tools/discord-bot serve
```
1. Invite the bot (see `tools/discord-bot/README.md`).
2. In `#announcements`, run `/setup` — the channel is now the release/launch channel.
3. `/config` to confirm.
4. Add a GitHub webhook in `#announcements` for release events only (keep the noise down).

---

## 7. Launch-day playbook

T-2 days:
- Create `#launch-upvotes` and `#launch-party`.
- Ping `@Launch Champion` with the PH link and support instructions.
- Set `NAEOS_ANNOUNCE_ON_STARTUP=true` on the bot + restart it so the PH
  announcement posts to `#announcements`.

T-0 (launch hour):
- Post the PH link in `#product-hunt` + `#announcements`.
- Post the go-live announcement (template in `templates.md`).
- Encourage members to upvote + comment (never brigade — organically).

D+1 → D+7:
- Post a thank-you recap in `#announcements` (numbers + feedback themes).
- Promote the top 3 feedback items as GitHub issues in `#roadmap`.
- Move `#launch-upvotes` to archived; delete `#launch-party`.

---

## 8. Recommended settings

| Setting | Value |
|---------|-------|
| Verification level | Medium |
| Default notification settings | @mentions |
| Explicit content filter | Scan media from non-roled members |
| Slowmode in `#general` | Off (or 5s during launch day) |
| Thread archive | 1 week |
| Community | Enable (gets Rules Screening + Onboarding) |
| Rules screening | On |
| Default member permissions | None (rely on roles) |

---

## File map

| File | Purpose |
|------|---------|
| `blueprint.md` | This guide |
| `templates.md` | Paste-ready welcome message, rules, community guidelines, launch announcements |
| `assets/server-icon.png` | 512×512 server icon |
| `assets/server-banner.png` | 960×540 banner (needs boost) |
| `assets/invite-splash.png` | 1920×1080 invite splash |
| `assets/generate.js` | Regenerates the assets (requires `sharp`) |
