#!/usr/bin/env bash
# Post a message to LinkedIn (member feed) via the LinkedIn Posts API.
#
# Requires in .env (gitignored):
#   LINKEDIN_ACCESS_TOKEN=<token with w_member_social scope>
#   LINKEDIN_URN=urn:li:person:<sub>   (optional; falls back to LINKEDIN_SUB)
#   LINKEDIN_SUB=<person sub from /v2/userinfo>
#
# Usage:
#   linkdin-post.sh "message text"            # literal text
#   linkdin-post.sh --file path/to/post.md    # read message from a file
#   linkdin-post.sh --dry-run "message"       # validate message only, no post
#
# Uses LinkedIn-Version 202601 (the active version that returns HTTP 201).
# Posts are PUBLISHED with MAIN_FEED distribution and PUBLIC visibility.
set -euo pipefail

log() { printf '[linkdin-post] %s\n' "$*" >&2; }

# --- load .env (do not echo secrets) ------------------------------------------
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

: "${LINKEDIN_ACCESS_TOKEN:?LINKEDIN_ACCESS_TOKEN missing in .env}"
DRY=0
FILE=""
while [ $# -gt 0 ]; do
  case "$1" in
    --file) FILE="$2"; shift 2 ;;
    --dry-run) DRY=1; shift ;;
    *) break ;;
  esac
done

if [ -n "$FILE" ]; then
  MSG="$(cat "$FILE")"
elif [ $# -gt 0 ]; then
  MSG="$*"
else
  echo "usage: linkdin-post.sh [--dry-run] [--file path] \"message\"" >&2
  exit 2
fi

# Resolve author URN.
AUTHOR="${LINKEDIN_URN:-}"
if [ -z "$AUTHOR" ]; then
  AUTHOR="${LINKEDIN_SUB:+urn:li:person:${LINKEDIN_SUB}}"
fi
: "${AUTHOR:?set LINKEDIN_URN or LINKEDIN_SUB in .env}"

if [ "$DRY" = "1" ]; then
  log "dry-run: would post to $AUTHOR"
  printf '%s\n' "$MSG"
  exit 0
fi

PAYLOAD="$(printf '%s' "$MSG" | python3 -c '
import json, sys
text = sys.stdin.read()
print(json.dumps({
  "author": "'"$AUTHOR"'",
  "commentary": text,
  "visibility": "PUBLIC",
  "distribution": {
    "feedDistribution": "MAIN_FEED",
    "targetEntities": [],
    "thirdPartyDistributionChannels": [],
  },
  "lifecycleState": "PUBLISHED",
  "isReshareDisabledByAuthor": False,
}))')"

RESP_HEADERS="$(mktemp)"
RESP_BODY="$(mktemp)"
HTTP_CODE="$(curl -s -o "$RESP_BODY" -D "$RESP_HEADERS" -w "%{http_code}" \
  -X POST "https://api.linkedin.com/rest/posts" \
  -H "Authorization: Bearer $LINKEDIN_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "LinkedIn-Version: 202601" \
  -H "X-Restli-Method: CREATE" \
  -d "$PAYLOAD")"

POST_ID="$(grep -i '^x-restli-id:' "$RESP_HEADERS" | tr -d '\r' | awk '{print $2}')"
rm -f "$RESP_HEADERS"

if [ "$HTTP_CODE" = "201" ]; then
  log "posted OK: $AUTHOR${POST_ID:+ (id: $POST_ID)}"
else
  log "FAILED (HTTP $HTTP_CODE)"
  cat "$RESP_BODY" >&2
  rm -f "$RESP_BODY"
  exit 1
fi
rm -f "$RESP_BODY"