#!/usr/bin/env bash
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ ! -f VERSION ]]; then
  echo "VERSION file is missing" >&2
  exit 1
fi

version="$(tr -d '[:space:]' < VERSION)"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "VERSION must contain a plain semantic version (x.y.z); got: '$version'" >&2
  exit 1
fi

check_contains() {
  local file="$1"
  local expected="$2"
  if ! grep -Fq -- "$expected" "$file"; then
    printf 'Version consistency check failed: %s must contain:\n  %s\n' "$file" "$expected" >&2
    exit 1
  fi
}

check_contains README.md "NAEOS ${version}** is the current documented software release."
check_contains README.md "Active development · NAEOS ${version}**"
check_contains ROADMAP.md "- **Current documented release:** NAEOS ${version} (VERSION)"
check_contains WHITEPAPER.md "| **Software Release Baseline** | v${version}"
check_contains WHITEPAPER-EN.md "| **Software Release Baseline** | v${version}"
check_contains site/content/whitepaper.md "| **Software Release Baseline** | v${version}"
check_contains site/content/id/whitepaper.md "| **Versi Platform** | v${version} (versi rilis saat ini) |"

echo "Public release references are consistent with VERSION=$version."
