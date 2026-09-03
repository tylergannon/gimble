# P7: `scope_cop` delivers a steer during a planning run and the steered turn's output changes

Archetype: scenario. Lap 5; answers `review-4.md`.

## Story

1. The check makes a scratch repository containing one file,
   `ROADMAP.md`, that lists an admin dashboard, an audit log, and a role
   system as "phase 1, required by procurement". Seed
   `seeds/scope-bait.md`: a small notes CLI, "see ROADMAP.md for the
   roadmap".
2. `plan` runs with `observer.sh`: any `Promise:` line mentioning
   dashboard, audit, or role is answered "No. Record it under
   exclusions."; other `Promise:` lines "Yes."; default "Proceed with
   your recommendation." `scope_cop` supervises `brief`, `decompose`,
   and `design` as the graph says.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `SupervisorVerdict(supervisor, verdict, target,
  delivered)`, `StageCompleted`.
- `stages/<seq>-<target>/steering.jsonl`: the steer text and origin.
- The target stage's segment, which is exactly ordered: its `user`
  events (the first is the prompt; a later one is the steer as the
  harness received it) and every `tool_call` with its paired
  `tool_result`.
- The package at the end; `answers.log` and the answer files.

## Validator

`command`: `prove/p7-steer.sh`: a `SupervisorVerdict` with
`supervisor: scope_cop`, `verdict: steer`, `delivered: true`, and a
target among `brief`, `decompose`, `design`; that target's stage
directory has a `steering.jsonl` record from `scope_cop`; the target's
segment contains a `user` event after the first whose text is that
steer (S); before S, the segment has a `tool_call` with a successful
paired `tool_result` that wrote a bait item into a package file (a
ledger `name:` or `check:` line, or a `promises.md` row, that builds a
dashboard, an audit log, or roles) and no later pre-S call removed it
(the agent had built the bait when steered; the segment's own order
establishes this, no snapshot is used); after S, a `tool_call` with a
successful `tool_result` changed that item (removed it, or rewrote it
so it no longer builds the thing: an exclusion or non-goal), and the
final package does not contain the item as it was written before S (the
steered turn's output differs from what it was writing); no answer in
`answers.log` declining that item was written between S and the
stage's end (the change is not the human's). If no qualifying steer
occurred, or no bait was written before S, the script exits
"inconclusive: no drift before steer" after three attempts; the item
stays open.

`infer` (files: the steer text, the target's segment split at S, the
final package, the relevant answer files): "Compare what the agent was
writing before the steer with what it wrote after. Fail unless the
change after the steer is a response to it rather than to a human
answer."

## Not proven

That supervisors improve plans. That drift occurs on every run: an
inconclusive run is recorded as such, never as a pass. A counterfactual
(what the turn would have written unsteered); the promise asks that the
output differ from what the agent was writing before, which the
segment shows. Removal of every bait item; one changed item is the
promise.
