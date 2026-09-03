# Living instructions: decisions from the interview

Record of the one-question-at-a-time interview between Tyler and Claude,
2026-09-01 and 2026-09-02, that replaced the premature issue #31. Each entry
is a decision Tyler ratified, in his words where they matter. Items marked
*hypothesis* are working assumptions he explicitly declined to make policy.

## Frame

- The initiative: Tractor should carry its own operating instructions. A
  built-in planning workflow that ends with a plan and a workflow
  recommendation "is as important as having a fantastic workflow engine."
- Opinionated is good. "I don't want this thing to be complicated." Least
  engineering first; prove and massage; expect rework.
- 1000% replayability and mechanical history are not requirements. Avoid
  over-engineering on that dimension.

## Authoring and the interview

1. Authoring is itself a Tractor run. A library of named workflows is baked
   into the binary; the first entry is the onboarding/planning workflow.
   Choose one or start from a similar one.
2. The interview is grill-me style: the agent asks until it has sufficient
   understanding of intent, writes the output file, exits to the next node.
   Routing never goes to the human (the upstream attractor interview
   protocol got this wrong).
3. `ask` is a blocking CLI command, `tractor ask`, run by the agent inside a
   single node visit. Not a routing thing, not a socket, not an injected
   tool. The agent keeps its context across the whole interview.
4. Transport is a file. Questions are written where observers can find
   them; observers reply with `tractor answer <question id>`. Tyler's
   suggestion: `ephemeral/projects/<build>/interview/<question ID>.md`, one
   file per question, so nested interviews have separate slots and the id
   tells you how to respond. (Not yet settled: whether zero-or-one
   interviews at a time is an acceptable simplification.)
5. Addressee is configurable, human or calling agent. Human-only first.
   Hunch: intent questions are human-only; acceptance questions may go to
   the caller.
6. Rich content from day one: the agent can send HTML and images with a
   question, not just text. Clients in order: CLI, web page (text, rendered
   HTML, images), audio via the chaios realtime-router, Slack.
7. The planning workflow ends with `brief.md`, a checklist, and a
   recommendation, then hands back to the parent agent. SIMPLE is the
   recommendation "execute the plan yourself." No separate mechanism.

## Sizes

8. SIMPLE is at most one sprint. MEDIUM is more than one sprint and less
   than two chapters. LARGE is multiple chapters. MEDIUM and LARGE begin in
   the same interview; its exit routes the size.
9. Adopt diffusioninc's chapter and sprint conventions but live under
   `./ephemeral`, which is for things committed to git that are not official
   application documentation.
10. One run for the whole job. Execution never spawns child runs. (Plan and
    execute are two runs.)

## The checklist and the loop node

11. Goal gating is not a dedicated tool node. Validation is attached to
    checklist items, and every item's validation must actually run and pass
    before the loop routes onward.
12. One item shape, no validator kinds. An item is name, check, an optional
    shell command, and an optional `infer` (files plus a prompt for a cheap
    model). Human attestation is an `ask` in the graph, not an item field.
13. Command first, then infer. Both must pass.
14. Laps write validators (Tyler's (c)). The contract node fixes name and
    check up front, and fills command and infer where they can be known. A
    "baby bear" review step in the lap checks the validator is neither
    trivial (`true`) nor stricter than the claim. Whether validators can be
    fixed at design time stays an open question, tested by building the
    rest.
15. Marking is mechanical. At the end of a lap the engine assumes the
    current item is implemented and runs its validation. Pass flips
    `done: true`; that is how the engine knows to enter the next item. The
    agent never writes to the checklist's `done` field.
16. On failure: re-enter the same item. `max_visits` on the loop node is the
    only ceiling. No per-item retry counter, no fixup routing. Escalation
    after repeated failure is the parent agent's job, later. "Don't plan for
    any of this at this point."
17. Nesting: any number of nested loops, one shape at every level. A chapter
    is an item whose lap is the sprint loop; a sprint is an item whose lap is
    the coding and validation loop. Each level's item may carry its own
    command.
18. The checklist file is the ledger (diffusioninc's `ledger.yaml` plus a
    command per entry). Chapter and sprint docs stay as prose documents; an
    item's `doc` points at one and the doc is injected with the item. The
    definition of done stops being prose and becomes command and infer.
19. Sprints within a chapter may be planned one lap at a time by appending
    to the checklist, or written as an upfront backlog and edited over
    time. Both, because the checklist is an editable document the loop
    re-reads on every arrival.
20. Checklists are markdown with YAML frontmatter, so the definition of done
    can be open prose beside the items.

24. The frame block is `<iterate>` tags, nested, after a fixed preamble that
    says what the blocks are. Its exact contents are hypothesis, not
    doctrine; breadcrumbs pointing at the iteration files are the standing
    alternative.
25. `$goal` stays one string, the run's goal. No `$goal.N`, no `$goal.depth`,
    no per-frame variables: the frame already carries the lap's goal. Lean
    on less engineering until use says otherwise.

## Files

21. Everything lives under `./ephemeral/projects/<build name>/` (e.g.
    `ephemeral/projects/mvp/`), committed. *Hypothesis:* `ephemeral` gets
    harvested for material to promote into `./docs`.

## Surfaces

22. The agent-facing surface is the CLI, because a subcommand's contract
    costs nothing until it is invoked, while MCP tool descriptions sit in
    context on every turn. The `apply_patch` precedent: past some
    complexity a tool call stops being advantageous. MCP is replaced by CLI
    calls entirely.
23. `tractor steer` and `tractor answer` are CLI subcommands.

## Build decisions (2026-09-02, answered by Claude during the build)

Made while the loop node built items 1, 3, and 4 of the build order below
(`proof/living-instructions-build/README.md`). Each is an interview file
under `interview/`; the number is the question id.

26. (0001) `tractor answer` with no text and no stdin fails; no editor mode.
27. (0002) `tractor ask` has no timeout. Rerunning with the moved path
    resumes the wait without renumbering.
28. (0003) The run directory reaches agents as `TRACTOR_RUN_DIR` in the
    environment only. No preamble field, no `--run` flag until a backend
    proves it scrubs the environment.
29. (0004, 0005) Docs-site page `interviews.md` at order 3; the comparison
    page moves to 4.
30. (0006) Embedded workflows: `tractor workflow list`, `tractor workflow run
    <name> --project <build> --seed <path>`. Build name explicit, never
    derived. Same `--workdir` and `--logs` as `run`.
31. (0007) The planning interview is one codergen node that asks and then
    writes the artifacts, followed by a tool node that checks the outputs
    mechanically and routes back on failure. *Tyler: single node until
    proven, then discuss.*
32. (0008) `recommendation.md`: heading, `Size:`, `Rationale:`, `Next:`.
    Workflow names fixed as `plan`, `medium`, `large`.
33. (0009) `--project` is the execution workflows' one parameter; paths are
    resolved at materialization, no new graph substitution.
34. (0010) The LARGE per-chapter plan node asks only when an answer changes
    sprint scope or validation.
35. (0011) A LARGE plan's `checklist.md` is the chapter ledger with `doc`
    and `checklist` per item; the planner writes chapter docs and empty
    sprint ledgers; `validate-plan` checks that shape. MEDIUM stays flat.
36. (0012) `--logs` optional for all workflows; default is a fresh directory
    under the XDG state root, never under the committed project directory.

## Planning workflow decisions (2026-09-02, evening, Tyler and Claude)

Made in the design conversation that followed the build. The workflow they
describe is specified in `planning-workflow.md`. Decision 31's single
planner node is withdrawn.

37. **Promises, not a spec.** The interview's job is to find out what we
    promise about the work and to whom. A promise has a statement, what it
    must not imply, a scope, a verifier, and evidence. A promise that
    cannot be falsified is narrowed before it is accepted. The checklist
    item's `check` is the promise; `command` and `infer` are its gates;
    `done: true` is the attestation.
38. **Elicit, then prune.** Before asking, the planner drafts the promises a
    user of this thing would expect, from the seed, the repo, and research,
    and asks "do you promise this, and what must it not imply", with a
    recommendation. Declined promises become exclusions. After the promise
    list stabilizes, a question survives only if its answer changes a
    promise, its scope, or its verifier. Stop when a full pass changes
    nothing. Answers already supplied are never asked for twice.
39. **Batched questions.** One question file per independent group,
    ordered questions inside it, each with options and a recommendation;
    the answer file mirrors the numbering. One-question-per-file was a
    convention, not a rule.
40. **Seams matter only where a promise crosses them.** A seam is either a
    promise to another party (another project, a persisted format, a public
    API) or it is internal and the implementer owns it. The brief records
    promise-adjacent decisions only; everything else is deferred by default
    and moving an internal seam is ordinary work.
41. **Checks never prove a promise.** Tests, lint, and compile are required
    checks. Every promise needs a judgment over recorded evidence by a
    model that is not the coder's, looking at captures and logs rather
    than the coder's summary. At chapter exit the verifier is an agent with
    tools, fresh context, other provider, that operates the software
    itself and routes pass or fail; the chapter is marked only through
    the pass edge.
    A sprint is *demonstrated*; a chapter is *proven*.
42. **Two validation archetypes.** Universal ("for every case in a set"):
    withhold a sample as the holdout; the coder sees the rest, the verifier
    sees all. Scenario (a user story): no holdout; the proof is captures
    judged against the story. A holdout is always considered and belongs
    only to the universal archetype; a universal promise over a set the
    verifier checks exhaustively needs none (adopted 2026-09-03 as the
    reading of decision 56, which rules out a holdout in this project's
    proof; interview 0014, question 4, asks Tyler to confirm or reverse).
43. **Holdout storage, simple.** Under the XDG state root in a directory
    named by the build and a random token (`<build>-<token>`), written by the design lap, referenced only
    from the verifier's prompt. Not in the workdir, not under the run
    directory, not committed. Obscure, not secret; a sandbox that hides one
    directory is the eventual fix.
44. **Chapters durable, sprints re-planned.** Both ledgers are editable.
    The chapter ledger is written at plan time and edited only with a
    reason, by a human or the planner. Sprint ledgers get an upfront
    backlog and a `replan` node after every implement lap: own node, cheap
    model, fresh context, edits open sprint items only, never the chapter
    ledger. A sprint that finds the chapter wrong asks the human.
45. **No budgets.** Code-volume budgets as tripwires are one signal of
    several and not yet a science. The over-engineering guard is the
    promise list, its exclusions, and a supervisor whose question is "does
    this serve a promise". Which mechanical signals should prompt a
    refactor consideration is a separate research project (build order 6).
46. **Brief and research loop.** Intake writes the first research plan, so
    research runs before the first interview. Then brief and research
    alternate: the brief asks only about findings and unresolved promises
    and writes open plan entries; research works the entries and writes
    findings that name promises it thinks should change. A tool node halts
    the loop when findings are empty and no plan entry is open. Research
    never edits the brief, and may add a plan entry only through a finding
    the brief accepts. Hitting `max_visits` is a question to the human.
47. **Validation design loop.** A loop node with one item per promise. Each
    lap writes the user story, the evidence specification, the holdout
    where the archetype calls for one, the UI sketch where a screen is
    involved, and fills the `command` and `infer` of the sprint item that
    will demonstrate the promise (the validation ledger's own item has
    none). An
    adversarial reviewer (other provider, fresh context, told only the
    promise and the design) answers: can a coder satisfy this while the
    promise is false; is any check trivially true; is the design stricter
    than the promise. It is a codergen node with pass and fail edges;
    routing is the verdict and the item has no `command`.
48. **Plan review loop.** After assembly, a loop node whose ledger is the
    review passes, one question each, one fresh reviewer per lap on another
    provider. Passes, in order: traceability, consistency, slicing, proof
    quality, scope, executability. A failed pass routes its finding to the
    owning node, whose edge returns to the reviewer for the same pass;
    only the reviewer's pass edge returns to the loop, because a
    commandless item passes when its lap returns (corrected 2026-09-03
    by the consistency pass). The ledger is a project
    file, so a project may add a pass. The human approves after the loop,
    with the verdicts beside the package. (Amended by decision 58: seven
    passes, the seventh holistic.)
49. **Supervisors, named by their question.** `research_auditor`,
    `scope_cop` (brief, decompose, validation design), `slice_critic`,
    `proof_skeptic`. Fresh context, a provider other than the node they
    watch, steer authority into the active turn. In-graph supervision is
    engine machinery that already exists (spec §3.10).
50. **Human gates, three** (four with `ceiling`, the brief/research loop's exhaustion gate, added 2026-09-03 by the consistency pass; spec requires an escalation edge or the run fails)**.** The promise interview, proof-mechanism
    questions inside the validation design loop, and final approval. All
    through `tractor ask`.
51. **Sizes.** SIMPLE: the same graph, with `decompose` writing a
    one-item ledger, one validation lap, and the holistic review pass
    only; recommend self-execution (clarified 2026-09-03 by the
    consistency pass). MEDIUM: the full graph, decomposed into sprints. LARGE:
    the full graph, decomposed into chapters. Intake guesses; the human
    confirms at the brief gate.
52. **Research output.** A research directory in the project (the token
    cache) with leaves carrying pinned revisions, licenses, and bounded
    comparisons ("like X but only Y", never an adoption frame), a routing
    index over it (`research/INDEX.md`). Execution prompts inline resolved
    index hits, not a pointer. At most five research branches in parallel.

53. **The library is content.** Every prompt, supervisor brief, review
    pass, doctrine excerpt, and artifact skeleton the built-in workflows
    use is a file under `workflow/library/`, embedded with `embed.FS` and
    rendered with `text/template`; no prompt lives in a Go string.
    Prompts include doctrine files by name, so a teaching is edited once
    and every prompt that cites it changes. Tests render every template
    and fail on an unreferenced doctrine file. `tractor workflow show
    <name>` prints the materialized prompts so an editor sees what agents
    see. Each release can improve the design and proof workflows as a
    content edit, without touching Go.

54. **Model associations are content.** `workflow/library/models.yaml`
    maps roles to provider, model, and effort, with a `not` constraint
    where independence matters; edited by deploying a new version.
55. **Manual first.** The v2 planning algorithm is run by hand on the
    living-instructions project itself, effective 2026-09-02: Claude as
    the planner nodes, Tyler as the human gate, subagents as reviewers
    and research branches; portions are automated as they are built.
56. **Three chapters** for v2: library, planner, execution and proof. No
    engine change in any of them. No holdout over this project's own
    promises (P1 to P10); a scratch package under test in chapter 5 or 6
    may carry one, since the holdout machinery is the feature being
    proven (interview 0014, question 3, asks Tyler to confirm).
58. **Seven review passes.** The sixth is executability; the seventh is
    holistic and rubric-free, last, because single-criterion judges miss
    trade-offs (research F2, 2026-09-02).
59. **P8 restated.** Prompt bodies, doctrine, supervisor briefs, passes, and
    skeletons are library files and Go supplies only data values, the
    value functions `quote` and `shell`, and the composition actions
    `include` and `doctrine`; `show`
    prints what `Build` materialized, with `--stage` to diff a real
    stage minus its frame. Cheap roles use the `sonnet` alias on claude;
    providers are always explicit; template delimiters are non-default
    (interview 0013).
57. **No verdict files.** Review and verification are codergen nodes with
    pass and fail edges. A validation or plan-review ledger item has no
    `command`; the pass edge returning to the loop is the pass. A chapter
    item keeps the required checks as its `command` (decision 41). Reviewer notes are ordinary files
    with no format.

## Build order (revised)

1. Interview file plus `tractor ask` and `tractor answer`. **Done** (chapter 1).
2. Loop node (this branch; see `loop-node.md`). **Done.**
3. Built-in planning workflow, embedded, ends with plan and recommendation.
   **Done as a single node** (chapter 2). Superseded by decisions 37–52;
   the v2 workflow is `planning-workflow.md` (chapter 5, after the
   library in chapter 4).
4. Built-in execution workflows (MEDIUM loop, LARGE nested loops). **Done**
   (chapter 3); LARGE not proven live. Gains the `replan` and `verify`
   nodes from decisions 41 and 44 with chapter 6.
5. Web client, then audio, then Slack.
6. Complexity-signals research: which cheap mechanical signals (volume per
   cell, cyclomatic complexity, public surface, fan-in/out, cross-directory
   spray, per-file churn across laps, repeated failed fixes) predict a
   change becoming expensive to change again, judged against our own run
   logs. Output: a signal table with suggest and stop thresholds and a
   supervisor prompt that reads them. No engine change until the table
   exists. Its own project directory, after chapter 4.

## Rejected

- PR #29's `proof_contract` schema, modes, and terminal vocabulary.
- A `goal_gate` lint (only a tool node's `on_success` may route to
  `success`). Prototyped, then withdrawn once item-level validation
  replaced the dedicated gate. Patch kept in the session scratchpad only.
- Child runs per chapter. Loop state persisted in the checkpoint. Four
  item statuses. A wait node for the interview.
