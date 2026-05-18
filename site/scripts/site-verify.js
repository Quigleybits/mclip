// site-verify.js
//
// Smoke-test the mclip.dev landing page. Asserts page renders, key sections
// are present, and the canonical config schema is served correctly.
//
// Usage (from site/): node scripts/site-verify.js [base-url]
// Defaults to http://localhost:4322 (matches `astro preview --port 4322`).

import { chromium } from "playwright";

const baseUrl = process.argv[2] || "http://localhost:4322";

const browser = await chromium.launch();
const page = await browser.newPage();

await page.goto(baseUrl, { waitUntil: "load" });

const title = await page.title();
if (!title.includes("MCLIP")) {
  throw new Error(`Title missing MCLIP: ${title}`);
}

const h1 = (await page.locator("h1").first().textContent()).trim();
if (h1 !== "MCLIP") {
  throw new Error(`Hero h1 mismatch: ${h1}`);
}

const specCount = await page.locator("ul.specs > li").count();
if (specCount !== 9) {
  throw new Error(`Expected 9 spec links, got ${specCount}`);
}

const moduleCount = await page.locator("ul.modules > li").count();
if (moduleCount !== 9) {
  throw new Error(`Expected 9 modules, got ${moduleCount}`);
}

const schemaLink = await page.locator('a[href="/schemas/config/v0.json"]').count();
if (schemaLink < 1) {
  throw new Error("Schema link not present on page");
}

await page.screenshot({
  path: "scripts/site-verify-screenshot.png",
  fullPage: true,
});

const schemaResp = await page.goto(`${baseUrl}/schemas/config/v0.json`);
if (!schemaResp.ok()) {
  throw new Error(`Schema endpoint not OK: ${schemaResp.status()}`);
}
const schemaText = await schemaResp.text();
const schema = JSON.parse(schemaText);

if (schema.$id !== "https://mclip.dev/schemas/config/v0.json") {
  throw new Error(`Schema $id wrong: ${schema.$id}`);
}
if (!schema.$defs || !schema.$defs.serverEntry) {
  throw new Error("Schema missing $defs.serverEntry");
}

await browser.close();

console.log("OK site verification passed");
console.log(`  title:       ${title}`);
console.log(`  hero h1:     ${h1}`);
console.log(`  spec links:  ${specCount}`);
console.log(`  modules:     ${moduleCount}`);
console.log(`  schema $id:  ${schema.$id}`);
console.log(`  screenshot:  scripts/site-verify-screenshot.png`);
