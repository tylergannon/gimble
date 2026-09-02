# Prompt Editing

Use this reference when reducing or improving existing prompt text.

## Audit Questions

- What exact behavior is the prompt supposed to cause?
- What is the model's real input after template rendering?
- Which parts are task intent, domain context, examples, schema descriptions,
  deterministic constraints, or verifier instructions?
- Which instructions are defensive reactions to one historical failure?
- Which words contradict code, schema, tests, or current product behavior?
- What proof would show the prompt now works?

## Rewrite Moves

1. Delete duplicates and generic warnings.
2. Replace case piles with a single discriminating principle.
3. Move exact output shape into schema or parser tests.
4. Move invariant enforcement into evals, fixtures, or scripts.
5. Keep examples only when they are representative and current.
6. Make exception wording explicit only when the exception is common and costly.

## Common Smells

- "Always" and "never" rules that are really form-specific exceptions.
- Long lists of edge cases without a classification principle.
- Prompt text that describes implementation internals instead of the task.
- Schema constraints duplicated in prose.
- Regressions handled by adding another warning instead of adding a test.
- Generated prompt fragments that no one has rendered end to end.

## Better Prompt Skeleton

```md
Task: [one sentence]

Inputs:
- [source text / file / fixture / schema]

Output:
- [format or schema]

Rules:
- [few domain rules that cannot be enforced elsewhere]

Check yourself:
- [short verification rubric]
```

## Proof

A prompt edit is not done until at least one current failure or representative
fixture has been rerun. For broad prompts, prefer a small held-out eval set over
one hand-picked example.
