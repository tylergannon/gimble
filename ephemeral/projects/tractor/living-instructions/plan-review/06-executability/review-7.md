1. No command or script finding. `chapters/04-library/sprints.md` names repository-root commands such as “`go build ./... && go vet ./... && go test ./workflow/...`”; every referenced proof script exists, shell-parses, and invokes the test names specified by its sprint document. Owning node: decompose.

2. No path or ledger-format finding. `loop-node.md` §2 requires “Markdown with YAML frontmatter” and paths “relative to the run workdir.” `chapters.md` and all six `sprints.md` files parse with an `items` list; every `doc` and `checklist` path resolves. Owning node: decompose.

3. No sizing finding. Chapter 4 explicitly defines “Five sprints, one agent turn each”; each has one bounded implementation and demonstration. Chapters 5 and 6 give each backlog entry one cohesive deliverable with its own proof—for example, “`intake` … writes `intake.md` and the first `research/plan.md`” and “`replan` in `medium` and `large`”—with no entry requiring a second implementation turn. Owning node: decompose.

ROUTE: pass


