# Building the living instructions

Conventions for the agent executing a sprint or planning a chapter of this
build. Read this before the sprint doc.

## Who is who

- **You** are the coding agent for one sprint, in a fresh context. The
  iterate blocks at the top of your prompt say which chapter and sprint.
- **The reviewer** is Claude, managing the build. Questions go to the
  reviewer through `tractor ask`; the reviewer answers or relays to Tyler.
- **Tyler** owns the project. Decisions he ratified are in `decisions.md`
  beside this file; the loop node's design is in `loop-node.md`. Read the
  parts your sprint touches. Do not reopen decisions; ask if one blocks you.

## Paths in these documents

A path that starts with a top-level repository directory (`workflow/`,
`cmd/`, `docs/`, `src/`, `skills/`, `engine/`, `ephemeral/`) is relative
to the repository root, where every command runs. Any other path in a
chapter or sprint document (`research/...`, `chapters/04-library/...`,
`prove/...`, `decisions.md`) is relative to this directory,
`ephemeral/projects/tractor/living-instructions/`; prefix it when you
open or copy the file.

## The repository

Go module `github.com/tylergannon/tractor`. Before you finish a sprint:

```
go build ./... && go vet ./... && go test ./...
golangci-lint run ./...
```

If you change `graph/graph.go`, regenerate the schema with `go generate ./graph`
and commit the generated files.

Code style: small functions, errors wrapped with context, tests beside the
code. Match the surrounding package. No new dependencies without asking.

Commit at the end of the sprint with `git add -A` and a short imperative
subject line. Include interview files and ledger changes in the commit.

## Ledgers

`chapters.md` and each chapter's `sprints.md` are markdown files with YAML
frontmatter. The engine marks items `done: true` after their validation
passes. You never write `done`. A planning turn may append sprint items and edit open ones; an
implementing turn does not edit the ledger, and no turn edits the proof
tooling: `chapters/*/prove/`, `validation/observer.sh`,
`validation/lib/`, the `base-commit.txt` pins, or the doctrine pages
under `content/`. Those are the planner's; your segment records every
write, and a write there fails the sprint.

## Asking questions

The interview directory for this build is

```
ephemeral/projects/tractor/living-instructions/interview
```

Write the question as a file anywhere (markdown, or HTML when a rendered
page says it better), then run

```
TRACTOR_INTERVIEW_DIR=ephemeral/projects/tractor/living-instructions/interview \
  go run ./cmd/tractor ask <path-to-question-file>
```

The command moves the file into the interview directory under the next
number, records the question in the run's timeline, blocks until the
reviewer writes `<n>.answer.md` beside it, and prints the answer. If your
shell tool times out while waiting, run the same command again with the
moved path (`interview/<n>.md`); it resumes waiting without renumbering.

Sprint 1 of chapter 1 builds this command. Until it exists, proceed from the
docs and note in your commit message what you would have asked.

Ask when a sprint doc leaves a decision open, or when you would otherwise
guess at something the reviewer would care about. One question per file.
State the options and your recommendation. Do not ask about things the docs
already settle. Do not batch ten questions into one file.

## Not in scope, ever

Replayability of runs, per-item retry counters, fixup routing, a web UI for
the interview, anything under `## Rejected` in `decisions.md`.
