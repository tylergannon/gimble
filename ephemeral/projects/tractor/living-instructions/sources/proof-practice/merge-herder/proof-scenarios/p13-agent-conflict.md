# P13: A Real Integration Conflict Is Resolved

- `HARNESS CONTROL -> environment`: create `A` and `B` whose changes conflict
  when applied in order; install fixture tests that state the intended behavior.
- `PRODUCT -> MergeHerder`: submit `A` and `B`; allow batch formation.
- `GitHub -> MergeHerder PR SNAPSHOT`: provide both member snapshots through the
  selected, still-TBD source.
- `MergeHerder -> GIT`: fetch `main`, `A`, and `B` as real objects.
- `MergeHerder -> WORKER`: start one integration session with repository,
  recorded base, ordered members, and authority to create only the release
  chain; receive resolved `R0=M-SA-SB` plus execution evidence.
- `MergeHerder -> GIT`: push `R` to exact resolved SHA `R0`.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `R`, `after=R0`.
- `WORKFLOW -> environment`: execute the real fixture workflow at exact `R0`;
  fixture behavior tests, not a copied expected patch, determine success.
- `GH WEBHOOK -> MergeHerder`: deliver run events for exact `R0`.
- `OBSERVE`: each source contributes one commit, source refs are unchanged, and
  only a workflow-green resolution may land.

Deterministic coordinator proof substitutes a controlled worker at the same
interaction. Agent-quality evaluation repeats the trace with the real coding
agent across multiple unseen conflict fixtures and reports a success rate.
