---
items:
  - name: P1 elicitation
    check: The design under validation/P1 proves P1 as stated in declaration.md and a coder cannot satisfy it while P1 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P1/design.md
    done: true
  - name: P2 halt
    check: The design under validation/P2 proves P2 and cannot be satisfied while P2 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P2/design.md
    done: true
  - name: P3 independent design review
    check: The design under validation/P3 proves P3 and cannot be satisfied while P3 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P3/design.md
    done: true
  - name: P4 re-entry
    check: The design under validation/P4 proves P4 and cannot be satisfied while P4 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P4/design.md
    done: true
  - name: P5 seven passes
    check: The design under validation/P5 proves P5 and cannot be satisfied while P5 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P5/design.md
    done: true
  - name: P6 verify before chapter done
    check: The design under validation/P6 proves P6 and cannot be satisfied while P6 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P6/design.md
    done: true
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
    done: true
  - name: P10 two seeds end to end
    check: The design under validation/P10 proves P10 and cannot be satisfied while P10 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P10/design.md
    done: true
---

# Validation designs

One item per promise (decision 47). In the manual run, Claude is the
`design` node and a codex session (`codex exec`, read-only, fresh) is the
`review` node; an item is marked when the reviewer routes pass. Items
carry no `command`: the pass edge is the verdict (decision 57). Reviews
are kept as `review-N.md` beside each design; lap N+1 answers review N.

Each `design.md` has: archetype; the story a verifier follows; the
evidence captured and where; the verifier (the ledger `command` and `infer` the promise's sprint
will carry, or the proof script by name; the declaration's word);
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
- **The observer's mirror catches forged records.** The observer copies
  every `timeline.jsonl` event as it arrives into its own append-only
  mirror with an arrival time and the stage it saw live at that moment
  (from the `StageStarted` and `StageCompleted` events it has already
  mirrored). The engine appends events between and around stages; a
  tool or agent turn can only write while its own stage is live. So an
  event that belongs to another stage or to a loop (`LoopValidated`,
  another node's `StageStarted` or `StageCompleted`, `PipelineCompleted`)
  arriving while a tool or codergen stage is live was forged by that
  stage, and a stage directory whose `StageStarted` arrived while some
  other stage was live is forged. Checks that rest on engine records
  require those records to have arrived when the engine, and nothing
  else, could have written them. The check also started the `tractor workflow run` process itself, so
  it knows when that process exited and with what status: a run whose
  process exited non-zero did not complete, whatever the timeline says,
  and any event whose engine-stamped `ts` is later than the process's
  exit time was written by something else (arrival order in the mirror
  may lag by scheduling; the event's own timestamp does not). Together these close the forgery the run
  directory's writability leaves open, short of a process an agent
  detaches to write during the run's own bookkeeping gaps, which is the
  harness sandbox's concern (out of scope: no engine or harness change)
  and is recorded under Not proven where a design leans on the mirror.
  The mirror lives with the
  observer, outside the run directory, at its random path, and every
  record the observer writes (mirror lines, snapshot manifests with the
  hash of each copied file, answers.log lines) is chained: each carries
  the HMAC of itself and the previous record under a key the observer
  holds only in memory and hands to the check on its own stdout when the
  run ends. A file edited after the fact breaks the chain, so a
  process that finds the observer's path gains nothing.
- **The proof tooling is the planner's, not the coder's.** The
  observer, the shared proof library, the chapter proof scripts, the
  ledgers, and the pinned baseline commits are planner-authored content
  committed before the sprint that uses them (as the doctrine pages
  are). The implement prompt forbids editing them; an implement turn's
  segment records every write it makes, so a `tool_call` whose
  arguments name a file under `prove/`, `validation/observer.sh`,
  `validation/lib/`, a ledger, or a baseline pin is a fail of that
  sprint, and every judge that reads a proof log also receives the
  scripts that wrote it and the segment of the turn under judgment, so
  a producer that constructs its own story is read, not trusted.
- **Out of scope for every check.** The engine and harness are excluded
  from this project (no engine change; `promises.md` exclusions). A game
  that needs the engine or a harness adapter to record false events,
  reuse a session it reports as fresh, or misroute, is a defect in code
  the coder is not allowed to touch; the required checks (`go test
  ./...`) and the diff review cover that boundary.

## Shared evidence tooling

`validation/observer.sh <run-dir> <rules-file>` is the scripted human
and the witness. Its own tree lives outside the run directory, at a path the check
draws at random under its scratch space and passes to the observer,
never inside the run directory and never at a predictable name, so
nothing under `TRACTOR_RUN_DIR` betrays its presence to the agents and
no agent can find it to edit it. It tails
`<run-dir>/timeline.jsonl` and:

- on every `StageStarted`, `StageCompleted`, `QuestionAsked`, and
  `SupervisorVerdict` event, copies the package directory
  (`ephemeral/projects/<build>/`, without `research/` leaves) into
  `<observer-tree>/<n>-<event>-<node>/`;
- on `QuestionAsked` only: copies the package (for the approval question,
  the last one a plan run asks, the whole package with research leaves),
  reads the question file,
  matches it against ordered rules (a substring or a `Promise:` line
  predicate, and an answer), composes the answer (when the question
  numbers its candidates the answer mirrors the numbering, one line per
  candidate, decision 39) and works a random token into it (a made-up
  word inside an ordinary sentence, different every time, with no fixed
  marker such as a label or a trailing line, so nothing distinguishes an
  observed run from a human-answered one), copies the package again
  immediately before writing the answer with `tractor answer` (so a
  change made between the ask and the answer is visible), and appends
  `<id> <rule> <ts> <candidate number> <nonce> <matched line>`
  (tab-separated) to `<observer-tree>/answers.log`. `tractor ask`
prints the answer it received, so the `tool_result` of the ask that
waited on a question contains that question's token: that is how a
check binds an ask call to its `QuestionAsked` event. It never answers a file it found by
  polling, so a question written without `tractor ask` is never
  answered.

The rules file is part of the scenario; the `observer/` tree is part of
the evidence. Written by the planner before chapter 5 sprint 5, which wires and
exercises it, and reused by every later scenario; the coder never
edits it (rule above). `validation/lib/run-plan.sh <seed> <rules>` makes a
scratch repository, runs `plan` with the observer attached, records the
run's stdout to `check.log`, waits for completion, and prints the run
directory. `validation/lib/timeline.sh` holds the `jq` idioms the proof
scripts share (last stage of a node, nearest preceding stage, event by
item, snapshot before and after a stage, paired tool result). Both are
written with the first proof script that needs them.
