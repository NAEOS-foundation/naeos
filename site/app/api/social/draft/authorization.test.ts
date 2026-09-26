// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import test from "node:test";
import assert from "node:assert/strict";
// Node's native TypeScript runner requires the explicit .ts extension.
// @ts-ignore TS5097: runtime resolution is handled by Node's strip-types loader.
import { evaluateAuthorization } from "./authorization.ts";

const base = {
  proposalId: "proposal-1",
  action: "publish-social-draft",
  target: { system: "social", resource: "linkedin" },
  policyVersion: "social-v1",
  requiredPolicyVersion: "social-v1",
  policyDecision: "review_required" as const,
};

const approval = {
  approvalId: "approval-1",
  approvedBy: "human-reviewer",
  approvedAt: "2026-09-26T00:00:00Z",
  proposalId: "proposal-1",
  action: "publish-social-draft",
  target: { system: "social", resource: "linkedin" },
  policyVersion: "social-v1",
};

test("reviewable proposal without approval remains unauthorized", () => {
  const result = evaluateAuthorization(base);

  assert.equal(result.decision, "deny");
  assert.equal(result.authorized, false);
  assert.ok(result.reasons.includes("approval-missing"));
});

test("explicit approval authorizes a matching proposal", () => {
  const result = evaluateAuthorization({ ...base, approval });

  assert.equal(result.decision, "authorized");
  assert.equal(result.authorized, true);
  assert.deepEqual(result.reasons, []);
});

test("policy deny fails closed even when an approval exists", () => {
  const result = evaluateAuthorization({
    ...base,
    policyDecision: "deny",
    approval,
  });

  assert.equal(result.decision, "deny");
  assert.equal(result.authorized, false);
  assert.ok(result.reasons.includes("policy-denied"));
});

test("policy version mismatch fails closed", () => {
  const result = evaluateAuthorization({
    ...base,
    policyVersion: "social-v2",
    approval: { ...approval, policyVersion: "social-v2" },
  });

  assert.equal(result.decision, "deny");
  assert.equal(result.authorized, false);
  assert.ok(result.reasons.includes("policy-version-mismatch"));
});

test("approval cannot be reused for another proposal", () => {
  const result = evaluateAuthorization({
    ...base,
    proposalId: "proposal-2",
    approval,
  });

  assert.equal(result.decision, "deny");
  assert.equal(result.authorized, false);
  assert.ok(result.reasons.includes("approval-proposal-mismatch"));
});

test("approval target and action must match the proposal", () => {
  const result = evaluateAuthorization({
    ...base,
    action: "publish-x-draft",
    target: { system: "social", resource: "x" },
    approval,
  });

  assert.equal(result.decision, "deny");
  assert.equal(result.authorized, false);
  assert.ok(result.reasons.includes("approval-action-mismatch"));
  assert.ok(result.reasons.includes("approval-target-mismatch"));
});
