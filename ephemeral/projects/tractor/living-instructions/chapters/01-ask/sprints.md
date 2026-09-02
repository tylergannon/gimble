---
chapter: Ask and answer
items:
  - name: ask and answer commands
    check: "`tractor ask` moves a question file into the interview directory, blocks until its answer file exists, and prints the answer; `tractor answer` writes that file; the sprint asked the reviewer at least one real question through the new command."
    doc: ephemeral/projects/tractor/living-instructions/chapters/01-ask/SPRINT-01.md
    command: go build ./... && go vet ./... && go test ./cmd/tractor/... && sh ephemeral/projects/tractor/living-instructions/chapters/01-ask/check-sprint-01.sh && test -f ephemeral/projects/tractor/living-instructions/interview/0001.answer.md
    done: true
  - name: run directory reaches agents
    check: Every harness turn and tool node started by a run sees TRACTOR_RUN_DIR; `ask` appends QuestionAsked to that run's timeline; a live pipeline whose agent asks a question completes once the question is answered.
    doc: ephemeral/projects/tractor/living-instructions/chapters/01-ask/SPRINT-02.md
    command: go build ./... && go vet ./... && go test ./engine/... ./harness/... ./cmd/tractor/... && sh ephemeral/projects/tractor/living-instructions/chapters/01-ask/check-sprint-02.sh
    done: true
  - name: docs and skill
    check: The spec, the docs site, the skill bundle, and llms.txt teach ask and answer accurately enough that an agent reading only them uses the commands correctly.
    doc: ephemeral/projects/tractor/living-instructions/chapters/01-ask/SPRINT-03.md
    command: go run ./cmd/tractor ask --help >/dev/null && go run ./cmd/tractor answer --help >/dev/null && grep -q 'tractor ask' docs/spec.md && grep -q 'tractor ask' skills/tractor/SKILL.md && grep -q 'tractor ask' llms.txt
    infer:
      files:
        - docs/spec.md
        - skills/tractor/SKILL.md
        - src/content/docs/*.md
      prompt: Run `go run ./cmd/tractor ask --help` and `go run ./cmd/tractor answer --help`. Judge whether an agent reading only the listed docs would invoke both commands correctly, including where the interview directory comes from and how the answer arrives. Fail on any claim the docs make that the help text contradicts.
---

# Chapter 1 sprints

Three sprints, one agent turn each. The commands are the definition of
done; the check scripts beside this file are part of it. Sprint docs carry
the open decisions the implementer must ask about.
