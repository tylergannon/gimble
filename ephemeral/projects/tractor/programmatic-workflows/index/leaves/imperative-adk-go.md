# Google ADK Go: reusable loop agents, DOT topology, incomplete typed-output boundary

Purpose: Google ADK Go offers maintained Go evidence for compositional loop agents and a structural renderer, plus a sharp warning about treating advertised model-output schemas as locally typed values. Snapshot: Apache-2.0, commit `b52d1fbeaa3b457bc739269a79846b569bc12c55`; GitHub API reported non-archived on 2026-09-08, latest release `v2.3.0` published 2026-08-31. Corpus and hashes: `imperative/manifest.json`.

## Findings

- `loopagent.New` produces a configured agent whose `Run` repeatedly invokes sub-agents in order, stopping on an escalation event or the optional iteration count. Its zero iteration value means unbounded looping, so each goal-seeking scope needs an explicit stop, escalation, or attention policy rather than inheriting this default. `imperative/adk-go/agent/workflowagents/loopagent/agent.go.txt:28-104` ([pinned source](https://github.com/google/adk-go/blob/b52d1fbeaa3b457bc739269a79846b569bc12c55/agent/workflowagents/loopagent/agent.go#L28-L104)).
- `LLMAgent.Config.OutputSchema` causes the framework to add a `set_model_response` tool so the agent can still call tools while producing a structured response. It remains a `*genai.Schema`, not a Go `Out` generic. `imperative/adk-go/agent/llmagent/llmagent.go.txt:310-352` ([pinned source](https://github.com/google/adk-go/blob/b52d1fbeaa3b457bc739269a79846b569bc12c55/agent/llmagent/llmagent.go#L310-L352)).
- The important counterevidence is in the implementation: when saving an output, it concatenates text and explicitly says `TODO: add output schema validation and unmarshalling`; it stores the string in state. Therefore this snapshot does *not* demonstrate typed per-call outputs suitable for Go routing decisions. `imperative/adk-go/agent/llmagent/llmagent.go.txt:520-559` ([pinned source](https://github.com/google/adk-go/blob/b52d1fbeaa3b457bc739269a79846b569bc12c55/agent/llmagent/llmagent.go#L520-L559)).
- The server can render a Graphviz DOT agent tree. It recognizes loop/sequential/parallel agent types and derives edges from known composition; it walks `SubAgents()` and LLM tools. This is source-derived structural visualization of declared object composition, not a semantic preview of arbitrary code or dynamic transfers. `imperative/adk-go/server/adkrest/internal/services/agentgraphgenerator.go.txt:71-132`, `imperative/adk-go/server/adkrest/internal/services/agentgraphgenerator.go.txt:174-205`, `imperative/adk-go/server/adkrest/internal/services/agentgraphgenerator.go.txt:273-340` ([pinned source](https://github.com/google/adk-go/blob/b52d1fbeaa3b457bc739269a79846b569bc12c55/server/adkrest/internal/services/agentgraphgenerator.go#L71-L340)).

## Mechanisms to copy

- Make loops ordinary composition with an explicit scope-appropriate stop, escalation, or attention policy.
- Generate DOT/Mermaid from declared composition and static source structure, labeling unresolved dynamic dispatch as uncertain; ADK's renderer is credible precisely because it only renders recognized agent types.
- Put `json.Unmarshal`/schema validation between every model call and Go routing. ADK's TODO is a concrete regression test target, not a mechanism to inherit.

Unknowns: this source inspection did not run an ADK server or model provider; the renderer's handling of duplicate names and dynamically constructed subagents needs dedicated tests before borrowing its model.
