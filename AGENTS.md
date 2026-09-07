# Tractor repository instructions

## What Tractor is for

`docs/five-arts.md` names the five arts of orchestration this project exists to
serve, and the tensions between them. Read it before proposing a feature, so a
change that serves one art at another's expense is proposed as that trade
rather than as a straight improvement.

## Releases

For every request to release, publish, tag, bump a version, or update Tractor's
Codex plugin or marketplace, load and follow the `release-tractor` repository
skill at `.agents/skills/release-tractor/SKILL.md` before changing release
state.

## Planning, design, and execution rules

Before planning, designing, or running an interview, loop, or workflow for
this repo, read `ephemeral/projects/tractor/workflow-designer/rules.md`. It
records the rules, each marked as decided by Tyler or proposed by an agent.

## Reference submodules

`reference/` holds upstream projects mounted as submodules for reading only.
Never commit inside one and never push from one. See `reference/README.md`.
