#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Reusable, framework-neutral evidence primitives for CrewAI compatibility v2.

This module does not itself prove framework enforcement. Integration tests must
enter through the real framework/adapter API and use SideEffectLedger to observe
the tool body independently.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import platform
import threading
import uuid
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from enum import Enum
from pathlib import Path
from typing import Any, Mapping


SCHEMA = "naeos.crewai-compatibility-v2.evidence/0.1"
CANONICALIZATION = "naeos-json-c14n/1"


class Result(str, Enum):
    PASS = "PASS"
    FAIL = "FAIL"
    BLOCKED = "BLOCKED"
    UNKNOWN = "UNKNOWN"


class Outcome(str, Enum):
    SUCCEEDED = "SUCCEEDED"
    FAILED = "FAILED"
    UNKNOWN = "UNKNOWN"


def _json_bytes(value: Any) -> bytes:
    """Deterministic JSON encoding; reject NaN and non-JSON values."""
    return json.dumps(
        value, sort_keys=True, separators=(",", ":"), ensure_ascii=False,
        allow_nan=False,
    ).encode("utf-8")


def candidate_digest(candidate: Mapping[str, Any]) -> str:
    envelope = {"canonicalization": CANONICALIZATION, "candidate": candidate}
    return "sha256:" + hashlib.sha256(_json_bytes(envelope)).hexdigest()


@dataclass(frozen=True)
class SideEffect:
    sequence: int
    action_id: str
    invocation_id: str
    tool_name: str
    payload_digest: str
    observed_at: str


class SideEffectLedger:
    """Test-owned append-only ledger; never infer effects from agent output."""

    def __init__(self) -> None:
        self._entries: list[SideEffect] = []
        self._lock = threading.Lock()

    def append(self, *, invocation_id: str, tool_name: str, payload: Mapping[str, Any],
               action_id: str | None = None) -> SideEffect:
        with self._lock:
            entry = SideEffect(
                sequence=len(self._entries) + 1,
                action_id=action_id or str(uuid.uuid4()),
                invocation_id=invocation_id,
                tool_name=tool_name,
                payload_digest=candidate_digest(payload),
                observed_at=datetime.now(timezone.utc).isoformat(),
            )
            self._entries.append(entry)
            return entry

    def snapshot(self) -> tuple[SideEffect, ...]:
        with self._lock:
            return tuple(self._entries)

    def count(self, invocation_id: str | None = None) -> int:
        entries = self.snapshot()
        return sum(1 for item in entries if invocation_id is None or item.invocation_id == invocation_id)

    def checksum(self) -> str:
        return "sha256:" + hashlib.sha256(_json_bytes([asdict(x) for x in self.snapshot()])).hexdigest()


@dataclass
class ScenarioEvidence:
    test_id: str
    run_id: str
    execution_path: str
    framework_version: str
    decision: str
    enforcement_observed: bool
    tool_body_entries: int
    side_effect_count: int
    result: Result
    outcome: Outcome
    assertions: dict[str, bool] = field(default_factory=dict)
    limitations: list[str] = field(default_factory=list)

    def validate(self) -> None:
        if self.result is Result.PASS and not all(self.assertions.values()):
            raise ValueError("PASS requires every recorded assertion to be true")
        if self.result is Result.PASS and self.limitations:
            raise ValueError("PASS cannot conceal unresolved limitations")
        if self.decision == "DENY" and (self.tool_body_entries or self.side_effect_count):
            if self.result is not Result.FAIL:
                raise ValueError("DENY with tool-body entry or side effect must be FAIL")
        if self.result is Result.BLOCKED and not self.limitations:
            raise ValueError("BLOCKED requires an explicit limitation")
        if self.outcome is Outcome.UNKNOWN and self.result is Result.PASS:
            raise ValueError("UNKNOWN outcome cannot be reported as PASS")


def classify_deny(*, enforcement_observed: bool, tool_body_entries: int,
                  side_effect_count: int, path_exercised: bool) -> Result:
    """Strict DENY oracle. An unexercised path is BLOCKED, never PASS."""
    if not path_exercised:
        return Result.BLOCKED
    if tool_body_entries > 0 or side_effect_count > 0:
        return Result.FAIL
    if not enforcement_observed:
        return Result.UNKNOWN
    return Result.PASS


def build_manifest(*, framework_version: str, naeos_commit: str,
                   execution_path: str, dependency_lock_sha256: str,
                   python_version: str | None = None) -> dict[str, Any]:
    return {
        "schema": SCHEMA,
        "created_at": datetime.now(timezone.utc).isoformat(),
        "run_id": str(uuid.uuid4()),
        "framework": "crewai",
        "framework_version": framework_version,
        "naeos_commit": naeos_commit,
        "execution_path": execution_path,
        "python_version": python_version or platform.python_version(),
        "dependency_lock_sha256": dependency_lock_sha256,
        "canonicalization": CANONICALIZATION,
    }


def run_primitive_self_test() -> dict[str, Any]:
    ledger = SideEffectLedger()
    invocation = "selftest-denied-001"
    # Model the body boundary only; this is explicitly not a CrewAI runtime test.
    denied = classify_deny(
        enforcement_observed=True, tool_body_entries=0,
        side_effect_count=ledger.count(invocation), path_exercised=True,
    )
    mutated = candidate_digest({"tool": "crm_read", "input": {"id": "A"}}) != candidate_digest(
        {"tool": "crm_read", "input": {"id": "B"}}
    )
    checks = {
        "deny_zero_effects_passes": denied is Result.PASS,
        "candidate_mutation_changes_digest": mutated,
        "unexercised_path_is_blocked": classify_deny(
            enforcement_observed=False, tool_body_entries=0,
            side_effect_count=0, path_exercised=False,
        ) is Result.BLOCKED,
        "deny_with_body_entry_fails": classify_deny(
            enforcement_observed=True, tool_body_entries=1,
            side_effect_count=1, path_exercised=True,
        ) is Result.FAIL,
        "missing_enforcement_is_not_pass": classify_deny(
            enforcement_observed=False, tool_body_entries=0,
            side_effect_count=0, path_exercised=True,
        ) is Result.UNKNOWN,
    }
    if not all(checks.values()):
        raise AssertionError(checks)
    return {
        "schema": SCHEMA,
        "scope": "primitive self-test only; no CrewAI runtime path exercised",
        "status": "PASS" if all(checks.values()) else "FAIL",
        "framework_compatibility": "NOT_EVALUATED",
        "checks": checks,
        "ledger_entries": len(ledger.snapshot()),
        "ledger_checksum": ledger.checksum(),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--self-test", action="store_true")
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    if not args.self_test:
        parser.error("use --self-test to validate the harness primitives")
    report = run_primitive_self_test()
    rendered = json.dumps(report, indent=2, sort_keys=True) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(rendered, encoding="utf-8")
    print(rendered, end="")
    return 0 if report["status"] == "PASS" else 1


if __name__ == "__main__":
    raise SystemExit(main())
