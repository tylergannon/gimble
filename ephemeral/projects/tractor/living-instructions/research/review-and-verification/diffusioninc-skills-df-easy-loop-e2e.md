# diffusioninc/skills — df-easy-loop-e2e

## Purpose
A Claude Code / Codex CLI orchestration skill that runs requirements -> plan -> cross-provider critique -> coding/validation/review loop, with a "validation holdout" the coder never sees. Closest running prior art to this project's holdout-plus-operating-verifier promise.

## Pinned
- Repo: https://github.com/diffusioninc/skills (main @ `2830b0024265d16fcb9fa6fe003b69ea685cc57d`, 2026-08-27)
- File: `.claude/skills/df-easy-loop-e2e/SKILL.md` (last changed `8039427048432c03a9fb153ae20b6ea6e5e4448a`, 2026-08-25); raw: https://raw.githubusercontent.com/diffusioninc/skills/main/.claude/skills/df-easy-loop-e2e/SKILL.md (unauthenticated raw fetch returned 404 on 2026-09-02; the authenticated contents API works)
- License: none declared (no LICENSE file; `licenseInfo: null`). Treat as all-rights-reserved; read, do not copy.

## Key concepts
- Holdout authored at requirements time by the interviewing agent, kept in `validation-holdout.md` alongside `requirements.md` (SKILL.md L241-242, L258-260).
- Coder never reads the holdout: `coding.reads` is plan, requirements, review only (L325-328); `review.reads` also omits it (L379-384). Only `validation` reads it (L354-356).
- Validation "must run the software, not only inspect code or run unit tests" and record commands and output (L368-370); any unmet holdout criterion fails validation even if everything else passes (L370-372).
- Review must route back to `coding` on holdout failure and must "not mention the holdout set or holdout criteria" in `review.md` (L404-409) — leak prevention by omission, not by sandbox.
- Cross-provider by role: plan on Claude (L266), critique on GPT (L283), plan-update on Claude (L300), coding + validation on GPT (L322, L351), review on Claude (L376).
- Orchestrator is forbidden to read sub-agent artifacts except the requirements Q&A handoff (L14-17, L200-204); routing is by `outcome.yaml` only (L37, L189-191).
- Validation runs with `context: empty` (L353); coding and review resume prior sessions (L324, L378).

## Bounded comparison
Like ours but only one validator and one reviewer per lap, and the validator (`gpt-5.6-sol`, L351) is the same model as the coder (L322) — provider separation exists between planner and critic, not between coder and verifier.

## Gotchas
- Holdout criteria are written by the same agent that writes requirements (L241-242); nothing checks that the holdout is not restated in `requirements.md`, which the coder does read. Leak by authorship, not by routing, is unguarded.
- `validation.md` records "whether every validation-holdout criterion is met" (L370-371) and `review` reads it (L384); review is told to withhold the criteria from `review.md`, so the seal depends on a prompt instruction to a resumed session, not on access control.
- "Run the software" is an instruction, not an enforced check: no artifact schema demands a command transcript, so a validator that only runs unit tests still satisfies the file contract.
- No evidence in the repo that the holdout catches defects the visible checks miss; the mechanism is asserted, not measured.

## Recipe
- To write a holdout that the coder cannot see, start at L241-242 and L325-328 (creates vs reads matrix).
- To make a validation verdict override visible checks, start at L363-373.
