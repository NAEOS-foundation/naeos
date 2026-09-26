#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

record="governance/examples/policy-decision-record.json"
validator="scripts/validate-policy-decision-record.sh"

run_expect_fail() {
  local fixture="$1"
  if bash "$validator" "$fixture"; then
    echo "validator unexpectedly accepted $fixture" >&2
    exit 1
  fi
}

bash "$validator" "$record"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

jq '.schema_version = "2.0.0"' "$record" > "$tmp/version.json"
run_expect_fail "$tmp/version.json"

jq '.evidence = []' "$record" > "$tmp/no-evidence.json"
run_expect_fail "$tmp/no-evidence.json"

jq '.evidence[0].digest = "invalid"' "$record" > "$tmp/bad-digest.json"
run_expect_fail "$tmp/bad-digest.json"

jq '.evidence[0].evidence_id = .evidence[0].evidence_id' "$record" > "$tmp/valid-id.json"
bash "$validator" "$tmp/valid-id.json"

jq '.verification.evidence_digest = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"' "$record" > "$tmp/unbound-verification.json"
run_expect_fail "$tmp/unbound-verification.json"

jq '.outcome = "verified" | .verification.status = "pending"' "$record" > "$tmp/unverified-outcome.json"
run_expect_fail "$tmp/unverified-outcome.json"

jq '.unexpected = true' "$record" > "$tmp/extra-property.json"
run_expect_fail "$tmp/extra-property.json"

echo "PDR validator regression tests: PASS"
