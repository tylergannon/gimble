# agent-browser

## Purpose
Vercel Labs' native Rust CLI that drives Chromium over CDP for AI agents, one shell command per action, emitting screenshots, accessibility snapshots, HAR, CDP traces, console logs, and pixel diffs.

## Pinned
- URL: https://github.com/vercel-labs/agent-browser
- Version: v0.36.0 (released 2026-09-01)
- License: Apache-2.0

## Key concepts
- Install: `npm install -g agent-browser && agent-browser install` downloads Chrome for Testing; existing Chrome, Brave, Playwright, and Puppeteer browsers are auto-detected (README "Requirements").
- Screenshot: `agent-browser screenshot [path] [--full] [--annotate] [--screenshot-format jpeg --screenshot-quality 80]`; `--annotate` overlays numbered labels matching snapshot refs (README lines 133-136). Artifact: image.
- Snapshot: `agent-browser snapshot -i|-c|-d 3|-s "#main" [--json]` returns the a11y tree with deterministic `@e1` refs (README "Snapshot"). Artifact: text.
- HAR: `agent-browser network har start [--content all|none]` then `network har stop [out.har]` (README lines 343-346). Artifact: JSON text.
- Trace: `agent-browser trace start` / `trace stop [path]` saves a Chrome DevTools Protocol trace as `.json`, not a Playwright trace (README "Debug", line 413). Artifact: large JSON, weakly judgeable.
- Diff: `agent-browser diff screenshot --baseline before.png [-o d.png] [-t 0.2]` and `diff url A B --screenshot` (README lines 401-403). Artifact: diff PNG.
- Console and errors: `agent-browser console --json`, `agent-browser errors` (README "Debug").
- Sessions: `--session <name>` or `AGENT_BROWSER_SESSION` isolate cookies and history; headless by default, `--headed` auto-starts Xvfb on Linux (README "Headed vs Headless").

## Bounded comparison
Like Playwright's screenshot and HAR surface but only Chromium over CDP, one process per command, and no video recording.

## Gotchas
- No video, GIF, or screencast command exists in the README (grep for video/record/webm/gif finds only HAR, trace, and React render lines); scenario evidence is a sequence of screenshots the verifier must take itself.
- Default command timeout is 25 s (`AGENT_BROWSER_DEFAULT_TIMEOUT`); slow CI pages fail into a screenshot of a half-loaded page.
- "Headless Chromium screenshots hide native scrollbars" (README line 97); baselines from headed runs will not match.
- A background daemon holds browser state between commands; a stale daemon from a prior run can leak session state into the evidence.
- Screenshots without a path go to a temp directory; always pass `--screenshot-dir` so the judge can find them.

## Recipe
- To capture a scenario as ordered images, start at README "Batch" (`agent-browser batch --bail "open URL" "click @e1" "screenshot step1.png"`) so one failed step halts the sequence.
- To capture request evidence, start at README lines 343-346 (`network har start` before `open`, `network har stop out.har` after the last action).
