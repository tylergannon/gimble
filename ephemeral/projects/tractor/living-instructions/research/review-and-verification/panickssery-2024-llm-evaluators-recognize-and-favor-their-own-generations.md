# Panickssery, Bowman, Feng — LLM Evaluators Recognize and Favor Their Own Generations

## Purpose
Establishes that self-preference in LLM judges is causally tied to self-recognition, the mechanism behind the "reviewer on another provider" promise.

## Pinned
- arXiv:2404.13076v1 (2024-04-15); NeurIPS 2024. License: CC BY 4.0.
- https://arxiv.org/abs/2404.13076

## Key concepts
- Setting: summarization (CNN/DailyMail, XSUM, 1,000 articles each); evaluators Llama-2-7b-chat, GPT-3.5, GPT-4 (Sec. 2.1).
- Out-of-the-box self-recognition: GPT-4 73.5% pairwise; all models above chance, easiest against human text (Sec. 2.3). Individual (non-comparative) setting is much weaker (Sec. 2.3).
- After fine-tuning on 500 examples, GPT-3.5 and Llama 2 exceed 90% self-recognition (Abstract; Sec. 3.2).
- Self-preference strength is linearly correlated with self-recognition ability (Fig. 1; Sec. 3.2).
- Controls: fine-tuning on unrelated tasks (length, vowel counting, readability; Appendix B) and constant/random controls do not produce the effect (Sec. 3.4); explicit source labels shift preference (Sec. 3.5).
- Indirect evidence the preference is not quality-driven: summed pairwise self-preference of two LLMs exceeds 1 (Sec. 5.2).

## Bounded comparison
Like ours but only tests same-model judging of same-model text; it does not measure whether a judge on a *different* provider is free of the bias, only that recognition drives it.

## Gotchas
- Summarization only (Sec. 5.2); no code or spec review, so magnitudes do not transfer.
- No human ground truth on quality, so "bias" is inferred from asymmetry, not measured against a gold standard (Sec. 5.2).
- Pairwise comparison is where recognition is strong; a single-artifact verdict (our review passes) sits in the weaker "individual" regime (Sec. 2.3), so the bias may be smaller in our format than the headline suggests.
- Does not test the case where the judge shares training data with the author (see the preference-leakage leaf).

## Recipe
- To reproduce the self-recognition probe on your own artifacts, start at Sec. 2.2-2.3 (pairwise and individual prompts).
- To argue provider separation from mechanism, start at Fig. 1 and Sec. 3.4 (controls).
