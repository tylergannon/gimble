# Issue 33 loop evaluator proof

Observed 2026-09-03 in `/Users/tyler/src/tractor`.

## Behavioral claims

- A first arrival with an empty ledger invokes the evaluator; when it writes an open item and returns `not_done`, the Runner dispatches that item and later completes it.
  - Demonstrated by `TestLoopEvaluatorPopulatesEmptyLedgerAndContinues`.
- Every passing lap invokes the evaluator. `not_done` re-reads the checklist, preserves an unchanged next item, or honors reordered/new open items. `done` routes to `on_done`.
  - Demonstrated by `TestLoopPassingLapEvaluatorAgreesAndAdvancesUnchanged`, `TestLoopEvaluatorReordersOpenItemsForNextLap`, `TestLoopAllItemsDoneEvaluatorAddsItemAndContinues`, and `TestLoopEvaluatorDoneRoutesToOnDone`.
- `not_done` with no open item is terminal, and failed validation does not invoke the evaluator.
  - Demonstrated by `TestLoopEvaluatorNotDoneWithoutOpenItemIsTerminal` and `TestLoopEvaluatorDoesNotRunAfterFailedValidation`.
- The infer judge retains its Flash default while the evaluator uses the pipeline default or its evaluator-specific override.
  - Demonstrated by `TestLoopInferJudgeModelIsIndependentOfPipelineDefaults`, `TestLoopEvaluatorUsesPipelineDefaultModelNotInferJudgeDefault`, and `TestLoopEvaluatorExplicitModelSelectionWins`.
- An infer judge and evaluator running in the same loop stage retain separate prompt and response artifacts.
  - Demonstrated by `TestLoopInferItemIsMarkedOnlyAfterJudgePasses`.

The focused Runner execution:

```text
$ go test -count=1 -run 'TestLoop(Evaluator|PassingLapEvaluator|AllItemsDoneEvaluator|InferItemIsMarkedOnlyAfterJudgePasses|InferJudgeModelIsIndependentOfPipelineDefaults)' -v ./engine
--- PASS: TestLoopInferItemIsMarkedOnlyAfterJudgePasses
--- PASS: TestLoopInferJudgeModelIsIndependentOfPipelineDefaults
--- PASS: TestLoopEvaluatorPopulatesEmptyLedgerAndContinues
--- PASS: TestLoopPassingLapEvaluatorAgreesAndAdvancesUnchanged
--- PASS: TestLoopEvaluatorReordersOpenItemsForNextLap
--- PASS: TestLoopAllItemsDoneEvaluatorAddsItemAndContinues
--- PASS: TestLoopEvaluatorNotDoneWithoutOpenItemIsTerminal
--- PASS: TestLoopEvaluatorDoneRoutesToOnDone
--- PASS: TestLoopEvaluatorUsesPipelineDefaultModelNotInferJudgeDefault
--- PASS: TestLoopEvaluatorExplicitModelSelectionWins
--- PASS: TestLoopEvaluatorDoesNotRunAfterFailedValidation
PASS
ok github.com/tylergannon/tractor/engine
```

## Required checks

```text
$ go build ./... && go test -count=1 ./...
exit 0; all packages passed

$ pnpm build
5 pages built; exit 0

$ go vet ./...
exit 0
```
