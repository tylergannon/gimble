# P9: an agent reading only the docs uses plan, show, and ask correctly

Archetype: scenario, judged by inference. Lap 9; answers `review-8.md`.

## Story

1. The check builds the binary and makes a scratch directory holding
   only: the promised sources, which are the whole docs site
   (`src/content/docs/`, every page), `docs/spec.md`, and the skill
   bundle (`skills/tractor/`); the binary; and an empty git repository.
   `llms.txt` and the library README are not included; P9 names the
   site, the spec, and the skill. `XDG_STATE_HOME` for the reader is
   set inside the scratch directory, so any run it starts without
   `--logs` also lands there.
2. It runs a one-node pipeline, `reader` (codergen, claude, fresh
   context, workdir the scratch directory), with one `observer.sh`
   attached to the reader run. The check watches the scratch directory
   for the nested plan run's directory; once that run's first question
   has been answered (its first `.answer.md` exists), the check attaches
   a second `observer.sh` with accept-all rules to the nested run so it
   can finish. The first question is the reader's alone. The prompt
   describes the tasks in plain words and names no command, flag, or
   path: start the built-in planning workflow on a one-line product and
   let it run to the end; ask your operator one question about it the
   way an agent inside a Tractor run is meant to, and wait for the
   answer; answer the first question the planning run asks; print, with
   the command the docs give for it, the prompt the planner node will
   receive; and say where the interview directory and a universal
   promise's holdout live and which planning step produces each. Run
   the commands, do not describe them, and finish with a short report.
3. Wait for `COMPLETED` of the reader run.

## Evidence

- The reader run directory: `timeline.jsonl` (its own `QuestionAsked`),
  its segment (every command it ran and what came back), `response.md`,
  `observer/answers.log`.
- The plan run directory the reader created (found under the scratch
  directory, whether by `--logs` or the state root): `timeline.jsonl`,
  `interview/`, the planner stage's `prompt.md` and segment (its
  `tractor ask` `tool_call` and paired `tool_result`), its own
  `observer/`.
- The promised sources at that commit, and the `--help` of `workflow
  run`, `workflow show`, `ask`, `answer`.
- The planner prompt as the check itself prints it with `show plan
  --node planner --raw` for the reader's project and seed, and the
  exit status of the check's own `show plan --node planner --stage
  <the planner stage dir>`.
- The holdout root as the software renders it: `show large --node
  verify --raw` for the reader's project, produced by the check.

## Validator

`command`: `prove/p9-docs.sh`: the reader run completed; the reader's
own timeline has a `QuestionAsked` answered through the observer (the
reader used `tractor ask`); exactly one plan run directory exists under
the scratch directory and its `timeline.jsonl` ends with
`PipelineCompleted` (the run the reader started finished; a run left
hanging or failed is a fail); its first `QuestionAsked` E was answered
by the reader: the reader's segment has a `tool_call` invoking `tractor
answer` for E's file at time A, and the plan run's planner segment has
a `tractor ask` `tool_call` before E with its paired `tool_result`
after A (the reader's answer released the planner's blocked ask); the
reader's segment has a `tool_call` invoking `tractor workflow show`
whose paired `tool_result` text contains, trailing whitespace aside,
the whole of the check's own `show plan --node planner --raw` output
for the same project (the headed form and the raw form both satisfy
this; a prefix does not), and the check's `show --stage` against the
plan run's planner stage exits 0 (what `show` prints is what the run
sent, frame aside); `response.md` names the interview directory the
plan run used (the directory of E's path) and the holdout root that
`show large --node verify --raw` renders.

`infer` (files: the reader's `response.md` and segment, the promised
sources, the help output, `planning-workflow.md` section 3): "For the
tasks only: did every command, flag, and path the agent used come from
the promised sources, and did it work as run? Does the agent's account
of where the interview directory and the holdout live and which
planning step produces each agree with the promised sources and with
`planning-workflow.md`? Fail on any command, flag, or path for these
tasks that the help text contradicts, on any statement in the promised
sources about these tasks that the help text contradicts, and on any
disagreement in the account."

## Not proven

Completeness of the docs beyond these tasks. Statements in `llms.txt`
or the library README; they are outside P9. The holdout handoff itself
and which node in a real run writes it (P6's `verify` reads it; chapter
6 sprint 2's own check covers the writer); here that the promised
sources, the agent, and the rendered verify prompt agree on where it
lives and what the docs say produces it.
