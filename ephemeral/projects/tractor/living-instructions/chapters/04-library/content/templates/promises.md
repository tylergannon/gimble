# Promises

One row per promise. Statement, must-not-imply, archetype, and verifier
are authoritative here; validation designs under `validation/<id>/`
elaborate them and never contradict them.

| Id | Statement | Must not imply | Archetype | Verifier |
|---|---|---|---|---|
| P1 | <observable, falsifiable sentence> | <what a reader might wrongly infer> | <universal or scenario> | <who decides, on what evidence> |

Status, maintained by the planner: `declared` when written here,
`designed` when its validation item is marked, `demonstrated` when its
sprint is marked, `proven` when a verify turn covering it routed pass.

| Id | Short name | Chapter | Item | Status |
|---|---|---|---|---|
| P1 | <short name> | <chapter number, or - for MEDIUM> | <`name` of the sprint ledger item that demonstrates it; several, comma-separated, when split> | declared |

The Item column is the mapping execution uses: a `verify` turn takes the
promises whose Chapter is its chapter and runs their designs under
`validation/<id>/`; a sprint item's `check` is the promise text or the
leg of it the Item column assigns.

Exclusions (promises considered and declined): <capability>, <capability>.
