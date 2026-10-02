#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

record="${1:-}"
if [[ -z "$record" || ! -f "$record" ]]; then
  echo "Legal evidence record is required" >&2
  exit 1
fi

command -v jq >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }

decision="$(jq -r '.decision // empty' "$record")"
case "$decision" in
  PASS)
    ;;
  HOLD)
    echo "Release legal gate: HOLD — release publication is blocked." >&2
    exit 1
    ;;
  *)
    echo "Release legal gate: invalid or missing decision: ${decision:-<empty>}" >&2
    exit 1
    ;;
esac

if jq -e '
  [.dco_status,.dependency_license_status,.attribution_status,.security_status,.provenance_status,.trademark_status]
  | any(. == "HOLD" or . == null)
' "$record" >/dev/null; then
  echo "Release legal gate: required legal evidence contains HOLD or missing status" >&2
  exit 1
fi

echo "Release legal gate: PASS — release publication may continue."
