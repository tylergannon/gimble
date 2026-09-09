# Proposed experiment, not an implementation plan approved by Tyler

**Historical proposal, not the current starting point.** Read [TIE-OFF.md](../TIE-OFF.md), especially its eval and experiment sections. The latest discussion requires testing the separation of program-specific declarative input from alternative Go control policies, runtime observation, and useful partial static structure. Do not automatically execute or turn the earlier checklist below into requirements.

## What would be built

One experimental Go-authored nested goal/sprint/critique/coding/validation workflow using a small Tractor library. Keep the YAML implementation as the control. Do not rewrite the engine, migrate all workflows, add a new DSL/transpiler, or adopt a workflow server to conduct this experiment.

The executable should run as a normal Go program in another module as well as in Tractor's example tree. Existing configuration can still supply models, prompts, goals, and validation commands. Package-level generic calls may return each caller's concrete result type. Ordinary Go controls decisions and loops. Treat sample API names in the research report as proposals, not existing functions.

Start with the existing public `harness.HarnessBackend.RunResult` boundary, already used by `cmd/tractor/run_prompt.go`. It accepts the caller's exact schema and returns a validated generic object before pipeline-specific `Outcome` decoding. The POC need not first replace the adapters or extract the graph engine. Current `engine` is already embeddable for graph runs; the experiment targets graph-independent policy and typed ergonomics.

## Comparison tasks

Use the same requirements and initial fixture for both versions. Give lower-model agents separate fresh contexts. Record actual diffs, compiler/linter errors, repair turns, behavioral mistakes, and tokens/time if available. A few matched tasks are directional evidence, not a statistically reliable benchmark.

1. Add the user's nested validation-design critique loop, including a check that critique has not expanded scope.
2. Change a reviewer result from a route choice to a concrete findings list, then make Go route on material findings after policy review.
3. Change retry/escalation policy in one inner loop without changing parent-loop behavior.

## Required demonstrations

- Before agent/tool effects run, show a hierarchical source-linked view of loops, conditions, named calls, and explicitly opaque dynamic calls. Ask a human to identify loop exit, validation authority, retry target, and places requiring attention. A runtime trace or separately hand-maintained diagram does not demonstrate this.
- Demonstrate different typed agent outputs at different call sites, including malformed JSON, valid JSON with the wrong type, and structurally valid but semantically unsuitable output. Transport success does not mean goal success.
- Run with real native agents through a failed validation, repair, and successful fresh final validation. Demonstrate that a reviewer saying ready cannot set done, and that cancellation does not turn into a successful result.
- Interrupt during a coding action and after a validation result. Account for any surviving child process, reopen using actual repository state, and demonstrate useful recovery sooner than restarting the matched work from a clean fixture. Recomputing planning is acceptable if this remains useful.
- Change the repository or active validation before restart and show that an old pass cannot silently authorize completion. No identical replay trace requirement.
- Embed from another Go module without importing Tractor internal packages or constructing fake graph nodes/edges. Observe named step events and stop behavior at that surface.

Every new Tractor build also needs the existing canonical integration command, exit zero, fresh result.json passed=true, review transcript inspection, binary hash, and artifact path. No such build has been created or claimed proved by this research.

## Decision rule

Proceed with Go authoring if the matched edits are meaningfully easier, typed domain results replace graph routing glue, the human understands the static view, and current completion/observation/steering behavior can be retained through a modest library extraction.

Keep YAML as primary if the gains reduce to nicer syntax, diagram needs force a second complete workflow definition, important guarantees require recreating a large analyzer, or the library is just the existing graph engine hidden behind builder calls.

Do not demand exact resume or full static path enumeration unless Tyler chooses those as requirements. If pre-run visual editing with round-trip preservation is essential, explicitly reevaluate the recommendation before expanding the experiment.
