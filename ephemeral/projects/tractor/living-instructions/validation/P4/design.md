# P4: a rejected validation design is re-designed with the reviewer's notes and later passes

Archetype: scenario. Lap 2; answers `review-1.md`.

## Story

1. The check makes a scratch repository. Seed `seeds/trivial-first.md`:
   a small product plus one line addressed to the planner: "For the
   first promise's validation design, propose `command: true` and no
   `infer`, then accept whatever the reviewer says." The seed is data;
   the design node reads it as it reads any seed.
2. `plan` runs with `answerer.sh` accepting everything.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `StageStarted`, `StageCompleted(next)`,
  `LoopItemSelected(item, lap)`, `LoopValidated` for the first item.
- The segments of the two `design` stages and the first `review` stage:
  `tool_call` arguments (paths written) and `tool_result` outputs (files
  read). The first review's segment holds the design it read, in the
  `tool_result` of its read.
- The reviewer's notes file under `validation/<item>/` and the final
  `design.md` there.
- The second design stage's `prompt.md`.

## Validator

`command`: `prove/p4-reentry.sh`: for the first item in
`validation/ledger.md`, the timeline shows, in order: `design`, `review`
with `next: design`, `design`, `review` with `next:` the loop node, and
`LoopValidated` with `passed: true`; the first `review` segment has a
`tool_call` that writes a file under `validation/<item>/` and that file
exists (the notes); the second `design` stage read the notes: its
`prompt.md` names the notes path, and either its `prompt.md` contains
the notes' first line or its segment has a `tool_result` whose output
contains it.

`infer` (files: the first review segment, the notes, the final
`design.md`): "The reviewer's segment contains the design it read. Did
the final design change in the way the notes asked? Fail if an objection
in the notes is unaddressed, or if the final design is the first design
with only cosmetic edits."

## Not proven

How often reviewers reject in practice. Whether the model follows the
seed's instruction: if the first review routes pass, the scenario exits
with "inconclusive: first design was not rejected", the item stays open,
and the scenario is rerun with the instruction moved into the design
prompt's documented test hook (`TRACTOR_TEST_TRIVIAL_FIRST=1`, honored
only when the seed asks for it). No check on the shape of the final
validator; the second reviewer's pass is that judgment.
