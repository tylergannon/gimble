# Domain results, routing policy, and library boundaries

1. Start with [Tractor](../leaves/static-tractor.md): the graph agent handler creates the closed next/notes schema, while public HarnessBackend.RunResult already returns the generic validated object before Outcome decoding. This is the smallest current experimental seam. Public engine embedding still takes a graph.
2. For schema + local decoding + corrective retry as a per-call API: [Orca](../leaves/imperative-orca.md).
3. For generic Go input/output types and named flow/step events: [Genkit Go](../leaves/imperative-genkit-go.md). Its typed action boundary should not be mistaken for verified model-wire output.
4. For a concrete warning that OutputSchema alone is insufficient: [ADK Go](../leaves/imperative-adk-go.md), especially the output-state validation/unmarshal TODO.
5. For typed Go durable function registration: [DBOS](../leaves/durable-dbos-go.md). Its persistence machinery is optional for this decision.

See [Guarantees](guarantees.md) for shape versus truth and who owns completion. Runtime program conditions can use domain values; models need not know graph successor IDs.
