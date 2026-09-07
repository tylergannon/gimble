import { chromium } from 'playwright';
import { readFileSync, writeFileSync, existsSync } from 'node:fs';

const S = process.env.S;
const yamlPath = `${S}/live/bake-off.yaml`;
const sidecar = `${S}/live/bake-off.layout.json`;
const url = 'http://127.0.0.1:7331/';

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });
const consoleErrors = [];
page.on('console', (m) => { if (m.type() === 'error' || m.type() === 'warning') consoleErrors.push(`${m.type()}: ${m.text()}`); });
page.on('pageerror', (e) => consoleErrors.push(`pageerror: ${e.message}`));
const remoteCalls = [];
page.on('request', (r) => { if (r.url().includes('/_app/remote/')) remoteCalls.push(`${r.method()} ${r.url().replace(url, '/')}`); });

await page.goto(url);
await page.waitForSelector('.node');
const nodes = await page.locator('.node').count();
const count = await page.locator('header .count').textContent();
const name = await page.locator('header .name').textContent();
console.log('1. loaded:', { nodes, count, name });
await page.screenshot({ path: `${S}/proof-1-loaded.png` });

// 2. drag a node; the sidecar must appear beside the pipeline.
if (existsSync(sidecar)) throw new Error('sidecar exists before any move');
const box = await page.locator('.node').first().boundingBox();
await page.mouse.move(box.x + 20, box.y + 20);
await page.mouse.down();
await page.mouse.move(box.x + 120, box.y + 80, { steps: 8 });
await page.mouse.up();
await page.waitForTimeout(1500);
if (!existsSync(sidecar)) throw new Error('sidecar not written after a move');
const layout = JSON.parse(readFileSync(sidecar, 'utf-8'));
console.log('2. moved a node; sidecar:', Object.keys(layout).length, 'entries, first:', Object.entries(layout)[0]);
await page.screenshot({ path: `${S}/proof-2-moved.png` });

// 3. an edit from the inspector: rename the selected node's label, wait for the save, read the file.
const yamlBefore = readFileSync(yamlPath, 'utf-8');
await page.locator('.node').first().click();
const labelField = page.locator('input').filter({ has: page.locator('xpath=.') }).first();
// find the Name field in the inspector
const nameInput = page.getByLabel('Name').first();
await nameInput.fill('Renamed by the browser');
await page.waitForTimeout(1500);
const yamlAfter = readFileSync(yamlPath, 'utf-8');
if (!yamlAfter.includes('Renamed by the browser')) throw new Error('the label edit did not reach the file:\n' + yamlAfter);
console.log('3. inspector edit reached disk; diff lines:', yamlAfter.split('\n').filter((l) => !yamlBefore.includes(l)));
await page.screenshot({ path: `${S}/proof-3-edited.png` });

// 4. a change from disk: the page must follow it without a reload.
const changed = yamlAfter.replace(/^name: (.*)$/m, 'name: $1-from-disk');
if (changed === yamlAfter) throw new Error('no name line to change');
writeFileSync(yamlPath, changed);
await page.waitForFunction(() => document.querySelector('header .name')?.textContent?.includes('from-disk'), null, { timeout: 5000 });
console.log('4. disk change followed; header now:', await page.locator('header .name').textContent());
await page.screenshot({ path: `${S}/proof-4-from-disk.png` });

// 5. a syntax error on disk puts the page in read-only mode with a banner.
writeFileSync(yamlPath, changed + 'nodes: [\n');
await page.waitForSelector('.banner', { timeout: 5000 });
console.log('5. syntax error banner:', (await page.locator('.banner').first().textContent()).trim().slice(0, 120));
await page.screenshot({ path: `${S}/proof-5-syntax-error.png` });
writeFileSync(yamlPath, changed);
await page.waitForFunction(() => !document.querySelector('.banner'), null, { timeout: 5000 });
console.log('5b. recovered after the file parses again');

console.log('remote calls:', remoteCalls);
console.log('console errors:', consoleErrors);
await browser.close();
