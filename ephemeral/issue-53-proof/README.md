# Issue 53 independent proof

Parent reviewer prepared and operated this proof separately from Sol's implementation.
Final reviewed production revision: d11fb2797958697d77034bc4ed910e17caf46895.
The following 9ab3044 commit only corrects a documented role name.

## Observed behavior

- A real three-harness loop rejected a shipping CLI's incorrect inclusive threshold even though its evidence-capture command exited zero. Fable repaired it and Flash passed the item. The Sol evaluator then returned `not_done` because expedited shipping was missing, appended open work, and dispatched it. Both items subsequently passed, the evaluator returned `done`, and the pipeline completed. An independent oracle executed all six real CLI scenarios successfully. [Native run receipt, program snapshots, decisions, and oracle](https://pub-49d826f028c94744bb6d55c4a63b56ed.r2.dev/proof/2026/09/04/5bbc9e94-47b4-4680-8d41-477aee8872d0-tractor-issue53-live-receipt.json).
- Nine actual model turns logged the correct role, provider, model, and effort. A transparent native observer captured Claude `claude-fable-5-1 --effort high`, agy `gemini-3.8-flash-medium`, and Codex `gpt-5.6-sol` with `effort: medium`. The receipt excludes raw prompts and harness transcripts.
- All 29 CLI preflight cases passed, including 23 invalid cases rejected before native invocation or run-log creation. [Preflight receipt](https://pub-49d826f028c94744bb6d55c4a63b56ed.r2.dev/proof/2026/09/04/7b73eec5-14ed-4f5e-b85b-b2aac2f69e75-tractor-issue53-final-preflight.json).
- CLI inspection demonstrated whole-object replacement for ordinary agents and supervisors, independent judge defaults, inherited evaluator defaults, branch overrides, fan-out templates, and fan-in defaults. [Resolution receipt](https://pub-49d826f028c94744bb6d55c4a63b56ed.r2.dev/proof/2026/09/04/02945839-c22f-4a8d-a091-4bd15a64d439-tractor-issue53-final-resolutions.json).
- In the real browser editor, judge version `3.7` survived save/reload and displayed `gemini-3.7-flash-medium`. A name-only agent override to Flash resolved to medium without inheriting pipeline Fable version `5.1` or high effort. [Loop screenshot](https://pub-49d826f028c94744bb6d55c4a63b56ed.r2.dev/proof/2026/09/04/d55c835f-243d-46e0-a69a-e83419c56c2a-tractor-issue53-editor-loop.png), [replacement screenshot](https://pub-49d826f028c94744bb6d55c4a63b56ed.r2.dev/proof/2026/09/04/113df63a-72b1-4fbe-9f8b-999a7f352238-tractor-issue53-editor-atomic.png), [saved YAML](https://pub-49d826f028c94744bb6d55c4a63b56ed.r2.dev/proof/2026/09/04/e872a8e0-dc5a-4070-8530-999587ae5cdb-tractor-issue53-editor-saved.yaml).

## Review findings and resolution

The initial native run exposed a cross-harness collision: both hidden loop roles used the same fidelity-none binding key. Commit 812ce0c scopes that key by role. The repeated native run passed.

Code review found that preflight normalized missing system defaults differently from per-turn resolution. Commit d11fb27 moved normalization into the shared path; independent focused tests passed for empty configuration and a Flash-only system configuration. [Regression results](https://pub-49d826f028c94744bb6d55c4a63b56ed.r2.dev/proof/2026/09/04/054c97ba-2973-48bf-80b6-b0a27e1edc60-tractor-issue53-review-regressions.txt).

Native proof tested 812ce0ce181f239bb702e46fa6783c7f50b9a713, binary SHA-256 ba70846822ef5f6c1557a31c2c643e3c5599a387ca9bcf9ed391eb8ebc643daf. Final CLI and regression checks tested d11fb27. The intervening production change normalizes missing system defaults; the live fixture supplies explicit model/effort objects for every role, so that change does not alter its selections. Editor proof tested f84f47a; neither later production fix changes editor code, schema, or its explicit selections.

## Required checks

Sol's implementation and fix commits passed the full Go suite and golangci-lint hooks. `go build ./...` passed. `go generate ./graph` preserved schema/checksum/generated-Go hashes. Editor checks reported zero errors/warnings, and its rebuilt bundle is committed. Documentation verification passed. All 14 shipped workflows (9 repository examples and 5 skill copies) validated. The parent independently ran final preflight/resolution scripts and both regression tests.

Review conclusion: no unresolved material findings. This proves the specified scenario and routing/default contracts; static resolution does not establish remote availability for every recognizable native model ID.

## Reproduction

Run `run_live.py /absolute/path/to/tractor` to copy the deliberately broken application into a fresh temporary workspace and execute native harnesses. Run `collect_live.py RUN_ROOT` after completion to produce the bounded receipt. `capture.py` records process outputs without deciding correctness; `snapshot` preserves software/evidence each visit; `verify_quote.py` remains outside the worker workspace.
