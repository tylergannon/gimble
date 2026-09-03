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
    check: Every prompt, doctrine page, and skeleton the built-in workflows use is an embedded library file with a render test, an orphan walk, and a `workflow show` command; `Build` returns what it returned before.
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/04-library/sprints.md
    command: >-
      go build ./... && go test ./... && test -d workflow/library && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/04-library/sprints.md)" -ge 4
  - name: The planner
    check: The `plan` workflow is the v2 graph with the brief/research loop, the validation design loop, the seven-pass review loop, and four supervisors, all as library content with `models.yaml` roles; P1 to P5 and P7 demonstrated by nested runs.
    doc: ephemeral/projects/tractor/living-instructions/chapters/05-planner/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/05-planner/sprints.md
    command: >-
      go build ./... && go test ./... && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/05-planner/sprints.md)" -ge 7
  - name: Execution and the live proof
    check: "`medium` and `large` carry `replan`, `large` carries `verify` with the holdout handoff, two seeds run end to end through plan and execution, and the proof record is written; P6, P9, P10."
    doc: ephemeral/projects/tractor/living-instructions/chapters/06-execution/CHAPTER.md
    checklist: ephemeral/projects/tractor/living-instructions/chapters/06-execution/sprints.md
    command: >-
      go build ./... && go test ./... && test -f ephemeral/projects/tractor/living-instructions/proof/planning-v2/README.md && test "$(grep -c 'done: true' ephemeral/projects/tractor/living-instructions/chapters/06-execution/sprints.md)" -ge 4
---

# Living instructions: the chapter ledger

Three chapters, in order. Each item's `checklist` is the chapter's sprint
ledger and its `doc` is the chapter doc. The engine marks a chapter done
when its sprint loop exits and the command above holds. The commands count
finished sprints so that a chapter whose planner appended nothing cannot
pass vacuously.

Chapter 1 has its sprints written up front. Chapters 2 and 3 start with an
empty ledger; the `plan` node interviews the reviewer and writes them.
