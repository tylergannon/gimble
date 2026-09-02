# Sprint 3: docs and skill

Teach `ask` and `answer` everywhere Tractor already teaches itself. Keep
each addition as short as the surrounding text.

- `docs/spec.md`: a subsection under the CLI or run sections describing both
  commands, the interview directory, `TRACTOR_INTERVIEW_DIR`,
  `TRACTOR_RUN_DIR`, the `QuestionAsked` timeline event, and the resume
  rule. Add `QuestionAsked` to the event table.
- `src/content/docs/`: a page for interviews (or a section in the page that
  fits), with the agent-side recipe from
  `ephemeral/projects/tractor/living-instructions/BUILD.md` "Asking
  questions", generalized.
- `skills/tractor/SKILL.md` and `llms.txt`: the two commands and when an
  agent uses them. The calling agent's side too: watch the timeline for
  `QuestionAsked`, open the file, answer with `tractor answer`.
- `README.md`: one line in the command list.

The infer judge for this sprint runs the real `--help` and fails on any
contradiction, so copy flag names from the help text, not from memory.

## Ask the reviewer

- Where the docs-site page should live and what it is called. Look at the
  existing pages first and recommend.
