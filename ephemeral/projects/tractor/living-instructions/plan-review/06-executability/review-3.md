1. `chapters/05-planner/sprints.md` schedules “`brief` … Proof: P1” before “The observer and the shared proof library (`validation/observer.sh`, `validation/lib/`).” But `validation/P1/design.md` requires “`validation/lib/run-plan.sh`” and “`observer.sh`.” Sprint 5 therefore cannot execute its required proof until sprint 6 has run. Owning node: `decompose`.

2. `chapters/04-library/content/doctrine/ledger.md` names “`prove/show-and-orphan-walk.sh`,” and `prove/prompts-are-library-files.sh` also refers to “show-and-orphan-walk.sh.” That script does not exist; the actual scripts are `show-equals-build.sh` and `orphan-walk-and-render.sh`. Owning node: `decompose`.

3. `chapters/04-library/content/doctrine/pyramid-index.md` cites “`research/INDEX.md`,” but that path is absent from the package. Owning node: `decompose`.

ROUTE: fail
