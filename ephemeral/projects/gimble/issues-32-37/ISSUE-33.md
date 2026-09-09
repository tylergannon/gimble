# Issue #33 — Loop node: an evaluator turn decides what happens after every passing lap

Task brief for the implementer. The text below is the issue verbatim,
as rewritten 2026-09-03 after Tyler's rulings.

## Definition of done

- The required behavior below holds in the code.
- The tests named below exist and pass.
- The docs named below say what the code now does.
- `go build ./... && go test -count=1 ./...` exits 0.
- The change is made in this checkout, at the path the run's workdir names.
  Do not run `git worktree add`, do not create a branch, do not commit, and do
  not push: the run's verification gate executes in the workdir, so work done
  anywhere else is not checked. This overrides any worktree or checkpoint step
  in the repository's agent protocol.

---

`engine/loop.go`.

## Bug

When a lap passes, the engine marks the item done and selects the next open item deterministically: first not-done in file order. No judgment happens there. When no open item remains it routes to `on_done` immediately. Nothing ever reads the checklist's markdown body, which is the prose definition of done.

Consequences: an empty ledger completes on first arrival; a ledger whose items were all demonstrated completes even when the definition of done is not met; and a plan that turns out to be wrong at lap two is not noticed until it has been worked to the end.

## Required behavior

The checklist file is markdown with YAML frontmatter: the body is the definition of done, the frontmatter is the ledger. The evaluator, not the ledger, decides what happens next.

One evaluator turn runs **after every lap whose validation passes**, and on any arrival that finds no open item (including an empty ledger on the first arrival). It does not run when validation fails; that lap re-enters its item per #32.

Inputs: the checklist body, the items with their last validation results, and the workspace.

The evaluator may edit the checklist as ordinary work: append, reorder, or rewrite open items. In the ordinary case it simply agrees that the next configured item is the right next step and changes nothing.

Verdicts:

- **done**: route to `on_done`.
- **not done**: the loop re-reads the file and selects the first open item. Next lap.
- **not done with no open item in the file**: terminal error.

The evaluator does **not** use the infer judge's model slot. It gets its own selection and defaults to the pipeline's default model, the same one the working agent runs on. The judge's slot and its Flash default are unchanged (#37).

The per-lap turn is expected to cost wall time. That is accepted; being able to be agile on the plan is the point. No per-loop knob for now.

## Not this

The evaluator is one harness call, not a programmable sequence of nodes. Ruled 2026-09-03.

An evaluator built from ordinary nodes already works at the loop's *exit*: point `on_done` at a node whose edges route to `success` or back to the loop, and the loop re-reads the checklist and picks up whatever it appended. Demonstrated with tool nodes and no models. It cannot serve the per-lap case, because `loop_body_exit` forbids a body node from routing to `success` or outside the body.

## Changes

- `engine/loop.go`: run the evaluator after a passing validation set and on the no-open-item branch, before routing to `on_done`.
- Evaluator prompt and a choice schema with two targets, `done` and `not_done`, like the infer judge.
- Its own model fields, defaulting to the pipeline default model. See #40: its prompt and response files must not collide with the judge's.
- Timeline event for the verdict.
- Tests: empty ledger gets items written and the loop continues; a passing lap where the evaluator agrees proceeds to the next item unchanged; an evaluator that reorders open items changes what runs next; all-items-done with a not-done verdict adds an item; not-done with no open item is a terminal error; done routes to `on_done`; the evaluator runs on the pipeline default model and not the judge's Flash default.
- Docs: `docs/spec.md` loop section, `src/content/docs/loops.md`, and §4 of `ephemeral/projects/tractor/loop-node/loop-node.md`.
