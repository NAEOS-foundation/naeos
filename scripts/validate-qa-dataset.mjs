#!/usr/bin/env node
// Copyright 2026 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

const datasetPath = resolve("tests/fixtures/qa-golden-v1.json");
const raw = readFileSync(datasetPath, "utf8");
const dataset = JSON.parse(raw);

assert.equal(dataset.dataset_id, "naeos-qa-golden-v1", "unexpected dataset_id");
assert.equal(dataset.version, 1, "unsupported dataset version");
assert.ok(Array.isArray(dataset.cases) && dataset.cases.length > 0, "cases must be a non-empty array");

const ids = new Set();
const categories = new Set([
  "grounding",
  "unsupported",
  "ambiguity",
  "conversation",
  "prompt-injection",
  "authorization",
  "robustness",
]);
const risks = new Set(["low", "medium", "high", "critical"]);
const requiredStrings = ["id", "category", "risk", "input", "expected_behavior"];

for (const [index, item] of dataset.cases.entries()) {
  const label = `cases[${index}]`;
  for (const field of requiredStrings) {
    assert.equal(typeof item[field], "string", `${label}.${field} must be a string`);
    assert.ok(item[field].trim().length > 0, `${label}.${field} must not be empty`);
  }
  assert.ok(!ids.has(item.id), `duplicate case id: ${item.id}`);
  ids.add(item.id);
  assert.ok(categories.has(item.category), `${label}.category is unsupported: ${item.category}`);
  assert.ok(risks.has(item.risk), `${label}.risk is unsupported: ${item.risk}`);
  assert.ok(Array.isArray(item.source_refs), `${label}.source_refs must be an array`);
  assert.ok(Array.isArray(item.assertions) && item.assertions.length > 0, `${label}.assertions must be non-empty`);
  assert.ok(Array.isArray(item.forbidden_behavior), `${label}.forbidden_behavior must be an array`);
  for (const sourceRef of item.source_refs) {
    assert.equal(typeof sourceRef, "string", `${label} source ref must be a string`);
    assert.ok(existsSync(resolve(sourceRef)), `${label} references missing source: ${sourceRef}`);
  }
  if (item.context !== undefined) {
    assert.ok(Array.isArray(item.context), `${label}.context must be an array when provided`);
    for (const [turnIndex, turn] of item.context.entries()) {
      assert.ok(["user", "assistant", "system"].includes(turn.role), `${label}.context[${turnIndex}].role is invalid`);
      assert.equal(typeof turn.content, "string", `${label}.context[${turnIndex}].content must be a string`);
    }
  }
}

console.log(`Validated ${dataset.cases.length} Q&A cases (${dataset.dataset_id} v${dataset.version}); all source references exist.`);
