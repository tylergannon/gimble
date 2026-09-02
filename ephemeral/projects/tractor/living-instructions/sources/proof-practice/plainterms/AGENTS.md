# Plainterms Agent Start

## Route

- App/issue/proof: `docs/coding-guide/integration-recipes/testing-proof.md`
- Docs: `docs/README.md`
- Architecture/location/drift: `docs/design/README.md`
- Code/frameworks: `docs/coding-guide/README.md`
- Product/glossary/invariants: `docs/product/README.md`
- Fixture scoring: `VALIDATE.md`, `SCORING.md`
- Authority: code/migrations > tests/generated registries > current docs > worklogs > old research.

## Operate

- Worktree only.
- Commit coherent checkpoints; do not park finished diffs.
- Log non-trivial work in `sprints/worklog/`.
- After dev/Supabase/Docker/Overmind, tear down before final.
- App proof: `pnpm run dev`; E2E loop: `pnpm run dev:e2e:loop`.
- Final: run `git status --short --branch`; report material residue.

## Proof

- App behavior/PR/cleanup: active Proof of Work.
- Proof of Work: start with CI Confidence coverage design; finish with CI
  Confidence test repair after behavior proof.
- Docs/policy/scripts/tooling: targeted repo proof only.
- No repo-local runbooks, probe banks, journals, or skills that restate global
  workflows.
