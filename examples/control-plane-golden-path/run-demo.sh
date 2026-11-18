#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [[ -n "${NAEOS_CONTROL_PLANE_DEMO_OUTPUT_DIR:-}" ]]; then
  OUTPUT_DIR="$NAEOS_CONTROL_PLANE_DEMO_OUTPUT_DIR"
  if [[ -e "$OUTPUT_DIR" ]]; then
    printf 'Refusing to reuse existing demo output directory: %s\n' "$OUTPUT_DIR" >&2
    printf 'Choose a new path or remove the directory yourself after reviewing its contents.\n' >&2
    exit 2
  fi
else
  OUTPUT_DIR="$(mktemp -d "${TMPDIR:-/tmp}/naeos-p1-6-golden-path.XXXXXX")"
  rmdir "$OUTPUT_DIR"
fi

export NAEOS_CONTROL_PLANE_DEMO_OUTPUT_DIR="$OUTPUT_DIR"
go run ./examples/control-plane-golden-path

echo
echo "P1.6 output: $OUTPUT_DIR"
echo "  - allow-side-effect.json exists only for ALLOW"
echo "  - deny-side-effect.json must not exist"
echo "  - result.json contains the verification summary"
