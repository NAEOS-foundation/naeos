#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
OUTPUT_DIR="/tmp/naeos-p1-7-policy-change"
rm -rf "$OUTPUT_DIR"
go run ./examples/control-plane-policy-change

echo
echo "P1.7 output: $OUTPUT_DIR"
echo "  - result.json contains the stale-authorization verification"
echo "  - no side-effect.json must exist"
