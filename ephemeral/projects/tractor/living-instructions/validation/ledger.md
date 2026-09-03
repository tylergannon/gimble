---
items:
  - name: P1 elicitation
    check: The design under validation/P1 proves P1 as stated in declaration.md and a coder cannot satisfy it while P1 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P1/design.md
  - name: P2 halt
    check: The design under validation/P2 proves P2 and cannot be satisfied while P2 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P2/design.md
  - name: P3 independent design review
    check: The design under validation/P3 proves P3 and cannot be satisfied while P3 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P3/design.md
  - name: P4 re-entry
    check: The design under validation/P4 proves P4 and cannot be satisfied while P4 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P4/design.md
  - name: P5 seven passes
    check: The design under validation/P5 proves P5 and cannot be satisfied while P5 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P5/design.md
    done: true
  - name: P6 verify before chapter done
    check: The design under validation/P6 proves P6 and cannot be satisfied while P6 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P6/design.md
  - name: P7 supervisor steer
    check: The design under validation/P7 proves P7 and cannot be satisfied while P7 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P7/design.md
    done: true
  - name: P8 library
    check: The design under validation/P8 proves P8 and cannot be satisfied while P8 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P8/design.md
  - name: P9 docs
    check: The design under validation/P9 proves P9 and cannot be satisfied while P9 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P9/design.md
  - name: P10 two seeds end to end
    check: The design under validation/P10 proves P10 and cannot be satisfied while P10 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P10/design.md
---

# Validation designs

One item per promise (decision 47). In the manual run, Claude is the
`design` node and a codex session (`codex exec`, read-only, fresh) is the
`review` node; an item is marked when the reviewer routes pass. Items
carry no `command`: the pass edge is the verdict (decision 57). Reviews
are kept as `review-N.md` beside each design; lap N+1 answers review N.

Each `design.md` has: archetype; the story a verifier follows; the
evidence captured and where; the validator (the ledger `command` and
`infer` the promise's sprint will carry, or the proof script by name);
and what the design does not prove.

## Rules every scenario follows

- **Fresh at check time.** A scenario's `command` creates its scratch
  repository and run directory itself (`mktemp -d`, `git init`, `tractor
  workflow run ... --logs <dir>`) after the coder's turn has ended.
  Nothing that exists before the check is evidence. The coder's
  deliverable is the binary and the library; the run that proves them is
  made by the check, and the check keeps its own transcript
  (`check.log`: every command it ran, the stdout it read, the handoffs
  it executed) as evidence.
- **Only recorded evidence.** A check reads what the run directory
  records (spec sections 5.6, 10, 12.4) and what the observer below
  captures. Run-directory facts used here: `timeline.jsonl` events with
  their documented fields; `stages/<seq>-<node>/`, where `tool.log`
  exists only for tool nodes, `prompt.md` and `response.md` only for
  codergen turns, `validation.json` only for loop turns (the engine's
  record of which item it validated and marked), and `steering.jsonl`
  holds delivered steers with origin and time; `response.md`, whose
  front matter carries the agent's own chosen `next` and whose body is
  the agent's notes; `events/<seq>-<node>.jsonl` segments, where every
  `tool_call` pairs with a `tool_result` by `call_id` and a steer
  appears as a second `user` event at the moment the harness received
  it; and `checkpoint.json`, whose `sessions` map records per node the
  harness that served it (`codex`, `claude`, `agy`, the routes of the
  providers `openai`, `anthropic`, `gemini`) and, by a NUL-prefixed
  `none:` key, that its thread mode was fresh. The run records the
  harness, never the model, so no design claims a model.
- **Content is coder-owned, so names prove nothing.** The graph, the
  prompts, and the edge conditions are library content the coder edits.
  A check never trusts a node's name or an edge's target alone. It binds
  a node to its kind by what the engine records for that kind, and binds
  a review or verify turn to its meaning by reading, with `infer`, the
  `prompt.md` it was given (was it told the promise, the design, the
  question) and the `response.md` it wrote (did its notes answer, and
  does the verdict in its notes agree with the `next` it chose). A
  relabelled graph or a gutted prompt cannot make a reviewer's notes
  answer a question it was not asked.
- **The engine marks; agents do not.** An item the loop validates gets
  a `LoopValidated` event and a `validation.json` in that loop stage.
  An item an agent hand-marks gets neither: the loop honours a `done`
  item without validating it (engine/loop.go). So where a promise says
  the engine marked every item, the check requires, for every item that
  is `done: true`, one `LoopValidated` with `passed: true` naming it and
  one loop stage whose `validation.json` names it. No snapshot timing is
  involved.
- **Snapshots are for authorship, and they race by model latency.**
  The observer's package copies at stage boundaries establish which
  stage changed a file. The engine appends `StageStarted` and runs the
  handler at once, so a copy taken at `StageStarted` can include a
  write the stage made before the copy finished; for a codergen stage
  the first write follows a model round trip, seconds after the event,
  and the copy takes milliseconds. Checks use the copy at the previous
  stage's `StageCompleted` as "before" and the stage's own
  `StageCompleted` as "after", and never use snapshots to time a loop
  stage's mark. The residual race is stated in each design that relies
  on a snapshot.
- **Semantic checks go to `infer`.** Where the promise is about meaning,
  the `command` binds identity and order and the `infer` judge reads the
  text. No `command` requires verbatim containment.
- **Inconclusive is not pass.** A scenario that depends on model
  behaviour it cannot force exits with a distinct message when the
  behaviour did not occur; the item stays open and the scenario is rerun.
- **Out of scope for every check.** The engine and harness are excluded
  from this project (no engine change; `promises.md` exclusions). A game
  that needs the engine or a harness adapter to record false events,
  reuse a session it reports as fresh, or misroute, is a defect in code
  the coder is not allowed to touch; the required checks (`go test
  ./...`) and the diff review cover that boundary.

## Shared evidence tooling

`validation/observer.sh <run-dir> <rules-file>` is the scripted human
and the witness. It tails `<run-dir>/timeline.jsonl` and:

- on every `StageStarted`, `StageCompleted`, `QuestionAsked`, and
  `SupervisorVerdict` event, copies the package directory
  (`ephemeral/projects/<build>/`, without `research/` leaves) into
  `<run-dir>/observer/<n>-<event>-<node>/`;
- on `QuestionAsked` only: copies the package, reads the question file,
  matches it against ordered rules (a substring or a `Promise:` line
  predicate, and an answer), composes the answer (when the question
  numbers its candidates the answer mirrors the numbering, one line per
  candidate, decision 39), copies the package again immediately before
  writing the answer with `tractor answer` (so a change made between
  the ask and the answer is visible), and appends
  `<id> <rule> <ts> <candidate number> <matched line>` (tab-separated) to
  `<run-dir>/observer/answers.log`. It never answers a file it found by
  polling, so a question written without `tractor ask` is never
  answered.

The rules file is part of the scenario; the `observer/` tree is part of
the evidence. Built once, in chapter 5 sprint 3, and reused by every
later scenario. `validation/lib/run-plan.sh <seed> <rules>` makes a
scratch repository, runs `plan` with the observer attached, records the
run's stdout to `check.log`, waits for completion, and prints the run
directory. `validation/lib/timeline.sh` holds the `jq` idioms the proof
scripts share (last stage of a node, nearest preceding stage, event by
item, snapshot before and after a stage, paired tool result). Both are
written with the first proof script that needs them.
