#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Probe the CrewAI OpenAI Agents adapter with the NAEOS tool boundary."""
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
from agents import Agent, Runner
from agents.items import ModelResponse
from agents.models.interface import Model, ModelTracing
from agents.usage import Usage
from openai.types.responses import ResponseFunctionToolCall, ResponseOutputMessage, ResponseOutputText
from crewai.agents.agent_adapters.openai_agents.openai_agent_tool_adapter import OpenAIAgentToolAdapter
from crewai.hooks.tool_hooks import clear_all_tool_call_hooks, register_before_tool_call_hook
from crewai.tools import BaseTool

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from scripts.crewai_compat_v2_harness import SideEffectLedger, candidate_digest, classify_deny
from scripts.naeos_tool_boundary import BoundaryDecision, PolicyEnforcedTool

LEDGER = SideEffectLedger()
HOOK_OBSERVATIONS: list[dict[str, Any]] = []
BOUNDARY_OBSERVATIONS: list[BoundaryDecision] = []


class SideEffectTool(BaseTool):
    name: str = "naeos_v2_openai_adapter_probe"
    description: str = "Records a local test side effect."

    def _run(self, invocation_id: str, value: str) -> str:
        LEDGER.append(invocation_id=invocation_id, tool_name=self.name, payload={"value": value})
        return "tool-body-executed"


class DeterministicAgentsModel(Model):
    def __init__(self) -> None:
        self.calls = 0

    async def get_response(
        self,
        system_instructions: str | None,
        input: Any,
        model_settings: Any,
        tools: list[Any],
        output_schema: Any,
        handoffs: list[Any],
        tracing: ModelTracing,
        *,
        previous_response_id: str | None,
        conversation_id: str | None,
        prompt: Any,
    ) -> ModelResponse:
        self.calls += 1
        if self.calls == 1:
            output = [ResponseFunctionToolCall(
                id="openai-adapter-item-001",
                call_id="openai-adapter-call-001",
                type="function_call",
                name="naeos_v2_openai_adapter_probe",
                arguments=json.dumps({"invocation_id": "openai-adapter-deny-001", "value": "probe"}),
            )]
        else:
            output = [ResponseOutputMessage(
                id="openai-adapter-final-001",
                type="message",
                role="assistant",
                content=[ResponseOutputText(
                    type="output_text",
                    text="OpenAI Agents adapter policy probe completed.",
                    annotations=[],
                    logprobs=[],
                )],
                status="completed",
            )]
        return ModelResponse(
            output=output,
            usage=Usage(),
            response_id=f"naeos-openai-adapter-response-{self.calls}",
        )

    async def stream_response(self, *args: Any, **kwargs: Any) -> Any:
        raise NotImplementedError("Streaming is not used by this deterministic probe.")


async def invoke_adapter() -> tuple[Any, int]:
    guarded = PolicyEnforcedTool(
        delegate=SideEffectTool(),
        authorize=lambda _name, _args: False,
        observe=BOUNDARY_OBSERVATIONS.append,
        policy_id="naeos.test.deny-openai-agents/1",
    )
    adapter = OpenAIAgentToolAdapter(tools=[guarded])
    adapter.configure_tools([])
    model = DeterministicAgentsModel()
    agent = Agent(
        name="NAEOS adapter boundary probe",
        instructions="Call the provided test tool once, then finish.",
        model=model,
        tools=adapter.converted_tools,
    )
    result = await Runner.run(agent, "Exercise the controlled tool once.")
    return result.final_output, model.calls


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
    model_calls = 0
    try:
        adapter_result, model_calls = asyncio.run(invoke_adapter())
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
        limitations.append("OpenAI Agents FunctionTool did not reach the NAEOS boundary.")
    if execution_error and not enforcement_observed:
        limitations.append(f"Adapter invocation raised before enforcement: {execution_error}")

    report = {
        "schema": "naeos.crewai-compatibility-v2.openai-agents-adapter/0.2",
        "framework": "crewai",
        "framework_version": crewai.__version__,
        "adapter": "OpenAIAgentToolAdapter",
        "adapter_sdk": "openai-agents",
        "adapter_sdk_version": version("openai-agents"),
        "python_version": platform.python_version(),
        "execution_path": "OpenAIAgentToolAdapter -> OpenAI Agents Runner.run -> FunctionTool callback -> PolicyEnforcedTool._run",
        "decision": "DENY",
        "invocation_completed": invocation_completed,
        "model_calls": model_calls,
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
        "candidate_digest_probe": candidate_digest({"tool": "naeos_v2_openai_adapter_probe", "decision": "DENY"}),
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
