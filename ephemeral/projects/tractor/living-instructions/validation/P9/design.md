# P9: an agent reading only the docs uses plan, show, and ask correctly

Archetype: scenario, judged by inference. Lap 14; answers `review-13.md`.

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
  the observer's `answers.log`.
- The plan run directory the reader created, found from its `--logs`
  argument, from a `Logs:` line anywhere in the reader's segment, or
  under the scratch state root: `timeline.jsonl`,
  `interview/`, the planner stage's `prompt.md` and segment (its
  `tractor ask` `tool_call` and paired `tool_result`), its own
  observer tree.
- The promised sources at that commit, and the `--help` of `workflow
  run`, `workflow show`, `ask`, `answer`.
- The planner prompt as the plan run recorded it: the check strips the
  frame from the planner stage's `prompt.md` itself, anchored at the
  start of the file: the frame preamble (`fixtures/frame-preamble.txt`,
  byte for byte) if present, then zero or more `<iterate ...>` ...
  `</iterate>` blocks that begin where the previous one ended, then
  blank lines; the remainder is T. Nothing later in the file is
  stripped, so a `</iterate>` planted in a prompt body stays in T, and
  T is never empty (an empty T fails).
- The holdout root as the software renders it: `show large --node
  verify --raw` for the reader's project, produced by the check.

## Validator

`command`: `prove/p9-docs.sh`: the reader run completed; the reader's
own timeline has a `QuestionAsked` answered through the observer, and
the reader's segment has a `tractor ask` `tool_call` whose paired
`tool_result` contains the observer's nonce for that question (the
reader used `tractor ask` and received the answer the observer wrote,
not canned text); the reader started exactly one plan
run, found by the check in this order: a `--logs` argument on the
reader's `workflow run plan` `tool_call`; a `Logs:` line in any
`tool_result` of the reader's segment (the run's stdout, whether read
at once or later from a redirect); or the one run directory under the
scratch state root; that directory exists and its `timeline.jsonl` ends
with `PipelineCompleted` (the run the reader started finished); its first
`QuestionAsked` E was answered by the reader: the reader's segment has
a `tool_call` invoking `tractor answer` for E's file, and the plan
run's planner segment has a `tractor ask` `tool_call` before E whose
paired `tool_result` contains the reader's answer text (the reader's
answer is what released the planner's blocked ask); the reader's
segment has a `tool_call` invoking `tractor workflow show` whose paired
`tool_result` text contains, trailing whitespace aside, the whole of T
(the reader printed, with the promised command, the prompt the run
actually sent; the check compares against the run's own record, not
against another `show` call, and a prefix or canned text fails);
`response.md` names the interview directory the plan run used (the
directory of E's path) and the holdout root that `show large --node
verify --raw` renders.

`infer` (files: the reader's `response.md` and segment, the observer's
`answers.log` for both runs, the planner stage's `prompt.md` and T, the
promised sources, the help output, `planning-workflow.md` section 3): "Is T the planner prompt the run sent, frame aside, and did the
reader print it whole? For the three promised commands (`workflow run plan`, `workflow show`, `ask`,
with `answer` as `ask`'s counterpart): did the agent's use of each come
from the promised sources, and did each work as run? Helper commands
the agent used to write a seed or a question file are not judged. Do the promised sources themselves state where the interview
directory and the holdout live and which planning step produces each
(quote the passages; if the sources are silent, fail: the agent may
have discovered the paths at run time, and P9 promises that the docs
teach them), and does the agent's account agree with those passages
and with `planning-workflow.md`? Fail on any use of the
promised commands that the help text contradicts, on any statement in
the promised sources about them that the help text contradicts, and on
any disagreement in the account."

## Not proven

Completeness of the docs beyond these tasks. Statements in `llms.txt`
or the library README; they are outside P9. The holdout handoff itself
and which node in a real run writes it (P6's `verify` reads it; chapter
6 sprint 2's own check covers the writer); here that the promised
sources, the agent, and the rendered verify prompt agree on where it
lives and what the docs say produces it. That `show` equals `Build`;
that is P8, and this check compares the reader's output with the run's
recorded prompt instead.
