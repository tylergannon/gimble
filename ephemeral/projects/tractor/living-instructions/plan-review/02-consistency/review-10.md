1. The research corpus is called both a “local library” and a “research directory (the token cache),” while “library” also names `workflow/library/`.

   - `declaration.md:14`: “researches prior art into a local library the plan can cite”
   - `planning-workflow.md:15-17`: “a research directory with a routing index”
   - `planning-workflow.md:243-246`: “The workflows are content. Everything an agent is told lives as a file under `workflow/library/`”

   `declaration.md` should call the research artifact the research directory/token cache. Owning node: `brief`.

2. The holdout ruling is simultaneously final and awaiting confirmation.

   - `declaration.md:153`: “**Holdout seeds.** Withdrawn; see P10. No holdout in this project.”
   - `decisions.md:190-194`: “a universal promise over a set the verifier checks exhaustively needs none (adopted 2026-09-03 …; interview 0014, question 4, asks Tyler to confirm or reverse).”
   - `decisions.md:279-283`: “No holdout over this project's own promises … (interview 0014, question 3, asks Tyler to confirm).”

   `declaration.md` should not call the question withdrawn while the governing decisions say confirmation remains outstanding. Owning node: `brief`.

3. The ceiling route can report workflow success without producing the approved package promised by the declaration.

   - `declaration.md:17-18`: “reviews the plan … and hands the human an approved package.”
   - `planning-workflow.md:91-94`: “`ceiling` … asks the human one question (continue with a higher ceiling, or stop) … and routes back to `brief` or to `success`”
   - `planning-workflow.md:179-181`: “**approve** … Yes routes to `success`.”

   The `ceiling` stop route in `planning-workflow.md` must not share the successful-completion terminal with approved output. Owning node: `brief`.

4. `Build` is assigned two incompatible responsibilities.

   - `planning-workflow.md:245-249`: “`workflow.go` keeps the graph surgery (`Build`, path resolution, validation)”
   - `chapters/04-library/SPRINT-01.md:50-54`: “`Build` itself is `Render` of each node's file and nothing more.”

   `SPRINT-01.md` should say that node payload materialization uses `Render`; `Build` still performs graph construction, resolution, and validation. Owning node: `decompose`.

5. The `workflow show` contract calls every node value a prompt in the declaration, but the sprint contract distinguishes three different values.

   - `declaration.md:60`: “prints every node of the graph, each node's prompt exactly as `Build` materialized it”
   - `chapters/04-library/SPRINT-02.md:23-25`: “`<prompt text, or tool_command, or checklist path>`”
   - `chapters/04-library/SPRINT-01.md:58-62`: “the prompt of every codergen node, the `tool_command` of every tool node, and the `checklist` of every loop node”

   `declaration.md` should promise each node’s materialized payload, with the three payload kinds named explicitly; `planning-workflow.md:286-288` repeats the same drift and should be conformed. Owning node: `brief`.

6. “Prompt” is also redefined as an umbrella name for doctrine and skeletons after P8 introduced them as separate concepts.

   - `declaration.md:60`: “Every prompt body, doctrine page, supervisor brief, pass, and skeleton”
   - `chapters/04-library/SPRINT-03.md:29-31`: “All three directories are prompts in P8's sense: text the library sends to an agent. Skeletons under `templates/`…”

   `SPRINT-03.md` should use the established umbrella term “agent-facing library files/content,” not rename doctrine and skeletons as prompts. Owning node: `decompose`.

7. `workflow show` is said to take the same flags as `run` while explicitly omitting one of `run`’s flags.

   - `chapters/04-library/SPRINT-02.md:17-18`: “takes the same flags as `run` (`--project`, `--seed`, `--workdir`), no `--logs`.”
   - `cmd/tractor/workflow.go:77`: `command.Flags().StringVar(&logsRoot, "logs", …)`

   `SPRINT-02.md` should say it shares only the named parameter flags with `run`. Owning node: `decompose`.

8. The declaration points to the wrong sprint as the authoritative doctrine-page list.

   - `declaration.md:111-114`: “the subset of `planning-workflow.md` §6a listed in `SPRINT-03.md`”
   - `chapters/04-library/SPRINT-04.md:11-18`: “## Pages … Only pages the four existing prompts can cite now,” followed by the page table.

   `declaration.md` should point to `SPRINT-04.md`. Owning node: `brief`.

9. Decision 37 says the chapter checklist check is the promise, but Chapter 4’s check is materially narrower than P8.

   - `decisions.md:157-162`: “The checklist item's `check` is the promise”
   - `declaration.md:60`: P8 includes “prompt body, doctrine page, supervisor brief, pass, and skeleton,” the exact template functions/actions, doctrine references, and exact `show` behavior.
   - `chapters.md:25-26`: “Every prompt, doctrine page, and skeleton … is an embedded library file with a render test, an orphan walk, and a `workflow show` command”

   `chapters.md` should carry the complete P8 claim or explicitly route the omitted portions to later chapter items. Owning node: `decompose`.

10. Chapter 5’s completion threshold has drifted from its stated backlog cardinality.

   - `chapters/05-planner/sprints.md:13`: “## Backlog sketch, in order”
   - `chapters/05-planner/sprints.md:51-53`: “14. Docs and skill…”
   - `chapters.md:36`: `test "$(grep -c 'done: true' …/chapters/05-planner/sprints.md)" -ge 12`

   No document authorizes dropping or combining two of the fourteen numbered units. `chapters.md` should use the reconciled count. Owning node: `decompose`.

11. Chapter 5 says it serves only P1–P5 and P7, while its own sprint backlog explicitly serves P9.

   - `chapters/05-planner/CHAPTER.md:9`: “Promises P1 to P5 and P7.”
   - `chapters/05-planner/sprints.md:51-53`: “Docs and skill … (the chapter 5 portion of P9's reader).”
   - `declaration.md:129-133`: “Chapter 6 … Promises P6, P9, P10.”

   `chapters/05-planner/CHAPTER.md` should name its partial P9 responsibility, or the sprint must stop claiming it. Owning node: `decompose`.

ROUTE: fail
