1. `verifier` and `validator` name the same proof mechanism. `declaration.md` says: “Each promise: the statement, what it must not imply, the archetype … and the verifier.” Its P1 verifier is a “Nested run with a scripted answerer …; inspector checks `brief.md`.” `validation/ledger.md` instead says each design contains “the validator (the ledger `command` and `infer` … or the proof script by name),” and every `validation/P*/design.md` uses `## Validator`. Change `validation/ledger.md` and the validation-design headings to `verifier`, or explicitly define a genuine distinction. Owning node: `design`.

2. The observer-token rule directly contradicts P2. `validation/ledger.md` requires the observer to work the token into “a made-up word inside an ordinary sentence” with “no fixed marker such as a label or a trailing line.” `validation/P2/design.md` requires: “Every answer ends with an `observer:` token line.” Change `validation/P2/design.md`. Owning node: `design`.

3. Unused agent-facing templates are simultaneously forbidden and permitted. `chapters/04-library/SPRINT-03.md` requires that “every template under `prompts/`, `supervisors/`, and `passes/` was executed at least once (a file none of the representative sets renders fails the test by name).” `validation/P8/design.md` says: “an unused prompt or supervisor file may exist and is not a defect.” Change `validation/P8/design.md`. Owning node: `design`.

4. Sprint 1’s byte-equality rule has a weaker duplicate gate. `chapters/04-library/SPRINT-01.md` requires that “every prompt and command `Build` returns is byte-identical to what it returned before this sprint.” Its ledger infer in `chapters/04-library/sprints.md` accepts either byte equality “or, where doctrine is present, a diff in which every removed line reappears in a doctrine page or skeleton.” Mere doctrine presence can therefore satisfy the infer while the sprint’s stated check is false. Change `chapters/04-library/sprints.md` so the sprint-1 gate is strictly byte-equal; later doctrine migrations already have their own sprint. Owning node: `decompose`.

ROUTE: fail


