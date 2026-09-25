// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import test from "node:test";
// Node's native TypeScript runner requires the explicit .ts extension.
// @ts-ignore TS5097: runtime resolution is handled by Node's strip-types loader.
import { evaluateSocialPolicy } from "./policy.ts";

test("reviewable GitHub signals require human approval", () => {
  for (const signal of ["release", "merged-pr", "bug", "ci-failure"]) {
    const result = evaluateSocialPolicy({ source: "github-live", signals: [signal] });
    assert.equal(result.decision, "review_required");
    assert.equal(result.authorized, false);
    assert.deepEqual(result.reasons, ["human-approval-required"]);
    assert.equal(result.policyVersion, "social-v1");
  }
});

test("unsupported GitHub signals fail closed", () => {
  const result = evaluateSocialPolicy({
    source: "github-live",
    signals: ["active-pr", "unknown"],
  });

  assert.equal(result.decision, "deny");
  assert.equal(result.authorized, false);
  assert.deepEqual(result.reasons, ["no-reviewable-signal"]);
  assert.equal(result.policyVersion, "social-v1");
});

test("unavailable source fails closed even with a supported signal", () => {
  const result = evaluateSocialPolicy({
    source: "unavailable",
    signals: ["release"],
  });

  assert.equal(result.decision, "deny");
  assert.equal(result.authorized, false);
  assert.deepEqual(result.reasons, ["source-unavailable", "human-approval-required"]);
  assert.equal(result.policyVersion, "social-v1");
});

test("unavailable source with no signal records both fail-closed reasons", () => {
  const result = evaluateSocialPolicy({
    source: "unavailable",
    signals: [],
  });

  assert.equal(result.decision, "deny");
  assert.equal(result.authorized, false);
  assert.deepEqual(result.reasons, ["source-unavailable", "no-reviewable-signal"]);
  assert.equal(result.policyVersion, "social-v1");
});

test("mixed supported and unsupported signals remain reviewable", () => {
  const result = evaluateSocialPolicy({
    source: "github-live",
    signals: ["active-pr", "release", "unknown"],
  });

  assert.equal(result.decision, "review_required");
  assert.equal(result.authorized, false);
  assert.deepEqual(result.reasons, ["human-approval-required"]);
  assert.equal(result.policyVersion, "social-v1");
});
