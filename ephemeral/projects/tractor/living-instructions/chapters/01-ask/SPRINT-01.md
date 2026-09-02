# Sprint 1: `tractor ask` and `tractor answer`

Add two subcommands to `cmd/tractor`. Both are plain cobra commands beside
`run` and `validate`. No engine changes in this sprint.

## `tractor ask <path>`

1. Resolve the interview directory: `--into <dir>` if given, else the
   `TRACTOR_INTERVIEW_DIR` environment variable. Neither: fail with a
   message that names both. Create the directory if needed.
2. If `<path>` is already inside the interview directory and matches the
   question naming below, skip to step 5 (resume waiting).
3. Pick the next number: one more than the highest numbered question
   already in the directory, starting at 1. Four digits, zero padded.
   Keep the source file's extension (`.md` or `.html`; anything else is an
   error). Move the file to `<dir>/<nnnn>.<ext>`.
4. If `TRACTOR_RUN_DIR` is set, append one line to
   `$TRACTOR_RUN_DIR/timeline.jsonl` in the engine's shape (see
   `engine/store.go` `appendTimeline`): `type` `QuestionAsked`, `question`
   set to the moved path (relative to the current directory when it is
   under it, absolute otherwise), plus `ts`. Open with `O_APPEND`; do not
   hold the file. If the variable is unset, print a one-line warning to
   stderr and continue.
5. Block until `<dir>/<nnnn>.answer.md` exists and is non-empty. Poll;
   pick an interval the reviewer would find reasonable and say so in
   `--help`. Then print the answer file's contents to stdout and exit 0.
6. Also print the question's moved path to stderr as soon as it is moved,
   so the agent knows the id even if the wait is cut short.

## `tractor answer <question-path> [text]`

Writes `<nnnn>.answer.md` beside the question. Text from the argument, or
from stdin when the argument is absent. Refuse to overwrite an existing
answer. Exit 0 and print the answer path.

## Tests

Unit tests in `cmd/tractor` for numbering, extension handling, resume
without renumber, the timeline line, and the refusal to overwrite. The
shell check `check-sprint-01.sh` beside this doc is the end-to-end
definition of done; read it before you start.

## Ask the reviewer

These are open on purpose. Ask through your new command once it builds,
one question per file, with your recommendation:

- Whether `answer` with neither text nor piped stdin should open `$EDITOR`
  or just fail.
- Whether `ask` should accept a `--timeout` and what it prints on timeout.

Do not finish the sprint with zero questions asked; the sprint's command
requires `interview/0001.answer.md` to exist.
