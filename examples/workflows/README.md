# Workflow proposals

Five workflows ported from the diffusioninc skills mounted at
`reference/diffusion-skills`. Each one is a graph the engine can run today;
the diagrams in `ephemeral/projects/tractor/workflows/diagrams/` were
rendered from these files with `tractor edit`.

They are proposals, not built-ins. Nothing in the binary loads them yet.

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

## What they need that does not exist yet

- `sprint-execute` and `chapter-loop` want ledgers as markdown with YAML
  frontmatter. Upstream keeps them as `ledger.yaml`; the checklist format is
  markdown with frontmatter so the definition of done can sit in prose beside
  the items.
- `promise-loop` needs the promise tool on PATH.
- All five would rather declare their doctrine as skills than carry it in
  prompts.
