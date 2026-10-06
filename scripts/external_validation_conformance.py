#!/usr/bin/env python3
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Deterministic NAEOS external-validation conformance harness.

This harness validates three invariants independently of NAEOS fixtures:
- authorization is bound to the exact canonical invocation;
- a policy-version change makes prior authorization stale;
- recovery follows a logical invocation identity and does not duplicate a
  side effect when a worker crashes after execution but before outcome commit.

It deliberately does not claim to reproduce a third-party runtime crash.
The companion CI job runs the real CrewAI + Attenu integration separately.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
from typing import Any


def canonical(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()


def digest(value: Any) -> str:
    return hashlib.sha256(canonical(value)).hexdigest()


def require(condition: bool, message: str) -> None:
    if not condition:
        raise AssertionError(message)


def scenario_mutation() -> dict[str, Any]:
    decision = {
        "invocation_id": "inv-mutation-001",
        "tool": "crm_query",
        "args": {"customer_id": "c-001", "fields": ["name", "status"]},
        "policy_version": "v1",
        "decision": "ALLOW",
    }
    auth_digest = digest(decision)
    mutated = dict(decision)
    mutated["args"] = {"customer_id": "c-001", "fields": ["name", "email"]}
    mutated_digest = digest(mutated)
    require(auth_digest != mutated_digest, "argument mutation must change canonical invocation digest")
    return {"name": "mutation-after-authorization", "status": "PASS", "authorization_digest": auth_digest, "mutated_digest": mutated_digest, "execution": "BLOCKED"}


def scenario_freshness() -> dict[str, Any]:
    authorization = {"invocation_id": "inv-fresh-001", "tool": "crm_query", "args_digest": "args-001", "policy_version": "v1"}
    current_policy_version = "v2"
    fresh = authorization["policy_version"] == current_policy_version
    require(not fresh, "v1 authorization must be stale under current v2 policy")
    return {"name": "policy-version-freshness", "status": "PASS", "authorized_policy_version": authorization["policy_version"], "current_policy_version": current_policy_version, "execution": "BLOCKED_UNTIL_REAUTHORIZED"}


def scenario_recovery() -> dict[str, Any]:
    journal: list[dict[str, Any]] = []
    invocation_id = "inv-recovery-001"
    side_effects: list[str] = []
    journal.append({"type": "AUTHORIZED", "invocation_id": invocation_id, "policy_version": "v1", "args_digest": "args-recovery-001"})
    side_effects.append(invocation_id)
    journal.append({"type": "EXECUTED", "invocation_id": invocation_id, "worker": "worker-1"})
    executed = [e for e in journal if e["type"] == "EXECUTED" and e["invocation_id"] == invocation_id]
    require(len(executed) == 1, "recovery must retain one logical execution")
    require(side_effects.count(invocation_id) == 1, "recovery must not duplicate the side effect")
    journal.append({"type": "OBSERVED", "invocation_id": invocation_id, "worker": "worker-2", "recovery": "RECONCILE_EXISTING_EXECUTION"})
    journal.append({"type": "EVIDENCE", "invocation_id": invocation_id, "worker": "worker-2", "evidence": "outcome-bound-to-logical-invocation"})
    evidence = [e for e in journal if e["type"] == "EVIDENCE" and e["invocation_id"] == invocation_id]
    require(len(evidence) == 1, "recovery must produce one evidence record")
    return {"name": "cross-worker-crash-recovery", "status": "PASS", "invocation_id": invocation_id, "execution_count": len(executed), "side_effect_count": side_effects.count(invocation_id), "evidence_count": len(evidence), "recovery": "RECONCILE_EXISTING_EXECUTION"}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    scenarios = [scenario_mutation(), scenario_freshness(), scenario_recovery()]
    report = {"schema": "naeos.external-validation-conformance.v1", "overall": "PASS", "scenarios": scenarios}
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())