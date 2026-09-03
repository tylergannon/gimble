# P4: a rejected validation design is re-designed with the reviewer's notes and later passes

Archetype: scenario. Lap 10; answers `review-9.md`.

## Story

1. The check makes a scratch repository. Seed `seeds/trivial-first.md`:
   a small product plus one line addressed to the planner: "For the
   first promise's validation design, propose `command: true` and no
   `infer`, then accept whatever the reviewer says." The seed is data;
   the design node reads it as it reads any seed.
2. `plan` runs with `observer.sh` accepting everything.
3. Wait until the validation loop has marked the item for the first
   promise (its `LoopValidated`), or the run ends, whichever comes
   first. What the run does afterwards is outside P4.

## Evidence

- `promises.md` from the package: the first promise's id X.
- `validation/ledger.md`: the item bound to X (its `doc` path under
  `validation/X/`, or its name X followed by a non-word character or
  the end; the file that `doc` names is the design under test, per the
  package layout in `planning-workflow.md` section 6:
  `validation/<promise>/design.md`).
- `timeline.jsonl`: `StageStarted`, `StageCompleted(next)`,
  `StageFailed`, `LoopItemSelected(item, lap)`, `LoopValidated` for that
  item; `validation.json` of the marking loop stage.
- For that item's laps: every completed review turn's `prompt.md` and
  `response.md`; every completed design turn after the first:
  `prompt.md` and segment (every `tool_call` and paired `tool_result`).
- Observer copies of `validation/X/` at the `StageCompleted` of each
  design and review stage for this item (D1, R1, D2, R2, ... ) and
  after the marking stage.
- `validation/ledger.md` after the marking stage.

## Validator

`command`: `prove/p4-reentry.sh`: the item under test is the one bound
to X (a ledger item serving no promise is not it) and its `doc` names
the design file F under `validation/X/`; the completed stages between
its first `LoopItemSelected` and its `LoopValidated` (`passed: true`,
with a matching `validation.json`) alternate design and review, with at
least two review turns; every review but the last has `next: design`
and the last has `next:` the loop node (rejected at least once, then
passed); retried attempts (a `StageFailed` stage with no `response.md`)
are skipped; the item is `done: true` in the ledger copy after the
marking stage; for the first review: the notes file is absent in D1
and present in R1, and F is byte-identical between D1 and R1 (the
reviewer wrote notes, not a redesign); the notes were available to
every design turn after the first: every non-empty line of the notes
appears in that turn's `prompt.md`, or every non-empty line appears in
the paired `tool_result` of a `tool_call` in its segment whose
arguments contain the notes path; F differs between the copies before
and after that design turn, and is byte-identical across the review
turn that follows it (each redesign is the design turn's, and the
reviewer left it alone).

`infer` (files: the proof tooling that produced the evidence
(`validation/observer.sh`, `validation/lib/`, the proof script named
above) and the segment of every turn this design judges; the seed, `promises.md`, the ledger item, every review
turn's `prompt.md` and `response.md`; every redesigning turn's
`prompt.md` and segment; the notes; the copies of F: D1, D2, ...): "Is
the item under test the validation design for the seed's first
promise, and is F the design the reviewers read (their prompts name
it)? Did each rejecting reviewer read the design in front of it and
reject it in its notes, with the route agreeing? Was the full text of
the notes in front of each redesigning turn, in its prompt or in a file
it read, and did its redesign change the design in a way that answers
those notes rather than cosmetically? Did the final reviewer read the
final design and accept it in its notes, with the route agreeing? Fail
on any no."

## Not proven

How often reviewers reject in practice, or that two laps suffice; any
number of rejections before the pass satisfies P4. Whether the run
completes after the item is marked; P4 ends at the mark. Whether the
model follows the seed's instruction: if the first review routes pass,
the scenario exits "inconclusive: first design was not rejected", the
item stays open, and the scenario is rerun (no test hook in the
product; a seed is data). Whether every objection was resolved; the
passing reviewer's judgment is that. The copies race by model latency
(ledger rules): each turn's first write follows a model round trip.
