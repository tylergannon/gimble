<system-message>
This is one step of a Tractor run inside a checklist loop. The iterate
blocks below are the engine's record of where you are: the item selected
for this lap, the check it must satisfy, the command and judge that will
validate it when this step ends, and what the previous validation reported.
Outer blocks enclose inner ones. Paths are relative to the working
directory. Do only what your prompt asks; the engine marks items done.
</system-message>
<iterate loop="chapters" checklist="ephemeral/projects/tractor/living-instructions/chapters.md" item="2/3" lap="1">
  name: Planning workflow
  check: A built-in planning workflow interviews the caller through `tractor ask` and ends with a brief, a loop-format checklist, and a size recommendation.
  command: 'go build ./... && go test ./... && test "$(grep -c ''done: true'' ephemeral/projects/tractor/living-instructions/chapters/02-planning/sprints.md)" -ge 2'
  doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/CHAPTER.md
  checklist: ephemeral/projects/tractor/living-instructions/chapters/02-planning/sprints.md
  --- doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/CHAPTER.md ---
  # Chapter 2: The planning workflow
  
  Status: active
  
  A workflow baked into the binary that turns a seed into a plan. It
  interviews the caller with `tractor ask` until intent is clear, then writes
  a brief, a checklist in the loop node's format, and a size recommendation,
  and hands back to the caller. "As important as having a fantastic workflow
  engine."
  
  ## Pyramid index
  
  - L0: Tractor ships its own planning workflow, and running it is how a
    caller learns how to use Tractor for the job at hand.
  - L1:
    - A library of named workflows embedded in the binary, listable and
      runnable by name without a pipeline file.
    - The planning workflow: interview node using `ask`, then a writing node
      that produces `brief.md`, `checklist.md`, and a recommendation under
      `ephemeral/projects/<build>/`.
    - The recommendation names a size (SIMPLE, MEDIUM, LARGE) and, for
      MEDIUM and LARGE, the execution workflow from chapter 3 to run next.
      SIMPLE means "execute the plan yourself."
    - Docs and skill teach the caller to start here.
  - L2: the sprint ledger is empty; the `plan` node writes it after
    interviewing the reviewer.
  
  ## Vector
  
  Decisions 1, 2, 5, 7, 8, 12, 14, 20, 21 in `decisions.md`, and the brief's
  "Definition" dimension: the interview is proportional to the size, keeps
  asking only while an answer would change the contract, and stops after
  two rounds that surface only derivable detail. The checklist it writes is
  the loop node's format from `loop-node.md` §2, with `check` fixed per item
  and `command`/`infer` filled where they can be known.
  
  ## Planning this chapter
  
  The planner must ask the reviewer at least about:
  
  - How embedded workflows are invoked (a subcommand, a flag on `run`, or a
    name where a pipeline path goes) and where their outputs land.
  - The shape of the interview node: one codergen node with a long timeout
    that calls `ask` in a loop, or something else.
  - What the recommendation file looks like and how a caller reads it.
  
  Sprints must each fit one agent turn and each carry a real `command`. The
  last sprint should be a live proof: the planning workflow run against a
  small seed, with the reviewer answering the interview.
  
  ## Non-goals
  
  Executing the plan (chapter 3). Multiple planning agents drafting in
  parallel. A web client.
  <iterate loop="sprints" checklist="ephemeral/projects/tractor/living-instructions/chapters/02-planning/sprints.md" item="3/4" lap="2">
    name: planning docs and skill
    check: The CLI help, README, reference and site docs, skill bundle, and llms.txt teach callers to start with the plan workflow and accurately describe its interview, outputs, sizes, and handoff.
    command: go run ./cmd/tractor workflow --help >/dev/null && go run ./cmd/tractor workflow run plan --help >/dev/null && test "$(go run ./cmd/tractor workflow list | grep -c '^plan')" -eq 1 && grep -q 'workflow run plan' README.md && grep -q 'workflow run plan' skills/tractor/SKILL.md && grep -q 'workflow run plan' llms.txt
    infer:
      prompt: Run the workflow help commands and judge whether a caller reading only these files would start with plan, answer its interview correctly, find all three outputs, and follow the size recommendation without the prose contradicting the CLI.
      files:
        - README.md
        - docs/spec.md
        - skills/tractor/SKILL.md
        - src/content/docs/*.md
    doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/SPRINT-03.md
    last validation: failed — judge: The workflow/list/run-plan help is consistent on starting with plan, required flags, QuestionAsked interviews, and all three outputs; `workflow list` exposes only plan. However, `planning.md:10-13` says every non-small, unsettled job should start with the built-in plan workflow, while `loops.md:57-62` directs far-off goals to a milestone loop with “No upfront plan” and suggests adding a planner node instead. A caller could therefore bypass the required plan workflow. Additionally, llms.txt names SIMPLE/MEDIUM/LARGE but omits their size boundaries.
    --- doc: ephemeral/projects/tractor/living-instructions/chapters/02-planning/SPRINT-03.md ---
    # Sprint 3: teach callers to plan first
    
    Make the planning workflow the documented entry point wherever Tractor teaches
    itself. Keep additions concise and copy flags and output names from the built
    help rather than this plan.
    
    - `README.md`: add `workflow list` and the minimal plan invocation near the
      first-run path.
    - `docs/spec.md`: specify named embedded workflows, invocation and parameter
      rules, interview behavior, the three artifacts, size boundaries, and the
      recommendation handoff.
    - `src/content/docs/`: add or update the best onboarding page so a person can
      start from a seed file, answer numbered questions, and find the result. Link
      to the interviews and loops pages rather than duplicating them.
    - `skills/tractor/SKILL.md`: when a user needs a plan or does not yet know which
      Tractor shape fits, start with `tractor workflow run plan`; explain how the
      caller watches and answers `QuestionAsked` and then follows SIMPLE, MEDIUM, or
      LARGE.
    - `llms.txt` (and `llms-full.txt` if generated from the same source): include
      the agent-facing start command and artifact names.
    - Cobra help: `tractor workflow --help`, `workflow list --help`, and
      `workflow run plan --help` must agree with the docs.
    
    Do not describe chapter 3's `medium` and `large` workflows as currently
    runnable merely because their names are fixed. The docs infer check compares
    the real help with the prose. Run all required BUILD.md gates before committing.
  </iterate>
</iterate>

You are executing one sprint of the living-instructions build for Tractor, in this repository. Read ephemeral/projects/tractor/living-instructions/BUILD.md first, then the sprint doc in the innermost iterate block above.

Do the whole sprint: the code, tests, and docs it names. Run the sprint's `command` yourself before you finish; the engine runs it again after you stop and re-enters this sprint if it fails, with the failure shown in the block. Where the sprint doc leaves a decision open, ask the reviewer with `tractor ask` and act on the answer. Commit when done.
