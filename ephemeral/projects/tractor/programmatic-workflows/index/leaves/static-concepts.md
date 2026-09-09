# Static leaf: desired-state reconciliation and Go CFG limits

## Purpose

This leaf covers the two conceptual boundaries that matter for a Go-authored workflow proposal: a reconciliation loop repeatedly drives observed state toward a declared goal, while a compiler CFG exposes syntactic control flow without recovering domain-level routes, conditions, or intent.

## Findings

- Kubernetes defines a controller as a non-terminating control loop that moves current state toward desired state; the controller is level-based, so it reads current state rather than trusting the triggering event. `corpus/concepts/kubernetes--website/content/en/docs/concepts/architecture/controller.md:9-17,26-37,108-118`; pinned source: https://github.com/kubernetes/website/blob/8ddcb47f4571c91c0d0544eae477c1efdfdeefe5/content/en/docs/concepts/architecture/controller.md
- The controller-runtime contract makes the continuation policy explicit in a typed result: `RequeueAfter` schedules a future reconcile, a non-nil error normally requeues with exponential backoff, and `TerminalError` suppresses requeue. `corpus/concepts/kubernetes-sigs--controller-runtime/pkg/reconcile/reconcile.go.txt:29-52,109-126,172-191`; pinned source: https://github.com/kubernetes-sigs/controller-runtime/blob/91e819d2dbfe0f8c6e37cacec3b04c3004afc97a/pkg/reconcile/reconcile.go
- This is a useful analogy for Tractor's goal-seeking loop, but it does not supply exact continuation: reconciliation re-derives the next action from durable desired/current state, whereas Tractor's checkpoint names a graph node and counters and then may replay a stage. `corpus/concepts/kubernetes-sigs--controller-runtime/pkg/reconcile/reconcile.go.txt:62-75,102-126`; `corpus/static/tractor/engine/state.go.txt:12-25`; pinned URLs above and https://github.com/tylergannon/tractor/blob/07c04ff4c62ff91c625d2e27b6427cd594f67f39/engine/state.go
- The `golang.org/x/tools/go/cfg` package is structurally useful: it creates blocks for a function, records successors, materializes implicit returns, and emits DOT. `corpus/concepts/golang--tools/go/cfg/cfg.go.txt:5-14,52-69,138-185,220-260`; pinned source: https://github.com/golang/tools/blob/2af88d6fb782ffc8a0f607598345de595cd60c27/go/cfg/cfg.go
- The CFG explicitly omits conditions on conditional edges, short-circuit semantics, and panic control flow. Therefore `go/cfg` can support structural lint and a control-flow picture, but it cannot by itself recover Tractor's ordinary-language edge predicates, goal promises, artifact declarations, or whether a function call is a node boundary. `corpus/concepts/golang--tools/go/cfg/cfg.go.txt:35-41`; pinned source above.
- The CFG builder delegates `mayReturn` to a caller for call expressions and uses that policy to remove infeasible successors after calls such as `panic`, `os.Exit`, or `log.Fatal`; this is a policy hook, not whole-program effect inference. `corpus/concepts/golang--tools/go/cfg/cfg.go.txt:138-151`; `corpus/concepts/golang--tools/go/cfg/builder.go.txt:15-21,47-52`; pinned source: https://github.com/golang/tools/blob/2af88d6fb782ffc8a0f607598345de595cd60c27/go/cfg/builder.go

## Themes

- desired-state convergence versus exact replay;
- source-level structural visualization versus semantic workflow contracts;
- compiler facts versus policy annotations and domain-specific lint.

## Gotchas and counterevidence

- Reconciliation is not a proof that a graph can be discarded: it assumes a durable desired/current state model and idempotent or recoverable effects. A function that calls an agent or shell command can have effects that are not derivable from a CFG.
- A CFG can show all syntactic branches in one function, but route meaning may be hidden behind calls, interfaces, reflection, goroutines, dynamic data, or agent-produced values. The source itself says the CFG omits conditions and panic flow.
- Kubernetes' `Requeue` field is deprecated in favor of a duration; importing its API shape directly would preserve a known confusing policy. `corpus/concepts/kubernetes-sigs--controller-runtime/pkg/reconcile/reconcile.go.txt:29-46`.

## Retrieval recipes

- For “can Go visualize this program before run?”, start at `corpus/concepts/golang--tools/go/cfg/cfg.go.txt:5-14` and `:220-260`, then inspect Tractor's route and lint leaves.
- For “what survives if exact serialized continuation is relaxed?”, start at `corpus/concepts/kubernetes--website/content/en/docs/concepts/architecture/controller.md:108-132` and `corpus/concepts/kubernetes-sigs--controller-runtime/pkg/reconcile/reconcile.go.txt:109-126`, then compare `corpus/static/tractor/engine/state.go.txt:12-25`.

## Status and license

- Corpus snapshot and metadata retrieved 2026-09-08; repository commits, file hashes, and licenses are in `corpus/concepts/manifest.json`.
- Kubernetes controller-runtime source is Apache-2.0; Kubernetes website is CC-BY-4.0; `golang/tools` is BSD-3-Clause. License snapshots are adjacent under `corpus/concepts/`.

## Open questions

- Which subset of Go calls should count as workflow nodes, and how would the author annotate dynamic calls or agent-selected routes?
- Does a proposed runtime need a durable desired-state ledger sufficient to re-derive the next work item after process loss, or does the user require node-level continuation?
