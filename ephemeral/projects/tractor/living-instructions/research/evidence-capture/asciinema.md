# asciinema (with agg)

## Purpose
Records a terminal session's byte stream with timing into an asciicast text file; `agg` renders that file to a GIF so a vision model can judge it.

## Pinned
- URL: https://github.com/asciinema/asciinema — v3.2.1 (released 2026-06-16), GPL-3.0, Rust rewrite since 3.0.
- URL: https://github.com/asciinema/agg — v1.9.0 (released 2026-05-29), GPL-3.0.
- Format spec: https://docs.asciinema.org/manual/asciicast/v3/

## Key concepts
- asciicast v3 is NDJSON: a JSON header (`version: 3`, `term: {cols, rows}`, `timestamp`, `idle_time_limit`, `command`) then events `[interval, code, data]` with codes `o` output, `i` input, `m` marker, `r` resize, `x` exit (spec "Header", "Event stream"). Artifact: text, but `o` data is raw ANSI.
- `asciinema rec --command "..." --headless --window-size 80x24 --idle-time-limit 2 --overwrite --return out.cast` (src/cli.rs `Record` struct). `--headless` is "enabled automatically when running in an environment where a terminal is not available"; `--return` propagates the recorded command's exit status.
- `-f txt` or a `.txt` path writes a plain-text log instead of asciicast (cli.rs `Record`, `--output-format`). `asciinema convert in.cast out.txt` converts an existing recording.
- `--rec-env` captures named environment variables into the header for auditing; `-I` captures keyboard input (passwords included).
- agg: `agg in.cast out.gif [--theme ... --font-size ... --idle-time-limit ... --fps-cap ... --last-frame-duration]`; "No ffmpeg or browser needed" (agg README). Artifact: GIF, judgeable by a vision model frame by frame.

## Bounded comparison
Like a screen recorder but only for the terminal byte stream; nothing is rendered to pixels until agg or a player replays it.

## Gotchas
- A `.cast` of a TUI (alternate screen, cursor addressing) is escape-sequence soup; a text judge cannot reconstruct the final screen from it. Render with agg, or use the `.txt` format and confirm on a sample that it reflects what the user would have seen.
- Interval timing depends on host load; two recordings of the same command differ in every interval, so byte-compare of `.cast` files is never a valid diff.
- `.zst` output (about 8% of original size per README) is not readable by a judge without decompression.
- GPL-3.0 on both tools; fine to run as a subprocess, relevant if code is linked or vendored.
- Recording input (`-I`) leaks secrets typed during the scenario into the evidence.

## Recipe
- To capture a CLI scenario, start at cli.rs `Record` flags: `asciinema rec --headless --return --window-size 100x30 --command "<cmd>" out.cast`, then `agg out.cast out.gif` for the judge.
