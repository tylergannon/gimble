# Playwright MCP

## Purpose
Microsoft's MCP server (and CLI) that lets an agent drive a Playwright browser through accessibility snapshots and emit screenshots, traces, video, network, and console artifacts into an output directory.

## Pinned
- URL: https://github.com/microsoft/playwright-mcp
- Version: v0.0.80 (released 2026-09-01)
- License: Apache-2.0

## Key concepts
- Design: "operates on accessibility snapshots, bypassing the need for screenshots or visually-tuned models" (README, top). `browser_snapshot` returns the a11y tree as text with element refs. Artifact: text, judgeable.
- `browser_take_screenshot` (png/jpeg, fullPage, or element); the README notes "You can't perform actions based on the screenshot, use browser_snapshot for actions" (README lines 1076-1078). Artifact: image.
- `browser_start_tracing` / `browser_stop_tracing` produce a Playwright trace; `browser_start_video` / `browser_stop_video` record webm with optional filename and size (README lines 1408-1442). Artifacts: zip and video, judgeable only after extraction.
- `browser_network_requests` and `browser_console_messages` return text lists; `--caps network` unlocks route tools (README "Network (opt-in via --caps=network)", line 1149).
- Flags: `--output-dir <path>`, `--headless` ("headed by default"), `--isolated`, `--save-session`, `--viewport-size 1280x720`, `--caps vision,pdf,devtools` (README config table lines 415-454).
- README states "CLI invocations are more token-efficient: they avoid loading large tool schemas and verbose accessibility trees into the model context."

## Bounded comparison
Like agent-browser but only reachable as an MCP server or its bundled CLI, and it yields Playwright-native trace zips and webm rather than CDP JSON.

## Gotchas
- Headed by default; CI must pass `--headless` or the launch fails without a display.
- "A persistent profile can only be used by one browser instance at a time" (README line 481); parallel verifiers need `--isolated` or distinct `--user-data-dir`.
- Every screenshot returned through MCP enters the operator's context as an image; for a separate judge, rely on `--output-dir` files, not the conversation.
- Version is 0.0.x; tool names have changed across releases, so pin the npm version in the verifier's invocation.

## Recipe
- To capture a scenario, start at the README config table (`--output-dir`, `--headless`) and call `browser_start_tracing` before the first navigation, `browser_stop_tracing` after the last assertion.
