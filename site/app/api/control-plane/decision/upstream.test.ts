// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import test from "node:test";
import { resolveControlPlaneUpstream } from "./upstream.ts";

test("requires an explicitly configured HTTPS origin", () => {
  assert.equal(resolveControlPlaneUpstream(undefined), null);
  assert.equal(resolveControlPlaneUpstream(""), null);
  assert.equal(resolveControlPlaneUpstream("   "), null);
  assert.equal(resolveControlPlaneUpstream("not a URL"), null);
  assert.equal(resolveControlPlaneUpstream("http://control-plane.example"), null);
});

test("rejects URL credentials and non-origin components", () => {
  assert.equal(resolveControlPlaneUpstream("https://user:pass@example.com"), null);
  assert.equal(resolveControlPlaneUpstream("https://example.com/api"), null);
  assert.equal(resolveControlPlaneUpstream("https://example.com/?debug=true"), null);
  assert.equal(resolveControlPlaneUpstream("https://example.com/#fragment"), null);
});

test("normalizes an explicitly configured HTTPS origin", () => {
  assert.equal(
    resolveControlPlaneUpstream("  https://staging-control-plane.example/  "),
    "https://staging-control-plane.example",
  );
});
