# Issue #36 — Loop node: evidence globs support ** and ./ prefixes

Task brief for the implementer. The text below is the issue verbatim.
It was filed against branch `worktree-goal-gates`, which is now merged;
the code it names is on `main` at the same paths.

## Definition of done

- The required behavior below holds in the code.
- The tests named below exist and pass.
- The docs named below say what the code now does.
- `go build ./... && go test -count=1 ./...` exits 0.
- The change stays in the working tree. Do not commit, branch, or push.

---

Branch `worktree-goal-gates`, `engine/loop.go` `matchEvidence`.

## Bug

`infer.files` globs are expanded with Go's `fs.Glob`. Two consequences the frame never tells the agent about:

- `**` matches one directory level, not recursively. `captures/**/*.png` misses every file more than one level down and the item fails with "no evidence files matched".
- A leading `./` fails `fs.ValidPath`, so `./captures/*.png` fails the item with "invalid evidence pattern".

Both are the syntax an agent will write. `infer.files` exists (decision 12) to hand the judge the evidence files for an item a shell command cannot prove; a pattern that silently matches nothing defeats it.

## Required behavior

- `**` matches zero or more directories.
- A leading `./` is stripped before matching.
- Absolute patterns and patterns escaping the workdir stay invalid.

## Changes

- `matchEvidence`: use a `**`-capable matcher (e.g. `github.com/bmatcuk/doublestar/v4` over `os.DirFS(workdir)`); normalize `./`.
- `TestMatchEvidence`: add a recursive case and a `./` case.
- Docs: `docs/spec.md` and `src/content/docs/loops.md` state the glob syntax.
