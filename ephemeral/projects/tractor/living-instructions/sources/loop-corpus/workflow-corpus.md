# Workflow and loop corpus

This corpus asks a practical question: what workflow shapes do people actually
repeat, and what makes those shapes work? “Adopt” means the insight should
influence Tractor guidance; it does not mean copy the source implementation.

## Attractor-family examples

### 1. Attempt -> mechanical check -> retry

- Sources: Microsoft Amplifier's
  [convergence loop](https://github.com/microsoft/amplifier-bundle-attractor/blob/main/examples/pipelines/00-convergence-loop.dot),
  [pipeline design principles](https://github.com/microsoft/amplifier-bundle-attractor/blob/main/docs/PIPELINE_DESIGN_PRINCIPLES.md),
  and numerous test/fix loops across the implementation inventory.
- Observation: the most repeated useful topology is a worker followed by an
  external check. Check failure returns concrete output to the worker; success
  exits.
- Adopt: default to two work-bearing nodes. Put “done” in a tool exit code when
  possible, keep its output in a readable file, and bound visits.
- Caution: Amplifier's practical examples have accumulated operational detail
  specific to that engine. Tractor guidance should preserve the shape, not the
  shell ingenuity.

### 2. Plan -> implement -> verify, with selective replanning

- Sources: upstream
  [Attractor spec example](https://github.com/strongdm/attractor/blob/main/attractor-spec.md),
  [martinemde/attractor `develop.dot`](https://github.com/martinemde/attractor/blob/main/pipelines/develop.dot),
  [arikWaisman/klaus plan-and-execute](https://github.com/arikWaisman/klaus/blob/main/pipelines/plan-and-execute.dot),
  and [jhugman/attractor-pi-dev](https://github.com/jhugman/attractor-pi-dev).
- Observation: planning pays when it produces a durable artifact or a genuine
  decision boundary. Mechanical failure normally returns to implementation;
  implementation should return to planning only when the approach, not merely
  the code, is wrong.
- Adopt: make the plan name files, constraints, risks, and verification; have
  implementation read it from the workspace. Do not replan on every red test.

### 3. Diagnose -> fix -> reproduce/test

- Sources: Amplifier's
  [practical bug-fix pipeline](https://github.com/microsoft/amplifier-bundle-attractor/blob/main/examples/pipelines/practical/bug-fix.dot),
  Superpowers'
  [systematic-debugging method](https://github.com/obra/superpowers), and
  dark-factory's slim bug-fix variants.
- Observation: repeated identical failure is evidence that more patching is not
  progress. Route back through diagnosis when the evidence is unchanged or the
  working theory has failed.
- Adopt: ask diagnosis to write a short current hypothesis; make verification
  reproduce the original symptom, not merely run a broad suite.
- Caution: a sophisticated “root-cause wall” is optional. A three-node loop
  already captures most of the value.

### 4. Draft -> independent review -> revise

- Sources: [samueljklee code review](https://github.com/samueljklee/attractor/blob/main/examples/code_review.dot),
  [tgoodwin feedback loop](https://github.com/tgoodwin/tractor/blob/main/examples/haiku_feedback.dot),
  [SWE-Review](https://arxiv.org/abs/2607.06065), and multi-lens review
  examples in Amplifier, Tracker, Factorial, and dark-factory.
- Observation: reviewer value comes from a different mandate or fresh context,
  not from adding another generic model call. Review must emit actionable
  findings that revision can see.
- Adopt: use fresh reviewer context (`fidelity: none`) when independence matters;
  route “no material findings” to success and findings to one revision node.
- Caution: multi-model “consensus” is much more expensive and rarely needed for
  ordinary work. One competent independent reviewer is the bread-and-butter
  shape.

### 5. Research -> decision/plan -> execution

- Sources: [Integral Engineering's research-plan-implement cycle](https://engineering.integral.de/posts/sdlc2-research-plan-implement-cycle/),
  [samueljklee research-then-build](https://github.com/samueljklee/attractor/blob/main/examples/research_then_build.dot),
  and dark-factory's minimal research pipeline.
- Observation: research is reusable when it records the source revision,
  constraints, and integration points. Otherwise it is just an expensive warmup
  embedded in the next prompt.
- Adopt: split research only for unfamiliar code, multiple repositories, live
  external facts, or a decision that deserves review. Let the workspace artifact
  carry the handoff.

### 6. One ready item -> implement/test -> record -> next item

- Sources: Geoffrey Huntley's
  [Ralph essay](https://ghuntley.com/ralph/),
  [snarktank/ralph](https://github.com/snarktank/ralph),
  [agenticloops-ai/ralph-loop](https://github.com/agenticloops-ai/ralph-loop),
  [jhugman/attractor-pi-dev's issue loop](https://github.com/jhugman/attractor-pi-dev),
  and [strongdm/agate](https://github.com/strongdm/agate).
- Observation: long-running autonomy works by shrinking each lap. One item fits
  in one context; progress and acceptance criteria live in files; verification
  and a visit/iteration cap provide backpressure.
- Adopt: Tractor can express the inner lap well. Prefer a bounded loop over one
  giant “finish the project” node.
- Caution: Huntley explicitly frames Ralph as best for greenfield work and says
  he would not use it unchanged in an existing codebase. The skill should not
  recommend autonomous backlog loops by default.

### 7. Parallel independent attempts -> fan-in synthesis

- Sources: Tractor's own
  [fan-out/fan-in example](https://github.com/tylergannon/tractor/blob/main/examples/parallel/fan-out-fan-in.json),
  Tracker's competitive implementation workflow, and Attractor-family
  multi-lens review examples.
- Observation: parallelism earns its cost when branches are genuinely
  independent and the fan-in can compare concrete artifacts. It is not a default
  replacement for a single strong worker.
- Adopt: reserve it for alternative designs, separate research questions, or
  independent review lenses. Give the fan-in an explicit comparison task.

### 8. Human gate at the high-leverage decision

- Sources: [Helix's plan/implement/review loop](https://www.helix.ml/docs/concept-agent-loop),
  [Integral Engineering](https://engineering.integral.de/posts/sdlc2-research-plan-implement-cycle/),
  [Sam McLeod's plan-act-review flow](https://smcleod.net/2025/04/my-plan-document-act-review-flow-for-agentic-software-development/),
  and Attractor human-gate examples.
- Observation: human attention compounds before an expensive or irreversible
  action—typically plan approval, scope choice, or ship—not after every agent
  step.
- Adopt: ask Codex to keep the loop autonomous until a judgment or authority
  boundary is reached. Tractor has no special human node; a codergen uses its
  available communication tools and routes on the answer.

## Broader workflow sources

### Official OpenAI: scored improvement loops

- Source: [Iterate on difficult problems](https://learn.chatgpt.com/use-cases/iterate-on-difficult-problems).
- Observation: define success before iterating; combine deterministic and
  judgment-based evaluation; keep machine-readable scores and a running log;
  change one major variable per lap; stop at explicit thresholds.
- Adopt: the skill should require an observable stop rule and should not use
  Tractor for “keep improving” without a threshold or bounded review rule.

### Official OpenAI: skills around plugin tools

- Sources: [Build skills](https://developers.openai.com/plugins/build/skills)
  and [Save workflows as skills](https://learn.chatgpt.com/use-cases/reusable-codex-skills).
- Observation: MCP supplies live operations; a skill supplies tool order,
  decision points, outputs, examples, and workflow boundaries. A skill should
  stay focused on a recognizable goal and move detail into references.
- Adopt: keep `use-tractor/SKILL.md` short; put complete JSON patterns in one
  reference; use the existing Tractor MCP tools rather than encoding another
  runtime.

### Sam McLeod: setup -> plan -> act -> review and iterate

- Source: [My Plan, Document, Act, Review flow](https://smcleod.net/2025/04/my-plan-document-act-review-flow-for-agentic-software-development/).
- Observation: the plan is durable state and the cycle scales down; small bugs
  should not drown in spec ceremony. Deterministic tools should handle
  deterministic outcomes.
- Adopt: teach a size threshold. A plan stage is optional, not a moral
  requirement.

### Integral Engineering: research -> plan -> implement

- Source: [Research, Plan, Implement](https://engineering.integral.de/posts/sdlc2-research-plan-implement-cycle/).
- Observation: research records constraints and integration points at a pinned
  revision; requirements precede architecture; real entry-point testing is a
  separate follow-up when needed.
- Adopt: when research is a node, tell it who consumes the artifact and what
  must be in it.

### Superpowers: brainstorm -> plan -> execute -> review -> finish

- Source: [obra/superpowers](https://github.com/obra/superpowers).
- Observation: task boundaries should have independently testable deliverables;
  plans and implementation are different sessions; review happens at meaningful
  gates, not bookkeeping granularity.
- Adopt: use separate Tractor nodes when the handoff merits a fresh context and
  can be made explicit in the workspace.

### Compound Engineering: research/plan/work/review/compound

- Source: [EveryInc/compound-engineering-plugin](https://github.com/EveryInc/compound-engineering-plugin).
- Observation: the useful repeatable units are distinct workflows rather than
  one omnipotent mega-loop. Research and review feed durable learning into later
  work.
- Adopt: the proposed Tractor skill should choose among a few small shapes, not
  prescribe a universal lifecycle graph.

### gstack: specialized plan, engineering, design, review, QA, and ship gates

- Sources: [garrytan/gstack](https://github.com/garrytan/gstack) and its
  [autoplan pipeline](https://github.com/garrytan/gstack/blob/main/autoplan/SKILL.md).
- Observation: different reviews answer different questions, but review depth
  should track the decision and risk. Shipping should verify plan completion,
  tests, scope drift, and user-visible behavior.
- Adopt: name a review node by its question (`review_security`,
  `review_requirements`), not merely `review`.
- Caution: its full review army is intentionally high ceremony and should not
  become Tractor's default.

### Plan-compliance research

- Source: [From Plan to Action: How Well Do Agents Follow the Plan?](https://arxiv.org/abs/2604.12147).
- Observation: adding task-relevant phases can degrade results when they do not
  align with the agent's own problem-solving strategy.
- Adopt: do not decompose cognitive micro-steps into graph nodes. Split on
  durable artifacts, independent checks, authority boundaries, or context
  isolation.

### Generate-review-revise research

- Source: [SWE-Review](https://arxiv.org/abs/2607.06065).
- Observation: an agentic reviewer that explores the repository and returns
  revision feedback improves issue resolution more than a one-shot fixed-context
  review.
- Adopt: make review an active repository task and feed its findings back; do
  not ask for a verdict on a pasted summary alone.

## Repeated advice across sources

1. Make each lap small enough to fit in one context and produce one coherent
   outcome.
2. Put state in inspectable artifacts, not only chat history.
3. Prefer a deterministic check for an observable condition.
4. Give every cycle a stop rule, visit cap, or escalation path.
5. Route different failure causes to different work when the distinction is
   observable; otherwise keep one correction path.
6. Use fresh context for independent review and shared thread state for repeated
   correction by the same worker.
7. Parallelize independent work, not sequential dependencies.
8. Ask a human at authority or taste boundaries, not as a substitute for tests.
9. Start with the smallest graph that changes the odds of success.
