// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { strict as assert } from "node:assert";
import { test } from "node:test";

import {
  executeLinkedInPublication,
  linkedinExecutionKey,
  type LinkedInExecutionReceiptStore,
} from "./linkedin-runtime.ts";
import type { ExternalProviderReceipt } from "./external-publisher.ts";

const base = {
  executionId: "execution-runtime-001",
  provider: "linkedin",
  authorization: { contractVersion: "authorization-v1", authorized: true },
  action: {
    proposalId: "proposal-runtime-001",
    action: "publish-social-draft",
    target: { system: "social", resource: "linkedin:urn:li:organization:12345" },
    policyVersion: "social-v1",
    decisionId: "decision-runtime-001",
  },
  authorUrn: "urn:li:organization:12345",
  commentary: "NAEOS runtime boundary test.",
};

function store(seed?: [string, ExternalProviderReceipt]): LinkedInExecutionReceiptStore {
  const entries = new Map<string, ExternalProviderReceipt>();
  const claims = new Set<string>();
  if (seed) entries.set(seed[0], seed[1]);
  return {
    async claim(key) {
      const existing = entries.get(key);
      if (existing) return { status: "existing", receipt: existing };
      if (claims.has(key)) return { status: "in-flight" };
      claims.add(key);
      return { status: "acquired" };
    },
    async put(key, receipt) { entries.set(key, receipt); claims.delete(key); },
    async release(key) { claims.delete(key); },
  };
}

function dependencies(storeImpl: LinkedInExecutionReceiptStore) {
  let calls = 0;
  return {
    receiptStore: storeImpl,
    credentials: { async getAccessToken() { return "test-token"; } },
    revalidateAuthorization: async () => true,
    fetchImpl: async () => {
      calls += 1;
      return new Response(null, { status: 201, headers: { "x-restli-id": "urn:li:share:runtime-001" } });
    },
    get calls() { return calls; },
  };
}

test("runtime execution records the provider observation", async () => {
  const deps = dependencies(store());
  const receipt = await executeLinkedInPublication(base, deps);
  assert.equal(receipt.status, "accepted");
  assert.equal(receipt.receiptId, "urn:li:share:runtime-001");
  assert.equal(deps.calls, 1);
  const replay = await executeLinkedInPublication(base, deps);
  assert.deepEqual(replay, receipt);
  assert.equal(deps.calls, 1);
});

test("runtime replays an existing receipt without provider execution", async () => {
  const existing: ExternalProviderReceipt = {
    provider: "linkedin",
    receiptId: "urn:li:share:existing-001",
    status: "accepted",
    observedAt: "2026-09-26T10:00:00.000Z",
    externalId: "urn:li:share:existing-001",
  };
  const key = linkedinExecutionKey(base);
  const deps = dependencies(store([key, existing]));
  const receipt = await executeLinkedInPublication(base, deps);
  assert.deepEqual(receipt, existing);
  assert.equal(deps.calls, 0);
});

test("runtime prevents concurrent provider execution", async () => {
  const receiptStore = store();
  let resolveFetch!: (response: Response) => void;
  const fetchResponse = new Promise<Response>((resolve) => { resolveFetch = resolve; });
  let calls = 0;
  const deps = {
    receiptStore,
    credentials: { async getAccessToken() { return "test-token"; } },
    revalidateAuthorization: async () => true,
    fetchImpl: async () => { calls += 1; return fetchResponse; },
  };
  const first = executeLinkedInPublication(base, deps);
  await new Promise((resolve) => setImmediate(resolve));
  await assert.rejects(executeLinkedInPublication(base, deps), /execution already in progress/);
  assert.equal(calls, 1);
  resolveFetch(new Response(null, { status: 201, headers: { "x-restli-id": "urn:li:share:concurrent-001" } }));
  const receipt = await first;
  assert.equal(receipt.receiptId, "urn:li:share:concurrent-001");
  assert.equal(calls, 1);
});

test("execution key binds identity and governance context without delimiter ambiguity", () => {
  const same = linkedinExecutionKey(base);
  const differentDecision = linkedinExecutionKey({ ...base, action: { ...base.action, decisionId: "decision-runtime-002" } });
  const colonVariant = linkedinExecutionKey({ ...base, action: { ...base.action, decisionId: "decision:runtime-001" } });
  assert.notEqual(same, differentDecision);
  assert.notEqual(same, colonVariant);
});
