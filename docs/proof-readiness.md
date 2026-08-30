# Promise-to-Proof contracts

A check can be valid and still prove the wrong promise. The failure that
motivated this feature was a boundary scenario being accepted as proof of the
primary user outcome. For example, showing that silence produces no invented
transcript is useful, but it does not show that qualifying speech produces the
expected visible words.

Tractor cannot discover product truth from fixture names, diagrams, tests, or
agent reports. A project can instead declare an optional `proof_contract`
before implementation and make its proof commitments executable.

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

## Contract design

A contract declares:

- `mode`: `delivery` or an explicitly non-terminal `discovery` prototype;
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
which cannot reach terminal product success.

Artifact requirements connect the design to execution without hard-coding a
domain architecture. Each relative workspace path has a role (`input`,
`output`, or `observation`) and a project-defined `architecture_edge`. Tractor
captures inputs before the assertion, captures outputs and observations after
it, copies them into engine-owned run evidence, and records SHA-256 digests.
The checkpoint evidence also records the run ID, case and node IDs, successful
graph edge, evidence mode, execution reference, and hashes of `outcome.json`
and `tool.log`. Terminal resume verifies those stored artifacts again.

See [`examples/proof-readiness/required-happy-path.yaml`](../examples/proof-readiness/required-happy-path.yaml)
for a complete generic contract. An ASR project could declare known human
speech and exact visible words as a primary case, then silence and an empty
transcript as a boundary case. Those facts remain project declarations, not
Tractor policy.

## Terminal success and fix routing

Each automated case binds to a distinct top-level `tool` node. Only execution
of that tool followed by its exit-zero `on_success` route creates proof
evidence. Nonzero exit routing is unchanged: it follows `on_error`, commonly
back to a fix node, and creates no proof pass.

Before admitting `success`, Tractor requires all declared primary and boundary
cases to have verified evidence from the same run. It refuses terminal success
when a case is missing, evidence is incomplete or changed, the contract is in
discovery mode, or a material scope gap remains. The failed gate saves a retry
continuation. A completed-node list, an architecture diagram, component
evidence for a separately declared operating-layer case, or a JSON record that
merely says `proven` cannot satisfy the gate.

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

No schema field is ASR-specific. The contract also does not assume that the
product is a formal state machine: `node` names only Tractor's executable
assertion, `architecture_edge` is a free-form evidence-chain label, and the
recorded graph edge is the assertion's actual success route. A game can put
state/seed in ordinary artifacts, while a time-dimensional simulation can put
versions, horizons, and constraints in ordinary qualifying criteria and
artifacts. Neither requires numbered product stages or recognizer concepts.

## Compatibility and known limits

`proof_contract` is optional. Existing pipelines and checkpoints continue
without migration, and no proof evidence is added to their checkpoints. A
pipeline that opts in must satisfy the generated schema and `proof_contract`
lint rule.

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
