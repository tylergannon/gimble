# odiff

## Purpose
Standalone pixel-diff binary that compares two screenshots and writes a diff PNG, for judging visual drift without a JS runtime.

## Pinned
- URL: https://github.com/dmtrKovalenko/odiff
- Version: v4.5.0 (released 2026-07-23); npm package `odiff-bin`
- License: MIT

## Key concepts
- CLI: `odiff base.png compare.png diff.png`; diff output is optional and must be PNG; without it, the diff is drawn inline in Ghostty/iTerm2/kitty/WezTerm (README "Usage").
- Options: `--threshold` (0-1 colour sensitivity), `--antialiasing` (ignore AA pixels), `--fail-on-layout` (reject different dimensions), `--diff-mask`, `--diff-overlay`, `--ignore-regions x1:y1-x2:y2`, `--output-diff-lines` (README "Options").
- Inputs: PNG, JPEG, WebP, TIFF, cross-format allowed (README "Supported formats").
- Performance: "6 times faster than imagemagick and pixelmatch" on full-page screenshots; written in Zig (README "Benchmarks").
- Artifact: diff PNG (image, judgeable) plus exit code (text); `--output-diff-lines` prints changed row coordinates for a text judge.

## Bounded comparison
Like pixelmatch but only a native binary, with layout-mismatch failure and ignore-regions built in.

## Gotchas
- Different dimensions are a hard failure with `--fail-on-layout`; a verifier that changes viewport between runs produces no diff, only an error.
- Threshold semantics differ from pixelmatch's default 0.1 and Playwright's 0.2; do not reuse numbers across tools.
- The diff PNG shows where, not why; pair it with the two source screenshots for the judge.

## Recipe
- To diff two scenario screenshots, start at README "Usage": `odiff --antialiasing --fail-on-layout before.png after.png diff.png`, keep all three files.
