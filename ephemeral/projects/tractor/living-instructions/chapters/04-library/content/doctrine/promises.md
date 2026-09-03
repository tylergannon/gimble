# Promises

Plan around promises, not a spec. A promise is what the work commits to
and to whom; the plan exists to keep it, and nothing in the plan may
exist that serves no promise.

A promise has five parts. Write all five before you accept it.

- **Statement.** One sentence, observable, falsifiable. "Given a seed
  that names three features, the planner asks about at least one
  promise the seed did not name." Not "the planner is thorough."
- **Must not imply.** What a reader might infer that you are not
  promising. Write it down so nobody, including you, proves the wrong
  thing.
- **Scope.** Which inputs, paths, screens, or callers the statement
  covers, and which it excludes.
- **Verifier.** Who or what decides it is true, on which evidence, and
  from which vantage. A promise the coder verifies for itself is not
  verified.
- **Evidence.** What is captured, where, and what makes it stale.

A promise that cannot be falsified with evidence you can capture is
narrowed until it can be. Prefer "no example in the seed fails when the
verifier runs it" over "the program works".

In a ledger, the item's `check` is the promise, its `command` and
`infer` are the gates, and `done: true` written by the engine is the
attestation. A declined promise is not deleted: it becomes an
exclusion, written down so the next reader does not reopen it.

When you draft promises from a seed, draft the ones a user of this
thing would expect, including ones the seed never mentions, and ask
about each. Anticipation is the job; the human prunes.

Source: decisions.md 37; sources/diffusioninc/.claude/skills/df-promise/SKILL.md ("Setup: interview before building").
