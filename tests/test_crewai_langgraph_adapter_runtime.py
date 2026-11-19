#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Probe the CrewAI LangGraph adapter through LangGraph's ToolNode runtime."""
from __future__ import annotations

import argparse
import asyncio
import inspect
import json
import platform
import sys
from importlib.metadata import version
from types import SimpleNamespace
from pathlib import Path
from typing import Any

import crewai
from pydantic import PrivateAttr

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from crewai.agents.agent_adapters.langgraph.langgraph_tool_adapter import LangGraphToolAdapter
from crewai.hooks.tool_hooks import clear_all_tool_call_hooks, register_before_tool_call_hook
from crewai.tools import BaseTool
from langchain_core.messages import AIMessage
from langgraph.prebuilt import ToolNode
from langgraph.runtime import ExecutionInfo

from scripts.crewai_compat_v2_harness import SideEffectLedger, candidate_digest, classify_deny


LEDGER = SideEffectLedger()
HOOK_OBSERVATIONS: list[dict[str, Any]] = []


class SideEffectTool(BaseTool):
    name: str = "naeos_v2_langgraph_probe"
    description: str = "Records a local test side effect."

    def _run(self, invocation_id: str, value: str) -> str:
        LEDGER.append(
            invocation_id=invocation_id,
            tool_name=self.name,
            payload={"value": value},
        )
        return "tool-body-executed"


async def invoke_adapter() -> Any:
    tool = SideEffectTool()
    adapter = LangGraphToolAdapter(tools=[tool])
    adapter.configure_tools([])
    converted = adapter.tools()[0]
    node = ToolNode([converted])
    runtime = SimpleNamespace(
        store=None,
        context=None,
        stream_writer=lambda *args, **kwargs: None,
        execution_info=ExecutionInfo(
            checkpoint_id="naeos-test-checkpoint",
            checkpoint_ns="",
            task_id="naeos-test-task",
        ),
        server_info=None,
    )
    message = AIMessage(
        content="",
        tool_calls=[
            {
                "name": converted.name,
                "args": {"invocation_id": "langgraph-adapter-deny-001", "value": "probe"},
                "id": "langgraph-deny-call-001",
                "type": "tool_call",
            }
        ],
    )
    return await node.ainvoke(
        {"messages": [message]},
        config={"configurable": {"__pregel_runtime": runtime}},
    )


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
    adapter_result = None
    invocation_completed = False
    unawaited_coroutine_returned = False
    try:
        adapter_result = asyncio.run(invoke_adapter())
        invocation_completed = True
        tool_messages = adapter_result.get("messages", []) if isinstance(adapter_result, dict) else []
        tool_content = tool_messages[-1].content if tool_messages else None
        unawaited_coroutine_returned = inspect.isawaitable(tool_content) or (
            isinstance(tool_content, str) and "coroutine object" in tool_content
        )
        if inspect.iscoroutine(tool_content):
            tool_content.close()
    except Exception as exc:
        execution_error = f"{type(exc).__name__}: {exc}"

    entries = LEDGER.snapshot()
    hook_called = bool(HOOK_OBSERVATIONS)
    path_exercised = invocation_completed or hook_called or bool(entries)
    result = classify_deny(
        enforcement_observed=hook_called,
        tool_body_entries=len(entries),
        side_effect_count=len(entries),
        path_exercised=path_exercised,
    )
    limitations = []
    if not path_exercised:
        limitations.append("LangGraph ToolNode did not reach the adapter invocation boundary.")
    if execution_error:
        limitations.append(f"Adapter invocation raised: {execution_error}")
    if unawaited_coroutine_returned:
        limitations.append("The converted StructuredTool returned an unawaited coroutine; tool execution and enforcement remain UNKNOWN.")

    report = {
        "schema": "naeos.crewai-compatibility-v2.langgraph-adapter/0.1",
        "framework": "crewai",
        "framework_version": crewai.__version__,
        "adapter": "LangGraphToolAdapter",
        "adapter_sdk": "langgraph",
        "adapter_sdk_version": version("langgraph"),
        "langchain_core_version": version("langchain-core"),
        "python_version": platform.python_version(),
        "execution_path": "LangGraphToolAdapter.configure_tools -> LangGraph ToolNode.ainvoke -> StructuredTool.ainvoke -> adapter wrapper -> BaseTool.run",
        "decision": "DENY",
        "invocation_completed": invocation_completed,
        "adapter_result_type": None if adapter_result is None else type(adapter_result).__name__,
        "tool_node_result_content_type": None if "tool_content" not in locals() or tool_content is None else type(tool_content).__name__,
        "adapter_returned_unawaited_coroutine": unawaited_coroutine_returned,
        "execution_error": execution_error,
        "path_exercised": path_exercised,
        "hook_dispatch_observed": hook_called,
        "hook_observations": HOOK_OBSERVATIONS,
        "tool_body_entries": len(entries),
        "side_effect_count": len(entries),
        "side_effect_ledger_checksum": LEDGER.checksum(),
        "adapter_result": None if adapter_result is None else str(adapter_result),
        "compatibility_result": result.value,
        "limitations": limitations,
        "assertions": {
            "deny_hook_was_dispatched": hook_called,
            "tool_body_not_entered": len(entries) == 0,
            "no_independent_side_effect": len(entries) == 0,
        },
        "candidate_digest_probe": candidate_digest(
            {"tool": "naeos_v2_langgraph_probe", "decision": "DENY"}
        ),
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
    raise SystemExit(
        0 if report["compatibility_result"] == "PASS"
        else 2 if report["compatibility_result"] == "FAIL"
        else 3
    )
