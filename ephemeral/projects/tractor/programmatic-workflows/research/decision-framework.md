# Decision framework (agent proposal)

The decision is not code versus graphs in the abstract. Compare four concrete products:

1. Keep YAML graph authoring and improve known pain points.
2. Offer Go builders that construct today's graph before execution.
3. Offer ordinary Go control flow calling a Tractor execution library.
4. Compile a deliberately restricted subset of Go into the graph engine.

Separate these axes: desired outcome declaration; control-flow authoring; execution representation; effect/result contracts; observation; recovery. None logically forces the same choice on every other axis.

Judge against the five arts: validated completion, manageable agent context, human authoring, visibility, and steering. Consider static visibility (readable outline, complete possible routes, visual editing/round-trip) separately from runtime visibility. Do not count an observed execution trace as a pre-run preview.

The key falsifiable hypotheses:

- H1: Agents make fewer orchestration mistakes and need fewer repair turns using ordinary Go on the same nested-loop change.
- H2: Per-call typed result values eliminate workflow-specific routing expressions without transferring semantic validation authority to the generating agent.
- H3: A structural pre-run view is enough for human comprehension even when not a complete graph of possible execution.
- H4: Re-entering through fresh validation and durable project artifacts saves useful work after interruption without serializing a Go stack.
- H5: The reusable execution library remains small enough to embed without pulling in a workflow server or importing internal engine packages.

Do not evaluate by source line count alone, successful compilation, a demo diagram produced from a different specification, aggregate tests, or popularity metrics. Code authorability is an empirical claim to test; current evidence can establish feasibility, not comparative agent success.

Provisional POC comparisons: actual nested sprint workflow, one schema change, one routing-policy change, a pre-run human walkthrough, same-process fake effects for fast exploration, then real native agents and interruption during a code change. Record interruption stage and time to regain useful work. No requirement for identical traces after recovery. Do require fresh final validation.

No implementation work is part of this research task. A future new Tractor build needs the repository's canonical integration proof in addition to changed-behavior evidence.
