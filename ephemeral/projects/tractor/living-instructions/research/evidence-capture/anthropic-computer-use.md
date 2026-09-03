# Anthropic computer use

## Purpose
Claude's `computer` toolset lets the model drive a desktop or browser by screenshot and coordinates; every step's screenshot is an artifact the harness receives as base64 PNG and may persist for a later judge.

## Pinned
- Docs: https://platform.claude.com/docs/en/docs/agents-and-tools/tool-use/computer-use-tool
- Tool version: `computer_toolset_20260801` (current, no beta header); `computer_20251124` (beta header, older models).
- Reference implementation: https://github.com/anthropics/anthropic-quickstarts/tree/main/computer-use-demo — HEAD 3313e9716fb5 (2026-08-25), MIT.

## Key concepts
- Actions include `screenshot`, `zoom` (region at full resolution), clicks, `left_click_drag`, `scroll`, `type`, `key`, `hold_key`, `wait` (docs "Core actions"). Batch actions run sequentially and stop at first failure.
- The application returns screenshots as `tool_result` image blocks (`media_type: image/png`, base64) (docs "Screenshot handling"); the API never stores them, so persistence is the harness's job.
- Reference demo: Docker image `ghcr.io/anthropics/anthropic-quickstarts:computer-use-demo-latest` with Xvfb, Mutter, Firefox, Streamlit; ports 5900 (VNC), 6080 (noVNC), 8501, 8080 (demo README "Quickstart").
- In the demo, `computer.py` sets `OUTPUT_DIR = "/tmp/outputs"` (line 16), saves each screenshot to a file before base64-encoding it (lines 228-240), and rescales with ImageMagick `convert` to `MAX_SCALING_TARGETS` (line 58). A run therefore leaves a PNG per step on disk inside the container.
- Resolution guidance: 1024x768 or 1280x720 for desktop, 1280x800 for web; image limits 1568 px long edge / 1.15 MP on older models, 2576 px / 3.75 MP on Opus 4.7+ (docs "Screenshot resolution & scaling"). Coordinates are in screenshot pixel space; Retina must be downscaled by 2.
- Related on this machine: the Claude in Chrome MCP (`mcp__claude-in-chrome__gif_creator`, `read_network_requests`, `read_console_messages`) records a GIF and request logs from a live Chrome, but needs the extension and a headed browser.

## Bounded comparison
Like Playwright screenshots but only pixel-and-coordinate driven, with no DOM, HAR, or trace, and only what the harness writes to disk survives.

## Gotchas
- Cost: one image per step at roughly 1 MP each; a 40-step scenario is 40 vision inputs for the operator and again for the judge.
- Screenshots are downscaled before the model sees them; small text can be unreadable to a judge unless `zoom` captures are also saved.
- The demo's `/tmp/outputs` lives inside the container; mount a volume or copy out, or the evidence vanishes with the container.
- Flakiness comes from the desktop, not the API: window focus, animation timing, and the 2 s `_screenshot_delay` (computer.py line 99) after each action.
- No built-in video, trace, or request log; the action log is the `tool_use` JSON in the transcript.

## Recipe
- To capture a desktop scenario, start at computer-use-demo `computer_use_demo/tools/computer.py` `OUTPUT_DIR` and mount it, then keep the `tool_use`/`tool_result` transcript as the step index for the PNGs.
