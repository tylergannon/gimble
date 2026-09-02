# Loop node: live proof (claim 7 of loop-node.md §11)

Second run, after the adversarial-review fixes. The first run (HEAD
`c1378a3`, run `26e8e41b`, nonce `2d171d08`) behaved identically; its
artifacts were replaced by this run's.

A real run of the new `loop` node through the Claude harness, against a
scratch git repository seeded with a three-item checklist. This directory
holds the evidence copied out of the run directory and the workspace.

## Identity

| | |
|---|---|
| Branch / HEAD | `worktree-goal-gates` at `0643b55a1186334248b23049bafdf2030e4ea0a8` |
| Binary | `go build -o tractor ./cmd/tractor` at that HEAD; SHA-256 `8a4dbe31d645a3b47f22ba695ea20ab29707827ffd2eacaf1b2edcc1e355488a` |
| Run id | `c4cdcdecb0fdb70466f4740f51027f4c` |
| Started / duration | 2026-09-02T17:58:06Z, 41.1 s |
| Pipeline | `pipeline.yaml` here (the `checklist-loop.yaml` example with `max_visits: 8`) |
| Model | claude-sonnet-5, fidelity none, reasoning low, for both body and judge |
| Result | `COMPLETED`; exit 0 |

## Unpredictable input

The checklist was seeded with a nonce minted from `/dev/urandom` at seed
time (`GREETING_10ae02e8`) so the agent could not satisfy the first
command from prior knowledge. The second command compares `count.txt`
against a value recomputed from `greet.sh` at validation time. The third
item has no command and is judged by a model reading `notes.md`.

## What happened

`timeline.jsonl`, loop events only:

```
LoopItemSelected  greet script  1/3 lap 1
LoopValidated     greet script  passed
LoopItemSelected  line count    2/3 lap 1
LoopValidated     line count    passed
LoopItemSelected  notes         3/3 lap 1
LoopValidated     notes         passed
LoopCompleted     3
PipelineCompleted 41.13s
```

Seven stages: `items, implement, items, implement, items, implement,
items`. Three laps, no retries, no failed validation.

## Claims demonstrated

1. **The engine marks items; the agent never does.** Every `tool_call`
   event across the three body turns was audited (`Write` of `greet.sh`,
   `count.txt`, `notes.md`; `Bash` checks; one `cat` that included
   `checklist.md`). No agent call wrote to the checklist. Yet
   `checklist.after.md` has `done: true` on all three items, and its
   markdown body is byte-identical to the seed.
2. **Validation ran as a command, in the run directory.**
   `stages/000003-items/validation.json` and `000005-items/validation.json`
   record the exact command, exit 0, and an empty log tail.
3. **The infer judge worked from the evidence, not the report.**
   `stages/000007-items/prompt.md` is the engine-built prompt;
   `judge-events.jsonl` shows the judge `cat` all three files itself
   before answering (in the first run it also ran `sh greet.sh` and
   `wc -l greet.sh`); `response.md`
   routes `next: pass` with the recomputed values in its notes.
4. **The frame reached the agent.** `lap2-implement-prompt.md` is the
   second body prompt: the `<tractor loop="items" ... item="2/3" lap="1">`
   block with the item's name, check, and command, then the two-line
   authored prompt.
5. **Nothing loop-related is in the checkpoint.** `checkpoint.json` holds
   only the standard fields; `frames.json` was `[]` at completion.

## Scope check: what this run did not prove

- The failure path (a command exiting nonzero, `last validation: failed` in
  the next prompt, re-entry of the same item). Unit-tested in
  `engine/loop_test.go`, not exercised live.
- A judge returning `fail`, and "no evidence files matched". Unit-tested
  with a scripted backend only.
- Nested loops with item-level `checklist:` inheritance. Unit-tested only.
- `--resume` after a crash mid-lap. Not exercised.
- Any provider other than Claude.
