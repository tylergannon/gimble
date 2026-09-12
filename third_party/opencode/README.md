# OpenCode session-state oracle

The oracle runs the untouched `createData` from OpenCode commit
`c55ee2a8152603f04a409163bd3edf79c425fbd7`. The source is MIT licensed;
the preserved notice is in `LICENSE`. The TypeScript reducer in
`web/src/lib/sessionstate` is an adaptation, not an upstream file.

Run `./prepare.sh` once in a fresh checkout. It fetches only the pinned commit,
checks out a detached worktree under `oracle/upstream`, verifies the authority
blobs and lockfile, and installs the locked Bun workspace dependencies. Normal
oracle runs use that prepared tree and do not access the ignored inspiration
checkout or a moving branch.

The harness does not edit or replace `createData`. It dynamically imports the
pinned file and wraps Solid's `createStore` only to retain the exact store proxy
that `createData` already mutates; event handling and API scheduling still run
through the upstream implementation.

The preparation contract is Bun 1.3.14, Solid 1.9.15 with the upstream patch,
and Effect 4.0.0-rc.112. `prepare.sh` rejects another Bun version. Delete
`oracle/upstream` to force a clean, hash-checked preparation.

Authority blobs:

| path | Git blob |
| --- | --- |
| `packages/client/src/solid/data.ts` | `83c6eeb32ecccb1e4e9fb589647ea8e33f96b71e` |
| `packages/schema/src/session-event.ts` | `6e43890ff8ae67d88f4990d710875f0d67b8db65` |
| `packages/schema/src/session-message.ts` | `7e8ce0482ae9db77bb1ddc8e9bb3fb0b4c41dcdf` |
| `packages/schema/src/session-inbox.ts` | `baadcc947b18c267e10fee76cb648f140ccdffe6` |
| `LICENSE` | `6439474beed8e0271df9862eff97ffd70ec2464c` |
| `bun.lock` | `30b8f7f4d5e89f3360728495701a4bbeda8a3732` |
