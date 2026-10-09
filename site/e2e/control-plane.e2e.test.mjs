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
  page.on("pageerror", (error) => pageErrors.push(error.message));
  const response = await page.goto(`${baseURL}${path}`, { waitUntil: "domcontentloaded", timeout: 45_000 });
  assert.ok(response, `No document response for ${path}`);
  assert.ok(response.status() < 400, `Unexpected HTTP ${response.status()} for ${path}`);
  await page.locator(".control-plane-dashboard").waitFor({ timeout: 20_000 });
  return { page, pageErrors };
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
