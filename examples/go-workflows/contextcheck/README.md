# Unadopted analyzer experiment

This code was written before Tyler rejected intraprocedural checking as
insufficient. It is preserved at his request, together with its tests.

It does not traverse the call graph. A successful exit does not validate a
workflow's context ownership. It is not enabled as a workflow check.
Its fixtures describe the earlier string-key API, before declared `Key` values.

Any adopted analyzer must traverse calls. That implementation is deferred;
the current work remains compiling workflow examples with stubbed backend
operations. See the [ownership note](../../../ephemeral/projects/gimble/programmatic-workflows/CONTEXT-OWNERSHIP.md).
