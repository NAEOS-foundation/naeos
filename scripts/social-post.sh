#!/usr/bin/env bash
# Post today's (or a given date's) content-calendar entries to community channels.
#
# Channels: discord (bot), slack (chat.postMessage), linkedin (Posts API via linkdin-post.sh)
#
# Requires in .env (gitignored):
#   DISCORD_TOKEN, DISCORD_ANNOUNCE_CHANNEL
#   NAEOS_SLACK_TOKEN, NAEOS_SLACK_ANNOUNCE_CHANNEL
#   LINKEDIN_ACCESS_TOKEN, LINKEDIN_URN
#
# Usage:
#   social-post.sh --dry-run [--date YYYY-MM-DD] [--platforms discord,slack,linkedin]
#   social-post.sh [--date YYYY-MM-DD] [--platforms discord,slack,linkedin]
#
# Defaults: date = today (UTC), platforms = all three. Unknown dates exit cleanly.
#
# Features:
#   - Retry logic: retries failed posts up to 3 times with exponential backoff
#   - Post logging: logs to .social-post.log to track what was posted
#   - Duplicate prevention: skips dates that were already posted
set -euo pipefail

POST_LOG=".social-post.log"
MAX_RETRIES=3

CALENDAR="brand/marketing/content-calendar.json"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log() { printf '[social-post] %s\n' "$*" >&2; }

# --- post logging & duplicate prevention --------------------------------------
already_posted() {
  local date="$1" platform="$2"
  [ -f "$POST_LOG" ] && grep -q "^${date}|${platform}|" "$POST_LOG" 2>/dev/null
}

log_post() {
  local date="$1" platform="$2" status="$3" detail="${4:-}"
  echo "${date}|${platform}|${status}|$(date -u +%FT%TZ)|${detail}" >> "$POST_LOG"
}

# --- retry wrapper -----------------------------------------------------------
retry() {
  local attempt=1 delay=2
  while true; do
    if "$@"; then
      return 0
    fi
    if [ "$attempt" -ge "$MAX_RETRIES" ]; then
      log "FAILED after $MAX_RETRIES attempts"
      return 1
    fi
    log "retry $attempt/$MAX_RETRIES in ${delay}s..."
    sleep "$delay"
    delay=$((delay * 2))
    attempt=$((attempt + 1))
  done
}

# --- load .env ---------------------------------------------------------------
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

DATE="$(date -u +%F)"
PLATFORMS="discord,slack,linkedin"
DRY=0
while [ $# -gt 0 ]; do
  case "$1" in
    --date) DATE="$2"; shift 2 ;;
    --platforms) PLATFORMS="$2"; shift 2 ;;
    --dry-run) DRY=1; shift ;;
    --show-log) [ -f "$POST_LOG" ] && cat "$POST_LOG" || echo "No posts logged yet"; exit 0 ;;
    --reset-log) rm -f "$POST_LOG"; log "post log cleared"; exit 0 ;;
    *) log "unknown option: $1"; exit 2 ;;
  esac
done
IFS=',' read -r -a PLATFORM_ARR <<< "$PLATFORMS"

[ -f "$CALENDAR" ] || { log "calendar not found: $CALENDAR"; exit 2; }
[ -x "$SCRIPT_DIR/linkdin-post.sh" ] || { log "linkdin-post.sh not executable"; exit 2; }

# Find the calendar entry for DATE.
ENTRY="$(python3 -c '
import json, sys
try:
    cal = json.load(open(sys.argv[1]))
except Exception as e:
    sys.exit("calendar parse error: %s" % e)
for e in cal:
    if e["date"] == sys.argv[2]:
        print(json.dumps(e)); break
' "$CALENDAR" "$DATE" 2>&1)" || { log "$ENTRY"; exit 2; }

if [ -z "$ENTRY" ] || [ "$ENTRY" = "None" ]; then
  log "no entry for $DATE — nothing to post"
  exit 0
fi

if [ "$(echo "$ENTRY" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("enabled", True))')" != "True" ]; then
  log "entry for $DATE is disabled (enabled: false)"
  exit 0
fi

ENABLED="$(echo "$ENTRY" | python3 -c 'import json,sys; print(",".join(json.load(sys.stdin)["platforms"]))')"
SLUG="$(echo "$ENTRY" | python3 -c 'import json,sys; print(json.load(sys.stdin)["slug"])')"
TITLE="$(echo "$ENTRY" | python3 -c 'import json,sys; print(json.load(sys.stdin)["title"])')"
log "entry: $DATE / $SLUG — $TITLE"

post_discord() {
  : "${DISCORD_TOKEN:?DISCORD_TOKEN missing}"
  : "${DISCORD_ANNOUNCE_CHANNEL:?DISCORD_ANNOUNCE_CHANNEL missing}"
  if already_posted "$DATE" "discord"; then
    log "discord: already posted for $DATE, skipping"
    return 0
  fi
  local text
  text="$(echo "$ENTRY" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("discord",""))')"
  [ -n "$text" ] || { log "discord: empty message, skipping"; return 0; }
  if [ "$DRY" = "1" ]; then
    log "discord [dry-run] channel=$DISCORD_ANNOUNCE_CHANNEL"; printf '%s\n' "$text"
    return 0
  fi
  do_post() {
    local code body
    body="$(python3 -c 'import json,sys; print(json.dumps({"content": sys.argv[1]}))' "$text")"
    code="$(curl -s -o /tmp/discord_resp -w "%{http_code}" -X POST \
      "https://discord.com/api/v10/channels/$DISCORD_ANNOUNCE_CHANNEL/messages" \
      -H "Authorization: Bot $DISCORD_TOKEN" -H "Content-Type: application/json" -d "$body")"
    if [ "$code" = "200" ]; then
      log "discord: posted OK"
      log_post "$DATE" "discord" "ok"
      return 0
    else
      log "discord: FAILED (HTTP $code)"; cat /tmp/discord_resp >&2; return 1
    fi
  }
  retry do_post
}

post_slack() {
  : "${NAEOS_SLACK_TOKEN:?NAEOS_SLACK_TOKEN missing}"
  : "${NAEOS_SLACK_ANNOUNCE_CHANNEL:?NAEOS_SLACK_ANNOUNCE_CHANNEL missing}"
  if already_posted "$DATE" "slack"; then
    log "slack: already posted for $DATE, skipping"
    return 0
  fi
  local text
  text="$(echo "$ENTRY" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("slack",""))')"
  [ -n "$text" ] || { log "slack: empty message, skipping"; return 0; }
  if [ "$DRY" = "1" ]; then
    log "slack [dry-run] channel=$NAEOS_SLACK_ANNOUNCE_CHANNEL"; printf '%s\n' "$text"
    return 0
  fi
  do_post() {
    local resp
    resp="$(curl -s -X POST "https://slack.com/api/chat.postMessage" \
      -H "Authorization: Bearer $NAEOS_SLACK_TOKEN" -H "Content-Type: application/json" \
      -d "$(python3 -c 'import json,sys; print(json.dumps({"channel": sys.argv[1], "text": sys.argv[2]}))' "$NAEOS_SLACK_ANNOUNCE_CHANNEL" "$text")")"
    if echo "$resp" | grep -q '"ok":true'; then
      log "slack: posted OK"
      log_post "$DATE" "slack" "ok"
      return 0
    else
      log "slack: FAILED"; echo "$resp" >&2; return 1
    fi
  }
  retry do_post
}

post_linkedin() {
  : "${LINKEDIN_ACCESS_TOKEN:?LINKEDIN_ACCESS_TOKEN missing}"
  : "${LINKEDIN_URN:?LINKEDIN_URN missing}"
  if already_posted "$DATE" "linkedin"; then
    log "linkedin: already posted for $DATE, skipping"
    return 0
  fi
  local text
  text="$(echo "$ENTRY" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("linkedin",""))')"
  [ -n "$text" ] || { log "linkedin: empty message, skipping"; return 0; }
  if [ "$DRY" = "1" ]; then
    log "linkedin [dry-run] author=$LINKEDIN_URN"; printf '%s\n' "$text"
    return 0
  fi
  log "linkedin: posting via linkdin-post.sh"
  if "$SCRIPT_DIR/linkdin-post.sh" "$text"; then
    log_post "$DATE" "linkedin" "ok"
    return 0
  else
    return 1
  fi
}

FAILED=0
for p in "${PLATFORM_ARR[@]}"; do
  case "$p" in
    discord)  [[ ",$ENABLED," == *",discord,"*  ]] && post_discord  || log "discord: not in entry platforms"; ;;
    slack)    [[ ",$ENABLED," == *",slack,"*  ]]   && post_slack    || log "slack: not in entry platforms"; ;;
    linkedin) [[ ",$ENABLED," == *",linkedin,"* ]] && post_linkedin || log "linkedin: not in entry platforms"; ;;
    *) log "unknown platform: $p"; FAILED=1; ;;
  esac
done
exit "$FAILED"