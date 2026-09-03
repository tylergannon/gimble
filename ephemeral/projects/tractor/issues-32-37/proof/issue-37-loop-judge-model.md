# Issue 37 proof

Proved from the uncommitted working tree based on `371f882794a8f322ac982568466fe50240cc2794` on 2026-09-03.

## Default judge through the CLI

A disposable pipeline set these pipeline defaults while leaving the loop's model fields unset:

```yaml
defaults:
  timeout: 5m
  llm_model: gpt-5.6-sol
  llm_provider: openai
  reasoning_effort: high
```

`go run ./cmd/tractor run ...` exited 0 and printed `COMPLETED`. The resulting checkpoint bound the infer turn to the `agy` harness:

```json
"\u0000none:items": {
  "harness": "agy",
  "session_id": "86c89c54-2918-4807-8079-8d0bf019da99"
}
```

The native agy log for that session recorded the concrete model:

```text
Print mode: starting (promptLength=650, model="gemini-3.8-flash-medium", conversationID="86c89c54-2918-4807-8079-8d0bf019da99")
```

The judge opened `evidence.txt`; `validation.json` records `"verdict": "pass"` and `"passed": true`. The timeline ends with:

```json
{"item":"Evidence","node":"items","passed":true,"summary":"passed","type":"LoopValidated"}
{"count":1,"node":"items","type":"LoopCompleted"}
{"duration":"11.513258875s","type":"PipelineCompleted"}
```

## Explicit loop override

`go test -count=1 -run TestLoopInferJudgeModelIsIndependentOfPipelineDefaults -v ./engine` exited 0 and observed both resolved turns through the Runner:

```text
judge selection: model=gemini-3.8-flash-medium provider=gemini reasoning_effort=medium
judge selection: model=claude-haiku-4-5 provider=anthropic reasoning_effort=low
--- PASS: TestLoopInferJudgeModelIsIndependentOfPipelineDefaults
```

The second case specifies only `llm_model` and `reasoning_effort` on the loop. Its provider is therefore also evidence that the OpenAI pipeline provider default did not leak into the judge and that provider auto-detection still works.

## Required gate

`go build ./... && go test -count=1 ./...` exited 0. All packages passed on 2026-09-03.
