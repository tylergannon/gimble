---
name: write-prompts
description: >
  Write, edit, or review prompts and skill instructions with concise,
  model-respecting guidance. Use when creating or improving prompt.ts files,
  system prompts, judge/eval prompts, agent skills, or when the user complains
  about prompt word salad, over-specific instructions, or brittle prompting.
---

# Write Prompts

Use this skill when the artifact is prompt text or prompt-like instructions.
The job is not to make the prompt longer. The job is to make the model's task
clearer and move non-language constraints into better mechanisms.

## Principles

- Assume the model is smart. Explain the task, context, constraints, and output
  contract; do not teach obvious reasoning.
- Delete before adding. Remove stale cases, duplicated rules, contradictions,
  and defensive prose before introducing new wording.
- Separate layers: prompt intent, examples, schema, code, tests, evals, tools,
  loop state, and memory are different surfaces.
- Put deterministic enforcement in deterministic places: types, schemas, tests,
  linters, validators, scripts, or harnesses.
- Prefer compact general principles over piles of narrow case law.
- When the prompt runs in a loop, design the loop's observations, verifier,
  state, halt policy, and escalation path too.

## Workflow

1. Identify the prompt consumer, trigger, rendered prompt surface, expected
   output, and current verifier.
2. Render or inspect the full prompt the model actually sees. Do not edit a
   helper fragment in isolation when generation changes the final prompt.
3. Classify the failure: missing context, ambiguous intent, bad examples,
   schema mismatch, absent verifier, stale policy, or model behavior.
4. Write the smallest change that targets that failure. Prefer removal or
   replacement over additive warnings.
5. Add or improve the check that proves the prompt works: fixture, eval,
   snapshot, unit test, judge rubric, or transcript comparison.
6. Verify with before/after evidence and call out any behavior still unproved.

## When To Read References

- Read [skill-writing.md](references/skill-writing.md) when writing or revising
  a `SKILL.md`.
- Read [prompt-editing.md](references/prompt-editing.md) when auditing an
  existing prompt file or reducing prompt sludge.
- Read [loop-prompts.md](references/loop-prompts.md) when the prompt is part of
  an agent loop, workflow, scheduler, or evaluator.

## Closeout

Report:

- prompt files changed,
- instruction classes deleted or added,
- verifier/eval/check added or run,
- before/after behavior proved,
- remaining assumptions or unresolved failures.

source: agents
