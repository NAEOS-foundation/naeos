#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Probe CrewAI's LangGraph tool adapter with the NAEOS execution boundary."""
from __future__ import annotations

import argparse
import asyncio
import json
import platform
import sys
from importlib.metadata import version
from pathlib import Path
from typing import Any

import crewai
from crewai.agents.agent_adapters.langgraph.langgraph_tool_adapter import LangGraphToolAdapter
from crewai.hooks.tool_hooks import clear_all_tool_call_hooks, register_before_tool_call_hook
from crewai.tools import BaseTool

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from scripts.crewai_compat_v2_harness import SideEffectLedger, candidate_digest, classify_deny
from scripts.naeos_tool_boundary import BoundaryDecision, PolicyEnforcedTool, normalize_langgraph_async_tool

LEDGER = SideEffectLedger()
HOOK_OBSERVATIONS: list[dict[str, Any]] = []
BOUNDARY_OBSERVATIONS: list[BoundaryDecision] = []


class SideEffectTool(BaseTool):
    name: str = "naeos_v2_langgraph_probe"
    description: str = "Records a local test side effect."

    def _run(self, invocation_id: str, value: str) -> str:
        LEDGER.append(invocation_id=invocation_id, tool_name=self.name, payload={"value": value})
        return "tool-body-executed"


async def invoke_adapter() -> tuple[Any, bool]:
    guarded = PolicyEnforcedTool(
        delegate=SideEffectTool(),
        authorize=lambda _name, _args: False,
        observe=BOUNDARY_OBSERVATIONS.append,
        policy_id="naeos.test.deny-langgraph/1",
    )
    adapter = LangGraphToolAdapter(tools=[guarded])
    adapter.configure_tools([])
    raw_converted = adapter.tools()[0]
    converted = normalize_langgraph_async_tool(raw_converted)
    result = await converted.ainvoke({
        "invocation_id": "langgraph-adapter-deny-001",
        "value": "probe",
    })
    normalized = getattr(converted, "coroutine", None) is not None
    return result, normalized


def run_probe() -> dict[str, Any]:
    global LEDGER
    LEDGER = SideEffectLedger()
    HOOK_OBSERVATIONS.clear()
    BOUNDARY_OBSERVATIONS.clear()
    clear_all_tool_call_hooks()

    def deny_every_tool(context: Any) -> bool:
        HOOK_OBSERVATIONS.append({"tool_name": context.tool_name, "decision": "DENY"})
        return False

    register_before_tool_call_hook(deny_every_tool)
    execution_error = None
    adapter_result = None
    invocation_completed = False
    async_callable_normalized = False
    try:
        adapter_result, async_callable_normalized = asyncio.run(invoke_adapter())
        invocation_completed = True
    except Exception as exc:
        execution_error = f"{type(exc).__name__}: {exc}"

    entries = LEDGER.snapshot()
    hook_called = bool(HOOK_OBSERVATIONS)
    enforcement_observed = any(item.decision == "DENY" for item in BOUNDARY_OBSERVATIONS)
    path_exercised = invocation_completed or enforcement_observed or bool(entries)
    result = classify_deny(
        enforcement_observed=enforcement_observed,
        tool_body_entries=len(entries),
        side_effect_count=len(entries),
        path_exercised=path_exercised,
    )
    limitations = []
    if not path_exercised:
        limitations.append("LangGraph StructuredTool.ainvoke did not reach the NAEOS boundary.")
    if execution_error and not enforcement_observed:
        limitations.append(f"LangGraph adapter invocation raised before enforcement: {execution_error}")

    report = {
        "schema": "naeos.crewai-compatibility-v2.langgraph-adapter/0.2",
        "framework": "crewai",
        "framework_version": crewai.__version__,
        "adapter": "LangGraphToolAdapter",
        "adapter_sdk": "langgraph",
        "adapter_sdk_version": version("langgraph"),
        "langchain_core_version": version("langchain-core"),
        "python_version": platform.python_version(),
        "execution_path": "LangGraphToolAdapter.configure_tools -> NAEOS async normalization -> StructuredTool.ainvoke -> adapter coroutine -> PolicyEnforcedTool.run",
        "decision": "DENY",
        "invocation_completed": invocation_completed,
        "async_callable_normalized": async_callable_normalized,
        "execution_error": execution_error,
        "path_exercised": path_exercised,
        "hook_dispatch_observed": hook_called,
        "hook_observations": HOOK_OBSERVATIONS,
        "enforcement_observed": enforcement_observed,
        "boundary_decisions": [item.__dict__ for item in BOUNDARY_OBSERVATIONS],
        "tool_body_entries": len(entries),
        "side_effect_count": len(entries),
        "side_effect_ledger_checksum": LEDGER.checksum(),
        "adapter_result": None if adapter_result is None else str(adapter_result),
        "compatibility_result": result.value,
        "limitations": limitations,
        "assertions": {
            "naeos_deny_observed": enforcement_observed,
            "tool_body_not_entered": len(entries) == 0,
            "no_independent_side_effect": len(entries) == 0,
        },
        "candidate_digest_probe": candidate_digest({"tool": "naeos_v2_langgraph_probe", "decision": "DENY"}),
    }
    clear_all_tool_call_hooks()
    return report


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    report = run_probe()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2, sort_keys=True))
    raise SystemExit(0 if report["compatibility_result"] == "PASS" else 2 if report["compatibility_result"] == "FAIL" else 3)
