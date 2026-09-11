// Package gimble runs agent workflows written as ordinary Go.
//
// A Runtime owns one project's durable run records and web application. The
// web application starts with the runtime unless WithNoWeb is used. Call
// Runtime.Run to establish the root scope for a workflow.
//
// Workflows use normal Go control flow. Scope names a bounded segment of work;
// Group provides an observable form of errgroup-style concurrency; Loop yields
// planner-selected tasks. Gimble supplies these runtime primitives, not named
// tactics: retries, critique rounds, bake-offs, and delivery methods remain
// visible in the workflow that needs them.
//
// NewSession creates a conversation owned by the current scope. Generate runs
// a blocking turn and returns either Text or a schema-bearing Output. Set,
// SetJSON, and ScopeText let the workflow explicitly choose which scoped data
// it places in a prompt; Generate does not inject context implicitly.
//
// Every operation follows context.Context. Returning from a scope closes its
// sessions, Group.Wait joins its children, and cancelling a run interrupts its
// agent work. See the package examples for complete, compiling uses of the
// runtime, groups, and planner loops.
package gimble
