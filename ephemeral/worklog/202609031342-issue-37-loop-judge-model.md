decision: ISSUE-37 explicitly requires the change to remain uncommitted in the current working tree, so no task worktree, branch, commit, or push is created despite the general repository protocol.
decision: Isolating the loop judge requires clearing its execution-time view of pipeline model defaults as well as stopping parse-time inheritance; the generic Codergen resolver otherwise reapplies those defaults.
decision: The judge's default provider comes from resolving the flash alias rather than a hard-coded provider fallback, so a loop that overrides only llm_model still auto-detects that model's provider.
decision: A real CLI loop run with OpenAI pipeline defaults routed its infer judge through agy as gemini-3.8-flash-medium and completed after the judge inspected the evidence file.
