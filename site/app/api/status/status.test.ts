// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import test from "node:test";

import { getLiveStatus, parseContributorCount, probeHttp } from "./status.ts";

function response(status: number, body: unknown = {}) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

test("HTTP probe treats 2xx-4xx as reachable and 5xx as down", async () => {
  const ok = await probeHttp("https://example.test", async () => response(404));
  const down = await probeHttp("https://example.test", async () => response(503));

  assert.equal(ok.ok, true);
  assert.equal(down.ok, false);
});

test("HTTP probe fails closed on fetch errors", async () => {
  const result = await probeHttp("https://example.test", async () => {
    throw new Error("network failure");
  });

  assert.equal(result.ok, false);
});

test("contributor pagination parser reads the last page", () => {
  assert.equal(
    parseContributorCount('<https://api.github.com/repos/x/contributors?page=10>; rel="last"'),
    10,
  );
  assert.equal(parseContributorCount(null), undefined);
});

test("live status produces a complete runtime payload", async () => {
  const fetcher = async (input: RequestInfo | URL) => {
    const url = String(input);

    if (url === "https://api.github.com/repos/NAEOS-foundation/naeos") {
      return response(200, {
        stargazers_count: 11,
        forks_count: 6,
        open_issues_count: 34,
      });
    }

    if (url === "https://api.github.com/repos/NAEOS-foundation/naeos/releases/latest") {
      return response(200, { tag_name: "v3.6.0" });
    }

    return response(200);
  };

  const payload = await getLiveStatus(fetcher);

  assert.equal(payload.source, "live");
  assert.equal(payload.github.stars, 11);
  assert.equal(payload.github.version, "3.6.0");
  assert.equal(payload.services.length, 4);
  assert.ok(payload.updatedAt);
});
