# Promise-to-Proof contracts

A check can be valid and still prove the wrong promise. The failure that
motivated this feature was a boundary scenario being accepted as proof of the
primary user outcome. For example, showing that silence produces no invented
transcript is useful, but it does not show that qualifying speech produces the
expected visible words.

Tractor cannot discover product truth from fixture names, diagrams, tests, or
agent reports. A project instead declares top-level workflow intent before
implementation: `mode: delivery` makes a `proof_contract` mandatory, while
`mode: discovery` permits contract-light learning without claiming delivery.

## Three different claims

- **Intended architecture** describes how the system is supposed to work.
  It is design context, not execution evidence.
- **Current-run evidence** records the exact assertion route and fingerprinted
  artifacts produced by this run. Static structure and prior runs cannot fill
  this role.
- **Terminal user outcome** is the observable promise to an actor doing a job.
  It is represented by one or more primary cases.

A **primary case** exercises input that meets explicit qualifying criteria and
checks the promised observable output. A **boundary case** checks an edge
condition such as empty or invalid input. Boundary evidence complements
primary evidence and never substitutes for it. A **correctness oracle** is the
assertion that decides whether the expected output occurred.

## Workflow mode and contract design

The top-level mode gives a route to `success` an honest meaning:

- `delivery` promises a user/product outcome. Validation fails without a
  `proof_contract`, and the run reports `COMPLETED` only after the contract's
  evidence gate passes.
- `discovery` promises learning rather than a product outcome. Its contract is
  optional, and a successful terminal route reports `LEARNING_COMPLETED`, not
  `COMPLETED`.

A proof contract declares:

- `intended_architecture` and the user-facing `primary_outcome`;
- primary and boundary cases, each with an actor, job, qualifying-input
  criteria, expected output, oracle, evidence surface/source, independence
  rule, status, capabilities, and optional artifact requirements;
- required dependencies/capabilities, known unknowns, and structured material
  `scope_gaps`; and
- every case required for `terminal_success`.

Case `status` is one of `proven`, `partial`, `blocked`, `simulated`, or
`unproven`. It states the project's pre-run assessment; even `proven` does not
replace current-run evidence. A delivery case must use the coherent triple
`tool_exit_zero` / `current_run` / `independent_execution`. Manual human
attestation is labeled explicitly and is accepted only in discovery mode,
where it cannot be mistaken for delivery evidence.

Artifact requirements connect the design to execution without hard-coding a
domain architecture. Each relative workspace path has a role (`input`,
`output`, or `observation`) and a project-defined `architecture_edge`. Tractor
captures inputs before the assertion, captures outputs and observations after
it, copies them into engine-owned run evidence, and records SHA-256 digests.
The checkpoint evidence also records the run ID, case and node IDs, successful
graph edge, execution reference, a digest of the complete case and bound tool
declarations, ordered snapshot digests, and hashes of `outcome.json` and
`tool.log`. Roles, architecture edges, capture timing, and snapshot locations
are derived from the digested contract instead of repeated in the checkpoint.
Terminal resume reconstructs those locations and verifies every stored
artifact again.

See [`examples/proof-readiness/required-happy-path.yaml`](../examples/proof-readiness/required-happy-path.yaml)
for a complete generic contract. A speech-transcription product could declare
known human speech and exact visible words as a primary case, then silence and
an empty transcript as a boundary case. Those facts remain project
declarations, not Tractor policy.

## Terminal meaning and fix routing

Each automated case binds to a distinct top-level `tool` node. Only execution
of that tool followed by its exit-zero `on_success` route creates proof
evidence. Nonzero exit routing is unchanged: it follows `on_error`, commonly
back to a fix node, and creates no proof pass.

Before reporting delivery `COMPLETED`, Tractor requires all declared primary
and boundary cases to have verified evidence from the same run. It refuses
delivery completion when a case is missing, evidence is incomplete or
changed, or a material scope gap remains. The failed gate saves a retry
continuation. A completed-node list, an architecture diagram, component
evidence for a separately declared operating-layer case, or a JSON record that
merely says `proven` cannot satisfy the gate.

Discovery still uses the graph's `success` pseudo-target as its authored end,
but the engine, CLI, MCP run store, resume path, and completion timeline expose
that end as `LEARNING_COMPLETED`. It is a successful learning run, not product
or user-outcome success.

A scope gap preserves the original promise alongside current proven behavior,
the missing capability, user impact, recommended next move, and whether a
decision is required. Tractor surfaces and blocks on the record; the workflow
or operator still decides whether an authorized in-scope fix can continue or a
cost, privacy, or direction choice needs human input.

## Applicability audit

The design was forward-checked against two unrelated domains without changing
either project:

| Contract concept | Small interactive game | 4D simulation / decision system |
|---|---|---|
| Primary case | Player takes an action under a declared rules version and deterministic/random seed. | Operator supplies a versioned scenario and complete, consistent constraints. |
| Qualifying input | Action is legal and unambiguous in the current game state; seed and rules version are recorded inputs. | Scenario/constraint versions meet declared completeness and consistency criteria. |
| Outcome | Visible resolved action, score, and result. | Calculated plan/decision and its user-visible presentation. |
| Oracle | Independently execute the seeded rule resolution and assert both calculated result and visible score/result. | Independently calculate/validate the decision and assert the visible plan/output. |
| Evidence chain | Action, seed, and rules artifacts on input edges; resolution and visible result artifacts on output edges. | Scenario and constraints on input edges; calculated plan and rendered decision on output edges. |
| Boundary case | Invalid or ambiguous action is rejected without changing score as though it were valid. | Incomplete or conflicting data produces the declared non-decision/error outcome, not a normal plan. |

No schema field is domain-specific. The contract also does not assume that
the product is a formal state machine: `node` names only Tractor's executable
assertion, `architecture_edge` is a free-form evidence-chain label, and the
recorded graph edge is the assertion's actual success route. A game can put
state/seed in ordinary artifacts, while a time-dimensional simulation can put
versions, horizons, and constraints in ordinary qualifying criteria and
artifacts. Neither requires numbered product stages or domain-specific
processing concepts.

## Compatibility and known limits

Workflow `mode` is optional only as a legacy compatibility path. Pipelines
that omit it keep the prior behavior: they validate and may report `COMPLETED`
without a proof contract. A legacy pipeline that already declares a contract
still receives the proof gate. Existing checkpoints need no migration, and no
proof evidence is added when no contract exists.

New and materially edited pipelines should migrate explicitly:

1. Choose `mode: delivery` and add a complete proof contract when the workflow
   promises a product/user outcome.
2. Choose `mode: discovery` when the authorized outcome is learning,
   feasibility, or unknown reduction; update status consumers to accept
   `LEARNING_COMPLETED` as terminal but not delivered.
3. Leave mode absent temporarily only when preserving an existing workflow's
   completion behavior is more important than immediate migration.

No automatic rewrite or deprecation warning is introduced in this change, so
existing CLI output and MCP records remain stable until a pipeline opts in.

The mechanism validates declarations, independently executes commands, and
preserves current-run provenance. It cannot know whether an author selected a
truly qualifying fixture, wrote the correct oracle, or made an
`operating_layer` command reach the claimed product surface. Local evidence is
auditable but not a cryptographically trusted remote attestation. Artifact
capture currently snapshots whole regular files in memory, so it is not
suited to very large evidence artifacts. Proof tools must be top-level nodes
rather than parallel-branch nodes. Required capabilities are declared and
referenced; their real availability must be exercised by the proof command.
Discovery and partial work can be represented honestly, but never as terminal
delivery success.
