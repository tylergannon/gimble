# P9: an agent reading only the docs uses plan, show, and ask correctly

Archetype: scenario, judged by inference. Lap 2; answers `review-1.md`.

## Story

1. The check builds the binary and makes a scratch directory holding
   only: the docs listed below, the binary, and an empty git
   repository. No other part of the Tractor repository is reachable.
2. It runs a one-node pipeline, `reader` (codergen, claude, fresh
   context, workdir the scratch directory), whose prompt is the four
   tasks: plan a one-line product with `tractor workflow run plan`;
   answer the first question the run asks through `tractor answer`;
   print the planner's prompt with `tractor workflow show`; and state
   where the holdout for a universal promise is stored and where the
   interview directory is. The agent must run the commands, not
   describe them, and finish with a short report.
3. Wait for `COMPLETED` of the reader run.

## Evidence

- The reader run directory: its segment (every command it ran and what
  came back) and `response.md`.
- The plan run directory the reader created inside the scratch directory
  (`Logs:` line in its output): `timeline.jsonl`, `interview/`.
- The docs at that commit: `docs/spec.md` section 3.1.2 and the library
  subsection, `src/content/docs/planning.md`, `interviews.md`,
  `skills/tractor/SKILL.md`, `llms.txt`, `workflow/library/README.md`.
- `--help` of `workflow run`, `workflow show`, `ask`, `answer`.

## Validator

`command`: `prove/p9-docs.sh`: the reader run completed; exactly one
plan run directory exists under the scratch directory; its
`timeline.jsonl` has a `QuestionAsked` and a later `StageCompleted`
(the run went on after the answer); the answer file for that question
exists and the reader's segment has a `tool_call` invoking `tractor
answer` on it; the reader's segment has a `tool_call` invoking `tractor
workflow show plan` whose `tool_result` contains a `== ` node header;
`response.md` names the interview directory under the project directory
and the holdout root (`$XDG_STATE_HOME/tractor/holdouts` or its
expanded default).

`infer` (files: the reader's `response.md` and segment, the docs, the
help output): "For the four tasks only: did every command, flag, and
path the agent used come from the docs, and did it work as run? Fail on
any command, flag, or path for these tasks that the help text
contradicts, and on any statement in the docs about these four tasks
that the help text contradicts."

## Not proven

Completeness of the docs beyond these four tasks; the holdout path is
documented as workflow state, not as CLI help, and the judge does not
expect it in `--help`.
