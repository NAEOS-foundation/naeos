# Copyright 2026 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Unit tests for v2 evidence primitives; not framework integration tests."""
import unittest
import threading

from scripts.crewai_compat_v2_harness import (
    ActionState, DecisionGate, DecisionGateError, ExecutionLedger, Result, SideEffectLedger, candidate_digest,
    classify_deny, run_primitive_self_test,
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


class DecisionLifecycleTests(unittest.TestCase):
    def setUp(self):
        self.now = 1000.0
        self.policy_version = "policy-v1"
        self.policy_available = True
        def policy_provider():
            if not self.policy_available:
                raise RuntimeError("simulated policy service outage")
            return self.policy_version
        self.gate = DecisionGate(policy_version_provider=policy_provider, clock=lambda: self.now)
        self.candidate = {
            "framework": "crewai", "path": "Flow.ToolAction",
            "tool_name": "mock_write", "action": "write",
            "input": {"record_id": "r-1", "value": "test"},
            "subject": "agent-1", "scope": "tenant-a/resource-r-1",
        }

    def issue(self, decision="PROCEED", expiry=1100.0):
        return self.gate.issue(candidate=self.candidate, policy_id="policy.main",
            subject="agent-1", scope="tenant-a/resource-r-1", expires_at=expiry,
            decision=decision, decision_id=f"decision-{decision.lower()}")

    def test_v2_01_fresh_proceed_authorizes_exact_candidate(self):
        record = self.issue()
        authorized = self.gate.authorize(decision_id=record.decision_id,
            candidate=self.candidate, subject="agent-1", scope="tenant-a/resource-r-1")
        self.assertEqual(authorized.decision, "PROCEED")
        self.assertEqual(authorized.candidate_digest, candidate_digest(self.candidate))

    def test_v2_02_candidate_mutation_is_rejected(self):
        record = self.issue()
        mutated = {**self.candidate, "input": {"record_id": "r-1", "value": "mutated"}}
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=mutated,
                subject="agent-1", scope="tenant-a/resource-r-1")

    def test_v2_03_tool_identity_mutation_is_rejected(self):
        record = self.issue()
        changed_tool = {**self.candidate, "tool_name": "other_tool"}
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=changed_tool,
                subject="agent-1", scope="tenant-a/resource-r-1")

    def test_v2_04_policy_version_change_invalidates_decision(self):
        record = self.issue()
        self.policy_version = "policy-v2"
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=self.candidate,
                subject="agent-1", scope="tenant-a/resource-r-1")

    def test_decision_expiry_rejects_nan_infinity_and_bool(self):
        for expiry in (float("nan"), float("inf"), float("-inf"), True):
            with self.subTest(expiry=expiry):
                with self.assertRaises(DecisionGateError):
                    self.issue(expiry=expiry)

    def test_v2_05_expired_decision_is_rejected(self):
        record = self.issue(expiry=1001.0)
        self.now = 1001.0
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=self.candidate,
                subject="agent-1", scope="tenant-a/resource-r-1")

    def test_v2_06_policy_provider_unavailable_fails_closed(self):
        record = self.issue()
        self.policy_available = False
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=self.candidate,
                subject="agent-1", scope="tenant-a/resource-r-1")
        with self.assertRaises(DecisionGateError):
            self.issue()

    def test_v2_07_subject_and_scope_mismatch_rejected(self):
        record = self.issue()
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=self.candidate,
                subject="agent-2", scope="tenant-a/resource-r-1")
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=self.candidate,
                subject="agent-1", scope="tenant-b/resource-r-1")

    def test_concurrent_authorize_consumes_decision_exactly_once(self):
        record = self.issue()
        barrier = threading.Barrier(8)
        accepted = []
        rejected = []

        def attempt():
            barrier.wait(timeout=5)
            try:
                self.gate.authorize(decision_id=record.decision_id,
                    candidate=self.candidate, subject="agent-1",
                    scope="tenant-a/resource-r-1")
                accepted.append(True)
            except DecisionGateError:
                rejected.append(True)

        threads = [threading.Thread(target=attempt) for _ in range(8)]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join(timeout=5)
        self.assertTrue(all(not thread.is_alive() for thread in threads))
        self.assertEqual(len(accepted), 1)
        self.assertEqual(len(rejected), 7)

    def test_v2_08_single_use_decision_replay_rejected(self):
        record = self.issue()
        args = dict(decision_id=record.decision_id, candidate=self.candidate,
            subject="agent-1", scope="tenant-a/resource-r-1")
        self.gate.authorize(**args)
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(**args)

    def test_v2_09_needs_review_requires_exact_explicit_approval(self):
        record = self.issue(decision="NEEDS_REVIEW")
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=self.candidate,
                subject="agent-1", scope="tenant-a/resource-r-1")
        self.gate.approve(decision_id=record.decision_id, candidate=self.candidate)
        authorized = self.gate.authorize(decision_id=record.decision_id, candidate=self.candidate,
            subject="agent-1", scope="tenant-a/resource-r-1")
        self.assertEqual(authorized.decision, "NEEDS_REVIEW")

    def test_v2_10_approval_candidate_substitution_rejected(self):
        record = self.issue(decision="NEEDS_REVIEW")
        substituted = {**self.candidate, "input": {"record_id": "r-1", "value": "different"}}
        with self.assertRaises(DecisionGateError):
            self.gate.approve(decision_id=record.decision_id, candidate=substituted)
        self.gate.approve(decision_id=record.decision_id, candidate=self.candidate)
        with self.assertRaises(DecisionGateError):
            self.gate.authorize(decision_id=record.decision_id, candidate=substituted,
                subject="agent-1", scope="tenant-a/resource-r-1")


class ResponseLossAndReconciliationTests(unittest.TestCase):
    def setUp(self):
        self.candidate = {
            "tool_name": "mock_write", "input": {"record": "r-9", "value": "committed"},
        }
        self.action_id = "action-v2-11-stable"
        self.invocation_id = "invocation-v2-11"
        self.exec_ledger = ExecutionLedger()
        self.effects = SideEffectLedger()
        self.exec_ledger.prepare(action_id=self.action_id, invocation_id=self.invocation_id,
            tool_name="mock_write", candidate=self.candidate)
        self.exec_ledger.mark_dispatched(action_id=self.action_id)

    def simulate_commit_then_response_loss(self):
        # Independent effect ledger records that the tool body committed.
        self.effects.append(action_id=self.action_id, invocation_id=self.invocation_id,
            tool_name="mock_write", payload=self.candidate["input"])
        # Simulate transport timeout after commit: caller cannot observe response.
        return self.exec_ledger.mark_unknown(action_id=self.action_id,
            reason="response lost after tool commit; client timeout")

    def test_v2_11_response_loss_is_unknown_and_retry_is_forbidden(self):
        unknown = self.simulate_commit_then_response_loss()
        self.assertEqual(unknown.state, ActionState.UNKNOWN)
        self.assertEqual(self.effects.count(self.invocation_id), 1)
        self.assertFalse(self.exec_ledger.retry_allowed(action_id=self.action_id))
        self.assertEqual(self.exec_ledger.get(self.action_id).state, ActionState.UNKNOWN)

    def test_v2_12_authoritative_reconciliation_finds_single_commit(self):
        self.simulate_commit_then_response_loss()
        reconciled = self.exec_ledger.reconcile(action_id=self.action_id,
            authoritative_effects=self.effects.snapshot())
        self.assertEqual(reconciled.state, ActionState.SUCCEEDED)
        self.assertEqual(reconciled.authoritative_effect_count, 1)
        self.assertFalse(self.exec_ledger.retry_allowed(action_id=self.action_id))

    def test_v2_12_absence_is_not_misreported_as_failure(self):
        self.simulate_commit_then_response_loss()
        reconciled = self.exec_ledger.reconcile(action_id=self.action_id,
            authoritative_effects=())
        self.assertEqual(reconciled.state, ActionState.UNKNOWN)
        self.assertIn("absence is not proof", reconciled.reconciliation_note)
        self.assertFalse(self.exec_ledger.retry_allowed(action_id=self.action_id))

    def test_v2_12_duplicate_effect_matches_remain_ambiguous_unknown(self):
        self.simulate_commit_then_response_loss()
        original = self.effects.snapshot()[0]
        duplicate = self.effects.append(invocation_id=self.invocation_id,
            tool_name="mock_write", payload=self.candidate["input"], action_id=self.action_id)
        reconciled = self.exec_ledger.reconcile(action_id=self.action_id,
            authoritative_effects=(original, duplicate))
        self.assertEqual(reconciled.state, ActionState.UNKNOWN)
        self.assertEqual(reconciled.authoritative_effect_count, 2)
        self.assertIn("ambiguous", reconciled.reconciliation_note)

    def test_v2_12_mismatched_effect_identity_cannot_confirm_success(self):
        self.simulate_commit_then_response_loss()
        wrong = self.effects.append(invocation_id="other-invocation",
            tool_name="different_tool", payload={"different": "payload"},
            action_id=self.action_id)
        reconciled = self.exec_ledger.reconcile(action_id=self.action_id,
            authoritative_effects=(wrong,))
        self.assertEqual(reconciled.state, ActionState.UNKNOWN)
        self.assertEqual(reconciled.authoritative_effect_count, 0)
        self.assertIn("absence is not proof", reconciled.reconciliation_note)

    def test_concurrent_dispatch_transition_is_single_winner(self):
        fresh = ExecutionLedger()
        fresh.prepare(action_id="race", invocation_id="race-inv",
            tool_name="write", candidate=self.candidate)
        barrier = threading.Barrier(8)
        successes = []
        failures = []

        def dispatch():
            barrier.wait(timeout=5)
            try:
                fresh.mark_dispatched(action_id="race")
                successes.append(True)
            except ValueError:
                failures.append(True)

        threads = [threading.Thread(target=dispatch) for _ in range(8)]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join(timeout=5)
        self.assertTrue(all(not thread.is_alive() for thread in threads))
        self.assertEqual(len(successes), 1)
        self.assertEqual(len(failures), 7)
        self.assertEqual(fresh.get("race").state, ActionState.DISPATCHED)

    def test_dispatch_state_transition_guards(self):
        fresh = ExecutionLedger()
        fresh.prepare(action_id="fresh", invocation_id="inv", tool_name="write",
            candidate=self.candidate)
        self.assertTrue(fresh.retry_allowed(action_id="fresh"))
        fresh.mark_dispatched(action_id="fresh")
        self.assertFalse(fresh.retry_allowed(action_id="fresh"))
        with self.assertRaises(ValueError):
            fresh.mark_dispatched(action_id="fresh")


if __name__ == "__main__":
    unittest.main()
