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

## Build order (revised)

1. Interview file plus `tractor ask` and `tractor answer`. **Done** (chapter 1).
2. Loop node (this branch; see `loop-node.md`). **Done.**
3. Built-in planning workflow, embedded, ends with plan and recommendation.
   **Done** (chapter 2); design discussion with Tyler still owed.
4. Built-in execution workflows (MEDIUM loop, LARGE nested loops). **Done**
   (chapter 3); LARGE not proven live.
5. Web client, then audio, then Slack.

## Rejected

- PR #29's `proof_contract` schema, modes, and terminal vocabulary.
- A `goal_gate` lint (only a tool node's `on_success` may route to
  `success`). Prototyped, then withdrawn once item-level validation
  replaced the dedicated gate. Patch kept in the session scratchpad only.
- Child runs per chapter. Loop state persisted in the checkpoint. Four
  item statuses. A wait node for the interview.
