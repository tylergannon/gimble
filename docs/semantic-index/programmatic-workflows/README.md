# Programmatic-workflows semantic index

This is a compact routing tree for Gimble's Go-workflow direction, executable
stub examples, recovered library POC, and authoring sketches. For concrete Go
programs, start with [the examples](sources/compiling-examples.md); consult the
recovered implementation before proposing new runtime primitives. Go adoption
can coexist with the graph language while preserving workflow legibility.

## Scope and token cache

The token cache is the local Gimble repository checkout. The indexed corpus is
intentionally narrower: the edited direction documents, five-arts framing,
preserved statements, POC source/recovery notes, Go-library/concurrency
sketches under `ephemeral/projects/gimble/programmatic-workflows/`, and the
stubbed programs under `examples/go-workflows/`. Citations
are repository-relative `path:line` anchors and should resolve from the root.

The graph engine is shipped. A Go library POC exists on a recovered unmerged
branch, with its workflow source preserved here for inspection. Automatic
scoped context remains proposed. New examples express the concurrency shapes
with ordinary `errgroup`, but their agent/command replies and workspaces are
stubs, not native runtime integration. The newer Go-library API sketches need
revision after Tyler's indirection correction.

## Route order

1. For code/API design, start with [the programs, POC, and sketches](routes/go-library/index.md).
   Otherwise start with [recipes.md](recipes.md) for a task-shaped question.
2. Use [topics.md](topics.md) to route by concept or source authority.
3. Use [themes.md](themes.md) for cross-cutting relationships and tradeoffs.
4. Open the linked leaf under `sources/` for the exact source anchors.

The edited syntheses are the best starting points for the current framing.
The stub programs make the ordinary-Go authoring shape inspectable; the POC
records prior native implementation. The newer proposal is not an accepted
replacement API.
The verbatim leaves preserve Tyler's original wording and corrections; use
them when the distinction between the source claim and later synthesis matters.

## Layout

- `recipes.md` — task-first routes.
- `topics.md` — concept and authority routes.
- `themes.md` — cross-source synthesis.
- `routes/` — narrow topic nodes used by the task and concept routes.
- `sources/` — one dense leaf per indexed writing.
- `.semantic-index/evals.jsonl` — small retrieval query set.
- `.semantic-index/state.json` — generated housekeeping metrics.

## Housekeeping

- Build mode: incremental update after POC recovery.
- Last indexed source review: 2026-09-09.
- Benchmark: run `python3 <df-semantic-index>/scripts/run_evals.py --index docs/semantic-index/programmatic-workflows --write-results` from the repository root.
- Known debt: citations point to mutable working-tree documents; the preserved
  verbatim notes provide stable wording. The restored POC source and concurrency
  sketches identify their historical revision; the complete library requires
  the local backup described in the recovery guide.
