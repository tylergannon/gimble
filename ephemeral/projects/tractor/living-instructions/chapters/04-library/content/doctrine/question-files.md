# Question files

You reach the human through files. `tractor ask <file>` moves the file
into the run's interview directory under the next number, records a
`QuestionAsked` event, blocks until `<n>.answer.md` appears beside it,
and prints the answer. `tractor answer <question> [text]` writes the
answer. Nothing else is a question.

Write the file, then run the command. If your shell tool times out
while waiting, run the same command again with the moved path
(`interview/<n>.md`); it resumes waiting without renumbering.

One file per independent group of questions. Inside the file, number
the questions; the answer file mirrors the numbering. Each question
carries its options and your recommendation.

A promise candidate is one line beginning `Promise:` followed by the
statement, so a reader, human or script, can find every candidate
without reading prose. Put the must-not-imply and your recommendation
on the lines after it.

```
## 2. Data loss

Promise: The ledger file is never truncated by a failed write.
Must not imply: durability across machine crashes.
Recommend: yes; write to a temp file and rename.
```

Ask when the sprint doc leaves a decision open, or when you would
otherwise guess at something the human would care about. Do not ask
what the docs settle. Do not batch unrelated questions into one file;
do not split one decision across files.

An answer is an instruction. Record what it changed (a promise, an
exclusion, a scope) in the brief before you continue, so the next lap
does not ask again.

Source: decisions.md 26-29, 39; BUILD.md ("Asking questions"); declaration.md section 4 (the question-file seam).
