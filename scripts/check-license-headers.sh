#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

# Verifies every tracked first-party source file carries the NAEOS Apache-2.0
# SPDX header. Covers Go, TypeScript/JavaScript, shell scripts, and Python.
set -euo pipefail

cd "$(dirname "$0")/.."

mapfile -t files < <(git ls-files '*.go' '*.ts' '*.tsx' '*.js' '*.mjs' '*.cjs' '*.sh' '*.py')
if [ "${#files[@]}" -eq 0 ]; then
  echo "no tracked source files to check"
  exit 0
fi

missing=0
for f in "${files[@]}"; do
  if ! head -n 8 "$f" | grep -q 'SPDX-License-Identifier: Apache-2.0'; then
    echo "missing header: $f"
    missing=$((missing + 1))
  fi
done

if [ "$missing" -gt 0 ]; then
  echo
  echo "error: $missing tracked source file(s) missing the Apache-2.0 SPDX header."
  echo "Add to each file (using // for Go/JS/TS, # for shell/Python — after any"
  echo "shebang or 'use client'/'use server' directive):"
  echo "  Copyright 2024-2026 NAEOS Foundation"
  echo "  SPDX-License-Identifier: Apache-2.0"
  exit 1
fi

echo "ok: all ${#files[@]} tracked first-party source files carry the Apache-2.0 SPDX header"
