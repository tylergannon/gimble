# Sprint 5: teach the execution handoff

Make the shipped instructions agree with the now-runnable workflow library and
CLI help.

- README: the shortest plan-to-execution path, including the printed `Next:`
  command and where the default logs path appears.
- `docs/spec.md`: parameters, default/override log allocation, MEDIUM's single
  checklist loop, LARGE's chapter planner plus nested sprint loop, blocking
  reviewer questions, and engine-owned validation and marking.
- `src/content/docs/`: update the planning and loops journey, or add one focused
  execution page if that is clearer. A caller should know when to use each
  size, how to watch it, and why one run covers the whole execution.
- `skills/tractor/SKILL.md`: replace the pending-handoff warning with the actual
  direct commands and monitoring/answering guidance. Keep copy-and-run custom
  graphs available for goals that do not begin with a planning artifact.
- `llms.txt` and `llms-full.txt` when generated from the same source: include
  the three built-in names and executable handoff.
- Cobra help: flags, examples, output, and size-specific requirements agree
  with the prose.

Do not claim child runs, per-item retry budgets, automatic escalation, or
parallel sprints. Do not bury that the engine runs the item validators and is
the only writer of `done`.

Add `check-sprint-05.sh` beside this document. It builds the CLI, captures
`workflow list` and relevant help, verifies all named documentation surfaces
contain the public commands and no pending-workflow warning, and runs any docs
generation or link/build check already used by the repository. The ledger's
infer judge provides the semantic cross-check. Run every BUILD.md gate before
committing.
