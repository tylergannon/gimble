# Sprint 2: the run directory reaches agents

`tractor ask` finds the timeline through `TRACTOR_RUN_DIR`. This sprint
makes the engine set it, for every process a run starts.

## Engine

- When a run starts (`engine.Runner.Run`, or wherever the logs root is
  fixed), set `TRACTOR_RUN_DIR` to the absolute logs root in the engine
  process's own environment, so every child inherits it. One run per
  process is the existing rule; do not build per-turn environments.
- Confirm each harness backend (`harness/claude`, `harness/codex`,
  `harness/agy`) passes the process environment through to the CLI it
  spawns. `harness/codex/rpc.go` captures `os.Environ()` when its process
  config is built; make sure that happens after the variable is set, or
  read the environment at spawn time. Tool nodes go through
  `engine/shell.go`; confirm the same there.
- Emit nothing new from the engine. The `QuestionAsked` line is written by
  `ask` (sprint 1).

## Tests

- A unit test per backend that the spawned process environment contains
  `TRACTOR_RUN_DIR` when it is set before the backend starts.
- An engine test with the scripted backend that a tool node's command sees
  the variable and that it equals the run's logs root.

## Live proof

`check-sprint-02.sh` beside this doc runs `proof-pipeline.yaml` with the
built binary and the codex backend. The pipeline's single codergen node is
told to ask the reviewer for a secret word with `tractor ask` and write the
answer to a file. The script watches the timeline for `QuestionAsked`,
answers with a nonce, and expects the run to complete with the nonce in the
file. Read both before you start. This is the chapter's review posture
made executable: a real agent, a real run, a real question.

## Ask the reviewer

- Whether `TRACTOR_RUN_DIR` should also be written into the `<system-message>`
  preamble so agents on backends that scrub the environment can pass
  `--run <dir>` to `ask` explicitly. Recommend one answer.
