# Sprint 3: the render test and the orphan walk

Two tests that make the library safe to edit, and the proof that they
bite. Exercisable at the end: add an uncited doctrine page and `go test
./workflow/` fails naming it; break a template action and it fails
naming the file.

## Read first

- `workflow/library.go` and the `BuildFrom(fs.FS)` seam from sprint 1.
- `chapters/04-library/prove/lib.sh` (the shared setup the proof
  scripts source) and `prove/show-equals-build.sh` (sprint 2).

## The work

`TestLibraryRendersAll`: for every workflow in the library, `Build`
with representative parameters and assert every prompt-bearing node's
prompt is non-empty and every template under `prompts/` was executed at
least once. Also execute every file under `doctrine/` and `templates/`
standalone with the same data, so a syntax error in a page fails here,
not at run time, and the failure names the file.

`TestLibraryNoOrphans`: walk the embedded tree; every file under
`doctrine/` must be named by a `doctrine` action in at least one file
under `prompts/`, `supervisors/`, or `passes/` that some node renders.
All three directories are prompts in P8's sense: text the library sends
to an agent. Skeletons under `templates/` are not in the walk; P8 does
not promise that every skeleton is used, and sprint 4's script checks
its own skeletons are cited. The failure names the orphan. Prove the
walk fails: the test constructs a synthetic `fs.FS` with one uncited
page and asserts the walk reports it by name.

The walk is written from scratch; no surveyed tool has one
(`research/prompt-libraries/goose.md` has the inverse and still drifted).

## Definition of done

The ledger item's command runs `prove/orphan-walk-and-render.sh`, which
proves, in a copy of the tree it makes itself: every rendered line of
every node is library text plus data values (each file of the node's
include closure is rendered standalone through `workflow.Render` by the
script's own program; with every data value form removed, each line of
the node's output equals some line of those renderings); with a distinct sentinel appended to every file
under `prompts/`, `supervisors/`, `passes/`, and `templates/`, each node
renders the sentinel of its header file and no sentinel of a file
outside that file's include closure (a conditional include may stay
silent), and every template file's sentinel appears in some node's
output; an injected uncited page, named at random, makes `TestLibraryNoOrphans`
fail naming it, and so does removing every citation of an existing page
chosen at random; an unclosed action with random text appended to a
doctrine page chosen at random (using the delimiters the README states)
makes `TestLibraryRendersAll` fail naming the page; when the library has
no doctrine page yet, the script plants one, cited from a prompt, so
every probe runs; and in the real tree both
tests run by name with `-v` and their `--- PASS:` lines are present.

## Not in this sprint

`show` itself (sprint 2). Doctrine pages (sprint 4). Docs (sprint 5).
