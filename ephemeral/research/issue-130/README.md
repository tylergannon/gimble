# Issue 130: native session observation

The implementation borrows OpenCode's `session.*` schema and matching
`packages/client/src/solid/data.ts` reducer at revision
`c55ee2a8152603f04a409163bd3edf79c425fbd7` (MIT).
That revision has a direct native V2 UI, exercised with schema-validated
streaming fixtures. The Go and TypeScript ports are compared with its actual
JavaScript implementation after every fixture prefix.

Every SSE connection begins with complete current state replacing the client
state, then incremental events. There is no cursor recovery.

- [Contract](contract.md): API and transport decisions.
- [Port contract](port-contract.md): upstream semantics and oracle.
- [Sprint artifact](SPRINT.md): original plan and scope adjustments.
- [Delivery](execution/DELIVERY.md): demonstrated behavior and remaining gaps.
- [Runnable browser proof](proof/): production handler and streaming scenario.

Downloaded source checkouts, exploratory reports, logs, and screenshots remain
local and ignored. Only the selected contract, sprint, delivery record, and
repeatable proof are committed. Upstream provenance and license notices are in
`third_party/opencode` and `internal/sessionstate/NOTICE.md`.
