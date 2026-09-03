# Research findings

Findings from research that the brief should consider. Each names the
promise or design point it bears on, the evidence, and a proposed change.
Research never edits the declaration; the brief lap asks about these and
records the outcome. An entry is closed when the brief has ruled.

## Closed by the brief (interview 0013, all accepted)

### F1. P8 "show is byte-equal to the engine's prompt.md" is ill-defined (R5, R1)

The engine composes `prompt.md` as frame plus `$goal`-expanded prompt
(`engine/codergen.go:42-58`); frames inline the selected item, the last
failure, and the doc file (`engine/frames.go:108-168`). Every node in
`medium` and `large` runs inside a loop, so only the frameless planner
node is even comparable. Proposal: restate P8 as "`show` prints, for each
node, the prompt exactly as `Build` materialized it for the given
parameters" and add an optional `--stage <dir>` mode that diffs against a
real stage with the frame stripped. R1 agrees from the other side: the
two tools that achieve byte-equality (gemini-cli, codex) do it by having
the live call write the string, not by recomputing it. See
`workflow-package-inventory/migration-inventory.md`,
`prompt-libraries/gemini-cli.md`, `prompt-libraries/codex.md`.

### F4. P8 "no built-in prompt lives in a Go string" needs a scope line (R1)

Every surveyed tool that moved prompt bodies to files still builds some
prompt text in code: crush formats git status lines with `fmt.Sprintf`,
sst/opencode assembles the `<env>` block in code, codex keeps header
sentences as consts. Proposal: restate as "every prompt body, doctrine
page, supervisor brief, pass, and skeleton is a library file; Go supplies
only data values (paths, names, commands) to templates". See
`prompt-libraries/crush.md`, `opencode-sst.md`, `codex.md`.

### F5. Template delimiters (R1)

`{{` in doctrine or skeleton text collides with `text/template`; fabric
avoided the engine for this reason. Proposal: set `Delims("<<", ">>")` or
similar before any page is authored, and say so in the library README.

### F2. Six one-question passes lose cross-cutting judgment (R2)

Multi-Crit (arXiv:2511.21662) shows separate-criterion judging is more
faithful per criterion but joint judging is where trade-offs between
criteria are seen; Mahmoud et al. (arXiv:2605.12474) show per-criterion
verifiers credit mere mention, and a cross-family panel still preferred
rubric-hacked output on rubric terms while rubric-free judges did not.
Proposal: a seventh pass, holistic and rubric-free: "would you accept
this plan as the plan for this product, and what is the one thing that
would stop you". See `review-and-verification/xiong-2025-multi-crit.md`
and `mahmoud-2026-reward-hacking-in-rubric-based-rl.md`.

### F3. Cheap-role model name (R4)

`claude-sonnet-5` is not a name any backend here accepts; Tractor
validates only non-emptiness, so a bad name fails at run time in the
vendor CLI. Reachable cheap names on this machine: the `sonnet` alias on
the claude backend; gemini flash names through agy with an explicit
`llm_provider: gemini`. Proposal: `sonnet` on claude for replan, judges,
and halt; always set the provider explicitly in `models.yaml`. See
`models-and-providers/tractor-provider-selection.md`.

## Closed without a promise change (no question asked)

- **Provider separation is supported by mechanism, not measured directly**
  (R2: Preference Leakage never measures an unrelated provider). Keep it;
  add a doctrine requirement that every reviewer and verifier reasons
  before it routes, which Chen et al. (arXiv:2504.03846) show is the
  largest single reducer of harmful self-preference. Content, not a
  promise.
- **The holdout must be unreachable, not merely unread** (R2: METR field
  report, ImpossibleBench). Our XDG-root location is outside the readable
  tree; the remaining hole is grepping stage prompts, already recorded in
  decision 43 as accepted.
- **An operating verifier has no measured precedent** (R2). Do not cite
  prior art as proving it; instrument it later. No change to P6.
- **Capture-format rule for scenario evidence** (R3): ordered per-step
  images plus one text artifact (HAR, accessibility snapshot, or console
  or terminal log); video is a human extra; TUI evidence is rendered
  frames, not the byte stream. Goes in the `validation-archetypes`
  doctrine page. No third archetype.
- **Pixel diffs are valid only in an identical environment** (R3).
  Universal-archetype screenshot baselines need the same container as the
  holdout capture. Doctrine, not promise.
- **Local tooling** (R3): Playwright is cached; vhs, asciinema,
  agent-browser, and a working ffmpeg are absent. This project's own
  software is a CLI, so its scenario evidence is transcripts and exit
  codes; no browser tooling is needed for P6 or P10.
- **`models.yaml` needs no engine change** (R4); supervisors must be
  authored into the YAMLs; effort is advisory for gemini names that
  encode it.
