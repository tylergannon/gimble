1. `chapters.md` adds chapter gates beyond the permitted checks. `declaration.md:49-50` says: “Required checks (`go build`, `go vet`, `go test`, `golangci-lint`) apply to every chapter.” `planning-workflow.md:137-139` says: “A chapter item carries only the required checks as `command`.” But `chapters.md:53-54` says: “The commands count finished sprints so that a chapter whose planner appended nothing cannot pass vacuously,” and its commands also test files and timeline events. `chapters.md` should change. Owning node: `decompose`.

2. The holdout’s visibility guarantee drifts. `decisions.md:195-199` says it is “referenced only from the verifier's prompt” and “Obscure, not secret; a sandbox that hides one directory is the eventual fix.” `chapters/06-execution/sprints.md:14-17` instead promises “a small scratch package with one universal promise whose holdout sample the coder never sees.” Non-disclosure in the coder’s prompt cannot guarantee that an unsandboxed coder never sees the filesystem-accessible sample. `chapters/06-execution/sprints.md` should change. Owning node: `decompose`.

ROUTE: fail


