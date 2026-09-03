---
items:
  - name: <unique name; for a validation ledger, the promise id>
    check: <the promise as observable behaviour>
    command: <shell from the workdir; exit 0 passes; omit for a chapter or a reviewed item>
    infer:
      files:
        - <glob relative to the workdir>
      prompt: <what the judge decides about those files>
    doc: <path to the sprint or design doc>
---

# <Ledger title>

<Prose the engine never reads: definition of done, context, notes. For
a chapter ledger, one paragraph per chapter naming its doc and sprint
ledger. For a sprint ledger, the backlog sketch below the items.>
