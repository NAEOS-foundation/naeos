// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { strict as assert } from "node:assert";
import { test } from "node:test";

import {
  LINKEDIN_API_VERSION,
  LINKEDIN_POSTS_ENDPOINT,
  publishLinkedInPost,
} from "./linkedin-publisher.ts";

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
      resource: "linkedin:urn:li:organization:12345",
    },
    policyVersion: "social-v1",
    decisionId: "decision-001",
  },
  authorUrn: "urn:li:organization:12345",
  commentary: "NAEOS engineering signal.",
};

function dependencies(
  response: Response,
  options: { revalidated?: boolean; token?: string } = {},
) {
  let calls = 0;
  return {
    get calls() {
      return calls;
    },
    credentials: {
      async getAccessToken() {
        return options.token ?? "test-token";
      },
    },
    revalidateAuthorization: async () => options.revalidated ?? true,
    fetchImpl: async (input: RequestInfo | URL, init?: RequestInit) => {
      calls += 1;
      assert.equal(String(input), LINKEDIN_POSTS_ENDPOINT);
      assert.equal(init?.method, "POST");
      assert.equal(
        new Headers(init?.headers).get("Linkedin-Version"),
        LINKEDIN_API_VERSION,
      );
      assert.equal(
        new Headers(init?.headers).get("X-Restli-Protocol-Version"),
        "2.0.0",
      );
      assert.equal(new Headers(init?.headers).get("Authorization"), "Bearer test-token");
      return response;
    },
  };
}

test("authorized LinkedIn publication maps provider receipt", async () => {
  const deps = dependencies(
    new Response(null, {
      status: 201,
      headers: { "x-restli-id": "urn:li:share:123" },
    }),
  );

  const receipt = await publishLinkedInPost(base, deps);

  assert.equal(receipt.provider, "linkedin");
  assert.equal(receipt.status, "accepted");
  assert.equal(receipt.receiptId, "urn:li:share:123");
  assert.equal(receipt.externalId, "urn:li:share:123");
  assert.equal(deps.calls, 1);
});

test("last-boundary revalidation denies provider call", async () => {
  const deps = dependencies(
    new Response(null, { status: 201 }),
    { revalidated: false },
  );

  const receipt = await publishLinkedInPost(base, deps);

  assert.equal(receipt.status, "rejected");
  assert.equal(deps.calls, 0);
});

test("missing credential denies provider call", async () => {
  const deps = dependencies(
    new Response(null, { status: 201 }),
    { token: "" },
  );

  const receipt = await publishLinkedInPost(base, deps);

  assert.equal(receipt.status, "rejected");
  assert.equal(deps.calls, 0);
});

test("invalid organization identity denies provider call", async () => {
  const deps = dependencies(new Response(null, { status: 201 }));

  const receipt = await publishLinkedInPost(
    { ...base, authorUrn: "urn:li:member:12345" },
    deps,
  );

  assert.equal(receipt.status, "rejected");
  assert.equal(deps.calls, 0);
});

test("LinkedIn rejection becomes provider observation", async () => {
  const deps = dependencies(
    new Response(null, {
      status: 403,
      headers: { "x-restli-id": "rejected-request-001" },
    }),
  );

  const receipt = await publishLinkedInPost(base, deps);

  assert.equal(receipt.status, "rejected");
  assert.equal(receipt.receiptId, "rejected-request-001");
  assert.equal(deps.calls, 1);
});
