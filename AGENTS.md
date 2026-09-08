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

## New-build proof

Before calling any new Tractor build proved, run:

```sh
go test -tags=integration ./internal/workflows -run '^TestCanonicalLoop$' -count=1 -v -timeout=25m
```

This launches the shipped `sprint-execute` workflow through a freshly built
Tractor binary with real native agents and the [canonical Go example](examples/loops/canonical/README.md).
To prove a separately built candidate, append `-args -tractor-binary /absolute/path/to/tractor`.
Require exit zero and the fresh artifact directory's `result.json` with
`passed: true`; inspect its review transcripts and report its binary hash and
artifact path. A skipped, failed, interrupted or stale run is not proof.
This is required in addition to ordinary checks and proof of changed behavior.

## Planning, design, and execution rules

Before planning, designing, or running an interview, loop, or workflow for
this repo, read `ephemeral/projects/tractor/workflow-designer/rules.md`. It
records the rules, each marked as decided by Tyler or proposed by an agent.

## Reference submodules

`reference/` holds upstream projects mounted as submodules for reading only.
Never commit inside one and never push from one. See `reference/README.md`.
