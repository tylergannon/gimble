# P4: a rejected validation design is re-designed with the reviewer's notes and later passes

Archetype: scenario. Lap 3; answers `review-2.md`.

## Story

1. The check makes a scratch repository. Seed `seeds/trivial-first.md`:
   a small product plus one line addressed to the planner: "For the
   first promise's validation design, propose `command: true` and no
   `infer`, then accept whatever the reviewer says." The seed is data;
   the design node reads it as it reads any seed.
2. `plan` runs with `observer.sh` accepting everything.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `StageStarted`, `StageCompleted(next)`,
  `LoopItemSelected(item, lap)`, `LoopValidated` for the first item.
- `stages/<seq>-review/response.md` for both review turns: chosen
  `next` and the notes with their verdict line.
- The first review's segment: the `tool_result` of its read of
  `design.md` (the first design, as the reviewer saw it).
- The reviewer's notes file under `validation/<item>/` and the observer
  snapshots around the first review stage.
- The second design stage: `prompt.md`, and its segment's `tool_call`
  and paired `tool_result` for the notes path.
- The final `design.md`.

## Validator

`command`: `prove/p4-reentry.sh`: for the first item in
`validation/ledger.md`, the timeline shows, in order: `design`, `review`
with `next: design`, `design`, `review` with `next:` the loop node, then
the loop stage whose `LoopValidated` marks the item; the first review's
`response.md` front matter says `next: design` and its body contains
`ROUTE: fail`, the second's says the loop node and `ROUTE: pass` (the
reviewer's words agree with the routes, so a relabelled graph cannot
fake the sequence); the notes file appears across the first review
stage (absent at its `StageStarted` snapshot, present at
`StageCompleted`); the second design stage's `prompt.md` names the
notes path, and its segment has a `tool_call` whose arguments contain
that path with a paired `tool_result` (same `call_id`) whose output
contains the notes' first line.

`infer` (files: the first review segment, the notes, the final
`design.md`): "The reviewer's segment contains the design it read. Did
the redesign respond to the notes? Fail only if the final design
ignores the notes: the objections are neither addressed nor answered,
or the final design is the first with cosmetic edits."

## Not proven

How often reviewers reject in practice. Whether the model follows the
seed's instruction: if the first review routes pass, the scenario exits
"inconclusive: first design was not rejected", the item stays open, and
the scenario is rerun with the instruction moved into the design
prompt's documented test hook (`TRACTOR_TEST_TRIVIAL_FIRST=1`, honored
only when the seed asks for it). Whether every objection was resolved;
the second reviewer's pass is that judgment.
