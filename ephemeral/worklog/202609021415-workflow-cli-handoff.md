# Workflow CLI and handoff sprint worklog

decision: the public handoff prints absolute artifact paths so callers can locate outputs even when workflow workdir differs from their shell directory

decision: workflow CLI tests inject only the shared pipeline runner call, preserving real input validation, graph materialization, environment setup, artifact parsing, and handoff formatting
