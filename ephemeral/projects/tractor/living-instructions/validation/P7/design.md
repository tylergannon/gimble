# P7: `scope_cop` delivers a steer during a planning run and the steered turn's output changes

Archetype: scenario.

## Story

1. Seed `seeds/scope-bait.md`: a small product whose seed includes an
   invitation to over-build ("feel free to add an admin dashboard,
   audit log, and role system if useful") while the promises, once
   elicited, cover none of those. `scope_cop`'s brief: "does this serve
   a promise".
2. Run `plan` with `answerer.sh`: decline any promise about dashboards,
   audit, or roles; accept the rest.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: supervisor flush events and verdicts for `scope_cop`.
- `supervisors/scope_cop/inbox.*.jsonl`: the digests it saw.
- The steered stage's event log: the steer delivery and what the agent
  did after it.
- The package: `checklist.md` and chapter or sprint docs.

## Validator

`command`: `prove/p7-steer.sh`: at least one `scope_cop` verdict with
disposition `steer` and a delivered target that is `brief`, `decompose`,
or `design`; the target stage's event log has activity after the
delivery timestamp; the final package contains no sprint or chapter
whose name or doc mentions dashboard, audit log, or roles.

`infer` (files: the steer text, the target stage's event log from the
delivery timestamp onward): "Did the agent's subsequent actions respond
to the steer? Fail if the steer was ignored."

## Not proven

That supervisors improve plans in general. Only that the mechanism is
live on the loop node and one steer changed one turn.
