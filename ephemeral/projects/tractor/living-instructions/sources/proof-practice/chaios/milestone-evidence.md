# Application proof milestone evidence

## Milestone 1: empty product shell and durable workspace

- Running-system claim: `CHAOS_PORT=43171 bun run start` started the real Bun
  entry point at `http://127.0.0.1:43171/`. `GET /` returned permanent browser
  chrome containing an exactly empty `#canvas`, composer, operational status,
  hidden recoverable-error region, and retry control. The browser assets were
  served successfully as CSS and JavaScript.
- Workspace claim: live `GET /api/bootstrap` returned HTTP 200 with application
  `application-proof`, status `Ready`, and `"canvas":{"components":[]}` from
  the tracked workspace manifest rather than an in-browser default.
- Recovery claim: the focused test supplies a malformed durable manifest and
  observes HTTP 500 with `"recoverable":true`; browser bootstrap exposes that
  failure in the error region and its retry control repeats the load.
- Engineering checks: `bun test` passed 3 tests with 0 failures, and
  `bun build src/server.ts --target=bun --outdir=<temporary-directory>` bundled
  the server successfully.
- Scope check: `GET /api/events` still returns HTTP 404 and the composer submit
  control remains disabled. SSE, mutation handling, compilation, CLI, and
  app-server integration remain for their own milestones. No spending-tracker
  source or feature was added.

## Milestone 2: Bun-owned state revisions over SSE

- Running-system claim: `CHAOS_PORT=43172 bun run start` started the public Bun
  entry point. Two simultaneous live `GET /api/events` clients each received
  the revision-0 empty snapshot. A live `POST /api/mutations` setting
  `proof.message` returned HTTP 202 with revision 1, after which both clients
  received the same `state.revision` event containing the Bun-accepted value.
- Browser-boundary claim: the served application client opens one EventSource,
  posts only validated `state.set` requests, retains the latest server snapshot,
  and fans each streamed revision out to every path subscriber. Its focused
  test delivered one revision to two independent subscribers.
- Validation claim: mutation bodies must contain JSON values and bounded safe
  paths; prototype-escaping path segments are rejected with HTTP 400.
- Engineering checks: `bun test` passed 6 tests with 0 failures. Independent
  Bun builds of the server target and browser entry target both succeeded.
- Scope check: the canvas remains genuinely empty and the composer remains
  disabled. State is deliberately process-local until the restoration
  milestone; no Svelte compilation, component installation, CLI, app-server,
  or spending-tracker behavior was added.

## Milestone 3: live content-addressed Svelte components

- Running-system claim: `CHAOS_PORT=43173 bun run start` started the public Bun
  entry point. One uninterrupted `/api/events` stream received an installation,
  a replacement more than ten seconds later, and a removal for `proof-card`;
  the browser lifecycle path consumes those events without navigation or reload.
- Compilation and identity claim: arbitrary proof source compiled through the
  one initialized `@rsvelte/compiler` and an in-memory virtual-file `Bun.build()`.
  The first served bundle was 48,463 bytes at SHA-256
  `faf2c46f72de20940d57dea48e9f2cd88675baafb4b89e62ae847e215509604f`;
  an independent `shasum -a 256` matched its artifact URL. Revised source
  returned `replaced: true` and the distinct hash
  `616b2250c9b55df7227b5448eb8c25e38f7499fd829fc4eb3ed914fa0d3a98e3`.
- Browser-boundary claim: the served client accepts only same-origin
  content-addressed descriptors. Its component host dynamically imports the
  bundle, mounts into the canvas, swaps an existing id after calling its bundled
  Svelte disposal boundary, and disposes/removes it on the removal event.
- Diagnostic claim: live malformed source returned HTTP 422 with the rsvelte
  `element_unclosed` diagnostic and emitted no accepted installation.
- Engineering checks: `bun test` passed 10 tests with 0 failures and 60
  expectations. Independent Bun builds of the server and browser entry targets
  succeeded; `git diff --check` passed.
- Scope check: component source, artifacts, and composition remain process-local
  until the restoration milestone. No CLI, app-server, composer turn, canned
  spending tracker, Vite, or Rolldown was added. The managing agent's complete
  Chrome-operated acceptance milestone remains unchecked.

## Milestone 4: application-local chaos CLI

- Running-system claim: `CHAOS_PORT=43174 bun run start` started the public Bun
  entry point, and the executable `product/workspace/chaos` reached it through
  `CHAOS_RUNTIME_URL=http://127.0.0.1:43174`. `chaos inspect` reported the
  durable application identity, revision-0 empty state, and empty composition.
- State claim: one live `chaos state patch` transaction added object and array
  values and returned accepted revision 1; `chaos state get /proof/via` then
  reported `"application-local CLI"` from Bun-owned state. State paths use safe
  JSON Pointers, and patches accept bounded add, replace, and remove operations.
- Compilation and installation claim: live `chaos install` compiled arbitrary
  `/tmp/chaios-milestone-four-proof.svelte`, installed `cli-proof` in `canvas`,
  and returned content hash
  `959d501ec5d874101549af7197e579ef14f4dfe6e2740522a2777e786db41795`.
  A following `chaos inspect` reported that exact live component descriptor;
  `chaos remove cli-proof` returned `{"removed":true,"id":"cli-proof"}`.
- Diagnostic claim: installing malformed source through the same executable
  exited 1, preserved HTTP 422, and printed the rsvelte `element_unclosed`
  diagnostic so Codex can repair the file and retry.
- Engineering checks: `bun test` passed 11 tests with 0 failures and 75
  expectations. Independent Bun builds of the server, browser entry, and CLI
  succeeded; `git diff --check` passed.
- Scope check: the CLI is a narrow wrapper around Bun HTTP operations. No
  app-server child, composer turn, persistence, canned spending tracker, Vite,
  or Rolldown was added, and the Chrome-operated acceptance milestone remains
  unchecked.

## Milestone 5: managed Codex app-server and persisted thread

- Running-system claim: `CHAOS_PORT=43175 bun run start` started the real Bun
  entry point and its direct child `codex app-server --listen stdio://` under
  Codex CLI 0.147.0. Startup completed the stable `initialize` / `initialized`
  handshake, verified the inherited account was ChatGPT-managed, and started
  thread `01a01317-f8fd-7123-b1f5-28800d34e9bb` with the application workspace
  as `cwd`, `approvalPolicy: "never"`, workspace-write sandboxing, and network
  access for the local runtime CLI.
- Persistence claim: the new thread was named through the stable app-server
  API without starting a turn, then atomically recorded in the application
  manifest. After SIGINT stopped Bun and its child, a second public
  `bun run start` resumed the exact same thread id; `/api/inspect` still
  reported revision-0 empty state and an empty canvas.
- Authentication and lifecycle claim: the integration rejects non-ChatGPT
  account modes rather than adding an API-key path. Both real shutdowns exited
  0, and process inspection confirmed each app-server was Bun's child and the
  first child was gone before restart.
- Engineering checks: `bun test` passed 13 tests with 0 failures and 86
  expectations. Independent Bun builds of the server and browser entry
  succeeded; `git diff --check` passed.
- Scope check: the composer remains disabled and no `turn/start`, agent-message
  rendering, component persistence, generated feature, spending tracker, Vite,
  or Rolldown was added. The managing agent's Chrome-operated acceptance
  milestone remains unchecked.

## Milestone 6: composer-driven real Codex turns

- Running-system claim: `CHAOS_PORT=43176 bun run start` started the public Bun
  entry point and resumed ChatGPT-managed thread
  `01a01317-f8fd-7123-b1f5-28800d34e9bb`. A live composer-equivalent
  `POST /api/turns` returned HTTP 202 for real turn
  `01a01326-7002-7f61-8bad-c4fba442c9ef`; the persisted rollout contains that
  turn and its `./chaos inspect` command activity.
- Operational-status claim: the uninterrupted live `/api/events` stream
  observed `Starting Codex turn`, `Interpreting request`, `Running application
  tools`, and terminal `No application change produced` statuses correlated to
  the real turn id. The composer disables during active work and becomes usable
  again on terminal status.
- Non-chat claim: app-server `agentMessage` items and deltas have no runtime
  event mapping and are never placed in browser state or markup. The focused
  integration injected distinctive assistant prose before command activity and
  established that the next browser event contained only compact operational
  status; the served shell contains no assistant or transcript surface.
- Scope and state claim: the transport-proof request explicitly performed no
  edits or installation. Live inspection remained at revision 0 with empty
  state and canvas, and the application workspace had no tracked changes. The
  spending tracker and complete Codex-to-Svelte-to-live-UI slice remain for the
  next milestone.
- Engineering checks: `bun test` passed 14 tests with 0 failures and 103
  expectations. Independent Bun builds of the server and browser entry
  succeeded; `git diff --check` passed. The managing agent's complete
  Chrome-operated acceptance milestone remains unchecked.

## Milestone 7: real Codex creates and repairs a live component

- Empty-start claim: `CHAOS_PORT=43177 CHAOS_RUNTIME_URL=http://127.0.0.1:43177
  bun run start` used the public entry point, resumed ChatGPT-managed thread
  `01a01317-f8fd-7123-b1f5-28800d34e9bb`, and the held-open event stream began
  with revision 0, empty state, and no canvas component.
- Creation claim: a real composer `POST /api/turns` requested a neutral Focus
  Pulse session counter. Turn `01a01332-3e93-7452-9310-fca947b204fe` inspected
  the application, authored new self-contained Svelte source, initialized
  Bun-owned state through `./chaos state patch`, and invoked `./chaos install`.
  The uninterrupted stream emitted component `focus-pulse` at artifact
  `211607fbdbac55a5ac96ec71b7fd1dc49eba0bad066da8218869232d84689024` and
  ended with `Application updated`.
- Repair claim: without restarting the runtime, a second composer request asked
  Codex to repair that component with a reset control while preserving its
  increment behavior. Turn `01a01335-202e-7710-a794-38cf92c5a426` revised the
  same source and reinstalled a distinct replacement artifact,
  `87a259f23ed83152727655bb46ad81e87dc2a37383b42ea32b1dea3006648eec`.
  The same live stream then observed server-owned reset and increment results
  as state revisions 4 and 5 and terminal `Application updated`.
- Artifact and causality claim: SHA-256 over the bytes served at the replacement
  artifact URL exactly matched its advertised content hash. The persisted Codex
  rollout records both turn ids, source creation/revision, `chaos` invocations,
  and runtime inspection; no model reasoning or authentication material is
  copied into this evidence.
- Durable implementation: the application workspace now states that Bun is
  already running and gives Codex the minimal inspect, author/revise, state
  patch, install, diagnostic repair, and final-inspect loop. A focused test
  guards that handoff and rejects spending-tracker instructions in the durable
  prompt.
- Engineering checks: `bun test` passed 15 tests with 0 failures and 109
  expectations. In-memory Bun builds succeeded for the server (44,902 bytes),
  browser entry (10,813 bytes), and CLI (3,346 bytes across two outputs), and
  `git diff --check` passed.
- Scope and cleanup: the Codex-authored proof source was removed after evidence
  capture, the tracked manifest remains an empty canvas, no spending tracker or
  persistence behavior was added, and no Vite or Rolldown dependency exists.
  The complete Chrome-operated acceptance milestone remains unchecked for the
  managing agent.

## Milestone 8: durable application restoration

- Running-system claim: `CHAOS_PORT=43178 CHAOS_WORKSPACE_DIRECTORY=<isolated
  workspace> bun run start` used the public entry point, resumed ChatGPT-managed
  thread `01a01317-f8fd-7123-b1f5-28800d34e9bb`, accepted state revision 1,
  and installed neutral component `restoration-proof`. SIGINT stopped Bun and
  its managed app-server child with exit 0; a second public start resumed the
  same thread without replaying either accepted change.
- Restoration claim: after the complete process restart, live `/api/inspect`
  returned revision 1 with the exact server-owned value and the same canvas
  placement. The initial SSE stream replayed that restored state snapshot and
  a component installation descriptor, so a newly loaded browser can mount the
  restored composition without a new Codex turn.
- Durable-material claim: the workspace retained the accepted Svelte source at
  `components/restoration-proof.svelte`, its manifest source reference, canvas
  composition, thread id, state revision and value, and artifact identity
  `75575fccb6d8f95eb535cf093207557fdd39ce9345b7536c4fba78ac868cbefc`.
  SHA-256 over both the durable bundle and the post-restart HTTP artifact bytes
  independently matched that identity.
- Engineering checks: `bun test` passed 16 tests with 0 failures and 122
  expectations. Independent Bun builds of the server, browser entry, and CLI
  succeeded; `git diff --check` passed.
- Scope check: the tracked application still begins with revision-0 empty state
  and canvas. No spending tracker, budgeting, advice, Vite, or Rolldown was
  added, and the managing agent's Chrome-operated acceptance milestone remains
  unchecked.

## Milestone 9: retained expenses and live budgeting revision

- Real-turn claim: `CHAOS_PORT=43179 CHAOS_RUNTIME_URL=http://127.0.0.1:43179
  CHAOS_WORKSPACE_DIRECTORY=<isolated workspace> bun run start` used the public
  entry point and the tracked empty application material. The exact initial
  spending-tracker request started real turn
  `01a01353-1944-7c30-9749-4409432d7197`, which authored and installed
  `spending-tracker` at artifact
  `c132ac9c2a9ed1d97970985cb2add927252a181fd839f9bdf241e61d786de3a5`.
- Retention claim: the generated expense form uses the application bridge for
  amount, merchant, category, and date mutations. Three live HTTP mutations,
  matching that bridge operation, produced SSE revisions 2 through 4 with
  Dining expenses of $18.75 and $86.40 and a Groceries expense of $64.20.
  Without restarting or reconnecting the held-open event stream, the exact
  budgeting request started later turn
  `01a01354-7e83-7e83-9e76-3dda7cba18de`; its accepted revision 5 retained all
  three expenses while adding the budget state and replaced the same component
  with artifact
  `a53c1c036788cad0b946a51c04d8e988bd442531355d58aa5160f3871c764a6a`.
- Usability and overspending claim: the accepted compiled Svelte revision
  subscribes independently to expenses and budgets, provides month, category,
  and limit controls, POSTs both state paths through `application.mutate`, and
  renders remaining amounts. Live budget mutations reached revisions 6 and 7;
  an independent calculation found Dining $105.15 spent against $80.00, or
  $25.15 over, and Groceries $64.20 against $100.00, or $35.80 remaining. The
  accepted UI labels the negative result `Over budget` and applies a distinct
  red border, background, amount, and progress treatment.
- No-reload and identity claim: one uninterrupted SSE connection observed the
  revision-0 empty snapshot, initial installation, all seven state revisions,
  and the later content-addressed replacement. Independent SHA-256 checks
  matched both durable bundle identities. No runtime restart or event-stream
  reconnect occurred between turns. The in-app browser reported no available
  browser surface, so this milestone does not claim screenshot proof; complete
  visible Chrome acceptance remains the managing agent's separate milestone.
- Engineering checks: `bun test` passed 16 tests with 0 failures and 122
  expectations. Independent Bun builds of the server, browser entry, and CLI
  succeeded; the focused persisted-state/source assertion and
  `git diff --check` passed.
- Scope check: the generated tracker, artifacts, expenses, and budgets exist
  only in the isolated proof workspace. The tracked application still starts
  empty, no feature source is preinstalled, and the advice, semantic-trace, and
  complete Chrome-operated milestones remain unchecked.
