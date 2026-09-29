#!/usr/bin/env bash
set -euo pipefail

OUTPUT_DIR="/tmp/naeos-p1-6-golden-path"
rm -rf "$OUTPUT_DIR"
go run ./examples/control-plane-golden-path

echo
echo "P1.6 output: $OUTPUT_DIR"
echo "  - allow-side-effect.json exists only for ALLOW"
echo "  - deny-side-effect.json must not exist"
echo "  - result.json contains the verification summary"
