# Loops

A loop is a thing that runs iteratively until it's done. In Tractor that is
two nodes pointing at each other. Start from whichever example matches your
moment, change the goal and the check, and run it:

```sh
tractor run examples/loops/<file>.yaml
```

- [`fix-until-green.yaml`](fix-until-green.yaml) — the hello world. An
  agent works, a command decides, failure routes back, `max_visits` stops
  it from running forever. For *"don't stop until it actually works."*
- [`critique-circle.yaml`](critique-circle.yaml) — three providers write a
  proposal on the same topic, then each critiques the other two. For
  *"have a couple of models cross-check this."*
- [`bake-off.yaml`](bake-off.yaml) — three attempts at the same task in
  isolated worktrees; a judge runs each one and merges the winner. For
  *"try a couple of approaches in parallel."*
- [`milestone-loop.yaml`](milestone-loop.yaml) — a chooser picks the next
  bite-sized runnable step, an implementer does it, a command checks,
  repeat until the goal's claims are demonstrable. For *"keep working on
  this after I close my laptop."*
- [`checklist-loop.yaml`](checklist-loop.yaml) — a `loop` node works
  through a checklist file, proving each item by running its command
  (and its judge) before marking it done and moving on. For *"work
  through this list and prove each item."*

## Writing node prompts

State the condition the agent is to bring about, never the steps — the
repository, its skills, and its tools already teach how. The engine
substitutes `$goal` (nothing else delivers the goal text), supplies the
workspace, and builds a structured routing choice from the node's edge
conditions. Never coach routing, never point at files that teach, never
script the checks: `prompt: $goal` is a complete, correct prompt.

## Checklists

`checklist-loop.yaml` iterates a markdown file whose YAML frontmatter is
the ledger. Copy [`checklist-loop.md`](checklist-loop.md) to the path the
pipeline names and edit its items:

```yaml
---
items:
  - name: Print a greeting                   # unique within the file
    check: Running `sh hello.sh` prints exactly HELLO_LOOPS
    command: test "$(sh hello.sh)" = HELLO_LOOPS   # exit 0 passes
  - name: Document the scripts
    check: NOTES.md tells a newcomer how to run both scripts
    infer:                                   # a cheap judge reads the evidence
      files: NOTES.md
      prompt: Judge whether a newcomer could run both scripts from these notes alone
---
Prose below the frontmatter is the definition of done the evaluator reads.
```

On every lap return the loop re-reads the file and validates the framed item
plus every done item (command, then `infer` judge). After a passing set, an
evaluator reads the definition of done and either exits or re-reads the ledger
and injects its first open item into the body's prompt as a frame: name, check,
command, the last failure, and the item's `doc` if it has one. The evaluator
may append, reorder, or rewrite open items. Agents never mark items; a
hand-edited `done: true` requests validation rather than bypassing it. Paths
are relative to the workdir. `max_visits` on the loop node is the budget, and
`validation.json` plus its per-item log paths say why an item stayed open.

## When a loop misbehaves

You don't need this table to write a loop; you need it when a loop annoys
you.

| Symptom | Fix |
|---|---|
| Runs forever | `max_visits` on the looping node. That's the budget; there is no other ceremony. |
| Says it's done when it isn't | Make "done" a `command` node. If the tests pass but the feature doesn't work, the command is checking the wrong thing — check the behavior you actually want. |
| Re-derives the same dead end every lap | Tell the prompt to keep a short notes file: "append what the next attempt should do differently; read it first." |
| Reviewer rubber-stamps | Don't tell it what to find or ask it to confirm your fix. Fresh session (`fidelity: none`), whole target, every round. A different provider makes the independence real. |
| Fan-in averages instead of deciding | Tell it to inspect the work itself and adjudicate each finding with evidence — never count votes or concatenate reports. |
| Agent guesses at a decision that wasn't its to make | Give it a door: an edge whose condition is "this decision isn't mine," leading to a node that asks a human (Slack, an issue) or writes a report and routes to `failure`. Agents improvise when forward is the only offered route. |
| Builds everything, nothing runs until the end | Steer the chooser to vertical slices: "the step is done when you can run something that proves it." Stack-order plans (schema → services → API → UI) are the model's default tic; say no to them in the prompt. |
| A long run starts believing its own stale plans | Per-lap plans are working notes, not authority; they live with the run, the code is the record. Don't commit them. |
