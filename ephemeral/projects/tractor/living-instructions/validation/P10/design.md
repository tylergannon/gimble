# P10: two known seeds plan and execute end to end

Archetype: scenario, twice.

## Story

Seeds written before chapter 5 sprint 1 starts, committed under
`seeds/`:

- `seeds/greeter.md`: a CLI that greets by name, with a flag for
  shouting and a file of names; expected to size MEDIUM.
- `seeds/ledger-tool.md`: a CLI that keeps a markdown ledger of items
  with add, done, and list, plus a validation subcommand and an
  export; expected to size LARGE with two chapters.

For each: scratch repository; `plan` with `answerer.sh` accepting
everything and answering "begin" to open prompts; the printed `Next:`
handoff is run verbatim; wait for `COMPLETED`.

## Evidence

- Both run directories in full (timeline, stages, events, supervisors).
- Both packages.
- The scratch repositories at the end, with the built software.

## Validator

`command`: `prove/p10-end-to-end.sh`: for each seed, `plan` completed,
`validate-plan` accepts the package, `recommendation.md` names the
expected size, the handoff command ran to `COMPLETED`, and the scratch
repository's own tests pass; for the LARGE seed, P6's script passes on
its run.

`infer` (files: each scratch repository's README and the seed): "Does
the built software do what the seed asked? Run it. Fail if a named
feature is missing or the program does not run."

## Not proven

Generality beyond seeds of this size.
