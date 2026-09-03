# Chen et al. — Do LLM Evaluators Prefer Themselves for a Reason?

## Purpose
Splits self-preference into legitimate (own answer was right) and harmful (own answer was wrong) on verifiable tasks, including code; shows reasoning at judge time cuts the harmful part.

## Pinned
- arXiv:2504.03846v2. License: as posted on arXiv.
- https://arxiv.org/abs/2504.03846

## Key concepts
- Benchmarks with gold answers: MATH500, MMLU, MBPP+ (executable tests) (Sec. 2).
- LSPR (Legitimate Self-Preference Ratio), Eq. 4, Sec. 3.2; HSPP (Harmful Self-Preference Propensity), Eq. 5, Sec. 4.1.
- Strong models prefer themselves mostly legitimately: Qwen2.5-72B LSPR 96.57% on MATH500 (Sec. 3).
- But when wrong, strong models are more likely to still prefer themselves: HSPP 86% on MATH500 vs 55% overall self-preference; capability correlates with HSPP (Fig. 5).
- Generating CoT before the verdict reduces HSPP across all models; long-CoT (DeepSeek-R1 style) lowest (Sec. 4.2, Fig. 6).
- MBPP+ (code) shows the weakest capability-HSPP correlation (Sec. 3-4, Tables 2/4/6).

## Bounded comparison
Like ours but only pairwise judging of short verifiable answers; our reviewers judge one long artifact without a competitor, which this paper does not test (Appendix D).

## Gotchas
- Raw self-preference rate overstates the problem: much of it is the better model being right (Sec. 3). Provider separation trades away legitimate preference too.
- Harmful self-preference is concentrated where the author is confidently wrong — exactly the case a reviewer exists for — and grows with model strength (Fig. 5).
- Pairwise only; no pointwise results, so the size of harmful self-preference in a single-verdict review is unknown (Appendix D).
- The reduction from reasoning is a judge-side effect; it does not license same-provider review, only says a reasoning judge is less harmful.

## Recipe
- To measure harmful vs legitimate self-preference on your own tasks, start at Sec. 3.2 (Eq. 4) and Sec. 4.1 (Eq. 5).
- To justify "reason before verdict" in reviewer prompts, start at Sec. 4.2 / Fig. 6.
