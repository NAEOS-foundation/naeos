#!/usr/bin/env bash
# Verifies every tracked Go file carries the NAEOS Apache-2.0 SPDX header.
set -euo pipefail

cd "$(dirname "$0")/.."

mapfile -t files < <(git ls-files '*.go')
if [ "${#files[@]}" -eq 0 ]; then
  echo "no tracked Go files to check"
  exit 0
fi

missing=0
for f in "${files[@]}"; do
  if ! head -n 5 "$f" | grep -q '^// SPDX-License-Identifier: Apache-2.0'; then
    echo "missing header: $f"
    missing=$((missing + 1))
  fi
done

if [ "$missing" -gt 0 ]; then
  echo
  echo "error: $missing tracked Go file(s) missing the Apache-2.0 SPDX header."
  echo "Add to each file:"
  echo "  // Copyright 2024-2026 NAEOS Foundation"
  echo "  // SPDX-License-Identifier: Apache-2.0"
  exit 1
fi

echo "ok: all ${#files[@]} tracked Go files carry the Apache-2.0 SPDX header"