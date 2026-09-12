# Antigravity `run-prompt` attestation

Model: `gemini-3.8-flash-low` through `agy` 1.2.1.

Proved code commit: `66787eee1329efc65eed824473743156cc89d453`.

The built `bin/gimble` ran `run-prompt` with an exact JSON Schema. The prompt
required Antigravity to use `run_command` to create `workspace/agy-proof.txt`,
read it, and report the current working directory. The command returned:

```json
{"content":"ANTIGRAVITY_OK","workdir":"/Users/tyler/.codex/worktrees/0c2f/gimble/ephemeral/attest/antigravity-run-prompt/workspace"}
```

`workspace/agy-proof.txt` is exactly `ANTIGRAVITY_OK\n`. The durable run under
`logs/runs/20260911-214914.run-prompt/` ends with `complete` and an empty
`recording_error`. Its session log includes translated step, tool input/call/
success, text delta/final, accounting, and execution-success events. The final
`observation.json` reduces those events to an idle session containing the
executed `run_command`, its output, and the assistant's response.
Provider provenance records the real agy conversation ID
`e786baf4-c7fa-4476-8fb1-e4c9716baf81`; no hidden bootstrap turn precedes the
visible prompt.
