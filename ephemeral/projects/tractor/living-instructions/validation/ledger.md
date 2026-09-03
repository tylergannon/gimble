---
items:
  - name: P1 elicitation
    check: The design under validation/P1 proves P1 as stated in declaration.md and a coder cannot satisfy it while P1 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P1/design.md
  - name: P2 halt
    check: The design under validation/P2 proves P2 and cannot be satisfied while P2 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P2/design.md
  - name: P3 independent design review
    check: The design under validation/P3 proves P3 and cannot be satisfied while P3 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P3/design.md
  - name: P4 re-entry
    check: The design under validation/P4 proves P4 and cannot be satisfied while P4 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P4/design.md
  - name: P5 seven passes
    check: The design under validation/P5 proves P5 and cannot be satisfied while P5 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P5/design.md
  - name: P6 verify before chapter done
    check: The design under validation/P6 proves P6 and cannot be satisfied while P6 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P6/design.md
  - name: P7 supervisor steer
    check: The design under validation/P7 proves P7 and cannot be satisfied while P7 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P7/design.md
  - name: P8 library
    check: The design under validation/P8 proves P8 and cannot be satisfied while P8 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P8/design.md
  - name: P9 docs
    check: The design under validation/P9 proves P9 and cannot be satisfied while P9 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P9/design.md
  - name: P10 two seeds end to end
    check: The design under validation/P10 proves P10 and cannot be satisfied while P10 is false.
    doc: ephemeral/projects/tractor/living-instructions/validation/P10/design.md
---

# Validation designs

One item per promise (decision 47). In the manual run, Claude is the
`design` node and a codex session (`codex exec`, read-only, fresh) is the
`review` node; an item is marked when the reviewer routes pass. Items
carry no `command`: the pass edge is the verdict (decision 57).

Each `design.md` has: archetype; the story a verifier follows; the
evidence captured and where; the validator (the ledger `command` and
`infer` the promise's sprint will carry, or the proof script by name);
and what the design does not prove.

## Shared evidence tooling

Most scenarios run a nested `tractor workflow run` with a scripted
answerer. `validation/answerer.sh <interview-dir> <rules-file>` polls the
directory, matches each new question against ordered rules (a substring
and an answer), writes the answer with `tractor answer`, and appends
`question-id<TAB>rule` to `answers.log`. The rules file is part of the
scenario; the log is part of the evidence. It is built once, in chapter
5 sprint 3, and reused by every later scenario.
