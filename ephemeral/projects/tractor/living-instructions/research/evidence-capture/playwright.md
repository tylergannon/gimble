# Playwright

## Purpose
Browser automation library plus test runner that produces screenshots, webm video, trace zips, HAR files, and pixel-diffed screenshot assertions from headless Chromium/Firefox/WebKit.

## Pinned
- URL: https://github.com/microsoft/playwright
- Version: v1.62.1 (released 2026-07-30)
- License: Apache-2.0

## Key concepts
- Screenshot: `page.screenshot({path, fullPage, clip, type: png|jpeg|webp, mask, animations: 'disabled', scale: 'css'|'device'})`; element screenshots via `locator.screenshot()`; buffer return for piping to a diff tool. https://playwright.dev/docs/screenshots ("Full page screenshots", "Capture into buffer", "Element screenshot"). Artifact: image, directly judgeable.
- Video: `recordVideo: {dir}` on a context, or test config `video: 'on'|'retain-on-failure'|'on-first-retry'`; webm, "video size defaults to the viewport size scaled down to fit 800x800"; file exists only after context close ("Make sure to await close"). https://playwright.dev/docs/videos. Artifact: video, not directly judgeable by an image-only model; extract frames with ffmpeg.
- Trace: `trace: 'on-first-retry'|'retain-on-failure'|'on'` writes trace.zip holding screenshots, DOM snapshots, network (headers and bodies), console. Open with `npx playwright show-trace trace.zip` or https://trace.playwright.dev ("loads the trace entirely in your browser"). https://playwright.dev/docs/trace-viewer ("Recording a trace", "Opening the trace viewer"). Artifact: zip; a model can judge it only after unzipping (resources/*.jpeg screenshots, *.trace NDJSON).
- HAR: `recordHar` on `browser.newContext()`, or `routeFromHAR(path, {update: true})`; CLI `npx playwright open --save-har=example.har --save-har-glob="**/api/**"`; `.zip` suffix bundles bodies. https://playwright.dev/docs/mock ("Recording a HAR file", "Recording HAR with CLI"). Artifact: JSON text, judgeable.
- Screenshot diff: `expect(page).toHaveScreenshot()` compares with pixelmatch; options `threshold` (YIQ, default 0.2), `maxDiffPixels`, `maxDiffPixelRatio`, `mask`, `stylePath`, `animations` (default disabled), `caret` (default hide). https://playwright.dev/docs/api/class-pageassertions#page-assertions-to-have-screenshot-1. Baseline named `{test}-{browser}-{platform}.png`. https://playwright.dev/docs/test-snapshots.
- On mismatch the matcher writes `-expected`, `-actual`, `-diff` files and attaches them: packages/playwright/src/matchers/toMatchSnapshot.ts lines 116-119 and 194-197.

## Bounded comparison
Like a full browser test harness but only for the three engines Playwright ships; no native desktop and no terminal.

## Gotchas
- Baselines are host-specific: "browser rendering can vary based on the host OS, version, settings, hardware, power source ... headless mode" (test-snapshots page). Generate and judge baselines in the same container image.
- Video is written only when the context closes; a crashed or killed run yields nothing. Video is downscaled to 800x800 max by default, which blurs small text for a judge.
- `trace: 'on'` for every test is documented as "not recommended" for performance; trace zips run to tens of MB with bodies.
- Trace and video are not model-readable as-is; the judgeable units are the PNG/JPEG screenshots inside them or the `-actual`/`-diff` PNGs.
- First run of `toHaveScreenshot` writes the baseline and fails; a verifier must seed baselines deliberately, not from the run under judgement.
- Headless uses `chromium_headless_shell` by default; `--headed` or `channel: 'chromium'` switches binaries and pixels.

## Recipe
- To capture a scenario as a per-step image set, start at https://playwright.dev/docs/trace-viewer "Recording a trace" with `retain-on-failure` (or `on` for a proof run), then unzip and hand the `resources/*.jpeg` files to the judge in order.
- To capture request evidence for an API-backed story, start at https://playwright.dev/docs/mock "Recording HAR with CLI" (`--save-har` with a glob on the API path).
