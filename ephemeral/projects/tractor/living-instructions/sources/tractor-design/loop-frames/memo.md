# Loops, frames, and the second level of Tractor

**Status:** complete design record as of 2026-08-26. Research done, no
implementation started. This document is the durable record of a long design
conversation and is meant to be re-read as the work proceeds.

**How to read this.** The memo was written in conversation order and preserves
its own history: early sections state positions that later evidence revised.
The current design is **§9 (final synthesis)** as amended by everything from
**§9a onward** — side-chat doctrine (§9a), runtime library (§10), CHA/OS
crossover and the ABI reframing (§11), the typed ABI and its corrections
(§12–§13a), positioning (§14), the substrate family (§15–§16), hygiene and
evidence-calibration rulings (§16a–§16b), and Go mechanics with the
isolation ladder (§17–§17a). Later sections amend earlier ones; when in
doubt the highest-numbered section wins. Sections 2–4 are pre-evidence
sketches — rationale, not final word (§2's frozen-items model is
superseded). Full verbatim research reports live in `research/`:
`workflow-engines-and-pl-theory.md`, `attractor-lineage-and-agent-frameworks.md`,
`humanlayer.md`. Cross-post: the locked-in kernel/substrate doctrine also
lives in `~/src/chaios/AGENTS.md` ("Shared doctrine with Tractor").

**Decision log (the conversation arc, in order):**

1. **Initial menu (pre-research).** First position rejected restoring the
   upstream context object and proposed only a deterministic history
   preamble; it dismissed declared iterators as "declarations that can lie."
2. **Tyler's correction — the steelman that reframed everything.** An
   engine-owned iterator is ground truth about *assignment*; the frame is
   what makes a lying implementation *detectable*, because validation needs
   stated intent. Led to the deixis doctrine (§0) and the fan-in precedent
   (§4.7 of the spec) as the in-spec license for injection.
3. **Deep research** (three subagents; reports in `research/`). Headlines:
   the two iterator types are the standard WCP-14/WCP-21 taxonomy; Step
   Functions' 2024 Variables retrofit is the single best precedent and its
   scoping rules are adopted wholesale; nobody anywhere injects engine-owned
   loop-frame context into prompts (unoccupied territory); upstream
   Attractor is StrongDM's, not Yegge's, and its cut context object shares
   nothing important with frames.
4. **HumanLayer revision.** They never inject iterator state — loop state
   is re-derived each lap from a mutable store (tickets, checkboxes,
   file existence). Adopted: type-1 items live in a workspace checklist
   the loop node re-reads and selects from on every arrival; engine owns
   parsing/selection/injection, never storage. Supersedes §2's
   freeze-at-activation model.
5. **Side chat (§9a).** Builder library is a compiler, not a runtime;
   restart-over-resume (wrong-goal, not crashes, is the real enemy);
   frames are ephemeral — compute, inject, forget; frames.jsonl dropped;
   goal gates resurrected.
6. **Runtime library reassessed (§10).** Restart-over-resume dissolves the
   Temporal-tax objection; one builder API with `Emit()` (YAML artifact)
   and `Run()` (composes engine runs — never a second interpreter of loop
   semantics).
7. **CHA/OS crossover (§11).** Tractor's durable contribution re-centered
   from the graph to the **agent calling convention (ABI)**; chaios
   independently ratified ephemeral frames and handles-not-text; bake-out =
   partial evaluation; graph = the detached/batch host of the ABI, embedded
   runtimes = the resident host.
8. **Typed ABI (§12–§13a).** go-gen-jsonschema types both call directions;
   call species fn / infer[T] / agent; "routing" retired for judgments and
   ordinary branching (§13); Tyler's correction: the load-bearing idea is
   the *typed return value entering the program's expression economy* (not
   frame-as-struct), which is the precondition for signature-preserving
   bake-out.
9. **Positioning (§14).** vs LangChain/LangGraph: overlap is plumbing
   (steal it); novelty is the doctrine stack (frames, no data plane,
   restart-by-judgment, coding agents as callees, bake-out teleology).
10. **Substrate family (§15–§16).** Authoritative: repo, data store,
    secret store; derived: process tree, frames, artifacts. Self-building
    = writes to authoritative substrates + re-derivation of derived ones
    (reconciliation, not command). Fixed kernel = "the residue of the
    interpreter that never specializes away" — **locked in**, cross-posted
    to chaios AGENTS.md. Secrets never flow through calls, frames, or
    evidence.
11. **Hygiene and calibration (§16a–§16b).** Prior art citable only as
    bounded one-sentence comparisons, never adoption frames. BEAM
    considered and rejected (model fluency, type enforcement, wrong half
    of OTP). Evidence vocabulary: "demonstrated" (n≈1, conditions stated)
    vs "proven" (full acceptance incl. human-visible leg); chaios claims
    recalibrated accordingly — its Chrome-operated acceptance milestone
    remains unchecked.
12. **Go mechanics (§17–§17a).** No code injection into running Go
    processes (`buildmode=plugin` prohibited); re-derive instead, at
    whole-process or child-process granularity; wazero/wasm as the
    in-process option if ever needed (imports = scoped syscalls).
    Isolation ladder: in-binary function → wasm instance → child process →
    browser realm. Tree width is a phase indicator: hardening shrinks the
    program toward one process; wide fan-out belongs to the factory, never
    the program.

**Glossary of coined terms** (used throughout, defined where introduced):
*deixis/deictic prompts* (§0) — prompts whose referents ("the current
work") the engine must resolve; *frame* — the engine-composed, ephemeral
rendering of iteration state injected into a prompt; *handles-not-text*
(§10) — deterministic code never branches on agent prose; *sensor/selector*
(§8.2) — deterministic re-derivation of loop position from the workspace;
*bake-out* (§11) — replacing agentic behavior with deterministic code;
*the ABI* (§11) — frames = arguments, choice schema = return type, `$goal`
= argument register, workspace = heap, budgets = fuel, run dir = trace,
runtime CLI = syscalls; *judgment* (§13) — a prose predicate evaluated by
inference, returning a closed value; its evaluator is the *judge*, and the
choice schema is its answer type. **"Routing" is retired** (§13): it
survives only as IR-level vocabulary for the compiled graph, like jumps in
assembly — at the source level it is ordinary branching on judged values.
Later sections predating §13 still say "routing"; read it as branching.
*fixed kernel* (§16, **locked in by Tyler and cross-posted to chaios
AGENTS.md**) — "the kernel is the residue of the interpreter that never
specializes away": loader, supervisor, syscall dispatcher, secret
resolver; everything else is malleable body.

**Companion docs:** PR #27 prompt doctrine (merged; `skills/tractor/SKILL.md`
"Writing node prompts"); `docs/attractor-archive-inventory.md`;
`docs/spec.md` §§3.3–3.4, 4.6–4.7, 5.4–5.6; `~/src/chaios`
`ephemeral/projects/application-design/` (esp. `application-document.md`,
`decisions-to-date.md`).

## 0. The corrected premise

Previous position ("declarations that can lie") was wrong, and the correction
matters because it is the whole argument:

An engine-owned iterator is ground truth about **assignment**. "We are on
chunk X" cannot be false the way an agent's self-report can, because the
engine made the assignment and holds the counter. When the implementer does
the wrong work, the declared frame is not the lie — it is the instrument that
makes the lie *detectable*, because a validator can only judge work against a
stated intent. Without an authoritative frame, "validate the current work"
sends the validator to reconstruct intent from the implementer's own notes —
the accused writes the indictment.

Doctrinal form: **dead-simple prompts are deictic** ("this round", "the
current work", "the next chunk"). Deixis needs a frame of reference. Today
Tractor resolves two indexicals by engine authority: `$goal` (what the run is
for) and the routing schema (what the choices are). PR #27 proved that every
indexical the engine resolves is a paragraph the author doesn't write. The
proposal on the table is the third indexical: *what the current work is*.

The spec already contains the precedent in ratified form. §4.7: the fan-in is
a codergen node "with the branch evidence appended to its prompt," and its
default prompt when empty is simply "Evaluate the results of the parallel
branches." — a two-word-class prompt made viable by deterministic engine
injection of engine-collected evidence. Parallel/fan-in is spatial fan-out
with an engine-rendered frame. A loop is temporal fan-out. The symmetry is
exact, and it means frame injection is not a new *kind* of thing in the spec;
it is the existing thing, applied to iteration.

The line that keeps this honest, and keeps faith with the razor that cut the
upstream context object:

> The engine may inject **facts it owns** (assignments it made, counters it
> keeps, evidence it collected, commits it snapshotted). The engine never
> **interprets prose** (conditions, notes, verdicts stay with choosers).

Everything below is designed on that line.

## 1. Tensions with the ratified spec (named, not dodged)

1. **§3.4: "The budget is engine bookkeeping only: handlers are not told the
   visit number."** Rationale given: the looping node's session already
   carries its history under default `compacted` fidelity. Two answers:
   (a) the shipped examples default to `fidelity: none`, where that rationale
   is void — a fresh session knows *nothing* and the preamble is the only
   orientation it gets; (b) the principled distinction is bookkeeping vs.
   semantics: `max_visits`/`node_visits` are traversal safety rails and should
   stay hidden, but a *declared iterator's* index and item are author-visible
   meaning, not bookkeeping. Amend §3.4's sentence to draw exactly that line.
2. **§5.5's three-surface data model.** Frames are a fourth surface only if
   they carry agent-authored data. Designed as below (items frozen from a
   workspace file; sub-goal captured through the choice schema like `next`;
   commits snapshotted by the engine), every frame datum is either already on
   an existing surface or is engine state of the same species as
   `node_visits`. The workspace remains the only data plane agents write
   freely.
3. **"Rich outcome-directed routing: deliberate exclusion."** Untouched.
   Nothing below lets the engine read a verdict or compare condition text.
   The one mechanical chooser added (the loop node advancing its index) is
   the same species as the tool node's exit-code chooser — §3.3's table gains
   a row, not an exception.

## 2. Iterator type 1: the array loop (`foreach`)

> **Superseded in part.** This section's frozen-items-in-checkpoint model
> was revised after the HumanLayer evidence: items now live in a workspace
> checklist file the loop node re-reads and selects from on every arrival
> (see §8.2 "Reconciliation" and §9 point 2). The region rules, chooser
> analysis, and trust-boundary argument below still stand.

For work where a discrete step already produced the list: migrate these 40
files, address these 6 review findings, run these conformance scenarios,
apply the codemod per package. The "stale plan" objection from the earlier
round is real only for *exploratory* work; mechanical fan-out over a known
list is common, valuable, and is what parallel already is — minus isolation,
minus concurrency.

Sketch:

```yaml
- id: plan
  type: codergen
  prompt: >
    $goal

    Break the work into independently demonstrable chunks and write them
    as a JSON array of strings to .tractor/chunks.json.
  edges: [{to: milestones}]

- id: milestones
  type: loop                       # new node type
  items_file: .tractor/chunks.json # read once at loop activation, frozen
  body: implement                  # body entry
  max_laps_per_item: 4             # optional; budget vocabulary TBD
  edges:
    - to: success                  # offered only when items are exhausted

- id: implement
  type: codergen
  prompt: Do the work for the current chunk.
  edges: [{to: check}]

- id: check
  type: tool
  tool_command: ./validate-current-chunk.sh   # or a codergen validator
  on_success: milestones           # lap ends: engine advances or re-offers
  on_error: implement
```

Semantics, each point chosen to reuse ratified machinery:

- **Items** come from a workspace file a prior node wrote. The engine
  validates shape only (JSON array of strings; violation = categorized
  Error), never meaning — the same trust boundary as a tool node's exit
  code, which is also agent-influenced world state read mechanically. Items
  are **prose sub-goals**, not structured objects (v1). A static `items:`
  list in YAML is also allowed for authored matrices.
- **Frozen at activation**, stored in the checkpoint. "The current chunk is
  X" is thereby stable ground truth for the whole activation — crash-safe,
  resume-safe. Re-planning is exiting the loop and re-entering it (a fresh
  activation re-reads the file); mid-activation replan is deliberately not
  v1.
- **The loop node's chooser is the iterator.** Arriving at the loop node
  with items remaining offers the body; exhausted offers the exit edges.
  This is mechanical state the engine owns — §3.3's chooser table gains:
  `loop` | the iterator | next item pending → body; exhausted → exit edges.
  No prose is interpreted. Completion-of-a-chunk judgment stays inside the
  body (the `check` node routes back to the loop = "this chunk is done" —
  or loops to `implement` = "not done"). Wait — correction: routing back to
  the loop node must distinguish "chunk done, advance" from "give up on
  chunk". V1 rule: any arrival at the loop node advances the iterator;
  abandoning a chunk without advancing is expressed by edges that bypass
  the loop node (escalation), same as budget-exhaustion routing today.
- **Body region** is delimited exactly like a branch node-set (§4.6): nodes
  reachable from `body` without passing through the loop node. The
  disjointness/entry lints transfer nearly verbatim (`loop_body_entry`,
  `loop_body_disjoint`). Nesting of loops is allowed (sequential — none of
  parallel's worktree/race constraints apply).
- **Budgets**: `max_visits` still bounds every node. The loop adds a
  per-item lap budget if wanted, but v1 can live on `max_visits` alone.
- **Checkpoint** grows a `loops` map: activation id → items, index, lap
  counters, per-lap HEAD snapshots. Resume is unchanged in kind.

## 3. Iterator type 2: the condition loop with a frame slot

Milestone-loop already is this shape; what is missing is only that STEP.md is
a per-example convention with no owner, no lifetime, no injection, and no
history. The upgrade:

- A loop node with no `items_file` is a **condition loop**: every arrival
  re-offers body and exit edges, and *the chooser is whoever routes into
  it* — i.e., exactly today's cycle semantics, now with a scope attached.
  (Alternative: the loop node itself is a codergen chooser with edge
  conditions. Both compose; v1 can make the bare loop node routing-neutral.)
- Any codergen node in the body may declare `sets: frame` (name TBD —
  `frame`, `sub_goal`). Its **choice schema** gains a required `frame`
  string property, described as "the sub-goal for this lap: the next
  functionally verifiable step toward the loop's purpose." This is the
  identical mechanism by which `next` already reaches the agent —
  engine-owned schema extension, structured output, no prose parsing. The
  engine stores the value in the loop scope and appends
  `{value, ts, head}` to `frames.jsonl` in the run directory: the stack's
  history, permanently auditable, one line per push.
- **The engine snapshots HEAD when a frame is set.** This is the quiet big
  win: "the current work" acquires a *diff*. A validator's preamble can say
  `sub-goal: <text>; work began at <commit>` — and `git diff <commit>..HEAD`
  is precisely "the current work," resolved by git, no self-report involved.
- Frame lifetime = loop activation. Exit pops the scope; inner frames never
  leak into outer prompts. History persists in the run dir (the popped
  frame is disposed from *scope*, not from *evidence* — a core dump is not
  a stack).

The example rewritten — every prompt at doctrine strength:

```yaml
- id: milestones
  type: loop
  purpose: Implement the plan, one verifiable step at a time.
  body: pick

- id: pick
  type: codergen
  sets: frame
  prompt: Choose the next functionally verifiable step toward $goal.
  edges:
    - {to: implement, condition: A next step remains.}
    - {to: milestones, condition: Every claim in the goal is demonstrable now.}

- id: implement
  type: codergen
  prompt: Do the work for this round.
  edges: [{to: validate}]

- id: validate
  type: codergen
  prompt: Validate the current work.
  edges:
    - {to: milestones, condition: The step is demonstrably complete.}
    - {to: implement, condition: The step is not complete.}
```

`pick` no longer says "record it as the only content of STEP.md" — the
mechanism ate the convention.

## 4. Injection: what the frame block actually contains

Rendered per active scope, innermost last, prepended (or appended, matching
fan-in) to every body codergen prompt. Everything in it is engine-owned:

```
<tractor>
goal: <run goal>
loop "milestones" — purpose: implement the plan (chunk 3/7)
  chunk: Wire the submit button to a working remote function
  chunk began at commit ad23hfa2 · lap 2
  loop "hardening" — purpose: make the step demonstrable
    sub-goal: make the failing submit-button integration test pass
    sub-goal set at commit 9f31c02
</tractor>
do the work for this round
```

Deliberately absent, v1: previous laps' notes, validator verdicts, tool
output tails. History is what sessions are for (`compacted` fidelity), and
what the run directory serves observers. The frame is the whiteboard of
record — assignments, not narrative. This keeps the block small, keeps
§3.4's razor mostly intact, and keeps the frame incapable of accumulating
prose that would need interpreting. (Fresh-session bodies that want the last
verdict can get it the honest way: the validator routes with notes, and the
next lap's `pick` sets a frame informed by its session memory.)

Frames also become a **steering surface** for free: `steer_run` rewriting
the current frame ("skip chunk 4; it's obsolete") is a precise, auditable
intervention — far better than injecting freeform steering text mid-session.

## 5. The procedural language, taken seriously

The claim to test: a small structured language with declared variables and
lexically scoped loops, instead of (or over) the graph.

The theory says the idea is *right*, not fucked up — with one inversion. A
graph with arbitrary cycles is flowchart-era control flow: goto. Dijkstra's
argument for structured programming was never aesthetics; it was that
**lexical structure is what makes program state comprehensible at a glance**
— scope and stack fall out of nesting for free. That is exactly why frames
feel awkward to bolt onto the graph and felt natural the moment the loop was
written as pseudocode. Every mature workflow system has faced this and the
outcome is consistent: either they went code-first (Temporal), or their
graph language grew a variables-and-scopes bolt-on years in (Step
Functions). [Agent 3 findings to confirm/cite.]

The compilation asymmetry decides the architecture: structured control flow
lowers to a CFG trivially; lifting an arbitrary CFG to structured form is
the hard Relooper problem. Therefore:

- **The graph stays the IR** — the executable, checkpointed, steerable,
  lintable substrate. Nothing about the engine changes for the DSL's sake.
- **The DSL is a front-end** that compiles to it. Tyler's example lowers
  almost embarrassingly directly:

```
declare the_plan
until agent_check("is the plan good? why or why not"):
    the_plan = agent("develop a plan for $the_goal")
```

→ condition loop node (scope of `the_plan`); `agent(...)` → codergen node;
assignment → `sets: the_plan` via the choice-schema mechanism; `agent_check`
→ a cheap codergen chooser with two conditioned edges — which is literally
§3.3's own advice ("when a decision point needs judgment, the author places
a cheap LLM node there"). The graph vocabulary is already the DSL's
instruction set: codergen = call, tool = syscall, edges = branches,
max_visits = fuel, frames = activation records.

**The decisive consequence: the runtime features are the prerequisite, the
language is sugar.** Building loop nodes + frames now is simultaneously the
useful standalone feature and the entire runtime the DSL would need later.
The DSL can be decided in six months with zero regret either way.

What keeps the DSL from becoming the chaos scenario, if built:

- Variables hold **prose**, set only by agents (`sets:`), consumed only by
  interpolation into prompts and by agent/tool judgment. No comparisons, no
  arithmetic, no indexing, no string ops. The moment `if $x == "yes"`
  exists, the engine is interpreting prose and the language is a bad Python.
- Control constructs: `foreach`, `until/while` (agent- or tool-checked),
  sequence, maybe `parallel`. Nothing else. No user-defined functions in
  v1 — that is what child pipelines are for, where `$goal` is already the
  argument register.
- Two-surface risk (YAML graphs vs DSL diverging) is real. Mitigation is
  positioning: DSL compiles to a plain graph file you can read, check in,
  and run — HCL-to-JSON, not a parallel dialect.

Honest assessment: worth wanting, wrong thing to build *first*. If the loop
+ frame layer lands and authors (human or agent) still reach for pseudocode,
that revealed preference is the go signal, and the compiler is then small.

## 6. The menu, re-scored

- **Loop node + scoped frames + injection (types 1 and 2, §§2–4)** — the
  recommendation. It is the fan-in precedent applied to time; it makes the
  doctrine's deictic prompts actually resolvable; it deletes per-example
  conventions (STEP.md) rather than adding surface; every datum injected is
  engine-owned ground truth. Cost: a new node type, region lints (machinery
  exists), checkpoint growth, one choice-schema extension.
- **Deterministic history preamble (arrival edge, last notes, visit
  numbers)** — demoted from earlier recommendation. §3.4 ratified against
  it for reasons that hold *when frames exist*: assignments belong in the
  frame; narrative belongs to sessions and the run dir. Revisit only if
  fidelity-none bodies still flounder with frames in place.
- **Workspace-file conventions only (STEP.md doctrine, no engine change)** —
  the do-nothing fallback. Cheap, already proven by milestone-loop; but no
  scope, no nesting, no injection, no history, no steering surface, and the
  prompt must re-teach the convention in every graph. Keep as the escape
  hatch, not the plan.
- **Unstructured context object / `context_updates` (upstream restoration)**
  — still rejected, now for a sharper reason: it is frames without owners,
  lifetimes, or ground truth — global mutable variables. The frame design
  delivers the 20% of it that was load-bearing. [Agent 2 to confirm how it
  behaved in the wild.]
- **Procedural DSL** — deferred, deliberately unblocked-for-later by the
  loop/frame runtime (§5).
- **Child pipelines as stack frames** — unchanged: the function-call story
  for coarse-grained composition; `$goal` is the argument register. Compose
  with loops later; prototype today via a tool node running `tractor run`.

## 7. Open questions (annotated with later resolutions)

- Loop-node routing on re-arrival: always-advance vs. explicit
  advance/repeat. **Resolved by §8.2/§9:** dissolved — the checklist file
  carries per-item status; "advance" = an agent marking the item done as
  ordinary workspace work; the loop node just re-selects.
- Frame property name; whether `sets:` may live outside any loop
  (run-scoped frame = global variable). **Open.** Lean lint-warn outside
  loops.
- Injection position: prepend vs append; block format; one block vs
  nested. **Open** (§4 sketches nested-innermost-last, prepended).
- `skip` semantics on the array loop. **Resolved by §8.2:** steering is an
  edit to the checklist file; no engine skip needed.
- Loop node `purpose:` prose vs reusing `prompt:`. **Open;** leaning
  `purpose:` (injected verbatim, never interpreted).
- **Added later:** checklist file format (markdown checkboxes vs JSON with
  status) and its validation contract; futility-detector thresholds; goal
  gate declaration shape (see §9a.2); which checkpoint machinery to
  forfeit under restart-over-resume (§9a.3); ABI extraction as a
  standalone nlspec section (§11).

## 8. Prior-art integration

### 8.1 Workflow engines + PL theory (agent report received)

Full report with citations: `research/workflow-engines-and-pl-theory.md`.
Distilled here.

**The taxonomy already exists and matches the two iterator types exactly.**
van der Aalst's Workflow Patterns separate *Multiple Instance* patterns from
*Structured Loop*: WCP-14 is "MI with a priori run-time knowledge" — count
discovered at runtime but **fixed at loop entry**, instances synchronized at
completion — which is precisely iterator type 1 with items frozen at
activation. WCP-21 is the structured pre/post-test loop with "a single entry
and exit point" and one body instance at a time — iterator type 2. WCP-10
(Arbitrary Cycles, multiple entries/exits) is what Tractor's raw graph
permits today. No mature system unifies MI and until-loops into one
construct; the two-type design is the industry-standard decomposition, not
an invention. (Also instructive: WCP-15 — instances addable mid-flight — is
the hard pattern most engines refuse; deferring mid-activation replan is the
mainstream choice.)

**The convergent foreach design (Camunda 8 MI + Step Functions Map + Argo
`withParam`) has five fields, and §2's sketch already matches four:**
1. Body = delimited sub-workflow, single entry/exit, run as a child scope.
2. Item source = expression/file **evaluated at body activation** (Camunda
   `inputCollection`; SF `ItemsPath`; Argo `withParam` fed by a previous
   step's output — "generate the JSON in another step" is literally the
   planner-writes-chunks.json pattern).
3. Per-iteration state: engine injects current item under a declared name
   plus a loop counter (`inputElement` + `loopCounter`; `{{item}}` +
   `map_index`); siblings never see each other's state.
4. **Exit semantics: iteration state destroyed; only a declared aggregation
   survives** (`outputElement→outputCollection`, SF `Output`). §2 lacked
   this — v1 answer: the workspace is the aggregation (sequential loop, one
   workdir), so no output plumbing needed; revisit if loops ever go
   parallel-MI.
5. Early exit via `completionCondition` — note sequential MI + completion
   condition is already most of type 2, confirming the two types can share
   one node type with `items_file` presence as the discriminator.

**Scoping rules worth copying verbatim — Step Functions Variables (2024),
the single most instructive precedent.** A mature, deliberately
data-plane-minimal declarative graph language retrofitting scoped variables
onto an existing graph — exactly Tractor's situation — and its published
rules are a complete answer sheet for §7's scope questions:
- inner scopes read outer variables; siblings see nothing;
- **shadowing is a validation-time error** (adopt as a lint; kills the
  Camunda-7 shadowing bug class);
- **variables are destroyed at scope exit**; passing data out requires an
  explicit output declaration (frames pop; the workspace is the export);
- remote/isolated iterations see only their item (for Tractor: a
  fidelity-none body session physically can't see anything else anyway —
  the rule and the mechanism agree);
- the "current frame" is itself modeled as a reserved well-known variable
  (`$states`) composed by the engine — the direct precedent for the
  injected `<tractor>` block.
Also decisive motivation-evidence: SF's own official loop tutorial is a
Lambda hand-threading `index`/`continue` flags through the data plane with a
Choice back-edge — the Böhm–Jacopini folk construction performed manually —
and Variables were pitched as the cure for exactly that pathology. A graph
language without frames makes authors build the program counter by hand in
the data plane; Tractor without frames makes authors build it in STEP.md
conventions and prompt prose. Same disease, same cure.

**Until-loops stay graph cycles, but blessing the structured form is what
makes frames attachable.** Airflow forbids cycles entirely; SF and Camunda
model until-loops as explicit back-edges; Argo uses template recursion
(WCP-22 — each recursive invocation is a fresh frame whose scope is its
parameters, i.e., Argo gets stack frames by making every call a function
call; philosophically the closest system to "child pipeline with $goal as
argument register"). Tractor's cycles are fine; the loop node's job for
type 2 is purely to mark the WCP-21 single-entry/exit region so the engine
knows where a frame's lifetime begins and ends.

**The code-first tax is now quantified.** Temporal buys lexical scoping with
the determinism contract plus a 51,200-event history cap and Continue-As-New
(a manual stack-spill where only explicitly passed state survives); Inngest
caps runs at 1,000 steps with step identity tied to invocation counters.
Code-first durability meters every iteration and constrains the code.
Tractor's graph-PC gets checkpoint/resume without either. And the compile
direction is unanimous across the industry: CDK/Workflow Studio → ASL, Hera
Python → Argo YAML, Airflow's Python executing at parse time to emit a
static DAG — code lowers to graph, nothing lifts graphs to code, and the
theory (Relooper needing dominator analysis and sometimes *added runtime
state*; Kosaraju's hierarchy; Knuth's "structure completely lost") says
lifting is impossible in general. §5's architecture (graph = IR, DSL/builder
= one-way front-end, never round-trip) is confirmed as the only sound shape.

**Design deltas taken from this evidence:**
- Name the semantics by pattern: type 1 = WCP-14 (sequential MI), type 2 =
  WCP-21. Freeze-at-activation is not a simplification to apologize for; it
  is the definition of the mainstream pattern.
- Add the **no-shadowing lint** for nested loop frames.
- State exit semantics explicitly in the spec: frame destroyed at loop
  exit, workspace is the only export surface (an earlier `frames.jsonl`
  audit idea was later dropped — §9a.3).
- Consider `completion_condition`-style early exit later, not v1; the
  chooser-routed exit edges already express it agentically.
- One node type, two modes, discriminated by `items_file`/`items` presence.

### 8.2 HumanLayer (agent report received)

Four generations of their stack examined (12-factor-agents doctrine → the
CodeLayer daemon → the shipping Riptide product → the `fold`/`electric`
runtime being built now). Two findings dominate.

**Finding 1 — convergent evolution, headline-grade.** HumanLayer's shipping
workflow format is nearly isomorphic to Tractor's graph: named steps whose
prompt is a reusable skill ("do the work for this phase"), prose-titled exit
edges, loops as plain cycles, **no variables or data plane** — state moves
through the task worktree and markdown artifacts. Two teams starting from
different doctrine (their 12-factor "own your context window" vs. our
razor-and-three-surfaces) landed on the same shape independently. Their one
delta: exits are offered to a *human* by default ("every satisfied exit
remains available for the user to choose"), with per-exit `autoAdvance`
opt-in — where Tractor turns edge conditions into a structured routing
question the *agent* answers. Validation of the whole bet.

**Finding 2 — the serious counter-argument: they never inject iterator
state. Anywhere. Across four generations.** Loop state is always
externalized and **re-derived at iteration start**: ralph loops run the same
constant prompt every pass and query Linear for the top item, moving ticket
statuses as they go; `implement_plan` reads checkboxes in the plan file and
picks up "from the first unchecked item"; their control-loop skill frames
every loop as set-point/sensor/actuator where the *sensor re-measures the
gap each run*; riptide phases are chosen by which artifact files exist.
The node prompt stays constant; the first action of each round is "read the
state store." Their loop variable is real — it just lives in a mutable
status store (tickets, checkboxes, files), not in an engine counter.

**Reconciliation — this reshapes iterator type 1.** The frozen-array-plus-
hidden-index model (§2, SF/Camunda-style) and the HumanLayer re-derive model
are not actually opposed; they differ on *where the item list and cursor
live*. Frozen-in-checkpoint: crisp count, but hidden state, a second source
of truth, and replanning requires exit/re-enter. Workspace-status-file:
the items file itself carries per-item status (checkboxes), the loop node
**re-reads it on every arrival and deterministically selects the first
pending item** — engine as sensor/selector, not counter. That version:
keeps §5.5's single data plane (the file IS the truth, diffable, in git);
makes mid-flight replanning trivial (any agent may legally edit the file —
WCP-15 flexibility without engine complexity); makes steering an edit
("skip chunk 4" = check it off with a note); and survives crashes with no
checkpoint growth at all. The engine still owns what matters for deixis:
*parsing, selection, and injection* are deterministic engine acts, so
"current chunk is X (3/7, lap 2)" remains engine-rendered ground truth —
about a file everyone can see. Revised type 1: `items_file` is a spec'd
lightweight checklist format; shape violations at read time are categorized
Errors; "advance" = a body node marking the item done as ordinary workspace
work; exit offered when nothing is pending; per-item lap counts derivable
(consecutive selections of the same item).

**Also worth stealing, smaller:**
- **Doom-loop detector** (fold): halt when N consecutive tool-call batches
  are identical (stable-JSON fingerprint). An engine-side lap-futility
  check — e.g. same item selected M laps with an empty diff — is the
  Tractor analog; cheap and catches the spin-out failure mode their
  doctrine warns about ("agents spin out trying the same broken approach").
- **Backpressure**: their scheduled loops no-op while the previous output
  PR is unreviewed. Maps to loop-node gating on an unconsumed artifact.
- **Handoff documents** — a templated frame (Task/Learnings/Next Steps)
  written on suspend, resumed by path — are serialized stack frames in
  production use; validates `frames.jsonl` and suggests its schema.
- **Artifact precedence rule**, copyable verbatim into doctrine: "later
  artifacts take priority when artifacts disagree; current code remains the
  source of truth for current behavior."
- **Context withholding as a feature**: they deliberately hide the ticket
  from the research phase. Frames must stay *small assignments*, not
  accumulating briefings — §4's minimalism confirmed from the other
  direction.
- Their doctrine quote that is secretly the frame thesis: "at any given
  point your input to an LLM is 'here's what's happened so far, what's the
  next step'" — the frame block is exactly that sentence, composed by the
  engine instead of by hope.

### 8.3 Attractor lineage + agent frameworks (agent report received)

**Lineage correction:** upstream Attractor is **StrongDM's** (Justin
McCarthy et al., ~Feb 2026, spec-only NLSpec release); the Yegge attribution
is folklore — his Gas Town writing is adjacent inspiration, and his own
bibliography has no attractor entry. Chain: strongdm/attractor →
tylergannon/attractor (archived) → Tractor.

**The upstream context object, precisely (strongdm spec §5.1):** a
thread-safe flat `Map<String, Any>` shared across all stages — "the primary
mechanism for passing data between nodes." Engine and handlers both write it
(`outcome.context_updates`, plus engine built-ins like `current_node` and
`internal.retry_count.*`); edge conditions are *evaluated by the engine*
against it (`key=value`, `&&`, truthiness); namespacing (`graph.*`,
`internal.*`, `stack.*`, `work.*`…) is **convention only, unenforced**.
Crucially: **prompts could not reference context at all** upstream — `$goal`
was deliberately the only template variable; what a node "saw" of prior
state came from context-fidelity summaries. And **loops are not
context-aware**: nothing resets per iteration; the spec's own loop example
has the author inventing and maintaining a `context.loop_state` key by hand
— the same hand-threaded program counter as Step Functions' loop tutorial
(§8.1). Parallel-branch context clones are the sole scoping mechanism, and
branch deltas are never merged back.

So the thing that was cut was: a flat, unscoped, convention-namespaced
global store whose main consumer was *engine-evaluated routing*, with no
prompt access and no loop awareness. The frame design shares almost nothing
with it except the word "state" — frames are scoped, declared, loop-owned,
engine-ground-truth, prompt-injected, and never engine-interpreted. Cutting
the context object and building frames are consistent acts.

**One germ worth stealing from the archived Go implementation:** it grew two
features upstream never had — `{{context.key}}` prompt interpolation, and
**`context_outputs`: keys a node *declares*, which the engine then requires
in the agent's structured outcome**. That declaration-first shape is exactly
the `sets: frame` mechanism in §3, independently invented in the lineage.
The precedent strengthens the design: the archive's one advance over
upstream was making state *declared*, not ambient.

**Experience reports:** essentially none exist (ecosystem ~6 months old);
the most substantive critique of the context object *is this repo's own
lineage* — the spec.failed.md quarantine and the archive inventory's
"deliberate exclusion" ruling. No external retrospective on whether the KV
store worked. Softness noted in our own inventory: "structured command
context — worth considering without creating a second workflow data plane."

**Frameworks survey — the direct answer to "does anyone have the
frame/stack idea":**

- **Nobody has loop-scoped frames.** No framework anywhere has per-loop-
  iteration variable scopes, a frame stack, or scope disposal on loop exit.
  Closest analogs are entered once per *invocation*, not per iteration:
  LangGraph subgraphs (own state schema, parent marshals in/out — a
  hand-written stack frame) and Mastra nested workflows.
- **Two real "current chunk is X" mechanisms exist**, both engine-owned
  per-invocation payloads: LangGraph's `Send` API (each mapped worker gets
  its own payload outside the global schema, fan-in via reducer) and
  **Mastra's `foreach`/`dountil`** — the best contract shape in class: the
  loop primitive owns (a) per-iteration input (the item), (b) an
  engine-maintained `iterationCount` surfaced at the decision point, (c)
  prior iteration's output feeding the next. Maps onto Tractor without a
  data plane: (a) = frame block line, (b) = iterator index, (c) = workspace
  and session.
- **Prompt context is always the author's job** — history accumulation
  (LangGraph messages, OpenAI SDK), explicit templating (ADK `{var}`,
  CrewAI placeholders), or raw code. **No mainstream framework synthesizes
  a state briefing into the next node's prompt**; engine-generated context
  summaries exist only in the Attractor lineage (fidelity modes). The
  injected frame block would be genuinely differentiating: engine-owned
  deixis resolution is unoccupied territory.
- **Nobody compiles a procedural language to a graph.** Mastra is the
  nearest existence proof: a procedural-looking fluent builder whose loop
  verbs provably serialize to a graph (`serializedStepGraph`), i.e., the
  "builder library that compiles, not executes" architecture (§5) is
  tractable and half-built elsewhere, but a structured DSL with loop scopes
  lowering to a checkpointable graph would be novel.
- **Cautionary tales confirmed:** ADK's prefix-scoping and upstream's
  namespace conventions are the same unenforced-convention weakness; ADK's
  LoopAgent is a loop primitive with *global* variables — the design to
  avoid. CrewAI/LangGraph loop bookkeeping is DIY counters in shared state.

**Design deltas taken:** none structural — the report confirms §§2–4
wholesale. Adopt Mastra's three-part contract as the loop node's normative
statement; cite `context_outputs` as lineage precedent for `sets:`; keep
frame names lint-enforced (no convention namespacing).

## 9. Final synthesis (all three reports in)

The design as it stands after evidence, superseding §2 where they differ:

1. **One `loop` node type, two modes** (items present → WCP-14 foreach;
   absent → WCP-21 condition loop), body delimited by branch-node-set-style
   region rules, single entry/exit, nesting allowed.
2. **Type 1 items live in the workspace, not the checkpoint** (HumanLayer
   revision, §8.2): a spec'd checklist file the loop node re-reads and
   deterministically selects from on every arrival. Engine owns parsing,
   selection, injection — never storage. Replanning and steering are file
   edits. Freeze-at-activation is dropped.
3. **Type 2 frames ride the choice schema** (`sets: frame`), engine
   snapshots HEAD at set-time, scope pops at loop exit. (~~history appends
   to `frames.jsonl`~~ — dropped by §9a.3: frames are ephemeral; the paper
   trail already exists in `stages/NNN-*/outcome.json` + git.)
4. **Injection**: one engine-rendered block per active scope, innermost
   last, containing only assignments and engine facts (item, index/count,
   lap, sub-goal, base commit). SF's `$states` is the precedent; no
   narrative, no accumulation (context-withholding evidence, §8.2). Scoping
   rules copied from SF Variables: inner-reads-outer, no siblings,
   shadowing is a lint error, destroyed at exit, workspace is the export.
5. **§3.4 amended, not repealed**: budgets/visit counters stay hidden
   (bookkeeping); declared iterator state is injected (semantics).
6. **Futility detection** (doom-loop analog): warn/halt when the same item
   is selected M consecutive laps with an empty diff. Optional, post-v1.
7. **DSL/builder-library deferred with confidence**: loop+frames is the
   runtime either front-end would need; Mastra's serialized stepFlow proves
   builder→graph tractable; compile direction one-way, never round-trip.

Positioning fact from the survey: **no existing system injects engine-owned
loop-frame context into prompts** — history accumulation, hand templating,
or raw code everywhere; engine-synthesized briefings exist only in the
Attractor lineage's fidelity modes. Deterministic deixis resolution is
unoccupied territory, and it is precisely what makes two-word prompts
honest. The competing philosophy (HumanLayer's re-derive-each-lap) is
absorbed rather than refuted: re-derivation supplies the *truth* (workspace
status file), injection supplies the *orientation* (the rendered frame) —
sensor and whiteboard, not rivals.

## 9a. Side-chat doctrine (ratified in a parallel session, recorded here)

Three decisions reached in a side conversation, load-bearing for everything
after §9:

1. **The Go library is a compiler, not an AST translator and not a
   runtime.** Running the Go program IS the compilation: builder calls
   (`g.Agent(...)`, `g.Until(...)`, `g.ForEach(...)`) append nodes/edges to
   an in-memory graph; the last line serializes the YAML Tractor lints and
   runs (the CDK/Mastra/Airflow-TaskFlow pattern). Native Go `for` executes
   at build time (metaprogramming — unrolls known collections); anything
   depending on an agent's runtime result physically cannot be native
   control flow because `g.Agent(...)` returns a node handle, not text —
   runtime decisions can only be builder constructs with prose conditions.
   The over-engineering hole closes structurally (same rule as JAX: Python
   unrolls at trace time; `lax.while_loop` is the runtime construct).
   Costs: emitted YAML is a build artifact people shouldn't hand-edit;
   lint diagnostics should thread back to Go source positions. All
   deferrable; none touches the engine.
2. **Resumability demoted; "wrong goal" promoted to the real enemy.** This
   is a software factory that assumes friction — steps failing is why
   loops exist. Runs rarely die; they reach the wrong goal, very often.
   Restart-from-scratch is acceptable policy; crash-resume machinery is
   not worth paying for. Doctrine: restart must be cheap, so loop state
   must be re-derivable from the workspace; checkpoints are a shortcut,
   never the truth. The failure mode that matters is **drift** — steps
   reinterpreting their assignment, validators judging against the
   implementer's self-report — which frames counter (validator judges
   against declared intent; paper trail = intent + diff + verdict per
   lap). Resurrects **goal gates** from the archive inventory as the most
   direct anti-wrong-goal mechanism: the run cannot terminate at success
   until an exit gate demonstrates the goal's claims. Failure-distribution
   coverage: frames (anti-drift), futility detector (anti-spin), goal gate
   (anti-wrong-goal at exit) — none of which care about crashes.
3. **Frames are ephemeral — the engine computes, injects, and forgets.**
   `frames.jsonl` is dropped. Authoritative state = the workspace, period;
   derived state (current item, sub-goal, what's left) is a function of
   goal + live workspace. The Temporal tax begins at the exact moment you
   persist derived state, because persisted derived state acquires a
   consistency obligation; a cache owes nothing. Rule: the engine may
   cache, inject, and forget; only agents persist, and only into the
   workspace. The paper trail exists by side effect — the pick node's
   structured outcome lands in `stages/NNN-pick/outcome.json` anyway,
   beside `prompt.md`/`response.md`, with code changes in git. Resume
   collapses to something coarse and dumb: re-enter the loop at its
   boundary; the picker re-runs against the live workspace and
   fast-forward happens by judgment against reality, not record replay —
   strictly more correct, since it accounts for partial work and external
   edits. Chunks of existing checkpoint machinery (`retry_visit`
   arithmetic, session-binding restore) become forfeiture candidates.
   Caveat, fixed by doctrine not persistence: work leaving no workspace
   trace (rejected approaches, decisions) won't re-derive — anything worth
   surviving a restart must be written to the workspace by an agent as
   ordinary work (decision notes, checklist annotations), per HumanLayer's
   production memory-file pattern.

## 10. The runtime library, reassessed under the new doctrine

Given §9a (restart-over-resume; workspace as only truth; ephemeral frames)
plus Tyler's verdict that the runtime idea is strong. Reassessment:

**Why it is now strong.** The case against an embedded runtime was the
Temporal tax: a host-language program counter is durable only via
deterministic replay, with its determinism contract and event caps. That
tax is charged *only if you replay*. Restart-over-resume never replays —
restart re-executes from `main()` and fast-forwards by judgment against the
live workspace. No replay → no determinism contract, no history caps, no
Continue-As-New. The dominant objection is gone because the doctrine
removed the requirement that created it.

**Fast-forward by construction.** A restarted program must not redo done
work. The answer is not memoization (that is persisted derived state — the
tax returning through the side door) but **condition-first structure**:
pre-test loops (`Until(check, body)`) run the check first, so re-execution
skips satisfied phases naturally; sensors/pickers re-derive position from
the workspace. This is PR #27's prompt doctrine lifted to control flow —
prompts state conditions to bring about; workflows are condition-gated;
restart is just running the program again. Idempotent restart is a property
of *authoring style*, and the library's API should make that style the
path of least resistance.

**The discipline that closes the over-engineering hole in runtime mode**
(same law as compile mode, same law as the engine):
- `agent(...)` returns a handle, never text. Go control flow cannot branch
  on what an agent said.
- Go code MAY branch on **workspace facts** (file existence/contents, exit
  codes) and on **construct outcomes** (`Until`/`Check` prose conditions,
  resolved through the engine's choice schema). That is exactly the
  engine's own law — read files and exit codes mechanically, never
  interpret prose — so the library grants no power the graph withholds.
- Frames are the call stack, literally: the library knows which
  `Until`/`ForEach` scopes enclose each `agent()` call and composes the
  injected frame block from them. Ephemeral by construction — process
  memory, rebuilt on restart by re-execution.

**What runtime mode genuinely loses: static reviewability.** A YAML graph
is a whole-program artifact you can lint, render, diff, and human-skim
*before tokens burn*; a program's structure is discovered by running it.
This is the Relooper asymmetry operationally, and it matters most for the
factory case where *agents author workflows* — the lint gate is the
governor on agent cleverness, and runtime mode weakens it. Steering and
observability are also library work to rebuild (write the same run
directory, poll steer surfaces) rather than free.

**The architecture decision that decides trap-vs-great: one execution
semantics, not two.** A library that *interprets* loop/frame semantics
in-process is a second implementation of the engine's walk — it will drift
from the graph engine forever. The strong shape instead: **the runtime
library composes engine runs.** `agent()` / a region of constructs lowers
to a small graph fragment; the library drives it through the existing
`start_run` path (lint included — per-fragment lint-before-tokens partially
restored); the Go program is the caller, engine runs are the frames,
`$goal` is the argument register — option E from round one, matured into a
library. One executor of nodes (the engine), one routing mechanism (the
choice schema), one observability surface (run directories), with the Go
stack supplying scope, nesting, and dynamic structure. Needs almost no new
engine surface: a Go client over the existing CLI/MCP contract, plus the
frame-preamble composition, buildable against today's Tractor.

**Resulting stance.** One builder API, two backends: `Emit()` serializes
static structure to YAML (the reviewable, distributable, MCP-startable
artifact — the governor stays for agent-authored workflows); `Run()`
executes by composing engine runs (the power tool — literal stack,
workspace-fact dynamism, unit-testable workflows in ordinary Go). The
handles-not-text discipline polices both identically. Loop nodes + ephemeral
frame injection in the engine remain the prerequisite either way — they are
what fragments lower *to*.

## 11. CHA/OS crossover: Tractor as the agent calling convention

Read: ~/src/chaios (README, GOAL.txt, application-design/{intent,
runtime-loop, application-document, decisions-to-date}, application-proof
milestone evidence). Vision: blur writing and running a program until the
line disappears — first run is 100% agent throwing up UI from a prompt; the
agent progressively rewrites portions into deterministic software,
recompiles, restarts; some projects' goal is to bake the agents out
entirely. Every project carries its own on-site software factory. Tyler is
not into the node/graph surface: the wanted shape is "writing a program
where any layer in the call stack might be an agent," with agents holding
tools to call back into the runtime.

### Independent convergences (chaios already ratified our laws)

1. **Ephemeral frames.** application-document.md: "the document is
   prompt-bearing source material and the runtime is its prompt renderer.
   The rendered prompt is an observation for one decision, not the
   application's storage format or complete state." That is §9's
   compute-inject-forget doctrine, ratified in chaios before we wrote it.
   Chaios's "what the application LLM sees" (bootstrap + seed prompt +
   state summary + current interface, rendered per decision) IS frame
   injection — deixis resolution by the runtime.
2. **Handles-not-text.** decisions-to-date §0: "A turn that produces only
   assistant prose has produced no application result" — no chat bubble;
   agents express results by mutating state or installing components. The
   same law as `agent()` returning a handle: deterministic code (and the
   UI) never consumes agent prose directly; effects land on authoritative
   state surfaces. Two projects, two scales, one invariant — strong
   evidence it is the right one.
3. **Codergen↔tool duality = prompt↔code regions.** The application
   document keeps "natural language where behavior is deliberately
   open-ended; code where behavior has become concrete," and the loop "can
   move behavior from prompt to code — or revise the prompt." Tractor's
   node duality is the same axis: hardening a workflow is replacing
   codergen nodes with tool nodes; hardening an application is replacing
   prompt regions with compiled components. Bake-out is one operation
   appearing at two scales.
4. **Deterministic fast path.** runtime-loop.md: "not every event requires
   inference... delegate ordinary work to application code and invoke a
   model when the program needs interpretation, synthesis, or
   modification." Same razor as the engine's mechanical choosers.

### The unifying frame: programs with holes, and the ABI they need

- **A chaios application is a program with holes** — regions whose behavior
  is still intent, executed by inference until materialized. Prior art:
  typed holes / live programming with holes (Hazel; GHC typed holes) —
  programs that run *around* their unfinished parts. The agent is the
  hole-filler; the first-ever run is one hole with a goal.
- **Baking agents out is partial evaluation.** An agent interpreting a
  prompt is an interpreter running a spec; specializing that interpreter
  with respect to a now-stable spec yields a program (first Futamura
  projection). This gives the principled answer to "what may stay
  agentic": not primarily realtime-vs-not, but **spec stability** — keep
  interpreting where intent is still open; specialize where it has
  hardened. Realtime/sensitive code is the limit case: spec must be fully
  concrete and the latency budget forbids inference.
- **What "any stack layer might be an agent" actually requires is a
  calling convention.** You cannot have agent stack frames without
  defining what a call into an agent *is*. This conversation built exactly
  that, piece by piece: frame injection = argument passing (deixis
  resolved by the caller's runtime); the choice schema = the return type
  (structured, enumerated, never prose); `$goal` = the argument register;
  the workspace = the heap (only agents write it, as ordinary work);
  budgets/max_visits = fuel; the run directory = the trace; ephemeral
  frames = activation records. And the chaos CLI (`chaos state patch`,
  `chaos install`) is the other half of any real ABI: **syscalls** —
  agent→runtime, narrow and enumerated, while runtime→agent is the call.
  Tractor's durable contribution to chaios is not the graph; it is this
  ABI.

### On graphs, conceded and re-centered

The graph was never the asset. After this conversation's arc (structured
DSL → builder → runtime library → chaios), the graph's residual value is:
(a) a reviewable, lintable artifact for *detached, agent-authored* work —
the governor where structure itself is machine-written; (b) the reference
implementation of the loop discipline. It is the assembly listing, not the
language. Consequence for the nlspec: re-center the portable core on the
**calling convention + loop discipline** (frames, choice schema, budgets,
workspace law, evidence layout), with the graph engine as one host of that
core — the detached/batch host — and an embedded runtime (chaios's Bun
loop; the Go `Run()` library) as the resident/interactive host. "Simple
enough to write in every language" then applies to the right thing: each
project's language hosts a thin tractor runtime — the on-site factory —
while the spec keeps them interoperable.

### The bake-out is itself a loop Tractor already knows how to run

Notice-stabilized-behavior → write deterministic replacement → prove parity
→ swap. Parity proof uses the **agent as oracle** (the previous agentic
behavior generates the characterization cases) plus a goal gate at exit
(§ side-chat: anti-wrong-goal). The reverse move — chaios's open question
"can code later become an editable statement of intent again?" — is the
cell-lifecycle condemn/harvest-first-rebuild protocol Tyler already
operates; application-document regions are cells. The frame machinery is
what makes the swap auditable: declared intent per region is the thing the
replacement is validated against.

### The serious version of "how deeply embedded"

Track agent calls as an **effect**. A call site either permits the agentic
effect or forbids it; realtime and sensitive layers forbid it statically.
Then the hardening frontier is visible in signatures — the program shows
you exactly where it is still soft — and the bake-out has a burndown
metric: count of live agentic call sites over time, per region. This is
the type-system articulation of chaios's own layer separation, and the Go
library's handles-not-text typing is its first concrete instance.

### What this changes about next steps: nothing structural, one reframe (see §12 for the typed-call refinement)

Engine loop node + ephemeral frame injection remain first — they are the
argument-passing half of the ABI, needed identically by the graph host and
by any resident runtime. The Go builder/`Run()` library becomes more
important, not less: it is the ABI embedded in a host language, the
template chaios's Bun runtime would follow in TS. Worth considering soon:
extract the calling convention into its own short nlspec section (or
document) so chaios can implement against it without importing the graph.

## 12. Typing the ABI: go-gen-jsonschema, call species, and the frame-as-struct

Tyler pointed at ~/src/go-gen-jsonschema (his Wire-inspired build-time
generator: Go types → LLM-optimized JSON Schemas + validation + decoding;
doc comments become field descriptions; deterministic property order;
enums auto-discovered from consts; interfaces become discriminated anyOf
unions; Optional[T]/Nullable[T]; drift-proof via go:generate + hooks).
Also in the repo: `tool_types.go`, a `nobuild` design sketch of
`BuildTool(name, desc, impl)` deriving a tool's parameter schema from a Go
function signature — inline param comments as descriptions, injected
dependencies (\*sql.DB, \*slog.Logger) excluded from the schema. Both
directions of a typed agent boundary already sketched in one repo.

### Typed returns generalize the choice schema

Tractor's choice schema (`next` enum + notes) is a degenerate case of a
declared return type. With generated schemas, an agent call becomes
`func(...) (T, error)` where T's schema rides structured output and the
result decodes back into T. Discriminated unions make T a **sum type**:
edge conditions become union variants, and routing becomes pattern
matching (`switch` on the returned variant) — the graph's `next` enum was
always a sum type in disguise. The harness adapters already validate exact
caller schemas locally (design.md), so the transport exists.

**Refined handles-not-text law:** deterministic code may branch on agent
output only through **closed, schema-validated types** — enums, bools,
bounded numbers, discriminated unions. Open string fields are prose:
display, log, or workspace material, never branch conditions. Mechanically
lintable, since the generator knows which fields are closed.

**The type is part of the prompt.** Doc comments become the descriptions
the model reads; field order is deterministic and prompt-relevant. In this
ABI the return type declaration is not just validation — it is a
micro-prompt, authored in Go doc comments, versioned with the code.

### Call species (from Tyler's point 1)

Three callee species, one convention:

| Species | Runs | Tools | Return | Cost |
|---|---|---|---|---|
| `fn` / tool node | deterministic code | — | typed value / exit code | free |
| `infer[T]` | in-process hand-rolled loop over the raw model API | local Go functions (BuildTool-style, scoped) | T via generated schema | cheap |
| `agent` | harness-backed coding agent (Codex/Claude) | filesystem, shell, workspace | typed via structured output + workspace effects | heavy |

`infer[T]` is for loops known NOT to need coding-agent powers — no search,
no file I/O — where tool calls correspond directly to local functions.
This is 12-factor Factor 8 ("own your control flow") and HumanLayer's
micro-agents, typed. It is also spec §3.3's "cheap LLM node at a decision
point," promoted to a first-class species. Budgets/fuel apply to both
inference species; evidence discipline (run-dir logging) applies to both.

### The frame is a struct (from Tyler's point 2)

The chaos-CLI callback is too coarse: ambient authority (`chaos state
patch` can patch anything) and one global tool surface. The refinement
that unifies everything:

**Declare the call-site scope as a Go struct: its fields are the values
injected into the frame (deixis rendering = serialization of the struct);
its method set is the tool surface exposed to the callee.** One
declaration yields, via generation: the frame block the prompt receives,
the tool schemas the model sees, and the dispatch glue. Scope sensitivity
is then just lexical scope: different call sites, different structs,
different fields and methods. Capability attenuation for free — a region's
agent sees only its region's operations. And the frame is ephemeral *by
construction*: it is a stack value.

Transport per species: for `infer[T]`, tools are direct function calls —
no IPC at all. For harness-backed `agent` calls, the runtime exposes the
frame's method set over a per-turn unix socket; MCP is the natural wire
protocol (harnesses attach MCP servers natively, tool lists are dynamic
per session, schemas are the generated ones). A CLI can remain as an
optional thin shim over the socket for shell-native agents, with --help
generated from the same schemas — the CLI becomes sugar, not the
mechanism.

Language-neutrality note for the nlspec: the ABI requirement is stated as
"typed boundary, generated schemas in both directions, per-frame tool
scoping, closed-types-only branching" — Go implements it Wire-style with
build-time codegen; Python with pydantic; TS with zod. The convention, not
the generator, is normative.

### Open questions added

- Sum-type returns vs. graph-mode routing: in graph mode, does a union
  return *replace* `next` (variant → edge mapping) or compose with it?
- Which struct fields render into the frame block vs. stay
  callee-invisible (unexported = not rendered is the obvious rule).
- Per-turn MCP server lifecycle inside the engine's codergen handler;
  socket naming/cleanup under crash.
- Evidence format for `infer[T]` turns (they bypass the harness run-log
  path today).

## 13. Vocabulary ruling: "routing" is retired

Ratified in conversation (Tyler, 2026-08-24): the design has moved from
open node routing to lexical structure, and the vocabulary follows.
Under lexical nesting there is no routing question — control flow is fixed
by program text; at runtime an expression gets a value and ordinary
conditional logic consumes it.

Replacements:
- **routing → branching** (no special term needed; it is normal
  conditional logic)
- **edge condition → judgment** — a prose predicate evaluated by
  inference, returning a closed value
- **chooser → judge**
- **successor choice → the value of an expression**
- the **choice schema** is revealed as the *answer type of a judgment*

Unifying law (subsumes §12's refined handles-not-text and the old chooser
doctrine): **inference produces a closed value; ordinary conditional logic
consumes it.** Prose is evaluated only by models; deterministic code
consumes only the closed result. `until(judge("is the plan good?"))` is
`infer[bool]` with a prose predicate; a `switch` on a returned union
variant is an ordinary switch.

"Routing" survives only at the IR level — the compiled graph still has
edges and offered successors, as compiler output still has jumps after
`goto` left the source language. `Emit()` lowers a switch-on-judgment to
edges + conditions, and the generated edge-condition prose can be the
source predicate text verbatim, keeping the graph readable as the
program's assembly listing. Source-level speech and docs should not use
"routing" for anything an author writes.

## 13a. Correction and sharpening (Tyler): the value, not the struct

Tyler's pushback on §12's emphasis, ratified: "frame is a struct" is a
Go-flavored implementation sketch for the argument side and is NOT the
load-bearing idea (other languages may do it differently; it may not even
be right for Go). The load-bearing idea is:

**The agent call is an expression. Its typed result enters the program's
ordinary value economy — normal comparison operators, boolean logic,
switches, loop conditions, against the result of an agentic turn.**

Consequences:

1. The graph's routing apparatus (edges, offered successors, choice-schema-
   as-mechanism) existed only because the agent's result was not a program
   value — it lived in outcome.json, so the engine needed a protocol to
   act on it. With typed returns the apparatus is not replaced; it
   evaporates. Even §13's "judgment construct" carried residue:
   `until(judge("is the plan good?"))` is just `for !planIsGood(...)`
   where the predicate's evaluator happens to be a model. The prose is
   the function's body; the call site is ordinary code.
2. Final form of the law: **inference returns typed values; the program
   computes with them like any others, except prose fields are opaque**
   (closed-types-only governs which fields may be compared).
3. **This is the precondition for seamless bake-out.** Ordinary function
   semantics on both sides means replacing the agentic callee with a
   deterministic implementation is invisible to every caller —
   signature-preserving substitution. It works in reverse for tests
   (deterministic fakes standing in for agents; ordinary assertions on
   agentic results). An agent returning through a routing layer can never
   be swapped for a function; an agent returning a T can. §11's hardening
   story reduces to swapping a function body.

## 14. Positioning: is this just LangChain/LangGraph?

Asked directly by Tyler; answered from the §8 research evidence.

**Genuine overlap (steal the plumbing, claim no novelty):** typed model
calls returning values (with_structured_output / Instructor / Pydantic-AI /
Mastra); builder APIs serializing to graphs (Mastra's serializedStepGraph);
in-process tool loops over local functions.

**Genuinely distinct (first two mechanically verified by the framework
survey):**
1. Engine-rendered frames / deixis resolution — no mainstream framework
   synthesizes the state briefing into the next step's prompt; context
   assembly is always the author's job.
2. No data plane — a rejection of LangGraph's core (state channels +
   reducers = a typed context blackboard, the thing Tractor already cut);
   workspace as heap, run dir as evidence, derived state ephemeral.
   HumanLayer converged on the same rejection independently.
3. Restart-over-resume via judgment vs. LangGraph's snapshot-restore of
   persisted derived state.
4. Coding agents as callees (harness-backed workspace turns), a layer the
   frameworks don't touch.
5. The bake-out teleology — agents as construction equipment that
   eventually vacates via signature-preserving substitution; frameworks
   treat permanent agent residence as the destination. This is what makes
   the discipline stack (closed-types-only, prose opaque, ordinary call
   semantics) load-bearing rather than stylistic.

One-liner: LangGraph builds programs that permanently contain agents; this
designs a calling convention letting agents temporarily inhabit — and
eventually vacate — an ordinary program. LangChain v1's failure into
LangGraph (embedded chains → explicit graph + persisted state) is evidence
the durability wall is real; the workspace/re-judgment answer to that wall
is only available because the substrate is a git repo, not a Python object.

## 15. The heap is plural: substrates (Tyler's correction to §14)

§14's "substrate is a git repo" was Tractor-parochial. In chaios (and the
general ABI) the heap is a mixture of substrates, each imposing a
different write discipline:

1. **Repo** (intent + materialization: code, application document, specs,
   checklists). Agents write freely as ordinary work. Versioned, diffable,
   judgment-friendly — restart-by-judgment and bake-out are cheap here.
   The only substrate where "agents persist freely" holds.
2. **Sandboxed data store** (authoritative application state; chaios's
   Bun-owned revisioned state). Runtime-owned; agents mutate only through
   typed, validated mutations on the scoped syscall surface — never
   freehand. User data is not the agent's material. Survival contract:
   outlives restarts AND recompiles/bake-outs → schema **migration** is
   where this substrate couples to the repo; it is the part of bake-out
   that is not just swapping a function body.
3. **Secret store.** Already ratified in chaios's application-document
   ("secret values, even if the document names how they are obtained" are
   runtime concerns). ABI law: secrets never flow through the call — not
   into frames (they are injected into prompts), not into the run dir
   (evidence is readable forever), not into model context. References
   flow; values resolve only inside deterministic code at point of use.
   Frame rendering must be secret-safe by construction (secret-typed
   fields serialize as names, no other code path).
(4. The ephemeral substrate — sessions/frames — already covered: cache
   owes nothing.)

Reformulations forced:
- **Restart-by-judgment** reads the repo AND projections of the data
  store. HumanLayer's ralph loops already demonstrated the non-git case:
  Linear ticket statuses as the mutable checklist substrate.
- **Wrong-goal recovery is substrate-dependent**: repo damage is cheap
  (revert/rebuild); data mutations may be irreversible. Doctrine: a
  syscall's risk tier is determined by which substrate it touches —
  repo-ops permissive; data-mutating ops carry validation gates and
  concentrate goal-gating/human confirmation; secret-touching ops don't
  exist (only reference passing). This gives §12's scoped capability
  surface its severity axis.
- §14 one-liner corrected: the advantage was never "git" but that **all
  authoritative state lives outside the model's reach in inspectable,
  runtime-owned substrates** (versioned repo, revisioned data store,
  sealed secret store), vs. LangGraph's derived, in-process,
  model-adjacent state object. Git was one instance of the principle.

## 16. The process tree joins the family (Tyler): self-building as re-derivation

The binary — more precisely **the process tree** (binaries, bundles,
running processes, in-memory state, supervision topology) — is named a
substrate. This completes the family and reveals its two-class structure:

- **Authoritative substrates:** repo, data store, secret store (§15).
- **Derived substrates:** process tree, frames/sessions, compiled
  artifacts — all rebuildable from authoritative substrates (compilation
  is a function of repo content; process state restores from the data
  store; the tree's shape is declared in the application document's
  startup section). Nothing authoritative may live only here.

**The radical claim, made safe:** the program never mutates its process
tree directly; it mutates authoritative substrates and the runtime
**re-derives** the tree. Self-building = repo writes + re-derivation. The
supervisor *reconciles* the live tree to the declared tree (the document)
— Erlang-supervision / Kubernetes-reconciliation shaped, not command
shaped. "Modify the runtime live" vs "hot restart" are not rival
architectures but **granularities of re-derivation**: region-level
(compile → content-addressed artifact → mount/dispose swap; chaios
milestone 3, demonstrated once under controlled conditions) vs
whole-process (stop, recompile, restart, restore; chaios milestone 8,
demonstrated once). Content-addressing makes the region path safe
(identity by hash, no staleness ambiguity). See §16b for the evidence
calibration — the Chrome-operated visible-browser acceptance remains
unchecked across all milestones.

**Prior art: Erlang/OTP.** "Let it crash" + supervision trees =
restart-over-resume at process granularity (§9a.2 is OTP philosophy
independently re-derived). OTP's `code_change` carries the hard lesson:
swapping code under live state needs a state-migration callback.
Generalized doctrine: **every authoritative→derived pair carries a
migration obligation at hardening boundaries** — repo→data (schema
migration, §15) and repo→process (code_change/restart) are the same
coupling. Migrations are the tax bake-out pays beyond swapping function
bodies.

Structural consequences:
1. **Fixed kernel vs malleable body.** The loader, supervisor, syscall
   dispatcher, and secret resolver cannot hot-swap themselves; kernel
   changes take the whole-process path, body changes the region path.
   Chaios already has this shape (Bun runtime + browser chrome = kernel;
   canvas regions = body). Futamura reading: the kernel is the interpreter
   residue that never specializes away.
2. **Bake-out materializes in the tree.** Agentic regions derive through
   model-host processes; baked-out regions are in-process code. §11's
   burndown metric is observable as model-attached processes disappearing
   from the tree. First run = agent hosts around a thin kernel; fully
   baked = none.
3. **Risk tier (§15 axis):** process-substrate syscalls (`install`,
   `restart`, mount/dispose — `chaos install` is this substrate's
   syscall) are middle-tier: recoverable by re-derivation, but with
   availability blast radius.

## 16a. Prior-art hygiene ruling, and the BEAM question (Tyler)

**Ruling (locked in):** prior art may be cited only in bounded
single-sentence comparisons — "like X but only …" / "like X without …" —
never as an adoption frame or identity. Prior art cited as identity
becomes a to-do list for agents (cargo-cult risk). Applies to OTP,
Futamura, Kubernetes, LangGraph, LiveView, all of it. §16's OTP citation,
restated in legal form: *like OTP, but only the supervision-shaped
reconciliation, per-process blast radius, and restart doctrine — without
behaviours, without distribution, and without the hot-upgrade machinery.*

**Should we actually use Erlang/Elixir/BEAM? No.** Grounds from our own
doctrine, not taste:
1. The factory's binding constraint is **model fluency in the substrate
   language** — agents write most of the code; Go/TS are tier-1 for
   models, Elixir is not.
2. BEAM undermines the enforcement half of the discipline: §12/§13a
   (closed-types-only, prose-opaque, generated schemas, signature-
   preserving bake-out) leans on static types + codegen; Elixir is
   dynamic, Dialyzer advisory. The values discipline is what keeps agents
   safe; the restart discipline is a few hundred lines of supervisor.
3. We'd buy the OTP machinery we need least (hot upgrades — which the
   Erlang community itself mostly avoids in favor of restarts, i.e. our
   own §9a doctrine) while the part we need (supervision-as-
   reconciliation) is a small, well-trodden pattern in Go.
4. (Weakened by §16b:) chaios has demonstrated region-granularity
   re-derivation once on Bun/Svelte under controlled conditions — nothing
   observed yet suggests the runtime is the bottleneck. The verdict rests
   on points 1–3, which do not depend on chaios evidence.

**Bookmarked exception:** Phoenix LiveView is the proven embodiment of
chaios's renderer boundary (server-owned state, DOM as projection). If the
Bun/SSE path hits a real wall there, study it — like LiveView, but with
agent-authored components and a content-addressed loader. A fallback
reference, not a plan.

## 16b. Evidence calibration for chaios claims (Tyler's correction)

Tyler challenged §16/§16a's "already proven / already solved" language;
verified against ephemeral/projects/application-proof/milestone-evidence.md
in the chaios repo. Actual epistemic status:

- Nine milestones exist, each backed by ONE real controlled execution with
  verifiable specifics (independently checked content hashes, real Codex
  turn/thread ids, an uninterrupted SSE stream across install → replace →
  remove, a real SIGINT restart restoring state and canvas, real turns
  where Codex authored then repaired a live component). Scope-check
  discipline in the evidence file is honest and good.
- BUT: every milestone's scope check states the **Chrome-operated
  visible-browser acceptance remains unchecked**; milestone 9 explicitly
  declines screenshot proof (no in-app browser surface available). The
  human-witnessed live-mount-without-reload leg has never been accepted.
  The advice and semantic-trace milestones are also outstanding.
- Correct claim strength: **feasibility demonstrated (n≈1 per path, under
  controlled conditions)** — not "proven," not "solved," not "a working
  vertical slice" in the accepted sense.

**Ruling (vocabulary for this memo and future docs):** "demonstrated"
means executed at least once with captured evidence and stated conditions;
"proven" is reserved for claims that passed their full acceptance
milestone including the human-visible leg. Check every load-bearing
"already proven" claim against the evidence file before repeating it.
Consequential edits applied: §16 (granularities), §16a point 4 (BEAM
verdict now rests on points 1–3).

## 17. Hot swap in Go: re-derive processes, don't inject code

Tyler asked what hot swapping / code injection looks like inside Go, and
whether Bun exists because Go can't.

**Ruling: no code injection into running Go processes.** Ranked options:
- `buildmode=plugin`: standing prohibition (agents reach for it) — exact
  toolchain lockstep, shared-dep version lockstep, and plugins can NEVER
  be unloaded: every swap leaks old code permanently. Disqualified by
  design for continuous re-derivation.
- Embedded interpreters (yaegi/starlark/goja): weak middle stage — slow,
  subset languages, unreal sandboxes. Subsumed by wasm.
- **WASM via wazero** (pure Go, no cgo): the one good in-process answer if
  ever needed — modules load/unload/swap cleanly, are content-addressable,
  and **host-function imports = §12's scoped syscall surface**: a module
  is instantiated with only the functions in its frame's scope, so
  capability attenuation is the linking model, not a discipline. Language-
  neutral (suits the nlspec); costs marshalling + TinyGo limits.

**The doctrine-aligned answer:** restart-over-resume + the substrate model
make process memory a cache, so process death is cheap, and Go's compile
speed makes re-derivation seconds-fast. Granularities:
- Whole-process re-exec (default; §9a applied to the program itself).
- **Region = child process**: a region compiles to its own small binary;
  the supervisor kills/re-execs it; it speaks to the kernel over the
  per-frame socket already designed as the syscall surface — §16's process
  tree taken literally (like OTP supervision, but with binaries re-derived
  from the repo instead of hot-loaded beam files). Content-addressing
  extends to binaries: artifact cache keyed by source hash; manifest
  declares the tree; supervisor reconciles.

Inversion worth recording: Go's weakness at hot code loading is neutralized
by doctrine ratified for independent reasons (§9a) — we chose the
architecture that makes Go viable before asking the question.

**Bun's role, scoped:** JS gets region swap nearly free (dynamic import of
content-addressed URLs, GC unloads, no lockstep) — which is why the
malleable presentation body lives there (chaios milestone 3's mechanism).
Division of labor: Go for fixed kernel + factory machinery; JS/browser for
the presentation body; child processes as the general region mechanism;
wasm as the eventual language-neutral region format if in-process swap
earns its way in. Bun was not a workaround for Go; it is the natural
runtime for that layer of the body.

## 17a. Tree width is a phase indicator (Tyler's pushback on §17)

Concern: does region-as-process mean a giant process tree for one program?
No — resolved by the bake-out teleology: **process-tree width is a phase
indicator, not an architecture constant.** Soft program = thin kernel +
agent-host processes. Hardening graduates regions into ordinary functions
in the next kernel binary (a hardened region is trusted, typed code that
needs no isolation boundary — zero IPC). **The mature program is one
process and a browser tab**; process count approaching 1 is the sharpest
form of the §11 hardening burndown.

The isolation ladder (process boundary = most expensive rung, spent only
where its properties are needed):
1. **Function in the kernel binary** — the default destination of every
   region; swap = whole-binary re-derivation (cheap by doctrine).
2. **wasm instance in-process** (wazero) — the "lightweight process"
   analog: soft server-side regions needing swap-without-restart or a
   sandbox; unloadable, content-addressed, capability-scoped, not an OS
   process.
3. **Child process** — only for agent hosts, different-runtime bodies
   (Bun), or heavyweight untrusted regions. A handful, not dozens.
4. **Browser realm** — presentation body.

The supervisor is unchanged: reconciliation compares the declared region
tree to the live one; nodes may be in-binary regions, wasm instances,
processes, or browser mounts — only mount/dispose mechanics differ per
rung. Supervision model yes; giant OS-process tree no — the tree is mostly
logical.

Pinned distinction: wide process fan-out is correct in the **factory**
(Tractor runs, parallel branches, agent fleets — elastic, transient
construction machinery), never in the **program**. Program tree small and
shrinking; factory fan-out wide and ephemeral.
