# Sprint 2: the v2 plan workflow

Rewrite `workflow/plan.yaml` (and its library prompts) as this graph:

- `interview` (codergen, thread `plan`): reads the seed from the workdir,
  asks the human through `tractor ask` until the job is understood, at most
  five questions per visit. Sizing rule for the plan it will write is
  decision 8: SIMPLE one sprint, MEDIUM a flat checklist, LARGE only for
  twelve or more sprints.
- `plan` (codergen, same thread): writes `brief.md`, `checklist.md` in the
  loop ledger format (`loop-node.md`; `chapters/01-ask/sprints.md` is an
  example), and `size.md` holding one word.
- `review` (codergen, fresh thread, `max_visits: 2`): a different model if
  the harness allows. Its prompt is decision 37's question, verbatim: could
  a frontier coding agent, given this plan and nothing else, arrive at a
  correct result; list what is missing that would send it astray; grade
  each finding blocking or note; route fail only for a blocking item.
  Writes `review.md`. Fail routes back to `plan`. When visits are exhausted
  the engine stops the run; that is the ceiling.
- `validate` (tool): the existing `validate-plan` command on
  `checklist.md`; success ends the run, error returns to `plan`.

Keep `medium.yaml` and `large.yaml` as they are. The pipeline must pass
`tractor validate`. Add a `workflow show plan` golden test.
