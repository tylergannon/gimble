# Nested and arbitrary context scopes

```sh
go run ./examples/go-workflows/scopes
```

`Chapters` and `Sprints` supply an item's `Context` automatically. Its index
contains inherited goal data plus current chapter, sprint, and attempt metadata.
The workflow passes that context to short agent calls; it never copies those
fields into a prompt.

The first sprint also creates two ordinary child scopes in an `errgroup`.
Each owns a different `review_focus` value. The chapter owns `chapter_focus`;
the root's `focus` remains visible throughout. After the group joins, a call with
the original sprint context has no reviewer-local values. After the inner
loop, the chapter context has no sprint metadata; after the outer loop, the
root has neither chapter nor sprint metadata. No pop or cleanup call changes
which scope is active: the caller chooses the context it passes.

After each sprint, the chapter updates its own `chapter_focus`. The next
sprint receives that update: scope creation keeps a parent relationship, and
each agent entry resolves current ancestor values. `Call.Context` is the fixed
snapshot received by that call; parent changes do not alter a running call's
prompt or files.

Each scope owns a private directory with a readable scope-name prefix for
value files, indices, and revision views. `ContextSnapshot.View` contains
`values/<key>.json` symlinks for all
effective values, including inherited values in ancestor layers. Its `index.json`
links to the authoritative `ContextSnapshot.Index`. Native tools can list and
read the effective context through those ordinary paths, without a FUSE mount.
The complete view is prepared before the agent callback receives its prompt.

Context files are read-only by convention. `DeclareContext` binds a key to its
owner before assignment; `SetContext` only accepts that owning scope. Attempts
to write an inherited key fail even before the parent assigns its first value.
The same display name in another scope denotes a separate binding, and both
remain visible. Any collision qualifies ALL effective keys with scope IDs to
avoid aliases; use the current index keys. Same-kind loops have distinct bindings.
The next snapshot publishes a new view pointing to the replacement file,
leaving earlier views intact. Direct writes through symlinks would modify their
targets; these links do not implement automatic copy-on-write.

Stdout captures exact prompts, scope paths, index/view locations, and decoded values
for every agent call. Context files are real and retained under the temporary
directory printed on stderr. Agent replies and command checks are canned;
scripted build completion drives the checks without executing software.
Parallel stage ordering may vary. Generated reports belong outside this repo.

Read `main.go` for orchestration, `input.go` for the typed nested input, and
`fixtures.go` for task data. `demo.go` contains thread-safe recording and the
scripted checks. `-example` prints JSON input; `-input file.json` replaces it.
