---
name: release-tractor
description: >
  Release, publish, tag, or bump Tractor and refresh its Codex plugin
  marketplace installation. Use for every Tractor version release, plugin
  version change, release tag, marketplace update, or request to ship a new
  version.
---

# Release Tractor

Ship a versioned Tractor release and prove the configured Codex marketplace
serves that exact version.

## Prepare the release

1. Read `AGENTS.md`, inspect the complete diff since the previous release, and
   confirm the requested semantic version. Do not invent a version when the
   user's intent leaves it materially ambiguous.
2. Work in a release worktree and require a clean, current `main` before and
   after the release, following the repository protocol.
3. Set `.codex-plugin/plugin.json` `version` to the plain release version on
   every release. Never ship the `+codex.<cachebuster>` suffix used for local
   plugin development.
4. Keep `.agents/plugins/marketplace.json` unchanged when its Tractor source
   still correctly tracks this repository's `main`. Change the catalog only
   when its source, ref, policy, category, or presentation metadata must change.
5. Update user-facing release notes or install documentation only when the
   shipped behavior or procedure changed.

## Verify before publishing

Run the repository's complete checks plus the plugin validations:

```sh
go test -race -count=1 ./...
go vet ./...
golangci-lint run ./...
python3 "${CODEX_HOME:-$HOME/.codex}/skills/.system/plugin-creator/scripts/validate_plugin.py" .
python3 "${CODEX_HOME:-$HOME/.codex}/skills/.system/skill-creator/scripts/quick_validate.py" .agents/skills/release-tractor
git diff --check
go test -tags=integration ./internal/workflows -run '^TestCanonicalLoop$' -count=1 -v -timeout=25m
```

The canonical loop proof is mandatory. Require its exit zero and fresh
`result.json` with `passed: true`, inspect the reviewer transcripts, and retain
the artifact path and binary hash in the release evidence. For a separately
built release binary, append `-args -tractor-binary <absolute-path>` to that
Go test command. It launches the actual workflow, with real native harnesses.
If source or binary changes afterward, rerun the proof.
A missing harness, quota failure or timeout blocks build-proof completion;
never replace this run with unit tests or a prior receipt.

Also run focused tests and live entry-point proof appropriate to changed
behavior. The canonical fixture proves the flat checklist loop; it does not
replace proof of services, nested loops or other changed surfaces.

## Publish and refresh Codex

1. Require the canonical loop proof above before publishing. Squash-merge the
   release PR to `main` and synchronize the root checkout.
2. For a versioned release, create and push `v<version>` at the verified merge
   commit. Do not move or overwrite an existing tag. Create a GitHub release
   only when requested or when the repository establishes that convention.
3. Refresh the Git-backed marketplace and reinstall the plugin:

```sh
codex plugin marketplace upgrade tractor --json
codex plugin add tractor@tractor --json
```

4. Verify `codex plugin list` reports `tractor@tractor` installed and enabled
   at the exact manifest version.
5. Start a fresh Codex process and exercise the installed marketplace copy. At
   minimum, call Tractor's deferred `get_pipeline_schema` MCP tool, or run a
   token-free example (`examples/yaml/multiline-tool.yaml`) through the
   installed binary; use stronger live run proof when runtime behavior changed.
6. Verify the remote tag, merge commit, clean synchronized root checkout, and
   installed plugin version before declaring the release complete.

The manifest version is the installed cache identity. Because the marketplace
entry tracks `main`, routine releases require a manifest version bump and the
upgrade/reinstall commands, not a cosmetic marketplace JSON edit.

## Closeout

Report the version, merge commit, tag, marketplace refresh result, installed
version, live proof, checks, and any uncompleted release step.
