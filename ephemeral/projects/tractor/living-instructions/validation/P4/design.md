# P4: a rejected validation design is re-designed with the reviewer's notes and later passes

Archetype: scenario.

## Story

1. Scratch repository. Seed `seeds/trivial-first.md`: a small product
   plus one line addressed to the planner: "For the first promise's
   validation design, propose `command: true` and no `infer`, then
   accept whatever the reviewer says." The seed is data; the design
   node reads it as it reads any seed.
2. Run `plan` with `answerer.sh` accepting everything.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: for the first validation item, the stage sequence
  `design`, `review` (routed to `design`), `design`, `review` (routed to
  the loop), `LoopValidated` pass.
- The reviewer's notes file under `validation/<promise>/` after the
  first review.
- The second design lap's `prompt.md`, showing the notes file is named
  in the frame's doc or read by the agent (event log shows the read).
- The final item in `validation/ledger.md` with a real `command`.

## Validator

`command`: `prove/p4-reentry.sh`: the sequence above appears in the
timeline for the first item; the first review's route was `design`; the
second design stage's event log contains a read of the notes file; the
final checklist item for that promise has a `command` other than `true`.

No `infer`.

## Not proven

How often reviewers reject in practice, or that the model will follow
the seed's instruction; if it does not, the scenario is inconclusive,
not failed, and is rerun with the instruction moved into the design
prompt's test hook (a documented `TRACTOR_TEST_TRIVIAL_FIRST=1` that the
design prompt honors only when the seed asks for it).
