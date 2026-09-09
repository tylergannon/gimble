# Reconsideration: declarative work, established development routines

**Superseded emphasis:** read [the conversation tie-off](../TIE-OFF.md) before continuing. Tyler's later corrections make program-owned plain JSON schemas, ordinary-Go control-flow freedom, comparative evals across shapes, and partial selected-call visualization central. Maintained routines remain useful, but are not a fixed catalog; names are undecided; linters are secondary. The one-routine/two-task experiment below does not answer the whole current question.

2026-09-08. Tyler's new direction is exploratory, not an approved implementation or settled API. This note revises the earlier agent recommendation; it does not change the pinned source findings.

## What changed in the reasoning

The previous recommendation assumed projects would repeatedly author general-purpose orchestration programs, and assessed whether Go could recover guarantees previously obtained from graph topology. That is too narrow. Tyler proposes moving the recurring method into Tractor itself, leaving goal, promises, scope, validation and other task-specific judgment as the main material to derive for each project.

The unit of reuse becomes a development routine, not merely an agent-call primitive. A project could call `develop.Micro(ctx, work)` or `develop.Big(ctx, work)`. These are illustrative names, not existing APIs. The routine's implementation can use ordinary Go, typed agent results and other routines. Project authors need not reconstruct its loop or review policy.

This changes the economics: careful implementation, focused analysis, human review and runtime demonstrations of a handful of maintained routines can benefit many projects. The same investment made separately for each generated workflow would be wasteful. The API records practiced engineering judgment about what to do after a failed check, when to replan, and how to respond to critique without expanding scope.

## Why a Go analyzer fits

A useful warning need not certify the workflow or the software. Within recognizable Tractor orchestration code, an analyzer can flag a loop with no identified validation operation, a discarded validation result, or a visible path that continues work or reports success without considering the relevant check. These are candidate diagnostics, not established capabilities.

Start with known calls and local syntax/control flow. Report the observation precisely: no recognized validation was found, or a visible branch bypasses it. Unknown callees should not silently become certified checks, nor should every ordinary polling/collection loop be forced to contain project validation. A validation in the loop condition or a known helper must count; mere text matching for a function named Validate is insufficient.

Go's [analysis framework](https://pkg.go.dev/golang.org/x/tools/go/analysis) supplies typed syntax, source diagnostics and modular facts, so knowledge about wrappers can be extended if actual usage calls for it. This is a plausible implementation basis, not a proposal to start by building a complete interprocedural proof system.

The earlier assertion that the compiler cannot establish whether validation ran conflated compile-time inspection, runtime occurrence, and semantic adequacy. A bounded analyzer can establish some path properties within its model. Runtime execution can establish that a check was invoked. Neither by itself establishes that the check demonstrates the intended software behavior. These are different, independently useful contributions.

## Where engineering judgment belongs

The system cannot eliminate judgment about adequate validation. It can concentrate and reuse the recurring portion of that judgment rather than asking each agent to reinvent the development method. A routine can consistently require a concrete demonstration, route actionable failure back to implementation, and return scope-expanding critique for consideration instead of treating every criticism as mandatory work.

The declarative work description should remain small and task-shaped. For a filtered CSV export, an appropriate demonstration could be operating the actual UI, downloading the file and inspecting whether it contains the filtered rows. It need not become a generic provenance/scoring system. A running check and a persuasive demonstration are distinct; documenting that distinction must not itself cause another validation subsystem to be invented.

Task-specific judgment still exists when choosing a routine and designing its checks. Those choices should remain visible for human correction. A complex routine can include ongoing discovery and validation design; this proposal does not require a complete immutable specification before development starts.

## Visualization and extensibility

**Tyler clarification:** runtime visibility is a central requirement: any observer should be able to understand the current work, previous work, and where time and tokens have gone. Feasibility must be considered now; detailed implementation is deferred.

Known routines give a pre-run viewer useful structure without understanding arbitrary programs. It can show the selected routine, bound goal/checks and its stages; expand a routine to show its implementation's loops and decisions; then overlay the current iteration and findings during a run. The built-in view can track source or be verified alongside the maintained routine. Avoid asking each project to separately author its diagram.

The feasible concept is a structural routine model joined to runtime observations. Analysis identifies recognizable routines, calls, loops and branches; runtime events identify which operation ran, its particular invocation, parent invocation, current work item, timestamps, outcome and reported usage. One validation call in source may execute many times: retain those occurrences separately and aggregate them for a collapsed view. Parallel branches may have multiple active operations. Future data-dependent work can remain abstract until runtime expansion supplies concrete items.

An observer could move between a goal/sprint hierarchy with current status, a timeline showing previous attempts and waits, and a heat map attributing time/tokens to routine stages. A graph alone is insufficient to explain elapsed time or distinguish repeated visits. The same structural vocabulary could support analysis warnings and visualization without requiring a single giant analysis engine.

Current Tractor already records stage starts/completions, sequence numbers and elapsed duration (`engine/runner.go:535-591`), timestamps timeline events (`engine/store.go:98-117`), and exposes harness usage events (`harness/contract.go:180-188`). Those are existing source seams, not evidence that the proposed routine-level UI or normalized accounting exists. Accurate totals must distinguish wall-clock elapsed time from overlapping child durations and avoid counting cumulative provider observations or parent/child aggregates twice. Missing usage stays unknown. Observer logs do not imply a requirement to serialize or replay program execution.

The conceptual acceptance question for a future experiment should include whether an observer can locate the current goal/sprint/attempt and identify where time and reported tokens went without reading raw logs. No viewer implementation or new telemetry contract is being prescribed here.

Ordinary Go remains available for composing routines and implementing new ones. Custom logic may have a less informative preview; this does not require making the common maintained path equally opaque. Library use, customization and routine authorship are different levels of responsibility.

Prefer a small family of routines distinguished by real differences in development behavior. Size names can be convenient, but size alone must not secretly choose rigor, reviewer counts or evidence machinery. Knowing what to build and having a usable validator differs from needing to discover the work; both can occur in a small project. Do not create a configuration matrix of every imaginable policy combination.

## Revised next experiment

Exercise one maintained routine with two genuinely different tasks, each with its own declared goal and practical validation. Reuse the same routine unchanged, show its pre-run structure, and implement one useful missing-validation warning on a deliberately faulty custom routine. Inspect whether the warning helps correct behavior without inducing ceremony or evasion.

The decisive question becomes: can we vary the work and the demonstration while keeping a useful development method stable? This is more directly responsive than beginning with arbitrary Go workflow extraction or rewriting graph lint wholesale.

## Historical boundary

The existing workflow-designer rules record an earlier rejection of automatic size routing into three canned graphs. This conversation explicitly reopens the design space. It is not authorization to restore the old implementation or its thresholds; routine names, selection criteria and contracts remain proposals. The difference being explored is a maintained, composable development library whose task-specific content remains declarative.
