#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Probe CrewAI Flow ToolAction through the framework's action builder.

This deliberately records a compatibility FAIL when the global before-tool
hook is bypassed and the tool body executes. A green unittest means the probe
observed and documented the framework behavior; it does NOT mean compatibility
passed.
"""
from __future__ import annotations

import json
import platform
import sys
import unittest
from pathlib import Path
from typing import Any

import crewai
from crewai.flow.flow_definition import FlowToolActionDefinition
from crewai.flow.runtime import Flow
from crewai.flow.runtime._actions import build_action
from crewai.hooks.tool_hooks import (
    clear_all_tool_call_hooks,
    register_before_tool_call_hook,
)
from crewai.tools import BaseTool

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from scripts.crewai_compat_v2_harness import SideEffectLedger, candidate_digest


LEDGER = SideEffectLedger()
HOOK_OBSERVATIONS: list[dict[str, Any]] = []


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


def run_probe() -> dict[str, Any]:
    global LEDGER
    LEDGER = SideEffectLedger()
    clear_all_tool_call_hooks()
    HOOK_OBSERVATIONS.clear()

    def deny_every_tool(context: Any) -> bool:
        HOOK_OBSERVATIONS.append({"tool_name": context.tool_name, "decision": "DENY"})
        return False

    register_before_tool_call_hook(deny_every_tool)
    flow = Flow()
    definition = FlowToolActionDefinition(
        call="tool",
        ref=f"{SideEffectTool.__module__}:SideEffectTool",
        with_={"invocation_id": "flow-tool-action-deny-001", "value": "probe"},
    )
    action = build_action(flow, definition)
    result = action()

    entries = LEDGER.snapshot()
    hook_called = bool(HOOK_OBSERVATIONS)
    body_executed = len(entries) > 0
    # Hook-based DENY should prevent the tool body and independent side effect.
    compatibility = "PASS" if hook_called and not body_executed else "FAIL"
    report = {
        "schema": "naeos.crewai-compatibility-v2.flow-tool-action/0.1",
        "framework": "crewai",
        "framework_version": crewai.__version__,
        "python_version": platform.python_version(),
        "execution_path": "Flow runtime build_action -> ToolAction.run -> BaseTool.run",
        "decision": "DENY",
        "hook_dispatch_observed": hook_called,
        "hook_observations": HOOK_OBSERVATIONS,
        "tool_body_entries": len(entries),
        "side_effect_count": len(entries),
        "side_effect_ledger_checksum": LEDGER.checksum(),
        "tool_result": str(result),
        "compatibility_result": compatibility,
        "assertions": {
            "deny_hook_was_dispatched": hook_called,
            "tool_body_not_entered": not body_executed,
            "no_independent_side_effect": len(entries) == 0,
        },
        "candidate_digest_probe": candidate_digest({"tool": "naeos_v2_flow_probe", "decision": "DENY"}),
        "interpretation": (
            "PASS means DENY was dispatched and no tool body/side effect occurred. "
            "FAIL means the Flow ToolAction path bypassed the hook or executed the body."
        ),
    }
    clear_all_tool_call_hooks()
    return report


class FlowToolActionRuntimeTests(unittest.TestCase):
    def test_deny_hook_blocks_tool_body_on_flow_tool_action(self) -> None:
        report = run_probe()
        expected = (
            "FAIL"
            if report["tool_body_entries"] > 0
            or report["side_effect_count"] > 0
            or not report["hook_dispatch_observed"]
            else "PASS"
        )
        self.assertEqual(report["compatibility_result"], expected, msg=json.dumps(report, indent=2, sort_keys=True))
        self.assertEqual(report["tool_body_entries"], report["side_effect_count"])

    @classmethod
    def tearDownClass(cls) -> None:
        clear_all_tool_call_hooks()


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    report = run_probe()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2, sort_keys=True))
    raise SystemExit(0 if report["compatibility_result"] == "PASS" else 2)
