#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

record="${1:-}"
if [[ -z "$record" || ! -f "$record" ]]; then
  echo "Legal evidence record is required" >&2
  exit 1
fi

command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

jq -e '
  (.schema_version == "1.0") and
  (.release_version | type == "string" and length > 0) and
  (.repository == "NAEOS-foundation/naeos") and
  (.commit_sha | type == "string" and test("^[a-f0-9]{40}$")) and
  (.release_date | type == "string" and test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$")) and
  (.license == "Apache-2.0") and
  (.dco_status | IN("PASS","HOLD","REVIEWED","NOT_APPLICABLE")) and
  (.dependency_license_status | IN("PASS","HOLD","REVIEWED","NOT_APPLICABLE")) and
  (.attribution_status | IN("PASS","HOLD","REVIEWED","NOT_APPLICABLE")) and
  (.security_status | IN("PASS","HOLD","REVIEWED","NOT_APPLICABLE")) and
  (.provenance_status | IN("PASS","HOLD","REVIEWED","NOT_APPLICABLE")) and
  (.trademark_status | IN("PASS","HOLD","REVIEWED","NOT_APPLICABLE")) and
  (.sbom | type == "object") and
  (.sbom.artifact | type == "string" and length > 0) and
  (.sbom.sha256 | type == "string" and test("^[a-f0-9]{64}$")) and
  (.exceptions | type == "array") and
  (.reviewer | type == "string" and length > 0) and
  (.decision | IN("PASS","HOLD"))
' "$record" >/dev/null

jq -e '
  ([keys[]] - [
    "schema_version","release_version","repository","commit_sha","release_date",
    "license","dco_status","sbom","dependency_license_status","attribution_status",
    "security_status","provenance_status","trademark_status","exceptions","reviewer","decision"
  ]) | length == 0
' "$record" >/dev/null

if jq -e '
  .decision == "PASS" and
  ([.dco_status,.dependency_license_status,.attribution_status,.security_status,.provenance_status,.trademark_status]
    | any(. == "HOLD"))
' "$record" >/dev/null; then
  echo "PASS decision cannot contain HOLD status" >&2
  exit 1
fi

if jq -e '
  any([
    .release_version,.commit_sha,.release_date,.reviewer,.sbom.artifact,.sbom.sha256
  ][]; test("^<.*>$"))
' "$record" >/dev/null; then
  echo "Placeholder values are not valid legal evidence" >&2
  exit 1
fi

echo "Legal evidence validation: PASS"
