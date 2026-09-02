# Sprint 1: the embedded `plan` workflow

Add the built-in workflow library and its first entry, `plan`. Keep the library
independent of Cobra: it owns names, descriptions, graph construction, and the
planning artifact contract. The CLI arrives in sprint 2.

## Registry and graph

- Embed the workflow definition in the binary and parse it through the same
  `graph` package used for pipeline files. Invalid embedded definitions must
  fail a test, not surprise a caller at runtime.
- The library must be able to list definitions in stable name order and build
  one by name from explicit parameters. Only `plan` is runnable in this
  chapter; `medium` and `large` are the fixed names chapter 3 will add.
- Materializing `plan` receives the project name, seed path, workdir, and the
  current executable path. Quote generated paths safely. Do not add general
  graph variables: decision 25 still leaves `$goal` as the only graph
  substitution.
- The graph is one codergen node followed by one tool node. The codergen has a
  deliberately long timeout and keeps the same context if artifact validation
  routes back. The tool mechanically validates the outputs, routes to success
  on exit 0, and routes back to the codergen on error. There is no second agent
  or human-routing node.

## Planner contract

The codergen reads the seed and relevant repository context, then owns the
whole interview and the write:

- It uses `tractor ask` in the configured interview directory, one Markdown or
  HTML question at a time. It asks only when the answer can change the
  contract. The depth is proportional to the apparent size.
- It explicitly checks intent, scope and non-goals, constraints, and the
  **Definition** of success. Once two passes over those dimensions reveal only
  details derivable from the seed, answers, or repository, it stops asking.
- It writes only under `ephemeral/projects/<project>/`: `brief.md`,
  `checklist.md`, and `recommendation.md`.
- `checklist.md` uses `loop-node.md` section 2 exactly. Each item has a stable
  `name` and observable `check`; it has a real `command` and/or `infer` wherever
  the repository makes one knowable. The planner never writes `done`.
- `recommendation.md` has a heading plus `Size: SIMPLE|MEDIUM|LARGE`, a short
  rationale, and `Next:`. SIMPLE is at most one sprint and says `Execute the
  plan yourself.` MEDIUM is more than one sprint but less than two chapters and
  names `tractor workflow run medium --project <project>`. LARGE is multiple
  chapters and names the equivalent `large` command.

Put mechanical artifact validation in a small testable Go function. It must at
least require all three non-empty files, parse `checklist.md` with package
`checklist`, require a non-empty item list, reject agent-written `done`, and
validate the recommendation fields and size-specific next action. The tool node
may reach that function through a hidden implementation command added in
sprint 2; do not use fragile grep as the parser.

## Tests

Tests named `TestBuiltInPlan`, `TestPlanArtifacts`, and `TestRecommendation`
demonstrate stable lookup, the exact
codergen-to-tool-to-codergen/success shape, the long-lived planner prompt, valid
and invalid loop checklists, and every recommendation size. Run all required
BUILD.md gates before committing. Add `check-sprint-01.sh` beside this document;
it runs those tests with `-count=1 -v` and fails unless the output proves all
three named tests actually ran and passed, so an empty or renamed test suite
cannot satisfy the ledger command.
