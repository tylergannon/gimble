---
chapter: The library
items:
  - name: prompts become library files
    check: "`workflow/library/` holds the three graphs and the four prompts as embedded files rendered with `text/template` under non-default delimiters; `Build` returns the same graphs it returned before, byte for byte in every prompt and command; no prompt body remains in a Go string."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-01.md
    command: go build ./... && go vet ./... && go test ./workflow/... ./cmd/tractor/... && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/check-sprint-01.sh
  - name: show, render test, orphan walk
    check: "`tractor workflow show <name>` prints each node's prompt and command as `Build` materialized them for the given parameters; `--stage <dir>` diffs against a stage's prompt.md with the frame stripped; a test renders every template with representative parameters and a test walks the embedded tree and fails on a doctrine or template file no prompt references, proven against a synthetic orphan."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-02.md
    command: go build ./... && go vet ./... && go test ./workflow/... ./cmd/tractor/... && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/check-sprint-02.sh
  - name: doctrine pages and skeletons
    check: "The doctrine pages and artifact skeletons the three existing prompts can cite exist under the library, each a page or less with a provenance pointer into `sources/`, and the migrated prompts include them so the orphan walk passes; the pages say what decisions 37 to 59 say, in the library's voice."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-03.md
    command: go build ./... && go test ./workflow/... && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/check-sprint-03.sh
    infer:
      files:
        - workflow/library/doctrine/*.md
        - workflow/library/templates/*
      prompt: Judge each doctrine page against decisions 37 to 59 in ephemeral/projects/tractor/living-instructions/decisions.md. Fail if any page contradicts a decision, exceeds about sixty lines, lacks a pointer to its source under ephemeral/projects/tractor/living-instructions/sources/, or contains the literal text "{{".
  - name: docs and skill
    check: "The spec, the docs site, the skill bundle, and llms.txt teach the library layout, the template contract, and `workflow show` accurately enough that an agent reading only them can add a doctrine page and see it in `show` output."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-04.md
    command: go run ./cmd/tractor workflow show --help >/dev/null && grep -q 'workflow show' docs/spec.md && grep -q 'workflow/library' docs/spec.md && grep -q 'workflow show' skills/tractor/SKILL.md && grep -q 'workflow show' llms.txt && grep -q 'workflow/library' src/content/docs/planning.md
    infer:
      files:
        - docs/spec.md
        - src/content/docs/planning.md
        - skills/tractor/SKILL.md
        - llms.txt
        - workflow/library/README.md
      prompt: Run `go run ./cmd/tractor workflow show --help`. Judge whether an agent reading only the listed docs would add a doctrine page under workflow/library, cite it from a prompt, and confirm it with `workflow show`. Fail on any claim the docs make that the help text or the library README contradicts.
---

# Chapter 4 sprints

Four sprints, one agent turn each except sprint 3, which Claude writes by
hand (decision, interview 0013 round: doctrine authorship). Commands are
the definition of done; check scripts beside this file are part of it.
Anchors for every file the migration touches are in
`research/workflow-package-inventory/migration-inventory.md`; read it
before the sprint doc.
