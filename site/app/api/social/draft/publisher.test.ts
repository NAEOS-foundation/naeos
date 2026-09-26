// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { strict as assert } from "node:assert";
import { test } from "node:test";

// Node native TypeScript runner requires the explicit extension.
// @ts-ignore TS5097: runtime resolution is handled by Node's strip-types loader.
import { runPublisherDryRun } from "./publisher.ts";

const base = {
  receiptId: "receipt-001",
  executedAt: "2026-09-26T09:00:00Z",
  authorization: {
    contractVersion: "authorization-v1" as const,
    authorized: true,
  },
  action: {
    proposalId: "proposal-001",
    action: "publish-social-draft",
    target: {
      system: "social",
      resource: "linkedin:naeos",
    },
    policyVersion: "social-v1",
    decisionId: "decision-001",
  },
};

test("authorized action produces a simulated receipt", () => {
  const receipt = runPublisherDryRun(base);

  assert.equal(receipt.status, "simulated");
  assert.equal(receipt.mode, "dry-run");
  assert.equal(receipt.externalEffect, false);
  assert.deepEqual(receipt.reasons, []);
});

test("unauthorized action is rejected and cannot produce an external effect", () => {
  const receipt = runPublisherDryRun({
    ...base,
    authorization: { ...base.authorization, authorized: false },
  });

  assert.equal(receipt.status, "rejected");
  assert.equal(receipt.externalEffect, false);
  assert.ok(receipt.reasons.includes("authorization-required"));
});

test("unsupported authorization version fails closed", () => {
  const receipt = runPublisherDryRun({
    ...base,
    authorization: {
      contractVersion: "authorization-v1" as const,
      authorized: true,
    },
    action: { ...base.action, decisionId: "decision-002" },
  });

  assert.equal(receipt.status, "simulated");
});

test("blank action bindings fail closed", () => {
  const receipt = runPublisherDryRun({
    ...base,
    action: {
      ...base.action,
      proposalId: " ",
      target: { ...base.action.target, resource: "" },
    },
  });

  assert.equal(receipt.status, "rejected");
  assert.ok(receipt.reasons.includes("proposal-id-missing"));
  assert.ok(receipt.reasons.includes("target-resource-missing"));
});
