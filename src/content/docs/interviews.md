---
title: Interviews
description: Let an agent pause inside one run, ask its caller a file-shaped question, and continue with the answer in the same turn.
eyebrow: Operator guide
order: 4
sourceLabel: Read the normative interview contract
sourceUrl: https://github.com/tylergannon/gimble/blob/main/docs/spec.md#311-blocking-interviews
---

An agent does not need a special graph node to ask for a decision. It writes
one Markdown or HTML question, runs `gimble ask`, and stays inside the same
node visit until its caller answers. The question, answer, and timeline event
are ordinary files.

## Ask from inside a run

Give the agent an interview directory in its operating instructions, then have
it write one question per file and run:

```sh
GIMBLE_INTERVIEW_DIR=ephemeral/projects/my-build/interview \
  gimble ask question.md
```

Alternatively, pass the directory explicitly with `--into <directory>`.
`gimble ask` moves the question to the next numbered `.md` or `.html` file,
records `QuestionAsked` in the run's `timeline.jsonl`, and blocks. Gimble sets
`GIMBLE_RUN_DIR` for agents and tool commands inside the run, so `ask` can
find that timeline without a run-path argument.

When a non-empty answer file appears beside the question, `ask` prints its
contents and the agent continues with its existing context. This is an
interview within one turn, not a route through the graph.

## Answer from the calling side

Watch the run's `timeline.jsonl`. When a `QuestionAsked` event appears, open
the file in its `question` field, decide, and pass that numbered path back:

```sh
gimble answer ephemeral/projects/my-build/interview/0001.md \
  "Use the simpler option."
```

`gimble answer` writes `0001.answer.md` beside the question. Omit the text to
read the answer from stdin. It refuses to overwrite an answer that already
exists. The blocked `ask` command notices the non-empty file and prints it.

## Resume an interrupted wait

If the shell running `ask` ends or times out, run the same command again with
the numbered path it printed and the same interview directory:

```sh
GIMBLE_INTERVIEW_DIR=ephemeral/projects/my-build/interview \
  gimble ask ephemeral/projects/my-build/interview/0001.md
```

This resumes waiting for `0001.answer.md`; it does not create another
question or another timeline event. Give interview nodes a timeout long enough
for the caller to respond.
