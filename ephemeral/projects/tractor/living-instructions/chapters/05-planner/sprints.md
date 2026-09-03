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
   the `not` constraint, a lint-time error when it cannot be satisfied.
   Proof: unit tests plus `workflow show` printing provider and model per
   node.
2. Intake node and the first research plan; the research parallel node
   (at most five branches), fan-in, index update, findings file, and the
   index quality tool node. Proof: nested run on a seed with a planted
   source; leaves and findings exist; the timeline shows the branches.
3. Brief node: elicit then prune, batched question files, promise list,
   exclusions, open plan entries; the halt tool node and the loop. Proof:
   P1 and P2 scenarios.
4. Decompose for MEDIUM and LARGE with the chapter and sprint skeletons.
   Proof: `validate-plan` shape rules pass on both sizes from one seed
   each.
5. Validation ledger generation; `design` with both archetypes and the
   holdout writer; `review` with pass and fail edges. Proof: P3 and P4
   scenarios.
6. Pass files, generated review ledger, `reviewer` with routes to the
   owning node; assemble; approve. Proof: P5 scenario.
7. The four supervisors in the YAML with `models.yaml` roles. Proof: P7
   scenario.
8. Docs and skill.
