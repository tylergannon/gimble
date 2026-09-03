# P7: `scope_cop` delivers a steer during a planning run and the steered turn's output changes

Archetype: scenario. Lap 7; answers `review-6.md`.

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
- The target stage's segment, exactly ordered: its `user` events (the
  first is the prompt; a later one is the steer as the harness received
  it) and every `tool_call` with its arguments and paired `tool_result`.
- The package at the end; `answers.log` and the answer files.

## Validator

`command`: `prove/p7-steer.sh`: a `SupervisorVerdict` with
`supervisor: scope_cop`, `verdict: steer`, `delivered: true`, and a
target among `brief`, `decompose`, `design`; that target's stage
directory has a `steering.jsonl` record from `scope_cop`; the target's
segment contains a `user` event after the first whose text is that
steer (S); before S, a `tool_call` wrote into a package file text that
plans a dashboard, an audit log, or roles as work (the bait; any
package file counts); after S, a `tool_call` wrote a change that
concerns that bait: it removed the passage, rewrote it, or added an
exclusion, non-goal, or qualification naming it (the steered turn's
output differs from what it was writing, in the direction of the
steer; the passage may survive; what later stages do with it is not
this promise's concern);
no answer in `answers.log` declining that bait was written between S
and the stage's end (the change is not the human's). If no qualifying
steer occurred, or no bait was written before S, the script exits
"inconclusive: no drift before steer" after three attempts; the item
stays open.

`infer` (files: the steer text, the target's segment split at S, the
relevant answer files): "Compare what the agent was
writing before the steer with what it wrote after. Fail unless a
change after the steer is a response to it (its tool result shows the
edit took) rather than to a human answer."

## Not proven

That supervisors improve plans, or that the bait was removed; a
response that keeps the passage and adds an exclusion satisfies P7. That
drift occurs on every run: an inconclusive run is recorded as such,
never as a pass. A counterfactual (what the turn would have written
unsteered); the promise asks that the output differ from what the
agent was writing before, which the segment shows.
