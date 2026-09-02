# Sprint 4: live planning proof

Prove the public workflow with a real agent and a real blocking interview. This
is not another unit-test sprint.

Add a small ambiguous seed beside this document for a one-sprint Tractor change:
enough information to establish intent, but with one material contract choice
the planner must ask about. Add `check-sprint-04.sh` beside it, following chapter
1's live proof posture:

1. Build the current Tractor binary into a temporary directory and create a
   minimal temporary git workdir containing the seed. Do not depend on an older
   installed Tractor.
2. Run `tractor workflow run plan` against that seed with an explicit project
   and logs directory, using a real supported harness and a bounded timeout.
3. Watch that run's timeline and project interview directory. When
   `QuestionAsked` appears, read the actual question and answer it through the
   built binary's `tractor answer`; the answer must settle the seed's material
   choice. Handle further material questions until the run completes, but fail
   if it asks indefinitely.
4. Require `PipelineCompleted`, at least one question and non-empty answer, and
   CLI output that names all three artifact paths plus `Size: SIMPLE` and
   `Next: Execute the plan yourself.`
5. Inspect the artifacts, not merely their existence: the brief incorporates
   the reviewer's chosen contract, the checklist parses through Tractor's real
   validator, contains exactly one open item with no `done`, and its check and
   command are specific to the seed.

Use `mktemp -d` and clean it on exit so validation is repeatable and never
overwrites a developer's project. On failure, print the run tail, question and
answer, and artifact contents needed to diagnose it. A green aggregate test
suite without this live run does not satisfy the sprint. Run all required
BUILD.md gates before committing.
