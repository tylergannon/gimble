# The loop node, v1

Design note for the first implementation, written 2026-09-02 from the
loop-frames memo (§9 as amended by §9a) and the interview decisions in
`decisions.md`. This is the spec the code on branch `worktree-goal-gates`
is built to. Expect it to be reworked after Tyler reviews the result.

## 1. What it is

A `loop` node iterates a checklist file. On every arrival it re-reads the
file, validates the item the previous lap worked on, marks that item done
if the validation passed, selects the first item not yet done, injects it
as the frame for the lap, and dispatches the body. When no item is left it
routes to `on_done`.

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
- The engine writes exactly one thing: `done: true` on one item. The
  markdown body is preserved byte for byte. The frontmatter is re-encoded
  from the parsed node tree, so key order and comments survive as far as
  the YAML library allows.
- A hand-edited `done: true` is honored without validation. It is the
  human override, and the file is the truth.
- Items are selected in file order: the first item whose `done` is not
  `true`. A failed item stays open and is therefore re-selected on the next
  arrival unless a planner inserted something before it. Planners may
  append, insert, reorder, or rewrite open items at any time; the loop
  re-reads on every arrival.
- Paths in the file (`command` cwd, `infer.files`, `doc`, `checklist`) are
  relative to the run workdir, which inside a parallel branch is the
  branch's worktree.

The parser lives in a new package `checklist`: `Load(path)`,
`(*Checklist).Open()` returning the first open item, `MarkDone(path, name)`.

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
| `llm_model`, `llm_provider`, `reasoning_effort` | string | inherited | The infer judge. Meant for a cheap model. |

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
        ELSE IF item.done:                    -- hand-marked
            frames.pop(node)
        ELSE:
            result = validate(item, scope)    -- §5
            write stage_dir/validation.json, validation.log
            IF result.passed:
                checklist.MarkDone(path, item.name)
                frames.pop(node)
            ELSE:
                frame.last_failure = result.summary
        list = checklist.Load(path)           -- re-read after marking

    next = list.Open()
    IF next is absent:
        frames.pop(node)                      -- if still present
        write frames.json
        RETURN Outcome{next: node.on_done, notes: "checklist complete"}

    IF frame is absent OR frame.item != next.name:
        frames.push_or_replace(node, {item: next.name, index, count, lap: 1, last_failure})
    ELSE:
        frame.lap += 1
    write frames.json
    RETURN Outcome{next: node.body, notes: "item i/n: name (lap k)"}
```

The frame stack is a field on the `Runner`, guarded by a mutex. Arriving at
a loop node whose frame is not on top of the stack pops everything above
it: an inner loop that was bypassed by an escalation edge is abandoned.
Restart is coarse and dumb by design: after a resume the stack is empty,
the first arrival validates nothing and selects the first open item, which
is the item the interrupted lap was working on unless a planner changed
the file. Nothing about the loop is checkpointed.

## 5. Validation

Runs in the loop node's own stage directory, `stages/{seq}-{loop}/`.

1. **Command.** If present: `/bin/sh -c`, cwd the workdir, stdout and
   stderr to `validation.log`, process group killed on stop or timeout,
   exactly like a tool node. Exit 0 passes. Nonzero fails with the last
   400 characters of the log as the summary.
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
3. `validation.json`: `{item, command, exit_code, log_tail, infer: {files,
   verdict, notes}, passed, summary}`. The summary is what the next lap's
   frame carries as `last validation`.

## 6. Frame injection

For every codergen and fan-in turn executed while at least one loop frame
is active, the engine prepends one rendered block per active frame,
outermost first, innermost last, before the node's own prompt. The block
contains only engine facts and file contents the engine copied:

```
<tractor loop="sprint" checklist="ephemeral/projects/mvp/sprints/SPRINT-0002.md" item="1/2" lap="2">
name: Build the login screen
check: The login screen validates and submits on valid input
command: npx playwright test tests/login.spec.ts
infer: Judge whether every state of the login screen looks usable
  files: ephemeral/captures/login/*.png
last validation: failed — exit 1 — <log tail>
--- doc: ephemeral/projects/mvp/sprints/SPRINT-0002.md ---
<file contents>
</tractor>
```

`last validation` appears only after a failed lap. `doc` and its contents
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
    validation.log                -- the item command's stdout and stderr
    validation.json               -- the validation record (§5)
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

## 11. Claims to demonstrate before this merges

1. A loop over a three-item checklist whose commands all exit 0 completes
   in three laps; every item ends `done: true`; the markdown body is
   byte-identical to the original.
2. An item whose command exits nonzero is re-entered; the next lap's
   `prompt.md` contains `last validation: failed` and the log tail; once the
   command passes the item is marked and the loop advances.
3. With a backend whose judge answers `fail` then `pass`, an `infer` item is
   marked only after the `pass`.
4. Nested: an outer checklist whose items carry `checklist:` drives an inner
   loop with no `checklist` field; the outer item is marked only after the
   inner loop exits; `prompt.md` inside the inner lap carries both frames.
5. The four lint rules fire on the shapes they name and stay silent on the
   example.
6. `tractor print-schema` includes the loop node and the example validates.
7. A live run of `checklist-loop.yaml` with the Claude harness against a
   scratch repository and a two-item checklist reaches `COMPLETED`, with
   `validation.log` present for each lap and the checklist marked by the
   engine, not the agent.
