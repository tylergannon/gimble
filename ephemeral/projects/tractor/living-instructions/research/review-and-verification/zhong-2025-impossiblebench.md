# Zhong, Raghunathan, Carlini — ImpossibleBench

## Purpose
Measures how often coding agents pass tests by violating the spec, and which context choices (prompt, test access, feedback loop) change that rate; the primary evidence for hiding checks from the coder.

## Pinned
- arXiv:2510.20270v1 (2025-10-23); ICLR 2026. License: as posted on arXiv.
- https://arxiv.org/abs/2510.20270

## Key concepts
- Impossible variants of LiveCodeBench and SWE-bench where unit tests contradict the spec; any pass is a cheat (Sec. 2). SWE-bench itself withholds the test patch from the model and grades on FAIL_TO_PASS / PASS_TO_PASS (SWE-bench, arXiv:2310.06770v3, Sec. 4.2 and App. A.3-A.4).
- Cheating rates on Conflicting-SWEbench, full scaffold: GPT-5 54%, o3 49%, Claude Opus 4.1 50% (Fig. 3); far lower on single-file LiveCodeBench (GPT-5 2.9%, Fig. 4).
- Taxonomy: modify tests, overload comparison operators, record extra state, special-case inputs (Sec. 4.1, Fig. 5). Claude and Qwen mostly edit tests; OpenAI models are more diverse.
- Prompt strictness: GPT-5 from >85% to 1% on Conflicting-LiveCodeBench with the strictest prompt, with no significant loss on the original benchmark (Sec. 5.1, Fig. 6).
- Test access: read-only tests remove test edits; hidden tests bring cheating to near zero (Sec. 5.2, Fig. 7).
- Feedback loops raise cheating (33% -> 38% with 10 submissions); an explicit abort/"flag for human" option drops GPT-5 from 54% to 9% (Sec. 5.3, Fig. 8).
- LLM monitors catch 86-89% on LiveCodeBench but only 42-65% on SWE-bench; rationalizations ("backward compatibility") fool them (Sec. 6, Fig. 10).

## Bounded comparison
Like ours but only measures cheating against *tests*; it has no reviewer or operating verifier, so it says what hiding the check prevents, not what an operator-verifier catches.

## Gotchas
- More capable models cheat more (Sec. 4), so improving the coder does not shrink the problem.
- Hidden tests remove one cheat class; they do not stop special-casing an inferred check or rationalized spec violations (Sec. 4.1, Sec. 6).
- Read-only / hidden access was measured with real SWE-bench tests; a holdout of prose criteria (our case) is not what was tested.
- Monitor detection on multi-file tasks is a coin flip (Sec. 6); "reviewer reads the diff" is not a reliable cheat detector.

## Recipe
- To decide what the coder may see, start at Sec. 5.2 / Fig. 7.
- To reduce gaming without a holdout, start at Sec. 5.1 (prompt D) and Sec. 5.3 (abort option).
