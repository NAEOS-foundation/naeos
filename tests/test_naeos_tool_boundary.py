#!/usr/bin/env python3
# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Unit tests for the NAEOS final tool-execution authorization boundary."""
from __future__ import annotations

import asyncio
import unittest
from typing import Any, ClassVar

from crewai.tools import BaseTool

from scripts.naeos_tool_boundary import BoundaryDecision, GuardedToolRegistry, PolicyEnforcedTool, require_guarded_tools


class RecordingTool(BaseTool):
    name: str = "naeos_boundary_unit_probe"
    description: str = "Records whether the underlying tool body was entered."
    entries: ClassVar[list[dict[str, Any]]] = []

    def _run(self, value: str) -> str:
        type(self).entries.append({"value": value})
        return "executed"


class AlternateRecordingTool(RecordingTool):
    name: str = "naeos_boundary_alternate_probe"


class ToolBoundaryTests(unittest.TestCase):
    def setUp(self) -> None:
        RecordingTool.entries = []
        self.decisions: list[BoundaryDecision] = []
        self.delegate = RecordingTool()

    def guarded(self, authorize: Any, observe: Any = None) -> PolicyEnforcedTool:
        return PolicyEnforcedTool(
            delegate=self.delegate,
            authorize=authorize,
            observe=observe or self.decisions.append,
            policy_id="naeos.unit-test/1",
        )

    def test_deny_is_observed_before_delegate_and_has_no_side_effect(self) -> None:
        tool = self.guarded(lambda _name, _args: False)
        with self.assertRaises(PermissionError):
            tool.run(value="denied")
        self.assertEqual(RecordingTool.entries, [])
        self.assertEqual(len(self.decisions), 1)
        self.assertEqual(self.decisions[0].decision, "DENY")

    def test_candidate_digest_binds_tool_policy_and_arguments(self) -> None:
        first = PolicyEnforcedTool(
            delegate=RecordingTool(), authorize=lambda _name, _args: False,
            observe=self.decisions.append, policy_id="naeos.digest-policy-a/1",
        )
        other_policy = PolicyEnforcedTool(
            delegate=RecordingTool(), authorize=lambda _name, _args: False,
            observe=self.decisions.append, policy_id="naeos.digest-policy-b/1",
        )
        other_tool = PolicyEnforcedTool(
            delegate=AlternateRecordingTool(), authorize=lambda _name, _args: False,
            observe=self.decisions.append, policy_id="naeos.digest-policy-a/1",
        )
        changed_args = {"value": "different"}
        for guarded, args in ((first, {"value": "same"}), (other_policy, {"value": "same"}), (other_tool, {"value": "same"}), (first, changed_args)):
            with self.assertRaises(PermissionError):
                guarded.run(**args)
        digests = [decision.arguments_digest for decision in self.decisions]
        self.assertEqual(len(digests), 4)
        self.assertEqual(len(set(digests)), 4)
        self.assertEqual(RecordingTool.entries, [])

    def test_allow_is_observed_then_delegate_executes(self) -> None:
        tool = self.guarded(lambda _name, _args: True)
        self.assertEqual(tool.run(value="allowed"), "executed")
        self.assertEqual(RecordingTool.entries, [{"value": "allowed"}])
        self.assertEqual([item.decision for item in self.decisions], ["ALLOW"])

    def test_async_deny_is_observed_before_delegate(self) -> None:
        tool = self.guarded(lambda _name, _args: False)
        with self.assertRaises(PermissionError):
            asyncio.run(tool.arun(value="async-denied"))
        self.assertEqual(RecordingTool.entries, [])
        self.assertEqual([item.decision for item in self.decisions], ["DENY"])

    def test_async_allow_executes_sync_delegate_after_observation(self) -> None:
        tool = self.guarded(lambda _name, _args: True)
        self.assertEqual(asyncio.run(tool.arun(value="async-allowed")), "executed")
        self.assertEqual(RecordingTool.entries, [{"value": "async-allowed"}])
        self.assertEqual([item.decision for item in self.decisions], ["ALLOW"])

    def test_policy_exception_fails_closed(self) -> None:
        def broken_policy(_name: str, _args: Any) -> bool:
            raise RuntimeError("policy unavailable")

        tool = self.guarded(broken_policy)
        with self.assertRaises(PermissionError):
            tool.run(value="must-not-run")
        self.assertEqual(RecordingTool.entries, [])
        self.assertEqual([item.decision for item in self.decisions], ["POLICY_ERROR"])

    def test_non_boolean_policy_result_fails_closed(self) -> None:
        tool = self.guarded(lambda _name, _args: "ALLOW")
        with self.assertRaises(PermissionError):
            tool.run(value="must-not-run")
        self.assertEqual(RecordingTool.entries, [])
        self.assertEqual([item.decision for item in self.decisions], ["POLICY_ERROR"])

    def test_guarded_registry_wraps_raw_tools_before_registration(self) -> None:
        registry = GuardedToolRegistry()
        guarded = registry.register(self.delegate, authorize=lambda _name, _args: False, observe=self.decisions.append, policy_id="naeos.registry-test/1")
        self.assertIsInstance(guarded, PolicyEnforcedTool)
        self.assertEqual(require_guarded_tools(registry.tools()), [guarded])
        with self.assertRaises(PermissionError):
            registry.tools()[0].run(value="registry-denied")
        self.assertEqual(RecordingTool.entries, [])
        self.assertEqual([item.decision for item in self.decisions], ["DENY"])

    def test_guarded_registry_rejects_raw_tool_and_duplicate_name(self) -> None:
        registry = GuardedToolRegistry()
        with self.assertRaises(TypeError):
            registry.register_guarded(self.delegate)
        registry.register(self.delegate, authorize=lambda _name, _args: False, observe=self.decisions.append)
        with self.assertRaises(ValueError):
            registry.register(self.delegate, authorize=lambda _name, _args: False, observe=self.decisions.append)
        with self.assertRaises(TypeError):
            require_guarded_tools([self.delegate])

    def test_observer_failure_fails_closed(self) -> None:
        def broken_observer(_event: BoundaryDecision) -> None:
            raise OSError("audit sink unavailable")

        tool = self.guarded(lambda _name, _args: True, observe=broken_observer)
        with self.assertRaises(OSError):
            tool.run(value="must-not-run")
        self.assertEqual(RecordingTool.entries, [])


if __name__ == "__main__":
    unittest.main(verbosity=2)
