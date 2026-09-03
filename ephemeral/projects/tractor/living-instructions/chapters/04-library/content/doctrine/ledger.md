# Ledgers

A ledger is a Markdown file with YAML front matter. The front matter is
the ledger; the body is prose the engine never reads. The engine
iterates the items, runs each item's validation when its lap returns,
and writes exactly one thing: `done: true` on one item.

Item fields:

- `name`: identity, unique in the file.
- `check`: the claim as observable behaviour. The promise, in prose.
- `command`: shell, run from the workdir; exit 0 passes.
- `infer`: `files` (globs) and `prompt`; a judge on another model reads
  the files and decides.
- `doc`: a prose document injected with the item.
- `checklist`: a sub-ledger for a nested loop.
- `done`: engine-owned. Absent or false means open.

Rules you must keep:

- Never write `done`. A hand-edited `done: true` is honored without
  validation; it is the human's override, not yours.
- An item with neither `command` nor `infer` passes when its lap
  returns. Use that only for items whose validation is their own
  sub-loop (a chapter) or their own review edge; never at the bottom.
- Items run in file order. A failed item stays open and is re-selected
  unless a planner inserts before it.
- Planners may append, insert, reorder, or rewrite open items, and
  only open items. A done item is history.
- The chapter ledger is durable: edited only with a reason, by the
  human or the planner, and the reason is recorded. Sprint ledgers are
  re-planned after every lap.
- Paths in the file are relative to the workdir.

Name a proof script for what it proves (`prove/show-equals-build.sh`),
never for the sprint that ran it.

Source: loop-node.md section 2; decisions.md 15, 44.
