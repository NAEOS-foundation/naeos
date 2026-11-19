#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Fail-closed authorization boundary for CrewAI BaseTool execution.

The wrapper is deliberately framework-adapter agnostic. It must be installed as
the tool object supplied to every execution path; global CrewAI hooks remain
defense-in-depth, not the authorization boundary.
"""
from __future__ import annotations

import asyncio
import inspect
from dataclasses import dataclass
from typing import Any, Callable, Mapping

from crewai.tools import BaseTool
from pydantic import PrivateAttr

from scripts.crewai_compat_v2_harness import candidate_digest


@dataclass(frozen=True)
class BoundaryDecision:
    tool_name: str
    decision: str
    arguments_digest: str
    policy_id: str


Authorizer = Callable[[str, Mapping[str, Any]], bool]
DecisionObserver = Callable[[BoundaryDecision], None]


class PolicyEnforcedTool(BaseTool):
    """BaseTool proxy that authorizes and observes before delegate execution.

    Both synchronous and asynchronous entry points are guarded. The evaluator
    must return the literal boolean True to allow execution. Exceptions,
    malformed decisions, and observer failures all fail closed.
    """

    _delegate: BaseTool = PrivateAttr()
    _authorize: Authorizer = PrivateAttr()
    _observe: DecisionObserver = PrivateAttr()
    _policy_id: str = PrivateAttr()

    def __init__(
        self,
        *,
        delegate: BaseTool,
        authorize: Authorizer,
        observe: DecisionObserver,
        policy_id: str = "naeos.default-deny-boundary/1",
    ) -> None:
        if not isinstance(delegate, BaseTool):
            raise TypeError("delegate must be a CrewAI BaseTool")
        if not callable(authorize) or not callable(observe):
            raise TypeError("authorize and observe callbacks are required")
        super().__init__(
            name=delegate.name,
            description=delegate.description,
            args_schema=delegate.args_schema,
        )
        self._delegate = delegate
        self._authorize = authorize
        self._observe = observe
        self._policy_id = policy_id

    def _decide(self, arguments: Mapping[str, Any]) -> None:
        digest = candidate_digest(dict(arguments))
        allowed = False
        try:
            decision = self._authorize(self.name, dict(arguments))
            if type(decision) is not bool:
                raise TypeError("policy evaluator must return literal bool")
            allowed = decision
        except Exception as exc:
            event = BoundaryDecision(
                tool_name=self.name,
                decision="POLICY_ERROR",
                arguments_digest=digest,
                policy_id=self._policy_id,
            )
            # Audit failure is itself fail-closed; preserve the original cause.
            self._observe(event)
            raise PermissionError(
                f"NAEOS policy evaluation failed closed for {self.name}"
            ) from exc

        event = BoundaryDecision(
            tool_name=self.name,
            decision="ALLOW" if allowed else "DENY",
            arguments_digest=digest,
            policy_id=self._policy_id,
        )
        # Do not execute unless the decision observation has been accepted.
        self._observe(event)
        if not allowed:
            raise PermissionError(f"NAEOS policy denied tool execution: {self.name}")

    def _run(self, **kwargs: Any) -> Any:
        self._decide(kwargs)
        return self._delegate.run(**kwargs)

    async def _arun(self, **kwargs: Any) -> Any:
        self._decide(kwargs)
        if type(self._delegate)._arun is BaseTool._arun:
            return await asyncio.to_thread(self._delegate.run, **kwargs)
        return await self._delegate.arun(**kwargs)


def normalize_langgraph_async_tool(tool: Any) -> Any:
    """Repair CrewAI's async StructuredTool conversion without changing its body.

    CrewAI 1.15.16 and 1.15.44 pass an async wrapper as StructuredTool.func.
    LangChain expects async callables in StructuredTool.coroutine; keeping the
    callable in func can return an unawaited coroutine and skip the tool body.
    """
    from langchain_core.tools import StructuredTool

    if (
        isinstance(tool, StructuredTool)
        and tool.coroutine is None
        and inspect.iscoroutinefunction(tool.func)
    ):
        return StructuredTool(
            name=tool.name,
            description=tool.description,
            coroutine=tool.func,
            args_schema=tool.args_schema,
        )
    return tool
