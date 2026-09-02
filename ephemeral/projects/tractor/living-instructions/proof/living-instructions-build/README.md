# The living-instructions build: proof by dogfood

Tractor built chapters 1 to 3 of its own living instructions with the loop
node, nested two deep, with GPT (`gpt-5.6-sol` through the codex backend) as
the coding agent and Claude answering the coding agent's questions through
the `tractor ask` command that the first sprint of the run created. This
directory holds the evidence copied out of the run directory. The ledgers,
sprint docs, check scripts, and interview files it refers to are committed
beside it under `ephemeral/projects/tractor/living-instructions/`.

## Identity

| | |
|---|---|
| Branch / HEAD at start | `worktree-goal-gates` at `358c164` (ledgers and pipeline committed) |
| HEAD at end | `a647731` (19 commits later; 116 files, +6047 −127) |
| Engine binary | `go build -o tractor ./cmd/tractor` at `358c164`; SHA-256 `6d95876b9e7ec9619d1e4fdda52bbcf5c22431c8513478a4c3a0e8602033a119` |
| Pipeline | `../../pipeline.yaml`: `chapters` loop → `plan` codergen → `sprints` loop → `implement` codergen |
| Run id | `de53a63f46b8519fe990610ffed94b44` |
| Started / ended | 2026-09-02T19:35:06Z → 21:51:26Z, 2h 16m 20s |
| Coder | codex backend, `gpt-5.6-sol`, reasoning high, fidelity none (fresh context per lap) |
| Judge | same model, `infer` items only |
| Result | `PipelineCompleted`, next `success`; `frames.json` = `[]` |
| Gates at end | `go build`, `go vet`, `go test ./...` (14 packages ok), `golangci-lint run` clean |

The engine binary predates sprint 2 of chapter 1, so it did not export
`TRACTOR_RUN_DIR`. Consequence noted under "Scope check".

## What happened

`timeline.jsonl`, loop events only. 40 stages, 18 agent turns, 18
validations, 4 loop completions.

```
chapters  Ask and answer                 selected 1/3
  sprints ask and answer commands        pass
  sprints run directory reaches agents   pass
  sprints docs and skill                 pass
chapters  Ask and answer                 pass
chapters  Planning workflow              selected 2/3   (plan node wrote 4 sprints)
  sprints embedded plan workflow         pass
  sprints workflow CLI and handoff       pass
  sprints planning docs and skill        FAIL  judge: planning.md contradicts loops.md
  sprints planning docs and skill        pass  (lap 2)
  sprints live planning proof            FAIL  exit 1: inspector rejected the planner's checklist command
  sprints live planning proof            pass  (lap 2)
chapters  Planning workflow              pass
chapters  Execution workflows            selected 3/3   (plan node wrote 6 sprints)
  sprints execution-ready planning artifacts   pass
  sprints embedded MEDIUM workflow             pass
  sprints embedded LARGE workflow              pass
  sprints execution workflow CLI and run dirs  pass
  sprints execution workflow docs and skill    pass
  sprints live plan-to-MEDIUM proof            pass
chapters  Execution workflows            pass
LoopCompleted chapters 3 → success
```

Twelve interview questions were asked through `tractor ask` and answered
through `tractor answer` (`../../interview/0001.md` to `0012.md`, each with
its `.answer.md`). Three were asked by implementing sprints, nine by the two
plan nodes. Every answer was given by Claude from `decisions.md`; none was
relayed to Tyler.

## Claims demonstrated

1. **Nested loops drive real work.** Two loop levels, three chapters, 13
   sprints. Inner loop resolved each chapter's ledger from the outer item's
   `checklist:` field (chapters 2 and 3 had no ledger until their plan node
   wrote one). `validations/000010-chapters.json`, `000025-chapters.json`,
   `000040-chapters.json` are the chapter-level validations; each ran only
   after its sprint loop emitted `LoopCompleted`.
2. **The engine marks; agents plan.** `summarize.py`-style audit of all 23
   event logs: the only agent writes to any ledger are the two plan nodes
   appending sprint items to `02-planning/sprints.md` and
   `03-execution/sprints.md`. No agent wrote `done`. All 16 `done: true`
   flags across the three ledgers are engine-written (commit `a647731` and
   the per-sprint commits carry them).
3. **The failure path works live.** Twice. `validations/000018-sprints.json`
   is a judge `fail` with notes; `lap2-docs-prompt.md` shows those notes
   arriving in the next lap's `<iterate>` block as `last validation:
   failed — judge: ...`; `000020-sprints.json` is the pass. `000022-sprints.json`
   is a command exit 1; `000024-sprints.json` the pass.
4. **`tractor ask` works from inside a run, on three backends' worth of
   plumbing.** Sprint 1 built it and used it in the same lap (questions 1
   and 2). Sprint 2's check (`01-ask/check-sprint-02.sh`) ran a nested
   pipeline whose codex agent asked for a nonce and wrote it back;
   `validations/000007-sprints.json` records `check-sprint-02: ok`.
5. **The planning workflow plans.** `validations/000024-sprints.json`: a
   nested `tractor workflow run plan` asked which greeting to print,
   received the scripted answer, and wrote brief, checklist, and a SIMPLE
   recommendation that the Go validator accepted.
6. **The planner's output executes.** `validations/000039-sprints.json`: a
   nested `plan` run produced a MEDIUM plan with two items, the printed
   `tractor workflow run medium --project live-medium-proof` handoff ran it
   with a real codex harness, and both items ended engine-marked with
   per-lap `validation.json`.
7. **Checklists are editable documents.** Chapters 2 and 3 started with
   `items: []`; the plan node appended items and sprint docs; the sprint
   loop picked them up on its next arrival. Decision 19, live.

## Scope check: what this run did not prove

- **`TRACTOR_RUN_DIR` from the engine.** The engine binary predates the
  change. Only three `QuestionAsked` events reached this run's timeline,
  from agents that set the variable by hand after finding the run
  directory in the process list. The other nine questions were seen by a
  directory watcher. The new behaviour was proven inside sprint 2's nested
  run, not by this run's own engine.
- **The LARGE workflow live.** `workflow/large.yaml` has unit tests
  (`TestBuiltInLarge`, `TestLargeRunsNestedPlanningChecklist`) and mirrors
  this pipeline, but no sprint ran it against a real harness.
- **Resume.** The run never crashed; `--resume` and the rewind-to-outer-loop
  rule were not exercised.
- **Chapter failure with all sprints done.** Every chapter command passed
  first time, so the plan-node-as-escape-hatch path was not exercised.
- **Validator quality.** Nobody reviewed the validators the plan nodes
  wrote (decision 14's "baby bear" review is not built). One inspector was
  stricter than its claim; see finding 2.
- **Reviewer independence.** Claude wrote chapter 1's sprints and answered
  every question; GPT wrote chapters 2 and 3's sprints and their check
  scripts, then implemented against them. Sprint 4 of chapter 2 loosened
  its own seed on lap 2 rather than its inspector.

## Findings for the loop node

1. **The failure summary loses the message.** `record.Summary` is
   `exit N — <last 400 runes of validation.log>`. Chapter 2 sprint 4's check
   printed its reason, then dumped diagnostics; the lap 2 frame carried the
   tail of a JSON timeline (`validations/000022-sprints.json`, `log_tail`).
   The agent reran the whole nested check to learn why. Fix: a larger tail
   and the `validation.log` path in the frame, now that agents have
   `TRACTOR_RUN_DIR`.
2. **Validator stricter than the claim, live.** The inspector wanted the
   literal greeting in the checklist command; the planner wrote a
   byte-exact hex comparison (`validations/000022-sprints.log`). Belongs to
   the planning-workflow phase per Tyler.
3. **Sub-second chapter validations** (`000010`, `000025`, `000040`) are
   real: Go's test cache made `go test ./...` instant. Fine, but a chapter
   command that only re-runs cached tests adds little; the sprint count
   check is what carried the weight.
