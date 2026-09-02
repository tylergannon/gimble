---
name: consensus
description: >
  Caller-side loop for resolving material findings from independent review of
  a design, implementation, or proof. Use after `/request-adversarial-review`
  when substantial work needs repeated review and adjudication.
---

# Consensus

The `agent` CLI selects the reviewer automatically. Never pass a model or agent
name as a positional argument; `<workdir>` must immediately follow `agent`.

1. Use `/request-adversarial-review` to start the review and retain the spawned
   reviewer's session ID and review-file path. Read the review artifact.
2. Fix each legitimate material finding. Challenge an incorrect or overstated
   finding with concrete evidence in the same reviewer session.
3. After fixes, do not summarize earlier findings, fixes, resolved areas, or
   expected remaining problems. Do not add positive or negative subject-matter
   limits, expected conclusions, or desired verdicts. Resume the same reviewer
   with a minimal prompt.

Bad:
```text
agent <workdir> --session <id> "/adversarial-review Re-check the fixes for
X and Y. Do not broaden into Z. Write to <next-review-file>."
```

Good:
```text
agent <workdir> --session <id> "/adversarial-review Re-review the entire current
work against the same authoritative sources. Inspect the current implementation
and proof. Write the complete review to <next-review-file>. Do not edit any
other files."
```

4. Use a new `ephemeral/reviews/...-round-NN.md` path for every round; never
   overwrite an earlier review artifact.
5. Re-review the entire current target; previous findings are not the review
   scope.
6. Read each artifact and stop when its outcome is `only nitpicks remain` or
   `no findings`.

If the reviewer returns `invalid review request`, discard that session and
restart with `/request-adversarial-review`; the rejected round cannot count as
consensus.

If evidence does not resolve a disputed finding after three exchanges, stop and
ask HITL to adjudicate. Consensus means either the independent reviewer has no
material findings or HITL has resolved the remaining dissent.
