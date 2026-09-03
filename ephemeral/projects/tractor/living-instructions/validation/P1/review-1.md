1. Yes. The planner can write a synthetic `interview/0001.md` directly, let `answerer.sh` log and answer it, copy its first line into `brief.md`, and complete without invoking or consuming `tractor ask`. The validator checks only the file/log/brief relationship (`design.md:27-34`), while a genuine ask requires a `QuestionAsked` event and blocks for the answer (`docs/spec.md:517-541`).

2. Yes. Although `timeline.jsonl` and `.answer.md` are listed as evidence (`design.md:18-23`), no stated check binds the declined ID to a `QuestionAsked` event, verifies the negative answer, or proves the planner read that answer before creating the exclusion. `answers.log` only records that the substring-based answerer noticed a file (`validation/ledger.md:49-54`).

3. Yes. Requiring the exclusion to contain the question’s first line verbatim (`design.md:27-30`) is stricter than the semantic promise. A correct planner could ask “Should data survive restarts?” and record the declined promise as “No persistence guarantee”; that satisfies P1 but fails exact containment. Likewise, an equivalent question lacking the literal “do you promise” phrasing could evade the scripted decline rule.

ROUTE: fail
