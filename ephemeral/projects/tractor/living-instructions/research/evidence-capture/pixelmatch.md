# pixelmatch

## Purpose
Dependency-free JavaScript pixel comparison library (also a small CLI) that returns the count of mismatched pixels and paints a diff image; the comparator Playwright's `toHaveScreenshot` uses.

## Pinned
- URL: https://github.com/mapbox/pixelmatch
- Version: v7.2.0 (released 2026-04-29)
- License: ISC

## Key concepts
- API: `pixelmatch(img1, img2, output, width, height, options)` on raw RGBA buffers; returns number of differing pixels (README "API").
- Options: `threshold` (0-1, default 0.1, YIQ colour distance), `includeAA` (default false), `alpha`, `diffColor`, `diffMask` (diff on transparent background), `windowSize` (max mismatches in any NxN window, for noise-resistant judging) (README "API").
- CLI: `pixelmatch image1.png image2.png output.png 0.1` (README "Command line").
- Artifact: diff PNG (image) plus a number (text). Both inputs must already be decoded to equal-size RGBA (pngjs or similar).

## Bounded comparison
Like odiff but only a JS library with a thin CLI, and it is what Playwright embeds, so its numbers match Playwright's `maxDiffPixels`.

## Gotchas
- Requires identical dimensions; it throws otherwise, so crop or fix viewport first.
- Pure JS: slow on full-page screenshots (odiff README benchmarks show about 6x).
- Anti-aliasing differences across hosts inflate the count unless `includeAA` stays false and `threshold` is tuned in the same container the baseline came from.

## Recipe
- To reproduce Playwright's verdict outside the test runner, start at README "API" with `threshold: 0.2` (Playwright's default) on the `-expected` and `-actual` PNGs.
