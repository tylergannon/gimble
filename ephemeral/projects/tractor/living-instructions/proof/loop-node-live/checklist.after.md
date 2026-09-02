---
title: Loop node live proof
items:
  - name: greet script
    check: Running `sh greet.sh` prints exactly GREETING_2d171d08 and nothing else
    command: test "$(sh greet.sh)" = GREETING_2d171d08
    done: true
  - name: line count
    check: count.txt contains the number of lines in greet.sh, as a bare integer
    command: test "$(cat count.txt)" = "$(wc -l < greet.sh | tr -d ' ')"
    done: true
  - name: notes
    check: notes.md explains what greet.sh prints and how count.txt is derived
    infer:
      files: notes.md
      prompt: Does notes.md accurately state what greet.sh prints and how count.txt is derived? Compare against the actual files.
    done: true
---

# Live proof of the loop node

Three items. The first two are checked by commands with a nonce the agent
cannot guess. The third is checked by a judge reading `notes.md`.
