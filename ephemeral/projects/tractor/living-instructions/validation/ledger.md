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
  - name: P6 verify before chapter done
    check: The design under validation/P6 proves P6 and cannot be satisfied while P6 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P6/design.md
  - name: P7 supervisor steer
    check: The design under validation/P7 proves P7 and cannot be satisfied while P7 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P7/design.md
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
  made by the check.
- **Only recorded evidence.** A check reads what the run directory
  records (spec sections 5.6, 10, 12.4) and what the observer below
  captures. Run-directory facts used here: `timeline.jsonl` events with
  their documented fields; `stages/<seq>-<node>/`, where `tool.log`
  exists only for tool nodes, `prompt.md` and `response.md` only for
  codergen turns, and `steering.jsonl` holds delivered steers with
  origin and time; `response.md`, whose front matter carries the agent's
  own chosen `next` and whose body is the agent's notes;
  `events/<seq>-<node>.jsonl` segments, where a steer appears as a
  second `user` event at the moment it was handed to the harness; and
  `checkpoint.json`, whose `sessions` map records per node the harness
  that served it (`codex`, `claude`, `agy`, the routes of the providers
  `openai`, `anthropic`, `gemini`) and, by a NUL-prefixed `none:` key,
  that its thread mode was fresh. The run records the harness, never the
  model, so no design claims a model.
- **Content is coder-owned, so names prove nothing.** The graph, the
  prompts, and the edge conditions are library content the coder edits.
  A check therefore never trusts a node's name or an edge's target alone.
  It binds a node to its kind by what the engine records for that kind
  (a `tool.log`; a `prompt.md`, `response.md`, and segment; a
  `LoopValidated`), and binds a route to its meaning by the agent's own
  words: the `response.md` body of a review or verify turn states its
  verdict, and the verdict must agree with the `next` it chose. A graph
  whose edges are relabelled cannot make a reviewer write "pass" when it
  found a defect.
- **Ledgers change only across the loop's own stage.** Where a promise
  says the engine marked an item, the check uses the observer's
  snapshots: the item is not `done` in the snapshot at the loop stage's
  `StageStarted`, a `LoopValidated` with `passed: true` for that item
  is emitted during that stage, and the item is `done` at its
  `StageCompleted`; and the item's `done` never changes across any other
  stage. An item hand-marked by an agent flips across that agent's
  stage, with no event. Commandless items validate as passed by
  construction (decision 57); the event still records that the engine,
  not an agent, marked them, which is what the promises claim.
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
  `<run-dir>/observer/<n>-<event>-<node>/`, so any file's state at any
  stage boundary is evidence the agents never wrote;
- on `QuestionAsked` only, reads the question file, matches it against
  ordered rules (a substring or a `Promise:` line predicate, and an
  answer), writes the answer with `tractor answer`, and appends
  `<id> <rule> <ts> <matched line>` (tab-separated) to
  `<run-dir>/observer/answers.log`. It never answers a file it found by
  polling, so a question written without `tractor ask` is never
  answered.

The rules file is part of the scenario; the `observer/` tree is part of
the evidence. Built once, in chapter 5 sprint 3, and reused by every
later scenario. `validation/lib/run-plan.sh <seed> <rules>` makes a
scratch repository, runs `plan` with the observer attached, waits for
completion, and prints the run directory. `validation/lib/timeline.sh`
holds the `jq` idioms the proof scripts share (last stage of a node,
nearest preceding stage, event by item, snapshot before and after a
stage). Both are written with the first proof script that needs them.
