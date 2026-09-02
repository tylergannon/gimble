# Living instructions: source material

Collected 2026-09-01 for the "Tractor knows how to use itself" initiative
(see `brief.md` and the GitHub issue it became). Everything an author of
the intake / contract / size-matched-loop flow should read is copied here
verbatim so it is findable later without chasing three repos.

## sources/agents/ — tylergannon/agents `skills/`

Copied from `~/src/agents` at main `bb98afb9974fd900edece8ca44fd4c01a4ca4eef`
(2026-08-20). `grilling` and `writing-great-skills` come from
`~/.agents/skills/` (borrowed skills installed on this machine).

| Skill | Why it matters here |
|---|---|
| `proof-of-work` | The contract is a claims list: observable behaviors, written before work, not a command checklist. Closeout names proved SHA, met/unmet claims. |
| `spec-writing` | Minimal spanning spec; membership test; definition of done as claims; proportionality by blast radius. |
| `spec-authoring` | The interview: grill HITL for goals, decompose, harden contracts, **stopping rule** (two rounds of only-derivable answers). |
| `spec-review` | Mechanical review passes for the produced spec. |
| `slice-design` | Vertical slices; acceptance (main point 100%, no demonstrated bugs, 90–95% items); handoff contract; exit valves. |
| `cell-lifecycle` | Maintenance vs rebuild; condemn triggers; catalytic converter. |
| `adversarial-review`, `request-adversarial-review`, `consensus` | Reviewer owns scope; re-review the whole target; never confirm-my-fix; three unresolved exchanges → HITL. |
| `write-prompts` (+ `references/loop-prompts.md`) | Loop anatomy (nine components) and six halt conditions; prompt hygiene for anything compiled into the binary. |
| `grilling` | One question at a time; recommend an answer; look up facts instead of asking. |
| `writing-great-skills` | How to make the eventual skill/prompt text predictable and short. |

## sources/diffusioninc/ — diffusioninc/skills (whole repo)

Tarball of main `2830b0024265d16fcb9fa6fe003b69ea685cc57d` (2026-08-27).
The `.agents/` mirror was dropped (byte-equivalent to `.claude/` modulo
path substitutions per their AGENTS.md). Includes their `docs/sprints/`
example run (intent, three drafts, three critiques, merge notes, final
sprint), the easyloop visualizer, the skills validator, and their own
`.promises/` state.

| Skill | Why it matters here |
|---|---|
| `df-easy-loop-e2e` | The closest existing thing to the MEDIUM tier: requirements interview (one multiple-choice question per visit, `Z. begin work`), **validation-holdout** criteria the reviewer judges against but never names, plan as `- [ ]` checklist, reviewer is the only agent that checks boxes, coder proposes checks with evidence. Orchestrated by the parent agent in the foreground. |
| `df-easy-loop-simple` | Same loop from an existing spec, no interview: the "good spec goes straight to work" path. |
| `df-chapter-create` | Chapters: 12–100 sprint vectors, under 3000 tokens, Pyramid Index, no fan-out. The LARGE tier's outer unit. |
| `df-sprint-plan` | Fan-out research/planning: orient → intent → three parallel drafts → cross-critique → interview → merge. Ledger CLI. |
| `df-sprint-execute` | Hand a sprint to one agent, review, update ledger. The LARGE tier's middle unit. |
| `df-semantic-index` | Token cache + semantic index: materialize all business-relevant tokens locally, build a routing tree over them, parallel read harness (≤5 workers per batch). The "collect the information locally" doctrine this directory follows. |
| `df-promise` | Bounded goal loops with durable evidence and gates; a badge only when every gate passes. Relevant to the checklist-as-contract loop. |

## sources/tractor/ — this repo's `skills/`

`tractor` (moments → examples, prompt doctrine, symptom table, "make done
honest") and `orchestrate-attractor-loops` (manual manager loop). Both are
inputs to consolidate, not to keep as-is.

## sources/tractor-design/ — design records from this repo's ephemeral

- `brief.md`, `wisdom.md`, `HANDOFF.md` (adoption-design, 2026-08-19/20):
  the agent funnel, named patterns (Pathfinder, Ladder, Consensus Loop,
  Coached Run), the principle → mechanism map.
- `loop-frames/memo.md` + `research/`: loop nodes, checklist items in the
  workspace, ephemeral frames, goal gates, the typed ABI. Current design is
  §9 as amended by §9a–§17a.

## sources/proof-practice/ — how Tyler actually chooses and records proof

Added 2026-09-01 after Tyler asked whether the homework on proof-of-work had
been done. The skill text is short; the practice lives in these artifacts.

| Path | What it shows |
|---|---|
| `nlspec-methodology/methodology.md`, `journal.md` (agents repo, 2026-07-28) | The substitution: replace an uncheckable objective with a checkable proxy. Claims classified cell-local / integration; every seam gets a claim; mocks must be authenticated; the stopping rule; decision 21 (no universal decision procedure: "a checklist a weak model can satisfy while doing the wrong thing"). Addendum 2 models phased authoring as an attractor pipeline: chat-session node, resume-not-replay, timeout must suspend, reserve goal gates for mechanical checks. |
| `core-tools/repo-proof-policy`, `proof-work` (pagerguild/core-tools) | Choose proof from the files touched, never a fixed ritual; static checks are supporting evidence; separate current-diff failures from inherited debt; record skipped gates and unproved behavior; verify the PR head SHA before trusting CI. |
| `plainterms/testing-proof.md`, `AGENTS.md`, `ci-confidence-*`, `ci-confidence-engineer-CHARTER.md` | A per-change-class proof table for a real product; the CI Confidence Pass (coverage design before implementation, test repair after behavior proof); layer ownership; "delete test theater". |
| `codex-autoupdate/*` (2026-08-24) | Tyler's proof style at full strength: one executable proof command per claim, red on baseline and green when fixed, aggregate runners that fail on a missing command; a separate fresh-operator validation with an identity block (HEAD, status, toolchain); an excluded capture recorded rather than hidden. Two of these validations were run through Tractor. |
| `merge-herder/proof-scenarios/` | Proof designed as ordered externally observable interactions per scenario, naming what the harness must keep real vs control, and listing known contract holes rather than filling them. |
| `pdx/issue52-incremental-update-proof.md` | A production proof artifact: pass flag, service, scenario, metrics table, product-visible result with citation ids, cleanup state. |
| `chaios/milestone-evidence.md` | Per-milestone structure: running-system claim, boundary claim, engineering checks, **scope check** naming what was not proven. Source of the "demonstrated vs proven" vocabulary. |
| `tractor/final-validation/*`, `claude-live-proof.md` | This repo's own proofs: binary built once with recorded SHA-256, unpredictable inputs the agent cannot guess, independent recomputation of the expected result, matched tool-call/result pairs counted. |

## sources/loop-corpus/ — Aug 18 loop-guidance research

`synthesis.md` (use-Tractor-or-not test; default worker → check shape; small
repertoire) and `workflow-corpus.md` (the repeated shapes across the
Attractor family and neighbors, with sources).
