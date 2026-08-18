#!/usr/bin/env bash
# Start the NAEOS Discord bot at 14:00 WIB on launch day (Aug 18, 2026).
# Usage: ./launch/discord-bot-launch.sh

set -euo pipefail

LAUNCH_HOUR=14
LAUNCH_MIN=0
LOGFILE="/tmp/discord-bot-launch.log"
BOT_BIN="/tmp/discord-bot"

cd "$(dirname "$0")/.."

# Build if binary is missing or stale
if [[ ! -x "$BOT_BIN" ]] || [[ "tools/discord-bot/" -nt "$BOT_BIN" ]]; then
  echo "[$(date '+%F %T')] Building discord-bot..." | tee -a "$LOGFILE"
  go build -o "$BOT_BIN" ./tools/discord-bot/
fi

# Load env
set -a
# shellcheck disable=SC1091
source .env
set +a

# Wait until launch hour
while true; do
  HOUR=$(date '+%-H')
  MIN=$(date '+%-M')
  if (( HOUR > LAUNCH_HOUR || (HOUR == LAUNCH_HOUR && MIN >= LAUNCH_MIN) )); then
    break
  fi
  REMAINING=$(( (LAUNCH_HOUR - HOUR) * 60 + (LAUNCH_MIN - MIN) ))
  echo "[$(date '+%F %T')] Waiting ${REMAINING}m until ${LAUNCH_HOUR}:$(printf '%02d' "$LAUNCH_MIN") WIB..." | tee -a "$LOGFILE"
  sleep 60
done

echo "[$(date '+%F %T')] Starting NAEOS Discord bot..." | tee -a "$LOGFILE"
exec "$BOT_BIN" serve >> "$LOGFILE" 2>&1
