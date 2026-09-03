# Tying off worktree-goal-gates

Working notes, started 2026-09-03. Branch `worktree-goal-gates`, worktree
`/Users/tyler/src/.worktrees/tractor/goal-gates`.

## The plan (Tyler, 2026-09-03)

1. Delete the ephemeral planning spiral (~25k lines under
   `ephemeral/projects/tractor/living-instructions/`).
2. Delete the `workflow` package: `workflow/`, `cmd/tractor/workflow.go`,
   `workflow_test.go`, `validate-plan`, the plan/medium/large YAMLs, the
   recommendation contract, docs that describe them (`planning.md`, parts of
   `loops.md`, `SKILL.md`, `llms.txt`, `spec.md` §3.12). They encode the
   misunderstanding: sizes were a heuristic to help the planner WRITE a
   workflow, not a router to three canned graphs.
3. Fix the loop node bugs below.
4. Squash-merge to main.

Keep: loop node, checklist package, frames, four lint rules, `tractor ask`
and `tractor answer`, `TRACTOR_RUN_DIR`, checklist-loop example, docs for
those.

## Loop node issues found in audit (2026-09-03)

1. **`done: true` on the framed item skips validation.** `engine/loop.go`
   Execute: `case item.Done: popFrame`. Prompt-only guard against the body
   agent flipping the flag. Test `TestLoopHonorsHandMarkedItemWithoutValidating`
   locks it in. Tyler: "This is fucking insane." See ruling below.
2. **Empty checklist validates as complete.** No open item → on_done. Nested:
   a chapter item with no command passes when its inner loop finds zero
   sprints. The "baby bear" reviewer (decision 14) was never built.
3. **`max_visits` on a nested loop is a whole-run budget.** Visits never
   reset per enclosing item. Inner loop `max_visits: 40` caps total inner laps
   across every outer item.
4. **Resume re-dispatches the outer body.** Rewind to the outermost loop
   re-runs whatever the outer body is (in the deleted LARGE shape, the
   chapter planner), which can duplicate items in an inner ledger.
5. **Smaller footguns.** `fs.Glob`: `**` not recursive, `./` prefix rejected,
   frame doesn't say so. Infer judge inherits the pipeline default model.
   Runner does `os.Setenv("TRACTOR_RUN_DIR")` process-wide.

## Rulings

### Issue 1 (Tyler, 2026-09-03) — filed as tylergannon/tractor#32

Validation is never skipped and never turned off. On every lap return:
run the validation of the item that just finished AND every item already
`done: true`. All pass → mark the finished item `done: true`. Any fail →
do not mark; re-enter the item with what failed (including a prior item
that regressed). `done: true` enables validation for that item on every
later lap; it never disables anything.

Cost of re-running every infer judge each lap: accepted, no per-item knob.
Tyler: "we need working software... If we spend more tokens on validation,
at least we will have working software that we can improve on."
Engine writes `done` both ways: a done item whose validation fails is set
back to `done: false`. `done: true` means only "passed on the last lap".

### Issue 2 (2026-09-03) — filed as tylergannon/tractor#33: the exit goal gate was dropped from v1

The memo §9a.2 goal gate ("the run cannot terminate at success until an
exit gate demonstrates the goal's claims") and the brief's "resurrected
goal gate at the exit" are not on the branch. loop-node.md §2: "body is
prose the engine never reads"; §1: no open item → on_done. Not listed in
§10 "Not in v1".

Proposed: when the loop finds no open item, run an evaluator turn (loop
node's model fields, like the infer judge) with the checklist body
(definition of done), the items with their last validation results, and
the workspace. Done → on_done. Not done → evaluator edits the checklist
(appends/rewrites items) as ordinary work; loop re-reads, selects first
open item. Not done and still no open item → terminal error. Empty ledger
on first arrival takes the same path. Awaiting Tyler's ruling.

### Issue 3 (2026-09-03) — filed as tylergannon/tractor#34

Ruled: a nested loop's `max_visits` counts arrivals per activation. On
frame push for a new enclosing item, reset the nested loop's visit count
and its body node-set's counts. Top-level loop unchanged.

### Issue 4 (2026-09-03) — filed as tylergannon/tractor#35, enhancement not bug

Tyler: re-running the outer body on resume is an inefficiency, not a bug;
no wrong result demonstrable. Filed as enhancement: rebuild frames from
checklist files on resume, resume at the innermost loop.

### Issue 5 (2026-09-03)

- Evidence globs: bug, filed as tylergannon/tractor#36. `**` recursive,
  strip `./`.
- Judge model: Tyler: "WRONG. judge model should be independently set.
  there should be a default model for judging screenshots. should be
  gemini 2.7 flash." Filed as tylergannon/tractor#37. Loop node model
  fields no longer inherit pipeline defaults; judge default = Gemini 2.7
  Flash via the agy harness.
- `os.Setenv("TRACTOR_RUN_DIR")`: not a bug, dropped.

## Issue index

| # | Title | Kind |
|---|---|---|
| #32 | validate every done item every lap; done never skips | bug |
| #33 | exit goal gate before on_done | bug |
| #34 | nested max_visits resets per enclosing item | bug |
| #35 | rebuild frames on resume | enhancement |
| #36 | evidence globs ** and ./ | bug |
| #37 | judge model independent, default Gemini 2.7 Flash | bug |

Still to do: delete `workflow/` + CLI + docs, delete the ephemeral spiral,
squash-merge.
