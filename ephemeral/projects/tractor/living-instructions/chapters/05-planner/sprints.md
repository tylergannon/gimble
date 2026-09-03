---
chapter: The planner
items: []
---

# Chapter 5 sprints

Empty ledger; the plan node writes the backlog when the chapter is
entered and the `replan` step (or, until it exists, the plan node on
re-entry) edits open items after every lap (decision 44).

## Backlog sketch, in order

1. `models.yaml` and per-role resolution in `Build`, explicit providers,
   the `not` constraint, a lint-time error when it cannot be satisfied;
   `workflow show` prints provider and model per node; the implement
   prompt allows a child Tractor run when the sprint's validator starts
   one. Proof: unit tests plus `show` output.
2. `intake`: reads the seed and the repository, writes `intake.md` and
   the first `research/plan.md`. Proof: a nested run on a one-line seed
   leaves both files with the documented shape.
3. The research parallel node (at most five branches) and its fan-in:
   leaves under `research/`, `INDEX.md`, `findings.md`. Proof: nested run
   on a seed with a planted source; leaves, index entry, and one finding
   exist; the timeline shows the branches.
4. The index quality tool node and the `halt` tool node with the loop's
   routes. Proof: the halt command under the three states of P2's
   design, and one nested run whose halt routes back once.
5. `brief`: elicit then prune, batched question files with `Promise:`
   lines, promise list, exclusions, open plan entries. Proof: P1.
6. The observer and the shared proof library (`validation/observer.sh`,
   `validation/lib/`). Proof: a nested run answered end to end by the
   observer with its tree beside the run directory.
7. The brief/research loop end to end. Proof: P2.
8. `decompose` for MEDIUM and LARGE with the chapter and sprint
   skeletons. Proof: `validate-plan` shape rules pass on both sizes from
   one seed each.
9. Validation ledger generation and `design` with both archetypes and
   the holdout writer. Proof: a nested run leaves one design per promise
   and, for a universal promise, a holdout under the state root.
10. `review` with pass and fail edges and the notes files. Proof: P3
    and P4.
11. Pass files, the generated review ledger, and `reviewer` with routes
    to the owning node. Proof: a nested run in which one pass fails, its
    owning node runs, and the pass is re-selected and marked.
12. `assemble` and `approve`, the human gate. Proof: P5.
13. The four supervisors in the YAML with `models.yaml` roles. Proof:
    P7.
14. Docs and skill: an agent reading only the docs starts a nested plan
    run on a one-line seed and answers its first question (the chapter
    5 portion of P9's reader). Proof: the reader's run directory.
