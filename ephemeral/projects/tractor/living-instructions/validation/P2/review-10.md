1. No cheaper qualifying game is apparent. The fresh scratch run, research segments, finding snapshots, nonce-bound `tractor ask`, tool logs, and final halt route jointly require the tested scenario to occur (`design.md:47-75`). Hard-coding this seed would not falsify P2 because arbitrary-seed convergence is explicitly excluded (`declaration.md:52`).

2. No substantive check is trivially true or relies on uncaptured evidence. The ask is bound to `QuestionAsked` through its nonce-bearing result, and `tool.log` plus `StageCompleted(next)` binds the halt to a real tool execution. The show-versus-run command gap is expressly conceded under “Not proven” (`design.md:77-88`).

3. Yes. The lexical finding rule is stricter than P2. A correct question could say, “Finding: the consumer accepts only a structured object encoding. Promise: replace comma-delimited output with that encoding?” This semantically cites and asks about the finding—as the infer explicitly permits “in any wording” (`design.md:73-75`)—but contains none of the required path, title, `F<n>`, or `JSON` tokens (`design.md:13-18`). It would match the generic `Promise:` rule, so the command’s requirement for exactly one finding-rule match (`design.md:53-54`) rejects a correct implementation.

ROUTE: fail
