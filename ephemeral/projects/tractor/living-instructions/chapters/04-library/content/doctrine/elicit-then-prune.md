# Elicit, then prune

Interviews have two phases. Do not mix them.

**Elicit.** Before you ask anything, draft the promises a user of this
thing would expect, from the seed, the repository, and the research.
Include the ones the seed does not mention: error handling, data loss,
what happens with no input, what a second user sees. For each, ask "do
you promise this, and what must it not imply", and give your
recommendation. A declined promise becomes an exclusion in the brief.

**Prune.** Once the promise list holds still, a question survives only
if its answer would change a promise, its scope, or its verifier. Ask
nothing whose answer you could derive from the seed, the repository,
or an earlier answer. Never ask for an answer already given.

**Stop.** Stop when a full pass over the promises changes nothing. A
lap that asks nothing and plans nothing is the last one; say so and
move on. Do not keep interviewing to feel thorough.

Two failure modes, named so you can catch yourself:

- Asking before drafting. The human ends up doing the anticipation
  that was your job.
- Asking after the list is stable. Every question then costs a human
  turn and changes nothing.

Write recommendations with every question. A question without a
recommendation is a request for the human to do the thinking.

Source: decisions.md 38; sources/proof-practice/nlspec-methodology/methodology.md (the membership test: ask only what two capable readers would derive differently).
