#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

record="${1:-}"
schema="${2:-governance/policy-decision-record.schema.json}"

if [[ -z "$record" || ! -f "$record" ]]; then
  echo "PDR record is required" >&2
  exit 1
fi
if [[ ! -f "$schema" ]]; then
  echo "PDR schema is required" >&2
  exit 1
fi

command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

jq -e '
  .record_version == "1.0.0" and
  (.record_id | type == "string" and length > 0) and
  (.policy_id | type == "string" and length > 0) and
  (.policy_version | type == "string" and length > 0) and
  (.schema_version | type == "string" and length > 0) and
  (.decision_id | type == "string" and length > 0) and
  (.evidence | type == "array" and length > 0) and
  ([.evidence[].digest] | all(test("^[a-f0-9]{64}$"))) and
  (.outcome | IN("allow","require_review","deny","verified"))
' "$record" >/dev/null

echo "PDR validation: PASS"
