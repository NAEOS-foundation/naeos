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
  (.schema_version == "1.0.0") and
  (.decision_id | type == "string" and length > 0) and
  (.evidence | type == "array" and length > 0) and
  ([.evidence[].evidence_id] | all(type == "string" and length > 0) and (length == (unique | length))) and
  ([.evidence[].digest] | all(test("^[a-f0-9]{64}$"))) and
  ((.required_gates // []) | type == "array" and all(type == "string" and length > 0) and (length == (unique | length))) and
  (.outcome | IN("allow","require_review","deny","verified")) and
  ((.change_classification // {}) | if . == {} then true else
    (.risk | IN("low","medium","high","critical","unknown")) and
    (.criticality | IN("low","medium","high","critical","unknown")) and
    (.decision | IN("allow","require_review","deny"))
  end) and
  ((.verification // {}) | if . == {} then true else
    (.status // "pending" | IN("pending","passed","failed")) and
    (if .status == "passed" then
      (.evidence_digest | type == "string" and test("^[a-f0-9]{64}$"))
    else true end)
  end)
' "$record" >/dev/null

jq -e '
  ([keys[]] - ["record_version","record_id","policy_id","policy_version","schema_version","decision_id","run_id","change_classification","required_gates","evidence","outcome","verification","created_at"]) | length == 0
' "$record" >/dev/null

if jq -e 'has("created_at")' "$record" >/dev/null; then
  jq -e '.created_at | type == "string" and test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(Z|[+-][0-9]{2}:[0-9]{2})$")' "$record" >/dev/null
fi

if jq -e 'has("verification") and (.verification.status == "passed")' "$record" >/dev/null; then
  jq -e '.verification.evidence_digest as $digest | [.evidence[].digest] | index($digest) != null' "$record" >/dev/null
fi

if jq -e '.outcome == "verified"' "$record" >/dev/null; then
  jq -e '.verification.status == "passed"' "$record" >/dev/null
fi

echo "PDR validation: PASS"
