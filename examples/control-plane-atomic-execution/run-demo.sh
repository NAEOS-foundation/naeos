#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

OUTPUT_DIR="/tmp/naeos-p1-8-atomic-execution"
rm -rf "$OUTPUT_DIR"
go run ./examples/control-plane-atomic-execution

echo
echo "P1.8 output: $OUTPUT_DIR"
echo "  - side-effect.json is created inside the atomic execution callback"
echo "  - result.json contains the verification summary"
