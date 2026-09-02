# Execution workflows planning worklog

decision: reviewer answer 0009 keeps --project as the sole workflow parameter; materialization resolves the project directory and writes concrete checklist and doc paths without new graph substitution

decision: reviewer answer 0010 allows the LARGE chapter planner to call tractor ask only when an answer changes sprint scope or validation; it plans silently when the chapter contract is derivable

decision: reviewer answer 0011 makes a LARGE plan checklist.md the chapter ledger; every chapter item points to a chapter doc and initially empty sprint ledger written beneath the project, while MEDIUM remains flat

decision: reviewer answer 0012 makes --logs optional for plan, medium, and large; each defaults to a fresh run directory under Tractor's existing state root and prints that path before running, never storing logs in the committed project
