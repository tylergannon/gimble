# P9: an agent reading only the docs uses plan, show, and ask correctly

Archetype: scenario, judged by inference.

## Story

A fresh agent (claude, fresh context, no repository access beyond the
listed docs and the built binary's `--help`) is asked to: plan a
one-line product with `tractor workflow run plan`, answer one question
through `tractor answer`, inspect the planner's prompt with `workflow
show`, and say where the holdout for a universal promise would be
stored. It writes the commands it would run and why.

## Evidence

- The agent's transcript.
- The docs at that commit: `docs/spec.md` §3.1.2 and the library
  subsection, `src/content/docs/planning.md`, `interviews.md`,
  `skills/tractor/SKILL.md`, `llms.txt`, `workflow/library/README.md`.
- `--help` output of `workflow run`, `workflow show`, `ask`, `answer`.

## Validator

`command`: the chapter 4 and chapter 6 docs sprints' grep checks
(commands, flags, and paths named in the docs exist in `--help`).

`infer` (files: the docs, the help output, the transcript): "Would the
commands the agent wrote work as written? Fail on any command, flag, or
path the help text contradicts, and on any claim in the docs the help
text contradicts."

## Not proven

Completeness of the docs beyond these four tasks.
