# Aider-AI/aider

## Purpose
Python coding agent whose prompts are class attributes composed with `str.format`, with a `--show-prompts` flag that prints the exact message list. Looked at only for the show flag; the storage model is the counterexample.

## Pinned
- Repo: https://github.com/Aider-AI/aider
- Commit: `5dc9490bb35f9729ef2c95d00a19ccd30c26339c` (main)
- License: Apache-2.0

## Key concepts
- Layout: one `aider/coders/<edit-format>_prompts.py` per edit format, each a subclass of `CoderPrompts` — https://raw.githubusercontent.com/Aider-AI/aider/5dc9490bb35f9729ef2c95d00a19ccd30c26339c/aider/coders/base_prompts.py L1-30 (`main_system`, `system_reminder`, `lazy_prompt`, `files_content_prefix` …); shared one-liners in `aider/prompts.py`.
- Inheritance is the fragment mechanism: a subclass overrides only the attributes that differ; shared reminders are attributes on the base.
- Rendering: `fmt_system_prompt` fills `{fence}`, `{platform}`, `{shell_cmd_prompt}`, `{final_reminders}`, `{language}` with `prompt.format(...)` — `aider/coders/base_coder.py` L1174-1225; `format_chat_chunks` L1226-1228 assembles system + reminder + files.
- Show: `aider --show-prompts` appends a fake `"Hello!"` user turn, calls `coder.format_messages().all_messages()` and prints via `utils.show_messages` — `aider/main.py` L1044-1050. Same function the live send uses, so the printed messages are what would be sent for that turn.
- No file-based templates, no render-all test, no orphan test.

## Bounded comparison
Like ours in printing the materialized prompt through the live formatting path, but only with prompt bodies as Python class attributes and `str.format` placeholders.

## Gotchas
- `str.format` on prompt text means any literal `{` in a body must be doubled; the same trap exists for `{{` in Go `text/template`.
- `--show-prompts` includes the repo map and file contents, so output is environment-dependent; byte-equality across runs is not a goal there.

## Recipe
- To write a show flag that reuses the message assembler, start at `aider/main.py` L1044-1050 and `base_coder.py` L1226.
