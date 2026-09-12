# Sprint 001: Token usage by scope

## Pyramid Index

- L0: Fold scopes and turns into the run observation, make the three
  adapters report the same cached/new split with per-field availability,
  and give the run page one tree of scopes, turns, and messages with
  per-model token totals and wall-clock bars, from the same code live and
  after the run.
- L1:
  - Observation: `RunSnapshot` gains `scopes` (key, name, parent, task,
    began, ended, error) and `turns` (scope, session, prompt, output type,
    started, ended, duration, error, interrupted), folded in Go from the
    lifecycle records the run already writes, mirrored by the browser fold,
    carried by the checkpoint.
  - Usage: one TypeScript module defines usage as an additive value per
    model and per field, each field a cell of tokens plus an unknown count.
    A scope's usage is its direct turns plus its child scopes. Go computes
    no totals. Tool-only assistant rows are not model calls.
  - Adapters: a live warm-cache probe on all three harnesses comes first.
    Then Antigravity's `input` excludes cache read, Claude's `output` no
    longer depends on a reasoning split, Codex's `input` no longer depends
    on a cache-write field, a negative derived value is unavailable, and
    `tokensAvailable` is deleted. The page reads `fieldAvailability` only.
  - Page: the run page becomes a collapsible tree from the root scope. Each
    row has a name, a bar on the run's wall clock, and five cells: input
    (new), cache read, cache write, output, reasoning. Expanding a row is
    how a person picks the level; an expanded row shows its per-model
    table, an expanded turn its prompt, an expanded message links to its
    transcript row. No tabs, no extras.
  - Proof: fake-adapter tests for the fold and the usage value; one live
    workflow on the cheap tier with a direct root turn, a warm-cache
    scope, a concurrent group, and a nested scope across all three
    harnesses; a reconcile script against the logs; live and finished
    screenshots; a restart against a directory holding only
    `observation.json`.
  - Out: cost and price tables; retyping `TurnEnded.Tokens`; the OpenCode
    projection port; a Usage tab; segmented bars, duration columns, a
    ticking clock.
- L2:
  - Architecture § "Where things live" answers the Go-versus-browser
    question; § "The usage value" defines the cell and the invariant;
    § "Provider accounting" is the normalization contract and what the
    probe must show; § "The page" is the layout.
  - Implementation Plan phases 0 to 5 carry files and tasks; Phase 0 is the
    probe that gates Phase 3.
  - Definition of Done maps each success criterion in the intent to what is
    seen working.

## Overview

Every model call already leaves a token record. All three harnesses emit
`session.step.ended` with `tokens {input, output, reasoning, cache {read,
write}}` and `session.step.started` with `model {providerID, id}`; the
projection keeps both on the assistant message row with `time.created` and
`time.completed` in milliseconds. Every message is placed by scope key,
session id, and turn id, and the scope tree is the prefix tree of keys
(`ephemeral/research/api/API.md` § Keys). The sum for any node is already
determined by data on disk. This sprint invents no identity and no event.

What is missing is structure and a view. The observation snapshot has no
scopes and no turns, so a finished run cannot say when a scope began or
what a turn's prompt was. The run page is a flat list of turns. And the
adapters disagree on what `input` means and mark a whole message
unavailable when one field is missing, which is why the only saved
Antigravity run shows "unavailable" beside real numbers.

This sprint:

1. Folds `scope_began`, `scope_ended`, `turn_started`, `turn_ended` into
   the observation store, its checkpoint, and the browser's mirror fold.
2. Defines usage once, in the browser, as an additive value with per-field
   availability, and rolls it up the scope tree from message rows.
3. Probes each harness live, then fixes the adapter inconsistencies the
   probe confirms and deletes `tokensAvailable`.
4. Replaces the run page's flat list with one tree: scopes, turns, and
   messages, each with a wall-clock bar and five token cells, per-model
   detail under any row, live and post-run from one code path.
5. Proves it with fake-adapter tests, one live run on the cheap tier whose
   page numbers reconcile with its logs, and a checkpoint-only restart.

Cost is out: no adapter reports it and a price table is a second source of
truth nobody has asked for. This sprint stops at tokens.

## Use Cases

1. **Where did the budget go.** Tyler opens a finished sprint run and sees,
   for the whole run, tokens per model split into input (new), cache read,
   cache write, output, reasoning. He expands `loop.1`, sees each
   `task.N` lap's share, then the coder's turns inside one lap with their
   prompts, then the individual model calls of the turn that blew up.
2. **Live watch.** A run is going. The tree grows as scopes begin, bars
   extend as frames arrive, totals rise as steps end. No reload; the same
   rows the finished run will show.
3. **Compare candidates.** A bake-off group has two `attempt` scopes whose
   bars overlap on the axis. The overlap says they ran concurrently; the
   cells say which was cheaper.
4. **Is the cache working.** A per-model line under the `research` scope
   shows cache read against input (new) for a Claude Haiku session across
   two turns. The split is comparable to the Codex candidates in the same
   run because all three adapters define `input` the same way.
5. **Honest gaps.** An Antigravity session cannot report cache write. Its
   rows show cache write as unavailable with the number of calls that
   lacked it; every other field still sums. The run's cache write cell
   shows the known sum and how many calls are unknown, never a zero that
   reads as "no cache writes".
6. **A program reads the checkpoint.** A post-mortem agent given the run
   directory reads `observation.json` and finds scopes with began and
   ended and turns with prompts, without replaying `run.jsonl`.
7. **Direct work in a parent.** A turn placed in the root scope beside
   child scopes appears as a row under the root and is counted once. Root
   equals its direct turns plus its child scopes.

## Architecture

### What is already true

- `run.jsonl` has every lifecycle record: `scope_began` (name, task),
  `scope_ended` (error), `turn_started` (prompt, output type), `turn_ended`
  (result, error, tokens, duration, interrupted), each with `seq`, `time`,
  scope key, session id, turn id (`events.go`, `event_persistence.go`).
- `run.observeLifecycle` (`run.go`) hands each record to
  `observation.Store.Lifecycle` with the fields the store folds today: run
  status, run name, session info. The record bytes are republished
  unchanged as a `lifecycle` frame.
- `RunSnapshot` is `run` (id, name, status, error, sessions) and
  `invocations` (per turn: scope, session, turn, projection snapshot,
  provenance keyed by normalized message id). The browser's
  `RunObservation` mirrors it and applies frames incrementally.
- Assistant rows carry `model {providerID, id}`, `tokens`, `cost`, `time
  {created, streamed, completed}`. The provenance sidecar's `accounting`
  carries `fieldAvailability {input, output, reasoning, cacheRead,
  cacheWrite}` and `rawProviderAccounting`. Every adapter attaches
  `accounting` on the `session.step.ended` sidecar, and that sidecar is the
  last one folded for its message (checked on the saved Antigravity log;
  Claude clears its message id after the step ends). A completed assistant
  row with no `accounting` sidecar is not a model call: the Antigravity
  tool step is one.
- `Placement.Scope` on an invocation is the key of the scope the turn ran
  in, which may be below the scope that created the session. That is the
  key usage is charged to.
- The root scope's key is `""` and its recorded `scope_began` name is `"."`
  (`path.Base("")`).
- `web/src/hooks.ts` parses the transported snapshot as open JSON; new
  fields cross without a transport change. `page.server.go` serializes
  whatever `RunSnapshot` holds.

### Where things live

The fold of scopes and turns is in Go, in the store, because the checkpoint
must carry them and the store is the one producer. The browser fold mirrors
it, as it already does for sessions, so a live page and a replayed
checkpoint hold the same structure.

The usage value and the rollup live in the browser only, in
`web/src/lib/observation/usage.ts`. The browser already holds every message
row for live and finished runs; the store stays a producer that never
decodes token JSON on the event path; a total computed in Go would still be
recomputed per frame in the browser, or the store would publish usage
frames, either of which is a second definition. The checkpoint is
sufficient without totals: scopes, turns, rows, and provenance determine
them. If a Go program later needs totals, it is a program reading
`observation.json` and summing by prefix, written when the workflow that
needs it exists.

`TurnEnded.Tokens` stays `[]JSONText`. The page reads message rows, which
also carry the model. The proof uses `turn_ended` tokens only as an
independent check that the page agrees with the log.

### The snapshot shape

Go, `internal/observation/snapshot.go`:

```go
// ScopeInfo is one scope instance, from scope_began and scope_ended.
// Times are Unix milliseconds, the unit message rows use. Ended is zero
// while the scope is open. The root is Scopes[""] and its Name is the run
// name, not the recorded ".".
type ScopeInfo struct {
	Key    string          `json:"key"`
	Name   string          `json:"name"`
	Parent string          `json:"parent"`         // key minus its last segment; "" for a child of the root and for the root
	Task   json.RawMessage `json:"task,omitempty"` // the Task scope_began carried, when it did
	Began  int64           `json:"began"`
	Ended  int64           `json:"ended,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// TurnInfo is one turn, from turn_started and turn_ended. It exists from
// turn_started, so a turn that fails before its first native message is
// still visible.
type TurnInfo struct {
	ID          string `json:"id"`
	Scope       string `json:"scope"`
	Session     string `json:"session"`
	Prompt      string `json:"prompt"`
	OutputType  string `json:"outputType"`
	Started     int64  `json:"started"`
	Ended       int64  `json:"ended,omitempty"`
	Duration    int64  `json:"duration,omitempty"` // ms; the record carries nanoseconds
	Error       string `json:"error,omitempty"`
	Interrupted bool   `json:"interrupted,omitempty"`
}

type RunSnapshot struct {
	Run         RunInfo               `json:"run"`
	Scopes      map[string]ScopeInfo  `json:"scopes"`
	Turns       map[string]TurnInfo   `json:"turns"`
	Invocations map[string]Invocation `json:"invocations"`
}
```

`observation.Lifecycle` grows a `Time time.Time` and one optional pointer
per folded variant (`ScopeBegan{Name, Task}`, `ScopeEnded{Error}`,
`TurnStarted{Prompt, OutputType}`, `TurnEnded{Duration, Error,
Interrupted}`), the pattern it already uses for `Session`.
`writeLifecycle` returns the record's time beside its bytes and
`run.observeLifecycle` forwards it, so the fold and the log agree to the
millisecond and nothing calls `time.Now()` twice. The task is marshalled
only when the `polytype.Optional[Task]` is present, so the snapshot does
not inherit the `""` the log shows for an absent task.

Browser, `web/src/lib/observation/index.ts`: `RunSnapshot` gains `scopes`
and `turns` with the same field names; `RunObservation` holds them as maps;
`replace` clears and refills them; `foldLifecycle` gains the four cases,
converting the record's RFC 3339 `time` with `Date.parse` and the duration
from nanoseconds to milliseconds. The browser fold also keeps `cancelled`
over a later `run_ended`, as the Go store does. A checkpoint written before
this sprint lacks `scopes` and `turns`; `loadCheckpoint` and the browser
default them to empty and the page shows an empty tree. No other shim.

### The usage value

`web/src/lib/observation/usage.ts` is the one definition:

```ts
export type Field = 'input' | 'cacheRead' | 'cacheWrite' | 'output' | 'reasoning'
export const FIELDS: readonly Field[]
/** tokens summed over calls that reported the field; unknown counts calls that did not. */
export type Cell = { tokens: number; unknown: number }
/** calls is the number of model calls that contributed. */
export type ModelUsage = { calls: number } & Record<Field, Cell>
/** keyed by `${providerID}/${modelID}`; a row with no model uses the session's configured model. */
export type Usage = Record<string, ModelUsage>

export function messageUsage(row: JSONObject, provenance: unknown, fallbackModel: string): Usage | undefined
export function add(into: Usage, from: Usage): Usage
export function empty(): Usage
export function total(usage: Usage): ModelUsage   // across models
```

Rules:

- A completed assistant row (`time.completed` set) whose provenance sidecar
  carries `accounting` is one model call. For each field, if
  `fieldAvailability[field]` is true, `tokens += n`; otherwise
  `unknown += 1`. A measured zero is a zero.
- A completed assistant row without an `accounting` sidecar is not a model
  call and contributes nothing (the Antigravity tool step). A row still in
  flight contributes nothing; it is pending, not unknown.
- `Usage` is a commutative monoid under `add`, so any grouping sums to the
  same value.
- A cell displays: `tokens` when `unknown === 0`; `tokens +N?` when both
  are known and some calls lacked the field (title: "N calls did not
  report this field"); `unavailable` when no call reported it; and a row
  with `calls === 0` shows `no model calls`.

Rollup, same module:

```ts
export type Rollup = {
	scopes: Map<string, Usage>   // by scope key; '' is the run
	turns: Map<string, Usage>    // by turn id
	messages: Map<string, Usage> // by `${turn}\0${messageID}`
}
export function rollup(observation: RunObservation): Rollup
export function contains(scope: string, key: string): boolean // scope === '' || key === scope || key.startsWith(scope + '/')
```

A turn's usage is the sum of its model-call rows. A scope's usage is the
sum over turns whose `invocation.scope` the scope contains. Containment is
equality or prefix followed by `/`, so `attempt.1` does not contain
`attempt.10`; the root contains everything. The invariant, per model and
per field including unknown counts:

```text
turn  = its model-call rows
scope = its direct turns + its immediate child scopes
run   = the root scope
```

Cost: `RunObservation` gains a per-turn revision, bumped in `apply` for
every event frame of that turn and reset by `replace`; `rollup` caches turn
usage by it. A frame recomputes one turn's rows and re-sums the tree. The
derived rollup is dropped whole on a replacement snapshot, so a reconnect
cannot count anything twice. Measured on the proof run, not asserted.

### Provider accounting

The contract, the one Codex already meets: `input` is what the provider
charged as fresh input, excluding anything counted as cache read or cache
write; `output` excludes `reasoning`; each of the five fields is available
on its own. A derived field is available only when every field it is
derived from is present, and unavailable when the subtraction goes
negative; the raw numbers stay in `rawProviderAccounting`.

| Field | Antigravity (`agy`) | Codex | Claude |
|---|---|---|---|
| input (new) | `input_tokens - cache_read_tokens`, hypothesis, gated on Phase 0 | `inputTokens - cachedInputTokens - cacheWriteInputTokens` (cache write only when present) | `input_tokens` (already excludes cache) |
| cache read | `cache_read_tokens` | `cachedInputTokens` | `cache_read_input_tokens` |
| cache write | not reported: unavailable | `cacheWriteInputTokens` when present | `cache_creation_input_tokens` |
| output | `output_tokens - thinking_tokens` | `outputTokens - reasoningOutputTokens` | `output_tokens - output_tokens_details.thinking_tokens` when the detail is present; otherwise the total, with reasoning unavailable |
| reasoning | `thinking_tokens` | `reasoningOutputTokens` | `output_tokens_details.thinking_tokens` when present |

Today all three adapters set `available = inputOK && outputOK &&
reasoningOK && cacheReadOK && cacheWriteOK` and Claude additionally marks
`output` unavailable unless the reasoning detail is present, Codex marks
`input` unavailable unless a cache-write field is present. Anthropic's
usage block carries no `output_tokens_details` and OpenAI has no cache
write, so live Claude and Codex messages are likely unavailable by
construction today. Phase 0 shows what each harness actually carries;
Phase 3 changes what the probe confirms. `tokensAvailable` is deleted from
all three sidecars and the page's `accounting()` with it: the page reads
`fieldAvailability` only. `rawProviderAccounting` stays so the proof can
check normalization against the raw numbers.

The OpenCode projection (`internal/sessionstate`, `web/src/lib/sessionstate`)
is not touched.

### The page

One tree, one level-picking gesture: expand. The run page is the tree; the
transcript sections it renders today sit below it and are the drill target
for a message.

```text
+---------------------------------------------------------------------------------------+
| Run usage-proof                     completed        live                             |
| claude-haiku-4-5   in 1,204  cache r 9,876  cache w 2,010  out 640  reason 0         |
| gpt-5.6-luna       in 3,110  cache r 0      cache w unavail  out 220  reason 180     |
| gemini-3.8-flash   in 14,980 cache r 0      cache w unavail  out 124  reason 0       |
+---------------------------------------------------------------------------------------+
| name                     | 0s ---------------- 84s  |  input | cache r | cache w | out | reason |
|--------------------------|--------------------------|--------|---------|---------|-----|--------|
| [-] usage-proof          | [======================] | 19,294 |   9,876 | 2,010+2?| 984 |    180 |
|   turn 1 (opener)        | [=]                      |    ... |         |         |     |        |
|   [-] research.1         | [=====]                  |        |         |         |     |        |
|       turn 1 (researcher)| [==]                     |        |         |         |     |        |
|       turn 2 (researcher)| [==]                     |        |         |         |     |        |
|   [-] bakeoff.1          |        [========]        |        |         |         |     |        |
|     [+] attempt.1        |        [======]          |        |         |         |     |        |
|     [+] attempt.2        |        [========]        |        |         |         |     |        |
|   [+] review.1           |                  [=====] |        |         |         |     |        |
+---------------------------------------------------------------------------------------+
```

- `RunViewer.svelte` keeps the observation, the SSE connection, and the
  revision counter, and renders a `UsageView` above the existing transcript
  sections.
- `UsageView.svelte` computes `rollup(observation)` in a `$derived` keyed
  on `revision`, computes the axis (start: root `began`; end: root `ended`,
  or while the run is running the latest timestamp the snapshot holds, so
  every frame extends it), and renders `UsageRow.svelte` for the root.
- `UsageRow.svelte` is recursive over three row kinds. Columns: a
  disclosure toggle and name (scope name with ordinal, the run name for the
  root; the turn's session name and turn number; the message's model and
  step ordinal); a bar positioned on the axis by the row's interval (scope
  began/ended, turn started/ended, message `time.created`/`time.completed`,
  open intervals to the axis end and marked open); five cells from
  `total(usage)`. Expanding a scope shows its child scopes (by `parent`)
  and the turns placed directly in it, in began/started order, and, when
  the scope has a task, the task's name and description. Expanding a turn
  shows its prompt (first line, the rest under a `<details>`) and its
  model-call rows. A message row has a link that scrolls to its
  `data-message-id` anchor in the transcript below. Under any expanded row,
  a `<details>` "by model" shows one line per model with the five cells.
- Bars encode when and how long; cells encode how much. No segmented bar,
  no duration column, no ticking clock (Tyler, interview).
- `MessageRow.svelte`'s footer prints the five fields from `messageUsage`
  with unavailable fields named, replacing the `accounting` string.
- Run `svelte-autofixer` on every component before calling it done, per
  `CLAUDE.md`.

## Implementation Plan

### Phase 0: Probe provider accounting live

Before touching normalization, capture what each harness actually reports.

Files: `ephemeral/attest/usage-by-scope/main.go` (the Phase 5 workflow, in
its first form), `ephemeral/attest/usage-by-scope/probe/accounting.md`.

1. Write the Phase 5 workflow and run it once with the page off, on the
   cheap tier: Claude `claude-haiku-4-5-20251001`, Codex `gpt-5.6-luna`,
   Antigravity `gemini-3.8-flash-low`. The `research` scope's two Haiku
   turns share a long identical preamble so the second hits a warm cache; a
   single `run-prompt` cannot do this because a cache read needs a second
   turn on one session.
2. From `sessions/*.jsonl`, extract `rawProviderAccounting` and the
   normalized `tokens` for every step and record them in `accounting.md`
   with the model used.
3. Decide from the numbers: does Antigravity's `input_tokens` include
   `cache_read_tokens` (expected yes: `total_tokens == input_tokens +
   output_tokens` on a cached turn and `cache_read_tokens <= input_tokens`);
   does Anthropic carry `output_tokens_details`; does Codex carry a cache
   write field. Phase 3 follows the answers. If a number contradicts the
   table above, the table changes here, not silently in code. If the cheap
   run shows no cache hit, say so and try one more time with a longer
   preamble; if it still does not, the Antigravity formula stays a
   hypothesis and Phase 3 leaves its `input` untouched with the reason
   recorded.

Phases 1 and 2 do not wait on this.

### Phase 1: Fold scopes and turns into the observation (Go)

Files: `internal/observation/snapshot.go`, `store.go`, `checkpoint.go`,
`store_test.go`; `event_persistence.go`; `run.go`; `gimble_test.go`.

1. Add `ScopeInfo`, `TurnInfo`, and the two maps on `RunSnapshot`.
   `Parent` is the key minus its last segment; the root's parent is `""`
   and the page tells it apart by `key === ''`.
2. `writeLifecycle` returns the record time with the bytes; `run.event`
   passes it to `observeLifecycle`; `Lifecycle` gains `Time` and the four
   optional pointers; `observeLifecycle` fills them in its existing switch,
   marshalling `Task` only when present.
3. `Store.Lifecycle`: on `ScopeBegan` insert `Scopes[key]` with name (the
   run name when key is `""`), parent, task, began; on `ScopeEnded` set
   ended and error; on `TurnStarted` insert `Turns[turn]` with scope,
   session, prompt, output type, started; on `TurnEnded` set ended,
   duration in ms, error, interrupted. An ended for an unknown key inserts
   a row with what it has.
4. `snapshotLocked` copies both maps, so `Close` persists them;
   `loadCheckpoint` defaults them.
5. `store_test.go`: feed began, started, ended, ended for a nested key
   (`lap.1/bakeoff.1/attempt.2`) and a root began; assert `Snapshot()` and,
   after `Close`, `loadCheckpoint` carry keys, parents, the root name, times,
   prompt, duration; assert a subscriber that connects after the fold
   starts from a snapshot that already has them.
6. `gimble_test.go`: after the fake attest run, read `observation.json` and
   assert its scope keys equal the set of scope keys in `run.jsonl`'s
   `scope_began` records, every `turn_started` turn is in `turns` with its
   prompt, and every ended is at or after its began.

### Phase 2: Mirror the fold and define usage (browser)

Files: `web/src/lib/observation/index.ts`, `usage.ts` (new),
`index.test.ts`, `usage.test.ts` (new),
`fixtures/antigravity-run-prompt.json` (new, the real checkpoint copied
from `ephemeral/attest/antigravity-run-prompt/...`), `web/package.json`
(add `usage.test.ts` to the `test` script, which is an explicit file list).

1. Types and fold as in § The snapshot shape, including cancellation
   precedence and empty defaults.
2. `turnRevision(turn)`: bumped in `apply` for event frames; reset by
   `replace`.
3. `usage.ts` as specified, with `rollup` and its per-turn cache, and
   `contains`.
4. Delete `accounting()` from `index.ts`.
5. Tests, with expected values written down independently of the code:
   the fold produces the same maps a Go snapshot would for a nested key;
   `add` is associative and commutative; a row with
   `fieldAvailability.cacheWrite: false` yields `cacheWrite.unknown === 1`
   and the other four summed; a completed row without an `accounting`
   sidecar contributes nothing and does not count as a call; an in-flight
   row contributes nothing; on the real Antigravity checkpoint,
   `rollup().scopes.get('')` has `calls === 2`, `cacheWrite.unknown === 2`,
   and per-field tokens equal to the sum of its two accounting rows; a
   synthetic three-level tree with a turn placed at each level reconciles
   parent = direct turns + children at every node; `attempt.1` does not
   contain `attempt.10`; two models in one scope stay separate; a
   replacement snapshot resets the rollup to exactly the new snapshot's
   value.

### Phase 3: Adapter accounting (gated on Phase 0)

Files: `agy/events.go`, `agy/events_test.go`, `claude/events.go`,
`claude/events_test.go`, `codex/events.go`, `codex/events_test.go`.

1. All three: delete `tokensAvailable`; keep `fieldAvailability` and
   `rawProviderAccounting`; a derived field whose subtraction goes negative
   is marked unavailable, not clamped.
2. agy, if Phase 0 confirms: `input = input_tokens - cache_read_tokens`,
   available iff both present; cache write stays unavailable. Test with
   the probe's raw numbers.
3. Claude, if Phase 0 confirms: `fieldAvailability.output = outputOK`;
   when the reasoning detail is absent, `output` is the total and
   `reasoning` is unavailable. Test with a usage block lacking
   `output_tokens_details`.
4. Codex, if Phase 0 confirms: `fieldAvailability.input = inputOK &&
   cachedOK`; subtract cache write only when present. Test with a usage
   block lacking `cache_write_input_tokens`.
5. `go test ./agy ./claude ./codex .`

### Phase 4: The page

Files: `web/src/lib/observation/RunViewer.svelte`, `UsageView.svelte`
(new), `UsageRow.svelte` (new), `MessageRow.svelte`.

1. `UsageView`: `$derived` rollup on `revision`; axis from root began to
   root ended or the latest known timestamp; per-model summary lines for
   the run in the header.
2. `UsageRow`: the three row kinds, interval bar as a positioned `<span>`
   in a fixed-width track, five cells with the four display states, task
   name and description on an expanded task scope, prompt on an expanded
   turn, "by model" `<details>`, message link to the transcript anchor.
   Expand state is a `Set<string>` on the view keyed by row id so
   re-renders keep it. Wide content scrolls inside its own container.
3. `RunViewer`: render `UsageView` above the transcript sections.
4. `MessageRow` footer from `messageUsage`.
5. `svelte-autofixer` on each component, then `cd web && pnpm run check &&
   pnpm test`, then `just build`.

### Phase 5: Proof

Files: `ephemeral/attest/usage-by-scope/main.go` (from Phase 0),
`ephemeral/attest/usage-by-scope/result.md`, `web/scripts/reconcile-usage.ts`.

The workflow, inline per `AGENTS.md`, with the page on:

```go
runtime := web.NewRuntime(ctx, logs, web.WithPort(8080))
err := runtime.Run(ctx, "usage-proof", func(ctx context.Context) error {
	// a direct turn in the root, beside the scopes below
	opener := gimble.NewSession(ctx, "opener", codex.New(), "gpt-5.6-luna", workdirA)
	if _, err := opener.Generate[gimble.Text](ctx, "In one line, what is a token budget?"); err != nil { return err }
	// research: one Haiku session, two turns, the second cache-warm
	if err := gimble.Scope(ctx, "research", func(ctx context.Context) error {
		s := gimble.NewSession(ctx, "researcher", claude.New(), "claude-haiku-4-5-20251001", workdirB)
		if _, err := s.Generate[gimble.Text](ctx, preamble+"Summarize this in one line."); err != nil { return err }
		_, err := s.Generate[gimble.Text](ctx, preamble+"Name its three most important identifiers.")
		return err
	}); err != nil { return err }
	// bakeoff: two concurrent Codex attempts, separate workdirs
	g := gimble.Group(ctx, "bakeoff")
	for i, dir := range []string{workdirC, workdirD} {
		g.Go("attempt", func(ctx context.Context) error {
			s := gimble.NewSession(ctx, "candidate", codex.New(), "gpt-5.6-luna", dir)
			_, err := s.Generate[gimble.Text](ctx, fmt.Sprintf("Write haiku %d about token budgets.", i+1))
			return err
		})
	}
	if err := g.Wait(); err != nil { return err }
	// review: one Antigravity turn nested two deep
	return gimble.Scope(ctx, "review", func(ctx context.Context) error {
		return gimble.Scope(ctx, "verdict", func(ctx context.Context) error {
			s := gimble.NewSession(ctx, "judge", agy.New(), "gemini-3.8-flash-low", workdirE)
			_, err := s.Generate[gimble.Text](ctx, "Pick the better of these two haiku: ...")
			return err
		})
	})
})
```

Exact names follow `go doc -all .` at implementation time; the shape is
what matters: a direct root turn, a warm-cache scope, a concurrent group,
a scope nested two deep, three harnesses.

1. Run it. While it runs, open `http://127.0.0.1:8080/runs/<id>` and
   capture the tree with the two `attempt` bars overlapping and totals
   rising (live). After it ends, reload and capture the finished view.
2. `bun web/scripts/reconcile-usage.ts <run dir>`: reads `observation.json`,
   `run.jsonl`, and `sessions/*.jsonl`; prints the rollup for every scope
   and turn; exits non-zero unless (a) every turn's per-field known tokens
   equal the sum of that turn's `turn_ended` tokens, (b) the root equals
   the sum over all turns, (c) per-model tokens equal the sum of
   `session.step.ended` tokens grouped by the `session.step.started`
   model, and (d) every scope equals its direct turns plus its children.
   Paste the output into `result.md`.
3. Copy only `runs/<id>/observation.json` into a fresh project directory,
   start the server against it, open the run: same tree, same numbers, no
   log read.
4. `result.md` records the models used and their resolved ids, the run id,
   the raw and normalized numbers for the cache-warm Haiku turn, and every
   field a harness left unavailable.

Checks, in order, before the validator looks: `just build` (Go web tests
embed the build), `go test -count=1 ./...`, `go vet ./...`, `cd web && pnpm
test`, `cd web && pnpm run check`.

## Files Summary

| File | Change |
| --- | --- |
| `internal/observation/snapshot.go` | `ScopeInfo`, `TurnInfo`, `Scopes`, `Turns` on `RunSnapshot` |
| `internal/observation/store.go` | `Lifecycle` time and four optional variants; fold; snapshot copies; root name |
| `internal/observation/checkpoint.go` | default the two maps on load |
| `internal/observation/store_test.go` | fold, checkpoint, late-subscriber test |
| `event_persistence.go` | `writeLifecycle` returns the record time |
| `run.go` | `observeLifecycle` fills scope and turn entries |
| `gimble_test.go` | checkpoint scope tree equals the log's prefix tree |
| `agy/events.go`, `agy/events_test.go` | `input` excludes cache read (per probe); `tokensAvailable` deleted |
| `claude/events.go`, `claude/events_test.go` | `output` independent of reasoning (per probe); `tokensAvailable` deleted |
| `codex/events.go`, `codex/events_test.go` | `input` independent of cache write (per probe); `tokensAvailable` deleted |
| `web/src/lib/observation/index.ts` | scopes, turns, fold cases, per-turn revision; `accounting()` deleted |
| `web/src/lib/observation/usage.ts` | new: the usage value, `messageUsage`, `add`, `rollup`, `contains` |
| `web/src/lib/observation/usage.test.ts` | new: monoid, availability, tool row, fixture, tree, prefix, replacement |
| `web/src/lib/observation/index.test.ts` | fold cases |
| `web/src/lib/observation/fixtures/antigravity-run-prompt.json` | new: real checkpoint fixture |
| `web/src/lib/observation/RunViewer.svelte` | renders `UsageView` above the transcript |
| `web/src/lib/observation/UsageView.svelte` | new |
| `web/src/lib/observation/UsageRow.svelte` | new |
| `web/src/lib/observation/MessageRow.svelte` | per-field footer |
| `web/package.json` | test script includes `usage.test.ts` |
| `web/scripts/reconcile-usage.ts` | new: proof script |
| `ephemeral/attest/usage-by-scope/` | probe results, workflow, result, screenshots |

Not changed: `internal/sessionstate/`, `web/src/lib/sessionstate/`,
`events.go` (the union), `jsonschema/`, `web/src/hooks.go`, `hooks.ts`,
`page.server.go`, `internal/observation/identity.go`, `http.go`,
`subscribe.go`, `registry.go`.

## Definition of Done

Gated on the intent's success criteria per `docs/definition-of-done.md`.
Each is a thing seen working, not a test count.

1. **Totals at every level.** On the proof run's page, live and after, the
   root row and every scope row show five cells and a by-model table.
   Fields a harness did not report show as unavailable or with an unknown
   count, never as a bare zero. Seen in the screenshots and printed by
   `reconcile-usage.ts` with the same numbers.
2. **Drill and reconcile.** Expanding `bakeoff.1` shows two `attempt` rows;
   expanding one shows its turn with the prompt; expanding the turn shows
   its model-call rows; the root shows the `opener` turn beside the scopes.
   The parent equals its direct turns plus its children at every step,
   checked by the three-level unit test and by `reconcile-usage.ts` (d) on
   the proof run.
3. **Timeline.** The two `attempt` bars overlap on the axis; `research.1`
   and `review.1` do not; the root bar spans the run. Seen in the finished
   screenshot; open bars marked open in the live one.
4. **Checkpoint alone.** A server started on a directory holding only the
   run's `observation.json` renders the same tree and numbers. Phase 5
   step 3.
5. **Comparable splits.** The Haiku session's second turn shows cache read
   greater than zero; the Codex turns show input and cache read; the
   Antigravity turn shows `input` equal to its raw `input_tokens` minus
   `cache_read_tokens` if Phase 0 confirmed the formula, and cache write
   unavailable. `result.md` holds the raw numbers beside the normalized
   ones; `reconcile-usage.ts` (c) checks per-model placement. If Phase 0
   could not produce a cache hit, this item records that and the formula
   stays a hypothesis; it is not a reason to hold the sprint.
6. **Tests.** `go test ./...` and `cd web && pnpm test` pass, including the
   fold, rollup, availability, tool-row, fixture, and replacement tests.
7. **Build.** `just build` succeeds; `pnpm run check` reports zero errors.

Exit at 90 to 95 percent: a presentation quirk (column widths, bars on very
short intervals) is filed as an issue and the branch is merged.

## Risks & Mitigations

- **Provider semantics assumed wrong.** The Gemini "input includes cache"
  reading rests on one run with zero cache reads. Phase 0 runs a warm-cache
  turn on each harness before any normalization change; each adapter edit
  is conditional on what it shows; the DoD names the evidence, not the
  formula.
- **Claude and Codex unavailable by construction today.** If Phase 0
  confirms it, every live message from both is currently unavailable and
  criterion 5 depends on Phase 3 for all three adapters. Phase 3 covers all
  three for this reason.
- **The cheap run shows no cache hit.** Retry once with a longer preamble;
  otherwise record it and leave the Antigravity `input` untouched.
- **Rollup cost per frame.** The per-turn revision cache limits
  recomputation to the turn that changed. Measured on the proof run.
- **Reconnect double counting.** The rollup is derived from current rows
  and dropped whole on a replacement snapshot; the replacement test proves
  it.
- **Tool-only rows counted as calls.** Rows without an `accounting` sidecar
  are not model calls; the real Antigravity fixture proves `calls === 2`.
- **Root named ".", `attempt.1` containing `attempt.10`.** Both are named
  in the fold and the `contains` test.
- **Timestamps from two clocks.** Scope times come from
  `LifecycleRecord.Time`, message times from `AgentEvent.Created`; both are
  `time.Now()` in one process. Message spans are observed event intervals
  (the Antigravity fixture shows about 1 ms), drawn as a visible tick, not
  a claim about inference time.
- **A checkpoint predating this sprint.** Empty tree, transcript still
  renders. No shim.
- **Task JSON.** `polytype.Optional[Task]` marshals as `""` when absent;
  the store takes a `json.RawMessage` filled only when present.
- **Scope of the page work.** Rows and cells land first, the bar second;
  totals without a bar already satisfy criteria 1, 2, 4, 5.

## Dependencies

- The three harness CLIs installed and logged in (`claude`, `codex`, `agy`
  1.2.1 or later) and the cheap tier: `claude-haiku-4-5-20251001`,
  `gpt-5.6-luna`, `gemini-3.8-flash-low`. Say which was used.
- `just build` before `go test ./...` in a fresh worktree, so the embedded
  web build exists.
- `web/` tests run with `bun` via `pnpm test`; the `svelte-autofixer` MCP
  tool for the components.
- The sealed `LifecycleEvent` union and `jsonschema/` are unchanged, so
  `go generate ./...` output does not move.
- No chapter, semantic index, price feed, or new production dependency.

## Open Questions

1. **Antigravity cache write.** If Phase 0 shows `agy` reporting a cache
   write field on some path, wire it; otherwise it stays unavailable and
   the page says so.
2. **Session card under its creating scope.** A session created in `loop.1`
   whose turns run in `loop.1/task.3` shows those turns under `task.3`, as
   `API.md` says. A second grouping by session is not in this sprint unless
   asked.
3. **Cost.** Bounded out. If a price table is wanted later, its home is a
   map from `${providerID}/${modelID}` to per-field prices in the browser,
   multiplying the same `Usage` value; nothing here blocks it.
4. **Extras.** Segmented bars, a duration column, tokens per second, and a
   ticking live clock were declined for this sprint and are cheap to add
   later on the same rows.
