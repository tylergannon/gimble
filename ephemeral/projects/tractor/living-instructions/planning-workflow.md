# The planning workflow, v2

Design note written 2026-09-02 from the conversation recorded as decisions
37–52 in `decisions.md`. It replaces the single-node `plan` workflow that
chapter 2 built (decision 31) and extends the `large` execution workflow
from chapter 3. Every node type it uses exists today: `codergen`,
`parallel`, `loop`, `tool`, `supervisor`, and the `tractor ask` transport.
No engine change is required; the work is in the `workflow` package, its
prompts, and the ledgers it materializes.

## 1. What it is

`tractor workflow run plan --project <build> --seed <path>` takes a
declarative description of the software and, optionally, the repository it
lives in, and produces a package the execution workflows can run: a brief,
a promise list, a research directory with a routing index, a chapter or
sprint ledger of vertical slices, a validation design per promise, and a
size recommendation. The package is proven before the human approves it:
a loop of fresh reviewers, one question each, finds nothing.

The organizing idea is the promise (decision 37). The interview finds out
what we promise about the work and to whom. Everything downstream is
arranged so that each promise gets a proof a coder cannot satisfy while the
promise is false, and so that nothing in the plan exists that serves no
promise.

## 2. The graph

```
intake ─▶ research ─▶ brief ─▶ research ─▶ halt? ──no──▶ brief …
                                            │
                                           yes
                                            ▼
                                        decompose
                                            ▼
                              validation design loop
                              (one lap per promise:
                               design ─▶ adversarial review)
                                            ▼
                                        assemble
                                            ▼
                              plan review loop
                              (one lap per pass:
                               fresh reviewer ─▶ verdict)
                                            ▼
                                        approve (ask)
                                            ▼
                                         success
```

Supervisors sit outside the walk: `research_auditor` over research,
`scope_cop` over brief, decompose, and validation design, `slice_critic`
over decompose, `proof_skeptic` over the validation design loop. Each runs
on a provider other than the node it watches, with fresh context and
steer authority into the active turn (spec §3.10).

Human gates, all through `tractor ask` with batched question files
(decision 39): the promise interview in `brief`, proof-mechanism questions
inside the validation design loop, and `approve`.

## 3. Nodes

**intake** (codergen, fresh context). Reads the seed and the repository.
Writes `intake.md`: a size guess, the promise candidates it can already
anticipate, and `research/plan.md` with the first open entries, each
naming what the entry must return and which node consumes it. Asks
nothing.

**research** (parallel, at most five branches, then a fan-in). Each branch
takes a segment of the open plan entries and writes leaf files under
`research/` with pinned revisions, licenses, integration points, and
bounded comparisons ("like X but only Y"; never an adoption frame). The
fan-in updates `research/INDEX.md` (the routing tree; incremental after
the first pass) and writes `research/findings.md`: any promise the
research thinks should change, with evidence. Research never edits the
brief and may add a plan entry only through a finding the brief accepts. A
tool node after the fan-in checks the index against the quality bar and
routes to the halt check.

**halt** (tool). Exits 0 when `research/findings.md` is empty and
`research/plan.md` has no open entry; routes to `decompose`. Otherwise
routes to `brief`. The enclosing `loop` node's `max_visits` is the
ceiling, and reaching it is a question to the human, not an exit.

**brief** (codergen, full fidelity across laps). First lap: elicit.
Drafts the promises a user of this thing would expect, from the seed, the
repository, and the research, and asks "do you promise this, and what must
it not imply", with a recommendation each; declined promises become
exclusions. Later laps: read `research/findings.md` and the unresolved
promises, ask only about those. Every lap: a question survives only if its
answer changes a promise, its scope, or its verifier; answers already
given are never asked again. Writes `brief.md` (intent, design principles,
exclusions, promise-adjacent decisions) and `promises.md`, and appends
open entries to `research/plan.md`. A lap that asks nothing and plans
nothing is the terminal one.

**decompose** (codergen). Vertical slices only; every promise is routed to
a slice and every slice serves a promise. MEDIUM writes `checklist.md` as
a flat sprint ledger with an upfront backlog and one sprint doc per item.
LARGE writes `checklist.md` as the chapter ledger (decision 35) with a
chapter doc (pyramid index, vector, review posture, non-goals) and a
sprint ledger per chapter; sprint ledgers may start with a backlog. The
chapter ledger is durable (decision 44); it is edited only with a reason.

**validation design loop** (loop over `validation/ledger.md`, one item per
promise; body `design` → `review` → back). `design` (codergen) chooses the
archetype (decision 42) and writes, under `validation/<promise>/`: the
user story a verifier follows, the evidence specification (what is
captured, where), the UI sketch when a screen is involved, and, for a
universal promise, the holdout sample under the XDG state root in a
random-token directory whose path is recorded only in the verifier prompt
the workflow will later materialize (decision 43). It then fills the
promise's checklist item with `command` (the required checks) and `infer`
(the judgment over the captured evidence). Where the mechanism of proof is
not derivable it asks the human. `review` (codergen, other provider, fresh
context) is told only the promise and the design and answers three
questions: can a coder satisfy this while the promise is false; is any
check trivially true; is the design stricter than the promise. It writes
`validation/<promise>/verdict.md`. The ledger item's own `command` is a
check that the verdict says pass, so the engine re-enters the item on a
failed review with the reviewer's notes in the frame.

**assemble** (codergen, cheap). Writes `recommendation.md` (decision 32)
and `plan-review/ledger.md` from the built-in pass list plus any passes
the project adds. Materializes nothing else; every artifact already exists.

**plan review loop** (loop over `plan-review/ledger.md`, one item per
pass; body `reviewer`). Each lap starts a fresh reviewer on a provider
other than the planner, whose prompt names the pass question and the
package and nothing else. It writes `plan-review/<pass>/verdict.md` with
findings that name the owning node. The item's `command` checks the
verdict. On failure the frame carries the findings back; the owning node
is re-run through an ordinary edge from `reviewer`'s failure route, and
the loop re-selects the pass. The built-in passes, in order:

1. **traceability**: every promise reaches a slice; every slice serves a
   promise. Scriptable.
2. **consistency**: one name per concept; no rule stated twice with drift
   across brief, chapter docs, sprint docs.
3. **slicing**: every slice is vertical and exercisable when done.
4. **proof quality**: across promises, no mechanism a coder can satisfy
   while the promise is false; holdouts present for every universal
   promise; nothing trivially true or stricter than its promise.
5. **scope**: nothing exceeds the promises; exclusions respected.
6. **executability**: every sprint fits one agent turn, every command
   runs from the workdir, every path resolves, every ledger parses. This
   is today's `validate-plan`, kept as the last pass.

**approve** (codergen, one `tractor ask`). Shows the human the package
summary and the six verdicts. Yes routes to `success`. No routes to
`brief` with the reason as the only open question.

## 4. Sizes

| Size | Path through the graph | Next |
|---|---|---|
| SIMPLE | intake, one research lap, brief, one validation lap, review, approve | "Execute the plan yourself." |
| MEDIUM | full graph; decompose into sprints | `tractor workflow run medium --project <build>` |
| LARGE | full graph; decompose into chapters | `tractor workflow run large --project <build>` |

Intake guesses; the human confirms at the brief gate; `recommendation.md`
records the confirmed size.

## 5. Execution side

The package is only honest if execution proves what planning designed.
Two nodes are added to the `large` chapter lap (and `replan` to `medium`):

```
chapters ─▶ plan ─▶ sprints ─▶ implement ─▶ replan ─▶ sprints … ─▶ verify ─▶ chapters
```

**replan** (codergen, cheap model, fresh context): after every implement
lap, reads the sprint ledger, the last validation, and the repository, and
edits open sprint items only. It never touches the chapter ledger; a
sprint that finds the chapter wrong asks the human.

**verify** (codergen, other provider than the coder, fresh context, tools):
runs once per chapter after its sprint loop exits. Reads the validation
design and the holdout path, operates the software itself, captures its
own evidence under the run directory, and writes
`validation/<chapter>/verdict.md`. The chapter item's `command` checks that
verdict; a chapter is *proven* only through this leg. Sprint items are
*demonstrated* by their own `command` and `infer`.

## 6. Project layout

```
ephemeral/projects/<build>/
  intake.md
  brief.md
  promises.md
  recommendation.md
  checklist.md                  chapter ledger (LARGE) or sprint ledger (MEDIUM)
  chapters/NN-<slug>/           CHAPTER.md, sprints.md, SPRINT-NN.md   (LARGE)
  research/                     plan.md, findings.md, INDEX.md, leaves
  validation/
    ledger.md                   one item per promise
    <promise>/                  story.md, evidence.md, sketch.*, verdict.md
  plan-review/
    ledger.md                   one item per pass
    <pass>/verdict.md
  interview/                    NNNN.md, NNNN.answer.md
```

Holdouts: `$XDG_STATE_HOME/tractor/holdouts/<build>-<token>/`.

## 6a. The library (decision 53)

The workflows are content. Everything an agent is told lives as a file
under `workflow/library/`, embedded with `embed.FS`, rendered with
`text/template` from the workflow `Parameters`. `workflow.go` keeps the
graph surgery (`Build`, path resolution, validation) and nothing an editor
would want to change.

```
workflow/library/
  README.md            what lives here; how prompts compose; how to test an edit
  workflows/           plan.yaml, medium.yaml, large.yaml (moved from workflow/)
  prompts/             one file per node, Go text/template
    plan/              intake.md research.md brief.md halt.sh decompose.md
                       design.md review.md assemble.md approve.md
    medium/            implement.md replan.md
    large/             plan.md implement.md replan.md verify.md
  supervisors/         research_auditor.md scope_cop.md slice_critic.md proof_skeptic.md
  passes/              01-traceability.md … 06-executability.md
                       (the plan-review ledger is generated from this directory)
  doctrine/            the teachings, included by prompts by name:
                       promises.md elicit-then-prune.md question-files.md
                       promise-adjacent-seams.md vertical-slices.md
                       proof-not-theater.md validation-archetypes.md
                       prior-art.md research-leaf.md chapter-doc.md sprint-doc.md
                       pyramid-index.md reviewer-independence.md
  templates/           skeletons the planner fills: brief.md promises.md
                       CHAPTER.md SPRINT.md story.md evidence.md verdict.md
                       recommendation.md ledger.md
```

Rules:

- A prompt includes doctrine with `{{doctrine "promises"}}`. A teaching is
  edited in one place and every prompt that cites it changes.
- Tests render every template with representative parameters, fail on an
  unreferenced doctrine or template file, and fail on any prompt that is
  a Go string.
- `tractor workflow show <name> [--project …]` prints the materialized
  prompts and supervisor briefs, so an editor sees exactly what agents
  will see before committing a content change.
- The skill bundle (`skills/tractor`) and the docs site teach the same
  things from the same files where they overlap (question-file format,
  ledger format, recommendation contract), so there is one source.
- The raw corpus under `sources/` stays as provenance. The library holds
  the distilled teaching, a page each, with a pointer back.

## 7. Not in this version

Budgets (decision 45). A secret holdout location. Multi-level supervision.
Fan-out drafts of chapter docs. Any engine change. The web client.

## 8. Claims to demonstrate before this replaces `plan`

1. Given a one-paragraph seed and an empty repository, the workflow asks
   about at least one promise the seed did not mention, and a declined
   promise appears under exclusions in `brief.md`.
2. The brief and research loop halts by the tool node, not by
   `max_visits`, on a seed whose research produces one promise-changing
   finding; the finding is asked, not applied.
3. Every promise in `promises.md` has a `validation/<promise>/verdict.md`
   that says pass, and at least one universal promise has a holdout
   outside the workdir and outside the run directory.
4. A validation design that a reviewer rejects is re-entered with the
   reviewer's notes in the frame and passes on the next lap.
5. All six review passes end `done: true` in `plan-review/ledger.md`, each
   marked by the engine after a verdict from a provider other than the
   planner's.
6. The package runs: `tractor workflow run medium` (or `large`) on the
   output reaches `COMPLETED`, and a chapter's `done: true` follows a
   `verify` verdict, never a sprint count alone.
7. `scope_cop` delivers at least one steer during the run, recorded in the
   timeline, and the steered turn's output changes.
