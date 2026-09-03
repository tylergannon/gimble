# Issue #33 — Loop node: exit goal gate evaluates the definition of done before on_done

Task brief for the implementer. The text below is the issue verbatim.
It was filed against branch `worktree-goal-gates`, which is now merged;
the code it names is on `main` at the same paths.

## Definition of done

- The required behavior below holds in the code.
- The tests named below exist and pass.
- The docs named below say what the code now does.
- `go build ./... && go test -count=1 ./...` exits 0.
- The change stays in the working tree. Do not commit, branch, or push.

---

Branch `worktree-goal-gates`, `engine/loop.go`.

## Bug

When the loop node arrives and finds no open item, it routes to `on_done` immediately. Nothing reads the checklist's markdown body, which is the prose definition of done. The exit goal gate from the loop-frames memo §9a.2 ("the run cannot terminate at success until an exit gate demonstrates the goal's claims") was dropped from the v1 loop node and is not listed in `loop-node.md` §10 "Not in v1".

Consequences: an empty ledger completes on first arrival. A ledger whose items were all demonstrated completes even when the definition of done is not met. A stale backlog is never reconciled against the goal.

## Required behavior

The checklist file is markdown with YAML frontmatter: the body is the definition of done, the frontmatter is the ledger. The checklist is editable on every lap. The loop evaluator, not the ledger, decides when the loop is done.

When the loop finds no open item:

1. Run one evaluator turn on the loop node's model fields (the same slot the infer judge uses). Inputs: the checklist body, the items with their last validation results, and the workspace.
2. Verdict done: route to `on_done`.
3. Verdict not done: the evaluator appends or rewrites items in the checklist as ordinary work. The loop re-reads the file and selects the first open item. Next lap.
4. Verdict not done and the file still has no open item: terminal error.

An empty ledger on first arrival takes the same path: the gate runs and the evaluator writes the first items.

## Changes

- `engine/loop.go`: the no-open-item branch runs the gate before routing to `on_done`.
- Gate prompt and choice schema with two targets, `done` and `not_done`, like the infer judge.
- Timeline event for the gate verdict; `prompt.md` and `response.md` in the loop node's stage directory.
- Tests: empty ledger gets items written and the loop continues; all-items-done with a not-done verdict adds an item; not-done with no edit is a terminal error; done routes to `on_done`.
- Docs: `docs/spec.md` loop section, `src/content/docs/loops.md`, `loop-node.md` §4.
