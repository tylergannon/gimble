# Issue 130: native V2 session data and timeline port contract

## Selected authority

The selected same-revision source is OpenCode commit
`c55ee2a8152603f04a409163bd3edf79c425fbd7` in
`ephemeral/inspiration/opencode-newer`.

The event/state authority is `packages/client/src/solid/data.ts` (blob
`83c6eeb32ecccb1e4e9fb589647ea8e33f96b71e`), exported through
`packages/client/src/solid/index.ts`. The renderer authority is
`packages/session-ui/src/timeline/projection.ts` (blob
`47c3390088163b016c93e6d82309d9068c31b719e`),
`timeline-row.ts` (blob `752375beb88d3079a5dc5e9bea8085f7d1efb0d8`),
`session-timeline.tsx`, and `session-timeline-row.tsx`. Its document contract
is `packages/session-ui/src/document.ts` (blob
`94798fbf2d237949c9d17520a63b17056eaaa7c2`):
`{sessionID, messages: SessionMessageInfo[], status, diffs}`.

The generated DTO/schema authority is `packages/client/src/promise`, which
imports `@opencode/schema` and `@opencode/protocol`. Do not substitute the
older vendored 1.17.13 client or older reducer. The newer client package is
version 2.0.0, MIT (package blob `10510526b32bc01ea7ba3d3f358c14aa171ab66f`);
session-ui is private version 2.0.0, MIT (blob
`5d50aeec187fcd28314ad885dc43714f556ccffc`). Runtime peers are `solid-js
>=1.9.0` and Effect `4.0.0-rc.112` (optional peers). The tested workspace
locks Solid 1.9.15 with its upstream patch and Effect 4.0.0-rc.112; browser
tooling uses Vite 7.3.6 and Playwright 1.59.1. Preserve the actual MIT notice,
including `Copyright (c) 2025 opencode`, with borrowed source.

Additional Git blob pins at the selected revision:

| Source | Blob |
| --- | --- |
| `packages/schema/src/session-event.ts` | `6e43890ff8ae67d88f4990d710875f0d67b8db65` |
| `packages/schema/src/session-message.ts` | `7e8ce0482ae9db77bb1ddc8e9bb3fb0b4c41dcdf` |
| `packages/schema/src/session-inbox.ts` | `baadcc947b18c267e10fee76cb648f140ccdffe6` |
| `LICENSE` | `6439474beed8e0271df9862eff97ffd70ec2464c` |
| `bun.lock` | `30b8f7f4d5e89f3360728495701a4bbeda8a3732` |

## Actual updater boundary

`createData(config: CreateDataInput)` requires `api: () => OpenCodeClient`,
`directory: string`, and an event object with typed `on(type, handler)` and
broad `listen(handler)` receiving `{name, details: OpenCodeEvent}`. Optional
connection status and error callback affect refresh behavior. The returned
`Data` exposes session, project, location, shell, and event methods; event
application itself is private `handleEvent` at `data.ts` line 591.

The updater synchronously mutates a Solid store for session info, family,
active status, messages, inbox/pending items, permissions, forms, and reverts;
it also schedules API effects. Model selection hydrates its derived message;
rename/viewed and most execution terminals can invalidate and re-fetch
sessions; creation/deletion and inbox events coordinate outboxes and pending
reads. Global/location events refresh model, agent, command, skill, VCS, MCP,
forms, shells, and other caches.

Native session event families handled by `handleEvent` are lifecycle/info
(`created`, `deleted`, `usage.updated`, `agent.selected`, `model.selected`,
`renamed`, `permissions.updated`, `moved`, `viewed`, `revert.*`), inbox
(`inbox.enqueued`, `inbox.delivered`, `inbox.cancelled`,
`inbox.delivery.changed`), transcript (`instructions.updated`, `synthetic`,
`shell.*`, `step.*`, `text.*`, `reasoning.*`, `tool.*`, `retry.scheduled`),
and execution/compaction (`execution.*`, `compaction.*`).

Text and reasoning events require `ordinal: NonNegativeInt` in the schema, but
the native updater currently ignores it: `data.ts` uses `findLast` for text
and the latest incomplete reasoning. A validated fixture with text starts at
ordinals 0 and 1 followed by delta ordinal 0 updates the second (latest) text;
the analogous reasoning fixture updates the latest incomplete reasoning. This
is an observed schema/updater limit, not a port choice. The actual producer
prevents this ordering: its fragment helper rejects a second open part when
single is true, and both text and reasoning pass true
(`publish-llm-event.ts:135,196-248`). Borrow that producer invariant too. Admission is represented by
`SessionInboxInfo` in `store.session.pending`; user/synthetic admissions are
materialized into messages immediately. Tool identity is `data.id`. Missing stream
targets are dropped, with no hydration sentinel. These are observed source
facts, not inferred semantics.

## Full state and effects

The private `Store` snapshot is session `info/family/active/message/
messageCursor/messageLoading/pending/permission/form`; project
`info/permission`; and location info, VCS, agents, commands, config,
integrations, MCP servers/resources, models, providers, references, websearch,
shells, and skills. A native port must preserve this boundary or explicitly
scope out non-session caches. `SessionTimeline` additionally requires `diffs`.

Every fixture prefix must compare decoded public state after event delivery,
then compare the effect transcript and eventual state after deterministic fake
responses settle. Record API method/arguments, request order, response
payloads and invalidations. Execution inventory confirmed that the 150ms
`settleMs` timer is used only by the excluded MCP cache, not session branches. Async responses are
part of behavior: immediate and post-refresh snapshots may differ.

## Executable oracle

`ephemeral/inspiration/opencode-newer/packages/client/issue-130-oracle.ts` is an
ignored probe inside the exact checkout. It creates `createData` inside Solid
`createRoot`, captures the private updater through `event.listen`, seeds an
already-loaded session, and uses a deterministic fake API. Its `send` function
selects the exact `EventManifest.ServerDefinitions` entry by event type and
calls `Schema.decodeUnknownSync(definition)(data)` before every dispatch. Run:

```sh
cd ephemeral/inspiration/opencode-newer
bun run packages/client/issue-130-oracle.ts
```

The validated run passed under Bun 1.3.14 and printed 14 immediate session
snapshots, one after each event. It exercises step start; two text starts
and an ordinal-0 delta; two reasoning starts and an ordinal-0 delta; tool
input start/end, called, success; step end; execution start; and normal
user interruption. Both ordinal-0 deltas update the latest part (ordinal 1),
confirming the source behavior with schema-accepted inputs.

This particular script does not test authoritative text/reasoning end,
inbox lifecycle, or tool input delta. Those must not be inferred from the
older unvalidated scratch run. The separate browser fixture proves text
and reasoning final replacement. The oracle stops after immediate event
application: it does not await or assert the normal-interruption API
refresh, record its request transcript, or prove eventual state. These
remain required parity fixtures before a complete Go port claim.

The decoded-state output is retained in
[evidence/v2-oracle-prefixes.jsonl](evidence/v2-oracle-prefixes.jsonl).
The exact per-definition EventManifest decoder succeeded inside the newer
checkout with its locked dependencies; no structural substitute is used.

The upstream harness is `packages/client/test/solid-data.test.ts` (blob
`4260aa18ac6feb682ea5e833654851a19764e8b0`), with sibling
`solid-compaction.test.ts`, `solid-eviction.test.ts`, and `solid-refresh.test.ts`.
Those tests use `createRoot`, fake event buses, `OpenCode.make` HTTP clients,
and deterministic `Response.json` responses. They are the source-backed model
for hydration fixtures. Session-ui tests directly seed `SessionMessageInfo`;
they prove rendering, not event production.

## Port recommendation and hard limit

Port the selected pair as explicit seams: Go representations of the newer
`SessionMessageInfo`/`SessionInboxInfo`/event DTOs with raw JSON maps, and a
state machine whose public state/effect transcript is compared against the JS
oracle after every fixture prefix. Preserve arbitrary metadata/provider state/
tool input/error JSON and optional-field absence. Then compare the resulting
message document with native session-ui projection fixtures. Do not invent a
replacement event vocabulary.

Extracting `handleEvent` into a pure function changes behavior because it
closes over Solid store proxies, message indexes, outboxes, sync invalidation,
timers, API reads, connection state, cleanup, and request races. A copied pure
reducer would lose inbox reconciliation, stale-read protection, status updates,
hydration, and refresh effects. Keep the JS implementation pristine and use
`createData` as oracle; only the Go port may be separate.

The older pinned `95daf90670b7c039c436c85537da5fbfe2205b41` reducer and its
1.17.13 client are a rejected alternative: ordinal streams, input admitted/
promoted pending Map, `callID` tool fields, and `missing` hydration results
differ materially from this selected native revision.
