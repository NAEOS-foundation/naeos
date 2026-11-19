#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Exercise CrewAI's public Crew.kickoff ReAct path under a denying hook."""
from __future__ import annotations

import argparse
import json
import platform
import sys
from pathlib import Path
from typing import Any

import crewai
from pydantic import PrivateAttr

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from crewai import Agent, Crew, Process, Task
from crewai.hooks.tool_hooks import clear_all_tool_call_hooks, register_before_tool_call_hook
from crewai.llms.base_llm import BaseLLM
from crewai.tools import BaseTool

from scripts.crewai_compat_v2_harness import SideEffectLedger, candidate_digest, classify_deny


LEDGER = SideEffectLedger()
TOOL_BODY_ENTRIES = 0
HOOK_OBSERVATIONS: list[dict[str, Any]] = []


class SideEffectTool(BaseTool):
    name: str = "naeos_v2_react_probe"
    description: str = "Records a local test side effect."

    def _run(self, invocation_id: str, value: str) -> str:
        global TOOL_BODY_ENTRIES
        TOOL_BODY_ENTRIES += 1
        LEDGER.append(
            invocation_id=invocation_id,
            tool_name=self.name,
            payload={"value": value},
        )
        return "tool-body-executed"


class DeterministicReActLLM(BaseLLM):
    _calls: int = PrivateAttr(default=0)

    def call(
        self,
        messages: Any,
        tools: Any = None,
        callbacks: Any = None,
        available_functions: Any = None,
        from_task: Any = None,
        from_agent: Any = None,
        response_model: Any = None,
    ) -> str:
        self._calls += 1
        if self._calls == 1:
            return (
                "Thought: I will call the controlled test tool.\n"
                "Action: naeos_v2_react_probe\n"
                'Action Input: {"invocation_id":"react-deny-001","value":"probe"}'
            )
        return "Thought: The tool attempt has been processed.\nFinal Answer: policy test completed"

    @property
    def call_count(self) -> int:
        return self._calls


def run_probe() -> dict[str, Any]:
    global LEDGER, TOOL_BODY_ENTRIES
    LEDGER = SideEffectLedger()
    TOOL_BODY_ENTRIES = 0
    HOOK_OBSERVATIONS.clear()
    clear_all_tool_call_hooks()

    def deny_every_tool(context: Any) -> bool:
        HOOK_OBSERVATIONS.append(
            {"tool_name": context.tool_name, "decision": "DENY"}
        )
        return False

    register_before_tool_call_hook(deny_every_tool)
    llm = DeterministicReActLLM(model="naeos-deterministic-react-probe")
    execution_error = None
    crew_result = None
    try:
        agent = Agent(
            role="Policy test agent",
            goal="Exercise exactly one controlled ReAct tool call.",
            backstory="A deterministic integration-test agent.",
            llm=llm,
            tools=[SideEffectTool()],
            allow_delegation=False,
            max_iter=3,
            max_retry_limit=0,
            max_execution_time=30,
            respect_context_window=False,
            verbose=False,
        )
        task = Task(
            description="Attempt the controlled tool call and then provide a concise result.",
            expected_output="A concise statement that the policy test completed.",
            agent=agent,
        )
        crew = Crew(
            agents=[agent],
            tasks=[task],
            process=Process.sequential,
            verbose=False,
        )
        crew_result = str(crew.kickoff())
    except Exception as exc:
        execution_error = f"{type(exc).__name__}: {exc}"

    entries = LEDGER.snapshot()
    body_entries = TOOL_BODY_ENTRIES
    hook_called = bool(HOOK_OBSERVATIONS)
    # A second LLM turn means the ReAct executor advanced beyond the initial
    # tool-call response; hook or body observations are stronger evidence.
    path_exercised = hook_called or body_entries > 0 or llm.call_count >= 2
    result = classify_deny(
        enforcement_observed=hook_called,
        tool_body_entries=body_entries,
        side_effect_count=len(entries),
        path_exercised=path_exercised,
    )
    limitations = []
    if not path_exercised:
        limitations.append("The public Crew.kickoff ReAct path did not reach an observable tool boundary.")
    if execution_error:
        limitations.append(f"Crew kickoff raised: {execution_error}")

    report = {
        "schema": "naeos.crewai-compatibility-v2.react/0.1",
        "framework": "crewai",
        "framework_version": crewai.__version__,
        "python_version": platform.python_version(),
        "execution_path": "public Crew.kickoff -> AgentExecutor/CrewAgentExecutor ReAct loop -> execute_tool_and_check_finality",
        "decision": "DENY",
        "llm_calls": llm.call_count,
        "execution_error": execution_error,
        "path_exercised": path_exercised,
        "hook_dispatch_observed": hook_called,
        "hook_observations": HOOK_OBSERVATIONS,
        "tool_body_entries": body_entries,
        "side_effect_count": len(entries),
        "side_effect_ledger_checksum": LEDGER.checksum(),
        "crew_result": crew_result,
        "compatibility_result": result.value,
        "limitations": limitations,
        "assertions": {
            "deny_hook_was_dispatched": hook_called,
            "tool_body_not_entered": body_entries == 0,
            "no_independent_side_effect": len(entries) == 0,
        },
        "candidate_digest_probe": candidate_digest(
            {"tool": "naeos_v2_react_probe", "decision": "DENY"}
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
