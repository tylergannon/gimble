# Local machine inventory (2026-09-02)

## Purpose
What evidence-capture tooling is already present on this Mac (macOS 26.5.2, Node 24.20.0 via vite-plus), recorded so a verifier knows what it can call without installing.

## Commands run
`which playwright npx agent-browser asciinema vhs ffmpeg chromium google-chrome`; `npx playwright --version`; plus inspection of `~/Library/Caches/ms-playwright`, `~/.npm/_npx`, `brew list`, `/Applications`.

## Present
- `npx`: /Users/tyler/.vite-plus/bin/npx
- `ffmpeg`: /Users/tyler/.local/bin/ffmpeg — broken: `dyld: Library not loaded: /opt/homebrew/opt/x264/lib/libx264.164.dylib`. Homebrew lists `ffmpeg 9.0.1_1` but the `~/.local/bin` shim shadows it and fails to start.
- Google Chrome: /Applications/Google Chrome.app, version 152.0.7977.65 (Info.plist). Not on PATH as `google-chrome`.
- Playwright browser cache (~/Library/Caches/ms-playwright): chromium-1200/1228/1234, chromium_headless_shell-1200/1228/1234, firefox-1532, webkit-2311, ffmpeg-1011 (Playwright's private ffmpeg for video).
- Playwright packages in the npx cache: `playwright` 1.61.0 and 1.60.0-alpha (two `~/.npm/_npx/*` entries), plus `@playwright/mcp` (`playwright-mcp` bin). Not installed globally; `npm ls -g` shows only corepack and npm.
- `gh`: /opt/homebrew/bin/gh; `claude` and `codex` CLIs in ~/.local/bin.
- Claude in Chrome MCP and a macOS computer-use MCP are attached to this session (headed only).

## Absent
- `playwright` (no global binary), `agent-browser`, `asciinema`, `vhs`, `ttyd`, `chromium`, `google-chrome` on PATH; `agg` not in brew.

## Notes
- `npx playwright --version` fails here with `EBADDEVENGINES Invalid devEngines.packageManager` because it runs from a directory whose package.json declares devEngines; run it from a scratch directory or pin `npx playwright@1.62.1`.
- Chromium build 1234 in the cache corresponds to Playwright 1.61.x; Playwright 1.62.1 will download a new build on first use.
- With ffmpeg broken, VHS cannot run; Playwright video relies on its own bundled ffmpeg-1011, which is present.
