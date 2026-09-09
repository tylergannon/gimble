# Gimble examples

These examples exercise Gimble through its real coding-agent harnesses. They
are intentionally small because their purpose is to prove orchestration, not to
spend tokens on a substantial coding task.

- [`steering/external-steering.json`](steering/external-steering.json) proves
  that an external operator can steer one active Codex turn. The workflow only
  completes if the live agent acts on the instruction.
- [`parallel/fan-out-fan-in.json`](parallel/fan-out-fan-in.json) proves that two
  tool branches execute concurrently in isolated Git worktrees and that one
  Codex fan-in turn reads both products into the main workspace.
- [`yaml/multiline-tool.yaml`](yaml/multiline-tool.yaml) proves that YAML
  comments and literal block commands work through the real CLI without an LLM
  turn.
- [`supervisor/live-steering.json`](supervisor/live-steering.json) proves that
  a Claude supervisor patrol can steer a live Codex worker and cause an
  independently verified workspace effect.
- [`loops/`](loops/README.md) are the starting points for iterative work:
  fix-until-green, critique circle, bake-off, a milestone loop, and a
  checklist loop. Copy one, change the goal and the check, run it.

Each directory documents the invocation and the observable success criteria.

Larger workflows do not live here: they ship inside the binary. Run `gimble
workflows` to list them, or read them in
[`internal/workflows/`](../internal/workflows/README.md).
