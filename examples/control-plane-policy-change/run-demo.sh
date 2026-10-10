#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
OUTPUT_DIR="/tmp/naeos-p1-7-policy-change"
rm -rf "$OUTPUT_DIR"

go test ./internal/controlplane -run TestPolicyChangeRevokingCapabilityBlocksStaleAndReauthorizedExecution -count=1
go run ./examples/control-plane-policy-change

echo
echo "P1.7 output: $OUTPUT_DIR"
echo "  - result.json records stale rejection, explicit v2 revocation, and verification"
echo "  - side-effect.json must not exist"
if [[ -e "$OUTPUT_DIR/side-effect.json" ]]; then
  echo "ERROR: side effect was observed despite policy denial" >&2
  exit 1
fi
