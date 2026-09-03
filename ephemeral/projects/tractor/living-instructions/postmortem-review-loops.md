# Postmortem: the review loops that ran all night

Written 2026-09-03 after the manual run of the v2 planning algorithm
(decision 55) ran unattended from 21:00 to 07:22, made 56 commits, and
built nothing. This is recorded because the mistakes are ordinary: any
agent running a fix-and-review loop will make them unless the prompts
and the loop forbid it. The built planner (chapter 5) must carry the
fixes below as content.

## What happened

Two loops ran by hand: the validation design loop (Claude designs, a
fresh codex session reviews, fail returns to design) and the plan
review loop (seven pass questions, fresh codex session each). Neither
had a lap ceiling. The reviewer prompts asked for attacks and made any
finding a fail. The author's only response to a finding was to add
text. Result:

| Document | Start | After 38 laps |
|---|---|---|
| validation/P8/design.md | 30 lines | 217 |
| validation/P10/design.md | 40 | 171 |
| validation/ledger.md | 55 | 210 |
| planning-workflow.md | 314 | 411 |
| declaration.md | 154 | 200 |

P8 failed 26 laps. The consistency pass failed 38 laps; its last
findings were wording drifts between restatements that earlier laps had
introduced. Total: 5,644 lines added, all prose.

## The five defects

1. **The reviewer was told to attack.** "Describe the cheapest way to
   game it" and "any contradiction is a fail" presuppose a finding
   exists. An agent asked to produce an attack produces one, every
   time, up to and including games that require rewriting the engine.
   The intent behind the gaming question (decision 41: a check must not
   be satisfiable while the promise is false) is right as evidence and
   wrong as the verdict. The verdict question is different: could a
   frontier coding agent, given this plan and nothing else, arrive at a
   correct result?
2. **Zero-tolerance criteria on prose cannot converge.** Five documents
   restating the same ten promises will always drift somewhere. Each
   fix adds text to one of them, which creates the next drift.
3. **The author had one move.** The loop as designed gave the reviewer
   the verdict (the pass edge marks the item) and the author no
   authority. So every finding was treated as an instruction, and the
   shortest answer to "you did not say X" is to say X. The author never
   rejected a finding as wrong, never conceded one under Not proven and
   marked the item done, and never asked the human. When the reviewer
   kept winning, the author edited the reviewer's prompt to raise the
   bar instead of stopping.
4. **No ceiling.** Decision 46 gives the brief/research loop a ceiling
   and makes reaching it a human question. The review loops had none
   and the author did not apply the rule to itself.
5. **No size signal.** A 30-line design becoming 217 lines is a halt
   condition by itself. Nothing watched the diff.

## What replaces them

**The reviewer's question.** One question, asked of a fresh reviewer:

> Could a frontier coding agent, given this plan and nothing else,
> arrive at a correct result? List what is missing or wrong that would
> send it astray. For each item say whether it is blocking (the coder
> would build the wrong thing, or could satisfy the check while the
> promise is false) or a note. Ignore wording, naming, and anything a
> competent coder resolves on its own. Route fail only if a blocking
> item remains.

The three adversarial questions (can a coder satisfy this while the
promise is false; is a check trivially true; is it stricter than the
promise) stay as things the reviewer looks for when grading an item
blocking, not as the verdict.

**The author's three moves.** For each finding the author does one of:
fix it; reject it with a one-line reason; or concede it under Not
proven. The response is recorded beside the review. Only a fix changes
the design, and a fix rewords before it adds. The author marks the item
done when no blocking finding is unanswered.

**Ceiling.** Two laps. A third fail is a question to the human with the
residual findings attached, never a third lap.

**Size tripwire.** A fix that grows the document by more than a fifth is
a redesign, not a fix; the author stops and asks.

**Stopping criterion, stated plainly.** Done is: no blocking finding
unanswered after at most two laps. Not: no finding at all. A plan is
finished when a coder can build the right thing from it, which is
90-95% right, not blessed.

## Where this lands

- Decision 60 in `decisions.md`.
- The `review` and `reviewer` prompts and the seven pass files in the
  library (chapter 5) carry the question above and the blocking/note
  grading. The `design` prompt carries the three moves.
- The loop bodies in `plan.yaml` carry `max_visits` of two on `review`
  and `reviewer` with the exhaustion edge routing to a `tractor ask`.
- Memory for the agent: `review-loop-ceiling` (cap loops; never loop
  unattended overnight).
