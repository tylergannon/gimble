# Source leaf: `docs/spec.md`

## Purpose

Normative current-state contract for Gimble's Attractor variant. It is included
as a comparison leaf so the proposed Go direction cannot be mistaken for the
implementation currently shipped on `main`.

## Key concepts

- The specification is authoritative over implementation and other docs:
  `docs/spec.md:12-19`.
- The current runner uses directed graphs defined in a typed JSON schema:
  `docs/spec.md:19-19`.
- JSON is the current pipeline definition format and pipeline authors declare
  graph structure rather than write control flow:
  `docs/spec.md:51-67`.
- A current run parses JSON, validates the graph, initializes state, executes
  the graph, and finalizes it: `docs/spec.md:484-505`.

## Retrieval recipes

- For a direct current-versus-proposed comparison, read this leaf alongside
  `sources/workflows-as-programs.md`.
- For the authoritative full contract, open `docs/spec.md` itself.

## Themes

Current implementation; typed JSON graph; declarative workflow; graph
execution; status boundary.
