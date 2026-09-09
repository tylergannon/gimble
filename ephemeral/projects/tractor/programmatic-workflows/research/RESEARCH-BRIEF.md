# Programming-based Tractor workflows: research brief

User request, 2026-09-08: investigate replacing graph authoring with Go programs for nested goal-seeking loops, typed per-call agent outputs, ordinary language routing and scoping, an embeddable Tractor library, and useful human visualization before execution. Exact serialized resume is not a priority; inexpensive recovery from repository state may suffice. Declarative goals/validation remain central. Research first; no production implementation or migration authorized by this brief.

Research questions: which choices are independent; what do actual implementations demonstrate; which projects are maintained; what can code honestly visualize before execution; what guarantees disappear or remain; is a narrow POC justified, and what would falsify it?

All research findings/recommendations are agent proposals, not user decisions. Pin repositories and copied documentation, retain licenses, distinguish source inspection from observed execution and vendor claims. Do not run fetched code. Do not modify reference submodules. Existing upstream-orchestrators findings are leads, not authoritative current facts.

Use the adapted semantic-index method in corpus/method/df-semantic-index-SKILL.md. Collect sources first, then treat that segment as immutable while producing leaves. Leaf format: purpose; question-specific findings with corpus-relative path:line and pinned URL; themes; gotchas; recipes for retrieval; current status evidence/date; license; unknowns and counterevidence. All corpus-relative paths start at corpus/. Root will integrate topic routes. Avoid a generic survey: study the mechanisms that would change this decision.

Source corpus is under ephemeral/projects/tractor/programmatic-workflows/corpus. Index under index/. Source snapshots can be selective copies with a manifest of repository/commit/original path/SHA256; no need for full huge repos. Preserve copyright/license files. No binaries or dependencies. Store metadata JSON locally. Workers own only their named corpus segment, leaves prefix, and research memo. Do not commit; root integrates and commits. You are not alone: preserve other agents' edits.
