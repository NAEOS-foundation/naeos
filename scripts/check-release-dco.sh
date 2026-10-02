#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

base="${1:-}"
head="${2:-HEAD}"

if [[ -z "$base" ]]; then
  echo "usage: $0 BASE [HEAD]" >&2
  exit 2
fi

git rev-parse --verify "$base" >/dev/null 2>&1 || {
  echo "base commit not found: $base" >&2
  exit 1
}

git rev-parse --verify "$head" >/dev/null 2>&1 || {
  echo "head commit not found: $head" >&2
  exit 1
}

range="$(git merge-base "$base" "$head")..$head"
commits="$(git rev-list --no-merges --first-parent "$range")"
bad=0

for hash in $commits; do
  committer_email="$(git show -s --format=%ce "$hash")"
  if [[ "$committer_email" == "noreply@github.com" ]]; then
    echo "Skipping GitHub-generated commit $hash"
    continue
  fi

  if ! git log -1 --format=%B "$hash" | grep -q '^Signed-off-by:'; then
    echo "::error::commit $hash is missing a Signed-off-by trailer"
    bad=1
  fi
done

if [[ "$bad" -ne 0 ]]; then
  echo "Release DCO verification: HOLD"
  exit 1
fi

echo "Release DCO verification: PASS"
