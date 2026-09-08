---
name: tractor
description: Run coding agents as detached pipelines with Tractor — loop until a check actually passes, fan out across Claude/Codex/Gemini, or have a second model review the work. Use when the user says "don't stop until the tests pass", "loop on this until it works", "keep working while I'm gone", "keep going after I close my laptop", "run this in the background", "try a couple of approaches in parallel", "have another model check this", "get a second opinion from Codex or Gemini" — or asks to start, check on, steer, or stop a Tractor run. Starts from copy-and-run examples; no graph authoring needed. Do not use for ordinary single-agent work that finishes in this session.
---

# Tractor

Tractor runs pipelines of real coding agents — the Claude Code, Codex, and
Gemini CLIs with native sessions — as detached processes that outlive this
session. A pipeline is a small YAML or JSON file: start from an example and
change two strings; you rarely author a graph.

MCP tools: `start_run`, `get_run_status`, `steer_run`, `stop_run`,
`get_pipeline_schema`. They take file paths, not inline graphs.

## The whole idea

```yaml
goal: Implement the TODO in cmd/server/routes.go and make the tests pass
start: implement
nodes:
  - id: implement
    type: agent # an agent turn
    max_visits: 5 # the budget; nothing else stops a loop
    prompt: $goal # the goal is the whole prompt
    edges:
      - to: check
  - id: check
    type: command # a command decides what "done" means
    command: go test ./...
    edges:
      success: success
      error: implement # failure routes back — that's the loop
```

An agent works, a command decides, failure routes back. Fan-out shapes add
`fan_out` branches (one per provider) and a `fan_in` node to judge.

## Writing node prompts

State the condition the agent is to bring about, never the steps: the
repository, its skills, and its tools already teach how. The engine
substitutes `$goal` (nothing else delivers the goal text), supplies the
workspace, and builds a structured routing choice from the node's edge
conditions — so never coach routing ("respond with pass"), never point at
files that teach ("read AGENTS.md first"), never script the checks.
"Validate the work done in this loop" beats a list of commands.

## Pick the user's moment

| The moment                               | Example                         |
| ---------------------------------------- | ------------------------------- |
| "Don't stop until it actually works"     | `examples/fix-until-green.yaml` |
| "Have another model check this"          | `examples/critique-circle.yaml` |
| "Try a couple of approaches in parallel" | `examples/bake-off.yaml`        |
| "Keep working on this after I leave"     | `examples/milestone-loop.yaml`  |
| "Work through a list, proving each item" | `examples/checklist-loop.yaml`  |

They ship beside this file, mirroring `examples/loops/` in the Tractor repo;
if neither is at hand, the pipeline above is a complete start.

## Or run a workflow that already ships

Those examples are single shapes to copy and edit. Whole workflows ship inside
the binary and run by name, with nothing to copy. `tractor workflows` lists
what the installed binary carries and the situation each one is for — run it
rather than trusting this table, which cannot know the installed version.

| The moment                                      | Workflow         |
| ----------------------------------------------- | ---------------- |
| "Work the sprint backlog until it's done"       | `sprint-execute` |
| "Run the whole chapter, not one sprint"         | `chapter-loop`   |
| "Plan the next sprint properly"                 | `sprint-plan`    |
| "Build what this spec describes while I'm gone" | `delivery-loop`  |

Pass the name where a pipeline path goes; a name resolves to a built-in only
when no file of that name exists.

Pass the user's ask as `--goal`: it replaces the pipeline's goal, and every
workflow carries that into its working prompts. `sprint-plan` and
`delivery-loop` are meaningless without one and refuse to start; the listing
marks them. `sprint-execute` and `chapter-loop` iterate a ledger that must
already be in the workspace — the listing names the path, and a missing one
fails on the first node — so for those a goal steers the run rather than
assigning it. None of them hardcodes a test command; an item's own command is
the gate, so they work in any language. `tractor workflows show <name>` prints the
pipeline, so redirect it into a file when the user wants one changed, then use
the copy. Prefer a built-in over authoring a graph when the moment matches one:
they carry engine-owned done, cross-provider validation, and supervisors
already.

1. Copy the chosen example into the target project (a git repo).
2. Replace its placeholders with the user's goal; command-gated loops also
   need the check node's `command`. Read the file — examples differ.
3. `start_run` with the copied file as `pipeline_path` and the repo as
   `workdir` — it lints the graph first and refuses to launch a broken one,
   so fix what it rejects and call it again. A fresh run needs an empty or
   absent `logs_root`; `resume: true` continues an existing run directory.
4. Report the returned `run_id`, process ID, and log paths. Run state lives
   under `~/.local/state/tractor/mcp-runs` and outlives this session and any
   MCP restart — a later session reconnects with the same `run_id`.

## Starting the software under test

Set `system_file` in the workflow to `Procfile` or a name such as
`Procfile.dev`, and set `services` to an array containing the one process the
workflow must reach. Tractor runs the whole Procfile through Overmind for the
run. Overmind supplies each process with `PORT`; Tractor exports the named
process's port as `TRACTOR_SERVICE_PORT` to every workflow shell. Construct the
address in the validation command because the workflow owns its protocol and
path. Multiple named services and other system-file formats are not supported
yet. A workflow without `system_file` and `services` behaves exactly as before;
declaring only one of those fields is invalid.

## Make "done" honest

In a command-gated loop, the command's exit code decides the route. Point it
at the closest observable proof of the user's claim — run the app, curl the
endpoint, assert on the artifact. In a checklist loop, commands supply
mechanical evidence, item judges assess inferred evidence, and the goal
evaluator decides whether the checklist's definition of done is satisfied.

## Checklists

`checklist-loop.yaml` puts a `loop` node in front of the agent. It iterates
a markdown file whose YAML frontmatter lists items — `name`, `check`,
optionally a `command` (exit 0 passes) and an `infer` judge (`files`,
`prompt`) — and on every lap return validates the framed item plus every done
item. After a passing set, an evaluator reads the checklist body's definition
of done and either exits or re-reads the evaluator-editable ledger and injects
its first open item into the prompt as a frame (name, check, command, last
failure, doc). The agent never marks items. Copy
`examples/checklist-loop.md` to the path the pipeline names and edit its
items; `validation.json` and its per-item log paths say why an item stayed
open.

## While it runs

`get_run_status` returns `current_node`, `last_stage`, and `last_response`
— enough for a progress update in the user's terms. `steer_run` reaches the
active steerable turn; when none is active it returns `accepted: false`, so
retry after the next status check rather than reporting a failure. Steer only
when new authoritative information arrives or the run is leaving scope; a
complete goal up front beats frequent correction. `stop_run` asks a run to
stop; calling it again after the graceful window forces it.

An agent inside a run can ask its caller a blocking question by writing one
question per Markdown or HTML file and running:

```sh
TRACTOR_INTERVIEW_DIR=ephemeral/projects/<build>/interview tractor ask question.md
```

The command moves the question to a numbered file, records `QuestionAsked`
when `TRACTOR_RUN_DIR` is available, and prints the answer once another
process runs `tractor answer <numbered-question> [text]`. Answers may come
from stdin; an existing answer is never overwritten. If a shell session ends
while waiting, rerun `tractor ask` with the numbered question path to resume.

On the calling side, watch the returned run path's `timeline.jsonl` for
`QuestionAsked`, open the event's `question` path, then run `tractor answer`
with that numbered path. The answer file unblocks the same agent turn; do not
steer or restart the run to deliver it.

## When a loop misbehaves

| Symptom                                       | Fix                                                                                                                                                                                                         |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Runs forever                                  | `max_visits` on the looping node.                                                                                                                                                                           |
| Says it's done when it isn't                  | The command is checking the wrong thing — check the behavior itself.                                                                                                                                        |
| Re-derives the same dead end every lap        | Have the prompt keep a short notes file: append what the next attempt should do differently, read it first.                                                                                                 |
| Reviewer rubber-stamps                        | Fresh session (`fidelity: none`), whole target every round, never say what to find. A different provider makes the independence real.                                                                       |
| Fan-in averages instead of deciding           | Tell it to inspect and run the work itself and adjudicate each finding on evidence — never count votes or concatenate reports.                                                                              |
| Guesses at a decision that wasn't its to make | Give it a door: an edge conditioned on "this decision isn't mine" leading to a node that asks the user or writes a report and routes to `failure`. Agents improvise when forward is the only route offered. |
| Builds everything, nothing runs until the end | Steer the chooser to vertical slices: a step is done when you can run something that proves it. Stack-order plans are the model's default tic; say no to them in the prompt.                                |

## Authoring beyond the examples

`get_pipeline_schema` returns the current graph schema, and `start_run`'s
lint diagnostics teach as they reject. Model consumers use an atomic `model`
object with required `name` and optional string `version` and `effort`;
provider and harness are derived. `tractor inspect-models <file>` displays the
effective selection and provenance for every node and hidden loop role. Keep
each prompt to the decision its node owns.

`tractor edit <file>` opens the pipeline in a browser graph editor with the
same lint. It reads and writes the file on disk, so keep editing the YAML
while the person has it open; the page reloads on every change and saves its
own edits back beside yours.
