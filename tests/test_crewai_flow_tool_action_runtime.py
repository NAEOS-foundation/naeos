#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Probe CrewAI Flow ToolAction through the public Flow runtime kickoff.

This is an execution-path probe, not a generic unit test of CrewAI hooks.
The independent ledger is changed only inside the test tool body.
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
from crewai.hooks.tool_hooks import (
    clear_all_tool_call_hooks,
    register_before_tool_call_hook,
)
from crewai.tools import BaseTool

from scripts.naeos_tool_boundary import BoundaryDecision, PolicyEnforcedTool\n\nfrom scripts.crewai_compat_v2_harness import (
    Result,
    SideEffectLedger,
    candidate_digest,
    classify_deny,
)

LEDGER = SideEffectLedger()
HOOK_OBSERVATIONS: list[dict[str, Any]] = []\nBOUNDARY_OBSERVATIONS: list[BoundaryDecision] = []


class ProbeFlow(Flow):
    """Concrete Flow subclass required by CrewAI's Pydantic metaclass."""


class SideEffectTool(BaseTool):
    name: str = "naeos_v2_flow_probe"
    description: str = "Records a local test side effect."

    def _run(self, invocation_id: str, value: str) -> str:
        LEDGER.append(
            invocation_id=invocation_id,
            tool_name=self.name,
            payload={"value": value},
        )
        return "tool-body-executed"


class GuardedFlowTool(PolicyEnforcedTool):\n    def __init__(self, **data: Any) -> None:\n        super().__init__(\n            delegate=SideEffectTool(),\n            authorize=lambda _name, _args: False,\n            observe=BOUNDARY_OBSERVATIONS.append,\n            policy_id="naeos.test.deny-flow/1",\n        )\n\n\nclass ToolActionProbeFlow(Flow):
    @start()
    def invoke_tool_action(self) -> str:
        definition = FlowToolActionDefinition(
            call="tool",
            ref=f"{GuardedFlowTool.__module__}:GuardedFlowTool",
            with_={
                "invocation_id": "flow-tool-action-deny-001",
                "value": "probe",
            },
        )
        action = build_action(self, definition)
        return action()


def run_probe() -> dict[str, Any]:
    global LEDGER
    LEDGER = SideEffectLedger()
    HOOK_OBSERVATIONS.clear()
    clear_all_tool_call_hooks()

    def deny_every_tool(context: Any) -> bool:
        HOOK_OBSERVATIONS.append(
            {"tool_name": context.tool_name, "decision": "DENY"}
        )
        return False

    register_before_tool_call_hook(deny_every_tool)
    execution_error = None
    execution_completed = False
    tool_result = None

    try:
        flow = ToolActionProbeFlow()
        tool_result = flow.kickoff()
        execution_completed = True
    except Exception as exc:  # retain evidence for runtime/API incompatibilities
        execution_error = f"{type(exc).__name__}: {exc}"

    entries = LEDGER.snapshot()
    hook_called = bool(HOOK_OBSERVATIONS)
    path_exercised = execution_completed or hook_called or bool(entries)
    result = classify_deny(
        enforcement_observed=enforcement_observed,
        tool_body_entries=len(entries),
        side_effect_count=len(entries),
        path_exercised=path_exercised,
    )

    limitations = []
    if not path_exercised:
        limitations.append(
            "Flow ToolAction did not reach an observable hook or tool-body boundary."
        )
    if execution_error:
        limitations.append(f"ToolAction invocation raised: {execution_error}")

    report = {
        "schema": "naeos.crewai-compatibility-v2.flow-tool-action/0.2",
        "framework": "crewai",
        "framework_version": crewai.__version__,
        "python_version": platform.python_version(),
        "execution_path": "public Flow.kickoff -> @start method -> build_action -> ToolAction.run -> BaseTool.run",
        "decision": "DENY",
        "execution_completed": execution_completed,
        "execution_error": execution_error,
        "path_exercised": path_exercised,
        "hook_dispatch_observed": hook_called,
        "hook_observations": HOOK_OBSERVATIONS,
        "tool_body_entries": len(entries),
        "side_effect_count": len(entries),
        "side_effect_ledger_checksum": LEDGER.checksum(),
        "tool_result": None if tool_result is None else str(tool_result),
        "compatibility_result": result.value,
        "limitations": limitations,
        "assertions": {
            "deny_hook_was_dispatched": hook_called,
            "tool_body_not_entered": len(entries) == 0,
            "no_independent_side_effect": len(entries) == 0,
        },
        "candidate_digest_probe": candidate_digest(
            {"tool": "naeos_v2_flow_probe", "decision": "DENY"}
        ),
    }
    clear_all_tool_call_hooks()
    return report


class FlowToolActionRuntimeTests(unittest.TestCase):
    def test_deny_classification_matches_independent_observations(self) -> None:
        report = run_probe()
        expected = classify_deny(
            enforcement_observed=report["enforcement_observed"],
            tool_body_entries=report["tool_body_entries"],
            side_effect_count=report["side_effect_count"],
            path_exercised=report["path_exercised"],
        )
        self.assertEqual(report["compatibility_result"], expected.value)
        self.assertEqual(report["tool_body_entries"], report["side_effect_count"])
        if expected is Result.BLOCKED:
            self.assertTrue(report["limitations"])


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    report = run_probe()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(
        json.dumps(report, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    print(json.dumps(report, indent=2, sort_keys=True))
    if report["compatibility_result"] == Result.PASS.value:
        raise SystemExit(0)
    if report["compatibility_result"] == Result.FAIL.value:
        raise SystemExit(2)
    raise SystemExit(3)
