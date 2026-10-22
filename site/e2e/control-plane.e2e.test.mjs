// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import { after, test } from "node:test";
import { chromium } from "playwright";

const baseURL = (process.env.BASE_URL || "http://127.0.0.1:3000").replace(/\/$/, "");
const browser = await chromium.launch({ headless: true });

after(async () => {
  await browser.close();
});

async function openPage(path, viewport = { width: 1440, height: 1000 }) {
  const page = await browser.newPage({ viewport });
  const pageErrors = [];
  const consoleErrors = [];
  const failedRequests = [];
  const scriptResponses = [];
  page.on("pageerror", (error) => pageErrors.push(error.message));
  page.on("console", (message) => {
    if (message.type() === "error") consoleErrors.push(message.text());
  });
  page.on("requestfailed", (request) => {
    failedRequests.push({ url: request.url(), error: request.failure()?.errorText || "unknown" });
  });
  page.on("response", (response) => {
    if (/\/_next\/static\/.*\.js(?:\?|$)/.test(response.url())) {
      scriptResponses.push({ url: response.url(), status: response.status() });
    }
  });
  const e2eToken = process.env.NAEOS_E2E_CHALLENGE_TOKEN;
  if (e2eToken) {
    const productionOrigin = new URL(baseURL).origin;
    await page.route("**/*", async (route) => {
      const request = route.request();
      const requestURL = new URL(request.url());
      const isControlPlaneDocument =
        requestURL.origin === productionOrigin &&
        /^\/(?:en|id)\/control-plane\/?$/.test(requestURL.pathname) &&
        request.method() === "GET";
      // Cloudflare's narrowly scoped E2E exception must cover the document and
      // the Next.js static assets required to hydrate it. Without the same
      // header on /_next/static requests, production can return 403 HTML for JS.
      const isNextStaticAsset =
        requestURL.origin === productionOrigin &&
        requestURL.pathname.startsWith("/_next/static/") &&
        request.method() === "GET";
      if (isControlPlaneDocument || isNextStaticAsset) {
        await route.continue({
          headers: {
            ...request.headers(),
            "x-naeos-e2e-token": e2eToken,
          },
        });
        return;
      }
      await route.continue();
    });
  }
  const response = await page.goto(`${baseURL}${path}`, { waitUntil: "domcontentloaded", timeout: 45_000 });
  assert.ok(response, `No document response for ${path}`);
  const headers = await response.allHeaders();
  const challengeResponse = response.status() === 403 && headers["cf-mitigated"] === "challenge";
  if (response.status() >= 400 && !challengeResponse) {
    const body = (await response.text()).slice(0, 1200).replace(/\\s+/g, " ");
    assert.fail(`Unexpected HTTP ${response.status()} for ${path}; headers=${JSON.stringify({
      server: headers.server,
      "cf-ray": headers["cf-ray"],
      "cf-mitigated": headers["cf-mitigated"],
      "content-type": headers["content-type"],
      location: headers.location,
    })}; body=${body}`);
  }
  try {
    // Cloudflare Managed Challenge can return an initial 403 while the real browser
    // executes the challenge. Require the actual dashboard to render before passing.
    await page.locator(".control-plane-dashboard").waitFor({ timeout: challengeResponse ? 30_000 : 20_000 });
  } catch {
    const body = (await page.locator("body").innerText().catch(() => "")).slice(0, 1200).replace(/\\s+/g, " ");
    assert.fail(`Control Plane did not render for ${path}; initial HTTP ${response.status()}; headers=${JSON.stringify({
      server: headers.server,
      "cf-ray": headers["cf-ray"],
      "cf-mitigated": headers["cf-mitigated"],
      "content-type": headers["content-type"],
      location: headers.location,
    })}; body=${body}`);
  }
  // Wait for the client component's explicit hydration marker instead of
  // depending on React's private DOM properties, which are implementation details.
  try {
    await page.locator('[data-testid="control-plane-live-demo"][data-hydrated="true"]').waitFor({
      state: "attached",
      timeout: 15_000,
    });
  } catch (error) {
    const diagnostics = await page.evaluate(() => ({
      url: location.href,
      readyState: document.readyState,
      markerCount: document.querySelectorAll('[data-testid="control-plane-live-demo"]').length,
      markerState: document.querySelector('[data-testid="control-plane-live-demo"]')?.getAttribute("data-hydrated") ?? null,
      scripts: Array.from(document.scripts, (script) => script.src).filter(Boolean),
      nextBuildId: document.querySelector('meta[name="next-build-id"]')?.getAttribute("content") ?? null,
    })).catch((diagnosticError) => ({ diagnosticError: String(diagnosticError) }));
    assert.fail(
      `Client hydration marker did not become true for ${path}; diagnostics=${JSON.stringify({
        ...diagnostics,
        pageErrors,
        consoleErrors: consoleErrors.slice(-20),
        failedRequests: failedRequests.slice(-30),
        scriptResponses: scriptResponses.slice(-30),
      })}; originalError=${String(error)}`,
    );
  }
  return { page, pageErrors };
}

async function assertScrolledIntoView(page, selector) {
  const element = page.locator(selector);
  await element.waitFor({ state: "visible" });
  await page.waitForFunction((sel) => {
    const node = document.querySelector(sel);
    if (!node) return false;
    const rect = node.getBoundingClientRect();
    return rect.top >= -1 && rect.top < window.innerHeight - 1 && rect.height > 0;
  }, selector, { timeout: 6_000 });
}

async function readPolicyProofState(proof) {
  const badge = await proof.locator("strong").first().innerText();
  const json = await proof.locator("pre").innerText();
  return { badge, json };
}

function assertJSONState(json, expected) {
  for (const [key, value] of Object.entries(expected)) {
    assert.match(json, new RegExp(`"${key}":\\s*"?${value}"?`), `expected policy-proof JSON ${key}=${value}`);
  }
}

test("English Control Plane renders truthful state and safe action boundaries", async () => {
  const { page, pageErrors } = await openPage("/en/control-plane");
  try {
    await page.getByRole("heading", { name: "Operate governance, not just logs." }).waitFor();
    const status = page.locator(".control-plane-console-header .console-status");
    await status.waitFor();
    assert.match(await status.getAttribute("class") || "", /is-(healthy|error|neutral)/,
      "status must use an explicit healthy/error/neutral state class");

    assert.equal(await page.getByRole("button", { name: "Agents", exact: true }).isDisabled(), true);
    assert.equal(await page.getByRole("button", { name: "Policies", exact: true }).isDisabled(), true);
    assert.equal(await page.getByRole("button", { name: "Exceptions", exact: true }).isDisabled(), true);
    assert.equal(await page.getByRole("button", { name: "View Policy", exact: true }).isDisabled(), true);

    const decisionsHeading = await page.locator(".control-plane-panel .panel-header h4").first().innerText();
    if (/sample/i.test(decisionsHeading)) {
      await page.locator(".control-plane-data-note").waitFor();
      assert.equal(await page.getByRole("button", { name: "View Evidence", exact: true }).isDisabled(), true,
        "sample decisions must not open evidence");
    }

    await page.getByRole("button", { name: "Runs", exact: true }).click();
    const runs = page.locator("#control-plane-runs");
    await runs.waitFor();
    assert.ok(await runs.isVisible(), "Runs section should remain visible after navigation");
    assert.deepEqual(pageErrors, [], "page should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Indonesian Control Plane renders localized copy and preserves disabled boundaries", async () => {
  const { page, pageErrors } = await openPage("/id/control-plane");
  try {
    await page.getByRole("heading", { name: "Operasikan governance, bukan sekadar melihat log." }).waitFor();
    assert.equal(await page.getByRole("button", { name: "Agents", exact: true }).isDisabled(), true);
    assert.equal(await page.getByRole("button", { name: "Policies", exact: true }).isDisabled(), true);
    assert.equal(await page.getByRole("button", { name: "Lihat Policy", exact: true }).isDisabled(), true);
    assert.deepEqual(pageErrors, [], "Indonesian page should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Control Plane dashboard fits a mobile viewport without horizontal shell overflow", async () => {
  const { page, pageErrors } = await openPage("/en/control-plane", { width: 390, height: 844 });
  try {
    const shell = page.locator(".control-plane-shell");
    await shell.waitFor();
    const dimensions = await shell.evaluate((element) => ({
      clientWidth: element.clientWidth,
      scrollWidth: element.scrollWidth,
    }));
    assert.ok(
      dimensions.scrollWidth <= dimensions.clientWidth + 2,
      `dashboard shell overflows horizontally: client=${dimensions.clientWidth}px scroll=${dimensions.scrollWidth}px`,
    );
    assert.deepEqual(pageErrors, [], "mobile page should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Policy Change Proof walks all six deterministic states and resets on replay", async () => {
  const { page, pageErrors } = await openPage("/en/control-plane");
  try {
    const proof = page.locator('[data-testid="policy-change-proof"]');
    await proof.waitFor({ state: "visible" });

    let state = await readPolicyProofState(proof);
    assert.equal(state.badge, "REQUEST");
    assertJSONState(state.json, { policy_version: 1, authorization: "NONE", execution: "NOT_REQUESTED", evidence: "PENDING" });

    const next = proof.getByRole("button", { name: "Next" });
    assert.ok(await proof.getByRole("button", { name: "Previous" }).isDisabled(), "Previous is disabled on the first step");

    await next.click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "ALLOW", "step 2 must authorize the capability");
    assertJSONState(state.json, { policy_version: 1, authorization: "NONE", execution: "NOT_REQUESTED" });

    await next.click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "AUTHORIZED", "step 3 must issue an authorization");
    assertJSONState(state.json, { policy_version: 1, authorization: "ACTIVE", execution: "READY" });

    await next.click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "POLICY CHANGED", "step 4 must surface the policy transition");
    assertJSONState(state.json, { policy_version: 2, authorization: "ACTIVE", execution: "READY" });

    await next.click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "BLOCKED", "step 5 must reject the stale authorization");
    assertJSONState(state.json, { policy_version: 2, authorization: "STALE", execution: "BLOCKED" });

    await next.click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "VERIFIED", "final step must retain verifiable evidence");
    assertJSONState(state.json, { policy_version: 2, authorization: "STALE", execution: "BLOCKED", evidence: "RETAINED" });
    assert.ok(await proof.getByRole("button", { name: "Complete" }).isDisabled(), "Complete is disabled on the final step");

    await proof.getByRole("button", { name: "Replay proof" }).click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "REQUEST", "replay must reset to the initial request step");
    assertJSONState(state.json, { policy_version: 1, authorization: "NONE", execution: "NOT_REQUESTED", evidence: "PENDING" });

    assert.deepEqual(pageErrors, [], "policy proof should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Policy Change Proof is localized in Indonesian and drives the same states", async () => {
  const { page, pageErrors } = await openPage("/id/control-plane");
  try {
    const proof = page.locator('[data-testid="policy-change-proof"]');
    await proof.waitFor({ state: "visible" });

    let state = await readPolicyProofState(proof);
    assert.equal(state.badge, "REQUEST");
    assertJSONState(state.json, { policy_version: 1, authorization: "NONE" });

    await proof.getByRole("button", { name: "Lanjutkan" }).click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "ALLOW");

    for (let step = 0; step < 3; step += 1) {
      await proof.getByRole("button", { name: "Lanjutkan" }).click();
    }
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "BLOCKED", "stale authorization must be blocked after the policy change");
    assertJSONState(state.json, { policy_version: 2, authorization: "STALE", execution: "BLOCKED" });

    await proof.getByRole("button", { name: "Lanjutkan" }).click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "VERIFIED");
    assertJSONState(state.json, { evidence: "RETAINED" });
    assert.ok(await proof.getByRole("button", { name: "Selesai" }).isDisabled(), "Selesai is disabled on the final step");

    await proof.getByRole("button", { name: "Ulangi proof" }).click();
    state = await readPolicyProofState(proof);
    assert.equal(state.badge, "REQUEST", "replay must reset the localized proof");

    assert.deepEqual(pageErrors, [], "Indonesian policy proof should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Control Plane live demo gates evaluation until an endpoint is configured", async () => {
  const { page, pageErrors } = await openPage("/en/control-plane");
  try {
    const demo = page.locator('[data-testid="control-plane-live-demo"]');
    await demo.waitFor({ state: "visible" });

    const allowButton = demo.getByRole("button", { name: "ALLOW · repository.read" });
    const denyButton = demo.getByRole("button", { name: "DENY · production.deploy" });
    assert.equal(await allowButton.getAttribute("aria-pressed"), "true", "ALLOW scenario is selected by default");
    assert.equal(await denyButton.getAttribute("aria-pressed"), "false");

    await denyButton.click();
    assert.equal(await denyButton.getAttribute("aria-pressed"), "true", "scenario selection must toggle aria-pressed");
    assert.equal(await allowButton.getAttribute("aria-pressed"), "false");

    const connected = await demo.getByText("ENDPOINT READY").isVisible().catch(() => false);
    if (connected) {
      assert.ok(await demo.getByRole("button", { name: "Evaluate Request" }).isEnabled(), "evaluation is enabled when endpoint is configured");
    } else {
      assert.ok(await demo.getByText(/^NOT CONNECTED$/).isVisible(), "demo must report NOT CONNECTED when no endpoint is configured");
      await demo.getByRole("button", { name: "Evaluate Request" }).click();
      const alert = demo.locator('[role="alert"]');
      await alert.waitFor({ state: "visible" });
      assert.match(await alert.innerText(), /Configure the Control Plane endpoint first/, "unconfigured evaluation must fail closed with instructions");
      assert.equal(await demo.locator('[aria-live="polite"]').count(), 0,
        "no decision result may render when evaluation is blocked");
    }

    assert.match(await demo.locator("small").innerText(), /Not configured|endpoint/i, "active endpoint label must be visible");
    assert.deepEqual(pageErrors, [], "live demo should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Indonesian live demo localizes the unconfigured evaluation gate", async () => {
  const { page, pageErrors } = await openPage("/id/control-plane");
  try {
    const demo = page.locator('[data-testid="control-plane-live-demo"]');
    await demo.waitFor({ state: "visible" });

    const connected = await demo.getByText("ENDPOINT READY").isVisible().catch(() => false);
    if (!connected) {
      assert.ok(await demo.getByText(/^BELUM TERHUBUNG$/).isVisible(), "demo must report BELUM TERHUBUNG when no endpoint is configured");
      await demo.getByRole("button", { name: "Evaluate Request" }).click();
      const alert = demo.locator('[role="alert"]');
      await alert.waitFor({ state: "visible" });
      assert.match(await alert.innerText(), /Konfigurasikan endpoint Control Plane terlebih dahulu/, "unconfigured evaluation must fail closed with localized instructions");
    }

    const allowButton = demo.getByRole("button", { name: "ALLOW · repository.read" });
    await allowButton.click();
    assert.equal(await allowButton.getAttribute("aria-pressed"), "true");
    assert.deepEqual(pageErrors, [], "Indonesian live demo should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Selecting a decision updates the decision detail panel", async () => {
  const { page, pageErrors } = await openPage("/en/control-plane");
  try {
    const list = page.locator(".decision-list");
    await list.waitFor({ state: "visible" });
    const rows = list.locator(".decision-row");
    const count = await rows.count();
    assert.ok(count >= 1, "expected at least one decision row");

    const detailHeader = page.locator("#control-plane-evidence .panel-header h4");
    const initialHeader = await detailHeader.innerText();
    const isSample = await page.locator(".control-plane-data-note").isVisible().catch(() => false);

    if (isSample) {
      assert.ok(count >= 3, "sample mode must render the three sample decisions");
      const targetRow = rows.nth(1);
      const targetText = await targetRow.innerText();
      await targetRow.click();
      assert.match(targetText, /test\.run/, "sample row 2 is the ALLOW test.run decision");
      assert.equal(await detailHeader.innerText(), "DEC-91A68", "detail panel must switch to the selected decision");
      const facts = await page.locator("#control-plane-evidence .decision-facts").innerText();
      assert.match(facts, /coding-agent-07/, "detail panel must show the selected decision agent");
      assert.match(facts, /engineering\/default-v2/, "detail panel must show the selected decision policy");
      assert.match(await targetRow.getAttribute("class") || "", /selected/, "selected decision row must be highlighted");
    } else if (count > 1) {
      const targetRow = rows.nth(count - 1);
      await targetRow.click();
      assert.notEqual(await detailHeader.innerText(), initialHeader, "detail panel must follow the live selection");
      assert.match(await targetRow.getAttribute("class") || "", /selected/, "selected decision row must be highlighted in live mode");
    }

    assert.deepEqual(pageErrors, [], "decision selection should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Control Plane live demo sends a Bearer token only when one is configured", async () => {
  const { page, pageErrors } = await openPage("/en/control-plane");
  try {
    const demo = page.locator('[data-testid="control-plane-live-demo"]');
    await demo.waitFor({ state: "visible" });

    const endpointInput = demo.locator('[data-testid="control-plane-live-demo-endpoint"]');
    const tokenInput = demo.locator('[data-testid="control-plane-live-demo-token"]');
    await tokenInput.waitFor({ state: "visible" });
    assert.equal(await tokenInput.getAttribute("type"), "password", "token field must mask its value");
    assert.equal(await tokenInput.getAttribute("autoComplete"), "off", "token field must disable autofill");
    // The public production endpoint uses the server-side proxy; this token is for custom endpoints.
    assert.equal(await tokenInput.getAttribute("placeholder"), "Token for custom endpoints only");

    const allowedPayload = JSON.stringify({
      status: "ALLOW",
      decision_id: "DEC-RESULT-1",
      request_id: "REQ-RESULT-1",
      allowed: true,
      reason: "capability_granted",
      policy_id: "POLICY-017",
      policy_version: 17,
      requested: "repository.read",
      evidence_endpoint: "/api/control-plane/evidence?decision_id=DEC-RESULT-1",
    });

    async function captureDecisionHeaders(fillToken) {
      const sentHeaders = {};
      await page.route("**/api/control-plane/decision", async (route) => {
        sentHeaders.authorization = route.request().headers()["authorization"] || null;
        await route.fulfill({ status: 200, contentType: "application/json", body: allowedPayload });
      });
      await endpointInput.fill(baseURL);
      await tokenInput.fill(fillToken);
      await demo.getByRole("button", { name: "Evaluate Request" }).click();
      await demo.locator('[aria-live="polite"]').waitFor({ state: "visible" });
      assert.match(await demo.locator('[aria-live="polite"] .decision-label + strong').innerText(), /ALLOW/,
        "a stubbed ALLOW decision must render");
      await page.unroute("**/api/control-plane/decision");
      return sentHeaders;
    }

    const withToken = await captureDecisionHeaders("test-bearer-token");
    assert.equal(withToken.authorization, "Bearer test-bearer-token",
      "decision request must carry the configured Bearer token");

    const withoutToken = await captureDecisionHeaders("");
    assert.equal(withoutToken.authorization, null,
      "decision request without a configured token must not send an Authorization header");

    assert.deepEqual(pageErrors, [], "token live demo should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Control Plane live demo surfaces an authentication failure from a protected endpoint", async () => {
  const { page, pageErrors } = await openPage("/en/control-plane");
  try {
    const demo = page.locator('[data-testid="control-plane-live-demo"]');
    await demo.waitFor({ state: "visible" });

    await page.route("**/api/control-plane/decision", async (route) => {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        headers: { "www-authenticate": "Bearer" },
        body: JSON.stringify({ status: "error", code: 401, message: "authentication required" }),
      });
    });

    await demo.locator('[data-testid="control-plane-live-demo-endpoint"]').fill(baseURL);
    await demo.getByRole("button", { name: "Evaluate Request" }).click();

    const alert = demo.locator('[role="alert"]');
    await alert.waitFor({ state: "visible" });
    assert.match(await alert.innerText(), /authentication required/, "401 from the endpoint must be surfaced to the user");
    assert.equal(await demo.locator('[aria-live="polite"]').count(), 0,
      "a rejected request must not render a decision result");

    await page.unroute("**/api/control-plane/decision");
    assert.deepEqual(pageErrors, [], "401 handling should not throw uncaught JavaScript errors");
  } finally {
    await page.close();
  }
});

test("Console navigation scrolls to the matching section in English and Indonesian", async () => {
  for (const lang of ["en", "id"]) {
    const { page, pageErrors } = await openPage(`/${lang}/control-plane`);
    try {
      const nav = page.locator(".control-plane-sidebar nav");
      await nav.waitFor({ state: "visible" });

      await nav.getByRole("button", { name: "Runs", exact: true }).click();
      await assertScrolledIntoView(page, "#control-plane-runs");

      await nav.getByRole("button", { name: "Decisions", exact: true }).click();
      await assertScrolledIntoView(page, "#control-plane-decisions");

      await nav.getByRole("button", { name: "Evidence", exact: true }).click();
      await assertScrolledIntoView(page, "#control-plane-evidence");

      await nav.getByRole("button", { name: "Verification", exact: true }).click();
      await assertScrolledIntoView(page, "#control-plane-evidence");

      assert.deepEqual(pageErrors, [], `console navigation (${lang}) should not throw uncaught JavaScript errors`);
    } finally {
      await page.close();
    }
  }
});

test("Control Plane hero and verified-path actions link to the documented destinations", async () => {
  for (const lang of ["en", "id"]) {
    const { page, pageErrors } = await openPage(`/${lang}/control-plane`);
    try {
      const id = lang === "id";
      const architectureHref = (await page.getByRole("link", { name: id ? "Lihat Arsitektur" : "Explore Architecture" }).getAttribute("href")) || "";
      assert.match(architectureHref, id ? /\/id\/docs\/architecture\/?$/ : /\/docs\/architecture\/?$/, "architecture CTA must point at the architecture doc");

      const githubHref = (await page.getByRole("link", { name: id ? "Lihat GitHub" : "View GitHub" }).getAttribute("href")) || "";
      assert.equal(githubHref.replace(/\/$/, ""), "https://github.com/NAEOS-foundation/naeos", "GitHub CTA must point at the repository");

      const goldenPathHref = (await page.getByRole("link", { name: id ? "Jalankan Golden Path" : "Run Verified Golden Path" }).getAttribute("href")) || "";
      assert.match(goldenPathHref, /\/docs\/VERIFIED-GOLDEN-PATH\.md$/, "golden path link must point at the verified golden path doc");

      const verifierHref = (await page.getByRole("link", { name: id ? "Lihat Verifier" : "Inspect Verifier" }).getAttribute("href")) || "";
      assert.match(verifierHref, /\/docs\/control-plane\/p1-11-independent-verifier-cli\.md$/, "verifier link must point at the independent verifier doc");

      assert.deepEqual(pageErrors, [], `control-plane actions (${lang}) should not throw uncaught JavaScript errors`);
    } finally {
      await page.close();
    }
  }
});
