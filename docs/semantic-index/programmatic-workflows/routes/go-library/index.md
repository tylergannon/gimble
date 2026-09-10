# Route: Go programs, existing library, and authoring sketches

Use this route before proposing Go API primitives, translating a builtin,
or looking for the POC that disappeared from the active branch list.

1. [Go programs with stubbed effects](../../sources/compiling-examples.md): start
   here for actual source and invocation commands. Bake-off and critique use
   ordinary `errgroup`; sprints use a Go iterator and explicit review/repair.
   These new examples use canned replies and symbolic workspaces.
2. [POC recovery and status](../../sources/poc-recovery.md): preserved branch,
   commits, full-source recovery, and implemented versus proposed behavior.
3. [Actual workflow source](../../sources/poc-workflows.md): the implemented
   iterator-based sprint/chapter/delivery routines after prompt/input extraction.
4. [Earlier concurrency sketches](../../sources/concurrency-shapes.md): ordinary
   `errgroup` critique circles and bake-offs, typed outcomes, and scoped context.
5. [Newer Go-library sketches](../../sources/go-library-sketches.md): additive
   roadmap, Polytype arguments, roles, and a curated CLI catalog. Tyler rejected
   the API's indirection; revise it against the earlier work, not vice versa.

For `SetContext`, inline versus external data, aggregate prompt budgets, and
waiting for a coherent index, use [filesystem-backed context](../../sources/context-files.md).
That example writes real context files while agent operations remain stubbed.
For automatic chapter/sprint contexts and arbitrary parallel scopes, follow
[context scopes and physical layers](../../sources/context-scopes.md).

For general rationale use [Go authorship](../go-authorship/index.md); for
current shipped behavior use [status and contract](../status-and-contract/index.md).
