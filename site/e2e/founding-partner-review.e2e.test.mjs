// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import { after, test } from "node:test";
import { chromium } from "playwright";

const baseURL = (process.env.BASE_URL || "http://127.0.0.1:3000").replace(/\/$/, "");
const supabaseURL = (process.env.NEXT_PUBLIC_SUPABASE_URL || "").replace(/\/$/, "");
const publishableKey = process.env.NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY || "";
assert.ok(supabaseURL, "NEXT_PUBLIC_SUPABASE_URL must point to the staging project");
assert.ok(publishableKey, "NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY must be configured for staging E2E");

const browser = await chromium.launch({ headless: true });
after(async () => browser.close());

async function openReviewPage(lang, viewport = { width: 1440, height: 1000 }) {
  const page = await browser.newPage({ viewport });
  const pageErrors = [];
  page.on("pageerror", (error) => pageErrors.push(error.message));
  const response = await page.goto(`${baseURL}/${lang}/founding-partners/review`, {
    waitUntil: "domcontentloaded",
    timeout: 45_000,
  });
  assert.ok(response, "review page must return a document response");
  assert.ok(response.status() < 400, `review page returned HTTP ${response.status()}`);
  await page.getByRole("heading", {
    name: lang === "id" ? "Review Founding Partner" : "Founding Partner Review",
  }).waitFor({ timeout: 20_000 });
  return { page, pageErrors };
}

test("Indonesian founder review route renders sign-in and exposes no self-enrollment", async () => {
  const { page, pageErrors } = await openReviewPage("id");
  try {
    const configured = await page.getByLabel("Email reviewer").isVisible().catch(() => false);
    if (configured) {
      await page.getByLabel("Password").waitFor({ state: "visible" });
      await page.getByRole("button", { name: "Masuk", exact: true }).waitFor({ state: "visible" });
    } else {
      await page.getByRole("alert").filter({ hasText: /belum dikonfigurasi/i }).waitFor({ state: "visible" });
    }
    assert.equal(await page.getByRole("link", { name: /daftar|register|sign up/i }).count(), 0,
      "reviewer self-enrollment must not be exposed");
    assert.deepEqual(pageErrors, [], "review page must not throw uncaught browser errors");
  } finally {
    await page.close();
  }
});

test("English founder review route renders the restricted reviewer sign-in", async () => {
  const { page, pageErrors } = await openReviewPage("en");
  try {
    const configured = await page.getByLabel("Reviewer email").isVisible().catch(() => false);
    if (configured) {
      await page.getByLabel("Password").waitFor({ state: "visible" });
      await page.getByRole("button", { name: "Sign in", exact: true }).waitFor({ state: "visible" });
    } else {
      await page.getByRole("alert").filter({ hasText: /not configured/i }).waitFor({ state: "visible" });
    }
    assert.equal(await page.getByRole("link", { name: /sign up|register/i }).count(), 0,
      "reviewer self-enrollment must not be exposed");
    assert.deepEqual(pageErrors, [], "review page must not throw uncaught browser errors");
  } finally {
    await page.close();
  }
});

test("founder review route remains within a narrow mobile viewport", async () => {
  const { page, pageErrors } = await openReviewPage("id", { width: 390, height: 844 });
  try {
    const dimensions = await page.evaluate(() => ({
      viewport: document.documentElement.clientWidth,
      document: document.documentElement.scrollWidth,
    }));
    assert.ok(
      dimensions.document <= dimensions.viewport + 2,
      `review page overflows horizontally: viewport=${dimensions.viewport}px document=${dimensions.document}px`,
    );
    assert.deepEqual(pageErrors, [], "mobile review page must not throw uncaught browser errors");
  } finally {
    await page.close();
  }
});

test("anonymous clients cannot execute reviewer RPCs or select private recruitment tables", async () => {
  const headers = { apikey: publishableKey, "Content-Type": "application/json" };
  const calls = [
    ["founding_partner_review_context", {}],
    ["founding_partner_review_list", { p_limit: 10, p_offset: 0, p_status: null }],
    ["founding_partner_review_detail", { p_application_id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" }],
    ["founding_partner_review_transition", {
      p_application_id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
      p_new_status: "Screening",
      p_note: "anonymous E2E must be rejected",
    }],
  ];

  for (const [name, body] of calls) {
    const response = await fetch(`${supabaseURL}/rest/v1/rpc/${name}`, {
      method: "POST",
      headers,
      body: JSON.stringify(body),
    });
    assert.ok(response.status >= 400, `anonymous request unexpectedly executed ${name}`);
  }

  const directTable = await fetch(
    `${supabaseURL}/rest/v1/founding_partner_applications?select=id`,
    { headers: { apikey: publishableKey, "Accept-Profile": "recruitment" } },
  );
  assert.ok(directTable.status >= 400,
    "the private recruitment schema must not be available through the PostgREST Data API");
});
