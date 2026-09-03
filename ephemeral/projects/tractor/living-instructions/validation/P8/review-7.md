1. Yes. P8 requires every pass to be file-backed, but the pass-specific proof mutates and traces only one pass file (`design.md:76-84`). The cheapest game is to file-back that pass while keeping the other six as Go strings; the general `Build` checks do not trace those ledger-fed passes. This violates the universal promise (`declaration.md:58`) while satisfying the validator.

2. Yes. The orphan walk accepts a doctrine reference from any file under `supervisors/` or `passes/` (`SPRINT-02.md:50-53`), but nothing requires every such file to be used by a built node or generated ledger. An unused pass or supervisor file can cite an otherwise orphaned doctrine page, making the test pass although no prompt sent to an agent references that page.

3. Yes. The 200-byte heuristic (`design.md:68-71`) rejects a valid library template containing a rendered line over 200 bytes with an authorized parameter interpolation: the rendered line is not verbatim in the source and is not shorter than 200 bytes. P8 permits arbitrary data values and specifies no line-length constraint (`declaration.md:71-78`), so this is a concrete false failure.

ROUTE: fail
