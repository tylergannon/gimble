# Tractor programs, declarative tasks, visualization, and evals: conversation tie-off

Date: 2026-09-08. Workspace: `/Users/tyler/.codex/worktrees/0f95/tractor`. Branch: `codex/programmatic-workflow-research`.

**Broader direction, 2026-09-09:** [Tractor's direction](../../../../docs/direction.md) now captures Tyler's emphasis on preserving graph legibility in Go that reads like pseudocode and helping authors understand the context-engineering consequences of workflow shapes. Future run analysis connects actual requests, semantic index/content, steering, outcomes, and cost. This is direction, not added first-version scope; [his statement](DIRECTION-CLARIFICATION-VERBATIM.md) is preserved verbatim.

**Subsequent implementation:** Tyler later authorized a time-bounded, narrow Go POC. [POC.md](POC.md) records the resulting primitives, iterator decision, runnable commands and live results. This document preserves the design conversation before that authorization; its statements that implementation had not begun are historical.

**Subsequent context clarification, 2026-09-09:** [Workflows as programs](../../../../docs/workflows-as-programs.md) now records the unified-context direction: goals, requirements, scoped state, the semantic index, and on-disk information belong to one agent information environment from the outset. Orchestration distributes attention and responsibility across agents and time. Agent success governs the design; fully realizing the program metaphor does not. [Tyler's clarification](CONTEXT-REFRAMING-VERBATIM.md) is preserved verbatim.

**Read this first when continuing.** It records the latest user corrections and supersedes the architectural emphasis and proposed experiments in [DECISION.md](DECISION.md), [established-shapes.md](research/established-shapes.md), and [poc-contract.md](research/poc-contract.md). Those documents preserve earlier thinking; they are not a settled implementation plan. The [semantic index](index/README.md) remains the entrypoint for source evidence. Do not restart the research survey or treat an older agent recommendation as Tyler's decision.

Tyler asked to reach a conversational tie-off and preserve the journey before context fills. This is a research/design handoff, not an implementation, release, or benchmark result. No production code has changed, no new Tractor binary has been built, and no migration or POC has been authorized or executed.

## 1. The current center of the idea

**Tyler's direction:** separate what state a task should achieve and how to recognize it from the program that advances toward it. The program is ordinary Go, with loops, function calls, checks, typed agent results, and supervisors that can steer particular steps. Each program publishes the JSON Schema for its own declarative input. That input is plain JSON; it describes the task's goals, requirements, promises, validation information, and possibly guidance for detecting and correcting drift. It is not a replacement graph-description or action-description language.

**Tyler's proposed third tier:** hold an example project and its established requirements/validations fixed, run them through different orchestration shapes, and compare tokens, wall time, and other meaningful outcomes. This makes the choice and evolution of shapes an empirical question. It is a central consequence of the separation, not an afterthought about measuring a single routine.

**Tyler's visibility requirement:** any observer should be able to see where work is, where it has been, and where time and tokens have gone. Useful static diagrams are still of interest. Analyze control flow with respect to selected agentic operations, checks, and steering relationships. Recognize supported shapes such as nested loops; other regions may remain opaque. Runtime visibility remains valuable even where static extraction is incomplete.

**Tyler's reason for Go is substantive:** the current edge-oriented authoring model resembles manually wiring GOTOs when viewed as a program. Moving to Go removes real representational constraints and transition/scoping bookkeeping, rather than merely making a graph prettier to type. A graph can remain a derived picture without remaining the executable authoring language.

**Agent synthesis, not a ratified API:** Tractor could provide an embeddable execution library and useful maintained Go programs/routines, plus program input-schema publication, observation, and an eval facility. Reuse of established routines and freedom to author Go are compatible. We have not selected a fixed catalog, a universal input type, naming conventions, a schema generator, an event format, or the extent of static analysis.

## 2. Terms that were getting overloaded

| Term | Meaning in this discussion |
|---|---|
| Declarative input shape | The JSON structure a particular program accepts: its task-specific goals, requirements and validation information |
| Orchestration shape | The Go control structure used to advance the work: loops, decomposition, checks, critique, supervisors and steering |
| Static diagram | A projection of program structure onto relevant operations and their control/steering relationships |
| Runtime picture | Actual invocations, active work, history, repeated iterations, outcomes, waits and usage |
| Validation | The task's means of recognizing the desired state; it can use code, an agent, or a combination |
| Compass | Optional guidance/observation for noticing divergence and finding a useful correction before final success |
| Eval | Comparing different orchestration programs/shapes against the same example task and established acceptance conditions |

Do not silently substitute one meaning of "shape" for another. A program owning its JSON input shape does not mean its control flow must be declaratively encoded as JSON. A static diagram need not be an executable graph. An eval of orchestration does not require turning every project's validation into a generic scoring framework.

## 3. The journey and the corrections to retain

### Initial question

Tyler questioned whether authoring graphs is the right abstraction for Tractor:

- Programs may be more natural for agents to author.
- Go brings scope and ordinary language tooling instead of reconstructing scope through lint rules.
- The predominant structures are loops within loops, with branching/fan-out, rather than arbitrary network design.
- Exact resume is an outlier; useful recovery from repository state may be enough if it regains useful progress faster than a fresh start.
- A Tractor library could support orchestration inside other projects as well as Tractor itself.
- Go should decide routing from domain-specific typed agent results, rather than every Codergen result having the same next/notes envelope.
- Declarative goals and validation remain essential: Tractor pulls work toward a declared desired state.

The initial example was an outer goal-validation loop, a loop to improve and critique the next sprint's validation design while rejecting scope expansion, and an inner coding/validation loop until the sprint passes. Treat that as motivating structure, not a complete workflow contract or approved algorithm.

Tyler requested lower-model research, copies of relevant source/docs, and a question-specific semantic index modeled after diffusioninc's df-semantic-index. That work was completed before the subsequent discussion.

### First agent recommendation: too centered on per-project workflow authorship

The agent recommended a small Go layer over the existing harness and a POC of a custom Go workflow, with a source-derived structural view. It overemphasized the difficulty of deriving a complete workflow from arbitrary Go and assessed analyzers against too broad a correctness standard.

Tyler corrected the framing: derive the task-specific "smooshy" content and use established orchestration shapes maintained inside Tractor. Possible names such as `develop.Big()`, `develop.Huge()`, `develop.Micro()`, and `plan.LFG()` were illustrative. No naming proposal, taxonomy, thresholds, or rigor levels had been agreed.

### Second agent recommendation: overcorrected toward known routines

The agent recognized the value of maintained development methods, but shifted too much emphasis onto a small catalog and warnings about its risks. It proposed using one routine on two tasks plus implementing a linter warning.

Tyler's latest response redirects that proposal:

- Obvious warnings that bad routines can be bad add nothing useful here.
- Warnings identify specific known bad shapes; they do not establish correctness. Do not inflate their role.
- Go linters are useful and relatively cheap. Build worthwhile ones, but put clever analyzer tricks behind the structural questions for now.
- The important new opportunity is running the same declared work through many different shapes and evaluating the results.
- Do not erase the underlying Go-versus-edge-language benefit by describing everything as a library-selection problem.
- Program-specific JSON input schemas, rather than a universal workflow graph schema, are central.
- The static-analysis target is control flow with respect to selected calls. This is narrower than general program-meaning discovery; supported nested shapes may be diagrammed while others are left alone.

### Runtime visibility clarification

Between these corrections, Tyler explicitly elevated live observer visibility: current work, prior work, time and token spending. The agent recorded a feasible conceptual separation between a routine's structure and its runtime occurrences. That remains useful, provided it is not used to dismiss the possibility of deriving structure from Go beyond a predeclared routine catalog.

## 4. The declarative boundary: each program publishes its own task schema

**User-stated direction.** Instead of publishing the JSON Schema of Tractor's graph definition language, publish the shape of goals and validations accepted by a particular program. Plain JSON is the interchange. The program owns its schema.

This shifts the authoring question from "which nodes and successor edges should this task have?" to "what information does this program require to pursue this task?" Different programs may ask for different information. Do not impose a global `Work{Goal, Promises, Validate}` type merely because it was used in an illustrative earlier response.

An example of the direction, not a prescribed schema:

```json
{
  "goal": "Export the currently filtered rows",
  "requirements": ["Preserve the selected columns"],
  "validation": {
    "fixture": "examples/filtered-export",
    "check": "exported_rows_match_visible_rows"
  }
}
```

Another program might accept a command and arguments, references to existing validation code, agent instructions with examples, or a combination. Those representations are decisions for that program's input contract. JSON does not serialize a Go function or closure; a program must define how an input refers to or configures executable validation. This is a concrete open boundary, not a reason to invent a universal validation DSL.

The distinction to preserve is between **declaring the desired state and how it will be recognized** and **encoding the algorithm that pursues it**. The first is input data; the second belongs in Go. Validations themselves may be executable software.

Publishing input schemas could support authoring assistance and input checking. Whether they are derived from Go types, written explicitly, or generated by a particular library remains undecided. No schema generator has been selected and no schema-publication API has been implemented.

A program's task-input schema and an individual agent call's output schema are separate contracts. Both matter:

- The former tells a caller what the orchestration program needs.
- The latter tells a model what domain-shaped result the Go caller expects.

Neither should be confused with the old graph envelope or with a required global result format.

## 5. Desired state, recognition, and directional feedback

Tyler's strongest formulation is that the whole problem should reduce to:

1. What state are you trying to achieve?
2. How do you know when you have got there?
3. Optionally/ideally, how do you notice you are pointed away from it and re-point toward it?

The third item is the compass. It may be a failing example, a check, supervisor feedback, or other task-specific guidance. It need not be a numerical score, a monotonic progress function, or a universal loss metric. A final pass/fail criterion can recognize arrival without explaining the next useful move; directional feedback can help choose that move without proving arrival.

**Agent inference:** an orchestration program can be understood as a policy that observes the current project and feedback, then chooses the next action. Several policies can pursue the same declared destination. This explains why the eval tier is natural and why Go control flow is useful. It does not prove that agent-driven changes monotonically improve the project.

Validation is not always a prompt. It may be entirely executable code, entirely an agent-operated assessment, or a combination. Preserve this breadth in any future API. Likewise, a supervisor's steering is not necessarily a next-node selection; it can correct the currently active step.

Tyler's engineering concern is adequacy: demonstrate that the software works without drifting into either superficial completion or elaborate provenance/scoring machinery that misses the actual requested behavior. Do not answer that concern by adding a new blanket proof bureaucracy.

## 6. Why ordinary Go still matters independently of routine reuse

The advantage is not exhausted by "agents know Go" or "we can provide reusable templates." Ordinary Go supplies lexical nesting, functions, conditions, returns, types and direct value flow. An edge-oriented program asks its author to encode those relationships by references to destinations and additional topology rules. In the initial example, the intent is several nested feedback loops, not a collection of unrelated jumps.

Established routines can themselves be written in Go, composed in Go, and evolved in Go. Custom programs may also be written in Go. The current direction does not authorize forcing all program authors to choose from a fixed number of routines, nor restricting executable Go to whatever subset the visualizer understands.

A Go graph builder is a distinct alternative. It preserves an explicit graph and may be useful in some circumstances, but ordinary Go evaluated during graph construction is not the same as ordinary Go deciding what happens after future agent results arrive. Do not declare the user's main interest satisfied merely by wrapping node/edge construction in Go methods.

Static diagrams or runtime relationship graphs may remain derived representations. Using a graph to explain a program does not require using a graph language to execute it.

Naming remains open. `Big`, `Huge`, `Micro`, and `LFG` were examples, not a claim that size is the correct taxonomy. Do not spend another response cautioning against an unproposed naming scheme or restoring old size thresholds.

## 7. Static visualization: projection onto relevant operations

**User question, still open:** can we map a program's flow with respect to particular calls, especially nested agent/check loops and supervisors with targeted steering? If a complete map is too much, can we diagram supported shapes and leave others alone?

**Agent assessment:** this is a plausible, bounded analysis problem worth testing. Lexical `for` loops, nested loops, `if`/`switch` branches and named calls are present in Go syntax. It is unnecessary to know a loop's future iteration count or an agent's answer to display the loop or both branch bodies. The existing graph also does not predict those values.

The relevant representation is a **control-flow projection**, not just a call graph. A call graph that says A calls B and C omits whether B precedes C, whether one is conditional, and whether the pair repeats. A useful projection preserves those relationships around selected operations and suppresses uninteresting implementation detail.

Candidate scope, not an approved support matrix:

| Source feature | Plausible diagram treatment |
|---|---|
| Direct recognized agent/check call | A named operation linked to source |
| Lexical loop containing relevant operations | A loop region; iteration count may be unknown |
| Nested loops | Nested regions, not unrolled hypothetical future instances |
| `if` or `switch` around relevant calls | Branch alternatives with source conditions |
| `break`, `continue`, or early return | Preserve their visible effect on the represented flow |
| Statically resolvable helper/routine | Expand or summarize it at a useful level |
| Known supervisor attachment | A steering relationship to its target, distinct from normal execution order |
| Dynamic dispatch or unresolved helper | Opaque region or unresolved call; observe actual children at runtime |
| Runtime fan-out over discovered work | A repeated/parallel region whose concrete children appear when known |

Do not promise support for all these features before a spike. In particular, wrappers, recursion, deferred calls, concurrency and dynamically configured relationships require deliberate handling. But do not use their existence to dismiss straightforward loops or require a second authored graph for every program.

Known maintained routines are an advantageous common case, not the only case worth analyzing. Their recognizable operations can help both source-derived views and source diagnostics. A partial extractor should be explicit about unsupported regions and should not prevent those programs from running simply because they cannot be fully drawn.

Supervisor relationships matter independently of ordinary control-flow edges. A supervisor may be active alongside a coding step and send guidance into that step. A picture that shows it only as another sequential box loses part of what the program does. Exactly how targets are declared and discovered remains open.

The original fear of "discovering the meaning of an arbitrary Go program" was too broad. The goal is to expose selected, useful structure, not reconstruct every domain meaning or predict every possible concrete execution. No extractor has been built or demonstrated yet.

## 8. Runtime visibility: core requirement, not optional decoration

Tyler wants observers to understand:

- where work is now;
- where it has been;
- what is active, waiting, failed, or complete;
- where time and tokens have gone;
- the relevant agentic steps and specified checks;
- nested work/attempt context and targeted supervision.

The conceptual model recorded earlier remains useful: associate a structural operation with each actual invocation of it. One validation call in the program may have many executions across sprints and attempts. Keep those occurrences distinguishable, while allowing aggregated views. Parallel work may have multiple active operations.

A source/routine view explains potential structure. Runtime events explain actual activity and can add children where the static view was incomplete. A hierarchy, timeline and aggregate spending view are complementary; a graph alone is not a complete explanation of elapsed time.

Examples of intended observer questions, not an agreed screen design:

- "Which sprint and attempt are active, and what check are we trying to satisfy?"
- "Are we spending our effort implementing, reviewing validation, or waiting?"
- "Which previous failed checks led to the current attempt?"
- "Which supervisor is steering this coding step, and what did it say?"

Time accounting should distinguish wall-clock elapsed time from overlapping child durations. Token accounting should count actual reported operation usage once rather than double-counting cumulative updates or parent/child totals. Unknown telemetry stays unknown. Detailed event fields, normalization and UI design have not been decided.

Current Tractor source provides useful seams: stage starts/completions, sequence numbers and elapsed duration in `engine/runner.go`; timestamped timeline events in `engine/store.go`; usage events in `harness/contract.go`. These facts do not demonstrate the proposed Go-program observer UI or complete cross-provider accounting.

Observer history does not require exact durable replay. Recording what happened is a separate commitment from reconstructing program execution from that record.

## 9. The eval tier: compare mechanisms while preserving the task

**Tyler proposal:** take an example project with established requirements and validations, run it through many orchestration shapes, and measure tokens, wall time and other useful outcomes.

This is a stronger test of the proposed separation than the previous agent suggestion to run one routine on two projects. That suggestion tests reuse across tasks; it does not identify a better way of pursuing a fixed task. Keep those two axes distinct:

| Comparison | What it can tell us |
|---|---|
| Same task and acceptance, different orchestration shapes | How mechanisms differ in outcome and resource use |
| Same shape, different tasks | Whether the mechanism generalizes or requires task-specific rewrites |
| Same shape/task, different models | Model sensitivity; a separate experimental variable |
| YAML versus Go authoring of matched behavior | Authoring friction and semantic mistakes; not the whole program-shape eval |

**Agent proposal for an informative first experiment, not a scheduled build:** one example project and established acceptance, at least two meaningfully different Go orchestration programs, and the same starting project state for each run. Both consume the same task semantics; both are judged against the same acceptance. Observe actual work and collect time/tokens through the runtime observation mechanism. Repeat sufficiently to see whether differences survive obvious stochastic variation before ranking methods confidently; no repetition count has been selected.

The important outcomes include whether the software actually reaches the declared state, how much work/time/tokens it takes, and where human intervention was needed. Failed or interrupted runs remain part of the comparison rather than disappearing from a success-only average. Do not select a "fastest" program simply because it stops early without fulfilling the task.

Program-specific input schemas do not prevent comparison, but they create an interface question. A compatible family can share a task schema; alternatively, a fixture adapter can bind the same requirements/validation to different input types. The binding must not quietly weaken the task for one candidate. A universal input schema for all Tractor programs is not implied or decided.

An orchestration shape may include generating or improving intermediate sprint checks. If that behavior is evaluated, the established end-goal acceptance still needs to remain comparable; the candidate cannot win merely by changing what counts as arrival. The concrete mechanism should remain appropriate to the example project, not grow into a general provenance tribunal.

The same per-invocation observations could feed the live observer view and the eval summaries. That is a useful potential architectural economy. It does not require building a generic analytics platform before running two candidates.

A compass is also a potentially controlled input: compare policies with the same guidance first if isolating orchestration, then vary guidance deliberately if that becomes the question. Again, no scoring taxonomy or experimental protocol has been adopted.

No eval has been run. No claim that Go is cheaper, faster, or more successful has been demonstrated. Source research established plausible mechanisms and tradeoffs, not comparative performance.

## 10. Linters: useful secondary tooling

Tyler's correction is explicit: warnings identify specific known bad shapes. They do not establish correctness. Linters and ordinary unit tests are not the basis for claiming that the software works.

Go analyzers are attractive because ordinary Go tooling makes useful diagnostics accessible. A missing-validation warning is a reasonable example, not the centerpiece of the architecture or a promised salvation from bad engineering. Other useful analyzer ideas can be retained for later, but clever tricks should be backburnered while the program/input/observation/eval boundaries are settled.

Do not require building a linter as the first proof that this architecture is worthwhile. Do not make absence of warnings an eval-success criterion or a software-completion verdict. Avoid another extended debate over precisely what a warning proves; that framing was a distraction in this conversation.

## 11. Recovery and durability: retain the original priority

Tyler has rarely needed resume outside testing and does not consider exact execution serialization a major value. Returning near a useful point faster than a clean start may suffice. Existing project state is valuable state.

The research found that current Tractor already reconstructs loop frames from ledgers and may rewind to an enclosing loop. It is not serializing a complete Go stack today. Ordinary Go authoring, task schemas, runtime observation and recovery policy are separable decisions.

A future implementation should account for active child work and external actions when resuming, but these considerations are not permission to import an enterprise durability architecture. Exact replay, checkpointing every value, provenance logs and exhaustive effect accounting are not requirements established here.

## 12. Research already completed and how to use it

Two Terra researchers and a Luna researcher collected selective source/docs and wrote question-specific semantic leaves. A fresh Luna reader performed an eight-question source-retrieval smoke check. Root integrated the findings and corrected several overstatements, particularly around static visibility and the existing generic harness boundary.

The corpus contains 90 source/documentation/license/metadata files, about 1.36 MB. The recorded audit has 14 leaf/family files, 6 topic routes, 165 checked local source anchors, no orphan leaves, and all 12 structural retrieval targets reachable. These are research-collection checks, not software proof. The eight-query exercise had no unindexed baseline and was not a blinded experiment: the reader reported opening expected-target data after retrieval. Its limits are recorded in [the review](index/.semantic-index/retrieval-review.md).

| Topic | Entry point |
|---|---|
| Current Tractor schema, generic results, graph lint, embedding and resume | [Tractor leaf](index/leaves/static-tractor.md) |
| Imperative coding-agent flows and typed outputs | [Orca](index/leaves/imperative-orca.md) |
| Ordinary typed Go function flows and named trace steps | [Genkit Go](index/leaves/imperative-genkit-go.md) |
| Explicit typed graphs versus general typed AI output | [Pydantic](index/leaves/imperative-pydantic-graph.md) |
| Agent composition, DOT rendering, output-state boundary | [ADK Go](index/leaves/imperative-adk-go.md) |
| Go dependency/query builders | [Dagger](index/leaves/static-dagger.md) |
| Functional tasks and documented visualization limitations | [LangGraph](index/leaves/static-langgraph.md) |
| Visualization by controlled execution and mocked outcomes | [Prefect](index/leaves/durable-prefect-visualization.md) |
| Durable code orchestration and its cost | [Temporal](index/leaves/durable-temporal-go.md), [Restate](index/leaves/durable-restate-go.md), [DBOS](index/leaves/durable-dbos-go.md) |
| Reconciliation and Go control-flow infrastructure | [Concepts](index/leaves/static-concepts.md) |
| Adapted semantic-index process | [Method](index/leaves/method-semantic-index.md) |

Read a relevant leaf and then its exact source citations, rather than the whole corpus. Snapshots are pinned; `.go.txt` preserves Go source without adding upstream packages to Tractor's package discovery. Manifests preserve original paths and hashes. No fetched program was executed. Some external prose is URL-only, and license uncertainty is recorded for the reference material where appropriate. Do not treat repository activity or a release as production proof.

If further collection/indexing is needed, retain Tyler's preference for Terra/Luna research workers. The adapted semantic index is a local routing tree over source text, not a requirement for embeddings or a vector database. The diffusioninc skill is already copied under `corpus/method/`; its public/raw URL returned 404 during collection but authenticated `gh api` access succeeded. Do not infer disappearance from that public-fetch failure.

### Source findings that materially affect the next step

- At Tractor source head `07c04ff4c62ff91c625d2e27b6427cd594f67f39`, public `harness.HarnessBackend.RunResult` accepts the caller's exact schema and returns a generic validated object before `Run` decodes the pipeline-specific `Outcome`. `cmd/tractor/run_prompt.go` already uses it. A typed-output experiment need not replace the provider integration layer first. The existing root is a JSON object; arbitrary Go values are not model outputs.
- `engine` is already a public graph-runner library. The desired Go-program ergonomics are not proof that Tractor has no library today. The built-in workflow catalog is under `internal/workflows`.
- Current agent graph routing uses offered edge descriptions and a closed next-ID enum. Domain-shaped replies plus ordinary Go routing would change that boundary.
- `rebuildLoopFrames` reads ledger state and can rewind on resume. Do not characterize the current graph engine as an exact serialized execution machine.
- The captured Go CFG package exposes per-function structure and documents omitted details. Lexical syntax plus type information is useful for a selected-call view; it is not an implemented Tractor visualizer.
- Genkit shows typed ordinary Go function flows; Orca shows imperative agent orchestration and typed model results; Pydantic Graph's static diagrams rely on materialized topology; LangGraph Functional API docs explicitly decline static graph visualization. These are design references, not proof of what Tractor's extractor can or cannot support.
- Temporal/Restate/DBOS provide durable execution references. Their additional machinery should not become a requirement merely because it exists. Prefect's visualization illustrates the distinction between a chosen execution scenario and a static structural view.

## 13. What is established, proposed, and open

### User-established direction and priorities

- Plain JSON, with each program defining/publishing its declarative input shape.
- Desired state, recognition of arrival, and optionally directional guidance are the task's conceptual core.
- Validation may be code, an agent, or a combination; it is not always a prompt.
- Ordinary Go should free authoring from graph-language limitations and permit typed per-call outputs and program-owned routing.
- Established reusable routines are attractive, but no names or fixed catalog have been decided.
- Runtime visibility for observers is very important, including time/token attribution and history.
- Investigate static structure with respect to selected operations, especially loops within loops and supervisor steering. Partial shape support may be sufficient.
- Compare different orchestration shapes on established example tasks using an eval.
- Linters are secondary practical tooling; do not oversell them or let them dominate the design.
- Exact durable replay has low priority; useful recovery can be enough.
- The goal is working software and an appropriate demonstration, not ceremony or provenance machinery.

### Agent proposals that remain unratified

- A program/task-input/observation/eval separation as the next architectural skeleton.
- A selected-call control-flow projection with partial static support and runtime expansion.
- A first informative comparison using one task and multiple Go mechanisms before broadening to many tasks.
- Reusing `RunResult` as the first low-level integration seam.
- Using the same operation observations for the runtime UI and eval metrics.
- Any example JSON field names, Go API signatures, routine taxonomy, schema derivation method or UI layout in these notes.

### Open questions, not a questionnaire to launch automatically

1. What is the smallest program boundary needed to publish its input schema and invoke it with JSON?
2. How does a particular program bind input validation data to code, agent execution or a hybrid?
3. Which small set of operations must runtime observation recognize, including scoped supervisor relationships?
4. Which source shapes can be drawn usefully first, and what is the visible treatment of unsupported structure?
5. How does the same task bind to multiple candidate programs without changing its acceptance meaning?
6. Which example project and genuinely different control policies would make the first eval informative?
7. What practical runtime visibility is needed in that experiment, and what can remain an elementary timeline/tree?
8. Which responsibilities should remain inside Tractor versus in the embedding program?
9. What compatibility or migration path, if any, follows evidence from the experiment? No deletion of the graph language has been authorized.

Naming can wait. A complete arbitrary-Go map can remain an open research limit. A new distributed runtime, universal validation DSL, and linter showcase are not prerequisites for addressing these structural questions.

## 14. Why the earlier first experiment did not get us far enough

The earlier proposal was one maintained routine used on two different tasks, plus one missing-validation warning. It could demonstrate reuse and analyzer feasibility, but it would not show whether one orchestration policy was better than another for the same task. It also did not adequately test program-owned schema publication, the freedom gained by ordinary Go routing, or runtime cost attribution across alternative shapes.

The latest direction suggests a more informative experiment, with exact scope still to be chosen:

1. Establish one example task's desired state and practical acceptance.
2. Supply that task as JSON to two different Go control policies, each publishing its input contract.
3. Let both use the existing Tractor agent-call boundary, including at least one domain-shaped result that Go uses to make a decision.
4. Make the observer able to locate active agent/check work and inspect its history and reported effort.
5. Compare actual attainment, wall time, tokens and any required intervention under comparable starting conditions.
6. Try extracting a small supported nested-loop shape, while preserving runtime visibility for an unsupported portion if present.

This is an **agent proposal**, not a newly approved implementation checklist or a demand to build all features before learning anything. It should be cut into the smallest runnable slice that answers the structural question. A linter is not required in that first slice. The second axis—more tasks using the same policy—can test generality afterward or alongside it if inexpensive.

The experiment should tell us whether the separation holds, whether Go removes practical control-flow friction, whether observers can understand a run, and whether the same acceptance makes method comparison meaningful. It must not declare success because two APIs compile or a diagram looks plausible.

## 15. Direct record of the latest annotations

These notes preserve the corrections with enough context to avoid reintroducing them after compaction. Numbering matches Tyler's latest twelve response annotations.

1. **On the warning that a shared routine can institutionalize overengineering:** Tyler: "obvi." The agent should drop this obvious caution rather than spend attention on it as a new insight.
2. **On evolving routines from experience:** Tyler proposed a third tier: take an example project and established requirements/validations, run many different shapes, and measure tokens, wall time and other outcomes. This is the eval opportunity created by separating task data from the program.
3. **On warnings and correctness:** Tyler: "warnings never establish correctness. They just warn against specific known bad shapes." Same principle as not relying on linters/unit tests to demonstrate the software works. Do not qualify this into a prolonged correctness discussion.
4. **On focusing on missing-validation lint:** Tyler said linters are cheap in Go and useful, but "cool tricks" need to be backburnered while the most important structural aspects are resolved. Keep analyzer possibilities without making them the centerpiece.
5. **On showing a known `develop.Big(ctx, work)` routine:** The selected passage remains relevant as one useful visualization case. The following correction prevents restricting the whole idea to that case.
6. **On "discovering the meaning of an arbitrary Go program":** Tyler wanted to map program flow with respect to specific calls, recognized a complete map might be too much, insisted a runtime picture remained valuable, and suggested analyzing a subset of shapes while leaving others alone. Do not treat full static completeness as a prerequisite for useful extraction.
7. **On the declarative portion:** Tyler: "This should just be plain JSON, defined in particular by any given program." Publish the shape of goals/validations accepted by that program instead of the graph-definition-language schema. The graph schema was a language; the new input should not recreate it.
8. **On a validation example expressed as prose:** Tyler: "it's not always a prompt. Sometimes it actually is just a program. Sometimes it's a combination." The whole problem reduces to desired state, recognizing arrival, and optionally/ideally a compass for recognizing divergence and correcting direction.
9. **On warning against size-based names:** Tyler said no naming convention has been decided and nobody had put forth a claim. The agent had mistaken illustrative names for a proposed taxonomy. Do not repeat that mistake.
10. **On one routine/two tasks as the first experiment:** Tyler: "how does this get us there?" The experiment needs an explicit link to the structural and comparative questions; reuse alone does not answer them.
11. **On framing the benefit mainly as reuse and better task definition:** Tyler emphasized freeing programs from real limitations/pitfalls of graph authoring, which looks like "a bunch of `GOTO` statements" as a program. Keep the language/control-flow improvement central alongside library reuse.
12. **On seeing an enclosing loop without knowing future sprints:** Tyler identified this as the inference boundary of interest: mostly loops within loops, with supervisors steering specific steps. Recognize loops and selected relationships without requiring prediction of concrete future work.

## 16. Workspace state, proof boundaries, and continuation instructions

All work is on the existing research branch under `ephemeral/projects/tractor/programmatic-workflows/`. Relevant preceding local commits:

- `ea6d5d5`: initial source corpus, semantic index and decision report.
- `a4e0a7d`: reconsideration around maintained routines.
- `f7917bb`: runtime visibility elevated to a central requirement.

This tie-off will be another local checkpoint. No remote publishing, PR, merge, release, implementation, native-agent program execution, or program-shape eval has occurred in this task. Research assertions are based on source inspection. The audit and retrieval exercise validate navigation/integrity only.

The repository's `docs/five-arts.md` remains relevant: actual completion, manageable decomposition, human authoring, visibility and steering are competing priorities. The latest discussion addresses all five; do not optimize only agent authoring or only visualization.

The older `ephemeral/projects/tractor/workflow-designer/rules.md` records a prior rejection of automatic routing into three canned size graphs. This conversation reopens the design space; it does not revive the old implementation or thresholds. Latest direct user instructions govern this exploration. Avoid treating an exploratory new proposal as a settled contradiction requiring another permission ceremony.

Before claiming any future new Tractor build proved, the repository still requires:

```sh
go test -tags=integration ./internal/workflows -run '^TestCanonicalLoop$' -count=1 -v -timeout=25m
```

For a separately built candidate, use `-args -tractor-binary /absolute/path/to/tractor`. Require exit zero and the fresh artifact directory's `result.json` with `passed: true`, inspect review transcripts, and report the binary hash and artifact path. This is additional to demonstrating changed behavior. It has not been run for a new build here because no such build exists.

### Resume here

Read this tie-off and only the source-index routes needed for the next question. Reflect the current separation back accurately: program-owned declarative JSON, ordinary Go mechanisms, observation of actual work, comparative evals, and useful partial static views. Preserve typed agent results, nested loops and scoped steering. Keep naming and linter tricks deferred. Propose or implement only the next work Tyler actually requests; do not automatically launch the old POC, another broad literature survey, or a migration.
