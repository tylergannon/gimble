---
chapter: The library
items:
  - name: prompts become library files
    check: "`workflow/library/` holds the three graphs and the four prompts as embedded files rendered with `text/template` under non-default delimiters; `Build` returns the same graphs it returned before, byte for byte in every prompt and command; no prompt body remains in a Go string."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-01.md
    command: test "$(cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/base-commit.txt)" = a5fca707f0abc8fd96c9af4f45ea759caebf864e && go build ./... && go vet ./... && go test ./workflow/... ./cmd/tractor/... && mkdir -p ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/prompts-are-library-files.sh > ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/prompts-are-library-files.log 2>&1; rc=$?; cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/prompts-are-library-files.log; test $rc -eq 0
    infer:
      files:
        - workflow/*.go
        - workflow/library/README.md
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/prompts-are-library-files.log
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/prompts-are-library-files.sh
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/lib.sh
      prompt: Read the script, its log, and the non-test Go of every module package the log lists. Does any Go string carry prompt text (instructions to an agent) rather than a data value? Did the log show every node's payload byte-equal to the pre-migration baseline built from the commit in prove/base-commit.txt (this sprint moves prompts without changing a byte; the diff mode the script has is for a later sprint and must not have been taken here)? Fail on prompt text in Go or on any node that was not byte-equal.
  - name: workflow show
    check: "`tractor workflow show <name>` prints every node `Build` returns and, for each, the prompt, command, or checklist exactly as `Build` materialized it for the given parameters, naming the library file each prompt came from; `--node --raw`, `--values`, and `--stage <dir>` (a diff against a stage's prompt.md with the frame stripped) work as the sprint doc says."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-02.md
    command: go build ./... && go vet ./... && go test ./workflow/... ./cmd/tractor/... && mkdir -p ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/show-equals-build.sh > ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/show-equals-build.log 2>&1; rc=$?; cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/show-equals-build.log; test $rc -eq 0
    infer:
      files:
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/show-equals-build.log
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/show-equals-build.sh
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/lib.sh
      prompt: Read the scripts and the log. Did the run compare every node Build returns against the Build dumper, did the headed body equal --raw for every node, did the --stage probe exit 1 with a diff naming the perturbed line, and did --values report exactly the derived values? Fail if any step is missing from the log or passed for a reason other than the one the script names.
  - name: render test and orphan walk
    check: "A test renders every template under prompts/, supervisors/, and passes/ with the representative parameter sets and names the file on a syntax error or on a template no set renders; a test walks the embedded tree and fails, naming the page, on a doctrine page that no node's rendering of an agent-facing library file (prompt body, supervisor brief, or pass) reaches under the representative parameter sets, directly or through a file it includes; both are proven against injected defects in a copy of the tree, and every prompt equals the standalone rendering of its library file."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-03.md
    command: go build ./... && go vet ./... && go test ./workflow/... && mkdir -p ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/orphan-walk-and-render.sh > ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/orphan-walk-and-render.log 2>&1; rc=$?; cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/orphan-walk-and-render.log; test $rc -eq 0
    infer:
      files:
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/orphan-walk-and-render.log
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/probe-*.log
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/orphan-walk-and-render.sh
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/lib.sh
        - workflow/library_test.go
      prompt: Read the scripts, the logs, and the tests. For each probe the log names (the random orphan page, the page whose citation was removed, the random broken action, the per-file sentinels, each listed skeleton, each doctrine page under a listed parameter set), did the test or check fail or pass for the reason the script names and name the right file, and does the test's code walk the tree and render rather than recognise probe names? Fail if any probe's outcome is generic, names the wrong file, or the test special-cases probes.
  - name: doctrine pages and skeletons
    check: "The doctrine pages and artifact skeletons under chapters/04-library/content/ are installed unchanged under workflow/library/doctrine/ and workflow/library/templates/, the four migrated prompts include them where they paraphrased them, and the orphan walk passes; nothing an agent was told before the sprint is lost, only moved into a page."
    doc: ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-04.md
    command: go build ./... && go test ./workflow/... && mkdir -p ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/prompts-are-library-files.sh > ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/prompts-are-library-files.log 2>&1 && sh ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/doctrine-pages.sh > ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/doctrine-pages.log 2>&1; rc=$?; cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/doctrine-pages.log; test $rc -eq 0
    infer:
      files:
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/doctrine-pages.log
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/doctrine-pages.sh
        - ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run/prompts-are-library-files.log
        - workflow/library/prompts/*/*
        - workflow/library/doctrine/*.md
        - workflow/library/templates/*
      prompt: The log shows each installed page and skeleton compared byte for byte with its copy under chapters/04-library/content/, and the prompts-are-library-files log shows, per node, the diff between the pre-migration prompt and the current one. Judge that diff. Is every removed passage now said by a doctrine page or skeleton the prompt includes, and is every added line an include's rendered text? Fail if anything an agent was told is lost, if a page or skeleton differs from content/, or if a prompt still paraphrases a page it now includes.
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
turn wires them. Each item's `command`, with its `infer` where it has
one, is its definition of done; the proof scripts beside this file are
part of it and are the planner's (BUILD.md: no turn edits them).
Anchors for every file the migration touches are in
`research/workflow-package-inventory/migration-inventory.md`; read it
before the sprint doc.
