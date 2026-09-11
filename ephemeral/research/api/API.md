# Gimble API: decisions so far

The rule: **as simple as possible.** Add a name only when a workflow that
exists needs it. Everything else is ordinary Go written in the workflow.

The research behind these decisions is in `gemini/`, `claude/`, and
`chatgpt/` next to this file. The old code in `ephemeral/legacy/` is
inspiration, not the API.

## Sessions and turns

```go
// The adapter holds the agent-specific code (Codex, Claude, agy). It is
// untyped: a raw JSON Schema goes in, raw JSON comes out.
type HarnessAdapter interface { /* ... */ }

// Session is a concrete struct wrapping one adapter session.
type Session struct { /* adapter, session id, model, workdir */ }

// Generate runs one turn and blocks until it ends. T is a polytype-generated
// type: its schema is sent with the prompt, and the result is validated and
// decoded into T.
func (s *Session) Generate[T Output](ctx context.Context, prompt string) (T, error)

// Steer injects a message into the turn that is running on this session.
// It is called from another goroutine while Generate blocks. The message
// lands at the worker's next model call.
func (s *Session) Steer(ctx context.Context, message string) error

// Interrupt stops the turn that is running on this session. Generate
// returns with an error. Cancelling Generate's context does the same thing;
// the method is for callers that hold the session but not the context.
func (s *Session) Interrupt(ctx context.Context) error

// Fork returns a new session, named for the graph, with the same
// conversation so far. The two sessions are independent after that.
func (s *Session) Fork(ctx context.Context, name string) (*Session, error)
```

- Session is a concrete struct, not an interface, so it can have a generic
  method (Go 1.27 allows generic methods on concrete types, never on
  interfaces). Unit tests fake the `HarnessAdapter` underneath; proof comes
  from live runs.
- Sessions are created with `gimble.NewSession(ctx, name, ...)`. The run
  is in the ctx, and so is the scope the session belongs to (see Scope).
- A follow-up is another `Generate` on the same session. There is no
  `FollowUp`.
- `Steer` is a primitive, not an option. `Supervise` is built on it, and a
  workflow can call it directly (a human at the console, a watchdog on a
  timer). The legacy adapters implement it: Codex through native
  `turn/steer` on the active turn, Claude by sending on the live SDK
  session, and agy, which has no native steer, by interrupting the process
  and replaying the message on the resumed turn.
- `Fork` exists because loading a session with research costs tokens and
  minutes, and a fork pays that once. Prime one session, then fork it for
  the coder and the validator so both start already knowing the codebase;
  or research, fork, attempt, and on failure attempt again from the fork
  instead of from scratch. Forks share one harness, so priming that has to
  cross models is a doc both sessions read, not a fork. Two forks in the
  same workdir cannot run turns at once.

## Context

Every operation takes a `context.Context` and stops when it is cancelled.
No cancel channels, no done markers. A cancelled `Generate` interrupts the
native turn and returns `ctx.Err()` (the legacy adapters do this today, and
it was live-attested: a SIGINT stops a Codex turn within seconds).

The rule for which context to derive from: **work derives from the
lifetime that owns it, not from whoever triggered it.**

- The process has a root context. SIGINT cancels it, and every run's
  context derives from it.
- The HTTP server lives as long as the process and sets
  `http.Server.BaseContext` to return the root, so every request context is
  a child of root and dies with it.
- A steer from a handler is a short call; `r.Context()` is fine.
- A turn started from a handler derives from the run's context, not the
  request's. A dropped browser tab must not kill twenty minutes of work.
- An observer streaming events derives from the request. Losing the client
  ends the stream, not the turn.
- "Cancel when either of two contexts dies" is `context.AfterFunc`, not a
  merged-context type:

  ```go
  ctx, cancel := context.WithCancel(root)
  defer context.AfterFunc(other, cancel)()
  ```
- No Futures. No ACP; Gimble has its own agent drivers.
- Names (`Generate`, `NewSession`, `Output`) are placeholders.

## Scope

A scope is a named span of the workflow: a segment, a lap, one candidate's
attempt. It is the unit of structure the page draws and the unit of data
the workflow hands to agents. The ctx says which scope you are in; the run
owns what is in it.

```go
err := gimble.Scope(ctx, "sprint", func(ctx context.Context) error {
	gimble.Set(ctx, "language", "go")                    // scalars
	gimble.SetJSON(ctx, "research", research)            // polytype-generated types
	lang, ok := gimble.Get[string](ctx, "language")      // nearest scope up the chain
	r, ok := gimble.GetJSON[Research](ctx, "research")   // validated, then decoded
	prompt := task.Text + "\n\n" + gimble.ScopeText(ctx) // the chain, rendered for a prompt
	// ...
	return nil // the scope ends here: its sessions are closed, then its ctx is cancelled
})

g := gimble.Group(ctx, "bakeoff") // a concurrent scope
g.Go("attempt", func(ctx context.Context) error { /* ... */ return nil })
err = g.Wait() // joins the goroutines, then ends the group's scope
```

Identity is in the ctx; data is in the run.

- A scope is a function. It begins when `Scope` calls the body and ends
  when the body returns. There is no handle to close, because a ctx cannot
  know when it goes out of scope and a function boundary can. Ending
  cancels the ctx, and nothing created in the scope outlives it: goroutines
  were joined by their group before the return, sessions are closed, the
  ctx is dead. That is the context rule made mechanical, and it is the rule
  structured concurrency settled on (Trio's nurseries, Kotlin's coroutine
  scopes, Swift's task groups). Cancelling alone bounds nothing in Go,
  since a cancelled ctx is only a signal; waiting is what bounds a
  lifetime, and `Wait` is where it happens.
- `Group` is the concurrent scope, and it is explicit, the way `errgroup`
  is. `g := gimble.Group(ctx, name)` opens it; `g.Go(name, fn)` starts a
  goroutine in a child scope of its own; `g.Wait()` joins them all and ends
  the group. The first error a goroutine returns cancels the group's ctx,
  interrupting the others' turns, and comes back from `Wait`; a goroutine
  whose failure must not stop its siblings returns nil and reports on a
  channel, as the bake-off below does. `g.Go` is the only way a workflow
  starts a goroutine; raw `go`, `sync.WaitGroup`, and `errgroup` are for
  Gimble's own code. Explicit rather than a `Go` that finds its group
  through the ctx, because then a helper cannot quietly attach goroutines
  to a caller's scope and the reader sees the join; Trio passes nurseries
  explicitly for the same reason. That every group is waited on every path
  before its function returns is a lint of the `lostcancel` shape.
- Sessions belong to the scope that created them and are closed when it
  ends. A session is a process and a workdir lease, the resource that
  leaks worst, and without this rule nothing ends one before the run
  does. With it, every agent process has a definite end the page can show,
  a workdir is known to be free at scope end, and a session that escapes
  its scope errors on its next `Generate` instead of running as a zombie.
  `Fork` loses nothing: create the researcher in the parent, prime it in a
  child (the priming turn is bounded by the child, the session by the
  parent), fork it later. A session created in a lap and wanted in the
  next lap is a smell; cross-lap state is made in the loop's scope, as the
  coder is in the Loop example. There is no `Close` on `Session`; Gimble
  calls it. Turns are bounded already: `Generate` blocks, and a
  goroutine's turn is bounded by its group's `Wait`.
- What the rule forbids, and what to do instead. None of it loses
  expressiveness; the fix is always to make the thing in the scope that
  should own it.
  - Fire and forget past your own scope. A watchdog for a turn lives in
    the turn's scope; the server lives in the run.
  - Returning while children run. `Wait` comes before the return, on every
    path, and blocks until they stop. A child that ignores ctx blocks
    forever, the same as a `Wait` does today; the harnesses abide ctx, so a
    turn stops in seconds.
  - A producer in one scope feeding a consumer in a sibling scope. They
    share a parent, and the join is explicit.
  - A goroutine or a session that should outlive a lap. Make it in the
    loop's scope, outside the range.
  - The cost that actually bites: a goroutine you forgot keeps its group's
    `Wait` blocked until it finishes instead of leaking. The page shows the
    group still open with a running child, which is the right failure.
- Nothing is stored in the ctx but a pointer to the scope. The values live
  in a store the run owns, keyed by scope, and every `Set` is a log event,
  so the page shows scope contents with no new endpoint. This is the line
  OpenTelemetry draws: the span is in the ctx and immutable there, and its
  attributes are written freely. The run itself is
  in the root ctx, which is why `NewSession`, `Loop`, and `Scope` are
  package functions and no run handle is threaded through anything.
- `Get` reads the nearest scope up the chain that has the key, like a
  variable: an inner scope shadows an outer one. `ScopeText` renders the
  chain the same way, innermost wins, and the workflow puts it in the
  prompt itself. `Generate` injects nothing.
- Two `Set` functions, because a Go constraint cannot mix a type set with
  an interface that has methods. `Set` takes scalars and derives their
  schema. `SetJSON` takes polytype-generated types, the same `Output` that
  `Generate` decodes into, so a value carries its schema and the field
  descriptions an agent reads. An arbitrary struct or a map does not
  compile, which is the point.

  ```go
  // Output is what polytype generates, and what Generate, SetJSON, and
  // GetJSON require.
  type Output interface {
  	Schema() json.RawMessage
  	ValidateJSON([]byte) error
  }
  ```

- Set-once per key per scope instance. The second `Set` of a key errors,
  naming the scope. That turns two goroutines racing on a key into a bug
  report, and it means a value cannot change under a turn that is reading
  it. Revision is shadowing: enter a child scope and set the key there.
  In-place replacement waits for a workflow that needs it.
- `Set` marshals at the call, so a value is a snapshot. A pointer mutated
  afterwards changes nothing, and agents, the page, and the files all see
  the same bytes.
- Keys and scope names are compile-time constants, like node names.
- `Set` on an ended scope, or on a ctx with no scope, is an error. The
  first can only happen through a ctx stored somewhere it should not be;
  the second is a chain severed by `context.Background()`. Both should be
  loud. `Get` on an ended scope is fine; the values are the record.

**Iterators own scopes.** A plain `for` body cannot end a scope per
iteration, so a `Set` inside a plain loop hits set-once on the second
pass. Gimble's iterators make the range body the scope's function: a fresh
scope every iteration, ended when the body returns. `Loop` does this per
lap, and there is a trivial one for a fixed slice:

```go
for ctx, prompt := range gimble.Each(ctx, "attempt", prompts) {
	// ctx is a child scope named "attempt", ended when this body returns
}
```

A group made inside a lap is waited inside the lap. One that should
outlive laps is made in the loop's scope, outside the range.

**Lints.** The rules above are checkable statically, in one analyzer over
the ssa form, and the runtime errors stay as the backstop for what it
cannot see:

- Two `Set` calls with the same constant key under the same scope, with no
  scope between them.
- A `Set` inside a plain `for` body on a ctx from outside the body. Use
  `Each`, or a `Scope` per iteration.
- A non-constant key or name.
- A raw `go` statement, `sync.WaitGroup`, or `errgroup` in a workflow
  package. Goroutines are `Group.Go`.
- A `Group` not waited on every path before its function returns, the
  `lostcancel` shape. This is what makes the bound hold.
- A ctx stored in a struct or a global, or one from `context.Background()`
  on a path that creates nodes or sets values, which is what the existing
  `contextcheck` linter flags. With a scope's ctx never escaping its
  function, "set after end" cannot happen.

**Files.** Later, and without changing the API: the store is written
through to the run directory. Go keeps reading from memory, so there is no
filesystem read race on the Go side; the files are for agents, the page,
and restart. Layout is the scope chain, as in
`run/scopes/sprint/attempt.3/research.json`, with instances of a repeated
scope given an ordinal by the runtime. The runtime is the only writer;
agents and the page read. Writes are atomic, temp file then rename, and a
scope's directory is frozen when the scope ends. The directory tree is the scope
tree is the graph. Once files exist, `ScopeText` can reference a large
value by path instead of inlining it.

## Run

A run is the root scope, and there is a runtime object behind it. The
workflow never holds it.

```go
func Run(ctx context.Context, name string, body func(ctx context.Context) error) error
```

The process is a server first. It serves the project's runs directory for
its whole life, live runs and past ones alike, and it hosts the workflows
compiled into it. `Run` creates a run inside it: a directory under the
project's runs, the event log, the session registry, the scope store. It
registers the run with the server as live, opens the root scope, and calls
the body. When the body returns, the root scope ends like any other,
sessions closed and ctx cancelled, the log gets its final event, and from
then on the run is served the way every past run is: from its log.

The wiring between workflow state and the server is the ctx, in both
directions:

- Every package function (`Scope`, `Group`, `NewSession`, `Loop`, `Set`)
  reads the current scope from its ctx, and the scope knows its run. The
  run is not a second parameter because the scope is already there and the
  run is its root; passing both would be the same pointer twice.
- Every remote function (`watchRun`, `steer`, and the rest) gets a request
  ctx that is a child of the process root, which carries the project: the
  runs directory and the registry of live runs. A watch tails a log from
  disk, live or past, and stops at the run's final event. `steer`,
  `interrupt`, and a scope cancel need a live run and error on a past
  one.
- Inside the run, the log has one writer and any number of readers, and
  the registries are maps under a mutex. Nothing is global, so two runs in
  one process do not see each other.
- Tests call `Run` with fake adapters and the server off.

## Concurrency: Scope and Group

Gimble has no built-in tactics for running agents in parallel. There is no
`BakeOff`, no `Race`, no `Parallel`, no `FirstK`. The legacy tree has
bake-off helpers and sketches; they are the pattern this section replaces.
What Gimble has is `Group`, the only way a workflow starts goroutines, with
`Wait` as the join. A workflow that needs a bake-off writes it out, like
this, so the reader sees every decision: how many candidates, what counts
as a finisher, when the losers stop.

**Wait for all of them** (the first error cancels the rest):

```go
var toaster Toaster
var car Car
g := gimble.Group(ctx, "build")
g.Go("toaster", func(ctx context.Context) (err error) { toaster, err = s1.Generate[Toaster](ctx, "make me a toaster"); return })
g.Go("car", func(ctx context.Context) (err error) { car, err = s2.Generate[Car](ctx, "make me a car"); return })
if err := g.Wait(); err != nil { // the first error cancelled the other turn
	return err
}
```

**Take the first k finishers** (`k = 1` is "first one"):

```go
type outcome struct {
	res Result
	err error
}
ctx, cancel := context.WithCancel(ctx) // to cut the losers before Wait
defer cancel()
g := gimble.Group(ctx, "bakeoff")
results := make(chan outcome, len(prompts)) // buffered: senders never block

for _, prompt := range prompts {
	g.Go("attempt", func(ctx context.Context) error {
		// One session per turn in flight. They share the name: the name is the
		// node, and the page stacks the instances under it.
		s, err := gimble.NewSession(ctx, "candidate", adapter, model, workdir)
		if err != nil {
			results <- outcome{err: err}
			return nil
		}
		res, err := s.Generate[Result](ctx, prompt)
		results <- outcome{res, err} // a failure travels on the channel,
		return nil                   // because returning it would cancel every other candidate
	})
}

var winners []Result
for range prompts {
	o := <-results
	if o.err != nil {
		continue // a failed candidate is not a finisher
	}
	winners = append(winners, o.res)
	if len(winners) == k {
		break
	}
}
cancel()                         // interrupt the losers' turns
if err := g.Wait(); err != nil { // let them finish shutting down
	return err
}
```

Why each line is there:

- The channel holds every result, so a losing goroutine never blocks on send.
- `cancel()` interrupts the losers' native turns; `Wait()` keeps the
  function from returning while their agent processes are still exiting.
- The goroutines return nil on a failed candidate. An error returned to the
  group cancels the siblings, and one failed candidate must not do that,
  so the failure travels on the channel instead. In the wait-for-all case
  the error is returned, because there a failure is the group's failure.
- The loop reads at most `len(prompts)` results, so it cannot hang when too
  many candidates fail. `len(winners) < k` means it came up short.
- A session runs one turn at a time, so concurrent turns need their own
  sessions.
- Each attempt is a scope of its own, so what it sets never collides with
  a sibling's, and the page shows N attempts under the bake-off.

When the candidates return different types, use one channel per type and a
`select`, written in the workflow that needs it.

## Loop: the first built-in

Gimble provides two orchestration shapes, `Loop` and `Supervise`. `Loop`
iterates a task file, where a planner agent decides each lap what comes
next, and yields each lap with a context of its own.

```go
loop := gimble.Loop(ctx, "sprint", goal, planner)
for ctx, task := range loop.Laps {
	implement(ctx, coder, task)
}
if err := loop.Err(); err != nil {
	return err
}
```

The file is markdown with YAML frontmatter: the goal as a prose Definition
of Done, and a list of steps. It lives in the run directory at the loop's
scope path, so the durable state and the graph are one tree; kill the
process, rerun, and the planner reads where things stand. The goal is a Go
string, and the loop writes it into the file when it starts.

Each lap the iterator:

1. Reloads the file.
2. Runs every step's `command:` and collects the exit codes. This is the one
   mechanical part: commands are cheap and deterministic, and it means the
   planner judges from facts. It cannot claim the tests pass unless `go test`
   exited 0.
3. Asks the planner, with the file, the workspace, and the command results:
   what is next? The planner may rewrite the backlog however it sees fit.
4. Yields the task the planner named, with a context for the lap: a child
   scope of the loop's, cancelled when the body returns. Or returns, if
   the planner named none.

Why a planner and not a mechanical walk: an agentic backlog goes stale in
proportion to the steps already taken. The agent may finish early, a step may
need repeating, the next few steps may need rewriting, or the sequence may be
"think up the next thing to do." A mechanical iterator (first open item, mark
done, sweep at the end) has to special-case each of those; a planner that
rewrites the backlog every lap handles all of them by construction. The
legacy loop (`ephemeral/legacy/program/loop.go`) was the mechanical version,
with an evaluator bolted on that reopened item 0 as a stand-in when the goal
failed. That hidden decision is what made it feel like magic.

What this means for the reader: `Loop` does one sentence of work. Each lap
the planner reads the file and the workspace, updates the file, and says
what is next or that we are done. The goal is in the workflow, the planner
is a session the workflow chose and primed, and the file is on disk, so
the decision is delegated in plain sight, not hidden in code.

Details:

- When the planner names nothing, the iterator returns and `range` falls
  through. That is all `iter.Seq2` needs: the function exits. How the
  planner's JSON says "nothing" is the planner type's business and never
  reaches the body. The yielded type is just `Task`; there is no `Done`
  field, because the body never sees a lap where nothing is next.
- The lap is a scope. When the body returns, the lap's sessions are closed
  and its ctx is cancelled; a group the body made was waited before that,
  by the lint. A lap can be interrupted from
  the page without stopping the loop: the body's `Generate` returns the
  context error, and the planner sees an interrupted lap when it plans the
  next.
- `Err` is checked after the range, the way `bufio.Scanner` does it,
  because the yielded pair is the context and the task. `Loop` fails only
  for real reasons: its file, a command that will not start, the planner's
  harness dying. A command that exits non-zero is a fact for the planner,
  not an error.
- The iterator never marks anything. There is no engine-owned `done` field,
  no `MarkDone`/`UnmarkDone`, no YAML rewriting. Agents and humans edit the
  file; the iterator only reads it and runs its commands.
- Lap limits and giving up belong to the body (`if task.Lap > maxLaps`), not
  to `Loop`.
- Cost is one planner call per lap plus the commands. The legacy loop
  re-judged every inference item every lap, which was O(n²) judge calls.

## Supervise: the second built-in

`Supervise` attaches reviewers to a turn. A reviewer is a session of its own
and an instruction saying what to watch for.

```go
func (s *Session) Generate[T Output](ctx context.Context, prompt string, opts ...Option) (T, error)

func Supervise(reviewer *Session, instruction string) Option
```

```go
coder := gimble.NewSession(ctx, "coder", adapter, model, workdir)
scope := gimble.NewSession(ctx, "scope", adapter, cheap, workdir)
taste := gimble.NewSession(ctx, "taste", adapter, cheap, workdir)

res, err := coder.Generate[Result](ctx, task.Text,
	Supervise(scope, "don't let it implement anything the goal doesn't require"),
	Supervise(taste, "don't let it over-engineer"),
)
```

Reviewers watch the turn while it runs and steer the worker the moment they
object. Steering is the point: a coding turn runs twenty minutes, and a
reviewer that first speaks at the end watches the worker build the plugin
system and then pays a second turn to tear it out.

What `Generate` does when reviewers are attached. This is the whole
definition; if `Supervise` ever needs more than this, inline it instead.

```go
for round := 1; ; round++ {
	transcript := newTranscript() // the worker's events, appended as they arrive

	// The worker's turn, in the background so the reviewers can run beside it.
	var res T
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		res, err = s.generate[T](ctx, prompt, transcript.append)
	}()

	// Each reviewer, self-paced: look whenever there is something new, steer
	// on objection, stop looking when the turn ends.
	var wg sync.WaitGroup
	for _, r := range reviewers {
		wg.Go(func() {
			for transcript.waitForNew(done) {
				review, _ := r.session.Generate[Review](ctx, r.instruction, prompt, transcript.soFar())
				if len(review.Objections) > 0 {
					s.Steer(ctx, strings.Join(review.Objections, "\n"))
				}
			}
		})
	}
	<-done
	wg.Wait()
	if err != nil {
		return res, err
	}

	// A final look at the finished answer.
	var objections []string
	for _, r := range reviewers {
		review, _ := r.session.Generate[Review](ctx, r.instruction, prompt, transcript.soFar(), res)
		objections = append(objections, review.Objections...)
	}
	if len(objections) == 0 {
		return res, nil // every reviewer approved
	}
	if round == maxRounds {
		return res, fmt.Errorf("supervise: %d rounds without approval", round)
	}
	prompt = followUp(objections) // same session: the worker remembers what it did
}
```

- Reviewers are attached per turn, not per session. The instruction is
  about the task, and the reader sees at the call site who is watching. A
  session carries no hidden reviewers. What is per session is the reviewer
  itself: its model, its workdir, and its memory of earlier objections. A
  workflow that wants the same watchers on every turn passes the same
  options every time.
- A review is `struct{ Objections []string }`. An empty list is approval;
  there is no `Approved` flag, for the same reason `Task` has no `Done`.
- Objections raised during the turn are steered. Objections raised at the
  final look become a follow-up turn on the worker's session, still typed
  `T` and still supervised. Same review type either way.
- Cadence is the reviewer's own speed: it looks again as soon as it is done
  looking and something new has happened. No knob. "Something new" is a
  tool result, not a token; that is what a review can judge.
- A steer lands at the worker's next model call, same as a person talking
  over its shoulder. The reviewer judged a snapshot that is now a few
  seconds old; that is fine.
- Reviewers are sessions, so a reviewer in round two remembers round one and
  can say "you still have not removed the plugin system."
- Reviewers get the instruction, the worker's prompt, the transcript so far,
  and the workdir. They read files themselves; nothing is diffed for them.
- Every review is an agent turn, and reviewers run for the whole worker
  turn. Reviewers should be on the cheap tier.
- Why a built-in when the bake-off is not: the bake-off's decisions differ
  per workflow (how many, what counts as a finisher, when losers stop).
  Supervision's mechanics never differ; its decisions are who reviews and
  what they watch for, and those are the arguments. And it needs the event
  stream, which a workflow does not otherwise see.

Hierarchy is composition, not a third shape. A `Loop` body that calls
`Generate` with `Supervise` is a manager over a supervised worker; a `Loop`
whose body runs a `Loop` is a manager over leads. The one thing these do not
give is an agent that delegates to agents on its own, without a Go program
in between. That would be `Generate` exposed to the agent as a tool, not a
new Go shape, and it waits for a workflow that needs it.

## Observability

The process serves a web page for every run in the project, live or past,
that shows the graph of agents: every scope, every session, its turns, the
reviewers watching each turn, every steer, and the transcript behind each
node. A person at that page can steer or interrupt any
session through the same `Steer` and `Interrupt` the reviewers use.

The graph has two sources. The runtime supplies what happened: every
instance, and every edge, because structure is declared in scopes. A static
pass over the source supplies the same shape before it has run.

**Nodes are named at the call site.** Everything that is a node in the
graph takes a name where it is created, and everything that groups nodes is
a scope:

```go
err := gimble.Scope(ctx, "sprint", body)
g := gimble.Group(ctx, "bakeoff")
g.Go("attempt", body)
coder := gimble.NewSession(ctx, "coder", adapter, model, workdir)
validator, err := researcher.Fork(ctx, "validator")
loop := gimble.Loop(ctx, "sprint", goal, planner)
```

- The name identifies the node, not the instance. Five bake-off candidates
  are five sessions all named `candidate`; a loop body that creates a coder
  every lap creates N sessions named `coder`. The runtime tells instances
  apart by session id and scope instance, and the page draws them as cards
  stacked under the node. So names are compile-time constants, and the
  static pass checks that. A name built at runtime would hide a node from
  it.
- The name is the join key between the static pass and the log. Nothing
  else would work: the runtime cannot know its call site without
  reflection, which is out, and a `file:line` key would churn on every
  edit anyway. A name at the call site is also visible to whoever reads the
  workflow.
- Structure comes from scopes. A node is created with a ctx, the ctx says
  which scope it is in, and scopes nest. So containment is exact at
  runtime, for one line per segment, and a loop's laps and a bake-off's
  attempts are scopes that `Loop`, `Each`, and `Group.Go` make. Nothing is
  inferred. A session created in one scope and used in another shows its
  turns where they ran, which is information, not a problem.
- Reviewers are sessions, so they are nodes like any other. A person
  steering from the page is a supervisor node too, one the runtime adds
  itself since it is not in the source.

**The runtime** writes to the log what happened: each scope as it begins
and ends, each instance as it is created (name, id, scope, parent for a
fork), each turn's events (model calls, tool calls, tool results: the
transcript), each `Supervise` attachment, each lap of a `Loop`, each `Set`,
and each `Steer` with who sent it. That is the whole live graph.

**The static pass** draws the same graph before it has run: the template
the page shows with nothing lit up yet. Over the ssa form of an entry point
it finds every `Scope`, `Group`, `NewSession`, `Fork`, `Loop`, `Each`,
`Supervise`, `Set`, and `Get`, with their constant names and keys, and
traces each ctx from the scope that made it to the calls that receive it.
Constant keys add data edges: this scope writes `research`, that node's
prompt reads it. The limitations are Go's existing ctx conventions: ctx
travels as a parameter or a lexical capture, never in a struct; it is
derived only through the `context.With` family and Gimble's own functions,
so a `context.Background()` severs the chain; a call through an interface
resolves to every implementation; and conditional scoping yields a set of
possible scopes, not an error. The same analyzer hosts the lints under
Scope. It is not needed for the live view, so it comes after the server,
and it needs no agent: what an agent was going to guess about structure is
now declared.

The server is a `tylergannon/skgo` app: SvelteKit pages, Go remote
functions beside them, one binary. The known surface so far is five remote
functions, and there will be more as the page finds what it needs:

```go
// listRuns names every run in the project, live and past.
func listRuns(ctx context.Context) ([]RunInfo, error)

// watchRun replays every event of one run from its log, then pushes each
// new one until the run ends or the browser goes away. The page draws the
// graph from these. A past run replays and stops at its final event.
func watchRun(ctx context.Context, runID string, yield func(Event) error) error

// watchSession is the zoomed-in view: one session's turns and transcript,
// replayed then live.
func watchSession(ctx context.Context, runID, sessionID string, yield func(Event) error) error

// steer and interrupt call the Session methods of the same name on a live
// run, and error on a past one. ctx is the request's; a turn does not end
// because a tab closed.
func steer(ctx context.Context, runID, sessionID, message string) error
func interrupt(ctx context.Context, runID, sessionID string) error
```

- The watches read logs from disk, so a past run and a live one go through
  the same code, and the log's final event is what tells a watch to stop.
  `Generate` writes events there, and the log feeds the reviewers'
  transcripts, the page, and the post-mortem. Whether a workflow can also
  see events stays open.
- A steer from the page is recorded like a reviewer's steer, attributed to
  the person. In the graph the person is one more supervisor node, attached
  to every session at once. `Steer` takes no source argument; the run
  attributes by call path (`Supervise` knows its reviewer, the `steer`
  command knows it is the console, anything else is the workflow).
- Contexts follow the rule above. A live query's `ctx` is the request's, so
  a watch returns when the browser leaves; `steer` and `interrupt` are
  short calls on the request context; nothing the page does can end a turn
  except `interrupt`.
- Build order: the skgo server comes first. The first commit is the server
  over an empty runs directory, and the first run writes a log into it, so
  every primitive after it
  (`Scope`, `Generate`, `Steer`, `Supervise`, `Fork`) is visible in the
  page the day
  it is written, and the event schema gets answered by what the page needs
  to draw rather than guessed here.

## Still open

- Plain-text turns: every turn typed (e.g. `struct{ Notes string }`), or a
  separate text path.
- `Compact`: left out until a workflow needs it.
- `Fork`: whether it takes a workdir or inherits the parent's; which
  harnesses fork natively and which resume from a transcript copy.
- `Steer` when no turn is running: error, or drop like the legacy adapters.
  The page argues for an error, since a message a person typed should not
  vanish silently.
- Validation: once, in `Generate`, or also in the adapter.
- Session construction: settled that sessions are created from the ctx,
  which carries the run, and carry a name. Still open: the arguments after
  the name.
- `Run`: written in above as the proposal; not yet agreed. How a run is
  started in the long-lived process: the workflow named on the command
  line, a command from the page, or a CLI client talking to the server.
  Run ids, the runs directory layout, and whether session ids are unique
  across runs.
- Sessions closing with their scope: written in under Scope with the
  argument; agreed in principle ("maybe both"), not yet firmly.
- Interrupting a scope from the page (a lap, an attempt). The Loop section
  promises it and the server does not list it yet; it is a command that
  cancels one scope's ctx, distinct from `interrupt` on a session.
- Events: what `Event` carries, decided by what the page needs to draw.
  Cost and timing per turn are the obvious additions once the graph works.
  Whether a workflow can see events, or only reviewers and the page.
- The static pass: whether a non-constant name or key fails the build or
  only warns, and how much of the template the page shows before a run.
- Restart: whether rerunning reuses the run directory, and so the loop's
  file and the scope values, or starts fresh. Only if runs resume does
  set-once make a rerun segment check `Get` before doing its work again.
- `Set`'s scalar type set: which kinds, and whether `[]string` is in it or
  goes through polytype.
- `Each`: the name, and whether it earns its place once a workflow uses it.
- What `Task` carries beyond the planner's text: a lap number, the command
  results, the step it corresponds to in the file.
- `Supervise`: the round cap (a fixed default, an option, or `ctx` only) and
  exactly what the review prompt shows the reviewer.
- Multi-level hierarchy: which shapes it needs beyond nesting, once a
  workflow models one.
