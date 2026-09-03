---
chapter: The library
items:
  - name: prompts become library files
    check: "`workflow/library/` holds the three graphs and the four prompts as embedded files rendered with `text/template` under non-default delimiters; `Build` returns the same graphs it returned before, byte for byte in every prompt and command; no prompt body remains in a Go string."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-01.md
    command: go build ./... && go vet ./... && go test ./workflow/... ./cmd/tractor/... && mkdir -p ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/prompts-are-library-files.sh > ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/prompts-are-library-files.log 2>&1; rc=$?; cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/prompts-are-library-files.log; test $rc -eq 0
    infer:
      files:
        - workflow/*.go
        - workflow/library/README.md
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/prompts-are-library-files.log
      prompt: Read the log and the non-test Go of every module package the log lists. Does any Go string carry prompt text (instructions to an agent) rather than a data value? Did the log show every node's payload byte-equal to the pre-migration baseline built from the commit in prove/base-commit.txt, or, where doctrine is present, a diff in which every removed line reappears in a doctrine page or skeleton the prompt includes and every added line is an include's rendered text? Fail on prompt text in Go, on a lost line, or on an invented one.
  - name: workflow show
    check: "`tractor workflow show <name>` prints every node the graph declares and, for each, the prompt, command, or checklist exactly as `Build` materialized it for the given parameters, naming the library file each prompt came from; `--node --raw`, `--values`, and `--stage <dir>` (a diff against a stage's prompt.md with the frame stripped) work as the sprint doc says."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-02.md
    command: go build ./... && go vet ./... && go test ./workflow/... ./cmd/tractor/... && mkdir -p ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/show-equals-build.sh > ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/show-equals-build.log 2>&1; rc=$?; cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/show-equals-build.log; test $rc -eq 0
    infer:
      files:
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/show-equals-build.log
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/show-equals-build.sh
      prompt: Read the script and its log. Did the run compare every node of every workflow against the Build dumper, did the --stage probe fail after the random perturbation for that reason, and did --values report only derived values? Fail if any step is missing from the log or passed for a reason other than the one the script names.
  - name: render test and orphan walk
    check: "A test renders every template with representative parameters and names the file on a syntax error; a test walks the embedded tree and fails, naming the page, on a doctrine page no rendered agent-facing library file (prompt body, supervisor brief, or pass) references; both are proven against injected defects in a copy of the tree, and every rendered line is library text plus data values."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-03.md
    command: go build ./... && go vet ./... && go test ./workflow/... && mkdir -p ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/orphan-walk-and-render.sh > ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/orphan-walk-and-render.log 2>&1; rc=$?; cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/orphan-walk-and-render.log; test $rc -eq 0
    infer:
      files:
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/orphan-walk-and-render.log
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/orphan-walk-and-render.sh
        - workflow/library_test.go
      prompt: Read the script, its log, and the tests. For each probe the log names (the random orphan page, the page whose citation was removed, the random broken action, the per-file sentinels), did the test or check fail for the injected reason and name the injected file, and does the test's code walk the tree rather than recognise probe names? Fail if any probe's failure is generic, names the wrong file, or the test special-cases probes.
  - name: doctrine pages and skeletons
    check: "The doctrine pages and artifact skeletons the four existing prompts can cite exist under the library, copied from chapters/04-library/content/, each under about sixty lines with a provenance pointer into `sources/` or `decisions.md`, and the migrated prompts include them so the orphan walk passes; the pages say what decisions 37 to 59 say, in the library's voice."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-04.md
    command: go build ./... && go test ./workflow/... && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/doctrine-pages.sh
    infer:
      files:
        - workflow/library/doctrine/*.md
        - workflow/library/templates/*
      prompt: Judge each doctrine page against decisions 37 to 59 in ephemeral/projects/tractor/living-instructions/decisions.md, and each skeleton against the artifact SPRINT-04.md's table assigns it (fixed parts filled, variable parts as angle-bracket placeholders, nothing blank). Fail if any page contradicts a decision, exceeds about sixty lines, lacks a pointer to its source under ephemeral/projects/tractor/living-instructions/sources/ or in ephemeral/projects/tractor/living-instructions/decisions.md.
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
turn wires them. Each item's `command`, with its `infer` where it has one, is its
definition of done;
the proof scripts beside this file are part of it.
Anchors for every file the migration touches are in
`research/workflow-package-inventory/migration-inventory.md`; read it
before the sprint doc.
