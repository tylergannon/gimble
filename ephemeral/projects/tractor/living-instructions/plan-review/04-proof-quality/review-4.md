1. `chapters/04-library/sprints.md` corrupts P8’s first three evidence lists through YAML folding:

   > `- workflow/library/README.md`  
   > `        - ephemeral/.../prompts-are-library-files.log`

   This parses as one nonexistent path containing `"README.md - ...log"`; the same defect combines the log and script paths for `show-equals-build` and `orphan-walk-and-render`. Yet `validation/P8/design.md` requires the judge to inspect “each script’s log” and the probe logs, while `engine/loop.go` supplies only files matched from `infer.files`. The semantic evidence is therefore not captured. Owning node: `design`. Promise: P8.

2. `chapters/04-library/sprints.md` promises doctrine and skeleton files are:

   > “copied from chapters/04-library/content/”

   But `prove/doctrine-pages.sh` checks only that installed files exist, have `Source:` lines, and pass render/orphan tests; the `infer.files` list contains only installed `workflow/library/doctrine/*.md` and `workflow/library/templates/*`. Neither command nor inference compares them with the planner-authored `content/` files. A coder can substitute different files and pass while this promise is false. Owning node: `design`. Promise: P8’s doctrine/skeleton sprint.

3. The same doctrine validation is stricter than P8. `declaration.md` says P8 must not imply:

   > “That the content is well written”

   but the ledger requires:

   > “the pages say what decisions 37 to 59 say, in the library’s voice”

   and its inference fails pages that contradict those decisions, exceed about sixty lines, or lack the prescribed provenance. P8 can be true—content is file-backed, rendered, referenced, and shown—while this validator rejects it for content quality. Owning node: `design`. Promise: P8.

ROUTE: fail

