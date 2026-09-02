---
name: repo-proof-policy
description: >
  Select and run proof gates for repository documentation, skills, policy,
  scripts, workflows, and source-index changes. Use when closing out docs
  maintenance or skill/tooling changes in any target repository.
---

# Repo Proof Policy

Use this skill before closing out documentation, skills, policy, or tooling
changes in a target repository. The job is to choose proof from the files
touched, not to run a fixed ritual.

## Workflow

1. Read the target repo's `AGENTS.md` and any repo-local proof policy such as
   `docs/policy/proof-gates.md`.
2. Classify the changed files: docs, skills, plugin metadata,
   submodules, workflows, scripts, or mixed.
3. Run every matching repo-local gate. If the repo has no gate, use
   [default-proof-gates.md](references/default-proof-gates.md).
4. Treat static checks as supporting evidence. For docs, verify the durable diff
   accurately represents its source material and is routed from the right index.
5. Separate failures caused by the current diff from inherited repo-wide debt.
   Record inherited debt, but do not treat it as a blocker when targeted proof
   for the changed surfaces passes and the target repo allows the change to
   proceed.
6. Record failures, skipped gates, unproved behavior, and any skill problems in
   the worklog. Use `skill_issue` or `skill_fix_request` when a shared skill
   caused friction or needs an upstream repair.

## Common Commands

Choose commands that exist in the target repo. Examples:

```sh
git diff --check
npm test
python3 <installed daily-docs-fold skill dir>/scripts/check-doc-indexes.py --repo .
```

When plugin metadata changes, parse manifests explicitly. For Core Tools itself:

```sh
node -e "for (const f of ['.codex-plugin/plugin.json','.claude-plugin/plugin.json','.claude-plugin/marketplace.json']) JSON.parse(require('fs').readFileSync(f,'utf8'));"
```

When submodules change:

```sh
git submodule status
```

## Closeout

Report the proof commands exactly, including skipped gates, inherited debt, and
the reason any behavior remains unproved. For PR work, pair this skill with
`proof-work` so exact-head checks, rebases, auto-merge, and merge verification
are handled instead of stopping at local proof.

source: pagerguild/core-tools
