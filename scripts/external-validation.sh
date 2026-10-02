#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${NAEOS_EXTERNAL_VALIDATION_OUTPUT_DIR:-${RUNNER_TEMP:-/tmp}/naeos-external-validation}"
RECORD="${OUTPUT_DIR}/external-validation-record.json"
DEMO_OUTPUT="${OUTPUT_DIR}/demo"
START_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
SHA="$(git -C "${ROOT_DIR}" rev-parse HEAD)"
REVIEWER="${NAEOS_VALIDATION_REVIEWER:-github-actions/${GITHUB_RUN_ID:-local}}"
mkdir -p "${OUTPUT_DIR}"
rm -rf "${DEMO_OUTPUT}"

set +e
go version > "${OUTPUT_DIR}/toolchain.txt" 2>&1
GO_VERSION_STATUS=$?
if [[ $GO_VERSION_STATUS -eq 0 ]]; then
  go -C "${ROOT_DIR}" build -o "${ROOT_DIR}/naeos" ./cmd/naeos > "${OUTPUT_DIR}/build.log" 2>&1
  BUILD_STATUS=$?
else
  BUILD_STATUS=$GO_VERSION_STATUS
fi

if [[ $BUILD_STATUS -eq 0 ]]; then
  NAEOS_BIN="${ROOT_DIR}/naeos" NAEOS_DEMO_OUTPUT_DIR="${DEMO_OUTPUT}"     "${ROOT_DIR}/examples/demo-cli/run-demo.sh" > "${OUTPUT_DIR}/demo.log" 2>&1
  DEMO_STATUS=$?
else
  DEMO_STATUS=$BUILD_STATUS
fi
set -e

END_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

python3 - "${OUTPUT_DIR}" "${RECORD}" "${SHA}" "${START_UTC}" "${END_UTC}" "${REVIEWER}" "${BUILD_STATUS}" "${DEMO_STATUS}" <<'PY'
import json, os, sys

out, record, sha, start, end, reviewer, build_status, demo_status = sys.argv[1:]
demo = os.path.join(out, "demo")
required = [
    "spec.yaml", "inspect.json", "validate.json", "invalid-policy.log",
    "context.md", "context.json", "run.json", "summary.md",
    os.path.join("generated", "README.md"),
    os.path.join("generated", "go.mod"),
    os.path.join("generated", "package.json"),
]
checks = []

def check(name, ok, evidence):
    checks.append({"name": name, "result": "PASS" if ok else "FAIL", "evidence": evidence})

check("Clean checkout commit", bool(sha), sha)
check("Go/toolchain", os.path.exists(os.path.join(out, "toolchain.txt")), os.path.join(out, "toolchain.txt"))
check("CLI build", build_status == "0", os.path.join(out, "build.log"))
check("Canonical demo", demo_status == "0", os.path.join(out, "demo.log"))

run = {}
run_path = os.path.join(demo, "run.json")
if os.path.isfile(run_path):
    try:
        with open(run_path, encoding="utf-8") as f:
            run = json.load(f)
    except Exception:
        run = {}

for rel in required:
    check(f"Evidence: {rel}", os.path.isfile(os.path.join(demo, rel)), os.path.join("demo", rel))

check("Run traceability", all(run.get(k) for k in ("run_id", "specification_hash", "neir_hash")), "demo/run.json")

policy_log = os.path.join(demo, "invalid-policy.log")
policy_text = open(policy_log, encoding="utf-8").read() if os.path.isfile(policy_log) else ""
check("Invalid policy rejected", "policy evaluation failed" in policy_text, "demo/invalid-policy.log")

generated = os.path.join(demo, "generated")
artifact_count = sum(1 for root, _, files in os.walk(generated) for name in files) if os.path.isdir(generated) else 0

record_data = {
    "record_version": "1.0.0",
    "record_type": "NAEOS External Validation",
    "repository": "NAEOS-foundation/naeos",
    "commit_sha": sha,
    "go_toolchain": open(os.path.join(out, "toolchain.txt"), encoding="utf-8").read().strip(),
    "demo_command": "NAEOS_BIN=./naeos NAEOS_DEMO_OUTPUT_DIR=<isolated-dir> ./examples/demo-cli/run-demo.sh",
    "start_time_utc": start,
    "end_time_utc": end,
    "reviewer_executor": reviewer,
    "exit_status": {"build": int(build_status), "demo": int(demo_status)},
    "run_id": run.get("run_id", ""),
    "specification_hash": run.get("specification_hash", ""),
    "neir_hash": run.get("neir_hash", ""),
    "generated_artifact_count": artifact_count,
    "observed_invalid_policy_rejection": "policy evaluation failed" in policy_text,
    "checks": checks,
    "deviations": [],
    "result": "PASS" if all(c["result"] == "PASS" for c in checks) else "FAIL",
    "evidence_directory": "demo/",
}
with open(record, "w", encoding="utf-8") as f:
    json.dump(record_data, f, indent=2, sort_keys=True)
    f.write("\n")

print(json.dumps(record_data, indent=2, sort_keys=True))
if record_data["result"] != "PASS":
    raise SystemExit(1)
PY
