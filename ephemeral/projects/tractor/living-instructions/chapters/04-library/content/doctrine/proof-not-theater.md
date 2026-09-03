# Proof, not theater

Checks never prove a promise. Build, vet, lint, and tests are required,
and they prove nothing about a promise; they prove the code is the
kind of code that can be run. A coder who writes the test writes the
verdict.

A promise is proven when a judge that is not the coder looks at
recorded evidence and decides. Recorded means captured by the run or
by an observer: a timeline, a transcript, a screenshot, a program's
output under a command the judge chose. Not the coder's summary of
what happened.

At chapter exit the verifier is an agent on another provider, fresh
context, with tools. It reads the validation design, operates the
software itself, captures its own evidence, and routes pass or fail. A
chapter is marked only through that route. Sprints are demonstrated by
their own `command` and `infer`; chapters are proven.

When you design a proof, ask three questions of it, as the reviewer
will:

- Could a coder satisfy this while the promise is false? Find the
  cheapest way. Close it.
- Is any check trivially true, or resting on evidence nobody captures?
- Is it stricter than the promise, so that a correct build fails it?

Evidence rules that hold everywhere:

- Make the run at check time. Nothing that existed before the check is
  evidence.
- A skipped check, a stale check, a check on a different commit: not
  evidence.
- Name a proof script for what it proves, never for the sprint number.
- Say what is not proven. A proof that claims everything proves less.

Source: decisions.md 41; sources/proof-practice/core-tools/proof-work/SKILL.md; research/review-and-verification (R2).
