# Issue #37 — Loop node: infer judge model is set independently, default Gemini 2.7 Flash

Task brief for the implementer. The text below is the issue verbatim.
It was filed against branch `worktree-goal-gates`, which is now merged;
the code it names is on `main` at the same paths.

## Definition of done

- The required behavior below holds in the code.
- The tests named below exist and pass.
- The docs named below say what the code now does.
- `go build ./... && go test -count=1 ./...` exits 0.
- The change is made in this checkout, at the path the run's workdir names.
  Do not run `git worktree add`, do not create a branch, do not commit, and do
  not push: the run's verification gate executes in the workdir, so work done
  anywhere else is not checked. This overrides any worktree or checkpoint step
  in the repository's agent protocol.


## Implementer note — corrected judge model id (2026-09-03)

The issue names "Gemini 2.7 Flash". No such model exists, and the version
number was never the decision: Tyler means the latest Gemini Flash.

The `agy` harness offers Gemini 3.6, 3.7 and 3.8 Flash, each as an id with the
reasoning effort baked in (`gemini-3.8-flash-low`, `-medium`, `-high`);
`agy models` is the list, and `harness/agy` skips `--effort` for an id that
already carries one. Gemini 3.8 Flash is the latest. Answering through `agy`
was verified on this machine on 2026-09-03.

Use **`gemini-3.8-flash-medium`** with provider `gemini` as the judge default.
Medium, not low: the judge decides pass or fail on screenshot evidence, which
is worth a little reasoning, and Flash is cheap at either effort. The `flash`
alias in `internal/modelalias` names that id. Everything else in the issue
stands.

---

Branch `worktree-goal-gates`, `engine/loop.go` `judge`, `graph/parse.go` `applyDefaults`.

## Bug

The infer judge runs on the loop node's `llm_model` / `llm_provider` / `reasoning_effort`, and those inherit the pipeline `defaults` block. So the judge runs on whatever model the body runs on unless every loop node overrides it. The judge is meant to be a cheap model that looks at evidence files, often screenshots.

## Required behavior

The judge model is set independently of the body and has its own default.

- Default judge model: Gemini 2.7 Flash, provider `gemini` (routes to the `agy` harness).
- A loop node may override with its own `llm_model` / `llm_provider` / `reasoning_effort`.
- Pipeline `defaults` do not apply to the judge.

## Changes

- `graph/parse.go`: stop inheriting `LLMModel`, `LLMProvider`, `ReasoningEffort` from `defaults` onto `LoopNode`. Keep `Timeout` inheritance.
- `engine/loop.go` `judge`: when the loop node sets no model, use the judge default constant instead of the runner's `DefaultModel` / `DefaultProvider`.
- `internal/modelalias`: add a `flash` alias for the Gemini 2.7 Flash id so the default can be named.
- Test: a pipeline with `defaults.llm_model` set and a loop node with none runs the judge on the default judge model; an explicit loop-node model wins.
- Docs: `docs/spec.md` loop node field table; `src/content/docs/loops.md`.
