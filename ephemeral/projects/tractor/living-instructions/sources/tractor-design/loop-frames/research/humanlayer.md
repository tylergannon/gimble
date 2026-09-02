# Research report: HumanLayer's workflow representation, state, and loops

Verbatim final report of the deep-research subagent (2026-08-24), commissioned
for the loop-frames design work. See ../memo.md §8.2 for the distillation and
the design revision it caused (workspace-checklist iteration replacing
frozen-in-checkpoint items).

---

All repos were cloned to session scratch (humanlayer, 12-factor-agents, fold, advanced-context-engineering-for-coding-agents, skills, rpi-coordination-template, agentcontrolplane, claudelayer, riptide-rpi, electric) — re-clone from github.com/humanlayer if needed.

### 1. What their workflow representation actually is — four generations, all converging on "event log + prompt-per-step + files as data plane"

**Gen 1 — 12-factor agents (doctrine, 2025):** workflows are *code you own*, not a framework graph. The agent is a while-loop over an **event-log thread**: `Thread { events: List[Event] }`, serialized to a custom XML-ish prompt via `thread_to_prompt()` (`12-factor-agents/content/factor-03-own-your-context-window.md:120-140`). Control flow is a hand-written switch over the LLM's structured "intent" output, with `break` for async waits and `continue` for sync tool results (`factor-08-own-your-control-flow.md:27-69`). Explicitly anti-"loop until you hit the goal": *"Agents, at least the good ones, don't follow the 'here's your prompt, here's a bag of tools, loop until you hit the goal' pattern"* (README.md:45). The blessed pattern is **micro-agents sprinkled into a deterministic DAG**, 3–20 steps each (`brief-history-of-software.md:106-155`).

**Gen 2 — CodeLayer / the humanlayer repo (`hld` Go daemon):** NO workflow graph at all. SQLite schema is `sessions` (with `parent_session_id` — sessions form a resume/fork **tree**) + `conversation_events` append-only event log (`humanlayer/hld/store/sqlite.go:83-183`). "Workflow" = markdown slash-command prompts in `.claude/commands/` (research_codebase, create_plan, implement_plan, ralph_*) and state lives entirely in **markdown artifacts** in a synced `thoughts/` repo.

**Gen 3 — current product (humanlayer.com, "Riptide"), the smoking gun:** workflows ARE explicit graphs-as-JSON, checked into the repo at `.humanlayer/workflows/` or `.agents/workflows/`. From the screenshot embedded in `advanced-context-engineering-for-coding-agents/images/product-review-workflow-json-shadow.png` (verbatim):

```json
{ "id": "implementation-review",
  "start": "implementation",
  "steps": {
    "implementation": { "prompt": "/rpi:implement-outline",
      "exits": [ { "title": "Review the code", "to": "code-review" },
                 { "title": "Create a pull request", "to": "pull-request" } ] },
    "code-review": { "prompt": "/code-review",
      "exits": [ { "title": "Revise the implementation", "to": "implementation" },
                 { "title": "Create a pull request", "to": "pull-request" } ] },
    "pull-request": { "prompt": "/rpi:describe-pr" } } }
```

Each step's prompt is a **skill/slash command** ("do the work for this phase"); edges are prose-titled **exits**; loops are just cycles (code-review → implementation). Per the sibling mockup (`images/product-review-mockup-handoffs-shadow.png`): exits become "**satisfied**" when the step's work supports them, and *"Every satisfied exit remains available for the user to choose"* — human picks by default; workflow authors opt into automation per-exit with `"autoAdvance": { "defaultArmed": true }`. The latest non-empty exit set persists in the composer so the user can keep iterating in-step without losing the handoff. docs.humanlayer.com confirms: four built-in workflows (RPI, PRD-Oriented, Oneshot, Freeform) as phase sequences; *"Design and outline handoffs always wait for you"* (docs.humanlayer.com/guide/skills-workflows, /reference/skills-workflows, /explanation/workflow-phases).

**Gen 4 — `electric` `packages/agents-runtime` (platform being built now):** 12-factor made literal. A durable **entity owns an append-only stream**; on every wake the runtime materializes the stream into typed TanStack DB collections and runs one code entrypoint `handler(ctx, wake)` (`electric/packages/agents-runtime/README.md`). Control flow is deterministic host code; the LLM is invoked inside via `ctx.useAgent(...)`. **`fold`** (their new agent core) is the same shape at the session level: log-based state, per-agent context computed as a **projection** of the log (`fold/packages/fold-core/src/Projection/Projection.ts`), a `doom-loop` stop condition that halts on repeated identical tool-call batches (`StopConditions/StopConditions.ts:1-92`), and subagents that can be **dispatched fresh, resumed by agent_id with context intact, or forked as "a copy of your own context"** (`Subagents/SubagentTool.ts:129`).

### 2. Exact mechanisms for state/context between steps and across loop iterations

- **Files are the data plane, period.** Numbered markdown artifacts per task (`ticket → research → design discussion → PRD → TDD → structure outline → plan`), with an explicit conflict rule: *"Later artifacts take priority when artifacts disagree. Current code remains the source of truth for questions about current behavior"* (docs.humanlayer.com/reference/skills-workflows). Files sync via task worktrees + cloud (`.humanlayer/tasks/<task-slug>/`).
- **Iteration progress inside a loop = checkboxes edited in the plan file itself**: "check for any existing checkmarks (- [x])… Trust that completed work is done. Pick up from the first unchecked item" (`humanlayer/.claude/commands/implement_plan.md:12,77-82`).
- **Cross-session "stack frames" = handoff documents**: a templated compaction (Task(s)/Critical References/Recent changes/Learnings/Artifacts/Next Steps) written to `thoughts/shared/handoffs/`, resumed by `/resume_handoff <path>` (`create_handoff.md`, `resume_handoff.md`).
- **Ralph loops (run-forever iteration): the loop variable lives in an external queue.** `ralph_research.md` / `ralph_impl.md` run the *same constant prompt* every pass; each pass queries Linear for "top 10 priority items in status X", picks ONE, moves its status (research needed → in progress → in review; ready for dev → in dev), writes an artifact, exits. Iteration state = ticket statuses + artifact files, never an engine counter. `riptide-rpi/skills/rpi-setup-humanlayer/SKILL.md` makes filesystem-as-state-machine explicit: *"Based on which files exist, suggest the next skill to use."*
- **Scheduled control loops (skills repo)**: `design-control-loop` frames loops as control theory — **set point / sensor / controller / actuator / disturbances / dampener**. The gap to the goal is **re-measured by the sensor every run** (state re-derived, not carried); the only carried state is a version-controlled **memory file** (`.github/agent-memory/<task>.md`) "loaded deterministically into the actuator's context after the controller on every run", holding only standing feedback, plus **flow control**: scheduled runs no-op while ≥1 loop-labeled PR is open (`skills/plugins/design-control-loop/skills/design-control-loop/SKILL.md`, phases E–G).
- **Within-session loop state**: fold compaction entries carry a structured summary (Goal / Constraints / Progress Done-InProgress-Blocked / Key Decisions / Next Steps / Critical Context) plus `postCompactionInstructions`, and the summary *replaces* the log range in projection (`fold/packages/fold-core/src/Compaction/CompactionPrompts.ts`).
- **Newest context-injection mechanism** (electric agents-runtime): `ctx.useContext({ sourceBudget, sources })` — named content sources each declaring a **cache tier** (`pinned | stable | slow-changing | volatile`) and a token `max`; assembly enforces budgets and logs overflows (`context-assembly.ts`, `types.ts:278-320`). Plus a **goal** object with `tokenBudget` (default 50k) and status (`goal-api.ts`).

### 3. Stated design philosophy (verbatim quotes)

- Context: *"At any given point, your input to an LLM in an agent is 'here's what's happened so far, what's the next step'"*; *"LLMs are stateless functions that turn inputs into outputs"* (factor 3). *"the contents of your context window are the ONLY lever you have"* (ace-fca.md:150).
- State: *"you can engineer your application so that you can infer all execution state from the context window… execution state is just metadata about what has happened so far"*; benefits include *"**Forking**: Can fork the thread at any point by copying some subset of the thread into a new context"* (factor 5).
- Control flow: own it in deterministic code; the #1 missing feature everywhere is interrupt/resume *"between the moment of tool **selection** and the moment of tool **invocation**"* (factor 8). Factor 12: the agent is a **stateless reducer / foldl** over the event list (hence the product name "fold").
- Scope: *"Even as models support longer and longer context windows, you'll ALWAYS get better results with a small, focused prompt and context"*; micro-agents of *"3-10, maybe 20 steps max"* (factor 10, brief-history:98).
- Loops: doctrine-level skepticism of naked "loop until solved" — *"Agents get lost when the context window gets too long — they spin out trying the same broken approach over and over again — literally that's it, but that's enough to kneecap the approach"* (brief-history:89-92). 2026 position (wsff.md): *"no amount of harness engineering or loopsmaxxing can solve what is fundamentally a model-training issue"*; the fix is front-loaded human alignment (Product review → System architecture → **Program design** → Vertical slices) and bounded review-gated loops, not autonomy.

### 4. Surprising / directly reusable findings

1. **Convergent evolution — the headline.** HumanLayer's shipping workflow format is nearly isomorphic to Tractor's: named steps whose prompt is a reusable skill ("do the work for this phase"), prose-labeled exit edges, loops as cycles. They have **no variables/data plane either** — state moves via the task worktree and markdown artifacts, exactly Tractor's bet. Their delta: exits are *offered as satisfied options* (default human-chosen, per-exit `autoAdvance` opt-in), where Tractor's engine turns edge conditions into a structured routing question the agent answers.
2. **No engine-owned loop iterator anywhere.** Across four generations they never inject "current chunk is X / iteration N". Loop state is always externalized and *re-derived at iteration start*: sensor re-measures the gap (control-loop skill), ticket statuses gate the queue (ralph), checkboxes in the plan file mark progress (implement_plan), file-existence determines the next phase (riptide-rpi). The node prompt stays constant; the first action of each round is "read the state store." For Tractor this argues: rather than an iterator primitive, give loops a deterministic *sensor/selector step* whose output is the round's work item, and let "done" be the sensor reporting an empty gap.
3. **The stack/scope metaphor exists but as documents and log-forks, not engine frames**: handoff docs are literally serialized stack frames (template in `create_handoff.md`); fold's subagent `fork: true` ("launch a copy of your own context") and thread-forking (factor 5 benefit #6) are the in-memory equivalent. Resumable subagents-by-id (`agent_id … resume this agent`) are persistent coroutine frames.
4. **Anti-runaway primitives worth stealing**: fold's doom-loop detector (stop after N identical tool-call batches — fingerprint = sorted-key stable JSON of tool calls); control-loop **PR bounding** (loop no-ops while its previous output is unreviewed) as backpressure; goal token budgets.
5. **Deliberate context *withholding* per phase**: in the QRSPI rebuild, research runs with the feature ticket hidden "to prevent premature opinions" (alexlavaee.me/blog/from-rpi-to-qrspi/ summarizing Dex's "Everything We Got Wrong About RPI" talk); also instruction-budget discipline — each step under ~40 instructions after an 85-instruction prompt silently dropped constraints.
6. **Artifact precedence rule** ("later artifacts win; code is truth for current behavior") is a cheap, prompt-level substitute for scoped variables — worth copying verbatim into loop-node prompts.

### 5. Could not determine

- The **full Riptide workflow JSON schema** (all step/exit fields, how "satisfied" is computed, autoAdvance semantics beyond `defaultArmed`) — the product is closed-source; evidence is limited to the screenshot in the ACE repo, docs.humanlayer.com prose, and grep of the public `electric` fork (which contains the runtime but not the workflow-schema package). Whether the *agent* can select an exit (vs. suggesting it to the human) is not documented publicly.
- Whether Riptide steps reuse one LLM session across a cycle iteration or start fresh per step (docs imply "different sessions for different phases" but don't pin per-loop behavior).
- Nothing runnable was cheap *and* informative without API keys: fold's TUI and the hld daemon both require provider auth/subscriptions, so behavior was verified by reading code (fold-core AgentRuntime/Projection/Subagents, hld session manager) rather than executing; the old SDKs were deleted from the humanlayer repo (`humanlayer.md`: "The humanlayer sdks were removed in #646").
- `agentcontrolplane` (K8s CRDs: Task = Agent + message + context window; sub-agents as tools) is alpha/dormant since mid-2025 — direction confirms the event-log/delegation-tree philosophy but adds nothing on loops.

Sources: repos above; https://docs.humanlayer.com (/explanation/workflow-phases, /reference/skills-workflows, /guide/skills-workflows, /explanation/tasks); [RPI→QRSPI writeup](https://alexlavaee.me/blog/from-rpi-to-qrspi/); wsff.md's own link list (talks: ACE 8/25, No Vibes Allowed 11/25, "Everything We Got Wrong About RPI" 3/26, WSFF keynote AIE 2026).
