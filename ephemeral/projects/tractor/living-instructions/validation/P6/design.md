# P6: `large` completes and every chapter is marked only after `verify` routed pass

Archetype: scenario. Lap 3; answers `review-2.md`.

## Story

1. A LARGE package produced at check time by `plan` on
   `seeds/ledger-tool.md` (the script runs `plan` itself; when the
   recommendation is not `large`, it exits "inconclusive: package sized
   medium" and the item stays open).
2. `tractor workflow run large --project <build> --logs <fresh dir>` with
   `observer.sh` accepting any question. `models.yaml` puts `implement`
   on one provider and `verify` on another.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `StageStarted`, `StageCompleted(name, next)`,
  `LoopItemSelected`, `LoopValidated(node, item, passed)`,
  `PipelineCompleted`.
- `stages/<seq>-verify/`: `prompt.md`, `response.md` (chosen `next`,
  the verifier's notes with its verdict line and what it ran).
- `observer/` snapshots of the chapter ledger and every chapter's
  `sprints.md` at every stage boundary.
- `checkpoint.json` `sessions`: harness for `verify` and `implement`.

## Validator

`command`: `prove/p6-verify-before-done.sh <large run dir>`: the last
`StageCompleted` has `next: success` and `PipelineCompleted` follows it;
every chapter item is `done: true`; for each, exactly one `chapters`
loop stage flips it, with a `LoopValidated` on `chapters` with
`passed: true` during that stage and no change across any other stage
(no agent marked it, whatever it read); for each such loop stage the
nearest preceding `StageCompleted` of a node other than the two loops is
a `verify` stage with `next: chapters`, whose `response.md` front matter
agrees and whose body contains `ROUTE: pass`; every `verify` stage
directory has `prompt.md`, `response.md`, and a segment in
`events/index.jsonl` (a codergen turn; a tool node has `tool.log` and
none of these); `checkpoint.json` records different harnesses for
`verify` and `implement`; the chapter ledger is byte-identical across
every `plan`, `implement`, `replan`, and `verify` stage; each `replan`
stage changes only the current chapter's `sprints.md` (the chapter from
the enclosing `LoopItemSelected`) and no other package file.

`infer` (files: each `replan` stage's before and after snapshots of the
sprint ledger): "Did any replan change an item that was `done: true`
before it ran? Fail if so."

## Not proven

That the software the package describes is good. That the verifier
operated the software rather than reading artifacts: its prompt requires
it and P10's judge runs the software independently, but this promise is
about order and marking only. The verifier's catch rate (research R2).
