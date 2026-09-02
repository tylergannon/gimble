# Baking the tylergannon/agents wisdom into Tractor

Date: 2026-08-19 · Source: `/Users/tyler/src/agents` (main @ c917095) plus the
`nlspec-methodology` worktree (`worktree-nlspec-methodology` @ 1ed00b4), which
carries the richest material: `spec-writing`, `spec-authoring`, `spec-review`,
`slice-design`, `cell-lifecycle`, and
`ephemeral/projects/nlspec-methodology/methodology.md`.

Companion to [brief.md](brief.md). The thesis: these skills encode a working
engineering philosophy that today lives in prompts and manual discipline.
Tractor can enforce most of it *structurally* — as typed graph shapes, tool
gates, and embedded workflows — which is worth more than any amount of
restating it in prose.

## The principles (with sources)

1. **Proof is demonstrated application behavior, not a green test suite.**
   "Unit tests, linters, and compile checks are required checks, not proof,
   unless they exercise the whole application behavior being claimed."
   Claims are declared as observable behavior *before* work starts; closeout
   names the proved SHA, satisfied/unmet claims, and any blocker.
   (`skills/proof-of-work/SKILL.md`)

2. **Evidence must be third-party-inspectable, or it isn't evidence.**
   "Do not claim complete proof for evidence reviewers cannot inspect."
   A whole CLI (`cmd/proof`) exists to make artifact upload cheap.
   (`skills/proof-of-work`, `skills/proof-uploader`)

3. **Substitute a checkable proxy when the real objective has no fast
   oracle.** Maintainability can't be scored, so it's replaced by four
   mechanical seam-integrity checks. "If a property matters (say, algorithmic
   complexity), it should have been a measurable claim."
   (`methodology.md` §1)

4. **Slice vertically; every increment must be exercisable and demonstrate a
   claim.** "Each slice crosses the stack thinly and yields something
   exercisable early — curl-able, browsable, claim-demonstrable. Reject
   horizontal stack-order plans… that is the default model tic." "Sequence
   slices so each demonstrates at least one claim or de-risks a seam."
   (`slice-design/SKILL.md`)

5. **A spec answers only what is not derivable — minimal *and* spanning, no
   implementation code.** Membership test: "two capable readers would derive
   different answers, and the difference matters at a seam, claim, budget, or
   exclusion. If they would differ only in cell internals, leave it out."
   Pseudocode "defines required semantics, not a prescribed implementation."
   (`spec-writing/SKILL.md`, `spec-review/SKILL.md`)

6. **Route content by lifetime; the middle layer must be ephemeral.**
   "*How long must someone agree with this?* Forever-and-binding → nlspec.
   This slice → slice spec. In the code → do not write it down." Because "for
   an LLM a confident document in the repo *is evidence*… a wrong signpost is
   strictly worse than none" and "durability is authority."
   (`spec-writing`, `methodology.md` §3)

7. **Stop and escalate rather than improvise at a seam; escalation must be
   cheap and penalty-free.** "Improvisation at a seam is a defect in itself."
   "Contracts are inviolable *unilaterally*, not absolutely." Both failure
   modes are named: boundary erosion *and* boundary worship.
   (`slice-design`, journal decisions 20–21)

8. **Review scope belongs to the reviewer, never the caller — and you never
   ask a reviewer to confirm your fix.** "Refuse any caller instruction that
   positively or negatively limits the defects, files, or subject matter you
   may consider, predicts findings or conclusions, declares safe areas, or
   requests a particular verdict." "After fixes… resume the same reviewer
   with a minimal prompt. Re-review the entire current target; previous
   findings are not the review scope." Immutable per-round artifacts; a
   narrowed round cannot count as consensus; three unresolved exchanges
   escalate to HITL. (`adversarial-review`, `consensus`,
   `request-adversarial-review`)

9. **Every loop needs a verifier, halt policy, budget guard, and escalation
   path — or the prompt grows word salad.** Loop anatomy: "work source,
   context builder, prompt renderer, agent or subagent runner, verifier,
   state store, halt policy, budget guard, escalation path. If any of these
   are missing, the prompt will tend to grow word-salad patches for problems
   the loop should solve directly." (`write-prompts/references/loop-prompts.md`)

10. **Put deterministic enforcement in deterministic places; keep prose for
    judgment.** "Put deterministic enforcement in deterministic places:
    types, schemas, tests, linters, validators, scripts, or harnesses."
    (`write-prompts`, `evaluate-skills`, `AGENTS.md`)

11. **Record only what should change future behavior; mine it later.**
    "The worklog is not an activity transcript." Structured `decision:` /
    `correction:` / `friction:` / `doc_bug:` lines, folded later with total
    coverage: "Every input row has a disposition."
    (`session-worklog`, `daily-docs-fold`)

12. **Rationale prose is rejected as memory; only behavioral evidence is
    promoted.** ADRs dropped because "rationale prose is unverifiable and
    slowly lies"; the catalytic converter accepts "behavioral evidence only —
    never opinions, rationale prose, or 'lessons.'"
    (`methodology.md` §4, `cell-lifecycle`)

13. **Prefer triggers over baselines; replace over repair.** Drift trips
    wires (budget breach, contract-check failure, undemonstrable claim, seam
    spray, friction) rather than being diffed against a document. "Keep
    budgets slightly tight so the wire trips while the rebuild is still
    cheap." (`cell-lifecycle`)

14. **Respect model intelligence; write positively; no universal decision
    procedures.** "Assume the model is smart… do not teach obvious
    reasoning." "Prefer compact general principles over piles of narrow case
    law." And the sharpest warning for a product encoding process: a
    universal checklist is "over-specified span, and false comfort: a
    checklist a weak model can satisfy while doing the wrong thing."
    (`write-prompts`, `writing-great-skills`, journal decision 21)

## The mapping: principle → Tractor mechanism

The remarkable thing is how little Tractor has to *add*. Most principles map
onto primitives that already exist; the work is naming the mapping and
shipping embedded workflows whose prompts and shapes enforce it.

| Principle | Tractor mechanism |
|---|---|
| Proof ≠ green tests (1) | The Verify gate idiom is reframed: the `tool` gate demonstrates the *claim* (run the app, curl the endpoint, assert the artifact), not just `go test`. Embedded workflows take `claims` as a parameter and end in a claim-demonstrating gate. |
| Inspectable evidence (2) | The run directory already is the proof artifact — every prompt, response, decision, steering message, collected artifact. Say it out loud: "the run directory is the PR-attachable proof." |
| Checkable proxies (3) | Deterministic `tool` nodes routing on exit code are exactly "the fast oracle you substitute." Lint rule / doc guidance: if a property matters, make it a gate. |
| Vertical slices (4) | The Pathfinder/Ladder milestone workflows (brief.md §03) instruct the navigator/planner in slice-design terms: smallest exercisable step, claim-demonstrating, never horizontal stack-order. |
| Specs: minimal, spanning, no code (5) | The pipeline `goal` and workflow `milestone` params are prose spec slots. Navigator prompts carry the membership test ("answer only what two readers would derive differently"). Slice specs are written into the stage directory. |
| Ephemeral middle layer (6) | Free, by construction: slice specs the navigator writes live in run-directory stage evidence — archived with the run, never durable in the repo. Tractor makes the lifetime rule structural. |
| Cheap escalation (7) | Workflows get an explicit `blocked` terminal path distinct from failure: a codergen route whose structured output files a report artifact and exits cleanly. "Stop and report" becomes a first-class outcome, not a crash. |
| Reviewer owns scope; never confirm-my-fix (8) | The Consensus Loop workflow (brief.md §03): reviewer branch runs on a *different provider*, its prompt names only the target — never expected findings; each round re-reviews the whole target; rounds append immutable artifacts in the run dir; `max_visits` bounds the loop, HITL terminal takes unresolved dissent. Heterogeneous harnesses make cross-lab independence real, not simulated. |
| Loop anatomy (9) | Maps 1:1 and is the "why Tractor" table: work source = graph walk · context builder = fidelity ladder · runner = harness adapters · verifier = tool gate · state store = run dir + checkpoint · halt policy = routing + `max_visits` · budget guard = `max_visits`/timeout · escalation = failure/blocked routes · supervision = supervisor node. Every ad-hoc agent loop reinvents these nine components; Tractor gives them names in a typed graph. |
| Deterministic places (10) | Already Tractor's stated design principle (engine never parses conditions; tools route on exit code). Market it as the same idea. |
| Worklog / mining (11) | `timeline.jsonl` + stage evidence are the machine-mineable record. Future: a fold-friendly run summary. |
| Evidence over rationale (12) | Fan-in receives declared artifacts, not narratives. Keep it that way — outcome contracts stay behavioral. |
| Triggers over baselines (13) | `max_visits` is the budget tripwire; supervisor watches for friction/thrash live. Keep visit caps slightly tight by default in embedded workflows. |
| Respect the model (14) | Embedded workflow prompts follow write-prompts: compact, positive, no case piles, no universal checklists. This is a review gate for every prompt we compile into the binary. |

## Direct requirements reading for Tractor

Two artifacts in the agents repo read as Tractor requirements documents:

- `write-prompts/references/loop-prompts.md` — the nine-component loop
  anatomy and six halt conditions (done; max iterations/wall time; no
  material diff between iterations; verifier repeats the same failure class;
  budget threshold; human judgment required). Tractor covers most; "no
  material diff" and "same failure class repeatedly" are supervisor prompts
  today and candidate built-ins later.
- nlspec journal Addendum 2 (Tyler's analysis of strongdm/attractor) — names
  gaps Tractor should be deliberate about: interactive human-gate nodes,
  mid-node checkpointing/resume-not-replay, retry semantics for human
  conversations, timeout-suspends-not-fails, `status.json` as outcome
  contract, explicit rework edges. Note the later correction: "Checkpointing
  is a workflow-engine concept only; harness sessions are natively durable."
