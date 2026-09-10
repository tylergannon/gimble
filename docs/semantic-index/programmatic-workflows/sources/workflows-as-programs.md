# Source leaf: `docs/workflows-as-programs.md`

## Purpose

Focused edited synthesis of the proposal to author Gimble workflows as Go
programs and organize each agent's objective, responsibility, and information
path together.

## Key concepts

- Core proposal and unimplemented context/indexing direction:
  `docs/workflows-as-programs.md:1-24`.
- General graph authoring is hard; a shorthand trends toward a programming
  language; Go supplies native control flow and concurrency:
  `docs/workflows-as-programs.md:26-53`.
- Reusable method and changing task/index data are separate:
  `docs/workflows-as-programs.md:55-62`.
- Source should preserve the legibility of a good diagram and read like
  pseudocode: `docs/workflows-as-programs.md:64-95`.
- One context includes immediate prompt material plus retrievable indexed
  files, organized around success: `docs/workflows-as-programs.md:97-122`.
- The first message should orient an agent toward outcome, responsibility,
  constraints, current state, and retrieval paths: `docs/workflows-as-programs.md:124-155`.
- Builder, supervisor, and checker divide attention without erasing the larger
  goal: `docs/workflows-as-programs.md:157-187`.
- Existing unmerged POC provides ordinary Go workflows, an iterator, typed
  agent results, and separated inputs/prompts:
  `docs/workflows-as-programs.md:189-197`.
- Earlier concurrency sketches remain proposed; newer library sketches need
  revision against that prior work: `docs/workflows-as-programs.md:199-206`.
- The shipped graph engine remains the README's subject:
  `docs/workflows-as-programs.md:208-210`.

## Retrieval recipes

- For “why Go?”, open `docs/workflows-as-programs.md:26-53`.
- For context/indexing, open `docs/workflows-as-programs.md:97-155`.
- For role decomposition and tradeoffs, open `docs/workflows-as-programs.md:157-187`.
- For the recovered POC and Go sketches, follow the links at
  `docs/workflows-as-programs.md:189-206`.

## Themes

Go as orchestration substrate; workflow legibility; one agent context;
success-oriented retrieval; scope and stack metaphor; distributed attention;
direction versus implementation; recovered iterator POC; Go concurrency sketches.
