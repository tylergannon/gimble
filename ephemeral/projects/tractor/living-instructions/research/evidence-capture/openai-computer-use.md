# OpenAI computer use

## Purpose
OpenAI's Responses API `computer` tool returns UI actions for a harness to execute against screenshots; the sample app wraps Playwright and records screenshots and events into a replay bundle an operator console can review.

## Pinned
- Docs: https://developers.openai.com/api/docs/guides/tools-computer-use (GA tool `{ type: "computer" }` with `model: "gpt-5.6"`, replacing `computer-use-preview`).
- Sample app: https://github.com/openai/openai-cua-sample-app — HEAD 3751c8baa637 (2026-03-05), MIT; default model `gpt-5.4`; Node 22.20.0, pnpm 10.26.0, Playwright Chromium.

## Key concepts
- Loop: model emits `computer_call` with batched `actions[]`; harness executes and replies with `computer_call_output` carrying an `input_image` screenshot (base64, `detail: "original"`); repeat until no `computer_call` (docs "The interaction loop").
- Three integration paths in the docs: built-in computer tool, custom harness (Playwright, Selenium, VNC, MCP), or code execution where the model scripts the UI.
- Safety: `pending_safety_checks` must be acknowledged; docs say "Treat only direct user instructions as permission" and never on-screen text.
- Sample app layout: `packages/replay-schema` (request/response/replay contracts), `packages/browser-runtime` (Playwright session), `packages/runner-core/src/responses-loop.ts` (the loop), `apps/runner` (Fastify with SSE, "artifact serving", "screenshot artifact routes"), `apps/demo-web` operator console with `ScreenshotPane.tsx` and `RunSummary.tsx` (README "Repo Map"; docs/architecture.md "Runtime Flow": the loop "emits events, screenshots, and final verification results back into the replay bundle").
- Modes: `native` (raw computer actions) and `code` (persistent Playwright JS REPL via `exec_js`) against the same browser (README "Execution Modes").

## Bounded comparison
Like Anthropic's computer toolset but only through the Responses API, and its reference harness is browser-only with a replay bundle instead of loose PNGs.

## Gotchas
- The sample "focuses exclusively on browsers"; README "Safety And Limitations" says do not point it at authenticated or high-stakes environments.
- Replay bundle file layout is not documented in README or docs/architecture.md; a verifier must read `packages/replay-schema/src/index.ts` to learn the on-disk shape before promising a judge a format.
- Sample last pushed 2026-03-05 and pins `gpt-5.4`; the docs have since moved to `gpt-5.6` and batched `actions[]`, so the sample may lag the API.
- Same cost profile as Anthropic: one full screenshot per step for operator and judge; no HAR or trace unless the harness adds Playwright's own.

## Recipe
- To capture a browser scenario with a reviewable bundle, start at `packages/runner-core/src/responses-loop.ts` and `packages/replay-schema/src/index.ts`, then serve the bundle through `apps/runner`'s artifact routes.
