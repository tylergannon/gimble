---
chapter: Execution workflows
items:
  - name: execution-ready planning artifacts
    check: The plan workflow emits and mechanically validates a flat multi-sprint checklist for MEDIUM and a chapter ledger with chapter docs and empty per-chapter sprint ledgers for LARGE.
    doc: ephemeral/projects/tractor/living-instructions/chapters/03-execution/SPRINT-01.md
    command: go build ./... && go vet ./... && sh ephemeral/projects/tractor/living-instructions/chapters/03-execution/check-sprint-01.sh
    done: true
  - name: embedded MEDIUM workflow
    check: The named medium workflow materializes the selected project's checklist path without new graph substitution and runs every planned sprint through one validating loop until the engine marks the checklist complete.
    doc: ephemeral/projects/tractor/living-instructions/chapters/03-execution/SPRINT-02.md
    command: go build ./... && go vet ./... && sh ephemeral/projects/tractor/living-instructions/chapters/03-execution/check-sprint-02.sh
    done: true
  - name: embedded LARGE workflow
    check: The named large workflow runs a project chapter ledger through a material planning turn and a nested sprint loop, using each chapter item's own checklist until both ledger levels are engine-marked complete.
    doc: ephemeral/projects/tractor/living-instructions/chapters/03-execution/SPRINT-03.md
    command: go build ./... && go vet ./... && sh ephemeral/projects/tractor/living-instructions/chapters/03-execution/check-sprint-03.sh
    done: true
  - name: execution workflow CLI and run directories
    check: A caller can list and run plan, medium, or large by project; all three allocate and print a fresh default run directory under Tractor's state root while preserving an explicit logs override.
    doc: ephemeral/projects/tractor/living-instructions/chapters/03-execution/SPRINT-04.md
    command: go build ./... && go vet ./... && sh ephemeral/projects/tractor/living-instructions/chapters/03-execution/check-sprint-04.sh
  - name: execution workflow docs and skill
    check: CLI help, README, reference and site docs, the Tractor skill, and llms.txt accurately teach the directly runnable MEDIUM and LARGE handoffs, their loop shapes, reviewer questions, validation ownership, and log discovery.
    doc: ephemeral/projects/tractor/living-instructions/chapters/03-execution/SPRINT-05.md
    command: go build ./... && go vet ./... && sh ephemeral/projects/tractor/living-instructions/chapters/03-execution/check-sprint-05.sh
    infer:
      files:
        - README.md
        - docs/spec.md
        - skills/tractor/SKILL.md
        - src/content/docs/*.md
      prompt: Compare these files with the built-in workflow help and judge whether a caller can choose and run MEDIUM or LARGE, find the allocated logs, answer a material question, and understand that the loop engine alone marks validated items done.
  - name: live plan-to-MEDIUM proof
    check: A real plan workflow creates a MEDIUM plan after a blocking reviewer interview, and the printed medium handoff runs that plan with a real coding harness to completion, leaving every item engine-marked done and per-lap validation evidence in the run logs.
    doc: ephemeral/projects/tractor/living-instructions/chapters/03-execution/SPRINT-06.md
    command: go build ./... && go vet ./... && go test ./... && sh ephemeral/projects/tractor/living-instructions/chapters/03-execution/check-sprint-06.sh
---

# Chapter 3 sprints

Six one-turn sprints. The first makes chapter 2's output executable at both
sizes, the next two establish the embedded loop graphs, the fourth exposes
their complete CLI lifecycle, the fifth teaches the handoff, and the last
proves the real plan-to-MEDIUM path with a live harness.
