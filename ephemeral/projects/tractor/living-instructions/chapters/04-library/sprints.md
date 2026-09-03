---
chapter: The library
items:
  - name: prompts become library files
    check: "`workflow/library/` holds the three graphs and the four prompts as embedded files rendered with `text/template` under non-default delimiters; `Build` returns the same graphs it returned before, byte for byte in every prompt and command; no prompt body remains in a Go string."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-01.md
    command: go build ./... && go vet ./... && go test ./workflow/... ./cmd/tractor/... && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/prompts-are-library-files.sh
  - name: workflow show
    check: "`tractor workflow show <name>` prints every node the graph declares and, for each, the prompt, command, or checklist exactly as `Build` materialized it for the given parameters, naming the library file each prompt came from; `--node --raw`, `--values`, and `--stage <dir>` (a diff against a stage's prompt.md with the frame stripped) work as the sprint doc says."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-02.md
    command: go build ./... && go vet ./... && go test ./workflow/... ./cmd/tractor/... && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/show-equals-build.sh
  - name: render test and orphan walk
    check: "A test renders every template with representative parameters and names the file on a syntax error; a test walks the embedded tree and fails, naming the page, on a doctrine page no rendered prompt references; both are proven against injected defects in a copy of the tree, and every rendered line is library text plus data values."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-03.md
    command: go build ./... && go vet ./... && go test ./workflow/... && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/orphan-walk-and-render.sh
  - name: doctrine pages and skeletons
    check: "The doctrine pages and artifact skeletons the four existing prompts can cite exist under the library, copied from chapters/04-library/content/, each a page or less with a provenance pointer into `sources/` or `decisions.md`, and the migrated prompts include them so the orphan walk passes; the pages say what decisions 37 to 59 say, in the library's voice."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-04.md
    command: go build ./... && go test ./workflow/... && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/doctrine-pages.sh
    infer:
      files:
        - workflow/library/doctrine/*.md
        - workflow/library/templates/*
      prompt: Judge each doctrine page against decisions 37 to 59 in ephemeral/projects/tractor/living-instructions/decisions.md. Fail if any page contradicts a decision, exceeds about sixty lines, lacks a pointer to its source under ephemeral/projects/tractor/living-instructions/sources/ or in ephemeral/projects/tractor/living-instructions/decisions.md.
  - name: docs and skill
    check: "The spec, the docs site, the skill bundle, and llms.txt teach the library layout, point at the library README for the template contract, and teach `workflow show` accurately enough that an agent reading only them can add a doctrine page and see it in `show` output."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-05.md
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

Five sprints, one agent turn each; sprint 4's pages are written by
Claude beforehand under `content/` (interview 0013, question 5) and the
turn wires them. Commands are
the definition of done; check scripts beside this file are part of it.
Anchors for every file the migration touches are in
`research/workflow-package-inventory/migration-inventory.md`; read it
before the sprint doc.
