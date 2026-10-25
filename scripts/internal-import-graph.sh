#!/usr/bin/env bash
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

module_path="github.com/NAEOS-foundation/naeos"

# Emit deterministic TSV edges: importing_package<TAB>imported_package.
# Keep only first-party edges so the output can be reviewed or piped into graph tools.
go list -f '{{ $from := .ImportPath }}{{ range .Imports }}{{ printf "%s\t%s\n" $from . }}{{ end }}' ./... |
  awk -F '\t' -v prefix="$module_path/" '
    index($1, prefix) == 1 && index($2, prefix) == 1 { print }
  ' |
  LC_ALL=C sort -u
