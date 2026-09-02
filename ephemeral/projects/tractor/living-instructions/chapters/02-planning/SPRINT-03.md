# Sprint 3: teach callers to plan first

Make the planning workflow the documented entry point wherever Tractor teaches
itself. Keep additions concise and copy flags and output names from the built
help rather than this plan.

- `README.md`: add `workflow list` and the minimal plan invocation near the
  first-run path.
- `docs/spec.md`: specify named embedded workflows, invocation and parameter
  rules, interview behavior, the three artifacts, size boundaries, and the
  recommendation handoff.
- `src/content/docs/`: add or update the best onboarding page so a person can
  start from a seed file, answer numbered questions, and find the result. Link
  to the interviews and loops pages rather than duplicating them.
- `skills/tractor/SKILL.md`: when a user needs a plan or does not yet know which
  Tractor shape fits, start with `tractor workflow run plan`; explain how the
  caller watches and answers `QuestionAsked` and then follows SIMPLE, MEDIUM, or
  LARGE.
- `llms.txt` (and `llms-full.txt` if generated from the same source): include
  the agent-facing start command and artifact names.
- Cobra help: `tractor workflow --help`, `workflow list --help`, and
  `workflow run plan --help` must agree with the docs.

Do not describe chapter 3's `medium` and `large` workflows as currently
runnable merely because their names are fixed. The docs infer check compares
the real help with the prose. Run all required BUILD.md gates before committing.
