# Compiling examples

Current executable sources are `bakeoff/main.go`, `critique/main.go`, `sprints/workflow.go`, `context/main.go`, and `scopes/main.go` under `examples/go-workflows/`.

Their shared API is `examples/go-workflows/internal/program/`: typed canned agent calls, canned command observations, symbolic worktrees, no-op integration, no-op context/scopes, and one-pass loops. Build with `go build ./examples/go-workflows/...`.
