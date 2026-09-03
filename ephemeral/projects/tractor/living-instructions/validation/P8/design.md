# P8: the library is content

Archetype: universal over the library's files. Exhaustive; no holdout.
Lap 4; answers `review-3.md`.

Reading of the promise: "prompt" means any text the library sends to an
agent, so files under `prompts/`, `supervisors/`, and `passes/` all
count as prompts for the orphan walk; the walk covers doctrine pages,
which P8 names, and not skeletons. `show` prints every node of the
graph (declaration section 3) and what `Build` materialized for each;
it never claims to reproduce frames.

## Story

Chapter 4 sprints 1 and 2; their proof scripts are the story. Sprint 3's
`prove/doctrine-pages.sh` demonstrates that sprint; it is not part of
P8's proof.

## Evidence

- The repository at chapter 4's end: `workflow/library/`,
  `workflow/library.go`, `workflow/testdata/`.
- The graph files `workflow/library/workflows/*.yaml`: the node ids and
  types, read by the check itself.
- `Build`'s output, captured by the check itself: a throwaway program
  the script writes into a copy of the tree, which calls
  `workflow.Build` and prints one node's prompt, command, or checklist.
- A mutated copy of the tree: every library prompt file and one
  doctrine page with a sentinel appended, one uncited doctrine page
  added.
- Output of `prove/prompts-are-library-files.sh` and
  `prove/show-and-orphan-walk.sh`.
- At chapter 6: a `plan` run the check makes itself, and its
  `stages/<seq>-<node>/prompt.md`.

## Validator

`command`: the two sprint scripts, then at chapter 6
`prove/p8-show-stage.sh`.

- `prove/prompts-are-library-files.sh`: library files exist; the
  planner's opening sentence is absent from Go; `go:embed` and
  `text/template` are used; the snapshot files exist and
  `TestBuildMatchesSnapshot` runs and passes. This script is the
  sprint's tripwire; the proof of content location is the next one.
- `prove/show-and-orphan-walk.sh`, three parts. Nodes: the node list
  the headed `show` prints for each workflow equals the id and type
  list the script reads from the workflow's YAML (an omitted node is a
  fail). Content and equality: in a copy of the tree the script builds
  its own `Build` dumper; `show --raw` for every node equals the
  dumper's output for the same parameters (bound to `Build`, not to a
  coder-owned snapshot or test); then with sentinels appended to every
  file under `prompts/`, `supervisors/`, and `passes/` and to one
  doctrine page, every codergen node's `show --raw` prints its file's
  sentinel and every prompt citing the page prints the page's sentinel
  (a prompt still living in Go cannot pass); `--stage` on a stage built
  from the frame preamble plus the dumper's output exits 0, and 1 after
  a byte is appended. Orphans: in the copy, `doctrine/zz-uncited.md`
  makes `go test -run TestLibraryNoOrphans` fail naming it; in the real
  tree `TestLibraryRendersAll` and `TestLibraryNoOrphans` run and pass.
- `prove/p8-show-stage.sh`: runs `plan` on `seeds/greeter.md` into a
  fresh run directory, then for the first stage of each codergen node
  runs `show plan --node <n> --stage <that stage dir>` with the
  parameters the script passed to `run`, and expects exit 0.

No `infer`.

## Not proven

That the content is good. That `show` reproduces frames (by design;
research F1). That the sprint 1 snapshot stays honest after content
edits; it is the coder's tripwire, and the check does not rely on it.
