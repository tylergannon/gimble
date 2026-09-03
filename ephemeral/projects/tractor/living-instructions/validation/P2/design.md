# P2: the brief/research loop halts through the tool node, and a finding is asked, not applied

Archetype: scenario.

## Story

1. Scratch repository with a `research/` token cache seeded with one
   planted document that contradicts a promise the seed states (seed
   says "exports CSV"; the planted doc, presented as the customer's
   integration guide, says the consumer only accepts JSON).
2. Run `plan` with `answerer.sh`. Rules: when a question cites the
   planted document, answer "yes, change the promise to JSON"; accept
   everything else.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: stage sequence for `brief`, `research`, `halt`,
  `decompose`; `LoopCompleted` for the brief loop.
- `research/findings.md` history: the finding present after lap 1,
  closed after the brief lap that asked.
- The question file that cites the finding, and its answer.
- `promises.md` before and after (the run's git history; the planner
  commits per lap).

## Validator

`command`: `prove/p2-halt.sh`: the last `halt` stage exited 0 and the
next stage started is `decompose`; the brief loop's `LoopCompleted` is
not preceded by a `max_visits` exhaustion event; exactly one
`QuestionAsked` file mentions the planted document's filename; the
commit that changed the CSV promise to JSON is authored by the `brief`
stage (stage index in the commit message trailer), not by `research`.

`infer` (files: findings.md at each lap, the question, promises.md
diff): "Did research change any promise directly? Fail if the promise
changed in a commit made during a research stage."

## Not proven

Convergence on arbitrary seeds. Only that the halt is the tool node's
decision and that findings route through the human.
