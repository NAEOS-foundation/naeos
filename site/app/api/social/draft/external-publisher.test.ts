// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { strict as assert } from "node:assert";
import { test } from "node:test";

import {
  isProviderReceiptValid,
  prepareExternalPublish,
} from "./external-publisher.ts";

const base = {
  executionId: "execution-001",
  provider: "linkedin",
  authorization: {
    contractVersion: "authorization-v1",
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

test("authorized handoff is ready without creating an external effect", () => {
  const result = prepareExternalPublish(base);

  assert.equal(result.contractVersion, "external-publisher-v1");
  assert.equal(result.ready, true);
  assert.equal(result.externalEffect, false);
  assert.deepEqual(result.reasons, []);
});

test("unauthorized handoff fails closed", () => {
  const result = prepareExternalPublish({
    ...base,
    authorization: { ...base.authorization, authorized: false },
  });

  assert.equal(result.ready, false);
  assert.ok(result.reasons.includes("authorization-required"));
  assert.equal(result.externalEffect, false);
});

test("unsupported authorization version fails closed", () => {
  const result = prepareExternalPublish({
    ...base,
    authorization: { contractVersion: "authorization-v0", authorized: true },
  });

  assert.equal(result.ready, false);
  assert.ok(result.reasons.includes("authorization-contract-version-unsupported"));
});

test("missing execution identity fails closed", () => {
  const result = prepareExternalPublish({
    ...base,
    executionId: " ",
  });

  assert.equal(result.ready, false);
  assert.ok(result.reasons.includes("execution-id-missing"));
});

test("missing provider fails closed", () => {
  const result = prepareExternalPublish({
    ...base,
    provider: " ",
  });

  assert.equal(result.ready, false);
  assert.ok(result.reasons.includes("provider-missing"));
});

test("missing action bindings fail closed", () => {
  const result = prepareExternalPublish({
    ...base,
    action: {
      ...base.action,
      proposalId: " ",
      target: { ...base.action.target, resource: "" },
      decisionId: "",
    },
  });

  assert.equal(result.ready, false);
  assert.ok(result.reasons.includes("proposal-id-missing"));
  assert.ok(result.reasons.includes("target-resource-missing"));
  assert.ok(result.reasons.includes("decision-id-missing"));
});

test("provider receipt is accepted only when it matches the expected provider", () => {
  const receipt = {
    provider: "linkedin",
    receiptId: "provider-receipt-001",
    status: "accepted" as const,
    observedAt: "2026-09-26T10:00:00Z",
    externalId: "provider-object-001",
  };

  assert.equal(isProviderReceiptValid(receipt, "linkedin"), true);
  assert.equal(isProviderReceiptValid(receipt, "x"), false);
});

test("unknown provider receipt is not execution evidence", () => {
  const receipt = {
    provider: "linkedin",
    receiptId: "provider-receipt-002",
    status: "unknown" as const,
    observedAt: "2026-09-26T10:00:00Z",
  };

  assert.equal(isProviderReceiptValid(receipt, "linkedin"), false);
});
