# Tractor for Humans & Agents — adoption design brief

Date: 2026-08-19 · Grounded in the repo at commit 472d5b4 ·
Published artifact: https://claude.ai/code/artifact/e8191720-2c11-4b64-a252-cb24ba6a355a

How to make a pipeline engine that top-level agents reach for unprompted, and
that a teammate understands in sixty seconds — and why those are two different
products sharing one binary.

## 01 · The core insight: two customers with opposite reading habits

Tractor has two customers, and they discover software in opposite ways.
**Humans** convert on one visceral demo plus a mental model they can draw on a
whiteboard. **Agents** never browse — they act on whatever is in their context
window at the moment an orchestration-shaped task arrives, which means their
entire product surface is a few kilobytes of skill descriptions, tool schemas,
and status output.

> **Thesis: docs are not the product surface for agents. The binary is.**
> Tool descriptions, embedded workflows, lint messages, and status output *are*
> the agent documentation — ship them compiled in.

The two paths meet at one point, and the current site already nails it: the
landing page's call to action is a *prompt the human copies and hands to their
agent*. That handoff is Tractor's signature interaction. Everything below is
built around widening both paths into it.

- Human lane (surface: the website): landing page → killer demo (Critique
  Circle, real run) → mental model (5 node types, one diagram) → copies the
  install prompt.
- Agent lane (surface: the binary): skill description matches the moment →
  named workflow, zero authoring → custom pipeline via schema + teaching lint.
- The copyable prompt is the bridge: the human path ends by pasting it; the
  agent path begins there.

## 02 · The agent funnel: what has to be true for unprompted use

"My top-level agents use it without being told" decomposes into four
conditions, each implemented by a specific artifact. Miss any one and the
funnel breaks — a perfect skill description with no zero-authoring entry point
still forces the agent to learn the graph language mid-task, and it will fall
back to spawning subagents instead.

| Stage | Condition | Implemented by |
|---|---|---|
| 1 · Trigger | Something in ambient context matches the task *shape*, not the feature list | Skill + MCP descriptions |
| 2 · First run | Value in the very first tool call — no schema, no authoring, no reading | Embedded named workflows |
| 3 · Legible progress | Status output the agent can narrate to its human without reading files | `get_run_status` as UX copy |
| 4 · Graduation | Authors custom graphs; lint errors teach instead of merely rejecting | Schema + 27 teaching lint rules |

Stages 1–3 are where the work is; stage 4 mostly exists already (the schema is
generated from the live Go type and the linter has 27 named rules).

### Stage 1 — write descriptions around moments, not mechanics

The trigger phrases that matter are the ones a human actually says to a
top-level agent: *"try a couple of approaches in parallel," "have another model
check this," "keep working on this after I close my laptop," "don't stop until
the tests pass."*

Current MCP server instructions (describe mechanics):

> "Pipeline definitions are files. Read the current schema only when authoring
> or changing a pipeline, validate before starting, and use the returned
> run_id for later operations."

Proposed (sells the moment):

> "Use Tractor when work should fan out across several models or approaches,
> be cross-checked by another model, pass a deterministic verification gate,
> or keep running after this session ends. Start with a named workflow — no
> pipeline authoring needed. Runs are detached processes; the run_id works
> across reconnects."

The same engineering applies to the Claude Code skill description — arguably
the single highest-leverage string in the whole project, because it is the
only Tractor text guaranteed to be in a top-level agent's context when the
task arrives.

### Stage 2 — named workflows compiled into the binary

The agent's first call should be
`start_run(workflow: "critique-circle", args: {task, artifact, providers})` —
a parameterized pipeline embedded via `go:embed`, validated at build time,
listed by a new `list_workflows` tool. This keeps the "schema and MCP surface
compile into one binary and cannot drift" guarantee, and it means the graph
language is something agents *graduate into*, not a prerequisite.

### Stage 3 — status output is UX copy

`get_run_status` already returns the right fields. The design move is treating
its text as narration: an agent should be able to relay `current_node` /
`last_stage` / `last_response` to a human verbatim and have it read as a
progress update, not a debug dump. That single property makes agents look good
for having chosen Tractor — which is what makes them choose it again.

## 03 · Named patterns: a short catalog of shapes covers almost everything

Patterns spread when they have names. "The tri-model artifact creation and
mutual critique possibility" is a paragraph; **Critique Circle** is a thing a
teammate can ask for. Each named pattern is simultaneously an embedded
workflow, a gallery page, and one diagram — the same asset serving both lanes.

- **Critique Circle** — N providers each produce an artifact; each then
  critiques the others' work; optional synthesis. Multi-model checks and
  balances in one node. (Unlocked by PR #18; the poem demo in
  `ephemeral/projects/tractor/poem-critique-demo/` is the reference and should
  be promoted to `examples/critique-circle`.)
- **Bake-off** — several branches implement the same task in isolated Git
  worktrees; a fan-in judge compares the declared artifacts and picks or
  merges. (parallel + fan_in + artifacts)
- **Verify Loop** — implement, then a deterministic `tool` gate routes by exit
  code — bounce back on failure, capped by `max_visits`. The gate demonstrates
  the *claim* (run the app, curl the endpoint, assert the artifact), not just
  the test suite: tests are required checks, not proof.
  (codergen + tool exit-code routing)
- **Coached Run** — a worker plus an advisory supervisor that watches live
  activity and steers into the active turn — catching thrash and drift no
  single visit can see. (supervisor node · steering socket)
- **Pathfinder** — *choose-your-own-adventure agile.* Give it a far-off
  milestone plus its observable claims. Each lap, a navigator codergen looks
  at the milestone and the repository *as it is now* and picks the smallest
  sensible next verifiable step — a vertical slice, exercisable early, never
  a horizontal stack-order plan. An implementer takes the slice; a proof gate
  demonstrates its claim; the navigator routes back for another lap, to
  `success` when the milestone claims are demonstrated, or to a `blocked`
  terminal that files a report (escalation is cheap and penalty-free). The
  slice specs the navigator writes land in stage evidence — ephemeral by
  construction, exactly where the lifetime rule wants them.
  (navigator codergen with routing + implementer + tool gate + max_visits)
- **Ladder** — *plan-first decomposition.* A planning codergen first turns
  the milestone into an ordered set of rungs — verifiable steps with claims,
  prose only, no code — which a human (or a spec-review pass) can approve
  before any implementation tokens burn. Then the loop climbs one rung at a
  time: implement → proof gate → check off → next rung. Use Ladder when the
  decomposition itself needs review up front; default to Pathfinder
  otherwise — the agents-repo wisdom is explicit that slice specs should be
  "derived fresh from the code as it is today," and upfront answers to
  questions no slice has hit yet are speculation.
- **Consensus Loop** — implement on one provider, then an independent
  reviewer on a *different* provider whose prompt names only the target —
  never expected findings, safe areas, or desired verdicts. Fix legitimate
  findings, challenge overstated ones with evidence, and re-review the
  *entire* target each round (never "confirm my fix"). Rounds append
  immutable review artifacts in the run directory; `max_visits` bounds the
  loop; unresolved dissent routes to a HITL terminal. Heterogeneous harnesses
  make the reviewer's independence real — a different lab, not a different
  system prompt.

Critique Circle is the landing-page hero: one `parallel` node, three
providers, three real harness CLIs, mutual critique with declared artifacts.
Five minutes of runtime and the entire differentiator story in a single
picture — and the actual poems and critiques from a real run make the demo
tangible in a way no feature list can. Pathfinder and Consensus Loop are the
workhorses — they're what a teammate points at a real backlog.

## 03b · Bake in the wisdom

The skills in tylergannon/agents encode a working engineering philosophy —
proof-of-work, vertical slicing, non-pedantic specs with no code, adversarial
review with reviewer-owned scope, loops with explicit halt policies. Today it
lives in prompts and manual discipline; Tractor can enforce most of it
*structurally*. The full distillation with sources and the
principle-by-principle mapping is in [wisdom.md](wisdom.md). The headlines:

- **The run directory is the proof artifact.** "Do not claim complete proof
  for evidence reviewers cannot inspect" — Tractor already satisfies this by
  construction; the product should say so.
- **Loop anatomy maps 1:1.** The write-prompts loop reference says every
  agent loop needs a work source, context builder, runner, verifier, state
  store, halt policy, budget guard, and escalation path — "or the prompt
  grows word-salad patches." Tractor gives each one a name in a typed graph.
  That table is the sharpest "why Tractor" pitch for engineers who already
  run ad-hoc loops.
- **Ephemeral middle layer, free.** Slice specs written mid-run live in stage
  evidence — archived with the run, never durable in the repo. "A wrong
  signpost is strictly worse than none"; Tractor makes the lifetime rule
  structural instead of disciplinary.
- **Escalation as a first-class outcome.** Embedded workflows route
  "stop and report" to a `blocked` terminal with a report artifact — distinct
  from failure, cheap and penalty-free, because "improvisation at a seam is a
  defect in itself."
- **Every compiled-in prompt passes the write-prompts bar.** Compact,
  positive, assumes the model is smart, no universal checklists ("a checklist
  a weak model can satisfy while doing the wrong thing").

## 04 · The human pitch: one message per audience

Tractor's honest competitive position rests on three claims nothing adjacent
can make. Everything else — YAML sugar, the closed schema, detached runs — is
supporting cast.

- **Heterogeneous by construction.** One graph drives Codex, Claude Code, and
  Gemini through their *real* CLIs with native sessions — not API calls
  pretending to be agents. No single-vendor orchestrator can offer mutual
  critique across labs.
- **Steerable while it runs.** Text injected into a live, mid-turn agent — by
  you or by an in-graph supervisor that watches for thrash and drift. Temporal
  signals a workflow; nothing else steers a running LLM turn.
- **Evidence on disk.** Every prompt, response, routing decision, steering
  message, and collected artifact lands in a browsable run directory. The
  audit trail isn't a feature; it's the substrate.

| Audience | They ask | Lead with | Prove it with |
|---|---|---|---|
| Engineer | "Why not just subagents / LangGraph?" | The three claims above, plus an honest "when not to use Tractor" (the archive inventory already wrote it). | A real run directory, opened live. Self-verifying examples that assert their own claims. |
| Eng manager | "Why should I trust agent output?" | Checks and balances: models critique each other, deterministic gates block hand-waving, every run leaves an audit trail. | Critique Circle output + a `stages/` walkthrough: "here's why the agent did what it did." |
| Top-level agent | (never asks — pattern-matches) | "When work should fan out, be cross-checked, pass a gate, or outlive this session — start a Tractor run." | A first tool call that succeeds without reading anything. |

## 05 · The website: the demo is the doc

Every Tractor example is self-verifying and every run leaves real evidence —
so the site should never show a screenshot of markdown when it can show an
actual run. Five surfaces, each visual-first, and every page carries the
signature affordance: not "copy this command" but **"copy this prompt for
your agent."**

- **Landing** — animated Critique Circle, real poems + critiques, copyable
  agent prompt.
- **How it works** — one diagram-led page: 5 node types, the walk, run
  directory, steering.
- **Pattern gallery** — card per named pattern: diagram + pipeline file +
  real run evidence.
- **Reference** — hosted JSON Schema URL, CLI, MCP tools, lint rules,
  spec.md rendered on-site.
- **Tractor vs X** — subagents · LangGraph · Temporal · honest, including
  when not to use.
- **llms.txt / llms-full.txt** — agents fetch, never browse: install, tools,
  workflow names; mirrors what the binary serves.

What survives from today's site: the visual identity (it's distinctive —
keep it), the copyable-prompt hero (it *is* the thesis),
`authoring-pipelines.md` (becomes the graduation path under Reference), and
the CI check that asserts rendered strings so docs can't rot.

## 06 · Build order, sequenced by leverage

1. **Ship the Claude Code plugin + skill** (worktree already exists), with the
   skill description engineered from the trigger-phrase list. Distribution
   parity with Codex; the skill description is the highest-leverage string in
   the project.
2. **Embedded named workflows** (`go:embed`, build-time validated) +
   `list_workflows` + workflow params on `start_run`. Promote the poem demo to
   `examples/critique-circle`. First wave: Critique Circle, Verify Loop,
   Pathfinder, Consensus Loop (the workhorses); Bake-off, Ladder, Coached Run
   follow. Creates the zero-authoring first run — the funnel stage that
   currently doesn't exist at all. Every compiled-in prompt is reviewed
   against the write-prompts bar (see wisdom.md, principle 14).
3. **Rewrite MCP server instructions + all six tool descriptions** around
   moments; audit `get_run_status` text as narration. Pure copy work on
   stages 1 and 3 of the funnel; no new machinery.
4. **Host the JSON Schema at a stable URL; fix the 0.3.0 / 0.4.0 version
   drift** (`cmd/tractor/mcp.go` vs `.codex-plugin/plugin.json`). Cheap,
   permanent wins: `$schema` editor completion, coherent version story.
5. **Site rebuild per section 05**; add `llms-full.txt`; render `spec.md`
   on-site. Depends on the named patterns and real Critique Circle run
   existing first — the site shows them, it doesn't invent them.
6. **Teaching lint**: each of the 27 rules gets a message that says what to do
   and links to a hosted anchor. Completes the graduation stage; turns
   validation errors into the graph-language tutorial.

### Definition of done

- A fresh top-level agent given *"have a couple of models cross-check this
  design"* starts a Tractor run without Tractor being mentioned.
- A teammate can explain what Tractor is — and draw Critique Circle — after
  sixty seconds on the landing page.
- Every named pattern on the site links to a real run's evidence, not a
  mockup.
