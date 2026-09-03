# Sprint 5: docs and skill

Teach the library and `show` wherever the workflows are already taught.
Anchors: `research/workflow-package-inventory/migration-inventory.md` §7.

- `docs/spec.md` §3.1.2: a subsection "The workflow library" stating the
  layout in one paragraph, pointing at `workflow/library/README.md` for
  the template contract (the README is authoritative; the spec does not
  restate it), naming the render and orphan tests, and teaching
  `workflow show` with `--node --raw`, `--values`, and `--stage`. State
  plainly that the raw output never reproduces frames or `$goal`, and
  that `--stage --goal` expands `$goal` for the comparison only.
- `src/content/docs/planning.md`: a short section "Editing what the
  planner is told" with the three-step loop: edit a page, run the tests,
  run `show`.
- `skills/tractor/SKILL.md` and `llms.txt`: one paragraph each; the skill
  description mentions `show`.
- `workflow/library/README.md` is the authoritative contract; the docs
  point at it rather than restating it.
- `README.md` "Start with a plan": one sentence and a link.

No claim the docs make may contradict `workflow show --help` or the
README. The `infer` judge on the ledger item checks exactly that.
