#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# Emit a Graphviz DOT graph of direct imports between packages in this Go module.
# External dependencies are intentionally omitted. The output is generated on
# demand and should not be committed as a repository snapshot.

if [[ $# -gt 1 ]] || [[ $# -eq 1 && "$1" != "--stats" ]]; then
  echo "usage: $0 [output.dot|--stats]" >&2
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
stats_tmp="$(mktemp)"
sorted_tmp="$(mktemp)"
trap 'rm -f "$tmp" "$stats_tmp" "$sorted_tmp"' EXIT
go list -f '{{.ImportPath}}|{{join .Imports ","}}' ./... > "$tmp"

render_graph() {
  awk -F'|' -v module="$module" -v prefix="$module/" '
    BEGIN {
      print "digraph GoPackages {"
      print "  rankdir=LR;"
      print "  graph [label=\"NAEOS direct Go package imports (in-module only)\", labelloc=t];"
    }
    {
      importer = $1
      if (importer == module) {
        importer = "."
      } else if (index(importer, prefix) == 1) {
        importer = substr(importer, length(prefix) + 1)
      } else {
        next
      }
      printf "  \"%s\";\n", importer

      count = split($2, imports, ",")
      for (i = 1; i <= count; i++) {
        imported = imports[i]
        if (imported == module) {
          imported = "."
        } else if (index(imported, prefix) == 1) {
          imported = substr(imported, length(prefix) + 1)
        } else {
          continue
        }
        printf "  \"%s\" -> \"%s\";\n", importer, imported
      }
    }
    END { print "}" }
  ' "$tmp"
}

render_stats() {
  awk -F'|' -v module="$module" -v prefix="$module/" '
    {
      importer = $1
      if (importer == module) {
        importer = "."
      } else if (index(importer, prefix) == 1) {
        importer = substr(importer, length(prefix) + 1)
      } else {
        next
      }
      nodes[importer] = 1

      count = split($2, imports, ",")
      for (i = 1; i <= count; i++) {
        imported = imports[i]
        if (imported == module) {
          imported = "."
        } else if (index(imported, prefix) == 1) {
          imported = substr(imported, length(prefix) + 1)
        } else {
          continue
        }
        nodes[imported] = 1
        outgoing[importer]++
        incoming[imported]++
        edges++
      }
    }
    END {
      print "PACKAGES=" length(nodes)
      print "EDGES=" edges + 0
      for (pkg in nodes) {
        printf "%s\t%d\t%d\n", pkg, incoming[pkg] + 0, outgoing[pkg] + 0
      }
    }
  ' "$tmp" > "$stats_tmp"

  sed -n '1,2p' "$stats_tmp"
  sed -n '3,$p' "$stats_tmp" | sort -t "$(printf '\t')" -k2,2nr -k1,1 > "$sorted_tmp"
  printf '\nTop packages by fan-in (incoming internal imports):\n'
  printf 'package\tfan_in\tfan_out\n'
  sed -n '1,15p' "$sorted_tmp"
  sort -t "$(printf '\t')" -k3,3nr -k1,1 "$sorted_tmp" > "$stats_tmp"
  printf '\nTop packages by fan-out (outgoing internal imports):\n'
  printf 'package\tfan_in\tfan_out\n'
  sed -n '1,15p' "$stats_tmp"
}

if [[ $# -eq 1 && "$1" == "--stats" ]]; then
  render_stats
elif [[ $# -eq 1 ]]; then
  render_graph > "$1"
else
  render_graph
fi
