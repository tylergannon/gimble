# Antigravity `run-prompt` attestation

Model: `gemini-3.8-flash-low` through `agy` 1.2.1.

Proved code commit: `e06570416ad7f975107c235348149d3da3b548ea`.

The built `bin/gimble` ran `run-prompt` with an exact JSON Schema. The prompt
required Antigravity to use `run_command` to create `workspace/agy-proof.txt`,
read it, and report the current working directory. The command returned:

```json
{"content":"ANTIGRAVITY_OK","workdir":"/Users/tyler/.codex/worktrees/0c2f/gimble/ephemeral/attest/antigravity-run-prompt/workspace"}
```

`workspace/agy-proof.txt` is exactly `ANTIGRAVITY_OK\n`. The durable run under
`logs/runs/20260911-214259.run-prompt/` ends with `complete` and an empty
`recording_error`. Its session log includes translated step, tool input/call/
success, text delta/final, accounting, and execution-success events. The final
`observation.json` reduces those events to an idle session containing the
executed `run_command`, its output, and the assistant's response.
