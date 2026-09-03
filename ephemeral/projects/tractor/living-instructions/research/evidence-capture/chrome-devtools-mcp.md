# Chrome DevTools MCP

## Purpose
Google's MCP server that gives an agent a Puppeteer-driven Chrome with DevTools access: screenshots, a11y snapshots, network and console listings, performance traces, and an experimental screencast.

## Pinned
- URL: https://github.com/ChromeDevTools/chrome-devtools-mcp
- Version: chrome-devtools-mcp-v1.8.0 (released 2026-08-25)
- License: Apache-2.0

## Key concepts
- `take_screenshot` with `format` (png/jpeg/webp), `quality`, `fullPage`, `filePath`, `uid` (element); returned inline as image content or saved when `filePath` is set (docs/tool-reference.md). Artifact: image.
- `take_snapshot` writes the a11y tree as text, optional `filePath` and `verbose` (tool-reference). Artifact: text.
- `list_network_requests` (paged, `resourceTypes`, `includePreservedRequests` across three navigations) and `get_network_request` with `requestFilePath` / `responseFilePath` (tool-reference). No HAR export tool.
- `list_console_messages` with `includeStackTraces` (tool-reference). Artifact: text.
- `performance_start_trace` / `performance_stop_trace` save a raw DevTools trace to `filePath`; `performance_analyze_insight` summarises it (tool-reference). Artifact: large JSON.
- `screencast_start` / `screencast_stop` write .webm/.mp4, gated behind `--experimentalScreencast=true` (tool-reference).
- Flags: `--headless`, `--isolated`, `--channel`, `--executablePath`, `--viewport`, `--browserUrl`, `--slim` (README "Configuration"). Requires Node LTS; "Officially supports Google Chrome and Chrome for Testing only" (README "Requirements").

## Bounded comparison
Like Playwright MCP but only Chrome, and with DevTools performance traces and an experimental screencast instead of Playwright trace zips.

## Gotchas
- Screencast is experimental and off by default; do not plan scenario video around it.
- Network bodies are only captured when explicitly fetched per request id; there is no one-shot HAR, so request evidence is a set of files the verifier must ask for.
- README warns the server "exposes content of the browser instance to the MCP clients"; a verifier should run it against a throwaway profile (`--isolated`).
- `npx -y chrome-devtools-mcp@latest` is the documented install; pin `@1.8.0` for reproducible evidence.

## Recipe
- To capture a scenario, start at docs/tool-reference.md `take_screenshot` with `filePath` per step and `take_snapshot` with `filePath` alongside each image.
