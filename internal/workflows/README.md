# Built-in workflows

The pipelines that ship inside the binary. They are embedded from this
directory, so editing one here changes what `tractor run <name>` runs after a
rebuild, and `workflows.go` is the catalogue that gives each one the line
`tractor workflows` prints. A test holds the catalogue and these files to each
other, so a workflow cannot ship undocumented or be documented without
shipping.

```sh
tractor workflows                       # the catalogue
tractor workflows show sprint-execute   # one pipeline, verbatim
tractor run sprint-execute --workdir . --logs .tractor/run
tractor run delivery-loop --goal "Build what docs/SPEC.md describes" \
  --workdir . --logs .tractor/run
```

A workflow either reads its work from the workspace — a ledger, or
`PROMISE_ID` in the environment — or works on whatever the operator names and
takes it from `--goal`, which replaces the pipeline's goal so every `$goal` in
a prompt expands to it. The catalogue records which, and one that needs a goal
refuses to start without it.

They were ported from the diffusioninc skills mounted at
`reference/diffusion-skills`. The diagrams in
`ephemeral/projects/tractor/workflows/diagrams/` were rendered from these
files with `tractor edit`.

None of them has been run end to end yet.

## What changed in the port

The upstream skills are pipeline engines written as prose. One of them embeds
a YAML workflow definition — agents, per-role models, empty or continued
context, visit directories, visit caps, an outcome file for routing, session
resume, artifact checks — and then spends two hundred lines instructing a
model to be that runtime by hand. The engine already is that runtime, so the
port is mostly deletion:

- Preflight blocks that check for installed CLIs and valid keys go away.
  `start_run` resolves harnesses and lints the graph before a token burns.
- Backgrounded shells, `wait`, and the recheck-then-do-not-relaunch policy go
  away. The engine runs the branches.
- The outcome file and its allowed-values discipline go away. Routing is a
  schema-enforced choice among the successors a node was offered.
- Checklist protocols go away. The engine flips done in both directions, so
  no rule is needed about which agent may check a box.

What is left over is doctrine: how a sprint is shaped, what a chapter may not
contain, what makes a promise falsifiable. That belongs in a skill attached to
the node that needs it, which is issue #56. Until then it lives in the prompts
here, marked in each file.

## Model roles

The same few roles recur across all five, so the choices follow one convention
rather than being decided node by node:

- **manager** — the planner, the merge, and every loop's goal evaluator: `fable`
- **workhorse** — the default for ordinary turns: `claude-sonnet-5`, raised to
  `claude-opus-5` where a node has to reason over a whole codebase
- **validation** — the reviewer, always on the other provider from whoever
  wrote the code: `gpt`
- **evidence** — the item judge over screenshots and other artifacts: the
  engine's own default, gemini flash

The convention is retyped in every file, and the workhorse cannot even be
named briefly: `gpt`, `flash`, and `fable` are maintained aliases, while
sonnet and opus have to be spelled as provider-native IDs. Named runtime
presets would let these files say `model: workhorse` and let the operator
decide what that resolves to, which is issue #58.

## The workflows

**`sprint-execute.yaml`** — the sprint loop. The sprint ledger is the
checklist; each item carries the command that demonstrates that sprint. The
reviewer runs on the other provider with fresh context and cannot leave the
loop, because leaving is the loop node's decision against the ledger's
definition of done.

**`chapter-loop.yaml`** — the same shape nested. A chapter item names its own
sprint ledger, the inner loop iterates it, and the coding and validation loop
sits inside that. One run, no child runs per chapter. A chapter's proof is its
item's command and infer judge, so the chapter exit gate is expressed where
the engine already owns it.

**`sprint-plan.yaml`** — competitive planning. Three providers draft the same
intent in isolated worktrees, then each critiques the two drafts it did not
write. The interview is a real gate: one node writes the questions, a command
node blocks on `tractor ask` until the answers appear, and the merge reads
them. Drop those two nodes for an unattended plan.

**`delivery-loop.yaml`** — spec to software. Plan, cross-provider critique,
plan update, then a loop over the plan's own checklist. The plan is written as
a checklist file, so the engine validates each item rather than the reviewer
being trusted to check boxes honestly.

**`promise-loop.yaml`** — one repository promise advanced by bounded runs.
Per-gate specifics live in each gate item's command, so the graph does not
need to know what any one promise measures. Non-fulfilment finishes the run
with a verdict instead of looping until something passes.

## Supervisors

Each workflow carries supervisors, named by the question they ask. A supervisor
runs on its own patrol clock outside the walk, sees the live snapshot and the
attempt digests, and either says ok or steers one named target with a message
delivered to it verbatim. Most patrols should be ok; a supervisor that steers
every time is noise.

They run on the workhorse tier rather than the manager tier. The judgment is
narrow — is this in the requirement or not — and it repeats on a clock for as
long as the run lasts, which is workhorse-shaped work. A supervisor keeps one
native session for the whole run, so it is not re-deriving the world each
patrol, but the nudge hands it paths rather than content, so every patrol
starts with a round of reading. That is issue #59; until it lands, the interval
is what governs the cost.

One limit is worth knowing before relying on them: a supervisor coaches turns
that are live when it patrols. A planning node that finishes in a couple of
minutes may never be seen. `done_is_demonstrable` patrols faster than the
others for that reason, but plan-shape doctrine is ultimately better placed in
the planner's own instructions than in a patrol.

**`serves_the_requirement`** watches the coding and validating agents for work
nobody asked for: speculative fixes, hardening against conditions the
requirement never mentions, edge cases outside the stated outcome, a second
mechanism where one already works.

**`proof_is_evidence`** watches the validating agents for proof becoming a
project of its own. Proof means running the software and recording what
happened. It is not an ideal to be approached and never reached, and it is not
a reason to invent requirements the work never promised. The steer is always
toward the concrete: run the thing, capture the output or the screenshot, say
whether it worked.

**`done_is_demonstrable`** watches the planning agents for definitions of done
that are lists of commands. A definition of done that reads

    run go test ./...
    run golangci-lint run ./...
    run lefthook run pre-commit

says nothing about whether the feature exists. One that reads

    Build the feature described in the goal document. Write the scenarios in
    Gherkin, implement them, use them to demonstrate the feature in a real
    browser, then collect screenshots and look at them to confirm the
    behavior is correct.

says what done looks like and leaves the route to the implementer.

## What they need that does not exist yet

- `sprint-execute` and `chapter-loop` want ledgers as markdown with YAML
  frontmatter. Upstream keeps them as `ledger.yaml`; the checklist format is
  markdown with frontmatter so the definition of done can sit in prose beside
  the items.
- `promise-loop` needs the promise tool on PATH.
- All five would rather declare their doctrine as skills than carry it in
  prompts.
