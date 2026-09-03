1. `decisions.md` contradicts `declaration.md` about chapters 1–3. `decisions.md`: “the commandless chapter item of chapters 1 to 3 is the pre-v2 shape”; `declaration.md`: “chapters 1 to 3 are built and keep the commands they were marked with.” Change `decisions.md`. Owning node: `brief`.

2. P3 loses its provider-independence requirement when restated. `declaration.md`: “after a `review` turn on a provider other than the planner's routed pass”; `planning-workflow.md`: “after a `review` turn routed pass.” Change `planning-workflow.md` claim 3. Owning node: `brief`.

3. P5 loses its fresh-context requirement when restated. `declaration.md`: “after a fresh reviewer on another provider routed pass”; `planning-workflow.md`: “after a reviewer on a provider other than the planner's routed pass.” Change `planning-workflow.md` claim 5. Owning node: `brief`.

4. P6 loses the operating-verifier requirement when restated. `declaration.md`: “a `verify` turn for that chapter that operated the software and routed pass”; `planning-workflow.md`: “a `verify` turn routing pass.” Change `planning-workflow.md` claim 6. Owning node: `brief`.

5. Chapter 5’s promise ownership drifts between its chapter document and the chapter ledger. `chapters/05-planner/CHAPTER.md`: “Promises P1 to P5, P7, P8's chapter 5 leg … and the chapter 5 portion of P9.” `chapters.md`: “P1 to P5 and P7 demonstrated by nested runs, and P8's chapter 5 leg” — P9 is omitted. Change `chapters.md`. Owning node: `decompose`.

6. `chapters/04-library/SPRINT-02.md` gives incompatible constructions for the `--stage` proof. The command must “strip the frame (everything from the start through the end of the outermost `</iterate>` block and the preamble before it),” but the definition of done builds the test stage from only “the frame preamble … plus that program's output,” with no `iterate` block to strip. Change `chapters/04-library/SPRINT-02.md`. Owning node: `decompose`.

ROUTE: fail
