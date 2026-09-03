---
title: Planning workflow v2
items:
  - name: prompts become library files
    check: Every prompt body the built-in workflows use lives in `workflow/library/` as a file, embedded and rendered by `Build`; `Build` returns byte for byte what it returned before; `tractor workflow show <name>` prints each node's rendered payload.
    doc: ephemeral/projects/tractor/living-instructions/planner-v2/SPRINT-01.md
    command: go build ./... && go vet ./... && golangci-lint run ./... && go test ./... && test -d workflow/library && go run ./cmd/tractor workflow show plan >/dev/null
  - name: the v2 plan workflow
    check: "`workflow/plan.yaml` is a graph of interview, plan, and review nodes; the interview asks through `tractor ask`; the plan node writes `brief.md`, `checklist.md` in the loop ledger format, and a size line; the review node asks the one question in decision 37 with `max_visits` 2; the pipeline lints."
    doc: ephemeral/projects/tractor/living-instructions/planner-v2/SPRINT-02.md
    command: go build ./... && go vet ./... && golangci-lint run ./... && go test ./... && go run ./cmd/tractor validate workflow/plan.yaml
  - name: run it once
    check: A `plan` run on a small seed reaches success with a brief, a checklist that `validate-plan` accepts, and a size line, with the reviewer's findings recorded in the run directory; whatever broke on the way is fixed.
    doc: ephemeral/projects/tractor/living-instructions/planner-v2/SPRINT-03.md
    command: go build ./... && go test ./... && test -f ephemeral/projects/tractor/living-instructions/planner-v2/proof/README.md
---

# Planning workflow v2

Three sprints, one agent turn each, MEDIUM. Read `BUILD.md` first. The
goal in one sentence: make Tractor's built-in `plan` workflow a real
interview-and-plan pipeline whose prompts live as editable files instead
of Go strings. Nothing here changes the engine.
