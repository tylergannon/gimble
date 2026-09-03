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
intake ─▶ research ─▶ index_gate ─▶ brief ─▶ research ─▶ index_gate ─▶ halt? ──no──▶ brief …
                                            │
                                           yes
                                            ▼
                                        decompose
                                            ▼
                              validation design loop
                              (one lap per promise:
                               design ─▶ review ─▶ pass | fail)
                                            ▼
                                        assemble
                                            ▼
                              plan review loop
                              (one lap per pass:
                               fresh reviewer ─▶ pass | fail)
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
inside the validation design loop, `approve`, and `ceiling` when the
brief/research loop exhausts its visits (decision 50 as amended).

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
fan-in updates `research/INDEX.md` (the routing index; incremental after
the first pass) and writes `research/findings.md`: any promise the
research thinks should change, with evidence. Research never edits the
brief and may add a plan entry only through a finding the brief accepts. A
tool node, `index_gate`, after the fan-in checks the index against the
quality bar; it routes to `brief` while no brief lap has run yet (no
`brief.md` exists), so the first interview always happens, and to
`halt` afterwards.

**halt** (tool). Exits 0 when `research/findings.md` is empty and
`research/plan.md` has no open entry; routes to `decompose`. Otherwise
routes to `brief`. Decision 38's stopping rule governs what `brief`
asks; the loop's exit is this predicate, which the brief's last lap
meets by leaving no finding and no open entry behind, so a lap that
asked nothing and planned nothing is followed by one research pass
that finds nothing and then the halt. `brief`'s `max_visits` is the
ceiling; its exhaustion
edge leads to `ceiling`, a codergen node that asks the human one
question (continue with a higher ceiling, or stop) through `tractor
ask` and routes back to `brief` or to `halted` (a terminal that ends the
run without a package; only `approve` reaches `success`) as answered,
so the ceiling is a human gate, not a failed run (spec: an exhausted node with
no escalation edge fails the run).

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
chapter doc (pyramid index, vector, review posture, non-goals; design
direction, architecture principles, sprint horizon, and planning notes
when the chapter has them) and a
sprint ledger per chapter; sprint ledgers start with a backlog. A
chapter planned before its predecessor is built carries its backlog as
prose in the ledger body and an empty item list; the `plan` node turns
the prose into items when the chapter is entered (decision 44 as
applied to this project's chapters 5 and 6). The
chapter ledger is durable (decision 44); it is edited only with a reason.

**validation design loop** (loop over `validation/ledger.md`, one item per
promise; body `design` → `review` → back). `design` (codergen) chooses the
archetype (decision 42) and writes, under `validation/<id>/`: the
user story a verifier follows, the evidence specification (what is
captured, where), the UI sketch when a screen is involved, and, for a
universal promise over a set too large to check whole, the holdout
sample under the XDG state root in a random-token directory whose path
is stored in the run's private workflow state and disclosed only to the
`verify` prompt the workflow later materializes (decision 43). It then fills the
sprint item that will demonstrate the promise, in MEDIUM's sprint
ledger or in the chapter's sprint ledger for LARGE, with `command` (the
required checks) and `infer` (the judgment over the captured evidence).
A chapter item carries only the required checks as `command`, because
the chapter is proven by `verify`, not demonstrated by a gate; the
validation ledger's own item stays without a command. Where the mechanism of proof is
not derivable it asks the human. `review` (codergen, other provider, fresh
context) is told only the promise and the design and answers three
questions: can a coder satisfy this while the promise is false; is any
check trivially true; is the design stricter than the promise. It routes
pass to the loop node and fail to `design`. The ledger item has no
`command`: only the pass edge returns to the loop, so a returning lap is
the pass and the engine marks the item. The reviewer's notes are ordinary
files under `validation/<id>/`, read by `design` on the next lap.

**assemble** (codergen, cheap). Writes `recommendation.md` (decision 32)
and `plan-review/ledger.md` from the built-in pass list plus any passes
the project adds. Materializes nothing else; every artifact already exists.

**plan review loop** (loop over `plan-review/ledger.md`, one item per
pass; body `reviewer`). Each lap starts a fresh reviewer on a provider
other than the planner, whose prompt names the pass question and the
package and nothing else. It routes pass to the loop node, which marks the
pass done, or fail to the owning node (brief, decompose, or design), whose
edge returns to `reviewer` for the same pass; only a passing reviewer
returns to the loop, since a commandless item passes when its lap
returns (loop-node.md section 2). The
reviewer's findings are ordinary files under `plan-review/<pass>/`. The
built-in passes, in order:

1. **traceability**: every promise reaches a slice; every slice serves a
   promise. Scriptable.
2. **consistency**: one name per concept; no rule stated twice with drift
   across brief, chapter docs, sprint docs.
3. **slicing**: every slice is vertical and exercisable when done.
4. **proof quality**: across promises, no mechanism a coder can satisfy
   while the promise is false; holdouts present for every universal
   promise whose set is not checked exhaustively; nothing trivially true or stricter than its promise.
5. **scope**: nothing exceeds the promises; exclusions respected.
6. **executability**: every sprint fits one agent turn, every command
   runs from the workdir, every path resolves, every ledger parses. This
   is today's `validate-plan`.
7. **holistic**: rubric-free. "Would you accept this as the plan for
   this product, and what one thing would stop you." Last, because
   single-criterion judges credit mere mention and miss trade-offs
   between criteria (research F2).

**approve** (codergen, one `tractor ask`). Shows the human the package
summary and the seven pass outcomes. Yes routes to `success`. No routes to
`brief` with the reason as the only open question.

## 4. Sizes

| Size | Path through the graph | Next |
|---|---|---|
| SIMPLE | the full graph; `decompose` writes a one-item sprint ledger; the validation loop runs one lap per promise and the review loop runs all seven passes, as for any size | "Execute the plan yourself." |
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
design and, for a universal promise with a holdout, the holdout path
the run's private workflow state carries (written there by the design
lap and rendered only into this prompt; a chapter whose promises are all
scenarios has none),
operates the software itself, captures its
own evidence under the run directory, and routes pass to the chapters
loop or fail to the sprint loop (or to a human question). A chapter is
*proven* only through this leg; its item carries the required checks as
`command` and nothing else. Sprint items are *demonstrated* by their own
`command` and `infer`.

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
    <id>/                       design.md (story, evidence, validator, not proven), sketch.*, review notes
  plan-review/
    ledger.md                   one item per pass
    <pass>/                     reviewer findings
  interview/                    NNNN.md, NNNN.answer.md
```

Holdouts: `$XDG_STATE_HOME/tractor/holdouts/<build>-<token>/`.

## 6a. The library (decision 53)

The workflows are content. Every prompt body, doctrine page, supervisor
brief, pass, and skeleton lives as a file under `workflow/library/`
(frames and `$goal` are the engine's, added at run time), embedded with `embed.FS`, rendered with
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
  passes/              01-traceability.md … 07-holistic.md
                       (the plan-review ledger is generated from this directory)
  doctrine/            the teachings, included by prompts by name:
                       promises.md elicit-then-prune.md question-files.md
                       promise-adjacent-seams.md vertical-slices.md
                       proof-not-theater.md validation-archetypes.md
                       prior-art.md research-leaf.md chapter-doc.md sprint-doc.md
                       pyramid-index.md reviewer-independence.md ledger.md
  templates/           skeletons the planner fills: brief.md promises.md
                       CHAPTER.md SPRINT.md design.md
                       recommendation.md ledger.md
```

Rules:

- A prompt includes doctrine with a `doctrine "promises"` action. The
template delimiters are non-default, chosen by the coder in chapter 4
sprint 1 before any page is installed in the library, because doctrine
text contains `{{` (research F5); pages written earlier avoid the
recommended pair.
  A teaching is edited in one place and every prompt that cites it
  changes.
- Tests render every template with representative parameters, fail on a
  doctrine page no rendered agent-facing library file (prompt body,
  supervisor brief, or pass) references, and fail on any prompt
  that is a Go string.
- `tractor workflow show <name> [--project …]` prints each node's payload (a
  prompt, a tool command, or a checklist path) and each supervisor's
  brief exactly as `Build` materialized them for those parameters; `--stage <dir>` diffs against a real stage with the frame
  stripped. Frames and `$goal` are engine additions the raw output never
  reproduces (research F1); `--stage --goal` expands `$goal` the
  engine's way for the comparison only.
- The skill bundle (`skills/tractor`) and the docs site teach the same
  things from the same files where they overlap (question-file format,
  ledger format, recommendation contract), so there is one source.
- The raw corpus under `sources/` stays as provenance. The library holds
  the distilled teaching, a page each, with a pointer back.

## 6b. Models (decision 54)

`workflow/library/models.yaml` maps roles to provider, model, and
reasoning effort, and is deployed with the binary. The provider is always
explicit; nothing relies on name-based detection (research R4). A role may carry
`not: <role>` meaning its provider must differ from that role's at
materialization. Initial table:

| Role | Provider, model, effort | Constraint |
|---|---|---|
| coder (`implement`) | codex, gpt-5.6-sol, high | |
| planner nodes (intake, brief, decompose, design, approve) | claude, claude-fable-5-1, high | |
| research branches and fan-in | claude, `sonnet` alias, medium | |
| validation reviewer, plan-review reviewer | codex, gpt-5.6-sol, high | not the provider of the node judged (the planner) |
| `verify` | claude, claude-fable-5-1, high | not the provider of the coder |
| supervisors | codex, gpt-5.6-sol, high | not the provider of the node watched (all watched nodes are claude) |
| `assemble`, `replan`, `infer` judges | claude, `sonnet` alias, medium | |
| `halt`, `index_gate` | tool nodes; no model | |
| answerer | human, or the calling agent | |

## 7. Not in this version

Budgets (decision 45). A secret holdout location. A holdout in this
project's own proof (none of its sets is large enough). Multi-level supervision.
Fan-out drafts of chapter docs. Any engine change. The web client.

## 8. Claims to demonstrate before this replaces `plan`

1. Given a seed that names three features and an empty repository, the
   workflow asks
   about at least one promise the seed did not mention, and a declined
   promise appears under exclusions in `brief.md`.
2. The brief and research loop halts by the tool node, not by
   `max_visits`, on a seed whose research produces one promise-changing
   finding; the finding is asked, not applied.
3. Every promise in `promises.md` ends `done: true` in the validation
   ledger after a `review` turn routed pass. (No holdout in this
   project's proof; a universal promise over a set the verifier checks
   exhaustively needs none, decision 42 as amended.)
4. A validation design that a reviewer rejects is re-entered with the
   reviewer's notes available (read from their files under
   `validation/<id>/`) and later passes.
5. All seven review passes end `done: true` in `plan-review/ledger.md`, each
   marked by the engine after a reviewer on a provider other than the
   planner's routed pass.
6. The package runs: `tractor workflow run large` on a LARGE output
   reaches `COMPLETED`, and every chapter's `done: true` follows a
   `verify` turn routing pass, never a sprint count alone (`medium` has
   no chapters and proves nothing here).
7. `scope_cop` delivers at least one steer during the run, recorded in the
   timeline, and the steered turn's output changes.
8. `show --stage` reports no diff against a stage built to the engine's
   frame shape (chapter 4) and against recorded stages of real runs
   (chapter 6), and the orphan walk fails when a doctrine page is added
   that no rendered agent-facing file cites.
