# Go workflow route

Start here for the Go API.

- [POC workflow source](../../sources/poc-workflows.md): the 88-line orchestration
  file. `program.Loop` owns validation and done; ordinary Go owns implement,
  review, repair, and nesting. This is the approved shape, ported into `program/`.
- [Recovered POC](../../sources/poc-recovery.md): the full library and its provenance.
- [Concurrency sketches](../../sources/concurrency-shapes.md): critique circles and
  bake-offs with ordinary `errgroup`, written inline in the workflow. There is
  no `workflows.BakeOff` and no worktree primitive; see AGENTS.md, "No wrappers".

The compiling stub examples that once sat above these were removed on 2026-09-10.
They had accreted `Sprints`, `Chapters`, `Scope`, and context calls through a
review loop and no longer matched the approved shape.
