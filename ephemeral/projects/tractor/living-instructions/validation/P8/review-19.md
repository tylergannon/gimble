1. Yes. The cheapest game is making headed `workflow show` print every node header but omit prompt bodies, while keeping `--node --raw` correct. The validator compares only headed node IDs, then compares `Build` solely with raw mode (`design.md:73-78`). Its infer rubric checks mutations and Go-supplied text, not headed-body presence (`design.md:147-158`), so the promised primary output can remain false.

2. Yes. The stage perturbation depends on uncaptured evidence. `show-equals-build.sh:55-57` suppresses all output and accepts any nonzero status, so a panic, parse error, or I/O failure passes just like the required diff exit status 1. The infer judge therefore cannot determine from the recorded log that the injected byte caused the intended comparison failure, contrary to `design.md:81-84,149-158`.

3. Yes. The promise requires each doctrine page to be rendered for some valid parameters, but the validator requires it under one of the README’s finite representative sets (`design.md:96-100`). A correct prompt could conditionally include `doctrine "acme"` when `.Project == "acme"` while the listed sets use other projects. That page is genuinely referenced and rendered for valid parameters, satisfying `declaration.md:60` and its template seam at `declaration.md:79`, yet the validator rejects it.

ROUTE: fail
