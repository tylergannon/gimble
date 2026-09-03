# Li et al. — Preference Leakage: A Contamination Problem in LLM-as-a-judge

## Purpose
Quantifies judge bias by *relatedness* (same model, inheritance, same family) rather than same-model only; the strongest available evidence on how much provider separation buys.

## Pinned
- arXiv:2502.01534v3 (2026-03-04). License: as posted on arXiv (not stated in the fetched HTML).
- https://arxiv.org/abs/2502.01534

## Key concepts
- Three relatedness types between data generator and judge: same model, inheritance, same family (Sec. 3).
- Preference Leakage Score (PLS): symmetric normalized win-rate advantage of a judge for its own student (Sec. 4.1, Eqs. 5-6).
- Judges GPT-4o-2024-11-20, Gemini-1.5-flash, LLaMA-3.3-70B; students Mistral-7B-v0.1 and Qwen-2.5-14B trained on judge-generated data; benchmarks Arena-Hard, AlpacaEval 2.0 (Sec. 4.1).
- PLS by relatedness (Table 2): same model ~23.6% avg; inheritance ~19-22%; same family same series ~8.9%; same family different series ~2.8%.
- Judges cannot explicitly recognize their students (29-53%, near chance) but a BERT classifier can (82.4%), so the leak rides on style features, not recognition (Sec. 5.5, Table 5).
- Subjective questions and dimensions leak far more than objective ones (Sec. 5.6, Fig. 3).
- Ranking distortion on real leaderboards exceeds that of egocentric bias (Sec. 5.4, Table 4).

## Bounded comparison
Unlike ours because the "student" is a model fine-tuned on the judge's outputs, not an agent whose plan the judge reads; the monotone drop from same-model to different-series is the transferable part.

## Gotchas
- No cross-provider row: the design never measures a judge on a fully unrelated provider as a baseline; "different series, same family" at 2.8% is the closest, so "another provider removes the bias" is extrapolated, not shown.
- Bias is stronger on subjective axes (Sec. 5.6); review questions like "is this plan clear" are subjective, "does this test exist" is not.
- The judge's inability to self-recognize (Sec. 5.5) means asking a reviewer "did you write this?" is not a usable detector.
- Mitigations tested (Table 7) are training-side; nothing here validates prompt-level debiasing.

## Recipe
- To size the bias you are avoiding, start at Table 2.
- To decide which review questions need provider separation most, start at Sec. 5.6 / Fig. 3.
