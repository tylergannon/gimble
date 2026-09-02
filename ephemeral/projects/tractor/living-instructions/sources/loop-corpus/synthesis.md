# Tractor authoring advice derived from the corpus

## The decision: use Tractor or stay in one Codex task?

Use Tractor when at least one of these is load-bearing:

- the work should continue through several bounded attempts;
- a different agent turn should inspect or judge the result with fresh context;
- a deterministic command can decide whether to continue;
- the handoff should survive a long run or context reset as a workspace artifact;
- several independent attempts or investigations should run concurrently;
- the run needs external monitoring, steering, stopping, or resume.

Stay in one Codex task when the work is short, its natural implementation loop
already fits comfortably in one session, and splitting it would only rename
ordinary thoughts (“think”, “code”, “double-check”) as nodes. A straight chain
can still be useful for repeatable handoffs, but a chain with no meaningful
handoff or check is usually ceremony.

## Default shape: worker -> check -> worker

Start with one `codergen` worker and one `tool` check. The worker owns the whole
coherent task, including whatever internal planning it needs. The tool runs the
real repository command. Exit 0 goes to `success`; nonzero returns to the worker.
Set `max_visits` on the worker so an unfixable check becomes an honest failure.

This shape is better than a long chain because the graph expresses the one fact
that matters outside the worker: what observable evidence admits exit.

## When to add a third node

Add `plan` when the plan is a reviewable or reusable artifact, the task crosses
several modules, or implementation may need to reject the approach. Have it
write a short plan file. Ordinary test failures return to implementation;
fundamental approach failures return to planning.

Add `diagnose` when patching without a current theory is likely to thrash. Have
it write the symptom, evidence, hypothesis, and next discriminating check.

Add `review` when success includes judgment that a command cannot establish.
Give the reviewer a specific question and fresh context. Let a revision node
consume concrete findings.

Do not add separate nodes for cognitive steps that one competent worker can
perform without a durable boundary.

## Write the goal and node prompts at different altitudes

The top-level `goal` describes the outcome and constraints. A node prompt
describes that stage's responsibility, inputs, artifact, and routing decision.
Do not paste the whole goal into every node; `$goal` already supplies it.

A useful stage prompt answers four questions:

1. What part of `$goal` does this stage own?
2. What workspace or run artifacts should it read?
3. What concrete effect or artifact should it produce?
4. If it chooses among edges, what evidence distinguishes the choices?

## Use the workspace as the ordinary handoff

Tractor's most robust data channel is the shared workspace. Plans, test logs,
review findings, and progress notes are durable, inspectable, and available to
later nodes. Different nodes have different threads by default; revisits of the
same node reuse and compact that node's session.

- Use default `compacted` fidelity for a worker that revisits its own work.
- Use `fidelity: "none"` for a reviewer that should form an independent view.
- Give multiple nodes the same explicit `thread_id` only when conversation
  continuity is more valuable than role separation. Prefer a file handoff when
  the information should be auditable.

## Make routing honest

- A `tool` routes only on exit status: `on_success` and `on_error`.
- A `codergen` with several edges chooses from prose `condition` values. Tell it
  what evidence warrants each choice; the engine does not parse the condition.
- Use `failure` when the goal is impossible or the budget is exhausted. Do not
  manufacture success to escape a loop.
- Set `max_visits` on every node reachable through a back-edge unless another
  explicit bound covers the whole cycle.

## Checks and review

Use checks in increasing cost order:

1. syntax, formatting, schema, or file-presence checks;
2. focused unit or reproduction test;
3. broader test/build/lint command;
4. independent semantic or product review;
5. live user-entry-point validation.

Not every task needs all five. The last required check should correspond to the
user's actual claim. “Tests pass” does not prove a UI flow works; “reviewer likes
it” does not prove it compiles.

## Failure feedback

The next attempt must be able to see why the prior one failed. The simplest
pattern is for the tool command to write its output to a known file while
preserving the real exit status, and for the worker prompt to read that file if
present. If the same evidence repeats, change the approach or exit; another
identical lap is not progress.

## Human judgment and external steering

Pause for a person when the next action requires authority, taste, or a scope
choice. Do not add human approval where a deterministic check answers the
question. External steering is useful for new authoritative information or
clear drift, not routine narration.

## Parallelism

Use `parallel` only when branches can work independently in isolated worktrees.
Typical cases are alternative implementations, separate research questions, or
independent review lenses. The `parallel.fan_in` prompt must inspect branch
artifacts and state how to select, merge, or summarize them. If one branch needs
another branch's result, keep them sequential.

## A small repertoire is enough

The proposed skill should teach these in order:

1. implement -> check -> retry;
2. plan -> implement -> check, with selective replanning;
3. diagnose -> fix -> reproduce;
4. draft -> review -> revise;
5. parallel attempts -> compare/synthesize, only on demand.

This is intentionally smaller than the feature vocabulary found in other
Attractor implementations. The corpus repeatedly shows that useful systems
converge on these shapes even when their syntax and runtime differ.

## Disagreements worth preserving

- Amplifier argues that `plan -> implement -> test` often belongs inside one
  adaptive worker and that graphs should encode only the convergence skeleton.
  Many other implementations and practitioners deliberately separate these
  stages for durable state and independent review. Tractor guidance should use
  a boundary test, not prohibit either form.
- Ralph practice favors one fresh context per small backlog item and strong
  eventual consistency. Plan-first systems favor reviewed plans before action.
  Both work in their intended domain: Ralph for well-specified, low-risk laps;
  explicit planning for uncertain or high-leverage decisions.
- Multi-model consensus appears in several repositories, but ordinary evidence
  favors one independent reviewer plus a deterministic check. Consensus should
be an advanced pattern, not a default example.
