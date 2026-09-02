---
items:
  - name: Print a greeting
    check: Running `sh hello.sh` prints exactly HELLO_LOOPS
    command: test "$(sh hello.sh)" = HELLO_LOOPS
  - name: Add two numbers
    check: Running `sh calc.sh add 2 3` prints 5
    command: test "$(sh calc.sh add 2 3)" = 5
  - name: Document the scripts
    check: NOTES.md tells a newcomer how to run both scripts
    command: test -s NOTES.md
    infer:
      files: NOTES.md
      prompt: Judge whether a newcomer could run both scripts from these notes alone
---

# Demo checklist

Three small claims about this workspace, worked through by
`checklist-loop.yaml`. The engine marks an item `done: true` only after its
command exits 0 and, where present, its `infer` judge passes; agents never
write that field. Edit the open items freely — the loop re-reads this file
on every arrival. Everything below the frontmatter is for agents and people;
the engine never reads it.
