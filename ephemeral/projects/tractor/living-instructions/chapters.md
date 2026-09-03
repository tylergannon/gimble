---
title: Living instructions
items:
  - name: Ask and answer
    check: An agent inside a run asks a question with `tractor ask`, the run's timeline records it, and `tractor answer` unblocks the agent with the answer.
    doc: ephemeral/projects/tractor/living-instructions/chapters/01-ask/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/01-ask/sprints.md
    command: >-
      go build ./... && go test ./cmd/tractor/... ./engine/... && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/01-ask/sprints.md)" -ge 3
    done: true
  - name: Planning workflow
    check: A built-in planning workflow interviews the caller through `tractor ask` and ends with a brief, a loop-format checklist, and a size recommendation.
    doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/02-planning/sprints.md
    command: >-
      go build ./... && go test ./... && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/02-planning/sprints.md)" -ge 2
    done: true
  - name: Execution workflows
    check: Built-in MEDIUM and LARGE execution workflows run a planning output to completion on the loop node.
    doc: ephemeral/projects/tractor/living-instructions/chapters/03-execution/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/03-execution/sprints.md
    command: >-
      go build ./... && go test ./... && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/03-execution/sprints.md)" -ge 2
    done: true
  - name: The library
    check: Every prompt body, doctrine page, supervisor brief, pass, and skeleton the built-in workflows use is an embedded library file rendered from data values, the two value functions `quote` and `shell`, and the two composition actions `include` and `doctrine`, every doctrine page is rendered by some agent-facing file, and `workflow show` prints every node's payload as `Build` materialized it with `--stage` diffing a stage built to the engine's frame shape from `chapters/04-library/fixtures/frame-preamble.txt` (P8 for the nodes that exist in chapter 4; its chapter 5 leg (the passes, the supervisor briefs, and the prompts, doctrine pages, and skeletons of nodes that first exist in chapter 5, such as `design.md`) is proven in chapter 5 and its real-stage leg in chapter 6); sprint 1 leaves `Build` returning what it returned before the migration byte for byte, and every content edit from sprint 4 on is a reviewed diff of that snapshot committed with the content.
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/04-library/sprints.md
    command: >-
      go build ./... && go vet ./... && golangci-lint run ./... && go test ./... && test -d workflow/library && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/04-library/sprints.md)" -ge 5 && test "$(grep -c '"type":"LoopValidated".*"node":"sprints".*"passed":true' "$TRACTOR_RUN_DIR/timeline.jsonl")" -ge 5
  - name: The planner
    check: The `plan` workflow is the v2 graph with the brief/research loop, the validation design loop, the seven-pass review loop, and four supervisors, all as library content with `models.yaml` roles; P1 to P5 and P7 demonstrated by nested runs, and P8's chapter 5 leg (the passes, the supervisor briefs, and the prompts, doctrine pages, and skeletons of nodes that first exist in chapter 5, such as `design.md`).
    doc: ephemeral/projects/tractor/living-instructions/chapters/05-planner/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/05-planner/sprints.md
    command: >-
      go build ./... && go vet ./... && golangci-lint run ./... && go test ./... && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/05-planner/sprints.md)" -ge 14 && test "$(grep -c '"type":"LoopValidated".*"node":"sprints".*"passed":true' "$TRACTOR_RUN_DIR/timeline.jsonl")" -ge 14
  - name: Execution and the live proof
    check: "`medium` and `large` carry `replan`, `large` carries `verify` with the holdout handoff, two seeds run end to end through plan and execution, and the proof record is written; P6, P9, P10, and P8's real-stage leg (`show --stage` against recorded stages of the seed runs)."
    doc: ephemeral/projects/tractor/living-instructions/chapters/06-execution/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/06-execution/sprints.md
    command: >-
      go build ./... && go vet ./... && golangci-lint run ./... && go test ./... && test -f ephemeral/projects/tractor/living-instructions/proof/planning-v2/README.md && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/06-execution/sprints.md)" -ge 5 && test "$(grep -c '"type":"LoopValidated".*"node":"sprints".*"passed":true' "$TRACTOR_RUN_DIR/timeline.jsonl")" -ge 5
---

# Living instructions: the chapter ledger

Six chapters, in order: three built (1 to 3) and three for the planning
workflow v2 (4 to 6). Each item's `checklist` is the chapter's sprint
ledger and its `doc` is the chapter doc. The engine marks a chapter done
when its sprint loop exits and the command above holds; in runs
started after chapter 6's sprint 2 lands (the seed runs are the first;
the run that builds chapter 6 was materialized without it), the `large`
workflow puts a `verify` turn between the sprint loop's exit and the
mark, and the mark follows only a pass (P6). The commands
count finished sprints so that a chapter whose planner appended nothing
cannot pass vacuously.

Chapter 1 has its sprints written up front. Chapters 2 and 3 start with an
empty ledger; the `plan` node interviews the reviewer and writes them.
