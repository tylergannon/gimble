# The loop node, v1

Design note for the first implementation, written 2026-09-02 from the
loop-frames memo (§9 as amended by §9a) and the interview decisions in
`decisions.md`. This is the spec the code on branch `worktree-goal-gates`
is built to. Expect it to be reworked after Tyler reviews the result.

## 1. What it is

A `loop` node iterates a checklist file. On every lap return it re-reads the
file and validates the framed item plus every item already marked done, in
file order. It marks the framed item done only when the whole set passes and
unmarks every failed item. It then selects the first item not yet done,
injects it as the frame for the lap, and dispatches the body. When no item is
left it routes to `on_done`.

The engine owns selection, validation, marking, and injection. It never
stores loop state anywhere durable: the checklist file is the truth, and the
engine caches, injects, and forgets (memo §9a.3). Agents never write the
`done` field.

## 2. The checklist file

Markdown with YAML frontmatter. The frontmatter is the ledger; the body is
prose the engine never reads (definition of done, context, notes).

```markdown
---
items:
  - name: Build the login screen
    check: The login screen validates and submits on valid input
    command: npx playwright test tests/login.spec.ts
    infer:
      files: ephemeral/captures/login/*.png
      prompt: Judge whether every state of the login screen looks usable
    doc: ephemeral/projects/mvp/sprints/SPRINT-0002.md
  - name: Reject a bad password
    check: A wrong password shows an inline error and keeps the form
    command: npx playwright test tests/login-error.spec.ts
    done: true
---

# Sprint 2: login

Definition of done in open prose. Anything here is for agents and people.
```

Item fields:

| Field | Required | Meaning |
|---|---|---|
| `name` | yes | Identity. Unique within the file. |
| `check` | yes | The claim, as observable behavior. Prose for agents. |
| `command` | no | Shell command run from the workdir. Exit 0 passes. |
| `infer.files` | with `infer` | One glob or a list of globs, relative to the workdir. |
| `infer.prompt` | with `infer` | What the judge is to decide about those files. |
| `doc` | no | Path to a prose document injected with the item. |
| `checklist` | no | Path to a sub-checklist. A nested loop without its own `checklist` field iterates this one. |
| `done` | no | Engine-owned. Absent or `false` means open. |

Rules:

- An item with neither `command` nor `infer` passes when its lap returns.
  This is what a chapter item looks like: its validation is its sprints.
  The baby-bear reviewer is the guard against abusing it at the bottom
  level.
- The engine writes `done` in both directions. Every failed item becomes
  `done: false`; the framed item becomes `done: true` only when the entire
  validation set passes. The
  markdown body is preserved byte for byte. The frontmatter is re-encoded
  from the parsed node tree, so key order and comments survive as far as
  the YAML library allows.
- A hand-edited `done: true` joins the next validation set. A failure
  unmarks it and makes it eligible for selection again.
- Items are selected in file order: the first item whose `done` is not
  `true`. A failed item stays open and is therefore re-selected on the next
  arrival unless a planner inserted something before it. Planners may
  append, insert, reorder, or rewrite open items at any time; the loop
  re-reads on every arrival.
- Paths in the file (`command` cwd, `infer.files`, `doc`, `checklist`) are
  relative to the run workdir, which inside a parallel branch is the
  branch's worktree.

The parser lives in package `checklist`: `Load(path)`,
`(*Checklist).Open()` returning the first open item, `MarkDone(path, name)`,
and `UnmarkDone(path, name)`. Both writes are atomic and parse the result
before replacing the file.

## 3. The node

```yaml
- id: sprint
  type: loop
  checklist: ephemeral/projects/mvp/sprints/SPRINT-0002.md
  body: implement
  on_done: success
  max_visits: 40
  timeout: 15m
```

| Field | Type | Default | Meaning |
|---|---|---|---|
| `checklist` | string | unset | Checklist path, relative to the workdir. Optional only when this loop node lies inside another loop's body, in which case it iterates the enclosing item's `checklist` field. |
| `body` | string | required | Entry node of the lap. |
| `on_done` | string | required | Target when no open item remains: a node ID or `success`/`failure`. |
| `max_visits` | integer | unset | Arrivals at the loop node, laps plus one. The only loop ceiling. |
| `timeout` | duration | inherited | Ceiling on one item's `command`. Also the infer turn's timeout. |
| `llm_model`, `llm_provider`, `reasoning_effort` | string | `flash` / `gemini` / `medium` | The infer judge, selected independently of pipeline `defaults` (issue #37). `flash` resolves to `gemini-3.8-flash-medium`. |

Routing targets of a loop node are `body` and `on_done`, so the existing
`edge_target_exists` and `dead_end` rules cover them and the offered-set
mechanism applies: if `body` has exhausted its own `max_visits` while items
remain, the arrival fails the run with a terminal error, the same way an
exhausted successor fails any node today.

The chooser table (spec §3.3) gains a row: `loop` | the checklist | an open
item remains → `body`; none → `on_done`. No prose is interpreted.

The body's last node routes back to the loop node with an ordinary edge.
That edge is the lap's "I think this item is implemented" claim, and the
engine tests the claim before believing it.

## 4. Arrival semantics

```
FUNCTION loop.execute(node, offered, scope):
    list = checklist.Load(resolve_checklist_path(node))
    frame = frames.top_for(node)              -- in-memory, may be absent

    IF frame is present:                      -- a lap just returned
        item = list.find(frame.item)
        IF item is absent:                    -- renamed or removed by a planner
            note "item vanished"; frames.pop(node)
        ELSE:
            validation_set = every done item plus item, in file order
            results = []
            FOR candidate, index IN validation_set:
                result = validate(candidate,
                    stage_dir/validation-{index}.log, scope) -- §5
                results.append(result)        -- failures do not short-circuit
            write {validations: results} to stage_dir/validation.json
            append LoopValidated with results
            FOR failed IN results where passed is false:
                checklist.UnmarkDone(path, failed.item)
            IF every result passed:
                checklist.MarkDone(path, item.name)
                frames.pop(node)
            ELSE:
                checklist.UnmarkDone(path, item.name)
                frame.last_failures = every failed result's item, summary,
                    and log path
        list = checklist.Load(path)           -- re-read after marking

    next = list.Open()
    IF next is absent:
        frames.pop(node)                      -- if still present
        write frames.json
        RETURN Outcome{next: node.on_done, notes: "checklist complete"}

    IF frame is absent OR frame.item != next.name:
        frames.push_or_replace(node, {item: next.name, index, count, lap: 1,
            last_failures: frame.last_failures})
    ELSE:
        frame.lap += 1
    write frames.json
    RETURN Outcome{next: node.body, notes: "item i/n: name (lap k)"}
```

The frame stack is a field on the `Runner`, guarded by a mutex. Arriving at
a loop node whose frame is not on top of the stack pops everything above
it. (With the §7 lint rules as written, an edge from an inner body straight
to the outer loop node is illegal, because the body node-set follows every
edge except through its own loop node; the pop rule is defensive, not a
supported shape. If escape edges are wanted later, the node-set definition
must stop at enclosing loop nodes.)
Restart is coarse and dumb by design: nothing about the loop is
checkpointed. On resume the engine rewinds the checkpoint's next node to
the outermost loop whose body contains it (added after review: resuming at
a body node ran it frameless, and resuming at a nested loop with no
`checklist` field failed outright). That loop's first arrival validates
nothing and selects the first open item, which is the item the
interrupted lap was working on unless a planner changed the file.

## 5. Validation

Runs in the loop node's own stage directory, `stages/{seq}-{loop}/`. Every
validated item gets its own `validation-{checklist position}.log`, and the
loop timeout applies independently to each command.

1. **Command.** If present: `/bin/sh -c`, cwd the workdir, stdout and
   stderr to that item's validation log, process group killed on stop or timeout,
   exactly like a tool node. Exit 0 passes. Nonzero fails with an excerpt
   of the log as the summary: the whole trimmed log up to 2000 characters,
   otherwise its first 600 and last 1400 joined by one
   `… (N runes omitted) …` line, so a check's stated reason survives the
   diagnostics it dumps afterwards.
2. **Infer.** Only if the command passed or is absent, and `infer` is
   present. The globs are expanded relative to the workdir. No match is a
   failure ("no evidence files"). The judge is one codergen turn through
   the configured backend with `fidelity: none`, no thread, the loop node's
   model fields, and the standard choice schema with two offered targets
   whose conditions are the verdicts:

   ```
   You are validating one checklist item. Judge only what the evidence
   shows; do not fix anything.

   item: <name>
   check: <check>
   <infer.prompt>

   Evidence, open and inspect every file:
   - <path>
   - <path>
   ```

   Targets: `pass` ("The evidence demonstrates the check.") and `fail`
   ("The evidence does not demonstrate the check, or is missing."). The
   outcome's `next` is the verdict and its `notes` the reason. The judge
   runs in a coding harness with tools, so it can open images itself; the
   harness contract's text-only content parts are not a limitation here.
   In simulation mode (no backend) the verdict is `pass`.
3. `validation.json`: `{validations: [{item, command, log_path, exit_code,
   log_tail, infer: {files, verdict, notes}, passed, summary}, ...]}` in
   checklist order. `LoopValidated` carries the same list and an aggregate
   `passed` value.

## 6. Frame injection

**Hypothesis, not doctrine (Tyler, 2026-09-02).** What the frame carries
and how it is shaped is to be tested by use. Two candidates: (a) the
rendered block below, inner loops nested inside outer ones so the
structure is explicit; (b) breadcrumbs only, a few paths pointing at the
current iteration files, with the agent reading them itself. v1 ships
(a) because it needs no agent discipline; if chapter docs make it fat,
switch to (b) without ceremony.

For every codergen and fan-in turn executed while at least one loop frame
is active, the engine prepends one rendered block per active frame, each
inner loop's block nested inside its enclosing loop's block and indented
one level, before the node's own prompt. The block contains only engine
facts and file contents the engine copied, after a fixed preamble that
says what the blocks are (tag name `iterate` and the preamble per Tyler,
2026-09-02; `tractor` was a non-helpful name):

```
<system-message>
This is one step of a Tractor run inside a checklist loop. The iterate
blocks below are the engine's record of where you are: the item selected
for this lap, the check it must satisfy, the command and judge that will
validate it when this step ends, and what the previous validation reported.
Outer blocks enclose inner ones. Paths are relative to the working
directory. Do only what your prompt asks; the engine marks items done.
</system-message>
<iterate loop="sprint" checklist="ephemeral/projects/mvp/sprints/SPRINT-0002.md" item="1/2" lap="2">
name: Build the login screen
check: The login screen validates and submits on valid input
command: npx playwright test tests/login.spec.ts
infer: Judge whether every state of the login screen looks usable
  files: ephemeral/captures/login/*.png
last validation: failed
  - item: Build the login screen
    summary: exit 1 — <log excerpt>
    validation log: <absolute path of this item's validation log>
--- doc: ephemeral/projects/mvp/sprints/SPRINT-0002.md ---
<file contents>
</iterate>
```

`last validation` appears only after a failed lap and lists every failure
with its item, summary, and distinct validation-log path, so the agent can
read every complete log. The block clears once the next validation set
passes. `doc` and its contents
appear only when the item has one; an unreadable doc renders as
`doc: <path> (unreadable: <error>)` rather than failing the turn. The
rendered text travels in `ExecutionScope.Frame`; the codergen handler
prepends it, separated by a blank line, and it lands in `prompt.md` like
everything else. Supervisors do not receive frames. `$goal` expansion is
unchanged.

`{logs_root}/frames.json` is rewritten whenever the stack changes: an
array of `{loop, checklist, item, index, count, lap}`, outermost first. It
is a cache for observers, never read by the engine.

## 7. Lint

The **body node-set** of a loop node is every node reachable from `body`
without passing through the loop node. It may contain further loop nodes,
parallel nodes, and edges to `failure` (an escape hatch). It is delimited
the way a parallel branch is (spec §4.6).

| Rule | Severity | Description |
|---|---|---|
| `loop_body_entry` | ERROR | A body node may be entered only from another body node, or, for the body root, from its loop node's `body` field. `on_done` may not name a body node. |
| `loop_body_returns` | ERROR | The loop node must be reachable from its body root within the body node-set; otherwise no lap ever returns to be validated. |
| `loop_checklist_required` | ERROR | A loop node without `checklist` must lie inside another loop node's body. |
| `loop_in_parallel` | ERROR | A loop node may not lie inside a parallel node's branch node-set. The frame stack is run-wide and branches run concurrently. |

`edge_target_exists`, `edge_target_unique`, `dead_end`, `max_visits_positive`,
and `reachability` apply to loop nodes through `RoutingTargets` and
`MaxVisits`. The rule count goes from 27 to 31.

## 8. Run directory additions

```
frames.json                       -- current frame stack, observer cache (§6)
stages/{seq}-{loop_id}/
    validation-{item_position}.log -- one command log per validated item
    validation.json               -- the ordered validation list (§5)
    outcome.json                  -- as for every node
```

The infer judge's turn writes its own `prompt.md` and `response.md` into
the same stage directory and allocates a run-log segment like a codergen
node.

## 9. Example

`examples/loops/checklist-loop.yaml`, mirrored into the skill bundle:

```yaml
name: checklist-loop
goal: Work through the checklist until every item is demonstrated.
start: items

defaults:
  timeout: 600s
  fidelity: none
  max_retries: 1
  llm_provider: anthropic
  llm_model: claude-sonnet-5
  reasoning_effort: low

nodes:
  - id: items
    type: loop
    checklist: ephemeral/projects/demo/checklist.md
    body: implement
    on_done: success
    max_visits: 20

  - id: implement
    type: codergen
    prompt: Implement the current checklist item so that its check holds.
    edges:
      - to: items
```

The prompt is deictic on purpose: the frame says what the current item is,
what will be run against it, and what went wrong last time.

## 10. Not in v1

Per-item attempt budgets. Fixup or escalation routing on repeated failure.
Frame persistence across restarts. Rendering the frame stack in run status
(`frames.json` exists for that later). Parallel laps. Futility detection.
Human attestation as an item field. Item-level `checklist` at more than one
level is supported by construction but untested beyond two levels.

## 10a. Implementation notes (2026-09-02)

Built on this branch as `graph.LoopNode`, package `checklist`,
`engine/loop.go`, `engine/frames.go`, `engine/shell.go` (the shell core
shared with the tool handler), and four lint rules. Small choices the note
left open, as implemented:

- Every entry in `validation.json.validations` carries `command`,
  `log_path`, `exit_code`, and `log_tail` (empty or zero when the item has
  no command); `infer` is present only when the judge ran. `summary` on a
  pass is the string `passed`. An item with no command still gets its own
  empty validation log.
- A vanished item is reported as an `item vanished: <name>;` prefix on the
  arrival's outcome notes.
- The item text in a frame is captured at push time; only the `doc` file
  is re-read at render time.
- After a failed validation set, the next selected item carries the full
  failure block, including failures belonging to earlier done items.
- `max_visits: N` on the loop node allows N−1 laps; the failure surfaces as
  the body's "every successor has exhausted its visit budget" before the
  body would run again.
- The judge's `prompt.md` and `response.md` land in the loop node's stage
  directory, beside `validation.json`.
- Timeline events: `LoopItemSelected {node,item,index,count,lap}`,
  `LoopValidated {node,passed,validations}`, `LoopCompleted {node,count}`.

Proof: unit tests for §11 claims 1–6 in the respective packages; the live
run for claim 7 is recorded in `proof/loop-node-live/`.

## 11. Claims to demonstrate before this merges

1. A loop over a three-item checklist whose commands all exit 0 completes
   in three laps; every item ends `done: true`; the markdown body is
   byte-identical to the original.
2. An item whose command exits nonzero is re-entered; the next lap's
   `prompt.md` contains its failed-validation summary and log path; once the
   command passes the item is marked and the loop advances.
3. Every done item is revalidated on each lap return. A regression is
   unmarked, the framed item remains open unless the whole set passes, and
   `validation.json`, `LoopValidated`, and the next frame carry the complete
   ordered results or failures with distinct log paths.
4. With a backend whose judge answers `fail` then `pass`, an `infer` item is
   marked only after the `pass`.
5. Nested: an outer checklist whose items carry `checklist:` drives an inner
   loop with no `checklist` field; the outer item is marked only after the
   inner loop exits; `prompt.md` inside the inner lap carries both frames.
6. The four lint rules fire on the shapes they name and stay silent on the
   example.
7. `tractor print-schema` includes the loop node and the example validates.
8. A live run of `checklist-loop.yaml` with the Claude harness against a
   scratch repository and a two-item checklist reaches `COMPLETED`, with
   per-item validation logs present for each lap and the checklist marked by the
   engine, not the agent.
