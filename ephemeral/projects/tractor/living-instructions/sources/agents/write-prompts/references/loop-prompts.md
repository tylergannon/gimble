# Loop Prompts

Use this reference when a prompt is part of an agent loop, workflow, scheduler,
or evaluator.

## Loop Anatomy

A useful loop has:

- work source,
- context builder,
- prompt renderer,
- agent or subagent runner,
- verifier,
- state store,
- halt policy,
- budget guard,
- escalation path.

If any of these are missing, the prompt will tend to grow word-salad patches for
problems the loop should solve directly.

## Prompt Surfaces

Keep these separate when possible:

- seed prompt: starts the work;
- per-iteration prompt: says what changed since last pass;
- verifier prompt: judges against the definition of done;
- summarizer prompt: persists state for the next pass;
- handoff prompt: tells a human what decision is needed.

## Halt Conditions

Every loop needs explicit stop rules:

- done condition is satisfied;
- max iterations or max wall time reached;
- no material diff from prior iteration;
- verifier returns the same failure class repeatedly;
- budget threshold reached;
- human judgment required.

## Prompt Rule

Write the smallest prompt that lets the loop observe and decide correctly. Put
repeatable enforcement in the verifier or tool path, not in repeated prose.
