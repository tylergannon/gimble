# P4: a rejected validation design is re-designed with the reviewer's notes and later passes

Archetype: scenario. Lap 5; answers `review-4.md`.

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
  `LoopItemSelected(item, lap)`, `LoopValidated` for the first item;
  `validation.json` of the marking loop stage.
- `stages/<seq>-review/prompt.md` and `response.md` for both review
  turns; `stages/<seq>-design/prompt.md` for the second design turn and
  its segment.
- `observer/` copies of `validation/<item>/` at: the `StageCompleted`
  before the first review (D1); the first review's `StageCompleted`
  (R1: notes present, design untouched); the second design's
  `StageCompleted` (D2); the end.

## Validator

`command`: `prove/p4-reentry.sh`: for the first item in
`validation/ledger.md`, the timeline shows, in order: `design`, `review`
with `next: design`, `design`, `review` with `next:` the loop node, then
a loop stage whose `LoopValidated` (`passed: true`) and `validation.json`
name the item; each review's `response.md` front matter agrees with its
`StageCompleted`; the notes file is absent in D1 and present in R1, and
`design.md` is byte-identical between D1 and R1 (the reviewer wrote
notes, not a redesign); the notes were available to the second design
turn: its `prompt.md` contains the notes' first line, or its segment
has a `tool_call` whose arguments contain the notes path with a paired
`tool_result` whose output contains that line; `design.md` in D2
differs from D1, and `design.md` at the end equals D2 (the design turn
made the redesign, and the second reviewer left it alone).

`infer` (files: both review turns' `prompt.md` and `response.md`, the
second design turn's segment, the notes, D1 and D2): "Did the first
reviewer read D1 and reject it in its notes, with the route agreeing?
Did the second design turn actually take in the notes (not merely echo
their path and first line) and did D2 respond to them? Did the second
reviewer read D2 and accept it in its notes, with the route agreeing?
Fail if any answer is no, or if D2 is D1 with cosmetic edits."

## Not proven

How often reviewers reject in practice. Whether the model follows the
seed's instruction: if the first review routes pass, the scenario exits
"inconclusive: first design was not rejected", the item stays open, and
the scenario is rerun with the instruction moved into the design
prompt's documented test hook (`TRACTOR_TEST_TRIVIAL_FIRST=1`, honored
only when the seed asks for it). Whether every objection was resolved;
the second reviewer's pass is that judgment. The copies race by model
latency (ledger rules): each turn's first write follows a model round
trip.
