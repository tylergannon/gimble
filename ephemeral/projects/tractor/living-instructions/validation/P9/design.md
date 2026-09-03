# P9: an agent reading only the docs uses plan, show, and ask correctly

Archetype: scenario, judged by inference. Lap 4; answers `review-3.md`.

## Story

1. The check builds the binary and makes a scratch directory holding
   only: the docs listed below, the binary, and an empty git
   repository. No other part of the Tractor repository is reachable.
2. It runs a one-node pipeline, `reader` (codergen, claude, fresh
   context, workdir the scratch directory), with `observer.sh` attached
   to the reader run. The prompt describes four tasks in plain words and
   names no command, flag, or path: start the built-in planning
   workflow on a one-line product; ask your operator one question about
   it the way an agent inside a Tractor run is meant to, and wait for
   the answer; answer the first question the planning run asks; print
   the prompt the planner node will receive; and say where the
   interview directory and a universal promise's holdout live. Run the
   commands, do not describe them, and finish with a short report.
3. Wait for `COMPLETED` of the reader run.

## Evidence

- The reader run directory: `timeline.jsonl` (its own `QuestionAsked`),
  its segment (every command it ran and what came back), `response.md`,
  `observer/answers.log`.
- The plan run directory the reader created inside the scratch directory
  (`Logs:` line in its output): `timeline.jsonl`, `interview/`, and
  `stages/<seq>-<planner node>/prompt.md`.
- The docs at that commit: `docs/spec.md` section 3.1.2 and the library
  subsection, `src/content/docs/planning.md`, `interviews.md`,
  `skills/tractor/SKILL.md`, `llms.txt`, `workflow/library/README.md`.
- The holdout root as the software itself renders it: `show large
  --node verify --raw` output for the reader's project, produced by the
  script.

## Validator

`command`: `prove/p9-docs.sh`: the reader run completed; the reader's
own timeline has a `QuestionAsked` answered through the observer (the
reader used `tractor ask` correctly); exactly one plan run directory
exists under the scratch directory; its `timeline.jsonl` has a
`QuestionAsked`, the reader's segment has a `tool_call` invoking
`tractor answer` for that question, and the plan run's timeline has any
later event after that call's `ts` (the answer unblocked the run); the
reader's segment has a `tool_result` whose text is contained in the
plan run's recorded `prompt.md` of the planner node (what the reader
printed is what the run sent, frame aside); `response.md` names the
interview directory the plan run actually used (the directory of its
`QuestionAsked` path) and the holdout root that `show large --node
verify --raw` renders.

`infer` (files: the reader's `response.md` and segment, the docs, the
help output): "For the four tasks only: did every command, flag, and
path the agent used come from the docs, and did it work as run? Fail
on any command, flag, or path for these tasks that the help text
contradicts, on any statement in the docs about these four tasks that
the help text contradicts, and if the docs' statement of where the
interview directory and the holdout live disagrees with what the run
and the verify prompt show."

## Not proven

Completeness of the docs beyond these four tasks. The holdout handoff
itself (P6's `verify` reads it); here that the docs, the agent, and the
rendered verify prompt agree on where it lives.
