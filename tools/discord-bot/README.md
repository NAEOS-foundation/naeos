# NAEOS Discord Bot

Community bot + release/launch notifier for the NAEOS Discord server, plus a
one-command **server provisioner** that builds the channel/role structure from
`launch/discord-server/blueprint.md`.

**Community commands** — `/help`, `/docs`, `/status`, `/release`, `/producthunt`, `/quickstart`, `/doctor`, `/setup`, `/config`, `/ping`
**Notifications** — posts new GitHub releases to a channel, plus an optional Product Hunt launch announcement on startup.
**Provisioning** — `provision` creates all categories, channels, roles, and permission overwrites idempotently.

## Requirements

- Go 1.25+
- A Discord application + bot token
- `discordgo` (already in the repo's `go.mod`)

## Setup

1. Create an application at https://discord.com/developers/applications
   - Go to **Bot** → reset & copy the token (this is `DISCORD_TOKEN`)
   - Go to **General Information** → copy the Application ID (this is `DISCORD_APPLICATION_ID`)
2. Invite the bot to your server with the `applications.commands` and `bot` scopes:
   ```
   https://discord.com/api/oauth2/authorize?client_id=<APPLICATION_ID>&permissions=2147485696&scope=bot%20applications.commands
   ```
   (permissions `2147485696` = Send Messages + Embed Links + Read Message History + Use Slash Commands)
   For full provisioning (creating channels/roles) invite with Administrator:
   ```
   https://discord.com/api/oauth2/authorize?client_id=<APPLICATION_ID>&permissions=8&scope=bot%20applications.commands
   ```
3. Create an announcement channel and copy its channel ID (right-click channel → Copy Channel ID) for `DISCORD_ANNOUNCE_CHANNEL` (or use `/setup`).

## Run

```bash
export DISCORD_TOKEN="your-bot-token"
export DISCORD_APP_ID="your-app-id"          # or DISCORD_APPLICATION_ID / DISCORD_CLIENT_ID
export DISCORD_GUILD_ID="your-guild-id"      # optional: instant slash commands in one server (alias: DISCORD_SERVER_ID)
export DISCORD_ANNOUNCE_CHANNEL="channel-id"        # optional: release notifications target
export NAEOS_PH_LAUNCH_URL="https://www.producthunt.com/posts/naeos"
export NAEOS_ANNOUNCE_ON_STARTUP="true"             # optional: post PH launch message on boot

go run ./tools/discord-bot serve
```

## Provision the server

With the bot invited (Administrator permission) to an empty server, build the
full structure from the blueprint:

```bash
set -a; source .env; set +a
export DISCORD_GUILD_ID="your-guild-id"   # optional: bot picks the first guild otherwise
go run ./tools/discord-bot provision
```

What it creates (idempotent — existing items are left alone):
- Roles: Administrator, Moderator, Core Contributor, Contributor, Launch Champion, Bot
- Categories + channels: WELCOME, COMMUNITY, ENGINEERING, HELP, VOICE, MODERATION
- Permission overwrites: locked channels (`#announcements`, `#rules`, `#faq`), private staff channels

After provisioning, run `serve` once, then `/setup` in `#announcements`.

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DISCORD_TOKEN` | yes | — | Discord bot token |
| `DISCORD_APP_ID` | yes | — | Discord application ID (aliases: `DISCORD_APPLICATION_ID`, `DISCORD_CLIENT_ID`) |
| `DISCORD_GUILD_ID` | no | — | Guild ID for guild-scoped commands (recommended while testing; alias: `DISCORD_SERVER_ID`) |
| `DISCORD_ANNOUNCE_CHANNEL` | no | — | Channel ID for release + launch announcements |
| `NAEOS_REPO` | no | `NAEOS-foundation/naeos` | GitHub repo to watch (`owner/repo`) |
| `GITHUB_TOKEN` | no | — | GitHub token (avoids rate limits) |
| `NAEOS_POLL_INTERVAL` | no | `15m` | Release poll interval (Go duration) |
| `NAEOS_STATE_FILE` | no | `.naeos-bot-state.json` | Where the last-seen release tag is stored |
| `NAEOS_ANNOUNCE_ON_STARTUP` | no | `false` | Post the Product Hunt announcement on startup |
| `NAEOS_PH_LAUNCH_URL` | no | — | Product Hunt launch URL |
| `NAEOS_PH_TAGLINE` | no | `Specify Once. Build Anywhere.` | Tagline for PH embed |
| `NAEOS_PH_GALLERY_URL` | no | repo hero image | Image shown in the PH announcement |
| `NAEOS_PH_RELEASE_NOTE` | no | v3.0.0 blurb | Short launch note |
| `NAEOS_BIN` | no | `naeos` (in PATH) | Local `naeos` binary path for `/doctor` |

## Commands

| Command | Description |
|---------|-------------|
| `/help` | List all commands |
| `/docs` | Documentation, whitepaper, GitHub links |
| `/status` | GitHub repo status (stars, issues, license) |
| `/release` | Latest stable release details |
| `/producthunt` | Product Hunt launch info + link |
| `/quickstart` | Install + run quick start |
| `/doctor` | Run `naeos doctor` on the host machine |
| `/setup` | Set the current channel as the announcement channel (persisted) |
| `/config` | Show resolved bot config (repo, channel, poll interval, state file) |
| `/ping` | Bot latency |

## Setting the announcement channel

Two ways:

1. **Via `/setup`** — run the command in the channel that should receive announcements.
   The channel ID is persisted to `NAEOS_STATE_FILE` and survives restarts.
2. **Via env** — set `DISCORD_ANNOUNCE_CHANNEL` to a channel ID. This takes precedence
   over the `/setup` value.

`/config` shows which channel is currently active.

## Release notifications

The bot polls the GitHub releases API every `NAEOS_POLL_INTERVAL` and posts an
embed to `DISCORD_ANNOUNCE_CHANNEL` for each new stable release. The last
announced tag is persisted in `NAEOS_STATE_FILE` (default `.naeos-bot-state.json`)
so restarts don't re-announce old releases.

## Testing

```bash
go test -race -count=1 -timeout 60s ./tools/discord-bot/
```

## Notes

- Slash command registration is global by default. During testing set
  `DISCORD_GUILD_ID` so commands appear instantly; remove it once verified
  (global registration can take up to an hour to propagate).
- The Product Hunt launch message posts on startup only when
  `NAEOS_ANNOUNCE_ON_STARTUP=true`. Remove it after launch day to avoid repeats.
