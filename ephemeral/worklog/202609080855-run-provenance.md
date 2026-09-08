decision: Tyler authorized issue #71 as foundational run provenance and explicitly required docs/spec.md to define the persisted contract.
decision: Persist provenance as an ordered invocations array so resume retains every argv, executable identity, pipeline source, and resolved graph hash instead of overwriting the original evidence.
friction: Terminal-checkpoint resume returned before manifest initialization, so its process invocation was invisible -> record provenance before the terminal fast path as well as ordinary fresh and resumed walks.
