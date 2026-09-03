# Promise-adjacent seams

A seam is a boundary between parts. Most of them are nobody's business
but the implementer's. Record a seam in the brief only when a promise
crosses it.

A seam crosses a promise when it is one of:

- a promise to another party: another project, a caller, a client that
  is not in this repository;
- a persisted format: a file, a ledger, a run directory, anything a
  later version or another tool must read;
- a public surface: a CLI flag, an API, a documented path.

For each such seam write the parties and the contract in one line, and
name the document that defines it. Everything else is internal: the
implementer owns it, may move it, and need not ask.

The test for whether a seam is placed well (Parnas): each side can
change how it does its job without the other side knowing. If a change
on one side forces a change on the other, the seam is in the wrong
place or the contract is too wide.

Two failures, both real:

- Boundary erosion: reaching across a seam because it was convenient.
  Ask instead; improvising at a seam you do not own is a defect.
- Boundary worship: recording every internal edge as a contract until
  nothing can move. Defer by default; a seam earns a line in the brief
  only through a promise.

Source: decisions.md 40; sources/proof-practice/nlspec-methodology/methodology.md (the Parnas seam-quality criterion).
