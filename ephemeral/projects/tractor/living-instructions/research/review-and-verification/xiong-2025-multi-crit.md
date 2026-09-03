# Xiong et al. — Multi-Crit: Benchmarking Multimodal Judges on Pluralistic Criteria-Following

## Purpose
Only source found that directly compares judging one criterion per call against judging all criteria in one call, on the same instances; evidence for the "one question per fresh reviewer" design.

## Pinned
- arXiv:2511.21662v2 (2026-03-12). License: as posted on arXiv.
- https://arxiv.org/abs/2511.21662

## Key concepts
- Protocol: each criterion judged in its own inference; the judge must follow a supplied, sometimes non-standard criterion rather than a general quality prior (Sec. 4, Implementation Details).
- Pluralistic accuracy is low even for proprietary models: best ~32.78% (o4-mini) on open-ended, ~53% on verifiable reasoning (Tables 3-4).
- Trade-off Sensitivity (TOS): whether the judge notices that criteria conflict within one instance (Sec. 3.5).
- Joint multi-criterion judging (all criteria in one pass) lowers accuracy for most models — GPT-4o from ~69.6% to ~50% on open-ended — and reduces trade-off sensitivity; GPT-5 is the exception with a small gain (Appendix E.2).
- Test-time scaling helps inconsistently; only o4-mini gains reliably (Sec. 4.2, Fig. 5). Fine-tuned critic models do not generalize to new criteria (Sec. 4.2).
- Judges track a criterion better when human annotators also agree on it (Fig. 6).

## Bounded comparison
Like ours but only image+text judging of short responses on 1-of-N criteria; no code, no specs, and no fresh-session isolation between criteria.

## Gotchas
- Separate-per-criterion beats joint for most models, but absolute criterion-following is poor (~30% on open-ended), so a one-question reviewer is more faithful, not reliable.
- Joint judging *lost* trade-off sensitivity; a design of six isolated questions therefore has no pass that sees conflicts between axes at all — the paper does not test a "seventh, cross-cutting" pass.
- The one model that improved under joint judging is the newest (GPT-5); the advantage of separation may shrink with model generation.
- Multimodal only; transfer to plan/spec review is assumed.

## Recipe
- To justify one criterion per reviewer call, start at Appendix E.2.
- To decide which questions need an integrative pass, start at Sec. 3.5 (TOS) and Fig. 6.
