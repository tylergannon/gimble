# VHS (charmbracelet)

## Purpose
Scripts a terminal session from a `.tape` file and renders it through ttyd in a headless Chrome into GIF, MP4, WebM, PNG frames, or single PNG screenshots.

## Pinned
- URL: https://github.com/charmbracelet/vhs
- Version: v0.11.0 (released 2026-03-10)
- License: MIT
- Runtime deps: ttyd (https://github.com/tsl0922/ttyd, 1.7.7, MIT), ffmpeg on PATH, a Chrome/Chromium found by go-rod `launcher.LookPath()` (vhs.go lines 139-141; go.mod `github.com/go-rod/rod v0.116.2`).

## Key concepts
- Tape DSL: `Output demo.gif|out.mp4|out.webm|frames/` (frames/ is "a directory of frames as a PNG sequence"); `Set Width/Height/FontSize/Theme/Framerate`; `Type`, `Enter`, `Sleep`, `Ctrl+C`; `Screenshot path.png` (README "Screenshot"). Artifacts: PNG (image, judgeable), GIF/MP4/WebM (video).
- `Wait`, `Wait+Screen /regex/`, `Wait+Line /regex/`, `Wait@10ms /re/`; "Default regex is `/>$/`, timeout is `15s`, and default scope is `Line`" (README "Wait"). This is the determinism lever: block on screen text instead of `Sleep`.
- `Require gum` fails early if a program is missing (README "Require"); `Env KEY "VALUE"`; `Set Shell fish`.
- `vhs record` produces a tape from a live session; `vhs publish` uploads to vhs.charm.sh (do not use for private evidence).
- CI: Docker image `ghcr.io/charmbracelet/vhs`; GitHub Action `charmbracelet/vhs-action` (README "Installation").
- Rendering is a real xterm.js terminal in Chrome, so TUIs (alternate screen, colours, box drawing) appear as the user would see them.

## Bounded comparison
Like asciinema but only script-driven from a tape, and it renders pixels through a headless Chrome instead of saving the byte stream.

## Gotchas
- Three external binaries (ttyd, ffmpeg, Chrome); on this machine ffmpeg is present but broken (see local-machine.md). When no Chrome is found, go-rod's launcher downloads a Chromium into its cache; that is rod's default, not documented by VHS, so verify on a clean CI box.
- `Sleep`-based tapes are timing-flaky under CI load; prefer `Wait+Screen`. `Wait` times out at 15 s by default and aborts the tape.
- Containers need `VHS_NO_SANDBOX=1` (vhs.go line 140) or Chrome refuses to start as root.
- `Output frames/` at default framerate produces hundreds of PNGs per minute; use `Screenshot` at named steps for judge-sized evidence.
- Font rendering differs between hosts; PNG baselines are only comparable within one container image.
- The tape is the only record of what was typed; keep it next to the outputs as the scenario's script.

## Recipe
- To capture a TUI or CLI scenario as judgeable images, start at README "Screenshot" and "Wait": after each `Type`/`Enter`, add `Wait+Screen /expected text/` then `Screenshot stepN.png`.
- To run it in CI, start at README "Installation" (`ghcr.io/charmbracelet/vhs` image) with `Require` lines for every binary the scenario needs.
