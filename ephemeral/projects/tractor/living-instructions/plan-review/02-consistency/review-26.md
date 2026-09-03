1. `declaration.md` applies four checks to every chapter:

   > “Required checks (`go build`, `go vet`, `go test`, `golangci-lint`) … apply to every chapter”

   But `chapters.md` gives chapter 1 only:

   > `go build ./... && go test ./cmd/tractor/... ./engine/...`

   Chapters 2 and 3 likewise omit `go vet` and `golangci-lint`. `declaration.md` should narrow the rule to chapters 4–6. Owning node: `brief`.

2. `planning-workflow.md` requires a distinct sprint item per promise:

   > “fills the sprint item … for the promise (one per promise…)”

   The same document says:

   > “SIMPLE … `decompose` writes a one-item sprint ledger; the validation loop runs one lap per promise”

   A multi-promise SIMPLE package cannot satisfy both rules. `planning-workflow.md` should define how multiple promises aggregate into the single SIMPLE item. Owning node: `decompose`.

3. `planning-workflow.md` says future LARGE chapters have:

   > “a backlog as prose in the ledger body and an empty item list; the `plan` node turns the prose into items when the chapter is entered”

   Both `chapters/05-planner/sprints.md` and `chapters/06-execution/sprints.md` implement that rule with:

   > `items: []`

   But the validation-design loop, which runs before execution enters any chapter, says:

   > “It then fills the sprint item the promise table's Item column names for the promise … in the chapter's sprint ledger for LARGE”

   Those sprint items do not yet exist. `planning-workflow.md` should specify where designs persist gates until the execution `plan` node creates the items. Owning node: `design`.

4. Chapter-item validation has three incompatible descriptions. `decisions.md` decision 37 says:

   > “The checklist item's `check` is the promise; `command` and `infer` are its gates”

   `planning-workflow.md` says:

   > “A chapter item carries … `command` … and no `infer`.”

   The referenced ledger contract in `loop-node.md` says:

   > “An item with neither `command` nor `infer` passes when its lap returns. This is what a chapter item looks like”

   `decisions.md` decision 37 and `loop-node.md` should be amended to distinguish legacy commandless chapter items from v2 command-only chapter items. Owning node: `decompose`.

5. `chapters/04-library/SPRINT-01.md` says:

   > “sprint 2's `show` is checked against the same [snapshot] files.”

   But `chapters/04-library/SPRINT-02.md` defines the proof as:

   > “for every node `show --raw` is byte-equal to a program … that calls `workflow.Build` directly”

   Snapshot files and a live `Build` dumper are different comparison authorities. `SPRINT-01.md` should change. Owning node: `decompose`.

6. `planning-workflow.md` says a failed plan-review pass remains inside its current lap:

   > “fail to the owning node … whose edge returns to `reviewer` for the same pass; only a passing reviewer returns to the loop”

   `chapters/05-planner/sprints.md` instead requires proof that:

   > “one pass fails, its owning node runs, and the pass is re-selected and marked”

   The pass cannot be re-selected without first returning to the loop. `chapters/05-planner/sprints.md` should say the reviewer reruns for the same selected pass. Owning node: `decompose`.

ROUTE: fail


