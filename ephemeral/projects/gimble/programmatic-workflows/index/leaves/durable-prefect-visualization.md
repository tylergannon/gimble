# Prefect: code-derived visualisation is a controlled execution

**Purpose.** Prefect is included as a precise counterexample to the claim that ordinary program code can be faithfully visualized before executing. Its visualization support deliberately executes flow code outside tasks and requires mock values to choose dynamic paths.

**Findings.**

- The official guide says both visualization methods execute code outside tasks and advises putting unwanted code in tasks (`durable/prefect/visualize-workflow-structure.mdx:7-15`; [pinned source](https://github.com/PrefectHQ/prefect/blob/34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef/docs/v3/how-to-guides/workflows/visualize-workflow-structure.mdx#L7-L15)). That is a pre-run trace mode, not static analysis.
- For loops and `if`/`else`, the guide tells authors to supply mock task return values to select the path visualized (`durable/prefect/visualize-workflow-structure.mdx:101-127`; [pinned source](https://github.com/PrefectHQ/prefect/blob/34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef/docs/v3/how-to-guides/workflows/visualize-workflow-structure.mdx#L101-L127)). A graph generated this way describes one chosen scenario, not all future agent/tool outcomes.
- The implementation tracks encountered tasks in a mutable tracker and derives edges from actual parameters/returned placeholder objects (`durable/prefect/visualization.py:72-149`; [pinned source](https://github.com/PrefectHQ/prefect/blob/34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef/src/prefect/utilities/visualization.py#L72-L149)). It then builds Mermaid/Graphviz edges from that recorded tracker (`durable/prefect/visualization.py:152-205`; [pinned source](https://github.com/PrefectHQ/prefect/blob/34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef/src/prefect/utilities/visualization.py#L152-L205)). This is source proof that the rendered topology follows exercised code paths.

**Implication for Tractor.** Static, human-useful visualization is not inherited for free from Go programs. A source-derived structural outline can represent both syntactic branches and mark opaque calls, runtime-sized fan-out, and unknown outcomes. A narrow declared graph is an optional product artifact when it needs to carry goals, validators, authority boundaries, or effect classes that source structure does not express. Neither should pretend to list the concrete dynamic subagents/branches selected by future agent answers or repository inspection; a scenario trace can complement either after planning or execution.

**Recovery.** Prefect does not bear on exact durable resume in this corpus; it is visualization evidence only. It should not be used to introduce a persistence requirement.

**Themes.** static contract versus dynamic trace; mock-value path selection; honest preflight visualization.

**Gotchas.** Calling a visualization routine can itself run non-task code. A single diagram without mock-path disclosure can conceal branches; treating it as a complete workflow contract would be misleading.

**Retrieval recipe.** For the safety boundary, start at `durable/prefect/visualize-workflow-structure.mdx:7`; for the dynamic-path limitation, `:101`; for tracker mechanics, `durable/prefect/visualization.py:72`.

**Status and license.** API evidence on 2026-09-08: active/non-archived, `main` `34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef`, commit date 2026-09-08, latest release `3.8.5` (2026-09-03), Apache-2.0; see `durable/status.json` and `durable/MANIFEST.md`.

**Unknowns and counterevidence.** No Prefect run/Graphviz output was produced. This source does not show every Prefect deployment/UI feature. It is sufficient only for the bounded proposition that its documented `Flow.visualize()` graph is constructed by executing flow code and may need mocked values for dynamic topology.
