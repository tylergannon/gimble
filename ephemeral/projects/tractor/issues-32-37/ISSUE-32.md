# Issue #32 — Loop node: validate every done item on every lap; done never skips validation

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

Branch `worktree-goal-gates`, `engine/loop.go`.

## Bug

On a lap return, the loop node validates the framed item only if its `done` field is not `true`:

```go
case item.Done:
    h.runner.popFrame(loop.ID)
```

Anything that writes `done: true` on the current item, including the body agent, skips the item's command and infer judge entirely. The only guard is a sentence in the frame preamble. `TestLoopHonorsHandMarkedItemWithoutValidating` locks the behavior in.

Also, an item is validated exactly once, on the lap it was implemented. Nothing checks that earlier items still hold.

## Required behavior

Validation is never skipped and never turned off. On every lap return:

1. Re-read the checklist.
2. Validation set = every item with `done: true` plus the framed item, in file order.
3. Run every validation in the set: command, then infer, per item. Record each result.
4. All pass: mark the framed item `done: true`, pop the frame, select the next open item.
5. Any fail: do not mark the framed item. Re-enter it with the full list of failures, including any prior item that regressed and the path to its validation log.

`done` reflects the last validation result and the engine writes it in both directions: an item in the set that fails is set back to `done: false`. `done: true` means only "passed on the last lap". A hand edit is honored for one lap, then tested.

## Changes

- `engine/loop.go`: replace the single-item validate with the set walk above. Loop `timeout` applies per command, not per arrival.
- `checklist`: add the un-mark operation (set `done: false`), same atomic rewrite and parse-back check as `MarkDone`.
- `validation.json` and the `LoopValidated` timeline event carry a list, one entry per validated item.
- Frame: `last validation` becomes a block listing each failure with its log path.
- Invert `TestLoopHonorsHandMarkedItemWithoutValidating` into a regression test: a hand-marked item that fails its command is un-marked and re-selected.
- Docs: `docs/spec.md` loop section, `src/content/docs/loops.md`, `loop-node.md` §4.

Re-running every infer judge on every lap is accepted. No per-item knob.
