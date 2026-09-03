# Chapter 4: the library

Status: active

Everything an agent is told by a built-in workflow becomes a file under
`workflow/library/`, embedded in the binary and rendered with
`text/template`. After this chapter, every later chapter adds content
and `workflow`-package Go, never engine code. Promise P8.

Paths in this chapter's documents follow `BUILD.md`: a path that does
not start with a top-level repository directory is relative to
`ephemeral/projects/tractor/living-instructions/`.

## Pyramid index

- L0: The built-in workflows' prompts, doctrine, and skeletons are
  editable embedded content with a render test, an orphan walk, and a
  `show` command.
- L1:
  - The three graphs and four prompts move out of Go strings into
    `workflow/library/` with no change to what `Build` returns.
  - `tractor workflow show <name>` prints what `Build` materialized;
    `--stage` diffs a real stage minus its frame; tests render every
    template and fail on an orphaned doctrine page.
  - Doctrine pages and artifact skeletons the current prompts can cite,
    written by Claude, wired by the coder.
  - Docs, spec, skill bundle, and `llms.txt` teach the library and `show`.
- L2: sprints 1 to 5 in `sprints.md`; anchors in
  `research/workflow-package-inventory/migration-inventory.md`.

## Vector

Decisions 53 and 59 (54, `models.yaml`, is chapter 5's). Research R1 (prior art: crush, goose, codex,
gemini-cli) and R5 (the exact seams). Go supplies data values, the
value functions `quote` and `shell`, and the composition actions
`include` and `doctrine` to templates, and nothing else; provider names are explicit; delimiters are
non-default because doctrine contains `{{`.

## Review posture

`Build`'s outputs are pinned by the snapshot test; in sprint 1 a change
to what a prompt says is a failed refactor, and from sprint 4 on a
change is a reviewed diff of the snapshot, committed with the content. The orphan walk is written
from scratch (no surveyed tool has one) and must be proven to fail on a
synthetic orphan.

## Non-goals

New nodes, new prompts for v2, `models.yaml` (chapter 5), any engine
change, byte-equality with the engine's `prompt.md` (impossible: research
F1).
