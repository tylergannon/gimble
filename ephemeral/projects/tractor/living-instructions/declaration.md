# Declaration: the planning workflow, v2

Written 2026-09-02 by hand, the way the v2 planner is meant to do it: a
declaration of the product, the state that exists when it is done, the
promises we make about it and how each is proven, the seams a promise
crosses, and the chapters that get there. Design detail is in
`planning-workflow.md`; the rulings are decisions 37–53.

## 1. Declaration

Tractor carries a built-in planning workflow. Given a declarative
description of some software, and optionally the repository it lives in,
the workflow interviews the caller about the promises the work will make,
researches prior art into a local library the plan can cite, decomposes
the work into vertical slices as chapters and sprints, designs a proof for
each promise that a coder cannot satisfy while the promise is false,
reviews the plan with independent reviewers one question at a time, and
hands the human an approved package. The built-in execution workflows run
that package to a proven state: sprints demonstrated by their own checks,
chapters proven by an adversarial verifier that operates the software.
Every prompt, supervisor brief, review pass, doctrine page, and artifact
skeleton the workflows use is editable content embedded in the binary, so
each release can improve how Tractor plans and proves work without
touching Go.

## 2. Final state

When this is done, the repository contains:

- `workflow/library/` with graphs, prompts, supervisors, passes, doctrine,
  and templates as files; no prompt in a Go string; a render test; and
  `tractor workflow show <name>`.
- `plan` v2: intake, the brief/research loop with its tool-node halt,
  decompose, the validation design loop, assemble, the plan review loop,
  approve; four supervisors.
- `medium` and `large` with `replan` after every implement lap and, in
  `large`, `verify` at chapter exit reading a holdout outside the workdir.
- Docs site, spec, skill bundle, and `llms.txt` teaching the above from
  the same library files where they overlap.
- A proof record under `proof/planning-v2/` in the established format:
  identity block, claims, scope check, findings.

## 3. Promises

Each promise: the statement, what it must not imply, the archetype
(decision 42), and the verifier. Required checks (`go build`, `go vet`,
`go test`, `golangci-lint`) apply to every chapter and prove nothing here.

| # | Promise | Must not imply | Archetype | Verifier |
|---|---|---|---|---|
| P1 | Given a seed that names three features and an empty repository, `plan` asks about at least one promise the seed did not name, and a declined promise appears under exclusions in `brief.md`. | That the planner anticipates every promise a user cares about. | Scenario | Nested run with a scripted answerer that declines the first anticipated promise; inspector checks `brief.md`. |
| P2 | On a seed whose research yields one promise-changing finding, the brief/research loop halts through the tool node, and the finding was asked, not applied. | That research converges on arbitrary seeds. | Scenario | Seed with a planted finding source in the token cache; timeline shows `halt` routing to `decompose` and one `QuestionAsked` citing the finding. |
| P3 | Every promise in `promises.md` is marked done in `validation/ledger.md` by the engine after a `review` turn on a provider other than the planner's routed pass. | That the designs are good, only that they were reviewed independently. | Universal over promises | Inspector over the ledger and the run's events (provider and routing per stage). No holdout: the set is small and checked exhaustively. |
| P4 | A validation design the reviewer routes to fail is re-designed with the reviewer's notes available and later routes to pass. | Anything about how often designs are rejected. | Scenario | A test seed whose first design lap is instructed to propose a trivial check; the timeline shows `review` routing to `design`, then to the loop; the item ends `done: true`. |
| P5 | All six review passes end `done: true` in `plan-review/ledger.md`, each marked by the engine after a fresh reviewer on another provider routed pass. | That the passes catch every defect. | Universal over passes | Inspector over ledger and events. |
| P6 | `tractor workflow run large` on a v2 package reaches `COMPLETED`, and every chapter's `done: true` is preceded in the timeline by a `verify` turn for that chapter routing pass. | That the software the package describes is good. | Scenario | Nested run on a small package; timeline order check. |
| P7 | `scope_cop` delivers at least one steer during a planning run, and the steered turn's output differs from what it was writing before the steer. | That supervisors improve plans. | Scenario | Timeline `steer` verdict with a delivered disposition; diff of the stage output before and after. |
| P8 | No built-in prompt lives in a Go string; every doctrine page is referenced by at least one prompt; `tractor workflow show plan` prints, for each node, text byte-equal to the prompt the engine materialized for that node in a real run. | That the content is well written. | Universal over library files | Render test plus a diff between `show` output and `stages/*/prompt.md` minus the frame block. Exhaustive, no holdout. |
| P9 | An agent reading only the docs site, spec, and skill bundle runs `workflow run plan`, `workflow show`, and `ask` correctly, including where the interview directory and the holdout come from. | That the docs are complete. | Scenario | `infer` judge over the docs, failing on any claim the `--help` text contradicts (the chapter 1 sprint 3 pattern). |
| P10 | For two small product descriptions written before chapter 5 starts, `plan` ends with an approved package that `validate-plan` accepts and `medium` or `large` runs to `COMPLETED`. | Generality beyond seeds of that size. | Scenario | Both seeds are known; the end-to-end run is the proof. |

No promise here carries a holdout. Two seeds cannot be overfitted in a way
a holdout would catch, and every other set is checked exhaustively. The
holdout stays in the workflow for projects whose sets are large.

## 4. Seams a promise crosses

Per decision 40, only these are contracts. Everything else is internal
and the implementer owns it.

| Seam | Parties | Contract |
|---|---|---|
| Package layout | `plan` → `medium`/`large`, the human, future releases | `planning-workflow.md` §6. Persisted format; changes need a reason and a docs edit. |
| Checklist item | ledgers ↔ engine | `loop-node.md` §2. Exists; unchanged. |
| Question and answer files | agents ↔ humans and the future web client | `tractor ask`/`answer` as built; batched questions (decision 39) are a convention inside the file, not a format change. |
| Holdout handoff | design lap → `verify` prompt materialization | Path under `$XDG_STATE_HOME/tractor/holdouts/<build>-<token>/`, recorded in the run's private workflow state, rendered only into `verify`. |
| Library template contract | content editors ↔ the workflow package | `{{doctrine "name"}}`, `{{template "name"}}`, and the `Parameters` fields templates may read. Documented in `workflow/library/README.md`; the render test is its check. |
| Supervisor digests and verdicts | engine ↔ supervisor turns | Spec §3.10. Exists; unchanged. |

## 5. Chapters

Three chapters, numbered after the three built. Each chapter carries a
docs-and-skill sprint, as chapter 1 did, so the docs never lag by more
than one chapter. Chapter 4's sprints are written up front as the
hand-written standard; later chapters start with a backlog sketch and are
re-planned each lap (decision 44). No engine change anywhere; chapters 5
and 6 are content plus workflow-package Go.

### Chapter 4: the library

Promise P8. Everything later is content, so this comes first.

1. Move the three graphs and three prompts into `workflow/library/`,
   embedded, rendered with `text/template`; `Build` unchanged from the
   outside; existing tests pass. Check: no `fmt.Sprintf` prompt remains;
   `go test ./workflow/...`.
2. `tractor workflow show <name> [--project …]`; render test that every
   template renders and every doctrine and template file is referenced.
   Check: `show plan` output equals a real run's `prompt.md` minus frames.
3. Doctrine pages, distilled from `sources/` one page each with a
   provenance pointer, and the artifact skeletons. Written by Claude, not
   the coder (see question 5). Check: page list matches
   `planning-workflow.md` §6a; `infer` judge that each page is under a
   screen and cites its source.
4. Docs and skill. Check: the chapter 1 sprint 3 pattern.

### Chapter 5: the planner

Promises P1 to P5, P7. Backlog sketch: `models.yaml` and per-role model
resolution; intake node and first research plan; research branches,
fan-in, index, quality tool node; brief node with elicit-then-prune and
batched questions; the halt tool node and the loop; decompose for MEDIUM
and LARGE; validation ledger generation and the `design` node with both
archetypes and the holdout writer; the `review` node with its pass and
fail edges; the pass files, the generated review ledger, the `reviewer` node
and the route back to the owning node; assemble and approve; the four
supervisors; docs and skill.

### Chapter 6: execution and the live proof

Promises P6, P9, P10. Backlog sketch: `replan` in `medium` and `large`;
`verify` in `large` with the holdout handoff; the two seeds; the live
end-to-end proof and the proof record; docs and skill; closeout.

## 6. Size and recommendation

LARGE: three chapters, each more than one sprint. Next:
`tractor workflow run large --project tractor/living-instructions` once
chapter 4 exists; until then, the existing `pipeline.yaml` shape (chapters
→ plan → sprints → implement) with the coder and reviewer as before. See
question 4.

## 7. Questions for Tyler

Batched, with a recommendation each (decision 39).

1. **Chapter split.** Ruled: three, since nothing in chapters 5 and 6
   is engine machinery.
2. **Roles.** Ruled: a `models.yaml` in the library maps roles to
   provider, model, and reasoning effort, with a `not` constraint for
   independence, deployed with the binary. Initial table in
   `planning-workflow.md` §6b.
3. **Holdout seeds.** Withdrawn; see P10. No holdout in this project.
4. **Dogfood timing.** Ruled: run the v2 algorithm by hand on this
   project now, Claude as the planner nodes, Tyler as the human gate,
   subagents as reviewers and research branches, and automate portions
   as they are built.
5. **Doctrine authorship.** Claude writes the doctrine pages and skeletons
   as chapter 4 sprint 3, from the extractions already made; the coder
   wires them. Recommend yes; the pages are the design, and writing them
   through a coder adds a lossy hop.
6. **Verdict file.** Withdrawn. Review and verification are codergen
   nodes with pass and fail edges; routing is the verdict, the item has
   no command, and the reviewer's notes are ordinary files.
