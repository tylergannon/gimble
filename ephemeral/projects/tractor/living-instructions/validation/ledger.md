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
  made by the check. The residual game, a binary that writes false
  events, is what the chapter's `verify` node exists for: it operates
  the software itself and reads the same run directory (decision 41).
- **Only recorded evidence.** A check reads what the run directory
  records (spec sections 5.6, 10, 12.4): `timeline.jsonl` events with
  their documented fields; `stages/<seq>-<node>/` contents, where
  `tool.log` exists only for tool nodes and `steering.jsonl` holds
  delivered steers with a timestamp and origin; `events/<seq>-<node>.jsonl`
  segments, whose `tool_call` events carry the arguments an agent used,
  including every path it wrote; and `checkpoint.json`, whose `sessions`
  map records, per node, the harness that served it and whether its
  thread mode was fresh (the key carries the NUL-prefixed `none:` marker
  the backend uses for no-thread nodes). The run records the harness,
  never the model, so no design claims a model.
- **Agents never mark ledgers.** Where a promise says the engine marked
  an item, the check requires a `LoopValidated` event with `passed: true`
  for that item and no `tool_call` in any agent segment whose arguments
  name the ledger path. An item hand-marked `done: true` is honored by
  the loop without a `LoopValidated` event (engine/loop.go), which is
  exactly what makes the event the discriminator.
- **Semantic checks go to `infer`.** Where the promise is about meaning
  (an exclusion matches a declined promise, a question asks rather than
  announces), the `command` binds identity and order and the `infer`
  judge reads the text. No `command` requires verbatim containment.
- **Inconclusive is not pass.** A scenario that depends on model
  behaviour it cannot force (P7's drift) exits with a distinct message
  when the behaviour did not occur; the item stays open and the scenario
  is rerun.

## Shared evidence tooling

`validation/answerer.sh <run-dir> <rules-file>` is the scripted human.
It tails `<run-dir>/timeline.jsonl` and acts only on `QuestionAsked`
events, never on files it finds by polling, so a question file an agent
wrote without `tractor ask` is never answered. On each event it: copies
the question file, `promises.md`, and `research/findings.md` as they are
at that moment into `<run-dir>/answers/<id>/`; matches the question
against ordered rules (a substring or a `Promise:` line predicate, and an
answer); writes the answer with `tractor answer`; and appends
`<id> <rule> <ts> <matched line>` (tab-separated) to
`<run-dir>/answers/answers.log`. The rules file is part of the scenario;
the `answers/` tree is part of the evidence. Built once, in chapter 5
sprint 3, and reused by every later scenario.

`validation/lib/run-plan.sh <seed> <rules>` makes a scratch repository,
runs `plan` with the answerer attached, waits for completion, and prints
the run directory. `validation/lib/timeline.sh` holds the `jq` idioms the
proof scripts share (last stage of a node, nearest preceding stage, event
by item, segment for a stage). Both are written with the first proof
script that needs them.
