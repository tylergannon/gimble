# Sprint 6: live plan-to-MEDIUM proof

Prove the public handoff across both built-in workflows with a real supported
coding harness. This is the chapter's acceptance run, not another mocked engine
test.

Add a small ambiguous MEDIUM seed beside this document and a repeatable
`check-sprint-06.sh` that:

1. Builds the current Tractor binary into a temporary directory and creates a
   minimal temporary git workdir. Give it two independently testable, initially
   failing Go package behaviors so a flat two-or-more-sprint plan is warranted.
2. Runs the real `plan` workflow against that seed. Watch its printed/default
   run path and project interview directory, wait for `QuestionAsked`, inspect
   each actual question, and answer through the built binary. The answer fixes
   the observable contracts and requires separate one-turn items with a MEDIUM
   recommendation. Bound time and question count without shortening the real
   interview protocol.
3. Require plan `PipelineCompleted`, non-empty questions and answers,
   mechanically validate the artifacts, and inspect—not merely grep—that
   `checklist.md` is a flat open MEDIUM ledger with item-specific commands and
   no `done`.
4. Invoke the exact printed `Next:` command, adding only explicit `--workdir`
   when the temporary repository is not the shell cwd. Let its default fresh
   log allocation stand and discover the path from stdout. Use the real harness
   for every sprint turn and answer any material execution question through the
   same project interview directory.
5. Require execution `PipelineCompleted`, the intended package behaviors and
   full Go tests to pass, every checklist item to contain engine-written
   `done: true`, at least two item selections, and a passing `validation.json`
   plus non-empty validation log artifact for every completed item. Git history
   must show the workflow's committed implementation rather than a harness
   reporting success without changing the repository.

Use `mktemp -d`, clean up on exit, and never depend on an installed Tractor
binary. On failure print the two run tails, timeline tails, interview, planning
artifacts, final checklist, git log/status, and failed validation evidence.
Run every BUILD.md gate before committing.
