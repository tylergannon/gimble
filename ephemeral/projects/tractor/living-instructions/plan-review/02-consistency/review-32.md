1. `decisions.md` says each validation lap “fills the `command` and `infer` of the sprint item that will demonstrate the promise.” `planning-workflow.md` instead allows split promises: “when a promise has a leg in a later chapter, as P8 and P9 do here, each item carries its leg’s `check` and the Item column lists them all.” Change decision 47 to permit multiple mapped items. Owner: `design`.

2. `decisions.md` unconditionally says: “After a failed `verify`, `replan` appends one open item from the findings.” `planning-workflow.md` says it “appends one open item … (or asks the human when the findings say the chapter is wrong),” and `chapters/06-execution/CHAPTER.md` repeats “appends a repair sprint or asks the human.” Change decision 44 to include the escalation case. Owner: `decompose`.

3. `chapters/05-planner/CHAPTER.md` says: “seven passes, each a fresh reviewer routing pass or fail to the owning node.” `planning-workflow.md` says pass routes “to the loop node” while only fail routes “to the owning node.” Change the chapter summary. Owner: `decompose`.

4. `chapters/04-library/SPRINT-02.md` opens by saying `show` “prints every node of the plan graph with its prompt.” Its detailed contract instead prints “`<prompt text, or tool_command, or checklist path>`.” Change the opening to say each node’s payload. Owner: `decompose`.

5. `chapters/04-library/SPRINT-05.md` requires a root-README edit: “`README.md` ‘Start with a plan’: one sentence and a link.” But `chapters/04-library/sprints.md` declares its `command` and `infer` the definition of done while neither checks nor reads root `README.md`; they cover only the spec, docs site, skill, `llms.txt`, and library README. Add root README to the ledger item or remove it from the sprint contract. Owner: `decompose`.

6. `decisions.md` decision 53 says: “`tractor workflow show <name>` prints the materialized prompts.” The governing P8 statement in `declaration.md` requires it to print “every node of the graph, each node’s payload (a prompt, a tool command, or a checklist path),” and `planning-workflow.md` repeats that payload contract. Amend decision 53 to the full payload rule. Owner: `brief`.

7. `planning-workflow.md` defines the consistency pass as checking drift “across brief, chapter docs, sprint docs.” The pass itself, `plan-review/02-consistency.md`, checks “`declaration.md`, `planning-workflow.md`, `decisions.md` (37 onward), `chapters.md`, and every `CHAPTER.md`, `sprints.md`, and `SPRINT-*.md` under chapters 04 to 06.” Change the workflow’s pass description to the actual scope. Owner: `brief`.

ROUTE: fail
