#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

: "${BASE_SHA:?BASE_SHA is required}"
: "${HEAD_SHA:?HEAD_SHA is required}"

git cat-file -e "${BASE_SHA}^{commit}"
git cat-file -e "${HEAD_SHA}^{commit}"

if git merge-base --is-ancestor "$BASE_SHA" "$HEAD_SHA"; then
  echo "Merge-base freshness: PASS"
  echo "Base commit $BASE_SHA is an ancestor of PR head $HEAD_SHA."
else
  echo "Merge-base freshness: BLOCKED" >&2
  echo "PR head $HEAD_SHA does not contain current base commit $BASE_SHA." >&2
  echo "Update the branch from the current base before merging." >&2
  exit 1
fi
