#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# Emit a Graphviz DOT graph of direct imports between packages in this Go module.
# External dependencies are intentionally omitted. The output is generated on
# demand and should not be committed as a repository snapshot.

if [[ $# -gt 1 ]]; then
  echo "usage: $0 [output.dot]" >&2
  exit 2
fi

if ! command -v go >/dev/null 2>&1; then
  echo "error: Go toolchain is required" >&2
  exit 1
fi

module="$(awk '$1 == "module" { print $2; exit }' go.mod)"
if [[ -z "$module" ]]; then
  echo "error: module path not found in go.mod (run from repository root)" >&2
  exit 1
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
go list -f '{{.ImportPath}}|{{join .Imports ","}}' ./... > "$tmp"

render_graph() {
  awk -F'|' -v prefix="$module/" '
    BEGIN {
      print "digraph GoPackages {"
      print "  rankdir=LR;"
      print "  graph [label=\"NAEOS direct Go package imports (in-module only)\", labelloc=t];"
    }
    {
      importer = $1
      if (index(importer, prefix) != 1) next
      importer = substr(importer, length(prefix) + 1)
      printf "  \"%s\";\n", importer

      count = split($2, imports, ",")
      for (i = 1; i <= count; i++) {
        imported = imports[i]
        if (index(imported, prefix) != 1) continue
        imported = substr(imported, length(prefix) + 1)
        printf "  \"%s\" -> \"%s\";\n", importer, imported
      }
    }
    END { print "}" }
  ' "$tmp"
}

if [[ $# -eq 1 ]]; then
  render_graph > "$1"
else
  render_graph
fi
