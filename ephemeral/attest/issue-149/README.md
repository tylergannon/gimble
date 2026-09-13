# Issue 149 token accounting attestation

Proved code commit: `96cecfc` (branch `claude/issue-148-plan-ae018a`).

Models, as the run log records them natively:

| scope | harness | model asked for | native model in `turn_ended` |
| --- | --- | --- | --- |
| `claude.1` | Claude Code | `haiku` | `claude-haiku-4-5-20251001` |
| `codex.1` | Codex | `gpt-5.6-luna` | `gpt-5.6-luna` |
| `agy.1` | Antigravity | `gemini-3.8-flash-low` | `gemini-3.8-flash-low` |

## What was run

`go run ./ephemeral/attest/issue-149 -port 8099 -hold 30m` (`main.go` here).
It clears `logs/`, starts `web.NewRuntime` with the web application on
127.0.0.1:8099, and inside one `runtime.Run` opens three child scopes named
`claude`, `codex` and `agy` concurrently through an errgroup. Each scope holds
one `gimble.NewSession` on its adapter and takes two `Generate[gimble.Text]`
turns; the second turn resumes the same conversation, so the resumed-turn
accounting path is exercised on all three harnesses. Both prompts are
two-sentence questions about latency and throughput that need no tools. After
the run the process holds the server open so the page and the live query can be
read.

Run: `20260912-204128.issue-149`, port 8099, page
`http://127.0.0.1:8099/runs/20260912-204128.issue-149`. The run completed with
no error in 9.8 s; all six turns ended `ok`.

## What was seen

### Per session, final `session.usage.updated` (the session's running total)

| session | input | output | reasoning | cache read | cache write | cost |
| --- | --- | --- | --- | --- | --- | --- |
| `claude.1/claude.1` (haiku) | 20 | 138 | 278 | 66163 | 8100 | $0.0249163 |
| `codex.1/codex.1` (gpt-5.6-luna) | 15192 | 80 | 47 | 30208 | 0 | $0 |
| `agy.1/agy.1` (gemini-3.8-flash-low) | 30186 | 105 | 399 | 0 | 0 | $0 |

Claude reports a real cost and heavy cache read/write with almost no fresh
input; Codex reports fresh input plus cache read and no cost; Antigravity
reports fresh input only, no cache and no cost. That is what each harness
states, not a guess by Gimble.

### The `query.live` remote function

`scopeUsage` is published as `1cqgzir/scopeUsage` in `web/skgo.remotes.json`.
It was streamed with curl at `/_app/remote/1cqgzir/scopeUsage?payload=…`, the
payload being devalue's flat form base64url-encoded, exactly as in
`web/observation_live_test.go`. Frames saved in `live-root.txt`,
`live-claude.1.txt`, `live-codex.1.txt`, `live-agy.1.txt`. Decoded:

| scope | input | output | reasoning | cache read | cache write | cost |
| --- | --- | --- | --- | --- | --- | --- |
| `""` (root) | 45398 | 323 | 724 | 96371 | 8100 | $0.0249163 |
| `claude.1` | 20 | 138 | 278 | 66163 | 8100 | $0.0249163 |
| `codex.1` | 15192 | 80 | 47 | 30208 | 0 | $0 |
| `agy.1` | 30186 | 105 | 399 | 0 | 0 | $0 |

The root scope is exactly the sum of the three children in all five token
fields and in cost: 20+15192+30186=45398, 138+80+105=323, 278+47+399=724,
66163+30208+0=96371, 8100+0+0=8100, 0.0249163+0+0=0.0249163. Because the run
had already finished, each stream delivered its opening value and ended at
once.

### Arithmetic against the durable logs

Excerpts in `usage-log-excerpts.txt` (every `session.usage.updated` and every
`session.step.ended`) and `turn-ended-records.txt` (all six `turn_ended`
lifecycle records with their `usage` arrays).

- Each session's final `usage.updated` tokens equal the sum of its two
  `step.ended` token sets, field by field. Checked for all three sessions; all
  three match.
- Claude published `usage.updated` four times (after each `step.ended` and
  again after each Claude Code `result`); Codex and Antigravity twice each.
- The Claude session's cost is the sum of its two turn reports:
  0.0195637 (turn 1) + 0.0053526 (turn 2) = 0.0249163, which is the session
  total and the root-scope cost. The `step.ended` events carry cost 0 for every
  harness; the money comes from the harness's own turn report.
- `turn_ended` carries the five fields plus cost per model, not opaque per-step
  JSON.
- `grep -c accounting` is 0 in all three session JSONLs, in `run.jsonl` and in
  `observation.json`.

### The run page

The Chrome extension was not connected in this session, so no screenshots were
taken; the page was fetched with curl instead, mid-run (`page-midrun.html` /
`.txt`, taken about 5 s in, run status `running`, turn 1 shown complete with
tokens and turn 2 still streaming) and after (`page-after.html` / `.txt`, run
status `completed`).

What the server-rendered page shows:

- Per-message tokens for every assistant message on all three harnesses, in the
  five fields: e.g. Claude turn 1 `10 in · 72 out · 116 reasoning · 29157 cache
  read · 7849 cache write · $0`, Codex turn 1 `10808 in · 35 out · 0 reasoning
  · 9984 cache read · 0 cache write · $0`, Antigravity turn 1 `14863 in · 51
  out · 185 reasoning · 0 cache read · 0 cache write · $0`.
- No "unavailable" or "partial" wording anywhere on the page.
- The "Run usage" line is present. In server-rendered HTML it reads
  `Run usage · connecting`, because its value arrives over the client-side
  `scopeUsage` stream; the value that line renders is the root-scope frame
  above, $0.0249163 with 45398/323/724/96371/8100.

## What looked wrong

1. **The per-session running total on the page is always zero.** Every session
   header in `page-midrun.txt` and `page-after.txt` reads
   `0 in · 0 out · 0 reasoning · 0 cache read · 0 cache write · $0`, for all
   three sessions, before and after the run. The cause is in the reducers, not
   in the data: `SessionTimeline.svelte` renders `usageOf(state.info[sessionID])`,
   and both reducers fold the running total only into an info record that
   already exists — `web/src/lib/sessionstate/index.ts:48` is
   `case 'session.usage.updated': if (this.state.info[sid]) Object.assign(...)`,
   and `internal/sessionstate/reduce.go:84` is the same shape. No Gimble event
   in this run creates `info[sid]` (the session event types emitted are inbox,
   execution, step, reasoning, text and usage.updated — no session-created or
   `session.updated`), so `info` stays `{}`. `observation.json` confirms it:
   every invocation snapshot has `"info": {}` while the run-level `usage` map
   holds the correct per-session totals. This is a real defect against
   definition-of-done item 4 ("the run page shows message tokens and session
   totals"). It was not fixed here.
2. **Dollars never appear on a message or on a session line.** Claude's cost is
   stated by the harness in its turn report, so `step.ended` and therefore every
   rendered message shows `$0`. On the page the Claude dollars are visible only
   through the "Run usage" line fed by `scopeUsage` (and would be visible in the
   session-total line if defect 1 were fixed). The dollars themselves are
   correct and nonzero: $0.0249163.
3. Not a code defect, noted for the record: the server-rendered HTML cannot show
   the live "Run usage" value, and no browser was available here, so that line's
   rendered text was not seen — only the frame the browser would render, taken
   from the same endpoint with the same payload encoding.

## Files

- `main.go` — the attest program.
- `README.md` — this file.
- `logs/runs/20260912-204128.issue-149/` — the durable run: `run.jsonl`,
  `observation.json`, `sessions/{claude.1,codex.1,agy.1}/*.jsonl`.
- `page-midrun.html`, `page-midrun.txt` — the run page during the run.
- `page-after.html`, `page-after.txt` — the run page after the run.
- `live-root.txt`, `live-claude.1.txt`, `live-codex.1.txt`, `live-agy.1.txt` —
  the `scopeUsage` SSE response (headers and frame) per scope.
- `usage-log-excerpts.txt` — every `session.usage.updated` and `step.ended`,
  the arithmetic, and the `grep -c accounting` counts.
- `turn-ended-records.txt` — all six `turn_ended` lifecycle records.
- `workspace/` — the (empty) agent working directory the sessions were given.
