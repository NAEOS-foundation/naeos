#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""CrewAI Flow ToolAction integration probe for response-loss reconciliation.

This exercises the public Flow.kickoff path. The side effect is committed by
the delegate tool body before a simulated transport timeout is raised.
"""
from __future__ import annotations

import argparse
import json
import platform
import unittest
from pathlib import Path
from typing import Any

import crewai
from crewai.flow import Flow, start
from crewai.flow.flow_definition import FlowToolActionDefinition
from crewai.flow.runtime._actions import build_action
from crewai.tools import BaseTool

from scripts.crewai_compat_v2_harness import (
    ActionState,
    ExecutionLedger,
    SideEffectLedger,
)
from scripts.naeos_tool_boundary import PolicyEnforcedTool

ACTION_ID = "response-loss-action-stable-001"
INVOCATION_ID = "response-loss-invocation-001"
TOOL_NAME = "naeos_v2_response_loss_commit_probe"
PAYLOAD = {"record_id": "record-42", "value": "committed-before-timeout"}

EFFECTS = SideEffectLedger()
EXECUTIONS = ExecutionLedger()
TOOL_BODY_ENTRIES = 0
BOUNDARY_OBSERVATIONS: list[dict[str, Any]] = []


class CommitThenLoseResponseDelegate(BaseTool):
    name: str = TOOL_NAME
    description: str = "Commits a controlled local effect, then the proxy simulates response loss."

    def _run(self, action_id: str, invocation_id: str, record_id: str, value: str) -> str:
        global TOOL_BODY_ENTRIES
        TOOL_BODY_ENTRIES += 1
        EFFECTS.append(
            action_id=action_id,
            invocation_id=invocation_id,
            tool_name=self.name,
            payload={"record_id": record_id, "value": value},
        )
        return "committed"


class ResponseLossTool(PolicyEnforcedTool):
    def __init__(self, **data: Any) -> None:
        super().__init__(
            delegate=CommitThenLoseResponseDelegate(),
            authorize=lambda _name, _args: True,
            observe=lambda event: BOUNDARY_OBSERVATIONS.append(event.__dict__),
            policy_id="naeos.test.response-loss/1",
        )

    def _run(self, **kwargs: Any) -> Any:
        self._decide(kwargs)
        result = self._delegate.run(**kwargs)
        # Simulate the transport/client response being lost after delegate commit.
        raise TimeoutError(f"simulated response loss after commit (delegate={result})")


class ResponseLossFlow(Flow):
    @start()
    def invoke_action(self) -> str:
        definition = FlowToolActionDefinition(
            call="tool",
            ref=f"{ResponseLossTool.__module__}:ResponseLossTool",
            with_={"action_id": ACTION_ID, "invocation_id": INVOCATION_ID, **PAYLOAD},
        )
        return build_action(self, definition)()


def run_probe() -> dict[str, Any]:
    global EFFECTS, EXECUTIONS, TOOL_BODY_ENTRIES
    EFFECTS = SideEffectLedger()
    EXECUTIONS = ExecutionLedger()
    TOOL_BODY_ENTRIES = 0
    BOUNDARY_OBSERVATIONS.clear()

    candidate = {"tool_name": TOOL_NAME, "input": PAYLOAD}
    EXECUTIONS.prepare(
        action_id=ACTION_ID,
        invocation_id=INVOCATION_ID,
        tool_name=TOOL_NAME,
        candidate=candidate,
    )
    EXECUTIONS.mark_dispatched(action_id=ACTION_ID)

    flow_error = None
    try:
        ResponseLossFlow().kickoff()
    except Exception as exc:
        flow_error = f"{type(exc).__name__}: {exc}"

    entries = EFFECTS.snapshot()
    dispatched_record = EXECUTIONS.get(ACTION_ID)
    if TOOL_BODY_ENTRIES == 1 and len(entries) == 1 and flow_error:
        unknown = EXECUTIONS.mark_unknown(
            action_id=ACTION_ID,
            reason="Flow ToolAction committed delegate effect; response was lost/timeout observed",
        )
    else:
        unknown = None

    retry_allowed_before_reconcile = EXECUTIONS.retry_allowed(action_id=ACTION_ID)
    reconciled = (
        EXECUTIONS.reconcile(action_id=ACTION_ID, authoritative_effects=entries)
        if unknown is not None
        else EXECUTIONS.get(ACTION_ID)
    )

    return {
        "schema": "naeos.crewai-compatibility-v2.response-loss-runtime/0.1",
        "framework": "crewai",
        "framework_version": crewai.__version__,
        "python_version": platform.python_version(),
        "execution_path": "public Flow.kickoff -> Flow ToolAction -> PolicyEnforcedTool -> delegate commit -> simulated response loss",
        "action_id": ACTION_ID,
        "invocation_id": INVOCATION_ID,
        "tool_name": TOOL_NAME,
        "payload_digest": reconciled.payload_digest,
        "flow_error": flow_error,
        "dispatch_state_before_unknown": dispatched_record.state.value,
        "tool_body_entries": TOOL_BODY_ENTRIES,
        "side_effect_count": len(entries),
        "side_effects": [entry.__dict__ for entry in entries],
        "boundary_observations": BOUNDARY_OBSERVATIONS,
        "state_after_response_loss": None if unknown is None else unknown.state.value,
        "retry_allowed_before_reconcile": retry_allowed_before_reconcile,
        "reconciled_state": reconciled.state.value,
        "authoritative_effect_count": reconciled.authoritative_effect_count,
        "reconciliation_note": reconciled.reconciliation_note,
        "assertions": {
            "flow_tool_action_path_reached": bool(flow_error and TOOL_BODY_ENTRIES),
            "single_committed_side_effect": len(entries) == 1,
            "response_loss_classified_unknown": unknown is not None and unknown.state is ActionState.UNKNOWN,
            "retry_forbidden_after_dispatch": not retry_allowed_before_reconcile,
            "stable_action_id_matched_authoritative_effect": reconciled.action_id == ACTION_ID and reconciled.authoritative_effect_count == 1,
            "reconciliation_resolves_success": reconciled.state is ActionState.SUCCEEDED,
        },
    }


class ResponseLossRuntimeTests(unittest.TestCase):
    def test_public_flow_tool_action_response_loss_reconciles_without_retry(self) -> None:
        report = run_probe()
        for name, assertion in report["assertions"].items():
            with self.subTest(assertion=name):
                self.assertTrue(assertion, report)
        self.assertEqual(report["tool_body_entries"], 1)
        self.assertEqual(report["side_effect_count"], 1)
        self.assertEqual(report["state_after_response_loss"], "UNKNOWN")
        self.assertFalse(report["retry_allowed_before_reconcile"])
        self.assertEqual(report["reconciled_state"], "SUCCEEDED")
        self.assertEqual(report["authoritative_effect_count"], 1)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    report = run_probe()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2, sort_keys=True))
    raise SystemExit(0 if all(report["assertions"].values()) else 2)
