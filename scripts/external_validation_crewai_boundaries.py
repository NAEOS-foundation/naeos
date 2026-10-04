#!/usr/bin/env python3
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Real CrewAI hook-runtime validation for mutation and policy freshness.

This test exercises CrewAI's public before-tool-call hook dispatch and an
actual CrewAI tool body. It does not claim that CrewAI itself provides the
policy model; the policy state is supplied by the validation harness.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
from typing import Any

from crewai.tools import BaseTool
from pydantic import PrivateAttr
from crewai.hooks.tool_hooks import (
    ToolCallHookContext,
    clear_all_tool_call_hooks,
    register_before_tool_call_hook,
    run_before_tool_call_hooks,
)


def digest(value: Any) -> str:
    payload = json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()
    return hashlib.sha256(payload).hexdigest()


class SideEffectTool(BaseTool):
    name: str = "crm_query"
    description: str = "Records a controlled local side effect."
    _effects: list[dict[str, Any]] = PrivateAttr()

    def __init__(self, effects: list[dict[str, Any]]) -> None:
        super().__init__()
        self._effects = effects

    def _run(self, customer_id: str, fields: list[str]) -> str:
        self._effects.append({"customer_id": customer_id, "fields": fields})
        return "executed"


def invoke(tool: SideEffectTool, tool_input: dict[str, Any]) -> tuple[bool, str | None]:
    context = ToolCallHookContext(
        tool_name=tool.name,
        tool_input=tool_input,
        tool=tool,
    )
    blocked = run_before_tool_call_hooks(context)
    if blocked:
        return False, None
    return True, tool.run(**tool_input)


def scenario_mutation() -> dict[str, Any]:
    effects: list[dict[str, Any]] = []
    tool = SideEffectTool(effects)
    state = {"authorized_digest": None}

    def authorize(context: ToolCallHookContext) -> None:
        state["authorized_digest"] = digest(
            {"tool": context.tool_name, "input": context.tool_input, "policy_version": "v1"}
        )

    def mutate(context: ToolCallHookContext) -> None:
        context.tool_input["fields"] = ["name", "email"]

    def enforce_binding(context: ToolCallHookContext) -> bool | None:
        current = digest(
            {"tool": context.tool_name, "input": context.tool_input, "policy_version": "v1"}
        )
        if current != state["authorized_digest"]:
            return False
        return None

    register_before_tool_call_hook(authorize)
    register_before_tool_call_hook(mutate)
    register_before_tool_call_hook(enforce_binding)

    allowed, _ = invoke(tool, {"customer_id": "c-001", "fields": ["name", "status"]})
    clear_all_tool_call_hooks()

    if allowed or effects:
        raise AssertionError("mutated invocation must be blocked before tool body execution")

    return {
        "name": "mutation-after-authorization",
        "status": "PASS",
        "execution": "BLOCKED",
        "side_effect_count": len(effects),
        "reason": "canonical invocation digest changed before execution boundary",
    }


def scenario_freshness() -> dict[str, Any]:
    effects: list[dict[str, Any]] = []
    tool = SideEffectTool(effects)
    policy = {"current_version": "v1"}
    authorization = {"policy_version": "v1", "tool": tool.name}
    
    def enforce_freshness(context: ToolCallHookContext) -> bool | None:
        if authorization["policy_version"] != policy["current_version"]:
            return False
        return None

    register_before_tool_call_hook(enforce_freshness)

    # Authorization is issued under v1; policy changes before the consequential call.
    policy["current_version"] = "v2"
    allowed, _ = invoke(tool, {"customer_id": "c-002", "fields": ["name"]})
    clear_all_tool_call_hooks()

    if allowed or effects:
        raise AssertionError("stale v1 authorization must be blocked under v2 policy")

    return {
        "name": "policy-v1-to-v2-freshness",
        "status": "PASS",
        "authorized_policy_version": authorization["policy_version"],
        "current_policy_version": policy["current_version"],
        "execution": "BLOCKED_UNTIL_REAUTHORIZED",
        "side_effect_count": len(effects),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    scenarios = [scenario_mutation(), scenario_freshness()]
    report = {
        "schema": "naeos.external-validation-crewai-boundaries.v1",
        "runtime": "CrewAI public before_tool_call hook + BaseTool execution",
        "overall": "PASS",
        "scenarios": scenarios,
        "evidence_boundary": (
            "Real CrewAI hook dispatch and tool body execution. The policy model "
            "is supplied by this validation harness; no claim is made that CrewAI "
            "natively implements NAEOS policy semantics."
        ),
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
