---
chapter: Planning workflow
items:
  - name: embedded plan workflow
    check: The binary contains a named plan workflow whose single long-lived agent interviews through tractor ask, writes the three fixed planning artifacts, and is re-entered when mechanical artifact validation fails.
    doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/SPRINT-01.md
    command: go build ./... && go vet ./... && sh ephemeral/projects/tractor/living-instructions/chapters/02-planning/check-sprint-01.sh
  - name: workflow CLI and handoff
    check: A caller can list built-in workflows and run plan with an explicit project, seed file, workdir, and logs directory; completion prints the artifact paths, size, and next action.
    doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/SPRINT-02.md
    command: go build ./... && go vet ./... && sh ephemeral/projects/tractor/living-instructions/chapters/02-planning/check-sprint-02.sh
  - name: planning docs and skill
    check: The CLI help, README, reference and site docs, skill bundle, and llms.txt teach callers to start with the plan workflow and accurately describe its interview, outputs, sizes, and handoff.
    doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/SPRINT-03.md
    command: go run ./cmd/tractor workflow --help >/dev/null && go run ./cmd/tractor workflow run plan --help >/dev/null && test "$(go run ./cmd/tractor workflow list | grep -c '^plan')" -eq 1 && grep -q 'workflow run plan' README.md && grep -q 'workflow run plan' skills/tractor/SKILL.md && grep -q 'workflow run plan' llms.txt
    infer:
      files:
        - README.md
        - docs/spec.md
        - skills/tractor/SKILL.md
        - src/content/docs/*.md
      prompt: Run the workflow help commands and judge whether a caller reading only these files would start with plan, answer its interview correctly, find all three outputs, and follow the size recommendation without the prose contradicting the CLI.
  - name: live planning proof
    check: A real plan workflow run against a small seed asks the reviewer at least one material question, consumes the answer in the same agent turn, completes, and produces a brief, a loop-format checklist, and a SIMPLE recommendation.
    doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/SPRINT-04.md
    command: go build ./... && go vet ./... && go test ./... && sh ephemeral/projects/tractor/living-instructions/chapters/02-planning/check-sprint-04.sh
---

# Chapter 2 sprints

Four sprints, each sized for one agent turn. The first two establish the
runtime contract, the third teaches it, and the last runs the real workflow
with a reviewer answering its interview.
