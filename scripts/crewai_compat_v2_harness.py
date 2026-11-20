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


class DecisionGateError(PermissionError):
    """Authorization evidence cannot be safely verified or consumed."""


@dataclass(frozen=True)
class AuthorizationDecision:
    decision_id: str
    candidate_digest: str
    policy_id: str
    policy_version: str
    subject: str
    scope: str
    expires_at: float
    decision: str


class DecisionGate:
    """Deterministic decision lifecycle primitive for isolated v2 scenarios.

    This is a test harness model, not a production authorization service.
    Decisions are bound to candidate digest, policy version, subject, scope,
    and expiry; PROCEED decisions are single-use. NEEDS_REVIEW requires an
    explicit approval bound to the exact same candidate digest.
    """

    VALID_DECISIONS = {"PROCEED", "NEEDS_REVIEW", "SILENCE"}

    def __init__(self, *, policy_version_provider: Any, clock: Any) -> None:
        if not callable(policy_version_provider) or not callable(clock):
            raise TypeError("policy_version_provider and clock must be callable")
        self._policy_version_provider = policy_version_provider
        self._clock = clock
        self._decisions: dict[str, AuthorizationDecision] = {}
        self._consumed: set[str] = set()
        self._approvals: dict[str, str] = {}
        self._lock = threading.Lock()

    def _current_policy_version(self) -> str:
        try:
            version = self._policy_version_provider()
        except Exception as exc:
            raise DecisionGateError("policy provider unavailable; fail closed") from exc
        if not isinstance(version, str) or not version.strip():
            raise DecisionGateError("policy version missing or malformed; fail closed")
        return version

    def issue(self, *, candidate: Mapping[str, Any], policy_id: str, subject: str,
               scope: str, expires_at: float, decision: str = "PROCEED",
               decision_id: str | None = None) -> AuthorizationDecision:
        if decision not in self.VALID_DECISIONS:
            raise DecisionGateError("invalid decision literal")
        if not all(isinstance(x, str) and x.strip() for x in (policy_id, subject, scope)):
            raise DecisionGateError("policy, subject, and scope are required")
        if not isinstance(expires_at, (int, float)) or expires_at <= self._clock():
            raise DecisionGateError("decision expiry must be in the future")
        policy_version = self._current_policy_version()
        record = AuthorizationDecision(
            decision_id=decision_id or str(uuid.uuid4()),
            candidate_digest=candidate_digest(candidate),
            policy_id=policy_id,
            policy_version=policy_version,
            subject=subject,
            scope=scope,
            expires_at=float(expires_at),
            decision=decision,
        )
        with self._lock:
            if record.decision_id in self._decisions:
                raise DecisionGateError("decision id already exists")
            self._decisions[record.decision_id] = record
        return record

    def approve(self, *, decision_id: str, candidate: Mapping[str, Any]) -> None:
        with self._lock:
            record = self._decisions.get(decision_id)
            if record is None or record.decision != "NEEDS_REVIEW":
                raise DecisionGateError("no review decision to approve")
            digest = candidate_digest(candidate)
            if digest != record.candidate_digest:
                raise DecisionGateError("approval candidate mismatch")
            if decision_id in self._consumed:
                raise DecisionGateError("decision already consumed")
            self._approvals[decision_id] = digest

    def authorize(self, *, decision_id: str, candidate: Mapping[str, Any],
                  subject: str, scope: str) -> AuthorizationDecision:
        with self._lock:
            record = self._decisions.get(decision_id)
            if record is None:
                raise DecisionGateError("decision not found")
            if decision_id in self._consumed:
                raise DecisionGateError("decision replay rejected")
            if candidate_digest(candidate) != record.candidate_digest:
                raise DecisionGateError("candidate mutation detected")
            if subject != record.subject or scope != record.scope:
                raise DecisionGateError("subject or scope mismatch")
            if self._clock() >= record.expires_at:
                raise DecisionGateError("decision expired")
            if self._current_policy_version() != record.policy_version:
                raise DecisionGateError("policy version changed; stale decision")
            if record.decision == "SILENCE":
                raise DecisionGateError("SILENCE is not executable")
            if record.decision == "NEEDS_REVIEW" and self._approvals.get(decision_id) != record.candidate_digest:
                raise DecisionGateError("explicit approval required for exact candidate")
            if record.decision != "PROCEED" and record.decision != "NEEDS_REVIEW":
                raise DecisionGateError("decision is not executable")
            self._consumed.add(decision_id)
            return record


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


class ActionState(str, Enum):
    PREPARED = "PREPARED"
    DISPATCHED = "DISPATCHED"
    UNKNOWN = "UNKNOWN"
    SUCCEEDED = "SUCCEEDED"
    FAILED = "FAILED"


@dataclass(frozen=True)
class ActionRecord:
    action_id: str
    invocation_id: str
    tool_name: str
    candidate_digest: str
    payload_digest: str
    state: ActionState
    authoritative_effect_count: int = 0
    reconciliation_note: str = ""


class ExecutionLedger:
    """Harness model for response-loss and authoritative effect reconciliation.

    Absence of an effect record is deliberately not treated as proof that an
    action did not execute. Reconciliation only upgrades UNKNOWN to SUCCEEDED
    when exactly one authoritative effect is found for the stable action ID.
    """

    def __init__(self) -> None:
        self._records: dict[str, ActionRecord] = {}
        self._lock = threading.Lock()

    def prepare(self, *, action_id: str, invocation_id: str, tool_name: str,
                candidate: Mapping[str, Any]) -> ActionRecord:
        if not all(isinstance(x, str) and x.strip() for x in
                   (action_id, invocation_id, tool_name)):
            raise ValueError("action_id, invocation_id, and tool_name are required")
        payload = candidate.get("input", candidate)
        record = ActionRecord(action_id, invocation_id, tool_name,
                             candidate_digest(candidate), candidate_digest(payload),
                             ActionState.PREPARED)
        with self._lock:
            if action_id in self._records:
                raise ValueError("duplicate action ID; do not create a new dispatch record")
            self._records[action_id] = record
        return record

    def mark_dispatched(self, *, action_id: str) -> ActionRecord:
        with self._lock:
            record = self._records.get(action_id)
            if record is None or record.state is not ActionState.PREPARED:
                raise ValueError("action must exist in PREPARED state to dispatch")
            record = dataclass_replace(record, state=ActionState.DISPATCHED)
            self._records[action_id] = record
            return record

    def mark_unknown(self, *, action_id: str, reason: str) -> ActionRecord:
        if not reason.strip():
            raise ValueError("UNKNOWN requires an explicit reason")
        with self._lock:
            record = self._records.get(action_id)
            if record is None or record.state is not ActionState.DISPATCHED:
                raise ValueError("only dispatched actions can become UNKNOWN")
            record = dataclass_replace(record, state=ActionState.UNKNOWN,
                                      reconciliation_note=reason)
            self._records[action_id] = record
            return record

    def reconcile(self, *, action_id: str, authoritative_effects: tuple[SideEffect, ...]) -> ActionRecord:
        with self._lock:
            record = self._records.get(action_id)
            if record is None or record.state is not ActionState.UNKNOWN:
                raise ValueError("only UNKNOWN actions can be reconciled")
            matches = tuple(effect for effect in authoritative_effects
                            if effect.action_id == action_id
                            and effect.invocation_id == record.invocation_id
                            and effect.tool_name == record.tool_name
                            and effect.payload_digest == record.payload_digest)
            if len(matches) == 1:
                record = dataclass_replace(record, state=ActionState.SUCCEEDED,
                    authoritative_effect_count=1,
                    reconciliation_note="one authoritative effect matched action ID, invocation, tool, and payload")
            elif len(matches) == 0:
                record = dataclass_replace(record, state=ActionState.UNKNOWN,
                    authoritative_effect_count=0,
                    reconciliation_note="no authoritative match; absence is not proof of non-execution")
            else:
                record = dataclass_replace(record, state=ActionState.UNKNOWN,
                    authoritative_effect_count=len(matches),
                    reconciliation_note="multiple authoritative matches; outcome ambiguous")
            self._records[action_id] = record
            return record

    def get(self, action_id: str) -> ActionRecord | None:
        with self._lock:
            return self._records.get(action_id)

    def retry_allowed(self, *, action_id: str) -> bool:
        """Only a prepared action may be dispatched; UNKNOWN is never retryable here."""
        with self._lock:
            record = self._records.get(action_id)
            return bool(record and record.state is ActionState.PREPARED)


# dataclasses.replace alias avoids mutation of frozen records and keeps transitions explicit.
from dataclasses import replace as dataclass_replace


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
