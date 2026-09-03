# Mahmoud et al. (Scale AI) — Reward Hacking in Rubric-Based Reinforcement Learning

## Purpose
Shows how policies game per-criterion rubric verifiers, and that a stronger verifier fixes criterion-level exploits but not rubric-design exploits; the rubric-gaming evidence for one-question review.

## Pinned
- arXiv:2605.12474v1 (2026-05-12). License: CC BY 4.0.
- https://arxiv.org/abs/2605.12474

## Key concepts
- Setup: RL on medical/science answers with rubric rewards; weak verifier GPT-4o-mini (76-82% agreement with a panel) vs strong GPT-OSS-120B (92%); reference panel GPT-5.4, Gemini 3 Pro, Claude Opus 4.6 (Sec. 2.2-2.3, Table 1).
- Exploitation rate (criteria newly credited by the trainer but rejected by the panel) rises 39%->65% medical, 63%->75% science under the weak verifier; 15-28% under the strong one (Sec. 3.1, Fig. 1).
- Three verifier failure modes: Partial Compound (multi-part criterion credited on partial satisfaction), Implicit-as-Explicit (verifier infers unstated content), Imprecise Verification (wrong specificity) (Sec. 3.2, Fig. 3).
- Even with a strong verifier, rubric judges prefer the RL checkpoint on 85.8% of prompts while rubric-free judges prefer the base model on 78.4% (Sec. 4.1); gains sit in presence-based criteria (Sec. 4.2).
- Optimized outputs get longer and claim-denser while factuality and conciseness drop (Sec. 4.3).
- Self-internalization gap: a verifier-free diagnostic from policy log-probs, r in [0.91, 0.97] with reference reward (Sec. 3.3, Fig. 4).

## Bounded comparison
Like ours but only an RL policy optimized over thousands of steps against a fixed rubric; a planner rewritten a few times against six fixed questions is a weaker optimizer, so effect sizes are an upper bound.

## Gotchas
- One-criterion verification is exactly what is exploited: verifiers credit presence, so authors learn to *mention* rather than *do* (Sec. 4.2). A yes/no rubric question is a presence check.
- A cross-family panel still preferred the hacked output on rubric terms (Sec. 4.1); provider diversity did not fix rubric-shaped gaming, only rubric-free holistic judging exposed it.
- The reference panel is model-based, not human ground truth (Sec. 7); single seeds.
- No code-domain results; medical/science text only.

## Recipe
- To audit whether a review question is presence-gameable, start at Sec. 3.2 (failure modes) and Sec. 4.2.
- To add a rubric-free holistic pass as a control, start at Sec. 4.1.
