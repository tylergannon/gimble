---
items:
  - name: traceability
    check: Every promise in promises.md reaches a chapter or sprint item, and every chapter and sprint item serves at least one promise.
    doc: ephemeral/projects/tractor/living-instructions/plan-review/01-traceability.md
    done: true
  - name: consistency
    check: One name per concept and no rule stated twice with drift across declaration.md, planning-workflow.md, decisions.md, the chapter docs, and the sprint docs.
    doc: ephemeral/projects/tractor/living-instructions/plan-review/02-consistency.md
  - name: slicing
    check: Every chapter and sprint is a vertical slice that is exercisable when done; no horizontal stack-order plan.
    doc: ephemeral/projects/tractor/living-instructions/plan-review/03-slicing.md
  - name: proof quality
    check: Across all promises, no validator can be satisfied while its promise is false, nothing is trivially true, and nothing is stricter than its promise.
    doc: ephemeral/projects/tractor/living-instructions/plan-review/04-proof-quality.md
  - name: scope
    check: Nothing in the plan exceeds the promises; the exclusions are respected.
    doc: ephemeral/projects/tractor/living-instructions/plan-review/05-scope.md
  - name: executability
    check: Every sprint fits one agent turn, every command runs from the workdir, every path resolves, every ledger parses.
    doc: ephemeral/projects/tractor/living-instructions/plan-review/06-executability.md
  - name: holistic
    check: A fresh reviewer, asked with no rubric whether they would accept this as the plan for this product, says yes.
    doc: ephemeral/projects/tractor/living-instructions/plan-review/07-holistic.md
---

# Plan review

Seven passes (decision 58), one fresh reviewer each on a provider other
than the planner's, told only the pass question and the package. In the
manual run the reviewer is a codex session; its answer is saved as
`<pass>/review-N.md` and its route is its last line. A fail names the
owning node; the planner fixes and the pass is re-selected.

The package under review: `declaration.md` (brief and promises),
`promises.md`, `chapters.md` and `chapters/04-*`, `05-*`, `06-*`,
`validation/`, `research/`, `planning-workflow.md`, `decisions.md`
37 onward.
