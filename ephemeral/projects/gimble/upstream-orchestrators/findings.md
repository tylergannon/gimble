# Upstream agent orchestrators

Research prompted by a naming question: whether the name Orca was available. It
is not, and the four projects carrying it turned out to be worth reading on
their merits. Sources were the public repositories and documentation sites; no
project was cloned, built, or run, so every judgement below rests on documented
behaviour rather than observed behaviour.

Everything here is agent research and agent proposal. Nothing has been ratified.

## The projects

### stablyai/orca

Not an orchestration engine. A desktop application that wraps agent CLIs in
terminals, with a Monaco editor, an embedded per-worktree browser, and a diff
viewer. Widely starred and actively developed.

There is no workflow file, no graph, and no lint step. Orchestration is an
experimental mode in which a coordinator model reads a prose skill guide and
issues CLI calls to create tasks and start workers. A worker reports completion
by composing its own CLI call in its terminal; the coordinator then reads the
prose body of that message and decides what happens next. Dependencies between
tasks are a JSON array typed at a shell prompt by a model at runtime.

Their documentation carries a large defensive vocabulary around absent and
unverifiable worker state, and warns the coordinator never to reconstruct
arguments it did not observe. That is the cost of having no engine-owned
routing decision to record.

Their genuine strengths are human-facing: git worktree management, a status
dashboard across worktrees, notification routing, a companion phone application
that acts as a remote control over a paired desktop, and diff review. Agents are
launched with permission checks disabled on the theory that the worktree is the
sandbox.

### ThakeeNathees/orca

A declarative orchestration language with an HCL-like block syntax and a
compiler that emits Python targeting LangGraph. Solo project, frozen partway
through a rewrite; the README points at a branch that does not exist, and the
actual work-in-progress branch has been idle for months. Test hygiene is better
than the project's visibility suggests, with golden-file codegen tests and
coverage gates.

Its control flow is a directed acyclic graph with conditional edges. Entry and
exit points are inferred from topology rather than declared, so an unattached
node silently becomes a second entry point instead of an error. Branching runs
the predecessor's raw output through a transform function and matches the result
against literal string keys; an unmatched key falls through to termination
without an error. Loops are expressible only by aliasing a node and routing back
to it, and there is no iteration bound. An open issue asks for one.

The compiler checks block schema conformance, references, member access, and
expression types, and performs constant folding. It does not check the workflow
graph for cycles or reachability, does not check that branch route keys are
exhaustive, and cannot verify that a transform's output range covers the routes.

Two ideas in it are good and are described under borrow candidates below. The
surrounding expression language, which includes lambdas, currying, recursion and
ternaries, exists to serve one optional string transform that in the project's
own test fixture is the identity function.

### jascal/orca-lang

A hierarchical statechart language whose surface syntax is Markdown headings and
tables. Machines declare context fields, events, states, transitions and guards.
It is aimed at a different problem than Tractor: models are the authors of these
machines, not the nodes inside them, and no model runs at execution time. It is
maintained, published to npm, and has real test suites across several runtime
languages.

Its verifier is the strongest static analysis of anything reviewed. It checks
reachability, deadlock, orphaned events and actions, completeness of event
handling, transition determinism, guard exhaustiveness, and cross-machine
invocation cycles. Author-declared properties are verified by bounded search
over the state space with a counterexample path returned on failure. Guard
expressions are deliberately restricted to a decidable fragment, with method
calls excluded, precisely so the verifier can be complete.

The Markdown surface is a real cost: no schema validation, no editor support,
and a specification that has to shout at the author because the format cannot
enforce its own rules. The bounded search also has a hard state ceiling.

### VirtusLab/orca

The most mature project of the four and the closest competitor. An imperative
embedded language in Scala, run as scripts. Five agent backends, a long series
of architecture decision records, and evidence of the authors running their own
review loop against their own codebase.

The important data point is that the one serious competitor chose imperative
scripting over a declarative graph, deliberately and with the trade-off
documented. Their bet is durability rather than static analysis. A stage is
simultaneously a memoised checkpoint, a commit boundary, a log entry, and a
resume unit, and the progress record is committed alongside the code the run
produced, so the record cannot drift from the working tree. A run interrupted
anywhere resumes by re-running the identical command.

Agent capability is a type-level axis rather than a configuration field, with a
documented matrix of how strictly each backend enforces it. The runtime owns
git, and every write-capable agent turn is told not to commit, push, or switch
branches.

The cost of their choice is that nothing is inspectable before it runs. There is
no graph to lint, visualise, or diff, and no way to answer whether a flow can
reach its final step without review other than by reading the program. Their
escape hatches are also load-bearing and enforced at runtime rather than by the
type system that is the project's whole pitch.

## Where Tractor stands

No project reviewed has schema-enforced route selection at an agent node. The
closest, ThakeeNathees, matches agent prose against string keys and terminates
silently on a miss. VirtusLab has schema-enforced structured output but no
concept of offered successors, because control flow is program control flow.
jascal has no agents at run time. stably reads a prose body and decides.

The constraint that the agent picks from a closed set the engine hands it, and
cannot invent a route, is where Tractor is ahead of all four. The README states
this but does not present it as a differentiator, and it should.

Bounded loops are similarly unmatched. ThakeeNathees has no iteration bound at
all, jascal bounds by guard on accumulated context, VirtusLab by an argument to
a library helper, stably not at all.

## Borrow candidates

Ordered by the reviewing agent's estimate of value against cost. None ratified.

### Author-declared graph properties, checked during lint

From jascal. The author states an intent the node and edge listing cannot
express, and the linter proves or refutes it before the run starts. Useful
properties are reachability of a node, a requirement that no path reaches one
node without passing through another, and a requirement that a node stays
reachable from everywhere.

Tractor already builds the whole graph before spending a token, so each of these
is a short search over a structure that is already in memory. This is the
clearest capability gap found and the cheapest to close.

### Exhaustive outcome handling with an explicit ignore

Also from jascal. A command node that does not enumerate an edge for an outcome
it can produce should be a lint error, which the author silences by naming the
case as ignored rather than by leaving it out. Tractor claims that a typo can
never silently change a run, but that claim currently covers only unknown
fields, not missing routes.

### Progress record committed with the work

From VirtusLab. Tractor's run directory sits outside the repository, so a killed
run can leave a working tree and a record that disagree. Committing the record
alongside the produced code removes that class of drift. Most valuable for
worktree fan-out, where a killed run currently leaves worktrees that nothing
accounts for.

### Roles as an indirection over backends

Also from VirtusLab. A workflow names the kind of agent a step needs rather than
a specific backend, and a settings file binds those roles to the real CLIs. This
makes cross-model fan-out a configuration change instead of a graph edit, and
keeps model names out of the pipeline body.

### Line-anchored diff comments batched into one revision prompt

From stably, and the one idea in that product worth taking. Review comments
attach to lines and follow those lines as the agent edits, then go to the agent
as a single batch. Their stated reason for batching is that feeding comments one
at a time makes the agent oscillate. Tractor has no analogue, and the web UI is
the natural home for it.

### Schemas expressed in the pipeline language itself

From ThakeeNathees, whose compiler defines its own block types as data files in
its own language, with one description string driving parse errors, editor hover
text, and documentation. This is the cheapest defence against node-type
documentation drifting away from the Go definitions.

### Separating node identity from node definition

Also from ThakeeNathees. Their workflow block can map a graph position to an
existing definition, so one agent definition can appear at two positions in a
graph. Tractor currently gets that only by copying the node.

## Rejected

The expression language around the ThakeeNathees branch transform, which buys
nothing over a typed configuration format for the one place it is used.

Markdown as a surface syntax, which trades away schema validation and editor
support for model authorability that a typed format already provides.

The whole of the stably product as an engine reference. It is worth studying as
an interface reference and nothing else.
