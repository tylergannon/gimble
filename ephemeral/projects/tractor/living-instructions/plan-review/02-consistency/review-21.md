1. `halt` is classified as both a tool command and a prompt. `planning-workflow.md` says, “**halt** (tool),” but its library layout places `halt.sh` under “`prompts/ one file per node`.” This also conflicts with `SPRINT-01.md`: “`Build` materializes each prompt-bearing node's prompt with `Render` … (tool commands and checklist paths are `Build`'s own).” Change `planning-workflow.md` to give tool commands a distinct location or keep the command in the workflow definition. Owning node: `decompose`.

2. The orphan rule differs on direct versus transitive use. `declaration.md` defines reference as: “every doctrine page must render for some node (the orphan walk's definition of referenced).” `chapters/04-library/sprints.md` instead requires that an agent-facing file “directly references” each page. A transitively included doctrine page satisfies the declaration but fails the sprint rule. Change `chapters/04-library/sprints.md` to match the declaration’s render/reachability rule, or explicitly strengthen the declaration. Owning node: `decompose`.

3. The final proof artifact has two names. `declaration.md` promises “A proof record under `proof/planning-v2/`,” and `chapters/06-execution/sprints.md` likewise says “the proof record under `proof/planning-v2/` written from both runs.” But `chapters/06-execution/CHAPTER.md` says, “The proof run is the artifact: identity block, claims, scope check, findings.” Change that chapter sentence to “The proof record is the artifact.” Owning node: `decompose`.

ROUTE: fail


