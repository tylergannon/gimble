# Embedded MEDIUM workflow

decision: Keep `Project`, `Workdir`, and `Executable` as common materialization parameters and move the planning seed under `PlanParameters`; MEDIUM therefore needs no dummy planning input.

decision: Scope this sprint to the embedded workflow library and engine proof because CLI exposure, user documentation, and live harness proof are separately assigned to sprints 4, 5, and 6.

decision: Keep `workflow list` filtered to the still-runnable plan CLI until sprint 4 adds execution-specific argument handling; the library registry itself now orders MEDIUM before plan.
