1. `chapters/04-library/prove/show-and-orphan-walk.sh` generates calls to `"fmt.Print(v.PromptValue())"` and then builds that program. But `graph/graph.go` defines `"func (n *LLMNodeFields) PromptValue(label string) string"`. Therefore the command in `chapters/04-library/sprints.md` invoking this script cannot compile its generated helper as written. Owning node: decompose.

2. `chapters/04-library/SPRINT-03.md` says, `"Written by Claude, not the coder"` and `"the coder's part is limited to the include wiring"`, while the sprint requires creating ten doctrine pages and six skeletons. The loop supplies only the implement turn; no separate Claude turn or precondition supplies that content. The sprint therefore cannot be completed by the loop node as written. Owning node: decompose.

3. `chapters/05-planner/sprints.md` backlog entry 6 combines `"Pass files, generated review ledger, reviewer with routes to the owning node; assemble; approve. Proof: P5 scenario."` Seven review-pass artifacts, ledger generation, routing, assembly, approval, and a nested proof are multiple independently executable slices, not one agent-turn sprint. Owning node: decompose.

ROUTE: fail
