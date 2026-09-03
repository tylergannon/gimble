1. Yes. The cheapest game is a shadow `validation/<item>/design.md` used only by this scenario: let real reviewers reject, revise, and accept it while leaving the declared design artifacts—`story.md`, `evidence.md`, and checklist `command`/`infer`—unchanged. The infer judge reads only the shadow design, notes, prompts, and segments (`design.md:56-65`), so it cannot detect that the operational design was never redesigned.

2. Yes. The byte-difference checks on `design.md` (`design.md:44-53`) depend on an uncaptured claim that this file is the authoritative validation design. The formal package layout instead identifies `story.md`, `evidence.md`, optional sketches, and review notes (`planning-workflow.md:192-210`); neither the command nor infer binds `design.md` to those artifacts or the resulting validator.

3. Yes. A correct implementation may write `story.md` and `evidence.md`, receive rejection notes, revise `evidence.md` and the relevant checklist fields, and later receive a passing review—exactly the workflow specified at `planning-workflow.md:105-122`. With no `design.md`, that implementation fails the validator’s required absence/presence and byte-comparison checks even though P4 is satisfied.

ROUTE: fail
