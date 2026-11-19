# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Unit tests for v2 evidence primitives; not framework integration tests."""
import unittest

from scripts.crewai_compat_v2_harness import (
    Result, SideEffectLedger, candidate_digest, classify_deny,
    run_primitive_self_test,
)


class HarnessPrimitiveTests(unittest.TestCase):
    def test_ledger_is_independent_and_append_only(self):
        ledger = SideEffectLedger()
        first = ledger.append(
            invocation_id="inv-1", tool_name="mock_write",
            payload={"record": "one"}, action_id="action-1",
        )
        self.assertEqual(ledger.count(), 1)
        self.assertEqual(ledger.count("inv-1"), 1)
        self.assertEqual(ledger.count("inv-2"), 0)
        self.assertEqual(ledger.snapshot()[0], first)
        self.assertTrue(ledger.checksum().startswith("sha256:"))

    def test_candidate_digest_is_order_independent_for_object_keys(self):
        self.assertEqual(
            candidate_digest({"tool": "read", "input": {"a": 1, "b": 2}}),
            candidate_digest({"input": {"b": 2, "a": 1}, "tool": "read"}),
        )

    def test_candidate_digest_binds_mutated_input(self):
        self.assertNotEqual(
            candidate_digest({"tool": "read", "input": {"id": "A"}}),
            candidate_digest({"tool": "read", "input": {"id": "B"}}),
        )

    def test_deny_with_side_effect_is_fail(self):
        self.assertEqual(
            classify_deny(enforcement_observed=True, tool_body_entries=1,
                          side_effect_count=1, path_exercised=True),
            Result.FAIL,
        )

    def test_unexercised_path_is_blocked(self):
        self.assertEqual(
            classify_deny(enforcement_observed=False, tool_body_entries=0,
                          side_effect_count=0, path_exercised=False),
            Result.BLOCKED,
        )

    def test_missing_enforcement_is_unknown_not_pass(self):
        self.assertEqual(
            classify_deny(enforcement_observed=False, tool_body_entries=0,
                          side_effect_count=0, path_exercised=True),
            Result.UNKNOWN,
        )

    def test_primitive_self_test(self):
        self.assertEqual(run_primitive_self_test()["status"], "PASS")


if __name__ == "__main__":
    unittest.main()
