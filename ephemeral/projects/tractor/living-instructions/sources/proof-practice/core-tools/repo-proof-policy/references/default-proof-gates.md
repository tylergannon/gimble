# Default Proof Gates

Use these gates only when the target repository has no more specific proof
policy.

| change class | minimum proof |
|---|---|
| Documentation only | `git diff --check`; verify changed docs are routed from the repo's root docs index or equivalent |
| Documentation maintenance fold | input ledger; source-backed durable diff; route check; `git diff --check` |
| Skills added or edited | inspect rendered skill list if available; run repo tests for skill distribution or installer behavior if present |
| Scripts | targeted script exercise; include the exact command and output summary |
| Plugin or package metadata | parse JSON/YAML/TOML manifests with the repo's normal parser or a deterministic command |
| Submodules or source indexes | `git submodule status` when applicable; verify pinned commits and source-index updates |
| Workflow changes | syntax/static check when available; record whether hosted execution was run |

For documentation claims, static checks are supporting evidence only. The proof
is that the durable diff accurately represents the source material and routes
future agents to the right place.
