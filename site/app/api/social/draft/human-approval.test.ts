// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import test from "node:test";
import assert from "node:assert/strict";
// Node's native TypeScript runner requires the explicit .ts extension.
// @ts-ignore TS5097: runtime resolution is handled by Node's strip-types loader.
import { evaluateHumanApproval } from "./human-approval.ts";

const base = {
  approvalId: "approval-1",
  approverId: "maintainer-1",
  approverIdentity: "human-reviewer",
  decision: "approve" as const,
  approvedAt: "2026-09-26T00:00:00Z",
  proposalId: "proposal-1",
  action: "publish-social-draft",
  target: { system: "social", resource: "linkedin" },
  policyVersion: "social-v1",
  decisionId: "decision-1",
};

test("complete explicit human approval is accepted as boundary evidence", () => {
  const result = evaluateHumanApproval(base);

  assert.equal(result.accepted, true);
  assert.deepEqual(result.reasons, []);
  assert.equal(result.authenticated, false);
});

test("missing approval identity fails closed", () => {
  const result = evaluateHumanApproval({ ...base, approverIdentity: "   " });

  assert.equal(result.accepted, false);
  assert.ok(result.reasons.includes("approver-identity-missing"));
});

test("missing proposal binding fails closed", () => {
  const result = evaluateHumanApproval({ ...base, proposalId: "" });

  assert.equal(result.accepted, false);
  assert.ok(result.reasons.includes("proposal-id-missing"));
});

test("missing policy decision binding fails closed", () => {
  const result = evaluateHumanApproval({ ...base, decisionId: "" });

  assert.equal(result.accepted, false);
  assert.ok(result.reasons.includes("decision-id-missing"));
});

test("rejection never becomes accepted", () => {
  const result = evaluateHumanApproval({ ...base, decision: "reject" });

  assert.equal(result.accepted, false);
  assert.ok(result.reasons.includes("human-rejected"));
});

test("blank approval fields fail closed", () => {
  const result = evaluateHumanApproval({
    ...base,
    approvalId: " ",
    approverId: " ",
    action: " ",
    target: { system: " ", resource: " " },
    policyVersion: " ",
  });

  assert.equal(result.accepted, false);
  assert.ok(result.reasons.includes("approval-id-missing"));
  assert.ok(result.reasons.includes("approver-id-missing"));
  assert.ok(result.reasons.includes("action-missing"));
  assert.ok(result.reasons.includes("target-system-missing"));
  assert.ok(result.reasons.includes("target-resource-missing"));
  assert.ok(result.reasons.includes("policy-version-missing"));
});
