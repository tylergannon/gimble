# Promises

The promise list for planning workflow v2, extracted from
`declaration.md` §3 so the validation and review loops have one file to
iterate. Statement, must-not-imply, archetype, and verifier are
authoritative there; this file tracks status only. Status moves to
`designed` when `validation/<id>/verdict.md` says pass, to `demonstrated`
when its sprint is engine-marked, and to `proven` when a `verify` verdict
covers it.

| Id | Short name | Chapter | Status |
|---|---|---|---|
| P1 | Elicitation adds a promise; declined becomes an exclusion | 5 | declared |
| P2 | Brief/research loop halts through the tool node | 5 | declared |
| P3 | Every promise has an independent passing validation verdict | 5 | declared |
| P4 | Rejected validation design is re-entered with notes and passes | 5 | declared |
| P5 | Six review passes engine-marked after independent verdicts | 5 | declared |
| P6 | `large` completes and chapters are marked only after `verify` | 6 | declared |
| P7 | `scope_cop` steers and the steered output changes | 5 | declared |
| P8 | Library is content: no Go-string prompts, no orphans, `show` equals materialized | 4 | declared |
| P9 | Docs alone are enough to use plan, show, and ask | 4, 5, 6 | declared |
| P10 | Two known seeds plan and execute end to end | 6 | declared |

Exclusions (promises considered and declined): a holdout in this
project's proof; budgets or complexity tripwires; a secret holdout
location; multi-level supervision; fan-out drafts of chapter docs; any
engine change; the web client.
