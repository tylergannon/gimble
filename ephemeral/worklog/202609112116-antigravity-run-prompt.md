# Antigravity harness and run-prompt

decision: Validation is the real `gimble run-prompt` command invoking `agy`, with the emitted result and durable Gimble events inspected; unit tests and builds are supporting checks.
decision: Port legacy behavior by hand against the current HarnessAdapter and event contracts; legacy is inspiration, not API authority.
friction: A fresh worktree's `go test -race ./...` fails in web tests because the embedded build lacks `skgo.manifest.json` -> run `just build` before the full Go suite so generated web assets exist.
decision: `agy` 1.2.1 rejects `/fork` in print mode and exposes no headless fork flag, so the adapter returns an explicit unsupported error instead of aliasing two Gimble sessions to one mutable conversation.
decision: The live Flash-low proof used `gimble run-prompt` to create and read `agy-proof.txt`, returned schema-valid JSON, projected the real tool and text steps into observation state, and ended the durable run with `recording_error: ""`.
