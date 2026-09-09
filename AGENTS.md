# Gimble repository instructions

## What Gimble is for

`docs/five-arts.md` names the five arts of orchestration this project exists to
serve, and the tensions between them. Read it before proposing a feature, so a
change that serves one art at another's expense is proposed as that trade
rather than as a straight improvement.

## Releases

For every request to release, publish, tag, bump a version, or update Gimble's
Codex plugin or marketplace, load and follow the `release-gimble` repository
skill at `.agents/skills/release-gimble/SKILL.md` before changing release
state.

## New-build proof

Before calling any new Gimble build proved, run:

```sh
go test -tags=integration ./internal/workflows -run '^TestCanonicalLoop$' -count=1 -v -timeout=25m
```

This launches the shipped `sprint-execute` workflow through a freshly built
Gimble binary with real native agents and the [canonical Go example](examples/loops/canonical/README.md).
To prove a separately built candidate, append `-args -gimble-binary /absolute/path/to/gimble`.
Require exit zero and the fresh artifact directory's `result.json` with
`passed: true`; inspect its review transcripts and report its binary hash and
artifact path. A skipped, failed, interrupted or stale run is not proof.
This is required in addition to ordinary checks and proof of changed behavior.

## Planning, design, and execution rules

Before planning, designing, or running an interview, loop, or workflow for
this repo, read `ephemeral/projects/gimble/workflow-designer/rules.md`. It
records the rules, each marked as decided by Tyler or proposed by an agent.

## Semantic index

The writings on the proposed move from JSON/DAG workflow definitions to
Go-authored programs are routed through
`docs/semantic-index/programmatic-workflows/README.md`. Read its entrypoint
before searching that topic broadly; follow its task routes to the edited
syntheses, preserved verbatim notes, and the current shipped specification.
The Go-program runtime and unified context/index behavior described there are
direction, not current implementation.

## Reference submodules

`reference/` holds upstream projects mounted as submodules for reading only.
Never commit inside one and never push from one. See `reference/README.md`.

## Repository size and generated artifacts

- Never commit Gimble run directories, event streams, stage transcripts,
  prompts, responses, checkpoints, or raw tool logs. Keep them local. Commit a
  small human-authored summary only when the result is worth preserving.
- Do not vendor large third-party source trees, research corpora, binaries,
  archives, nested repositories, or generated dependency caches. Record our
  findings and cite the upstream source. If the repository genuinely needs to
  attach a substantial external codebase, use a submodule pinned to a reviewed
  upstream commit.
- Exercise judgment before staging broad changes. Review the staged diff and
  investigate unexpected file counts, added lines, or bytes. Do not split
  unwanted material across commits merely to evade the limits.
- `scripts/check-staged-content.sh` enforces the mechanical floor. Its large
  commit bypass is for an explicitly human-approved exception; agents must not
  set `GIMBLE_ALLOW_LARGE_COMMIT=1` on their own. The run-log prohibition is
  not bypassable.
