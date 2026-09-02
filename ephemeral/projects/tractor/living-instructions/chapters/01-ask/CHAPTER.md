# Chapter 1: Ask and answer

Status: active

An agent running inside a Tractor node can stop and ask a question. The
question is a file; the answer is a file beside it; the run's timeline says
a question is waiting. Nothing else. This is the transport that the planning
workflow (chapter 2) and every interview after it will use.

## Pyramid index

- L0: A blocking CLI command lets an agent ask a file-shaped question and
  wait for a file-shaped answer, with the run's timeline as the doorbell.
- L1:
  - `tractor ask <path>` moves the question into the interview directory,
    numbered, records `QuestionAsked` in the run timeline, blocks until the
    answer file exists, prints the answer.
  - `tractor answer <question> [text]` writes the answer file.
  - The engine tells agents where the run is, so `ask` can find the
    timeline without being told.
  - Docs and the skill bundle teach both commands.
- L2: sprints 1 to 3 in `sprints.md`.

## Vector

Decisions 3, 4, 22, and 23 in `decisions.md`. The interview is not routing:
the agent keeps its context and blocks inside one node visit. Transport is
files under `ephemeral/projects/<build>/interview/`. The agent-facing
surface is the CLI. The calling agent, not Tractor, is responsible for
noticing the timeline event, opening the question, and answering.

## Review posture

Work advances this chapter when a real agent in a real run asks and gets
answered, end to end, with no human step other than writing the answer.
Unit tests are necessary and not sufficient; every sprint here has a shell
check that exercises the built binary.

## Non-goals

A web page, server, socket, or notification channel. Addressing questions
to anyone but the reviewer. Rich rendering beyond "the file may be HTML".
Multiple concurrent interviews in one run (one directory, one counter).
